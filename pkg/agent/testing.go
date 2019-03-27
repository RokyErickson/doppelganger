package agent

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/doppelganger"
)

const (
	buildDirectoryName = "build"
)

func CopyBundleForTesting() error {

	executablePath, err := os.Executable()
	if err != nil {
		return errors.Wrap(err, "unable to compute test executable path")
	}
	testDirectory := filepath.Dir(executablePath)

	doppelgangerSourcePath, err := doppelganger.SourceTreePath()
	if err != nil {
		return errors.Wrap(err, "unable to compute Doppelganger source tree path")
	}

	agentBundlePath := filepath.Join(doppelgangerSourcePath, buildDirectoryName, BundleName)

	bundleCopyFile, err := os.Create(filepath.Join(testDirectory, BundleName))
	if err != nil {
		return errors.Wrap(err, "unable to create agent bundle copy file")
	}
	defer bundleCopyFile.Close()

	bundleFile, err := os.Open(agentBundlePath)
	if err != nil {
		return errors.Wrap(err, "unable to open agent bundle file")
	}
	defer bundleFile.Close()

	if _, err := io.Copy(bundleCopyFile, bundleFile); err != nil {
		return errors.Wrap(err, "unable to copy bundle file contents")
	}

	return nil
}
