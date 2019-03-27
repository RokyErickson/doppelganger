package doppelganger

import (
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
)

func SourceTreePath() (string, error) {

	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("unable to compute script path")
	}

	return filepath.Dir(filepath.Dir(filepath.Dir(filePath))), nil
}
