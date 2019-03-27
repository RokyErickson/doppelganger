package configuration

import (
	"os"

	"github.com/RokyErickson/doppelganger/pkg/encoding"
	"github.com/RokyErickson/doppelganger/pkg/filesystem"
	"github.com/RokyErickson/doppelganger/pkg/sync"
)

type Configuration struct {
	Synchronization struct {
		Mode sync.SynchronizationMode `toml:"mode"`

		MaximumEntryCount uint64 `toml:"maxEntryCount"`

		MaximumStagingFileSize ByteSize `toml:"maxStagingFileSize"`
	} `toml:"sync"`

	Ignore struct {
		Default []string `toml:"default"`

		VCS sync.IgnoreVCSMode `toml:"vcs"`
	} `toml:"ignore"`

	Symlink struct {
		Mode sync.SymlinkMode `toml:"mode"`
	} `toml:"symlink"`

	Watch struct {
		Mode filesystem.WatchMode `toml:"mode"`

		PollingInterval uint32 `toml:"pollingInterval"`
	} `toml:"watch"`

	Permissions struct {
		DefaultFileMode filesystem.Mode `toml:"defaultFileMode"`

		DefaultDirectoryMode filesystem.Mode `toml:"defaultDirectoryMode"`
		DefaultOwner         string          `toml:"defaultOwner"`

		DefaultGroup string `toml:"defaultGroup"`
	} `toml:"permissions"`
}

func loadFromPath(path string) (*Configuration, error) {
	result := &Configuration{}

	if err := encoding.LoadAndUnmarshalTOML(path, result); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return result, nil
}

func Load() (*Configuration, error) {
	return loadFromPath(filesystem.DoppelgangerConfigurationPath)
}
