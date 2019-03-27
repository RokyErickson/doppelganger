package local

import (
	"crypto/sha1"
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/filesystem"
)

const (
	cachesDirectoryName = "caches"

	stagingDirectoryName = "staging"

	alphaName = "alpha"

	betaName = "beta"

	stagingPrefixLength = 1
)

func pathForCache(session string, alpha bool) (string, error) {

	cachesDirectoryPath, err := filesystem.Doppelganger(true, cachesDirectoryName)
	if err != nil {
		return "", errors.Wrap(err, "unable to compute/create caches directory")
	}

	endpointName := alphaName
	if !alpha {
		endpointName = betaName
	}

	cacheName := fmt.Sprintf("%s_%s", session, endpointName)

	return filepath.Join(cachesDirectoryPath, cacheName), nil
}

func pathForStagingRoot(session string, alpha bool) (string, error) {

	endpointName := alphaName
	if !alpha {
		endpointName = betaName
	}

	stagingRootName := fmt.Sprintf("%s_%s", session, endpointName)

	return filesystem.Doppelganger(false, stagingDirectoryName, stagingRootName)
}

func pathForStaging(root, path string, digest []byte) (string, string, error) {

	if len(digest) == 0 {
		return "", "", errors.New("entry digest too short")
	}
	prefix := fmt.Sprintf("%x", digest[:1])

	stagingName := fmt.Sprintf("%x_%x", sha1.Sum([]byte(path)), digest)

	return filepath.Join(root, prefix, stagingName), prefix, nil
}
