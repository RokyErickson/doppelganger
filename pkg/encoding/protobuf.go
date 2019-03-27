package encoding

import (
	"bufio"
	"encoding/binary"
	"io"

	"github.com/pkg/errors"

	"github.com/golang/protobuf/proto"
)

const (
	protobufEncoderInitialBufferSize = 32 * 1024

	protobufEncoderMaximumPersistentBufferSize = 1024 * 1024

	protobufDecoderReaderBufferSize = 32 * 1024

	protobufDecoderInitialBufferSize = 32 * 1024

	protobufDecoderMaximumAllowedMessageSize = 100 * 1024 * 1024

	protobufDecoderMaximumPersistentBufferSize = 1024 * 1024
)

func LoadAndUnmarshalProtobuf(path string, message proto.Message) error {
	return loadAndUnmarshal(path, func(data []byte) error {
		return proto.Unmarshal(data, message)
	})
}

func MarshalAndSaveProtobuf(path string, message proto.Message) error {
	return marshalAndSave(path, func() ([]byte, error) {
		return proto.Marshal(message)
	})
}

type ProtobufEncoder struct {
	writer io.Writer

	buffer *proto.Buffer
}

func NewProtobufEncoder(writer io.Writer) *ProtobufEncoder {
	return &ProtobufEncoder{
		writer: writer,
		buffer: proto.NewBuffer(make([]byte, 0, protobufEncoderInitialBufferSize)),
	}
}

func (e *ProtobufEncoder) EncodeWithoutFlush(message proto.Message) error {

	if err := e.buffer.EncodeMessage(message); err != nil {
		return errors.Wrap(err, "unable to encode message")
	}

	return nil
}

func (e *ProtobufEncoder) Flush() error {

	data := e.buffer.Bytes()

	if len(data) > 0 {
		if _, err := e.writer.Write(data); err != nil {
			return errors.Wrap(err, "unable to write message")
		}
	}

	if cap(data) > protobufEncoderMaximumPersistentBufferSize {
		e.buffer.SetBuf(make([]byte, 0, protobufEncoderMaximumPersistentBufferSize))
	} else {
		e.buffer.Reset()
	}

	return nil
}

func (e *ProtobufEncoder) Encode(message proto.Message) error {

	if err := e.EncodeWithoutFlush(message); err != nil {
		return err
	}

	return e.Flush()
}

type ProtobufDecoder struct {
	reader *bufio.Reader

	buffer []byte
}

func NewProtobufDecoder(reader io.Reader) *ProtobufDecoder {
	return &ProtobufDecoder{
		reader: bufio.NewReaderSize(reader, protobufDecoderReaderBufferSize),
		buffer: make([]byte, protobufDecoderInitialBufferSize),
	}
}

func (d *ProtobufDecoder) bufferWithSize(size int) []byte {

	if cap(d.buffer) >= size {
		return d.buffer[:size]
	}

	result := make([]byte, size)

	if size <= protobufDecoderMaximumPersistentBufferSize {
		d.buffer = result
	}

	return result
}

func (d *ProtobufDecoder) Decode(message proto.Message) error {

	length, err := binary.ReadUvarint(d.reader)
	if err != nil {
		return errors.Wrap(err, "unable to read message length")
	}

	if length > protobufDecoderMaximumAllowedMessageSize {
		return errors.New("message size too large")
	}

	messageBytes := d.bufferWithSize(int(length))

	if _, err := io.ReadFull(d.reader, messageBytes); err != nil {
		return errors.Wrap(err, "unable to read message")
	}

	if err := proto.Unmarshal(messageBytes, message); err != nil {
		return errors.Wrap(err, "unable to unmarshal message")
	}

	return nil
}
