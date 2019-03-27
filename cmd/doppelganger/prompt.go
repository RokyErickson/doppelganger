package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"

	promptpkg "github.com/RokyErickson/doppelganger/pkg/prompt"
	promptsvcpkg "github.com/RokyErickson/doppelganger/pkg/service/prompt"
)

func promptMain(arguments []string) error {

	if len(arguments) != 1 {
		return errors.New("invalid number of arguments")
	}
	prompt := arguments[0]

	prompter := os.Getenv(promptpkg.PrompterEnvironmentVariable)
	if prompter == "" {
		return errors.New("no prompter specified")
	}

	daemonConnection, err := createDaemonClientConnection()
	if err != nil {
		return errors.Wrap(err, "unable to connect to daemon")
	}
	defer daemonConnection.Close()

	promptService := promptsvcpkg.NewPromptingClient(daemonConnection)

	request := &promptsvcpkg.PromptRequest{
		Prompter: prompter,
		Prompt:   prompt,
	}
	response, err := promptService.Prompt(context.Background(), request)
	if err != nil {
		return errors.Wrap(err, "unable to invoke prompt")
	}

	fmt.Println(response.Response)

	return nil
}
