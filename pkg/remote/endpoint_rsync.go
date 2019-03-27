package remote

import (
	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/encoding"
	"github.com/RokyErickson/doppelganger/pkg/rsync"
)

const (
	rsyncTransmissionGroupSize = 10
)

type protobufRsyncEncoder struct {
	encoder  *encoding.ProtobufEncoder
	buffered int
	error    error
}

func newProtobufRsyncEncoder(encoder *encoding.ProtobufEncoder) *protobufRsyncEncoder {
	return &protobufRsyncEncoder{encoder: encoder}
}

func (e *protobufRsyncEncoder) Encode(transmission *rsync.Transmission) error {
	if e.error != nil {
		return errors.Wrap(e.error, "previous error encountered")
	}

	if err := e.encoder.EncodeWithoutFlush(transmission); err != nil {
		e.error = errors.Wrap(err, "unable to encode transmission")
		return e.error
	}

	e.buffered++

	if e.buffered == rsyncTransmissionGroupSize {
		if err := e.encoder.Flush(); err != nil {
			e.error = errors.Wrap(err, "unable to write encoded messages")
			return e.error
		}
		e.buffered = 0
	}

	return nil
}

func (e *protobufRsyncEncoder) Finalize() error {

	if e.error != nil {
		return errors.Wrap(e.error, "previous error encountered")
	}

	if err := e.encoder.Flush(); err != nil {
		return errors.Wrap(err, "unable to write encoded messages")
	}

	e.buffered = 0

	return nil
}

type protobufRsyncDecoder struct {
	decoder *encoding.ProtobufDecoder
}

func newProtobufRsyncDecoder(decoder *encoding.ProtobufDecoder) *protobufRsyncDecoder {
	return &protobufRsyncDecoder{decoder: decoder}
}

func (d *protobufRsyncDecoder) Decode(transmission *rsync.Transmission) error {

	return d.decoder.Decode(transmission)
}

func (d *protobufRsyncDecoder) Finalize() error {
	return nil
}
