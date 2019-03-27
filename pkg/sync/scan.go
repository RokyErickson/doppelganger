package sync

import (
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"

	"golang.org/x/text/unicode/norm"

	"github.com/golang/protobuf/ptypes"

	fs "github.com/RokyErickson/doppelganger/pkg/filesystem"
)

const (
	scannerCopyBufferSize       = 32 * 1024
	defaultInitialCacheCapacity = 1024
)

type scanner struct {
	root                   string
	hasher                 hash.Hash
	cache                  *Cache
	ignorer                *ignorer
	ignoreCache            IgnoreCache
	symlinkMode            SymlinkMode
	newCache               *Cache
	newIgnoreCache         IgnoreCache
	buffer                 []byte
	deviceID               uint64
	recomposeUnicode       bool
	preservesExecutability bool
}

func (s *scanner) file(path string, file fs.ReadableFile, metadata *fs.Metadata, parent *fs.Directory) (*Entry, error) {
	if file != nil {
		defer file.Close()
	}

	executable := s.preservesExecutability && anyExecutableBitSet(metadata.Mode)

	modificationTimeProto, err := ptypes.TimestampProto(metadata.ModificationTime)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert modification time format")
	}

	var digest []byte
	cached, hit := s.cache.Entries[path]
	match := hit &&
		(metadata.Mode&fs.ModeTypeMask) == (fs.Mode(cached.Mode)&fs.ModeTypeMask) &&
		modificationTimeProto.Seconds == cached.ModificationTime.Seconds &&
		modificationTimeProto.Nanos == cached.ModificationTime.Nanos &&
		metadata.Size == cached.Size &&
		metadata.FileID == cached.FileID
	if match {
		digest = cached.Digest
	}

	if digest == nil {

		if file == nil {
			file, err = parent.OpenFile(metadata.Name)
			if err != nil {
				return nil, errors.Wrap(err, "unable to open file")
			}
			defer file.Close()
		}

		s.hasher.Reset()

		if copied, err := io.CopyBuffer(s.hasher, file, s.buffer); err != nil {
			return nil, errors.Wrap(err, "unable to hash file contents")
		} else if uint64(copied) != metadata.Size {
			return nil, errors.New("hashed size mismatch")
		}

		digest = s.hasher.Sum(nil)
	}

	s.newCache.Entries[path] = &CacheEntry{
		Mode:             uint32(metadata.Mode),
		ModificationTime: modificationTimeProto,
		Size:             metadata.Size,
		FileID:           metadata.FileID,
		Digest:           digest,
	}

	return &Entry{
		Kind:       EntryKind_File,
		Executable: executable,
		Digest:     digest,
	}, nil
}

func (s *scanner) symbolicLink(path, name string, parent *fs.Directory, enforcePortable bool) (*Entry, error) {

	target, err := parent.ReadSymbolicLink(name)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read symbolic link target")
	}

	if enforcePortable {
		target, err = normalizeSymlinkAndEnsurePortable(path, target)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("invalid symbolic link (%s)", path))
		}
	} else if target == "" {
		return nil, errors.New("symbolic link target is empty")
	}

	return &Entry{
		Kind:   EntryKind_Symlink,
		Target: target,
	}, nil
}

func (s *scanner) directory(path string, directory *fs.Directory, metadata *fs.Metadata, parent *fs.Directory) (*Entry, error) {

	if directory != nil {
		defer directory.Close()
	}

	if metadata.DeviceID != s.deviceID {
		return nil, errors.New("scan crossed filesystem boundary")
	}

	var err error
	if directory == nil {
		directory, err = parent.OpenDirectory(metadata.Name)
		if err != nil {
			return nil, errors.Wrap(err, "unable to open directory")
		}
		defer directory.Close()
	}

	directoryContents, err := directory.ReadContents()
	if err != nil {
		return nil, errors.Wrap(err, "unable to read directory contents")
	}

	contents := make(map[string]*Entry, len(directoryContents))
	for _, c := range directoryContents {
		name := c.Name

		if fs.IsTemporaryFileName(name) {
			continue
		}

		if s.recomposeUnicode {
			name = norm.NFC.String(name)
		}

		contentPath := pathJoin(path, name)

		var kind EntryKind
		switch c.Mode & fs.ModeTypeMask {
		case fs.ModeTypeDirectory:
			kind = EntryKind_Directory
		case fs.ModeTypeFile:
			kind = EntryKind_File
		case fs.ModeTypeSymbolicLink:
			kind = EntryKind_Symlink
		default:
			continue
		}

		isDirectory := kind == EntryKind_Directory
		ignoreCacheKey := IgnoreCacheKey{contentPath, isDirectory}
		ignored, ok := s.ignoreCache[ignoreCacheKey]
		if !ok {
			ignored = s.ignorer.ignored(contentPath, isDirectory)
		}
		s.newIgnoreCache[ignoreCacheKey] = ignored
		if ignored {
			continue
		}

		var entry *Entry
		if kind == EntryKind_File {
			entry, err = s.file(contentPath, nil, c, directory)
		} else if kind == EntryKind_Symlink {
			if s.symlinkMode == SymlinkMode_SymlinkPortable {
				entry, err = s.symbolicLink(contentPath, name, directory, true)
			} else if s.symlinkMode == SymlinkMode_SymlinkIgnore {
				continue
			} else if s.symlinkMode == SymlinkMode_SymlinkPOSIXRaw {
				entry, err = s.symbolicLink(contentPath, name, directory, false)
			} else {
				panic("unsupported symlink mode")
			}
		} else if kind == EntryKind_Directory {
			entry, err = s.directory(contentPath, nil, c, directory)
		} else {
			panic("unhandled entry kind")
		}

		if err != nil {
			return nil, err
		}

		contents[name] = entry
	}

	return &Entry{
		Kind:     EntryKind_Directory,
		Contents: contents,
	}, nil
}

func Scan(root string, hasher hash.Hash, cache *Cache, ignores []string, ignoreCache IgnoreCache, symlinkMode SymlinkMode) (*Entry, bool, bool, *Cache, IgnoreCache, error) {
	if cache == nil {
		cache = &Cache{}
	}

	ignorer, err := newIgnorer(ignores)
	if err != nil {
		return nil, false, false, nil, nil, errors.Wrap(err, "unable to create ignorer")
	}

	if symlinkMode == SymlinkMode_SymlinkPOSIXRaw && runtime.GOOS == "windows" {
		return nil, false, false, nil, nil, errors.New("raw POSIX symlinks not supported on Windows")
	}

	initialCacheCapacity := defaultInitialCacheCapacity
	if cacheLength := len(cache.Entries); cacheLength != 0 {
		initialCacheCapacity = cacheLength
	}
	newCache := &Cache{
		Entries: make(map[string]*CacheEntry, initialCacheCapacity),
	}

	initialIgnoreCacheCapacity := defaultInitialCacheCapacity
	if ignoreCacheLength := len(ignoreCache); ignoreCacheLength != 0 {
		initialIgnoreCacheCapacity = ignoreCacheLength
	}
	newIgnoreCache := make(IgnoreCache, initialIgnoreCacheCapacity)

	s := &scanner{
		root:           root,
		hasher:         hasher,
		cache:          cache,
		ignorer:        ignorer,
		ignoreCache:    ignoreCache,
		symlinkMode:    symlinkMode,
		newCache:       newCache,
		newIgnoreCache: newIgnoreCache,
		buffer:         make([]byte, scannerCopyBufferSize),
	}

	rootObject, metadata, err := fs.Open(root, false)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, false, newCache, newIgnoreCache, nil
		} else {
			return nil, false, false, nil, nil, errors.Wrap(err, "unable to probe scan root")
		}
	}

	s.deviceID = metadata.DeviceID

	if rootType := metadata.Mode & fs.ModeTypeMask; rootType == fs.ModeTypeDirectory {
		rootDirectory, ok := rootObject.(*fs.Directory)
		if !ok {
			panic("invalid directory object returned from root open operation")
		}

		if decomposes, err := fs.DecomposesUnicode(rootDirectory); err != nil {
			rootDirectory.Close()
			return nil, false, false, nil, nil, errors.Wrap(err, "unable to probe root Unicode decomposition behavior")
		} else {
			s.recomposeUnicode = decomposes
		}

		if preserves, err := fs.PreservesExecutability(rootDirectory); err != nil {
			rootDirectory.Close()
			return nil, false, false, nil, nil, errors.Wrap(err, "unable to probe root executability preservation behavior")
		} else {
			s.preservesExecutability = preserves
		}

		if rootEntry, err := s.directory("", rootDirectory, metadata, nil); err != nil {
			return nil, false, false, nil, nil, err
		} else {
			return rootEntry, s.preservesExecutability, s.recomposeUnicode, newCache, newIgnoreCache, nil
		}
	} else if rootType == fs.ModeTypeFile {
		rootFile, ok := rootObject.(fs.ReadableFile)
		if !ok {
			panic("invalid file object returned from root open operation")
		}

		if preserves, err := fs.PreservesExecutabilityByPath(filepath.Dir(root)); err != nil {
			rootFile.Close()
			return nil, false, false, nil, nil, errors.Wrap(err, "unable to probe root parent executability preservation behavior")
		} else {
			s.preservesExecutability = preserves
		}

		if rootEntry, err := s.file("", rootFile, metadata, nil); err != nil {
			return nil, false, false, nil, nil, err
		} else {
			return rootEntry, s.preservesExecutability, false, newCache, newIgnoreCache, nil
		}
	} else {
		panic("invalid type returned from root open operation")
	}
}
