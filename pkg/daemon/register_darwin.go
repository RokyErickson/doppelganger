package daemon

import (
	"fmt"
	"github.com/polydawn/gosh"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/filesystem"
)

const RegistrationSupported = true

const launchdPlistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>io.doppelganger.doppelganger</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>daemon</string>
		<string>run</string>
	</array>
	<key>LimitLoadToSessionType</key>
	<string>Aqua</string>
	<key>KeepAlive</key>
	<true/>
</dict>
</plist>
`

const (
	libraryDirectoryName             = "Library"
	libraryDirectoryPermissions      = 0700
	launchAgentsDirectoryName        = "LaunchAgents"
	launchAgentsDirectoryPermissions = 0755
	launchdPlistName                 = "io.doppelganger.doppelganger.plist"
	launchdPlistPermissions          = 0644
)

func Register() error {

	if registered, err := registered(); err != nil {
		return errors.Wrap(err, "unable to determine registration status")
	} else if registered {
		return nil
	}

	lock, err := AcquireLock()
	if err != nil {
		return errors.New("unable to alter registration while daemon is running")
	}
	defer lock.Unlock()

	targetPath := filepath.Join(filesystem.HomeDirectory, libraryDirectoryName)
	if err := os.MkdirAll(targetPath, libraryDirectoryPermissions); err != nil {
		return errors.Wrap(err, "unable to create Library directory")
	}

	targetPath = filepath.Join(targetPath, launchAgentsDirectoryName)
	if err := os.MkdirAll(targetPath, launchAgentsDirectoryPermissions); err != nil {
		return errors.Wrap(err, "unable to create LaunchAgents directory")
	}

	executablePath, err := os.Executable()
	if err != nil {
		return errors.Wrap(err, "unable to determine executable path")
	}

	plist := fmt.Sprintf(launchdPlistTemplate, executablePath)

	targetPath = filepath.Join(targetPath, launchdPlistName)
	if err := filesystem.WriteFileAtomic(targetPath, []byte(plist), launchdPlistPermissions); err != nil {
		return errors.Wrap(err, "unable to write launchd agent plist")
	}

	return nil
}

func Unregister() error {

	if registered, err := registered(); err != nil {
		return errors.Wrap(err, "unable to determine registration status")
	} else if !registered {
		return nil
	}

	lock, err := AcquireLock()
	if err != nil {
		return errors.New("unable to alter registration while daemon is running")
	}
	defer lock.Unlock()

	targetPath := filepath.Join(
		filesystem.HomeDirectory,
		libraryDirectoryName,
		launchAgentsDirectoryName,
		launchdPlistName,
	)

	if err := os.Remove(targetPath); err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "unable to remove launchd agent plist")
		}
	}

	return nil
}

func registered() (bool, error) {

	targetPath := filepath.Join(
		filesystem.HomeDirectory,
		libraryDirectoryName,
		launchAgentsDirectoryName,
		launchdPlistName,
	)

	if info, err := os.Lstat(targetPath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "unable to query launchd agent plist")
	} else if !info.Mode().IsRegular() {
		return false, errors.New("unexpected contents at launchd agent plist path")
	}

	return true, nil
}

func RegisteredStart() (bool, error) {

	if registered, err := registered(); err != nil {
		return false, errors.Wrap(err, "unable to determine daemon registration status")
	} else if !registered {
		return false, nil
	}

	targetPath := filepath.Join(
		filesystem.HomeDirectory,
		libraryDirectoryName,
		launchAgentsDirectoryName,
		launchdPlistName,
	)

	load := gosh.Gosh("launchctl", "load", targetPath)
	load.Run()

	return true, nil
}

func RegisteredStop() (bool, error) {

	if registered, err := registered(); err != nil {
		return false, errors.Wrap(err, "unable to determine daemon registration status")
	} else if !registered {
		return false, nil
	}

	targetPath := filepath.Join(
		filesystem.HomeDirectory,
		libraryDirectoryName,
		launchAgentsDirectoryName,
		launchdPlistName,
	)

	unload := gosh.Gosh("launchctl", "unload", targetPath)
	unload.Run()
	return true, nil
}
