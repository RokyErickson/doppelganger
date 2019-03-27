package agent

import (
	"fmt"
	"os"
	"runtime"

	"github.com/pkg/errors"

	"github.com/google/uuid"

	"github.com/RokyErickson/doppelganger/pkg/prompt"
)

func Install() error {

	destination, err := installPath()
	if err != nil {
		return errors.Wrap(err, "unable to compute agent destination")
	}

	executablePath, err := os.Executable()
	if err != nil {
		return errors.Wrap(err, "unable to determine executable path")
	}

	if err = os.Rename(executablePath, destination); err != nil {
		return errors.Wrap(err, "unable to relocate agent executable")
	}

	return nil
}

func install(transport Transport, prompter string) error {

	goos, goarch, posix, err := probe(transport, prompter)
	if err != nil {
		return errors.Wrap(err, "unable to probe remote platform")
	}

	if err := prompt.Message(prompter, "Extracting agent..."); err != nil {
		return errors.Wrap(err, "unable to message prompter")
	}
	agentExecutable, err := executableForPlatform(goos, goarch)
	if err != nil {
		return errors.Wrap(err, "unable to get agent for platform")
	}
	defer os.Remove(agentExecutable)

	if err := prompt.Message(prompter, "Copying agent..."); err != nil {
		return errors.Wrap(err, "unable to message prompter")
	}
	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return errors.Wrap(err, "unable to generate UUID for agent copying")
	}
	destination := agentBaseName + randomUUID.String()
	if goos == "windows" {
		destination += ".exe"
	}
	if posix {
		destination = "." + destination
	}
	if err = transport.Copy(agentExecutable, destination); err != nil {
		return errors.Wrap(err, "unable to copy agent binary")
	}

	if runtime.GOOS == "windows" && posix {
		if err := prompt.Message(prompter, "Setting agent executability..."); err != nil {
			return errors.Wrap(err, "unable to message prompter")
		}
		executabilityCommand := fmt.Sprintf("chmod +x %s", destination)
		run(transport, executabilityCommand)
	}

	if err := prompt.Message(prompter, "Installing agent..."); err != nil {
		return errors.Wrap(err, "unable to message prompter")
	}
	var installCommand string
	if posix {
		installCommand = fmt.Sprintf("./%s %s", destination, ModeInstall)
	} else {
		installCommand = fmt.Sprintf("%s %s", destination, ModeInstall)
	}
	run(transport, installCommand)
	return nil
}
