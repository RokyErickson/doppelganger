package sync

import (
	"github.com/pkg/errors"
)

func (m SynchronizationMode) IsDefault() bool {
	return m == SynchronizationMode_SynchronizationModeDefault
}

func (m *SynchronizationMode) UnmarshalText(textBytes []byte) error {
	text := string(textBytes)

	switch text {
	case "two-way-safe":
		*m = SynchronizationMode_SynchronizationModeTwoWaySafe
	case "two-way-resolved":
		*m = SynchronizationMode_SynchronizationModeTwoWayResolved
	case "one-way-safe":
		*m = SynchronizationMode_SynchronizationModeOneWaySafe
	case "one-way-replica":
		*m = SynchronizationMode_SynchronizationModeOneWayReplica
	default:
		return errors.Errorf("unknown synchronization mode specification: %s", text)
	}

	return nil
}

func (m SynchronizationMode) Supported() bool {
	switch m {
	case SynchronizationMode_SynchronizationModeTwoWaySafe:
		return true
	case SynchronizationMode_SynchronizationModeTwoWayResolved:
		return true
	case SynchronizationMode_SynchronizationModeOneWaySafe:
		return true
	case SynchronizationMode_SynchronizationModeOneWayReplica:
		return true
	default:
		return false
	}
}

func (m SynchronizationMode) Description() string {
	switch m {
	case SynchronizationMode_SynchronizationModeDefault:
		return "Default"
	case SynchronizationMode_SynchronizationModeTwoWaySafe:
		return "Two Way Safe"
	case SynchronizationMode_SynchronizationModeTwoWayResolved:
		return "Two Way Resolved"
	case SynchronizationMode_SynchronizationModeOneWaySafe:
		return "One Way Safe"
	case SynchronizationMode_SynchronizationModeOneWayReplica:
		return "One Way Replica"
	default:
		return "Unknown"
	}
}
