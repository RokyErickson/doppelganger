package remote

import (
	"fmt"
	"io"
)

type magicNumberBytes [3]byte

var serverMagicNumber = magicNumberBytes{0x05, 0x27, 0x87}

var clientMagicNumber = magicNumberBytes{0x87, 0x27, 0x05}

func sendMagicNumber(writer io.Writer, magicNumber magicNumberBytes) error {
	_, err := writer.Write(magicNumber[:])
	return err
}

func receiveAndCompareMagicNumber(reader io.Reader, expected magicNumberBytes) (bool, error) {

	var received magicNumberBytes
	if _, err := io.ReadFull(reader, received[:]); err != nil {
		return false, err
	}

	return received == expected, nil
}

type handshakeTransportError struct {
	underlying error
}

func (e *handshakeTransportError) Error() string {
	return fmt.Sprintf("handshake transport error: %v", e.underlying)
}

func IsHandshakeTransportError(err error) bool {
	_, ok := err.(*handshakeTransportError)
	return ok
}
