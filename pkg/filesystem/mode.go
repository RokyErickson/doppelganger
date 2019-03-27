package filesystem

import (
	"strconv"

	"github.com/pkg/errors"
)

const (
	ModePermissionsMask = Mode(0777)

	ModePermissionUserRead = Mode(0400)

	ModePermissionUserWrite = Mode(0200)

	ModePermissionUserExecute = Mode(0100)

	ModePermissionGroupRead = Mode(0040)

	ModePermissionGroupWrite = Mode(0020)

	ModePermissionGroupExecute = Mode(0010)

	ModePermissionOthersRead = Mode(0004)

	ModePermissionOthersWrite = Mode(0002)

	ModePermissionOthersExecute = Mode(0001)
)

func ParseMode(value string, mask Mode) (Mode, error) {
	if m, err := strconv.ParseUint(value, 8, 32); err != nil {
		return 0, errors.Wrap(err, "unable to parse numeric value")
	} else if mode := Mode(m); mode&mask != mode {
		return 0, errors.New("mode contains disallowed bits")
	} else {
		return mode, nil
	}
}

func (m *Mode) UnmarshalText(textBytes []byte) error {

	text := string(textBytes)

	if result, err := ParseMode(text, ModePermissionsMask); err != nil {
		return err
	} else {
		*m = result
	}

	return nil
}
