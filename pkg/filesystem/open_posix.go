// +build !windows

package filesystem

import (
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
)

func Open(path string, allowSymbolicLinkLeaf bool) (io.Closer, *Metadata, error) {
	flags := os.O_RDONLY | syscall.O_NOFOLLOW
	if allowSymbolicLinkLeaf {
		flags = os.O_RDONLY
	}
	file, err := os.OpenFile(path, flags, 0)
	if err != nil {
		return nil, nil, err
	}

	fileMetadata, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, errors.Wrap(err, "unable to query file metadata")
	}

	rawMetadata, ok := fileMetadata.Sys().(*syscall.Stat_t)
	if !ok {
		file.Close()
		return nil, nil, errors.New("unable to extract raw file metadata")
	}

	metadata := &Metadata{
		Name:             filepath.Base(path),
		Mode:             Mode(rawMetadata.Mode),
		Size:             uint64(rawMetadata.Size),
		ModificationTime: fileMetadata.ModTime(),
		DeviceID:         uint64(rawMetadata.Dev),
		FileID:           uint64(rawMetadata.Ino),
	}

	switch metadata.Mode & ModeTypeMask {
	case ModeTypeDirectory:
		return &Directory{
			file:       file,
			descriptor: int(file.Fd()),
		}, metadata, nil
	case ModeTypeFile:
		return file, metadata, nil
	default:
		file.Close()
		return nil, nil, ErrUnsupportedOpenType
	}
}
