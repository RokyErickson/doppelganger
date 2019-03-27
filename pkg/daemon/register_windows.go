package daemon

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	"golang.org/x/sys/windows/registry"
)

const RegistrationSupported = true

const (
	rootKey = registry.CURRENT_USER

	runPath = "Software\\Microsoft\\Windows\\CurrentVersion\\Run"

	runKeyName = "Doppelganger"
)

func Register() error {

	key, err := registry.OpenKey(rootKey, runPath, registry.SET_VALUE)
	if err != nil {
		return errors.Wrap(err, "unable to open registry path")
	}
	defer key.Close()

	executablePath, err := os.Executable()
	if err != nil {
		return errors.Wrap(err, "unable to determine executable path")
	}

	command := fmt.Sprintf("\"%s\" daemon start", executablePath)

	if err := key.SetStringValue(runKeyName, command); err != nil {
		return errors.Wrap(err, "unable to set registry key")
	}

	return nil
}

func Unregister() error {

	key, err := registry.OpenKey(rootKey, runPath, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return errors.Wrap(err, "unable to open registry path")
	}
	defer key.Close()

	if err := key.DeleteValue(runKeyName); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "unable to remove registry key")
	}

	return nil
}

func RegisteredStart() (bool, error) {
	return false, nil
}

func RegisteredStop() (bool, error) {
	return false, nil
}
