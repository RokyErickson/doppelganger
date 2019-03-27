// +build !windows,!darwin

package daemon

import (
	"github.com/pkg/errors"
)

const RegistrationSupported = false

func Register() error {
	return errors.New("daemon registration not supported on this platform")
}

func Unregister() error {
	return errors.New("daemon deregistration not supported on this platform")
}

func RegisteredStart() (bool, error) {
	return false, nil
}

func RegisteredStop() (bool, error) {
	return false, nil
}
