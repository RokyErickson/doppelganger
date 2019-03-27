// +build !windows

package filesystem

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

const (
	executabilityProbeFileNamePrefix = TemporaryNamePrefix + "executability-test"
)

func PreservesExecutabilityByPath(path string) (bool, error) {

	file, err := ioutil.TempFile(path, executabilityProbeFileNamePrefix)
	if err != nil {
		return false, errors.Wrap(err, "unable to create test file")
	}

	defer func() {
		file.Close()
		os.Remove(file.Name())
	}()

	if err = file.Chmod(0700); err != nil {
		return false, errors.Wrap(err, "unable to mark test file as executable")
	}

	if info, err := file.Stat(); err != nil {
		return false, errors.Wrap(err, "unable to check test file executability")
	} else {
		return info.Mode()&0111 == 0100, nil
	}
}

func PreservesExecutability(directory *Directory) (bool, error) {

	name, file, err := directory.CreateTemporaryFile(executabilityProbeFileNamePrefix)
	if err != nil {
		return false, errors.Wrap(err, "unable to create test file")
	}

	defer func() {
		file.Close()
		directory.RemoveFile(name)
	}()

	osFile, ok := file.(*os.File)
	if !ok {
		panic("opened file is not an os.File object")
	}

	if err = osFile.Chmod(0700); err != nil {
		return false, errors.Wrap(err, "unable to mark test file as executable")
	}

	if info, err := osFile.Stat(); err != nil {
		return false, errors.Wrap(err, "unable to check test file executability")
	} else {
		return info.Mode()&0111 == 0100, nil
	}
}
