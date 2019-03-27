package remote

import (
	contextpkg "context"
	"net"

	"github.com/pkg/errors"

	"github.com/golang/protobuf/proto"

	"github.com/RokyErickson/doppelganger/pkg/compression"
	"github.com/RokyErickson/doppelganger/pkg/doppelganger"
	"github.com/RokyErickson/doppelganger/pkg/encoding"
	"github.com/RokyErickson/doppelganger/pkg/rsync"
	"github.com/RokyErickson/doppelganger/pkg/session"
	"github.com/RokyErickson/doppelganger/pkg/sync"
)

type endpointClient struct {
	connection        net.Conn
	encoder           *encoding.ProtobufEncoder
	decoder           *encoding.ProtobufDecoder
	lastSnapshotBytes []byte
}

func NewEndpointClient(
	connection net.Conn,
	root,
	session string,
	version session.Version,
	configuration *session.Configuration,
	alpha bool,
) (session.Endpoint, error) {
	if magicOk, err := receiveAndCompareMagicNumber(connection, serverMagicNumber); err != nil {
		connection.Close()
		return nil, &handshakeTransportError{errors.Wrap(err, "unable to receive server magic number")}
	} else if !magicOk {
		connection.Close()
		return nil, &handshakeTransportError{errors.New("server magic number incorrect")}
	}
	if err := sendMagicNumber(connection, clientMagicNumber); err != nil {
		connection.Close()
		return nil, &handshakeTransportError{errors.Wrap(err, "unable to send client magic number")}
	}

	serverMajor, serverMinor, serverPatch, err := doppelganger.ReceiveVersion(connection)
	if err != nil {
		connection.Close()
		return nil, &handshakeTransportError{errors.Wrap(err, "unable to receive server version")}
	}

	if err := doppelganger.SendVersion(connection); err != nil {
		connection.Close()
		return nil, &handshakeTransportError{errors.Wrap(err, "unable to send client version")}
	}

	versionMatch := serverMajor == doppelganger.VersionMajor &&
		serverMinor == doppelganger.VersionMinor &&
		serverPatch == doppelganger.VersionPatch
	if !versionMatch {
		connection.Close()
		return nil, errors.New("version mismatch")
	}

	reader := compression.NewDecompressingReader(connection)
	writer := compression.NewCompressingWriter(connection)

	encoder := encoding.NewProtobufEncoder(writer)
	decoder := encoding.NewProtobufDecoder(reader)

	request := &InitializeRequest{
		Root:          root,
		Session:       session,
		Version:       version,
		Configuration: configuration,
		Alpha:         alpha,
	}
	if err := encoder.Encode(request); err != nil {
		connection.Close()
		return nil, errors.Wrap(err, "unable to send initialize request")
	}

	response := &InitializeResponse{}
	if err := decoder.Decode(response); err != nil {
		connection.Close()
		return nil, errors.Wrap(err, "unable to receive transition response")
	} else if err = response.ensureValid(); err != nil {
		connection.Close()
		return nil, errors.Wrap(err, "invalid initialize response")
	} else if response.Error != "" {
		connection.Close()
		return nil, errors.Errorf("remote error: %s", response.Error)
	}

	return &endpointClient{
		connection: connection,
		encoder:    encoder,
		decoder:    decoder,
	}, nil
}

func (e *endpointClient) Poll(context contextpkg.Context) error {
	request := &EndpointRequest{Poll: &PollRequest{}}
	if err := e.encoder.Encode(request); err != nil {
		return errors.Wrap(err, "unable to send poll request")
	}

	completionContext, forceCompletionSend := contextpkg.WithCancel(context)
	defer forceCompletionSend()

	completionSendResults := make(chan error, 1)
	go func() {
		<-completionContext.Done()
		completionSendResults <- errors.Wrap(
			e.encoder.Encode(&PollCompletionRequest{}),
			"unable to send poll completion request",
		)
	}()

	responseReceiveResults := make(chan error, 1)
	go func() {
		response := &PollResponse{}
		if err := e.decoder.Decode(response); err != nil {
			responseReceiveResults <- errors.Wrap(err, "unable to receive poll response")
		} else if err = response.ensureValid(); err != nil {
			responseReceiveResults <- errors.Wrap(err, "invalid poll response")
		} else if response.Error != "" {
			responseReceiveResults <- errors.Errorf("remote error: %s", response.Error)
		}
		responseReceiveResults <- nil
	}()

	var completionSendErr, responseReceiveErr error
	select {
	case completionSendErr = <-completionSendResults:
		responseReceiveErr = <-responseReceiveResults
	case responseReceiveErr = <-responseReceiveResults:
		forceCompletionSend()
		completionSendErr = <-completionSendResults
	}

	if responseReceiveErr != nil {
		return responseReceiveErr
	} else if completionSendErr != nil {
		return completionSendErr
	}

	return nil
}

func (e *endpointClient) Scan(ancestor *sync.Entry) (*sync.Entry, bool, error, bool) {

	engine := rsync.NewEngine()
	var baseBytes []byte
	if e.lastSnapshotBytes != nil {
		baseBytes = e.lastSnapshotBytes
	} else {
		buffer := proto.NewBuffer(nil)
		buffer.SetDeterministic(true)
		if err := buffer.Marshal(&sync.Archive{Root: ancestor}); err != nil {
			return nil, false, errors.Wrap(err, "unable to marshal ancestor"), false
		}
		baseBytes = buffer.Bytes()
	}

	baseSignature := engine.BytesSignature(baseBytes, 0)

	request := &EndpointRequest{
		Scan: &ScanRequest{
			BaseSnapshotSignature: baseSignature,
		},
	}
	if err := e.encoder.Encode(request); err != nil {
		return nil, false, errors.Wrap(err, "unable to send scan request"), false
	}

	response := &ScanResponse{}
	if err := e.decoder.Decode(response); err != nil {
		return nil, false, errors.Wrap(err, "unable to receive scan response"), false
	} else if err = response.ensureValid(); err != nil {
		return nil, false, errors.Wrap(err, "invalid scan response"), false
	}

	if response.TryAgain {
		return nil, false, errors.New(response.Error), true
	}

	snapshotBytes, err := engine.PatchBytes(baseBytes, baseSignature, response.SnapshotDelta)
	if err != nil {
		return nil, false, errors.Wrap(err, "unable to patch base snapshot"), false
	}

	archive := &sync.Archive{}
	if err := proto.Unmarshal(snapshotBytes, archive); err != nil {
		return nil, false, errors.Wrap(err, "unable to unmarshal snapshot"), false
	}
	snapshot := archive.Root

	if err = snapshot.EnsureValid(); err != nil {
		return nil, false, errors.Wrap(err, "invalid snapshot received"), false
	}

	e.lastSnapshotBytes = snapshotBytes

	return snapshot, response.PreservesExecutability, nil, false
}

func (e *endpointClient) Stage(paths []string, digests [][]byte) ([]string, []*rsync.Signature, rsync.Receiver, error) {

	if len(paths) == 0 {
		return nil, nil, nil, nil
	}

	request := &EndpointRequest{
		Stage: &StageRequest{
			Paths:   paths,
			Digests: digests,
		},
	}
	if err := e.encoder.Encode(request); err != nil {
		return nil, nil, nil, errors.Wrap(err, "unable to send stage request")
	}

	response := &StageResponse{}
	if err := e.decoder.Decode(response); err != nil {
		return nil, nil, nil, errors.Wrap(err, "unable to receive stage response")
	} else if err = response.ensureValid(); err != nil {
		return nil, nil, nil, errors.Wrap(err, "invalid scan response")
	} else if response.Error != "" {
		return nil, nil, nil, errors.Errorf("remote error: %s", response.Error)
	}

	if len(response.Paths) == 0 {
		return nil, nil, nil, nil
	}

	encoder := newProtobufRsyncEncoder(e.encoder)
	receiver := rsync.NewEncodingReceiver(encoder)

	return response.Paths, response.Signatures, receiver, nil
}

func (e *endpointClient) Supply(paths []string, signatures []*rsync.Signature, receiver rsync.Receiver) error {

	request := &EndpointRequest{
		Supply: &SupplyRequest{
			Paths:      paths,
			Signatures: signatures,
		},
	}
	if err := e.encoder.Encode(request); err != nil {
		return errors.Wrap(err, "unable to send supply request")
	}

	decoder := newProtobufRsyncDecoder(e.decoder)
	if err := rsync.DecodeToReceiver(decoder, uint64(len(paths)), receiver); err != nil {
		return errors.Wrap(err, "unable to decode and forward rsync operations")
	}

	return nil
}

func (e *endpointClient) Transition(transitions []*sync.Change) ([]*sync.Entry, []*sync.Problem, error) {

	request := &EndpointRequest{
		Transition: &TransitionRequest{
			Transitions: transitions,
		},
	}
	if err := e.encoder.Encode(request); err != nil {
		return nil, nil, errors.Wrap(err, "unable to send transition request")
	}

	response := &TransitionResponse{}
	if err := e.decoder.Decode(response); err != nil {
		return nil, nil, errors.Wrap(err, "unable to receive transition response")
	} else if err = response.ensureValid(len(transitions)); err != nil {
		return nil, nil, errors.Wrap(err, "invalid transition response")
	} else if response.Error != "" {
		return nil, nil, errors.Errorf("remote error: %s", response.Error)
	}

	results := make([]*sync.Entry, len(response.Results))
	for r, result := range response.Results {
		results[r] = result.Root
	}

	return results, response.Problems, nil
}

func (e *endpointClient) Shutdown() error {

	return e.connection.Close()
}
