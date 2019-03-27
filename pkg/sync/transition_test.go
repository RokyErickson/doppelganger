package sync

import (
	"bytes"
	"fmt"
	"hash"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/filesystem"
)

const (
	defaultFilePermissionMode = filesystem.ModePermissionUserRead | filesystem.ModePermissionUserWrite |
		filesystem.ModePermissionGroupRead |
		filesystem.ModePermissionOthersRead

	defaultDirectoryPermissionMode = filesystem.ModePermissionUserRead | filesystem.ModePermissionUserWrite | filesystem.ModePermissionUserExecute |
		filesystem.ModePermissionGroupRead | filesystem.ModePermissionGroupExecute |
		filesystem.ModePermissionOthersRead | filesystem.ModePermissionOthersExecute
)

type testEntryDecomposer struct {
	creation bool

	transitions []*Change
}

func (d *testEntryDecomposer) decompose(path string, entry *Entry) {

	if entry == nil {
		return
	}

	shallowEntry := entry.copySlim()

	if d.creation {
		d.transitions = append(d.transitions, &Change{Path: path, New: shallowEntry})
	}

	if entry.Kind == EntryKind_Directory {
		for name, entry := range entry.Contents {
			d.decompose(pathJoin(path, name), entry)
		}
	}

	if !d.creation {
		d.transitions = append(d.transitions, &Change{Path: path, Old: shallowEntry})
	}
}

func testDecomposeEntry(path string, entry *Entry, creation bool) []*Change {

	decomposer := &testEntryDecomposer{creation: creation}

	decomposer.decompose(path, entry)

	return decomposer.transitions
}

type testProvider struct {
	servingRoot string

	contentMap map[string][]byte

	hasher hash.Hash
}

func newTestProvider(contentMap map[string][]byte, hasher hash.Hash) (*testProvider, error) {

	servingRoot, err := ioutil.TempDir("", "doppelganger_provide_root")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create serving directory")
	}

	return &testProvider{
		servingRoot: servingRoot,
		contentMap:  contentMap,
		hasher:      hasher,
	}, nil
}

func (p *testProvider) Provide(path string, digest []byte) (string, error) {

	content, ok := p.contentMap[path]
	if !ok {
		return "", errors.New("unable to find content for path")
	}

	p.hasher.Reset()
	p.hasher.Write(content)
	if !bytes.Equal(p.hasher.Sum(nil), digest) {
		return "", errors.New("requested entry digest does not match expected")
	}

	temporaryFile, err := ioutil.TempFile(p.servingRoot, "doppelganger_provide")
	if err != nil {
		return "", errors.Wrap(err, "unable to create temporary file")
	}

	_, err = temporaryFile.Write(content)
	temporaryFile.Close()
	if err != nil {
		os.Remove(temporaryFile.Name())
		return "", errors.Wrap(err, "unable to write file contents")
	}

	return temporaryFile.Name(), nil
}

func (p *testProvider) finalize() error {
	return os.RemoveAll(p.servingRoot)
}

func testTransitionCreate(temporaryDirectory string, entry *Entry, contentMap map[string][]byte, decompose bool) (string, string, error) {

	parent, err := ioutil.TempDir(temporaryDirectory, "doppelganger_simulated")
	if err != nil {
		return "", "", errors.Wrap(err, "unable to create temporary root parent")
	}

	recomposeUnicode, err := filesystem.DecomposesUnicodeByPath(parent)
	if err != nil {
		os.RemoveAll(parent)
		return "", "", errors.Wrap(err, "unable to determine Unicode decomposition behavior")
	}

	root := filepath.Join(parent, "root")

	var transitions []*Change
	if decompose {
		transitions = testDecomposeEntry("", entry, true)
	} else {
		transitions = []*Change{{New: entry}}
	}

	provider, err := newTestProvider(contentMap, newTestHasher())
	if err != nil {
		os.RemoveAll(parent)
		return "", "", errors.Wrap(err, "unable to create test provider")
	}
	defer provider.finalize()

	if entries, problems := Transition(
		root,
		transitions,
		nil,
		SymlinkMode_SymlinkPOSIXRaw,
		defaultFilePermissionMode,
		defaultDirectoryPermissionMode,
		nil,
		recomposeUnicode,
		provider,
	); len(problems) != 0 {
		os.RemoveAll(parent)
		return "", "", errors.New("problems occurred during creation transition")
	} else if len(entries) != len(transitions) {
		os.RemoveAll(parent)
		return "", "", errors.New("unexpected number of entries returned from creation transition")
	} else {
		for e, entry := range entries {
			if !entry.Equal(transitions[e].New) {
				os.RemoveAll(parent)
				return "", "", errors.New("created entry does not match expected")
			}
		}
	}

	return root, parent, nil
}

func testTransitionRemove(root string, expected *Entry, cache *Cache, symlinkMode SymlinkMode, decompose bool) error {

	var transitions []*Change
	if decompose {
		transitions = testDecomposeEntry("", expected, false)
	} else {
		transitions = []*Change{{Old: expected}}
	}

	var recomposeUnicode bool
	if expected != nil && expected.Kind == EntryKind_Directory {
		if r, err := filesystem.DecomposesUnicodeByPath(root); err != nil {
			return errors.Wrap(err, "unable to determine Unicode decomposition behavior")
		} else {
			recomposeUnicode = r
		}
	}

	if entries, problems := Transition(
		root,
		transitions,
		cache,
		symlinkMode,
		defaultFilePermissionMode,
		defaultDirectoryPermissionMode,
		nil,
		recomposeUnicode,
		nil,
	); len(problems) != 0 {
		return errors.New("problems occurred during removal transition")
	} else if len(entries) != len(transitions) {
		return errors.New("unexpected number of entries returned from removal transition")
	} else {
		for _, entry := range entries {
			if entry != nil {
				return errors.New("post-removal entry non-nil")
			}
		}
	}

	// Success.
	return nil
}

type testContentModifier func(string, *Entry) (*Entry, error)

func testTransitionCycle(temporaryDirectory string, entry *Entry, contentMap map[string][]byte, decompose bool, modifier testContentModifier) error {

	root, parent, err := testTransitionCreate(temporaryDirectory, entry, contentMap, decompose)
	if err != nil {
		return errors.Wrap(err, "unable to create test content")
	}
	defer os.RemoveAll(parent)

	expected := entry

	if modifier != nil {
		if e, err := modifier(root, expected.Copy()); err != nil {
			return errors.Wrap(err, "modifier failed")
		} else {
			expected = e
		}
	}

	snapshot, preservesExecutability, _, cache, ignoreCache, err := Scan(root, newTestHasher(), nil, nil, nil, SymlinkMode_SymlinkPortable)
	if !preservesExecutability {
		snapshot = PropagateExecutability(nil, expected, snapshot)
	}
	if err != nil {
		return errors.Wrap(err, "unable to perform scan")
	} else if cache == nil {
		return errors.New("nil cache returned")
	} else if ignoreCache == nil {
		return errors.New("nil ignore cache returned")
	} else if modifier == nil && !snapshot.Equal(expected) {
		return errors.New("snapshot not equal to expected")
	}

	if err := testTransitionRemove(root, expected, cache, SymlinkMode_SymlinkPortable, decompose); err != nil {
		return errors.Wrap(err, "unable to remove test content")
	}

	return nil
}

func testTransitionCycleWithPermutations(entry *Entry, contentMap map[string][]byte, modifier testContentModifier, expectFailure bool) error {
	for _, decompose := range []bool{false, true} {

		caseName := "composed"
		if decompose {
			caseName = "decomposed"
		}

		for _, temporaryDirectory := range testingTemporaryDirectories() {
			err := testTransitionCycle(temporaryDirectory.path, entry, contentMap, decompose, modifier)
			if expectFailure && err == nil {
				return errors.Errorf("transition cycle succeeded unexpectedly in %s case for %s temporary directory", caseName, temporaryDirectory.name)
			} else if !expectFailure && err != nil {
				return errors.Wrap(err, fmt.Sprintf("transition cycle failed in %s case for %s temporary directory", caseName, temporaryDirectory.name))
			}
		}
	}

	return nil
}

func TestTransitionNilRoot(t *testing.T) {
	if err := testTransitionCycleWithPermutations(testNilEntry, nil, nil, false); err != nil {
		t.Error(err)
	}
}

func TestTransitionFile1Root(t *testing.T) {
	if err := testTransitionCycleWithPermutations(testFile1Entry, testFile1ContentMap, nil, false); err != nil {
		t.Error(err)
	}
}

func TestTransitionFile2Root(t *testing.T) {
	if err := testTransitionCycleWithPermutations(testFile2Entry, testFile2ContentMap, nil, false); err != nil {
		t.Error(err)
	}
}

func TestTransitionFile3Root(t *testing.T) {
	if err := testTransitionCycleWithPermutations(testFile3Entry, testFile3ContentMap, nil, false); err != nil {
		t.Error(err)
	}
}

func TestTransitionDirectory1Root(t *testing.T) {
	if err := testTransitionCycleWithPermutations(testDirectory1Entry, testDirectory1ContentMap, nil, false); err != nil {
		t.Error(err)
	}
}

func TestTransitionDirectory2Root(t *testing.T) {
	if err := testTransitionCycleWithPermutations(testDirectory2Entry, testDirectory2ContentMap, nil, false); err != nil {
		t.Error(err)
	}
}

func TestTransitionDirectory3Root(t *testing.T) {
	if err := testTransitionCycleWithPermutations(testDirectory3Entry, testDirectory3ContentMap, nil, false); err != nil {
		t.Error(err)
	}
}

func TestTransitionSwapFile(t *testing.T) {

	modifier := func(root string, expected *Entry) (*Entry, error) {

		_, _, recomposeUnicode, cache, ignoreCache, err := Scan(root, newTestHasher(), nil, nil, nil, SymlinkMode_SymlinkPortable)
		if err != nil {
			return nil, errors.Wrap(err, "unable to perform scan")
		} else if cache == nil {
			return nil, errors.New("nil cache returned")
		} else if ignoreCache == nil {
			return nil, errors.New("nil ignore cache returned")
		}

		transitions := []*Change{{
			Path: "file",
			Old:  testFile1Entry,
			New:  testFile2Entry,
		}}

		contentMap := map[string][]byte{
			"file": testFile2Contents,
		}

		provider, err := newTestProvider(contentMap, newTestHasher())
		if err != nil {
			return nil, errors.Wrap(err, "unable to create creation provider")
		}
		defer provider.finalize()

		if entries, problems := Transition(
			root,
			transitions,
			cache,
			SymlinkMode_SymlinkPortable,
			defaultFilePermissionMode,
			defaultDirectoryPermissionMode,
			nil,
			recomposeUnicode,
			provider,
		); len(problems) != 0 {
			return nil, errors.New("file swap transition failed")
		} else if len(entries) != 1 {
			return nil, errors.New("unexpected number of entries returned from swap transition")
		} else if !entries[0].Equal(testFile2Entry) {
			return nil, errors.New("file swap transition returned incorrect new file")
		} else {
			expected.Contents["file"] = entries[0]
		}

		return expected, nil
	}

	if err := testTransitionCycleWithPermutations(testDirectory1Entry, testDirectory1ContentMap, modifier, false); err != nil {
		t.Error(err)
	}
}

func TestTransitionSwapFileOnlyExecutableChange(t *testing.T) {
	modifier := func(root string, expected *Entry) (*Entry, error) {
		_, _, recomposeUnicode, cache, ignoreCache, err := Scan(root, newTestHasher(), nil, nil, nil, SymlinkMode_SymlinkPortable)
		if err != nil {
			return nil, errors.Wrap(err, "unable to perform scan")
		} else if cache == nil {
			return nil, errors.New("nil cache returned")
		} else if ignoreCache == nil {
			return nil, errors.New("nil ignore cache returned")
		}

		executableEntry := testFile1Entry.Copy()
		executableEntry.Executable = true

		transitions := []*Change{{
			Path: "file",
			Old:  testFile1Entry,
			New:  executableEntry,
		}}

		if entries, problems := Transition(
			root,
			transitions,
			cache,
			SymlinkMode_SymlinkPortable,
			defaultFilePermissionMode,
			defaultDirectoryPermissionMode,
			nil,
			recomposeUnicode,
			nil,
		); len(problems) != 0 {
			return nil, errors.New("file swap transition failed")
		} else if len(entries) != 1 {
			return nil, errors.New("unexpected number of entries returned from swap transition")
		} else if !entries[0].Equal(executableEntry) {
			return nil, errors.New("file swap transition returned incorrect new file")
		} else {
			expected.Contents["file"] = entries[0]
		}

		return expected, nil
	}

	if err := testTransitionCycleWithPermutations(testDirectory1Entry, testDirectory1ContentMap, modifier, false); err != nil {
		t.Error(err)
	}
}

func TestTransitionCaseConflict(t *testing.T) {

	expectCaseConflict := runtime.GOOS == "windows" || runtime.GOOS == "darwin"

	if err := testTransitionCycleWithPermutations(testDirectoryWithCaseConflict, testDirectoryWithCaseConflictContentMap, nil, expectCaseConflict); err != nil {
		t.Error("case conflict behavior not as expected:", err)
	}
}

func TestTransitionFailRemoveModifiedSubcontent(t *testing.T) {

	modifier := func(root string, expected *Entry) (*Entry, error) {
		if err := ioutil.WriteFile(filepath.Join(root, "file"), testFile3Contents, 0600); err != nil {
			return nil, errors.Wrap(err, "unable to modify file content")
		}
		return expected, nil
	}

	if err := testTransitionCycleWithPermutations(testDirectory1Entry, testDirectory1ContentMap, modifier, true); err != nil {
		t.Error(err)
	}
}

func TestTransitionFailRemoveModifiedRootFile(t *testing.T) {

	modifier := func(root string, expected *Entry) (*Entry, error) {
		if err := ioutil.WriteFile(root, testFile3Contents, 0600); err != nil {
			return nil, errors.Wrap(err, "unable to modify file content")
		}
		return expected, nil
	}

	if err := testTransitionCycleWithPermutations(testFile1Entry, testFile1ContentMap, modifier, true); err != nil {
		t.Error(err)
	}
}

func TestTransitionFailCreateInvalidPathCase(t *testing.T) {

	modifier := func(root string, expected *Entry) (*Entry, error) {

		_, _, recomposeUnicode, cache, ignoreCache, err := Scan(root, newTestHasher(), nil, nil, nil, SymlinkMode_SymlinkPortable)
		if err != nil {
			return nil, errors.Wrap(err, "unable to perform scan")
		} else if cache == nil {
			return nil, errors.New("nil cache returned")
		} else if ignoreCache == nil {
			return nil, errors.New("nil ignore cache returned")
		}

		if err := os.Rename(filepath.Join(root, "directory"), filepath.Join(root, "directory-temp")); err != nil {
			return nil, errors.Wrap(err, "unable to rename directory to temporary name")
		}
		if err := os.Rename(filepath.Join(root, "directory-temp"), filepath.Join(root, "DiRecTory")); err != nil {
			return nil, errors.Wrap(err, "unable to rename directory to different case name")
		}

		transitions := []*Change{{Path: "directory/new", New: testFile1Entry}}

		contentMap := map[string][]byte{
			"directory/new": testFile1Contents,
		}

		provider, err := newTestProvider(contentMap, newTestHasher())
		if err != nil {
			return nil, errors.Wrap(err, "unable to create creation provider")
		}
		defer provider.finalize()

		if entries, problems := Transition(
			root,
			transitions,
			cache,
			SymlinkMode_SymlinkPortable,
			defaultFilePermissionMode,
			defaultDirectoryPermissionMode,
			nil,
			recomposeUnicode,
			provider,
		); len(problems) == 0 {
			return nil, errors.New("transition succeeded unexpectedly")
		} else if len(entries) != 1 {
			return nil, errors.New("unexpected number of entries returned from creation transition")
		} else if entries[0] != nil {
			return nil, errors.New("failed creation transition returned non-nil entry")
		}

		if err := os.Rename(filepath.Join(root, "DiRecTory"), filepath.Join(root, "directory-temp")); err != nil {
			return nil, errors.Wrap(err, "unable to rename directory to temporary name")
		}
		if err := os.Rename(filepath.Join(root, "directory-temp"), filepath.Join(root, "directory")); err != nil {
			return nil, errors.Wrap(err, "unable to rename directory to original name")
		}

		return expected, nil
	}

	if err := testTransitionCycleWithPermutations(testDirectory1Entry, testDirectory1ContentMap, modifier, false); err != nil {
		t.Error(err)
	}
}

func TestTransitionFailRemoveInvalidPathCase(t *testing.T) {

	modifier := func(root string, expected *Entry) (*Entry, error) {
		if err := os.Rename(filepath.Join(root, "directory"), filepath.Join(root, "directory-temp")); err != nil {
			return nil, errors.Wrap(err, "unable to rename directory to temporary name")
		}
		if err := os.Rename(filepath.Join(root, "directory-temp"), filepath.Join(root, "DiRecTory")); err != nil {
			return nil, errors.Wrap(err, "unable to rename directory to different case name")
		}
		return expected, nil
	}

	if err := testTransitionCycleWithPermutations(testDirectory1Entry, testDirectory1ContentMap, modifier, true); err != nil {
		t.Error(err)
	}
}

func TestTransitionFailRemoveUnknownContent(t *testing.T) {

	modifier := func(root string, expected *Entry) (*Entry, error) {
		if err := filesystem.WriteFileAtomic(filepath.Join(root, "new test file"), []byte{0}, 0600); err != nil {
			return nil, errors.Wrap(err, "unable to create unknown content")
		}
		return expected, nil
	}

	if err := testTransitionCycleWithPermutations(testDirectory1Entry, testDirectory1ContentMap, modifier, true); err != nil {
		t.Error(err)
	}
}

func TestTransitionFailOnParentPathIsFile(t *testing.T) {

	var parent string
	if file, err := ioutil.TempFile("", "doppelganger_simulated"); err != nil {
		t.Fatal("unable to create temporary file:", err)
	} else if err = file.Close(); err != nil {
		t.Fatal("unable to close temporary file:", err)
	} else {
		parent = file.Name()
	}
	defer os.Remove(parent)

	root := filepath.Join(parent, "root")

	transitions := []*Change{{New: testDirectory1Entry}}

	provider, err := newTestProvider(testDirectory1ContentMap, newTestHasher())
	if err != nil {
		t.Fatal("unable to create test provider:", err)
	}
	defer provider.finalize()

	if entries, problems := Transition(
		root,
		transitions,
		nil,
		SymlinkMode_SymlinkPortable,
		defaultFilePermissionMode,
		defaultDirectoryPermissionMode,
		nil,
		false,
		provider,
	); len(problems) != 1 {
		t.Error("transition succeeded unexpectedly")
	} else if len(entries) != 1 {
		t.Error("transition returned invalid number of entries")
	} else if entries[0] != nil {
		t.Error("failed creation transition returned non-nil entry")
	}
}
