package filesystem

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"golang.org/x/sys/windows"
)

func Open(path string, allowSymbolicLinkLeaf bool) (io.Closer, *Metadata, error) {

	if !filepath.IsAbs(path) {
		return nil, nil, errors.New("path is not absolute")
	}

	path16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to convert path to UTF-16")
	}

	flags := uint32(windows.FILE_ATTRIBUTE_NORMAL | windows.FILE_FLAG_BACKUP_SEMANTICS | windows.FILE_FLAG_OPEN_REPARSE_POINT)
	if allowSymbolicLinkLeaf {
		flags = uint32(windows.FILE_ATTRIBUTE_NORMAL | windows.FILE_FLAG_BACKUP_SEMANTICS)
	}
	handle, err := windows.CreateFile(
		path16,
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		flags,
		0,
	)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, err
		}
		return nil, nil, errors.Wrap(err, "unable to open path")
	}

	var rawMetadata windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &rawMetadata); err != nil {
		windows.CloseHandle(handle)
		return nil, nil, errors.Wrap(err, "unable to query file metadata")
	}

	if rawMetadata.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		windows.CloseHandle(handle)
		return nil, nil, ErrUnsupportedOpenType
	}

	isDirectory := rawMetadata.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0

	var file *os.File
	if isDirectory {
		file, err = os.Open(path)
		if err != nil {
			windows.CloseHandle(handle)
			return nil, nil, errors.Wrap(err, "unable to open file object for directory")
		}
	} else {
		file = os.NewFile(uintptr(handle), path)
	}

	fileMetadata, err := file.Stat()
	if err != nil {
		if isDirectory {
			windows.CloseHandle(handle)
		}
		file.Close()
		return nil, nil, errors.Wrap(err, "unable to query file metadata")
	}

	metadata := &Metadata{
		Name:             fileMetadata.Name(),
		Mode:             Mode(fileMetadata.Mode()),
		Size:             uint64(fileMetadata.Size()),
		ModificationTime: fileMetadata.ModTime(),
	}

	if isDirectory {
		return &Directory{
			handle: handle,
			file:   file,
		}, metadata, nil
	} else {
		return file, metadata, nil
	}
}
