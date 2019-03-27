// +build !windows

package filesystem

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"golang.org/x/sys/unix"
)

func ensureValidName(name string) error {

	if name == "." {
		return errors.New("name is directory reference")
	} else if name == ".." {
		return errors.New("name is parent directory reference")
	}

	if strings.IndexByte(name, os.PathSeparator) != -1 {
		return errors.New("path separator appears in name")
	}

	return nil
}

type Directory struct {
	file *os.File

	descriptor int
}

func (d *Directory) Close() error {
	return d.file.Close()
}

func (d *Directory) CreateDirectory(name string) error {

	if err := ensureValidName(name); err != nil {
		return err
	}

	return mkdirat(d.descriptor, name, 0700)
}

const maximumTemporaryFileRetries = 256

func (d *Directory) CreateTemporaryFile(pattern string) (string, WritableFile, error) {

	if err := ensureValidName(pattern); err != nil {
		return "", nil, err
	}

	var prefix, suffix string
	if starIndex := strings.LastIndex(pattern, "*"); starIndex != -1 {
		prefix, suffix = pattern[:starIndex], pattern[starIndex+1:]
	} else {
		prefix = pattern
	}

	for i := 0; i < maximumTemporaryFileRetries; i++ {

		name := prefix + strconv.Itoa(i) + suffix

		descriptor, err := openat(d.descriptor, name, os.O_RDWR|os.O_CREATE|os.O_EXCL|unix.O_CLOEXEC, 0600)
		if err != nil {
			if os.IsExist(err) {
				continue
			}
			return "", nil, errors.Wrap(err, "unable to create file")
		}

		file := os.NewFile(uintptr(descriptor), name)

		return name, file, nil
	}

	return "", nil, errors.New("exhausted potential file names")
}

func (d *Directory) CreateSymbolicLink(name, target string) error {

	if err := ensureValidName(name); err != nil {
		return err
	}

	return symlinkat(target, d.descriptor, name)
}

func (d *Directory) SetPermissions(name string, ownership *OwnershipSpecification, mode Mode) error {

	if err := ensureValidName(name); err != nil {
		return err
	}

	if ownership != nil && (ownership.ownerID != -1 || ownership.groupID != -1) {
		if err := fchownat(d.descriptor, name, ownership.ownerID, ownership.groupID, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			return errors.Wrap(err, "unable to set ownership information")
		}
	}

	mode = mode & ModePermissionsMask
	if mode != 0 {
		if err := fchmodat(d.descriptor, name, uint32(mode), unix.AT_SYMLINK_NOFOLLOW); err != nil {
			return errors.Wrap(err, "unable to set permission bits")
		}
	}

	return nil
}

func (d *Directory) open(name string, wantDirectory bool) (*os.File, int, error) {

	if err := ensureValidName(name); err != nil {
		return nil, 0, err
	}

	var descriptor int
	for {
		if d, err := openat(int(d.descriptor), name, os.O_RDONLY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0); err == nil {
			descriptor = d
			break
		} else if runtime.GOOS == "darwin" && err == unix.EINTR {
			continue
		} else {
			return nil, 0, err
		}
	}

	expectedType := ModeTypeFile
	if wantDirectory {
		expectedType = ModeTypeDirectory
	}
	var metadata unix.Stat_t
	if err := unix.Fstat(descriptor, &metadata); err != nil {
		unix.Close(descriptor)
		return nil, 0, errors.Wrap(err, "unable to query file metadata")
	} else if Mode(metadata.Mode)&ModeTypeMask != expectedType {
		unix.Close(descriptor)
		return nil, 0, errors.New("path is not of the expected type")
	}

	file := os.NewFile(uintptr(descriptor), name)

	return file, descriptor, nil
}

func (d *Directory) OpenDirectory(name string) (*Directory, error) {

	file, descriptor, err := d.open(name, true)
	if err != nil {
		return nil, err
	}

	return &Directory{
		file:       file,
		descriptor: descriptor,
	}, nil
}

func (d *Directory) ReadContentNames() ([]string, error) {

	names, err := d.file.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	if offset, err := unix.Seek(d.descriptor, 0, 0); err != nil {
		return nil, errors.Wrap(err, "unable to reset directory read pointer")
	} else if offset != 0 {
		return nil, errors.New("directory offset is non-zero after seek operation")
	}

	results := names[:0]
	for _, name := range names {

		if name == "." || name == ".." {
			continue
		}

		results = append(results, name)
	}

	return names, nil
}

func (d *Directory) ReadContentMetadata(name string) (*Metadata, error) {

	if err := ensureValidName(name); err != nil {
		return nil, err
	}

	var metadata unix.Stat_t
	if err := fstatat(d.descriptor, name, &metadata, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return nil, err
	}

	modificationTime := extractModificationTime(&metadata)

	return &Metadata{
		Name:             name,
		Mode:             Mode(metadata.Mode),
		Size:             uint64(metadata.Size),
		ModificationTime: time.Unix(modificationTime.Unix()),
		DeviceID:         uint64(metadata.Dev),
		FileID:           uint64(metadata.Ino),
	}, nil
}

func (d *Directory) ReadContents() ([]*Metadata, error) {

	names, err := d.ReadContentNames()
	if err != nil {
		return nil, errors.Wrap(err, "unable to read directory content names")
	}

	results := make([]*Metadata, 0, len(names))

	for _, name := range names {

		if m, err := d.ReadContentMetadata(name); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, errors.Wrap(err, "unable to access content metadata")
		} else {
			results = append(results, m)
		}
	}

	return results, nil
}

func (d *Directory) OpenFile(name string) (ReadableFile, error) {
	file, _, err := d.open(name, false)
	return file, err
}

const readlinkInitialBufferSize = 128

func (d *Directory) ReadSymbolicLink(name string) (string, error) {

	if err := ensureValidName(name); err != nil {
		return "", err
	}

	for size := readlinkInitialBufferSize; ; size *= 2 {

		buffer := make([]byte, size)

		count, err := readlinkat(d.descriptor, name, buffer)
		if err != nil {
			return "", &os.PathError{
				Op:   "readlinkat",
				Path: name,
				Err:  err,
			}
		}

		if count < 0 {
			return "", errors.New("unknown readlinkat failure occurred")
		}

		if count < size {
			return string(buffer[:count]), nil
		}
	}
}

func (d *Directory) RemoveDirectory(name string) error {

	if err := ensureValidName(name); err != nil {
		return err
	}

	return unlinkat(d.descriptor, name, _AT_REMOVEDIR)
}

func (d *Directory) RemoveFile(name string) error {

	if err := ensureValidName(name); err != nil {
		return err
	}

	return unlinkat(d.descriptor, name, 0)
}

func (d *Directory) RemoveSymbolicLink(name string) error {
	return d.RemoveFile(name)
}

func Rename(
	sourceDirectory *Directory, sourceNameOrPath string,
	targetDirectory *Directory, targetNameOrPath string,
) error {

	if sourceDirectory != nil {
		if err := ensureValidName(sourceNameOrPath); err != nil {
			return errors.Wrap(err, "source name invalid")
		}
	}

	if targetDirectory != nil {
		if err := ensureValidName(targetNameOrPath); err != nil {
			return errors.Wrap(err, "target name invalid")
		}
	}

	var sourceDescriptor, targetDescriptor int
	if sourceDirectory != nil {
		sourceDescriptor = sourceDirectory.descriptor
	}
	if targetDirectory != nil {
		targetDescriptor = targetDirectory.descriptor
	}

	return renameat(
		sourceDescriptor, sourceNameOrPath,
		targetDescriptor, targetNameOrPath,
	)
}

func IsCrossDeviceError(err error) bool {
	return err == unix.EXDEV
}
