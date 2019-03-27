package sync

import (
	"github.com/pkg/errors"
)

func (c *Conflict) CopySlim() *Conflict {

	alphaChanges := make([]*Change, len(c.AlphaChanges))
	for a, change := range c.AlphaChanges {
		alphaChanges[a] = change.copySlim()
	}

	betaChanges := make([]*Change, len(c.BetaChanges))
	for b, change := range c.BetaChanges {
		betaChanges[b] = change.copySlim()
	}

	return &Conflict{
		AlphaChanges: alphaChanges,
		BetaChanges:  betaChanges,
	}
}

func (c *Conflict) Root() string {

	if len(c.AlphaChanges) == 1 && len(c.BetaChanges) == 1 {
		if len(c.AlphaChanges[0].Path) < len(c.BetaChanges[0].Path) {
			return c.AlphaChanges[0].Path
		} else {
			return c.BetaChanges[0].Path
		}
	} else if len(c.AlphaChanges) == 1 && len(c.BetaChanges) != 1 {
		return c.AlphaChanges[0].Path
	} else if len(c.BetaChanges) == 1 && len(c.AlphaChanges) != 1 {
		return c.BetaChanges[0].Path
	} else {
		panic("invalid conflict")
	}
}

func (c *Conflict) EnsureValid() error {

	if c == nil {
		return errors.New("nil conflict")
	}

	if len(c.AlphaChanges) == 0 {
		return errors.New("conflict has no changes to alpha")
	} else {
		for _, change := range c.AlphaChanges {
			if err := change.EnsureValid(); err != nil {
				return errors.Wrap(err, "invalid alpha change detected")
			}
		}
	}
	if len(c.BetaChanges) == 0 {
		return errors.New("conflict has no changes to beta")
	} else {
		for _, change := range c.BetaChanges {
			if err := change.EnsureValid(); err != nil {
				return errors.Wrap(err, "invalid beta change detected")
			}
		}
	}

	if len(c.AlphaChanges) != 1 && len(c.BetaChanges) != 1 {
		return errors.New("both sides of conflict have zero or multiple changes")
	}

	return nil
}
