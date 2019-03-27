package compression

import (
	"compress/flate"
	"io"

	"github.com/pkg/errors"
)

const (
	defaultCompressionLevel = 6
)

func NewDecompressingReader(source io.Reader) io.Reader {

	return flate.NewReader(source)
}

type automaticallyFlushingFlateWriter struct {
	compressor *flate.Writer
}

func (w *automaticallyFlushingFlateWriter) Write(buffer []byte) (int, error) {
	count, err := w.compressor.Write(buffer)
	if err != nil {
		return count, err
	} else if err = w.compressor.Flush(); err != nil {
		return 0, errors.Wrap(err, "unable to flush compressor")
	}
	return count, nil
}

func NewCompressingWriter(destination io.Writer) io.Writer {
	compressor, _ := flate.NewWriter(destination, defaultCompressionLevel)

	return &automaticallyFlushingFlateWriter{compressor}
}
