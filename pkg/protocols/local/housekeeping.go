package local

import (
	"os"
	"path/filepath"
	"time"

	"github.com/RokyErickson/doppelganger/pkg/filesystem"
)

const (
	maximumCacheAge       = 30 * 24 * time.Hour
	maximumStagingRootAge = maximumCacheAge
)

func HousekeepCaches() {
	cachesDirectoryPath, err := filesystem.Doppelganger(false, cachesDirectoryName)
	if err != nil {
		return
	}

	cachesDirectoryContents, err := filesystem.DirectoryContentsByPath(cachesDirectoryPath)
	if err != nil {
		return
	}

	now := time.Now()

	for _, c := range cachesDirectoryContents {

		cacheName := c.Name()
		fullPath := filepath.Join(cachesDirectoryPath, cacheName)
		if stat, err := os.Stat(fullPath); err != nil {
			continue
		} else if now.Sub(stat.ModTime()) > maximumCacheAge {
			os.Remove(fullPath)
		}
	}
}

func HousekeepStaging() {
	stagingDirectoryPath, err := filesystem.Doppelganger(false, stagingDirectoryName)
	if err != nil {
		return
	}

	stagingDirectoryContents, err := filesystem.DirectoryContentsByPath(stagingDirectoryPath)
	if err != nil {
		return
	}

	now := time.Now()

	for _, c := range stagingDirectoryContents {

		stagingRootName := c.Name()
		fullPath := filepath.Join(stagingDirectoryPath, stagingRootName)
		if stat, err := os.Stat(fullPath); err != nil {
			continue
		} else if now.Sub(stat.ModTime()) > maximumStagingRootAge {
			os.RemoveAll(fullPath)
		}
	}
}
