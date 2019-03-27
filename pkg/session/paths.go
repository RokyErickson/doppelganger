package session

import (
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/filesystem"
)

const (
	sessionsDirectoryName = "sessions"
	archivesDirectoryName = "archives"
)

func pathForSession(sessionIdentifier string) (string, error) {
	sessionsDirectoryPath, err := filesystem.Doppelganger(true, sessionsDirectoryName)
	if err != nil {
		return "", errors.Wrap(err, "unable to compute/create sessions directory")
	}

	return filepath.Join(sessionsDirectoryPath, sessionIdentifier), nil
}

func pathForArchive(session string) (string, error) {
	archivesDirectoryPath, err := filesystem.Doppelganger(true, archivesDirectoryName)
	if err != nil {
		return "", errors.Wrap(err, "unable to compute/create archives directory")
	}

	return filepath.Join(archivesDirectoryPath, session), nil
}
