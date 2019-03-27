package ssh

import (
	"os"
	"strings"

	"github.com/RokyErickson/doppelganger/pkg/prompt"
)

func setPrompterVariables(prompter string) map[string]string {

	environment := make(map[string]string)

	environment["LC_ALL"] = "C"

	if prompter == "" {

		filteredEnvironment := environment
		for _, e := range os.Environ() {
			if strings.HasPrefix(e, "SSH_ASKPASS=") {
				environment["SSH_ASKPASS"] = ""
			}
		}
		environment = filteredEnvironment
	} else {

		if doppelgangerPath, err := os.Executable(); err != nil {
			panic("can't find executable")
		} else {
			environment["SSH_ASKPASS"] = doppelgangerPath
		}
		environment["DISPLAY"] = "doppelganger"
		environment[prompt.PrompterEnvironmentVariable] = prompter
	}
	return environment
}
