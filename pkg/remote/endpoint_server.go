package remote

import (
	contextpkg "context"
	"net"

	"github.com/pkg/errors"

	"github.com/golang/protobuf/proto"

	"github.com/RokyErickson/doppelganger/pkg/compression"
	"github.com/RokyErickson/doppelganger/pkg/doppelganger"
	"github.com/RokyErickson/doppelganger/pkg/encoding"
	"github.com/RokyErickson/doppelganger/pkg/protocols/local"
	"github.com/RokyErickson/doppelganger/pkg/rsync"
	"github.com/RokyErickson/doppelganger/pkg/session"
	"github.com/RokyErickson/doppelganger/pkg/sync"
)

type endpointServer struct {
	encoder  *encoding.ProtobufEncoder
	decoder  *encoding.ProtobufDecoder
	endpoint session.Endpoint
}

func ServeEndpoint(connection net.Conn, options ...EndpointServerOption) error {
	defer connection.Close()

	if err := sendMagicNumber(connection, serverMagicNumber); err != nil {
		return &handshakeTransportError{errors.Wrap(err, "unable to send server magic number")}
	}

	if magicOk, err := receiveAndCompareMagicNumber(connection, clientMagicNumber); err != nil {
		return &handshakeTransportError{errors.Wrap(err, "unable to receive client magic number")}
	} else if !magicOk {
		return &handshakeTransportError{errors.New("client magic number incorrect")}
	}

	if err := doppelganger.SendVersion(connection); err != nil {
		return &handshakeTransportError{errors.Wrap(err, "unable to send server version")}
	}

	clientMajor, clientMinor, clientPatch, err := doppelganger.ReceiveVersion(connection)
	if err != nil {
		return &handshakeTransportError{errors.Wrap(err, "unable to receive client version")}
	}

	versionMatch := clientMajor == doppelganger.VersionMajor &&
		clientMinor == doppelganger.VersionMinor &&
		clientPatch == doppelganger.VersionPatch
	if !versionMatch {
		return errors.New("version mismatch")
	}

	reader := compression.NewDecompressingReader(connection)
	writer := compression.NewCompressingWriter(connection)

	encoder := encoding.NewProtobufEncoder(writer)
	decoder := encoding.NewProtobufDecoder(reader)

	endpointServerOptions := &endpointServerOptions{}
	for _, o := range options {
		o.apply(endpointServerOptions)
	}

	request := &InitializeRequest{}
	if err := decoder.Decode(request); err != nil {
		err = errors.Wrap(err, "unable to receive initialize request")
		encoder.Encode(&InitializeResponse{Error: err.Error()})
		return err
	}

	if endpointServerOptions.root != "" {
		request.Root = endpointServerOptions.root
	}

	if endpointServerOptions.configuration != nil {
		if err := endpointServerOptions.configuration.EnsureValid(
			session.ConfigurationSourceTypeAPIEndpointSpecific,
		); err != nil {
			err = errors.Wrap(err, "override configuration invalid")
			encoder.Encode(&InitializeResponse{Error: err.Error()})
			return err
		}
		request.Configuration = session.MergeConfigurations(
			request.Configuration,
			endpointServerOptions.configuration,
		)
	}

	if endpointServerOptions.connectionValidator != nil {
		err := endpointServerOptions.connectionValidator(
			request.Root,
			request.Session,
			request.Version,
			request.Configuration,
			request.Alpha,
		)
		if err != nil {
			err = errors.Wrap(err, "endpoint configuration rejected")
			encoder.Encode(&InitializeResponse{Error: err.Error()})
			return err
		}
	}

	if err := request.ensureValid(); err != nil {
		err = errors.Wrap(err, "invalid initialize request")
		encoder.Encode(&InitializeResponse{Error: err.Error()})
		return err
	}

	endpoint, err := local.NewEndpoint(
		request.Root,
		request.Session,
		request.Version,
		request.Configuration,
		request.Alpha,
		endpointServerOptions.endpointOptions...,
	)
	if err != nil {
		err = errors.Wrap(err, "unable to create underlying endpoint")
		encoder.Encode(&InitializeResponse{Error: err.Error()})
		return err
	}
	defer endpoint.Shutdown()

	if err = encoder.Encode(&InitializeResponse{}); err != nil {
		return errors.Wrap(err, "unable to send initialize response")
	}

	server := &endpointServer{
		endpoint: endpoint,
		encoder:  encoder,
		decoder:  decoder,
	}

	return server.serve()
}

func (s *endpointServer) serve() error {

	request := &EndpointRequest{}

	for {

		*request = EndpointRequest{}
		if err := s.decoder.Decode(request); err != nil {
			return errors.Wrap(err, "unable to receive request")
		} else if err = request.ensureValid(); err != nil {
			return errors.Wrap(err, "invalid endpoint request")
		}

		if request.Poll != nil {
			if err := s.servePoll(request.Poll); err != nil {
				return errors.Wrap(err, "unable to serve poll request")
			}
		} else if request.Scan != nil {
			if err := s.serveScan(request.Scan); err != nil {
				return errors.Wrap(err, "unable to serve scan request")
			}
		} else if request.Stage != nil {
			if err := s.serveStage(request.Stage); err != nil {
				return errors.Wrap(err, "unable to serve stage request")
			}
		} else if request.Supply != nil {
			if err := s.serveSupply(request.Supply); err != nil {
				return errors.Wrap(err, "unable to serve supply request")
			}
		} else if request.Transition != nil {
			if err := s.serveTransition(request.Transition); err != nil {
				return errors.Wrap(err, "unable to serve transition request")
			}
		} else {
			return errors.New("invalid request")
		}
	}
}

func (s *endpointServer) servePoll(request *PollRequest) error {

	if err := request.ensureValid(); err != nil {
		return errors.Wrap(err, "invalid poll request")
	}

	pollContext, forceResponse := contextpkg.WithCancel(contextpkg.Background())
	defer forceResponse()

	responseSendResults := make(chan error, 1)
	go func() {
		if err := s.endpoint.Poll(pollContext); err != nil {
			s.encoder.Encode(&PollResponse{Error: err.Error()})
			responseSendResults <- errors.Wrap(err, "polling error")
		}
		responseSendResults <- errors.Wrap(
			s.encoder.Encode(&PollResponse{}),
			"unable to send poll response",
		)
	}()

	completionReceiveResults := make(chan error, 1)
	go func() {
		request := &PollCompletionRequest{}
		completionReceiveResults <- errors.Wrap(
			s.decoder.Decode(request),
			"unable to receive completion request",
		)
	}()

	var responseSendErr, completionReceiveErr error
	select {
	case responseSendErr = <-responseSendResults:
		completionReceiveErr = <-completionReceiveResults
	case completionReceiveErr = <-completionReceiveResults:
		forceResponse()
		responseSendErr = <-responseSendResults
	}

	if responseSendErr != nil {
		return responseSendErr
	} else if completionReceiveErr != nil {
		return completionReceiveErr
	}

	return nil
}

func (s *endpointServer) serveScan(request *ScanRequest) error {

	if err := request.ensureValid(); err != nil {
		return errors.Wrap(err, "invalid scan request")
	}

	snapshot, preservesExecutability, err, tryAgain := s.endpoint.Scan(nil)
	if tryAgain {
		response := &ScanResponse{
			Error:    err.Error(),
			TryAgain: true,
		}
		if err := s.encoder.Encode(response); err != nil {
			return errors.Wrap(err, "unable to send scan retry response")
		}
		return nil
	} else if err != nil {
		s.encoder.Encode(&ScanResponse{Error: err.Error()})
		return errors.Wrap(err, "unable to perform scan")
	}

	buffer := proto.NewBuffer(nil)
	buffer.SetDeterministic(true)
	if err := buffer.Marshal(&sync.Archive{Root: snapshot}); err != nil {
		return errors.Wrap(err, "unable to marshal snapshot")
	}
	snapshotBytes := buffer.Bytes()

	engine := rsync.NewEngine()

	delta := engine.DeltafyBytes(snapshotBytes, request.BaseSnapshotSignature, 0)

	response := &ScanResponse{
		SnapshotDelta:          delta,
		PreservesExecutability: preservesExecutability,
	}
	if err := s.encoder.Encode(response); err != nil {
		return errors.Wrap(err, "unable to send scan response")
	}

	return nil
}

func (s *endpointServer) serveStage(request *StageRequest) error {

	if err := request.ensureValid(); err != nil {
		return errors.Wrap(err, "invalid stage request")
	}

	paths, signatures, receiver, err := s.endpoint.Stage(request.Paths, request.Digests)
	if err != nil {
		s.encoder.Encode(&StageResponse{Error: err.Error()})
		return errors.Wrap(err, "unable to begin staging")
	}

	response := &StageResponse{
		Paths:      paths,
		Signatures: signatures,
	}
	if err = s.encoder.Encode(response); err != nil {
		return errors.Wrap(err, "unable to send stage response")
	}

	if len(paths) == 0 {
		return nil
	}

	decoder := newProtobufRsyncDecoder(s.decoder)
	if err = rsync.DecodeToReceiver(decoder, uint64(len(paths)), receiver); err != nil {
		return errors.Wrap(err, "unable to decode and forward rsync operations")
	}

	return nil
}

func (s *endpointServer) serveSupply(request *SupplyRequest) error {

	if err := request.ensureValid(); err != nil {
		return errors.Wrap(err, "invalid supply request")
	}

	encoder := newProtobufRsyncEncoder(s.encoder)
	receiver := rsync.NewEncodingReceiver(encoder)

	if err := s.endpoint.Supply(request.Paths, request.Signatures, receiver); err != nil {
		return errors.Wrap(err, "unable to perform supplying")
	}

	return nil
}

func (s *endpointServer) serveTransition(request *TransitionRequest) error {

	if err := request.ensureValid(); err != nil {
		return errors.Wrap(err, "invalid transition request")
	}

	results, problems, err := s.endpoint.Transition(request.Transitions)
	if err != nil {
		s.encoder.Encode(&TransitionResponse{Error: err.Error()})
		return errors.Wrap(err, "unable to perform transition")
	}

	wrappedResults := make([]*sync.Archive, len(results))
	for r, result := range results {
		wrappedResults[r] = &sync.Archive{Root: result}
	}

	response := &TransitionResponse{Results: wrappedResults, Problems: problems}
	if err = s.encoder.Encode(response); err != nil {
		return errors.Wrap(err, "unable to send transition response")
	}

	return nil
}
