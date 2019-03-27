package sync

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"golang.org/x/text/unicode/norm"

	"github.com/golang/protobuf/ptypes"

	"github.com/RokyErickson/doppelganger/pkg/filesystem"
)

const (
	crossDeviceRenameTemporaryNamePrefix = filesystem.TemporaryNamePrefix + "cross-device-rename"
)

type Provider interface {
	Provide(path string, digest []byte) (string, error)
}

type transitioner struct {
	root                           string
	cache                          *Cache
	symlinkMode                    SymlinkMode
	defaultFilePermissionMode      filesystem.Mode
	defaultDirectoryPermissionMode filesystem.Mode
	defaultOwnership               *filesystem.OwnershipSpecification
	recomposeUnicode               bool
	provider                       Provider
	problems                       []*Problem
}

func (t *transitioner) recordProblem(path string, err error) {
	t.problems = append(t.problems, &Problem{Path: path, Error: err.Error()})
}

func (t *transitioner) nameExistsInDirectoryWithProperCase(
	name string,
	directory *filesystem.Directory,
) (bool, error) {
	names, err := directory.ReadContentNames()
	if err != nil {
		return false, errors.Wrap(err, "unable to read directory contents")
	}

	for _, n := range names {
		if !t.recomposeUnicode && n == name {
			return true, nil
		} else if t.recomposeUnicode && norm.NFC.String(n) == name {
			return true, nil
		}
	}

	return false, nil
}

func (t *transitioner) walkToParentAndComputeLeafName(
	path string,
	createRootParent bool,
	validateLeafCasing bool,
) (*filesystem.Directory, string, error) {
	if path == "" {
		rootParentPath, rootName := filepath.Split(t.root)
		if rootName == "" {
			return nil, "", errors.New("root path is a filesystem root")
		}

		if createRootParent {
			if err := os.MkdirAll(rootParentPath, os.FileMode(t.defaultDirectoryPermissionMode)); err != nil {
				return nil, "", errors.Wrap(err, "unable to create parent component of root path")
			}
		}

		if rootParent, _, err := filesystem.OpenDirectory(rootParentPath, true); err != nil {
			return nil, "", errors.Wrap(err, "unable to open synchronization root parent directory")
		} else {
			return rootParent, rootName, nil
		}
	}

	components := strings.Split(path, "/")
	parentComponents := components[:len(components)-1]
	leafName := components[len(components)-1]

	parent, _, err := filesystem.OpenDirectory(t.root, false)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to open synchronization root")
	}

	for _, component := range parentComponents {
		if found, err := t.nameExistsInDirectoryWithProperCase(component, parent); err != nil {
			parent.Close()
			return nil, "", errors.Wrap(err, "unable to verify parent path casing")
		} else if !found {
			parent.Close()
			return nil, "", errors.New("parent path does not exist or has incorrect casing")
		}

		if p, err := parent.OpenDirectory(component); err != nil {
			parent.Close()
			return nil, "", errors.Wrap(err, "unable to open parent component")
		} else {
			parent.Close()
			parent = p
		}
	}

	if validateLeafCasing {
		if found, err := t.nameExistsInDirectoryWithProperCase(leafName, parent); err != nil {
			parent.Close()
			return nil, "", errors.Wrap(err, "unable to verify path leaf name casing")
		} else if !found {
			parent.Close()
			return nil, "", errors.New("leaf name does not exist or has incorrect casing")
		}
	}

	return parent, leafName, nil
}

func (t *transitioner) ensureExpectedFile(parent *filesystem.Directory, name, path string, expected *Entry) error {

	cached, ok := t.cache.Entries[path]
	if !ok {
		return errors.New("unable to find cache information for path")
	}

	metadata, err := parent.ReadContentMetadata(name)
	if err != nil {
		return errors.Wrap(err, "unable to grab file statistics")
	}

	modificationTimeProto, err := ptypes.TimestampProto(metadata.ModificationTime)
	if err != nil {
		return errors.Wrap(err, "unable to convert modification time format")
	}

	match := metadata.Mode == filesystem.Mode(cached.Mode) &&
		modificationTimeProto.Seconds == cached.ModificationTime.Seconds &&
		modificationTimeProto.Nanos == cached.ModificationTime.Nanos &&
		metadata.Size == cached.Size &&
		metadata.FileID == cached.FileID &&
		bytes.Equal(cached.Digest, expected.Digest)
	if !match {
		return errors.New("modification detected")
	}

	return nil
}

func (t *transitioner) ensureExpectedSymbolicLink(parent *filesystem.Directory, name, path string, expected *Entry) error {

	target, err := parent.ReadSymbolicLink(name)
	if err != nil {
		return errors.Wrap(err, "unable to read symlink target")
	}

	if t.symlinkMode == SymlinkMode_SymlinkPortable {
		target, err = normalizeSymlinkAndEnsurePortable(path, target)
		if err != nil {
			return errors.Wrap(err, "unable to normalize target in portable mode")
		}
	}

	if target != expected.Target {
		return errors.New("symlink target does not match expected")
	}

	return nil
}

func (t *transitioner) ensureNotExists(parent *filesystem.Directory, name string) error {

	_, err := parent.ReadContentMetadata(name)

	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.Wrap(err, "unable to determine path existence")
	}

	return errors.New("path exists")
}

func (t *transitioner) removeFile(parent *filesystem.Directory, name, path string, expected *Entry) error {
	if err := t.ensureExpectedFile(parent, name, path, expected); err != nil {
		return errors.Wrap(err, "unable to validate existing file")
	}

	return parent.RemoveFile(name)
}

func (t *transitioner) removeSymbolicLink(parent *filesystem.Directory, name, path string, expected *Entry) error {

	if t.symlinkMode == SymlinkMode_SymlinkIgnore {
		return errors.New("symbolic link removal requested with symbolic links ignored")
	}

	if err := t.ensureExpectedSymbolicLink(parent, name, path, expected); err != nil {
		return errors.Wrap(err, "unable to validate existing symbolic link")
	}

	return parent.RemoveSymbolicLink(name)
}

func (t *transitioner) removeDirectory(parent *filesystem.Directory, name, path string, expected *Entry) bool {
	directory, err := parent.OpenDirectory(name)
	if err != nil {
		t.recordProblem(path, errors.Wrap(err, "unable to open directory"))
		return false
	}

	contents, err := directory.ReadContents()
	if err != nil {
		directory.Close()
		t.recordProblem(path, errors.Wrap(err, "unable to read directory contents"))
		return false
	}

	unknownContentEncountered := false
	for _, c := range contents {
		contentName := c.Name
		if t.recomposeUnicode {
			contentName = norm.NFC.String(contentName)
		}

		contentPath := pathJoin(path, contentName)

		entry, ok := expected.Contents[contentName]
		if !ok {
			t.recordProblem(contentPath, errors.New("unknown content encountered on disk"))
			unknownContentEncountered = true
			continue
		}

		if entry.Kind == EntryKind_Directory {
			if !t.removeDirectory(directory, contentName, contentPath, entry) {
				continue
			}
		} else if entry.Kind == EntryKind_File {
			if err = t.removeFile(directory, contentName, contentPath, entry); err != nil {
				t.recordProblem(contentPath, errors.Wrap(err, "unable to remove file"))
				continue
			}
		} else if entry.Kind == EntryKind_Symlink {
			if err = t.removeSymbolicLink(directory, contentName, contentPath, entry); err != nil {
				t.recordProblem(contentPath, errors.Wrap(err, "unable to remove symbolic link"))
				continue
			}
		} else {
			t.recordProblem(contentPath, errors.New("unknown entry type found in removal target"))
			continue
		}

		delete(expected.Contents, contentName)
	}

	directory.Close()

	if !unknownContentEncountered && len(expected.Contents) == 0 {
		if err := parent.RemoveDirectory(name); err != nil {
			t.recordProblem(path, errors.Wrap(err, "unable to remove directory"))
		} else {
			return true
		}
	}

	return false
}

func (t *transitioner) remove(path string, entry *Entry) *Entry {

	if entry == nil {
		return nil
	}

	parent, name, err := t.walkToParentAndComputeLeafName(path, false, true)
	if err != nil {
		t.recordProblem(path, errors.Wrap(err, "unable to walk to transition root"))
		return entry
	}
	defer parent.Close()

	if entry.Kind == EntryKind_Directory {

		entryCopy := entry.Copy()

		if !t.removeDirectory(parent, name, path, entryCopy) {
			return entryCopy
		}
	} else if entry.Kind == EntryKind_File {
		if err := t.removeFile(parent, name, path, entry); err != nil {
			t.recordProblem(path, errors.Wrap(err, "unable to remove file"))
			return entry
		}
	} else if entry.Kind == EntryKind_Symlink {
		if err := t.removeSymbolicLink(parent, name, path, entry); err != nil {
			t.recordProblem(path, errors.Wrap(err, "unable to remove symlink"))
			return entry
		}
	} else {
		t.recordProblem(path, errors.New("removal requested for unknown entry type"))
		return entry
	}

	return nil
}

func (t *transitioner) findAndMoveStagedFileIntoPlace(
	path string,
	target *Entry,
	parent *filesystem.Directory,
	name string,
) error {

	mode := t.defaultFilePermissionMode
	if target.Executable {
		mode = markExecutableForReaders(mode)
	}

	stagedPath, err := t.provider.Provide(path, target.Digest)
	if err != nil {
		return errors.Wrap(err, "unable to locate staged file")
	}

	if err := filesystem.SetPermissionsByPath(stagedPath, t.defaultOwnership, mode); err != nil {
		return errors.Wrap(err, "unable to set staged file permissions")
	}

	renameErr := filesystem.Rename(nil, stagedPath, parent, name)
	if renameErr == nil {
		return nil
	}

	if !filesystem.IsCrossDeviceError(renameErr) {
		return errors.Wrap(renameErr, "unable to relocate staged file")
	}

	stagedFile, err := os.Open(stagedPath)
	if err != nil {
		return errors.Wrap(err, "unable to open staged file")
	}

	temporaryName, temporary, err := parent.CreateTemporaryFile(crossDeviceRenameTemporaryNamePrefix)
	if err != nil {
		stagedFile.Close()
		return errors.Wrap(err, "unable to create temporary file for cross-device rename")
	}

	_, copyErr := io.Copy(temporary, stagedFile)

	stagedFile.Close()
	temporary.Close()

	if copyErr != nil {
		parent.RemoveFile(temporaryName)
		return errors.Wrap(copyErr, "unable to copy file contents")
	}

	if err := parent.SetPermissions(temporaryName, t.defaultOwnership, mode); err != nil {
		parent.RemoveFile(temporaryName)
		return errors.Wrap(err, "unable to set intermediate file permissions")
	}

	if err := filesystem.Rename(parent, temporaryName, parent, name); err != nil {
		parent.RemoveFile(temporaryName)
		return errors.Wrap(err, "unable to relocate intermediate file")
	}

	os.Remove(stagedPath)

	return nil
}

func (t *transitioner) swapFile(path string, oldEntry, newEntry *Entry) error {

	parent, name, err := t.walkToParentAndComputeLeafName(path, false, true)
	if err != nil {
		return errors.Wrap(err, "unable to walk to transition root")
	}
	defer parent.Close()

	if err := t.ensureExpectedFile(parent, name, path, oldEntry); err != nil {
		return errors.Wrap(err, "unable to validate existing file")
	}

	if bytes.Equal(oldEntry.Digest, newEntry.Digest) {

		mode := t.defaultFilePermissionMode
		if newEntry.Executable {
			mode = markExecutableForReaders(mode)
		}

		if err := parent.SetPermissions(name, t.defaultOwnership, mode); err != nil {
			return errors.Wrap(err, "unable to change file permissions")
		}

		return nil
	}

	return t.findAndMoveStagedFileIntoPlace(path, newEntry, parent, name)
}

func (t *transitioner) createFile(parent *filesystem.Directory, name, path string, target *Entry) error {

	if err := t.ensureNotExists(parent, name); err != nil {
		return errors.Wrap(err, "unable to ensure path does not exist")
	}

	return t.findAndMoveStagedFileIntoPlace(path, target, parent, name)
}

func (t *transitioner) createSymbolicLink(parent *filesystem.Directory, name, path string, target *Entry) error {

	if t.symlinkMode == SymlinkMode_SymlinkIgnore {
		return errors.New("symbolic link creation requested with symbolic links ignored")
	} else if t.symlinkMode == SymlinkMode_SymlinkPortable {
		if normalized, err := normalizeSymlinkAndEnsurePortable(path, target.Target); err != nil || normalized != target.Target {
			return errors.New("symbolic link was not in normalized form or was not portable")
		}
	}

	if err := t.ensureNotExists(parent, name); err != nil {
		return errors.Wrap(err, "unable to ensure path does not exist")
	}

	return parent.CreateSymbolicLink(name, target.Target)
}

func (t *transitioner) createDirectory(parent *filesystem.Directory, name, path string, target *Entry) *Entry {

	if err := t.ensureNotExists(parent, name); err != nil {
		t.recordProblem(path, errors.Wrap(err, "unable to ensure path does not exist"))
		return nil
	}

	if err := parent.CreateDirectory(name); err != nil {
		t.recordProblem(path, errors.Wrap(err, "unable to create directory"))
		return nil
	}

	created := target.copySlim()

	if err := parent.SetPermissions(name, t.defaultOwnership, t.defaultDirectoryPermissionMode); err != nil {
		t.recordProblem(path, errors.Wrap(err, "unable to set directory permissions"))
		return created
	}

	var directory *filesystem.Directory
	if len(target.Contents) > 0 {

		created.Contents = make(map[string]*Entry)

		if d, err := parent.OpenDirectory(name); err != nil {
			t.recordProblem(path, errors.Wrap(err, "unable to open new directory"))
			return created
		} else {
			directory = d
			defer directory.Close()
		}
	}

	for name, entry := range target.Contents {

		contentPath := pathJoin(path, name)

		if entry.Kind == EntryKind_Directory {
			if c := t.createDirectory(directory, name, contentPath, entry); c != nil {
				created.Contents[name] = c
			}
		} else if entry.Kind == EntryKind_File {
			if err := t.createFile(directory, name, contentPath, entry); err != nil {
				t.recordProblem(contentPath, errors.Wrap(err, "unable to create file"))
			} else {
				created.Contents[name] = entry
			}
		} else if entry.Kind == EntryKind_Symlink {
			if err := t.createSymbolicLink(directory, name, contentPath, entry); err != nil {
				t.recordProblem(contentPath, errors.Wrap(err, "unable to create symbolic link"))
			} else {
				created.Contents[name] = entry
			}
		} else {
			t.recordProblem(contentPath, errors.New("creation requested for unknown entry type"))
		}
	}

	return created
}

func (t *transitioner) create(path string, target *Entry) *Entry {

	if target == nil {
		return nil
	}

	parent, name, err := t.walkToParentAndComputeLeafName(path, false, false)
	if err != nil {
		t.recordProblem(path, errors.Wrap(err, "unable to walk to transition root parent"))
		return nil
	}
	defer parent.Close()

	if target.Kind == EntryKind_Directory {
		return t.createDirectory(parent, name, path, target)
	} else if target.Kind == EntryKind_File {
		if err := t.createFile(parent, name, path, target); err != nil {
			t.recordProblem(path, errors.Wrap(err, "unable to create file"))
			return nil
		} else {
			return target
		}
	} else if target.Kind == EntryKind_Symlink {
		if err := t.createSymbolicLink(parent, name, path, target); err != nil {
			t.recordProblem(path, errors.Wrap(err, "unable to create symlink"))
			return nil
		} else {
			return target
		}
	} else {
		t.recordProblem(path, errors.New("creation requested for unknown entry type"))
		return nil
	}
}

func Transition(
	root string,
	transitions []*Change,
	cache *Cache,
	symlinkMode SymlinkMode,
	defaultFilePermissionMode filesystem.Mode,
	defaultDirectoryPermissionMode filesystem.Mode,
	defaultOwnership *filesystem.OwnershipSpecification,
	recomposeUnicode bool,
	provider Provider,
) ([]*Entry, []*Problem) {
	transitioner := &transitioner{
		root:                           root,
		cache:                          cache,
		symlinkMode:                    symlinkMode,
		defaultFilePermissionMode:      defaultFilePermissionMode,
		defaultDirectoryPermissionMode: defaultDirectoryPermissionMode,
		defaultOwnership:               defaultOwnership,
		recomposeUnicode:               recomposeUnicode,
		provider:                       provider,
	}

	var results []*Entry

	for _, t := range transitions {

		fileToFile := t.Old != nil && t.New != nil &&
			t.Old.Kind == EntryKind_File &&
			t.New.Kind == EntryKind_File
		if fileToFile {
			if err := transitioner.swapFile(t.Path, t.Old, t.New); err != nil {
				results = append(results, t.Old)
				transitioner.recordProblem(t.Path, errors.Wrap(err, "unable to swap file"))
			} else {
				results = append(results, t.New)
			}
			continue
		}

		if r := transitioner.remove(t.Path, t.Old); r != nil {
			results = append(results, r)
			continue
		}

		results = append(results, transitioner.create(t.Path, t.New))
	}

	return results, transitioner.problems
}
