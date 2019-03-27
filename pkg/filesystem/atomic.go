package filesystem

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	atomicWriteTemporaryNamePrefix = TemporaryNamePrefix + "atomic-write"
)

func WriteFileAtomic(path string, data []byte, permissions os.FileMode) error {

	temporary, err := ioutil.TempFile(filepath.Dir(path), atomicWriteTemporaryNamePrefix)
	if err != nil {
		return errors.Wrap(err, "unable to create temporary file")
	}

	if _, err = temporary.Write(data); err != nil {
		temporary.Close()
		os.Remove(temporary.Name())
		return errors.Wrap(err, "unable to write data to temporary file")
	}

	if err = temporary.Close(); err != nil {
		os.Remove(temporary.Name())
		return errors.Wrap(err, "unable to close temporary file")
	}

	if err = os.Chmod(temporary.Name(), permissions); err != nil {
		os.Remove(temporary.Name())
		return errors.Wrap(err, "unable to change file permissions")
	}

	if err = os.Rename(temporary.Name(), path); err != nil {
		os.Remove(temporary.Name())
		return errors.Wrap(err, "unable to rename file")
	}

	return nil
}
