package session

import (
	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/configuration"
	"github.com/RokyErickson/doppelganger/pkg/filesystem"
	"github.com/RokyErickson/doppelganger/pkg/sync"
)

type ConfigurationSourceType uint8

const (
	ConfigurationSourceTypeSession ConfigurationSourceType = iota
	ConfigurationSourceTypeGlobal
	ConfigurationSourceTypeCreate
	ConfigurationSourceTypeSessionEndpointSpecific
	ConfigurationSourceTypeCreateEndpointSpecific
	ConfigurationSourceTypeAPIEndpointSpecific
)

func (c *Configuration) EnsureValid(source ConfigurationSourceType) error {

	if c == nil {
		return errors.New("nil configuration")
	}

	endpointSpecific := source == ConfigurationSourceTypeSessionEndpointSpecific ||
		source == ConfigurationSourceTypeCreateEndpointSpecific ||
		source == ConfigurationSourceTypeAPIEndpointSpecific

	if endpointSpecific {
		if !c.SynchronizationMode.IsDefault() {
			return errors.New("synchronization mode cannot be specified on an endpoint-specific basis")
		}
	} else {
		if !(c.SynchronizationMode.IsDefault() || c.SynchronizationMode.Supported()) {
			return errors.New("unknown or unsupported synchronization mode")
		}
	}

	if endpointSpecific {
		if !c.SymlinkMode.IsDefault() {
			return errors.New("symbolic link handling mode cannot be specified on an endpoint-specific basis")
		}
	} else {
		if !(c.SymlinkMode.IsDefault() || c.SymlinkMode.Supported()) {
			return errors.New("unknown or unsupported symlink mode")
		}
	}

	if !(c.WatchMode.IsDefault() || c.WatchMode.Supported()) {
		return errors.New("unknown or unsupported watch mode")
	}

	if source != ConfigurationSourceTypeSession && len(c.DefaultIgnores) > 0 {
		return errors.New("deprecated default ignores configuration field specified")
	}
	for _, ignore := range c.DefaultIgnores {
		if !sync.ValidIgnorePattern(ignore) {
			return errors.Errorf("invalid default ignore pattern: %s", ignore)
		}
	}

	if endpointSpecific && len(c.Ignores) > 0 {
		return errors.New("ignores cannot be specified on an endpoint-specific basis")
	}
	for _, ignore := range c.Ignores {
		if !sync.ValidIgnorePattern(ignore) {
			return errors.Errorf("invalid ignore pattern: %s", ignore)
		}
	}

	if endpointSpecific {
		if !c.IgnoreVCSMode.IsDefault() {
			return errors.New("VCS ignore mode cannot be specified on an endpoint-specific basis")
		}
	} else {
		if !(c.IgnoreVCSMode.IsDefault() || c.IgnoreVCSMode.Supported()) {
			return errors.New("unknown or unsupported VCS ignore mode")
		}
	}

	if c.DefaultFileMode != 0 {
		if err := sync.EnsureDefaultFileModeValid(filesystem.Mode(c.DefaultFileMode)); err != nil {
			return errors.Wrap(err, "invalid default file permission mode specified")
		}
	}

	if c.DefaultDirectoryMode != 0 {
		if err := sync.EnsureDefaultDirectoryModeValid(filesystem.Mode(c.DefaultDirectoryMode)); err != nil {
			return errors.Wrap(err, "invalid default directory permission mode specified")
		}
	}

	if c.DefaultOwner != "" {
		if kind, _ := filesystem.ParseOwnershipIdentifier(c.DefaultOwner); kind == filesystem.OwnershipIdentifierKindInvalid {
			return errors.New("invalid default owner specification")
		}
	}

	if c.DefaultGroup != "" {
		if kind, _ := filesystem.ParseOwnershipIdentifier(c.DefaultGroup); kind == filesystem.OwnershipIdentifierKindInvalid {
			return errors.New("invalid default group specification")
		}
	}

	return nil
}

func snapshotGlobalConfiguration() (*Configuration, error) {
	configuration, err := configuration.Load()
	if err != nil {
		return nil, errors.Wrap(err, "unable to load global configuration")
	}

	result := &Configuration{
		SynchronizationMode:    configuration.Synchronization.Mode,
		MaximumEntryCount:      configuration.Synchronization.MaximumEntryCount,
		MaximumStagingFileSize: uint64(configuration.Synchronization.MaximumStagingFileSize),
		SymlinkMode:            configuration.Symlink.Mode,
		WatchMode:              configuration.Watch.Mode,
		WatchPollingInterval:   configuration.Watch.PollingInterval,
		Ignores:                configuration.Ignore.Default,
		IgnoreVCSMode:          configuration.Ignore.VCS,
		DefaultFileMode:        uint32(configuration.Permissions.DefaultFileMode),
		DefaultDirectoryMode:   uint32(configuration.Permissions.DefaultDirectoryMode),
		DefaultOwner:           configuration.Permissions.DefaultOwner,
		DefaultGroup:           configuration.Permissions.DefaultGroup,
	}

	if err := result.EnsureValid(ConfigurationSourceTypeGlobal); err != nil {
		return nil, errors.Wrap(err, "global configuration invalid")
	}

	return result, nil
}

func MergeConfigurations(lower, higher *Configuration) *Configuration {
	result := &Configuration{}

	if !higher.SynchronizationMode.IsDefault() {
		result.SynchronizationMode = higher.SynchronizationMode
	} else {
		result.SynchronizationMode = lower.SynchronizationMode
	}

	if higher.MaximumEntryCount != 0 {
		result.MaximumEntryCount = higher.MaximumEntryCount
	} else {
		result.MaximumEntryCount = lower.MaximumEntryCount
	}

	if higher.MaximumStagingFileSize != 0 {
		result.MaximumStagingFileSize = higher.MaximumStagingFileSize
	} else {
		result.MaximumStagingFileSize = lower.MaximumStagingFileSize
	}

	if !higher.SymlinkMode.IsDefault() {
		result.SymlinkMode = higher.SymlinkMode
	} else {
		result.SymlinkMode = lower.SymlinkMode
	}

	if !higher.WatchMode.IsDefault() {
		result.WatchMode = higher.WatchMode
	} else {
		result.WatchMode = lower.WatchMode
	}

	if higher.WatchPollingInterval != 0 {
		result.WatchPollingInterval = higher.WatchPollingInterval
	} else {
		result.WatchPollingInterval = lower.WatchPollingInterval
	}

	result.DefaultIgnores = append(result.DefaultIgnores, lower.DefaultIgnores...)
	result.DefaultIgnores = append(result.DefaultIgnores, higher.DefaultIgnores...)

	result.Ignores = append(result.Ignores, lower.Ignores...)
	result.Ignores = append(result.Ignores, higher.Ignores...)

	if !higher.IgnoreVCSMode.IsDefault() {
		result.IgnoreVCSMode = higher.IgnoreVCSMode
	} else {
		result.IgnoreVCSMode = lower.IgnoreVCSMode
	}

	if higher.DefaultFileMode != 0 {
		result.DefaultFileMode = higher.DefaultFileMode
	} else {
		result.DefaultFileMode = lower.DefaultFileMode
	}

	if higher.DefaultDirectoryMode != 0 {
		result.DefaultDirectoryMode = higher.DefaultDirectoryMode
	} else {
		result.DefaultDirectoryMode = lower.DefaultDirectoryMode
	}

	if higher.DefaultOwner != "" {
		result.DefaultOwner = higher.DefaultOwner
	} else {
		result.DefaultOwner = lower.DefaultOwner
	}

	if higher.DefaultGroup != "" {
		result.DefaultGroup = higher.DefaultGroup
	} else {
		result.DefaultGroup = lower.DefaultGroup
	}

	return result
}
