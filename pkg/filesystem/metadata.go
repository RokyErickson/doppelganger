package filesystem

import (
	"time"
)

type Metadata struct {
	Name string

	Mode Mode

	Size uint64

	ModificationTime time.Time

	DeviceID uint64

	FileID uint64
}
