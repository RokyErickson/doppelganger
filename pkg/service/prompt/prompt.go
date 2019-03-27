package prompt

import (
	"github.com/pkg/errors"
)

func (r *PromptRequest) ensureValid() error {

	if r == nil {
		return errors.New("nil prompt request")
	}

	if r.Prompter == "" {
		return errors.New("empty prompter identifier")
	}

	if r.Prompt == "" {
		return errors.New("empty prompt")
	}

	return nil
}
