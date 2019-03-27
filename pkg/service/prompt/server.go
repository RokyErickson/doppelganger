package prompt

import (
	"context"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/prompt"
)

type Server struct{}

func New() *Server {
	return &Server{}
}

type asyncPromptResponse struct {
	response string
	error    error
}

func (s *Server) Prompt(ctx context.Context, request *PromptRequest) (*PromptResponse, error) {
	if err := request.ensureValid(); err != nil {
		return nil, errors.Wrap(err, "invalid prompt request")
	}

	asyncResponse := make(chan asyncPromptResponse, 1)
	go func() {
		response, err := prompt.Prompt(request.Prompter, request.Prompt)
		asyncResponse <- asyncPromptResponse{response, err}
	}()
	select {
	case <-ctx.Done():
		return nil, errors.New("prompting cancelled while waiting for response")
	case r := <-asyncResponse:
		if r.error != nil {
			return nil, errors.Wrap(r.error, "unable to prompt")
		} else {
			return &PromptResponse{Response: r.response}, nil
		}
	}
}
