package session

import (
	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/session"
)

func (r *CreateRequest) ensureValid(first bool) error {
	if r == nil {
		return errors.New("nil create request")
	}

	if first {
		if err := r.Alpha.EnsureValid(); err != nil {
			return errors.Wrap(err, "invalid alpha URL")
		}

		if err := r.Beta.EnsureValid(); err != nil {
			return errors.Wrap(err, "invalid beta URL")
		}

		if err := r.Configuration.EnsureValid(session.ConfigurationSourceTypeCreate); err != nil {
			return errors.Wrap(err, "invalid session configuration")
		}

		if err := r.ConfigurationAlpha.EnsureValid(session.ConfigurationSourceTypeCreateEndpointSpecific); err != nil {
			return errors.Wrap(err, "invalid alpha-specific configuration")
		}

		if err := r.ConfigurationBeta.EnsureValid(session.ConfigurationSourceTypeCreateEndpointSpecific); err != nil {
			return errors.Wrap(err, "invalid beta-specific configuration")
		}

		if r.Response != "" {
			return errors.New("non-empty response")
		}
	} else {
		if r.Alpha != nil {
			return errors.New("alpha URL present")
		}

		if r.Beta != nil {
			return errors.New("beta URL present")
		}

		if r.Configuration != nil {
			return errors.New("configuration present")
		}

		if r.ConfigurationAlpha != nil {
			return errors.New("alpha-specific configuration present")
		}

		if r.ConfigurationBeta != nil {
			return errors.New("beta-specific configuration present")
		}

	}
	return nil
}

func (r *CreateResponse) EnsureValid() error {
	if r == nil {
		return errors.New("nil create response")
	}

	var fieldsSet int
	if r.Session != "" {
		fieldsSet++
	}
	if r.Message != "" {
		fieldsSet++
	}
	if r.Prompt != "" {
		fieldsSet++
	}
	if fieldsSet != 1 {
		return errors.New("incorrect number of fields set")
	}

	return nil
}

func (r *ListRequest) ensureValid() error {
	if r == nil {
		return errors.New("nil list request")
	}

	return nil
}

func (r *ListResponse) EnsureValid() error {
	if r == nil {
		return errors.New("nil list response")
	}

	for _, s := range r.SessionStates {
		if err := s.EnsureValid(); err != nil {
			return errors.Wrap(err, "invalid session state")
		}
	}

	return nil
}

func (r *FlushRequest) ensureValid(first bool) error {

	if r == nil {
		return errors.New("nil flush request")
	}
	if first {

	} else {
		if r.Specifications != nil {
			return errors.New("non-empty specifications on message acknowledgement")
		}
	}

	return nil
}

func (r *FlushResponse) EnsureValid() error {
	if r == nil {
		return errors.New("nil flush response")
	}

	return nil
}

func (r *PauseRequest) ensureValid(first bool) error {
	if r == nil {
		return errors.New("nil pause request")
	}

	if first {

	} else {
		if r.Specifications != nil {
			return errors.New("non-empty specifications on message acknowledgement")
		}
	}

	return nil
}

func (r *PauseResponse) EnsureValid() error {

	if r == nil {
		return errors.New("nil pause response")
	}

	return nil
}

func (r *ResumeRequest) ensureValid(first bool) error {

	if r == nil {
		return errors.New("nil resume request")
	}

	if first {

	} else {

		if r.Specifications != nil {
			return errors.New("non-empty specifications on message acknowledgement")
		}
	}

	return nil
}

func (r *ResumeResponse) EnsureValid() error {

	if r == nil {
		return errors.New("nil resume response")
	}

	var fieldsSet int
	if r.Message != "" {
		fieldsSet++
	}
	if r.Prompt != "" {
		fieldsSet++
	}
	if fieldsSet > 1 {
		return errors.New("multiple fields set")
	}

	return nil
}

func (r *TerminateRequest) ensureValid(first bool) error {
	if r == nil {
		return errors.New("nil terminate request")
	}

	if first {

	} else {
		if r.Specifications != nil {
			return errors.New("non-empty specifications on message acknowledgement")
		}

	}

	return nil
}

func (r *TerminateResponse) EnsureValid() error {

	if r == nil {
		return errors.New("nil terminate response")
	}

	return nil
}
