package agent

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/shibukawa/extstat"

	"github.com/RokyErickson/doppelganger/pkg/filesystem"
	"github.com/RokyErickson/doppelganger/pkg/process"
)

const (
	maximumAgentIdlePeriod = 30 * 24 * time.Hour
)

func Housekeep() {

	agentsDirectoryPath, err := filesystem.Doppelganger(false, agentsDirectoryName)
	if err != nil {
		return
	}

	agentDirectoryContents, err := filesystem.DirectoryContentsByPath(agentsDirectoryPath)
	if err != nil {
		return
	}

	agentName := process.ExecutableName(agentBaseName, runtime.GOOS)

	now := time.Now()

	for _, c := range agentDirectoryContents {

		agentVersion := c.Name()
		if stat, err := extstat.NewFromFileName(filepath.Join(agentsDirectoryPath, agentVersion, agentName)); err != nil {
			continue
		} else if now.Sub(stat.AccessTime) > maximumAgentIdlePeriod {
			os.RemoveAll(filepath.Join(agentsDirectoryPath, agentVersion))
		}
	}
}
