// +build go1.11

package doppelganger

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	VersionMajor = 0

	VersionMinor = 1

	VersionPatch = 0

	VersionTag = ""
)

var Version string

func init() {

	if VersionTag != "" {
		Version = fmt.Sprintf("%d.%d.%d-%s", VersionMajor, VersionMinor, VersionPatch, VersionTag)
	} else {
		Version = fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)
	}
}

type versionBytes [12]byte

func SendVersion(writer io.Writer) error {

	var data versionBytes
	binary.BigEndian.PutUint32(data[:4], VersionMajor)
	binary.BigEndian.PutUint32(data[4:8], VersionMinor)
	binary.BigEndian.PutUint32(data[8:], VersionPatch)

	_, err := writer.Write(data[:])
	return err
}

func ReceiveVersion(reader io.Reader) (uint32, uint32, uint32, error) {

	var data versionBytes
	if _, err := io.ReadFull(reader, data[:]); err != nil {
		return 0, 0, 0, err
	}

	major := binary.BigEndian.Uint32(data[:4])
	minor := binary.BigEndian.Uint32(data[4:8])
	patch := binary.BigEndian.Uint32(data[8:])

	return major, minor, patch, nil
}

func ReceiveAndCompareVersion(reader io.Reader) (bool, error) {

	major, minor, patch, err := ReceiveVersion(reader)
	if err != nil {
		return false, err
	}

	return major == VersionMajor &&
		minor == VersionMinor &&
		patch == VersionPatch, nil
}
