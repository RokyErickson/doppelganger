package prompt

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/google/uuid"
)

var registryLock sync.Mutex

var registry = make(map[string]chan Prompter)

func RegisterPrompter(prompter Prompter) (string, error) {

	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return "", errors.Wrap(err, "unable to generate UUID for prompter")
	}
	identifier := randomUUID.String()

	holder := make(chan Prompter, 1)
	holder <- prompter

	registryLock.Lock()
	registry[identifier] = holder
	registryLock.Unlock()

	return identifier, nil
}

func UnregisterPrompter(identifier string) {

	registryLock.Lock()
	holder, ok := registry[identifier]
	if !ok {
		panic("deregistration requested for unregistered prompter")
	}
	delete(registry, identifier)
	registryLock.Unlock()

	<-holder
	close(holder)
}

func Message(identifier, message string) error {

	if identifier == "" {
		return nil
	}

	registryLock.Lock()
	holder, ok := registry[identifier]
	registryLock.Unlock()
	if !ok {
		return errors.New("prompter not found")
	}

	prompter, ok := <-holder
	if !ok {
		return errors.New("unable to acquire prompter")
	}

	err := prompter.Message(message)

	holder <- prompter

	if err != nil {
		errors.Wrap(err, "unable to message")
	}

	return nil
}

func Prompt(identifier, prompt string) (string, error) {

	registryLock.Lock()
	holder, ok := registry[identifier]
	registryLock.Unlock()
	if !ok {
		return "", errors.New("prompter not found")
	}

	prompter, ok := <-holder
	if !ok {
		return "", errors.New("unable to acquire prompter")
	}

	response, err := prompter.Prompt(prompt)

	holder <- prompter

	if err != nil {
		return "", errors.Wrap(err, "unable to prompt")
	}

	return response, nil
}
