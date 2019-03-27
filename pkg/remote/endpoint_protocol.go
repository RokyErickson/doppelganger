package remote

import (
	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/session"
)

func (r *InitializeRequest) ensureValid() error {

	if r == nil {
		return errors.New("nil initialize request")
	}

	if r.Session == "" {
		return errors.New("empty session identifier")
	}

	if !r.Version.Supported() {
		return errors.New("unsupported session version")
	}

	if r.Root == "" {
		return errors.New("empty root path")
	}

	if err := r.Configuration.EnsureValid(session.ConfigurationSourceTypeSession); err != nil {
		return errors.Wrap(err, "invalid configuration")
	}

	return nil
}

func (r *InitializeResponse) ensureValid() error {

	if r == nil {
		return errors.New("nil initialize response")
	}

	return nil
}

func (r *PollRequest) ensureValid() error {

	if r == nil {
		return errors.New("nil poll request")
	}

	return nil
}

func (r *PollCompletionRequest) ensureValid() error {

	if r == nil {
		return errors.New("nil poll completion request")
	}

	return nil
}

func (r *PollResponse) ensureValid() error {

	if r == nil {
		return errors.New("nil poll response")
	}

	return nil
}

func (r *ScanRequest) ensureValid() error {

	if r == nil {
		return errors.New("nil scan request")
	}

	if err := r.BaseSnapshotSignature.EnsureValid(); err != nil {
		return errors.Wrap(err, "invalid base snapshot signature")
	}

	return nil
}

func (r *ScanResponse) ensureValid() error {

	if r == nil {
		return errors.New("nil scan response")
	}

	for _, operation := range r.SnapshotDelta {
		if err := operation.EnsureValid(); err != nil {
			return errors.Wrap(err, "invalid snapshot delta operation")
		}
	}

	if r.Error != "" {
		if len(r.SnapshotDelta) > 0 {
			return errors.New("non-empty snapshot delta present on error")
		} else if r.PreservesExecutability {
			return errors.New("executability preservation information present on error")
		}
	}

	return nil
}

func (r *StageRequest) ensureValid() error {

	if r == nil {
		return errors.New("nil stage request")
	}

	if len(r.Paths) == 0 {
		return errors.New("no paths present")
	}

	if len(r.Digests) != len(r.Paths) {
		return errors.New("digest count does not match path count")
	}

	return nil
}

func (r *StageResponse) ensureValid() error {

	if r == nil {
		return errors.New("nil stage response")
	}

	if len(r.Paths) != len(r.Signatures) {
		return errors.New("number of paths not equal to number of signatures")
	}

	for _, signature := range r.Signatures {
		if err := signature.EnsureValid(); err != nil {
			return errors.Wrap(err, "invalid rsync signature")
		}
	}

	if r.Error != "" {
		if len(r.Paths) > 0 {
			return errors.New("paths/signatures present on error")
		}
	}

	return nil
}

func (r *SupplyRequest) ensureValid() error {

	if r == nil {
		return errors.New("nil supply request")
	}

	if len(r.Paths) != len(r.Signatures) {
		return errors.New("number of paths does not match number of signatures")
	}

	for _, s := range r.Signatures {
		if err := s.EnsureValid(); err != nil {
			return errors.Wrap(err, "invalid base signature detected")
		}
	}

	return nil
}

func (r *TransitionRequest) ensureValid() error {

	if r == nil {
		return errors.New("nil transition request")
	}

	for _, change := range r.Transitions {
		if err := change.EnsureValid(); err != nil {
			return errors.Wrap(err, "invalid transition")
		}
	}

	return nil
}

func (r *TransitionResponse) ensureValid(expectedCount int) error {

	if r == nil {
		return errors.New("nil transition response")
	}

	if len(r.Results) != expectedCount {
		return errors.New("unexpected number of results returned")
	}

	for _, result := range r.Results {
		if err := result.EnsureValid(); err != nil {
			return errors.Wrap(err, "invalid result returned")
		}
	}

	for _, problem := range r.Problems {
		if err := problem.EnsureValid(); err != nil {
			return errors.Wrap(err, "invalid problem returned")
		}
	}

	return nil
}

func (r *EndpointRequest) ensureValid() error {

	if r == nil {
		return errors.New("nil endpoint request")
	}

	set := 0
	if r.Poll != nil {
		set++
	}
	if r.Scan != nil {
		set++
	}
	if r.Stage != nil {
		set++
	}
	if r.Supply != nil {
		set++
	}
	if r.Transition != nil {
		set++
	}
	if set != 1 {
		return errors.New("invalid number of fields set")
	}

	return nil
}
