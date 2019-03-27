package filesystem

import (
	"os/user"

	"github.com/pkg/errors"
)

func mustComputeHomeDirectory() string {

	currentUser, err := user.Current()
	if err != nil {
		panic(errors.Wrap(err, "unable to lookup current user"))
	}

	if currentUser.HomeDir == "" {
		panic(errors.New("empty home directory found"))
	}

	return currentUser.HomeDir
}
