package agent

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/process"
)

const (
	BundleName = "doppelganger-agents.tar.gz"
)

func executableForPlatform(goos, goarch string) (string, error) {

	executablePath, err := os.Executable()
	if err != nil {
		return "", errors.Wrap(err, "unable to determine executable path")
	}

	bundlePath := filepath.Join(filepath.Dir(executablePath), BundleName)

	bundle, err := os.Open(bundlePath)
	if err != nil {
		return "", errors.Wrap(err, "unable to open agent bundle")
	}
	defer bundle.Close()

	bundleDecompressor, err := gzip.NewReader(bundle)
	if err != nil {
		return "", errors.Wrap(err, "unable to decompress agent bundle")
	}
	defer bundleDecompressor.Close()

	bundleArchive := tar.NewReader(bundleDecompressor)

	var header *tar.Header
	for {
		if h, err := bundleArchive.Next(); err != nil {
			if err == io.EOF {
				break
			}
			return "", errors.Wrap(err, "unable to read archive header")
		} else if h.Name == fmt.Sprintf("%s_%s", goos, goarch) {
			header = h
			break
		}
	}

	if header == nil {
		return "", errors.New("unsupported platform")
	}

	targetBaseName := process.ExecutableName(agentBaseName, goos)

	file, err := ioutil.TempFile("", targetBaseName)
	if err != nil {
		return "", errors.Wrap(err, "unable to create temporary file")
	}

	if _, err := io.CopyN(file, bundleArchive, header.Size); err != nil {
		file.Close()
		os.Remove(file.Name())
		return "", errors.Wrap(err, "unable to copy agent data")
	}

	if runtime.GOOS != "windows" && goos != "windows" {
		if err := file.Chmod(0700); err != nil {
			file.Close()
			os.Remove(file.Name())
			return "", errors.Wrap(err, "unable to make agent executable")
		}
	}

	if err := file.Close(); err != nil {
		os.Remove(file.Name())
		return "", errors.Wrap(err, "unable to close temporary file")
	}

	return file.Name(), nil
}
