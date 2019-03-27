package filesystem

import (
	"os"
)

type Mode os.FileMode

const (
	ModeTypeMask = Mode(os.ModeType)

	ModeTypeDirectory = Mode(os.ModeDir)

	ModeTypeFile = Mode(0)

	ModeTypeSymbolicLink = Mode(os.ModeSymlink)
)
