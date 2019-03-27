package filesystem

import (
	"io"
)

type ReadableFile interface {
	io.Reader
	io.Seeker
	io.Closer
}

type WritableFile interface {
	io.Writer
	io.Closer
}
