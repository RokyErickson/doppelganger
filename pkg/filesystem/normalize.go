package filesystem

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/pkg/errors"
)

func tildeExpand(path string) (string, error) {

	if path == "" || path[0] != '~' {
		return path, nil
	}

	pathSeparatorIndex := -1
	for i := 0; i < len(path); i++ {
		if os.IsPathSeparator(path[i]) {
			pathSeparatorIndex = i
			break
		}
	}

	var username string
	var remaining string
	if pathSeparatorIndex > 0 {
		username = path[1:pathSeparatorIndex]
		remaining = path[pathSeparatorIndex+1:]
	} else {
		username = path[1:]
	}

	var homeDirectory string
	if username == "" {
		homeDirectory = HomeDirectory
	} else {
		if u, err := user.Lookup(username); err != nil {
			return "", errors.Wrap(err, "unable to lookup user")
		} else {
			homeDirectory = u.HomeDir
		}
	}

	return filepath.Join(homeDirectory, remaining), nil
}

func Normalize(path string) (string, error) {

	path, err := tildeExpand(path)
	if err != nil {
		return "", errors.Wrap(err, "unable to perform tilde expansion")
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return "", errors.Wrap(err, "unable to compute absolute path")
	}

	return path, nil
}
