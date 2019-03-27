package doppelganger

import (
	"bytes"
	"testing"
)

func TestVersionSendReceiveAndCompare(t *testing.T) {

	buffer := &bytes.Buffer{}

	if err := SendVersion(buffer); err != nil {
		t.Fatal("unable to send version:", err)
	}

	if buffer.Len() != 12 {
		t.Fatal("buffer does not contain expected byte count")
	}

	if match, err := ReceiveAndCompareVersion(buffer); err != nil {
		t.Fatal("unable to receive version:", err)
	} else if !match {
		t.Error("version mismatch on receive")
	}
}

func TestVersionReceiveAndCompareEmptyBuffer(t *testing.T) {

	buffer := &bytes.Buffer{}

	match, err := ReceiveAndCompareVersion(buffer)
	if err == nil {
		t.Error("version received from empty buffer")
	}
	if match {
		t.Error("version match on empty buffer")
	}
}
