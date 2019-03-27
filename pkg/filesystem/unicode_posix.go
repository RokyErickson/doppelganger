// +build !windows

package filesystem

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const (
	composedFileNamePrefix = TemporaryNamePrefix + "unicode-test-\xc3\xa9ntry"

	decomposedFileNamePrefix = TemporaryNamePrefix + "unicode-test-\x65\xcc\x81ntry"
)

func DecomposesUnicodeByPath(path string) (bool, error) {

	file, err := ioutil.TempFile(path, composedFileNamePrefix)
	if err != nil {
		return false, errors.Wrap(err, "unable to create test file")
	} else if err = file.Close(); err != nil {
		return false, errors.Wrap(err, "unable to close test file")
	}

	composedFilename := filepath.Base(file.Name())
	decomposedFilename := strings.Replace(
		composedFilename,
		composedFileNamePrefix,
		decomposedFileNamePrefix,
		1,
	)

	defer func() {
		if os.Remove(filepath.Join(path, composedFilename)) != nil {
			os.Remove(filepath.Join(path, decomposedFilename))
		}
	}()

	contents, err := DirectoryContentsByPath(path)
	if err != nil {
		return false, errors.Wrap(err, "unable to read directory contents")
	}

	for _, c := range contents {
		name := c.Name()
		if name == decomposedFilename {
			return true, nil
		} else if name == composedFilename {
			return false, nil
		}
	}

	return false, errors.New("unable to find test file after creation")
}

func DecomposesUnicode(directory *Directory) (bool, error) {

	composedName, file, err := directory.CreateTemporaryFile(composedFileNamePrefix)
	if err != nil {
		return false, errors.Wrap(err, "unable to create test file")
	} else if err = file.Close(); err != nil {
		return false, errors.Wrap(err, "unable to close test file")
	}

	decomposedName := strings.Replace(
		composedName,
		composedFileNamePrefix,
		decomposedFileNamePrefix,
		1,
	)

	defer func() {
		if directory.RemoveFile(composedName) != nil {
			directory.RemoveFile(decomposedName)
		}
	}()

	names, err := directory.ReadContentNames()
	if err != nil {
		return false, errors.Wrap(err, "unable to read directory content names")
	}

	for _, name := range names {
		if name == decomposedName {
			return true, nil
		} else if name == composedName {
			return false, nil
		}
	}

	return false, errors.New("unable to find test file after creation")
}
