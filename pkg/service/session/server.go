package session

import (
	"context"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/prompt"
	"github.com/RokyErickson/doppelganger/pkg/session"
)

type Server struct {
	manager *session.Manager
}

func New() (*Server, error) {
	manager, err := session.NewManager()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create session manager")
	}

	return &Server{
		manager: manager,
	}, nil
}

func (s *Server) Shutdown() {
	s.manager.Shutdown()
}

func (s *Server) Create(stream Sessions_CreateServer) error {
	request, err := stream.Recv()
	if err != nil {
		return errors.Wrap(err, "unable to receive request")
	} else if err = request.ensureValid(true); err != nil {
		return errors.Wrap(err, "received invalid create request")
	}

	prompter, err := prompt.RegisterPrompter(&createStreamPrompter{stream})
	if err != nil {
		return errors.Wrap(err, "unable to register prompter")
	}

	session, err := s.manager.Create(
		request.Alpha,
		request.Beta,
		request.Configuration,
		request.ConfigurationAlpha,
		request.ConfigurationBeta,
		prompter,
	)

	prompt.UnregisterPrompter(prompter)

	if err != nil {
		return err
	}

	if err := stream.Send(&CreateResponse{Session: session}); err != nil {
		return errors.Wrap(err, "unable to send response")
	}

	return nil
}

func (s *Server) List(_ context.Context, request *ListRequest) (*ListResponse, error) {
	if err := request.ensureValid(); err != nil {
		return nil, errors.Wrap(err, "received invalid list request")
	}

	stateIndex, states, err := s.manager.List(request.PreviousStateIndex, request.Specifications)
	if err != nil {
		return nil, err
	}

	return &ListResponse{
		StateIndex:    stateIndex,
		SessionStates: states,
	}, nil
}

func (s *Server) Flush(stream Sessions_FlushServer) error {
	request, err := stream.Recv()
	if err != nil {
		return errors.Wrap(err, "unable to receive request")
	} else if err = request.ensureValid(true); err != nil {
		return errors.Wrap(err, "received invalid flush request")
	}

	prompter, err := prompt.RegisterPrompter(&flushStreamPrompter{stream})
	if err != nil {
		return errors.Wrap(err, "unable to register prompter")
	}

	err = s.manager.Flush(request.Specifications, prompter, request.SkipWait, stream.Context())

	prompt.UnregisterPrompter(prompter)

	if err != nil {
		return err
	}

	if err := stream.Send(&FlushResponse{}); err != nil {
		return errors.Wrap(err, "unable to send response")
	}

	return nil
}

func (s *Server) Pause(stream Sessions_PauseServer) error {
	request, err := stream.Recv()
	if err != nil {
		return errors.Wrap(err, "unable to receive request")
	} else if err = request.ensureValid(true); err != nil {
		return errors.Wrap(err, "received invalid pause request")
	}

	prompter, err := prompt.RegisterPrompter(&pauseStreamPrompter{stream})
	if err != nil {
		return errors.Wrap(err, "unable to register prompter")
	}

	err = s.manager.Pause(request.Specifications, prompter)

	prompt.UnregisterPrompter(prompter)

	if err != nil {
		return err
	}

	if err := stream.Send(&PauseResponse{}); err != nil {
		return errors.Wrap(err, "unable to send response")
	}

	return nil
}

func (s *Server) Resume(stream Sessions_ResumeServer) error {
	request, err := stream.Recv()
	if err != nil {
		return errors.Wrap(err, "unable to receive request")
	} else if err = request.ensureValid(true); err != nil {
		return errors.Wrap(err, "received invalid resume request")
	}

	prompter, err := prompt.RegisterPrompter(&resumeStreamPrompter{stream})
	if err != nil {
		return errors.Wrap(err, "unable to register prompter")
	}

	err = s.manager.Resume(request.Specifications, prompter)

	prompt.UnregisterPrompter(prompter)

	if err != nil {
		return err
	}

	if err := stream.Send(&ResumeResponse{}); err != nil {
		return errors.Wrap(err, "unable to send response")
	}

	return nil
}

func (s *Server) Terminate(stream Sessions_TerminateServer) error {

	request, err := stream.Recv()
	if err != nil {
		return errors.Wrap(err, "unable to receive request")
	} else if err = request.ensureValid(true); err != nil {
		return errors.Wrap(err, "received invalid terminate request")
	}

	prompter, err := prompt.RegisterPrompter(&terminateStreamPrompter{stream})
	if err != nil {
		return errors.Wrap(err, "unable to register prompter")
	}

	err = s.manager.Terminate(request.Specifications, prompter)

	prompt.UnregisterPrompter(prompter)

	if err != nil {
		return err
	}

	if err := stream.Send(&TerminateResponse{}); err != nil {
		return errors.Wrap(err, "unable to send response")
	}

	return nil
}
