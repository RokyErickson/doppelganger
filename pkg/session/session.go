package session

import (
	"crypto/sha1"
	"hash"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/filesystem"
	"github.com/RokyErickson/doppelganger/pkg/sync"
)

func (v Version) Supported() bool {
	switch v {
	case Version_Version1:
		return true
	default:
		return false
	}
}

func (v Version) Hasher() hash.Hash {
	switch v {
	case Version_Version1:
		return sha1.New()
	default:
		panic("unknown or unsupported session version")
	}
}

func (v Version) DefaultSynchronizationMode() sync.SynchronizationMode {
	switch v {
	case Version_Version1:
		return sync.SynchronizationMode_SynchronizationModeTwoWaySafe
	default:
		panic("unknown or unsupported session version")
	}
}

func (v Version) DefaultSymlinkMode() sync.SymlinkMode {
	switch v {
	case Version_Version1:
		return sync.SymlinkMode_SymlinkPortable
	default:
		panic("unknown or unsupported session version")
	}
}

func (v Version) DefaultWatchMode() filesystem.WatchMode {
	switch v {
	case Version_Version1:
		return filesystem.WatchMode_WatchModePortable
	default:
		panic("unknown or unsupported session version")
	}
}

func (v Version) DefaultWatchPollingInterval() uint32 {
	switch v {
	case Version_Version1:
		return 10
	default:
		panic("unknown or unsupported session version")
	}
}

func (v Version) DefaultIgnoreVCSMode() sync.IgnoreVCSMode {
	switch v {
	case Version_Version1:
		return sync.IgnoreVCSMode_PropagateVCS
	default:
		panic("unknown or unsupported session version")
	}
}

func (v Version) DefaultFileMode() filesystem.Mode {
	switch v {
	case Version_Version1:
		return filesystem.ModePermissionUserRead |
			filesystem.ModePermissionUserWrite
	default:
		panic("unknown or unsupported session version")
	}
}

func (v Version) DefaultDirectoryMode() filesystem.Mode {
	switch v {
	case Version_Version1:
		return filesystem.ModePermissionUserRead |
			filesystem.ModePermissionUserWrite |
			filesystem.ModePermissionUserExecute
	default:
		panic("unknown or unsupported session version")
	}
}

func (v Version) DefaultOwnerSpecification() string {
	switch v {
	case Version_Version1:
		return ""
	default:
		panic("unknown or unsupported session version")
	}
}

func (v Version) DefaultGroupSpecification() string {
	switch v {
	case Version_Version1:
		return ""
	default:
		panic("unknown or unsupported session version")
	}
}

func (s *Session) EnsureValid() error {
	if s == nil {
		return errors.New("nil session")
	}

	if s.Identifier == "" {
		return errors.New("invalid session identifier")
	}

	if !s.Version.Supported() {
		return errors.New("unknown or unsupported session version")
	}

	if s.CreationTime == nil {
		return errors.New("missing creation time")
	}

	if err := s.Alpha.EnsureValid(); err != nil {
		return errors.Wrap(err, "invalid alpha URL")
	}

	if err := s.Beta.EnsureValid(); err != nil {
		return errors.Wrap(err, "invalid beta URL")
	}

	if err := s.Configuration.EnsureValid(ConfigurationSourceTypeSession); err != nil {
		return errors.Wrap(err, "invalid configuration")
	}

	if err := s.ConfigurationAlpha.EnsureValid(ConfigurationSourceTypeSessionEndpointSpecific); err != nil {
		return errors.Wrap(err, "invalid alpha-specific configuration")
	}

	if err := s.ConfigurationBeta.EnsureValid(ConfigurationSourceTypeSessionEndpointSpecific); err != nil {
		return errors.Wrap(err, "invalid beta-specific configuration")
	}

	return nil
}
