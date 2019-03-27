package filesystem

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	doppelgangerConfigurationName = ".doppelganger.toml"

	DoppelgangerDirectoryName = ".doppelganger"

	doppelgangerDirectoryPermissions os.FileMode = 0700
)

var HomeDirectory string

var DoppelgangerConfigurationPath string

func init() {

	HomeDirectory = mustComputeHomeDirectory()

	DoppelgangerConfigurationPath = filepath.Join(HomeDirectory, doppelgangerConfigurationName)
}

func Doppelganger(create bool, subpath ...string) (string, error) {

	components := make([]string, 0, 2+len(subpath))
	components = append(components, HomeDirectory, DoppelgangerDirectoryName)
	root := filepath.Join(components...)
	components = append(components, subpath...)
	result := filepath.Join(components...)

	if create {
		if err := os.MkdirAll(result, doppelgangerDirectoryPermissions); err != nil {
			return "", errors.Wrap(err, "unable to create subpath")
		} else if err := markHidden(root); err != nil {
			return "", errors.Wrap(err, "unable to hide Doppelganger directory")
		}
	}

	return result, nil
}
