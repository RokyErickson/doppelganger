package sync

import (
	"github.com/pkg/errors"
)

func (c *Cache) EnsureValid() error {
	if c == nil {
		return errors.New("nil cache")
	}

	for _, e := range c.Entries {
		if e == nil {
			return errors.New("nil cache entry detected")
		} else if e.ModificationTime == nil {
			return errors.New("cache entry will nil modification time detected")
		}
	}

	return nil
}

type ReverseLookupMap struct {
	map20 map[[20]byte]string
}

func (m *ReverseLookupMap) Lookup(digest []byte) (string, bool) {
	if len(digest) == 20 {
		var key [20]byte
		copy(key[:], digest)

		result, ok := m.map20[key]

		return result, ok
	}

	return "", false
}

func (c *Cache) GenerateReverseLookupMap() (*ReverseLookupMap, error) {
	result := &ReverseLookupMap{}

	digestSize := -1

	for p, e := range c.Entries {
		if digestSize == -1 {
			digestSize = len(e.Digest)
			if digestSize == 20 {
				result.map20 = make(map[[20]byte]string, len(c.Entries))
			} else {
				return nil, errors.New("unsupported digest size")
			}
		} else if len(e.Digest) != digestSize {
			return nil, errors.New("inconsistent digest sizes")
		}

		if digestSize == 20 {
			var key [20]byte
			copy(key[:], e.Digest)
			result.map20[key] = p
		} else {
			panic("invalid digest size allowed")
		}
	}

	return result, nil
}
