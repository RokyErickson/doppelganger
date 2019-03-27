package session

import (
	"github.com/pkg/errors"
)

func (s Status) Description() string {
	switch s {
	case Status_Disconnected:
		return "Waiting to connect"
	case Status_HaltedOnRootDeletion:
		return "Halted due to root deletion"
	case Status_HaltedOnRootTypeChange:
		return "Halted due to root type change"
	case Status_ConnectingAlpha:
		return "Connecting to alpha"
	case Status_ConnectingBeta:
		return "Connecting to beta"
	case Status_Watching:
		return "Watching for changes"
	case Status_Scanning:
		return "Scanning files"
	case Status_WaitingForRescan:
		return "Waiting 5 seconds for rescan"
	case Status_Reconciling:
		return "Reconciling changes"
	case Status_StagingAlpha:
		return "Staging files on alpha"
	case Status_StagingBeta:
		return "Staging files on beta"
	case Status_Transitioning:
		return "Applying changes"
	case Status_Saving:
		return "Saving archive"
	default:
		return "Unknown"
	}
}

func (s *State) EnsureValid() error {
	if s == nil {
		return errors.New("nil state")
	}

	if err := s.Session.EnsureValid(); err != nil {
		return errors.Wrap(err, "invalid session")
	}

	if err := s.StagingStatus.EnsureValid(); err != nil {
		return errors.Wrap(err, "invalid staging status")
	}

	for _, c := range s.Conflicts {
		if err := c.EnsureValid(); err != nil {
			return errors.Wrap(err, "invalid conflict detected")
		}
	}

	for _, c := range s.AlphaProblems {
		if err := c.EnsureValid(); err != nil {
			return errors.Wrap(err, "invalid alpha problem detected")
		}
	}

	for _, c := range s.BetaProblems {
		if err := c.EnsureValid(); err != nil {
			return errors.Wrap(err, "invalid beta problem detected")
		}
	}

	return nil
}

func (s *State) Copy() *State {
	result := &State{}
	*result = *s

	if s.Session != nil {
		result.Session = &Session{}
		*result.Session = *s.Session
	}

	return result
}
