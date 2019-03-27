package filesystem

import (
	"os"

	"github.com/pkg/errors"
)

func DirectoryContentsByPath(path string) ([]os.FileInfo, error) {

	directory, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open directory")
	}
	defer directory.Close()

	contents, err := directory.Readdir(0)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read directory contents")
	}

	return contents, nil
}
