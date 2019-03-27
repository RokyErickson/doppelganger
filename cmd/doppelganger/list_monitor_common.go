package main

import (
	"fmt"

	"github.com/dustin/go-humanize"

	"github.com/RokyErickson/doppelganger/pkg/filesystem"
	sessionpkg "github.com/RokyErickson/doppelganger/pkg/session"
	"github.com/RokyErickson/doppelganger/pkg/sync"
	urlpkg "github.com/RokyErickson/doppelganger/pkg/url"
)

func printEndpoint(name string, url *urlpkg.URL, configuration *sessionpkg.Configuration, version sessionpkg.Version) {

	fmt.Println(name, "configuration:")

	fmt.Println("\tURL:", url.Format("\n\t\t"))

	watchModeDescription := configuration.WatchMode.Description()
	if configuration.WatchMode.IsDefault() {
		watchModeDescription += fmt.Sprintf(" (%s)", version.DefaultWatchMode().Description())
	}
	fmt.Println("\tWatch mode:", watchModeDescription)

	if configuration.WatchMode != filesystem.WatchMode_WatchModeNoWatch {
		var watchPollingIntervalDescription string
		if configuration.WatchPollingInterval == 0 {
			watchPollingIntervalDescription = fmt.Sprintf("Default (%d seconds)", version.DefaultWatchPollingInterval())
		} else {
			watchPollingIntervalDescription = fmt.Sprintf("%d seconds", configuration.WatchPollingInterval)
		}
		fmt.Println("\tWatch polling interval:", watchPollingIntervalDescription)
	}

	var defaultFileModeDescription string
	if configuration.DefaultFileMode == 0 {
		defaultFileModeDescription = fmt.Sprintf("Default (%#o)", version.DefaultFileMode())
	} else {
		defaultFileModeDescription = fmt.Sprintf("%#o", configuration.DefaultFileMode)
	}
	fmt.Println("\tFile mode:", defaultFileModeDescription)

	var defaultDirectoryModeDescription string
	if configuration.DefaultDirectoryMode == 0 {
		defaultDirectoryModeDescription = fmt.Sprintf("Default (%#o)", version.DefaultDirectoryMode())
	} else {
		defaultDirectoryModeDescription = fmt.Sprintf("%#o", configuration.DefaultDirectoryMode)
	}
	fmt.Println("\tDirectory mode:", defaultDirectoryModeDescription)

	defaultOwnerDescription := "Default"
	if configuration.DefaultOwner != "" {
		defaultOwnerDescription = configuration.DefaultOwner
	}
	fmt.Println("\tDefault file/directory owner:", defaultOwnerDescription)

	defaultGroupDescription := "Default"
	if configuration.DefaultGroup != "" {
		defaultGroupDescription = configuration.DefaultGroup
	}
	fmt.Println("\tDefault file/directory group:", defaultGroupDescription)
}

func printSession(state *sessionpkg.State, long bool) {

	fmt.Println("Session:", state.Session.Identifier)

	if long {

		fmt.Println("Configuration:")

		configuration := state.Session.Configuration

		synchronizationMode := configuration.SynchronizationMode.Description()
		if configuration.SynchronizationMode.IsDefault() {
			defaultSynchronizationMode := state.Session.Version.DefaultSynchronizationMode()
			synchronizationMode += fmt.Sprintf(" (%s)", defaultSynchronizationMode.Description())
		}
		fmt.Println("\tSynchronization mode:", synchronizationMode)

		if configuration.MaximumEntryCount == 0 {
			fmt.Println("\tMaximum entry count: Unlimited")
		} else {
			fmt.Println("\tMaximum entry count:", configuration.MaximumEntryCount)
		}

		if configuration.MaximumStagingFileSize == 0 {
			fmt.Println("\tMaximum staging file size: Unlimited")
		} else {
			fmt.Printf(
				"\tMaximum staging file size: %d (%s)\n",
				configuration.MaximumStagingFileSize,
				humanize.Bytes(configuration.MaximumStagingFileSize),
			)
		}

		symlinkModeDescription := configuration.SymlinkMode.Description()
		if configuration.SymlinkMode == sync.SymlinkMode_SymlinkDefault {
			defaultSymlinkMode := state.Session.Version.DefaultSymlinkMode()
			symlinkModeDescription += fmt.Sprintf(" (%s)", defaultSymlinkMode.Description())
		}
		fmt.Println("\tSymbolic link mode:", symlinkModeDescription)

		ignoreVCSModeDescription := configuration.IgnoreVCSMode.Description()
		if configuration.IgnoreVCSMode == sync.IgnoreVCSMode_IgnoreVCSDefault {
			defaultIgnoreVCSMode := state.Session.Version.DefaultIgnoreVCSMode()
			ignoreVCSModeDescription += fmt.Sprintf(" (%s)", defaultIgnoreVCSMode.Description())
		}
		fmt.Println("\tIgnore VCS mode:", ignoreVCSModeDescription)

		if len(configuration.DefaultIgnores) > 0 {
			fmt.Println("\tDefault ignores:")
			for _, p := range configuration.DefaultIgnores {
				fmt.Printf("\t\t%s\n", p)
			}
		}

		if len(configuration.Ignores) > 0 {
			fmt.Println("\tIgnores:")
			for _, p := range configuration.Ignores {
				fmt.Printf("\t\t%s\n", p)
			}
		} else {
			fmt.Println("\tIgnores: None")
		}

		alphaConfigurationMerged := sessionpkg.MergeConfigurations(
			state.Session.Configuration,
			state.Session.ConfigurationAlpha,
		)
		printEndpoint("Alpha", state.Session.Alpha, alphaConfigurationMerged, state.Session.Version)

		betaConfigurationMerged := sessionpkg.MergeConfigurations(
			state.Session.Configuration,
			state.Session.ConfigurationBeta,
		)
		printEndpoint("Beta", state.Session.Beta, betaConfigurationMerged, state.Session.Version)
	}
}
