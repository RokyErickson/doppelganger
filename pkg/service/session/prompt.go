package session

import (
	"github.com/pkg/errors"
)

type createStreamPrompter struct {
	stream Sessions_CreateServer
}

func (p *createStreamPrompter) sendReceive(request *CreateResponse) (*CreateRequest, error) {
	if err := p.stream.Send(request); err != nil {
		return nil, errors.Wrap(err, "unable to send request")
	}

	if response, err := p.stream.Recv(); err != nil {
		return nil, errors.Wrap(err, "unable to receive response")
	} else if err = response.ensureValid(false); err != nil {
		return nil, errors.Wrap(err, "invalid response received")
	} else {
		return response, nil
	}
}

func (p *createStreamPrompter) Message(message string) error {
	_, err := p.sendReceive(&CreateResponse{Message: message})
	return err
}

func (p *createStreamPrompter) Prompt(prompt string) (string, error) {
	if response, err := p.sendReceive(&CreateResponse{Prompt: prompt}); err != nil {
		return "", err
	} else {
		return response.Response, nil
	}
}

type flushStreamPrompter struct {
	stream Sessions_FlushServer
}

func (p *flushStreamPrompter) sendReceive(request *FlushResponse) (*FlushRequest, error) {
	if err := p.stream.Send(request); err != nil {
		return nil, errors.Wrap(err, "unable to send request")
	}

	if response, err := p.stream.Recv(); err != nil {
		return nil, errors.Wrap(err, "unable to receive response")
	} else if err = response.ensureValid(false); err != nil {
		return nil, errors.Wrap(err, "invalid response received")
	} else {
		return response, nil
	}
}

func (p *flushStreamPrompter) Message(message string) error {
	_, err := p.sendReceive(&FlushResponse{Message: message})
	return err
}

func (p *flushStreamPrompter) Prompt(_ string) (string, error) {
	return "", errors.New("prompting not supported on flush message streams")
}

type pauseStreamPrompter struct {
	stream Sessions_PauseServer
}

func (p *pauseStreamPrompter) sendReceive(request *PauseResponse) (*PauseRequest, error) {
	if err := p.stream.Send(request); err != nil {
		return nil, errors.Wrap(err, "unable to send request")
	}

	if response, err := p.stream.Recv(); err != nil {
		return nil, errors.Wrap(err, "unable to receive response")
	} else if err = response.ensureValid(false); err != nil {
		return nil, errors.Wrap(err, "invalid response received")
	} else {
		return response, nil
	}
}

func (p *pauseStreamPrompter) Message(message string) error {
	_, err := p.sendReceive(&PauseResponse{Message: message})
	return err
}

func (p *pauseStreamPrompter) Prompt(_ string) (string, error) {
	return "", errors.New("prompting not supported on pause message streams")
}

type resumeStreamPrompter struct {
	stream Sessions_ResumeServer
}

func (p *resumeStreamPrompter) sendReceive(request *ResumeResponse) (*ResumeRequest, error) {
	if err := p.stream.Send(request); err != nil {
		return nil, errors.Wrap(err, "unable to send request")
	}

	if response, err := p.stream.Recv(); err != nil {
		return nil, errors.Wrap(err, "unable to receive response")
	} else if err = response.ensureValid(false); err != nil {
		return nil, errors.Wrap(err, "invalid response received")
	} else {
		return response, nil
	}
}

func (p *resumeStreamPrompter) Message(message string) error {
	_, err := p.sendReceive(&ResumeResponse{Message: message})
	return err
}

func (p *resumeStreamPrompter) Prompt(prompt string) (string, error) {
	if response, err := p.sendReceive(&ResumeResponse{Prompt: prompt}); err != nil {
		return "", err
	} else {
		return response.Response, nil
	}
}

type terminateStreamPrompter struct {
	stream Sessions_TerminateServer
}

func (p *terminateStreamPrompter) sendReceive(request *TerminateResponse) (*TerminateRequest, error) {
	if err := p.stream.Send(request); err != nil {
		return nil, errors.Wrap(err, "unable to send request")
	}

	if response, err := p.stream.Recv(); err != nil {
		return nil, errors.Wrap(err, "unable to receive response")
	} else if err = response.ensureValid(false); err != nil {
		return nil, errors.Wrap(err, "invalid response received")
	} else {
		return response, nil
	}
}

func (p *terminateStreamPrompter) Message(message string) error {
	_, err := p.sendReceive(&TerminateResponse{Message: message})
	return err
}

func (p *terminateStreamPrompter) Prompt(_ string) (string, error) {
	return "", errors.New("prompting not supported on terminate message streams")
}
