package prompt

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/havoc-io/gopass"
)

func PromptCommandLine(prompt string) (string, error) {

	class := Classify(prompt)

	var getter func() ([]byte, error)
	if class == PromptKindEcho || class == PromptKindBinary {
		getter = gopass.GetPasswdEchoed
	} else {
		getter = gopass.GetPasswd
	}

	fmt.Print(prompt)

	result, err := getter()
	if err != nil {
		return "", errors.Wrap(err, "unable to read response")
	}

	return string(result), nil
}
