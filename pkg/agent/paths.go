package agent

import (
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/doppelganger"
	"github.com/RokyErickson/doppelganger/pkg/filesystem"
	"github.com/RokyErickson/doppelganger/pkg/process"
)

const (
	agentsDirectoryName = "agents"
	agentBaseName       = "doppelganger-agent"
)

func installPath() (string, error) {

	parent, err := filesystem.Doppelganger(true, agentsDirectoryName, doppelganger.Version)
	if err != nil {
		return "", errors.Wrap(err, "unable to compute parent directory")
	}

	executableName := process.ExecutableName(agentBaseName, runtime.GOOS)

	return filepath.Join(parent, executableName), nil
}
