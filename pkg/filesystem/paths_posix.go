// +build !windows

package filesystem

import (
	"os"

	"github.com/pkg/errors"
)

func mustComputeHomeDirectory() string {

	home, ok := os.LookupEnv("HOME")
	if !ok {
		panic(errors.New("HOME environment variable not present"))
	} else if home == "" {
		panic(errors.New("HOME environment variable empty"))
	}

	return home
}
