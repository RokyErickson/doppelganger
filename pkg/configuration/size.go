package configuration

import (
	"github.com/dustin/go-humanize"
)

type ByteSize uint64

func (s *ByteSize) UnmarshalText(textBytes []byte) error {

	text := string(textBytes)

	value, err := humanize.ParseBytes(text)
	if err != nil {
		return err
	}
	*s = ByteSize(value)

	return nil
}
