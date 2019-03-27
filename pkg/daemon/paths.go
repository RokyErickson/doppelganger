package daemon

import (
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/filesystem"
)

const (
	daemonDirectoryName = "daemon"
)

func subpath(name string) (string, error) {
	daemonRoot, err := filesystem.Doppelganger(true, daemonDirectoryName)
	if err != nil {
		return "", errors.Wrap(err, "unable to compute daemon directory")
	}

	return filepath.Join(daemonRoot, name), nil
}
