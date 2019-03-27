package filesystem

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pkg/errors"

	"golang.org/x/sys/windows"

	aclapi "github.com/hectane/go-acl/api"
)

func ensureValidName(name string) error {

	if name == "." {
		return errors.New("name is directory reference")
	} else if name == ".." {
		return errors.New("name is parent directory reference")
	}

	if strings.IndexByte(name, os.PathSeparator) != -1 {
		return errors.New("path separator appears in name")
	} else if strings.IndexByte(name, '/') != -1 {
		return errors.New("alternate path separator appears in name")
	}

	return nil
}

type Directory struct {
	handle windows.Handle

	file *os.File
}

func (d *Directory) Close() error {

	if err := d.file.Close(); err != nil {
		windows.CloseHandle(d.handle)
		return errors.Wrap(err, "unable to close file object")
	}

	if err := windows.CloseHandle(d.handle); err != nil {
		return errors.Wrap(err, "unable to close file handle")
	}

	return nil
}

func (d *Directory) CreateDirectory(name string) error {

	if err := ensureValidName(name); err != nil {
		return err
	}

	return os.Mkdir(filepath.Join(d.file.Name(), name), 0700)
}

func (d *Directory) CreateTemporaryFile(pattern string) (string, WritableFile, error) {

	if err := ensureValidName(pattern); err != nil {
		return "", nil, err
	}

	file, err := ioutil.TempFile(d.file.Name(), pattern)
	if err != nil {
		return "", nil, err
	}

	name := filepath.Base(file.Name())

	return name, file, nil
}

func (d *Directory) CreateSymbolicLink(name, target string) error {

	if err := ensureValidName(name); err != nil {
		return err
	}

	return os.Symlink(target, filepath.Join(d.file.Name(), name))
}

func (d *Directory) SetPermissions(name string, ownership *OwnershipSpecification, mode Mode) error {

	if err := ensureValidName(name); err != nil {
		return err
	}

	path := filepath.Join(d.file.Name(), name)

	if ownership != nil && (ownership.ownerSID != nil || ownership.groupSID != nil) {

		var information uint32
		if ownership.ownerSID != nil {
			information |= aclapi.OWNER_SECURITY_INFORMATION
		}
		if ownership.groupSID != nil {
			information |= aclapi.GROUP_SECURITY_INFORMATION
		}

		if err := aclapi.SetNamedSecurityInfo(
			path,
			aclapi.SE_FILE_OBJECT,
			information,
			ownership.ownerSID,
			ownership.groupSID,
			0,
			0,
		); err != nil {
			return errors.Wrap(err, "unable to set ownership information")
		}
	}

	mode = mode & ModePermissionsMask
	if mode != 0 {
		if err := os.Chmod(path, os.FileMode(mode)); err != nil {
			return errors.Wrap(err, "unable to set permission bits")
		}
	}

	return nil
}

func (d *Directory) openHandle(name string, wantDirectory bool) (string, windows.Handle, error) {

	if err := ensureValidName(name); err != nil {
		return "", 0, err
	}

	path := filepath.Join(d.file.Name(), name)

	path16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return "", 0, errors.Wrap(err, "unable to convert path to UTF-16")
	}

	handle, err := windows.CreateFile(
		path16,
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		if os.IsNotExist(err) {
			return "", 0, err
		}
		return "", 0, errors.Wrap(err, "unable to open path")
	}

	var metadata windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &metadata); err != nil {
		windows.CloseHandle(handle)
		return "", 0, errors.Wrap(err, "unable to query file metadata")
	}

	if metadata.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		windows.CloseHandle(handle)
		return "", 0, errors.New("path pointed to symbolic link")
	} else if wantDirectory && metadata.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 {
		windows.CloseHandle(handle)
		return "", 0, errors.New("path pointed to non-directory location")
	}

	return path, handle, nil
}

func (d *Directory) OpenDirectory(name string) (*Directory, error) {

	path, handle, err := d.openHandle(name, true)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open directory handle")
	}

	file, err := os.Open(path)
	if err != nil {
		windows.CloseHandle(handle)
		return nil, errors.Wrap(err, "unable to open file object for directory")
	}

	return &Directory{
		handle: handle,
		file:   file,
	}, nil
}

func (d *Directory) ReadContentNames() ([]string, error) {

	names, err := d.file.Readdirnames(0)
	if err != nil {
		return nil, err
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

	metadata, err := os.Lstat(filepath.Join(d.file.Name(), name))
	if err != nil {
		return nil, err
	}

	return &Metadata{
		Name:             name,
		Mode:             Mode(metadata.Mode()),
		Size:             uint64(metadata.Size()),
		ModificationTime: metadata.ModTime(),
	}, nil
}

func (d *Directory) ReadContents() ([]*Metadata, error) {

	contents, err := d.file.Readdir(0)
	if err != nil {
		return nil, err
	}

	results := make([]*Metadata, 0, len(contents))

	for _, content := range contents {

		name := content.Name()
		if name == "." || name == ".." {
			continue
		}

		results = append(results, &Metadata{
			Name:             name,
			Mode:             Mode(content.Mode()),
			Size:             uint64(content.Size()),
			ModificationTime: content.ModTime(),
		})
	}

	return results, nil
}

func (d *Directory) OpenFile(name string) (ReadableFile, error) {

	_, handle, err := d.openHandle(name, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open file handle")
	}

	file := os.NewFile(uintptr(handle), name)

	return file, nil
}

func (d *Directory) ReadSymbolicLink(name string) (string, error) {

	if err := ensureValidName(name); err != nil {
		return "", err
	}

	return os.Readlink(filepath.Join(d.file.Name(), name))
}

func (d *Directory) RemoveDirectory(name string) error {

	if err := ensureValidName(name); err != nil {
		return err
	}

	path := filepath.Join(d.file.Name(), name)

	path16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return errors.Wrap(err, "unable to convert path to UTF-16")
	}

	return windows.RemoveDirectory(path16)
}

func (d *Directory) RemoveFile(name string) error {

	if err := ensureValidName(name); err != nil {
		return err
	}

	path := filepath.Join(d.file.Name(), name)

	path16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return errors.Wrap(err, "unable to convert path to UTF-16")
	}

	return windows.DeleteFile(path16)
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
		sourceNameOrPath = filepath.Join(sourceDirectory.file.Name(), sourceNameOrPath)
	}

	if targetDirectory != nil {
		if err := ensureValidName(targetNameOrPath); err != nil {
			return errors.Wrap(err, "target name invalid")
		}
		targetNameOrPath = filepath.Join(targetDirectory.file.Name(), targetNameOrPath)
	}

	return os.Rename(sourceNameOrPath, targetNameOrPath)
}

const (
	_ERROR_NOT_SAME_DEVICE = 0x11
)

func IsCrossDeviceError(err error) bool {
	if linkErr, ok := err.(*os.LinkError); !ok {
		return false
	} else if errno, ok := linkErr.Err.(syscall.Errno); !ok {
		return false
	} else {
		return errno == _ERROR_NOT_SAME_DEVICE
	}
}
