package sync

import (
	"crypto/sha1"
	"fmt"
	"hash"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/pkg/errors"
)

func testCreateScanCycle(temporaryDirectory string, entry *Entry, contentMap map[string][]byte, ignores []string, symlinkMode SymlinkMode, expectEqual bool) error {

	root, parent, err := testTransitionCreate(temporaryDirectory, entry, contentMap, false)
	if err != nil {
		return errors.Wrap(err, "unable to create test content")
	}
	defer os.RemoveAll(parent)

	hasher := newTestHasher()

	snapshot, preservesExecutability, _, cache, ignoreCache, err := Scan(root, hasher, nil, ignores, nil, symlinkMode)
	if !preservesExecutability {
		snapshot = PropagateExecutability(nil, entry, snapshot)
	}
	if err != nil {
		return errors.Wrap(err, "unable to perform scan")
	} else if cache == nil {
		return errors.New("nil cache returned")
	} else if ignoreCache == nil {
		return errors.New("nil ignore cache returned")
	} else if expectEqual && !snapshot.Equal(entry) {
		return errors.New("snapshot not equal to expected")
	} else if !expectEqual && snapshot.Equal(entry) {
		return errors.New("snapshot should not have equaled original")
	}

	return nil
}

func testCreateScanCycleWithPermutations(entry *Entry, contentMap map[string][]byte, ignores []string, symlinkMode SymlinkMode, expectEqual bool) error {
	for _, temporaryDirectory := range testingTemporaryDirectories() {
		if err := testCreateScanCycle(temporaryDirectory.path, entry, contentMap, ignores, symlinkMode, expectEqual); err != nil {
			return errors.Wrap(err, fmt.Sprintf("create/scan cycle failed for %s temporary directory", temporaryDirectory.name))
		}
	}

	return nil
}

func TestScanNilRoot(t *testing.T) {
	if err := testCreateScanCycleWithPermutations(testNilEntry, nil, nil, SymlinkMode_SymlinkPortable, true); err != nil {
		t.Error("creation/scan cycle failed:", err)
	}
}

func TestScanFile1Root(t *testing.T) {
	if err := testCreateScanCycleWithPermutations(testFile1Entry, testFile1ContentMap, nil, SymlinkMode_SymlinkPortable, true); err != nil {
		t.Error("creation/scan cycle failed:", err)
	}
}

func TestScanFile2Root(t *testing.T) {
	if err := testCreateScanCycleWithPermutations(testFile2Entry, testFile2ContentMap, nil, SymlinkMode_SymlinkPortable, true); err != nil {
		t.Error("creation/scan cycle failed:", err)
	}
}

func TestScanFile3Root(t *testing.T) {
	if err := testCreateScanCycleWithPermutations(testFile3Entry, testFile3ContentMap, nil, SymlinkMode_SymlinkPortable, true); err != nil {
		t.Error("creation/scan cycle failed:", err)
	}
}

func TestScanDirectory1Root(t *testing.T) {
	if err := testCreateScanCycleWithPermutations(testDirectory1Entry, testDirectory1ContentMap, nil, SymlinkMode_SymlinkPortable, true); err != nil {
		t.Error("creation/scan cycle failed:", err)
	}
}

func TestScanDirectory2Root(t *testing.T) {
	if err := testCreateScanCycleWithPermutations(testDirectory2Entry, testDirectory2ContentMap, nil, SymlinkMode_SymlinkPortable, true); err != nil {
		t.Error("creation/scan cycle failed:", err)
	}
}

func TestScanDirectory3Root(t *testing.T) {
	if err := testCreateScanCycleWithPermutations(testDirectory3Entry, testDirectory3ContentMap, nil, SymlinkMode_SymlinkPortable, true); err != nil {
		t.Error("creation/scan cycle failed:", err)
	}
}

func TestScanDirectorySaneSymlinkSane(t *testing.T) {
	if err := testCreateScanCycleWithPermutations(testDirectoryWithSaneSymlink, nil, nil, SymlinkMode_SymlinkPortable, true); err != nil {
		t.Error("sane symlink not allowed inside root with sane symlink mode:", err)
	}
}

func TestScanDirectorySaneSymlinkIgnore(t *testing.T) {
	if err := testCreateScanCycleWithPermutations(testDirectoryWithSaneSymlink, nil, nil, SymlinkMode_SymlinkIgnore, false); err != nil {
		t.Error("sane symlink not allowed inside root with ignore symlink mode:", err)
	}
}

func TestScanDirectorySaneSymlinkPOSIXRaw(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	if err := testCreateScanCycleWithPermutations(testDirectoryWithSaneSymlink, nil, nil, SymlinkMode_SymlinkPOSIXRaw, true); err != nil {
		t.Error("sane symlink not allowed inside root with POSIX raw symlink mode:", err)
	}
}

func TestScanDirectoryInvalidSymlinkNotSane(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	if testCreateScanCycleWithPermutations(testDirectoryWithInvalidSymlink, nil, nil, SymlinkMode_SymlinkPortable, true) == nil {
		t.Error("invalid symlink allowed inside root with sane symlink mode")
	}
}

func TestScanDirectoryInvalidSymlinkIgnore(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	if err := testCreateScanCycleWithPermutations(testDirectoryWithInvalidSymlink, nil, nil, SymlinkMode_SymlinkIgnore, false); err != nil {
		t.Error("invalid symlink not allowed inside root with ignore symlink mode:", err)
	}
}

func TestScanDirectoryInvalidSymlinkPOSIXRaw(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	if err := testCreateScanCycleWithPermutations(testDirectoryWithInvalidSymlink, nil, nil, SymlinkMode_SymlinkPOSIXRaw, true); err != nil {
		t.Error("invalid symlink not allowed inside root with POSIX raw symlink mode:", err)
	}
}

func TestScanDirectoryEscapingSymlinkSane(t *testing.T) {
	if testCreateScanCycleWithPermutations(testDirectoryWithEscapingSymlink, nil, nil, SymlinkMode_SymlinkPortable, true) == nil {
		t.Error("escaping symlink allowed inside root with sane symlink mode")
	}
}

func TestScanDirectoryEscapingSymlinkIgnore(t *testing.T) {
	if err := testCreateScanCycleWithPermutations(testDirectoryWithEscapingSymlink, nil, nil, SymlinkMode_SymlinkIgnore, false); err != nil {
		t.Error("escaping symlink not allowed inside root with ignore symlink mode:", err)
	}
}

func TestScanDirectoryEscapingSymlinkPOSIXRaw(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	if err := testCreateScanCycleWithPermutations(testDirectoryWithEscapingSymlink, nil, nil, SymlinkMode_SymlinkPOSIXRaw, true); err != nil {
		t.Error("escaping symlink not allowed inside root with POSIX raw symlink mode:", err)
	}
}

func TestScanDirectoryAbsoluteSymlinkSane(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	if testCreateScanCycleWithPermutations(testDirectoryWithAbsoluteSymlink, nil, nil, SymlinkMode_SymlinkPortable, true) == nil {
		t.Error("escaping symlink allowed inside root with sane symlink mode")
	}
}

func TestScanDirectoryAbsoluteSymlinkIgnore(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	if err := testCreateScanCycleWithPermutations(testDirectoryWithAbsoluteSymlink, nil, nil, SymlinkMode_SymlinkIgnore, false); err != nil {
		t.Error("escaping symlink not allowed inside root with ignore symlink mode:", err)
	}
}

func TestScanDirectoryAbsoluteSymlinkPOSIXRaw(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	if err := testCreateScanCycleWithPermutations(testDirectoryWithAbsoluteSymlink, nil, nil, SymlinkMode_SymlinkPOSIXRaw, true); err != nil {
		t.Error("escaping symlink not allowed inside root with POSIX raw symlink mode:", err)
	}
}

func TestScanPOSIXRawNotAllowedOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip()
	}
	if testCreateScanCycleWithPermutations(testDirectoryWithSaneSymlink, nil, nil, SymlinkMode_SymlinkPOSIXRaw, true) == nil {
		t.Error("POSIX raw symlink mode allowed for scan on Windows")
	}
}

func TestScanInvalidIgnores(t *testing.T) {
	if testCreateScanCycleWithPermutations(testDirectory1Entry, testDirectory1ContentMap, []string{""}, SymlinkMode_SymlinkPortable, true) == nil {
		t.Error("scan allowed with invalid ignore specification")
	}
}

func TestScanIgnore(t *testing.T) {
	if err := testCreateScanCycleWithPermutations(testDirectory1Entry, testDirectory1ContentMap, []string{"second directory"}, SymlinkMode_SymlinkPortable, false); err != nil {
		t.Error("unexpected result when ignoring directory:", err)
	}
}

func TestScanIgnoreDirectory(t *testing.T) {
	if err := testCreateScanCycleWithPermutations(testDirectory1Entry, testDirectory1ContentMap, []string{"directory/"}, SymlinkMode_SymlinkPortable, false); err != nil {
		t.Error("unexpected result when ignoring directory:", err)
	}
}

func TestScanFileNotIgnoredOnDirectorySpecification(t *testing.T) {
	if err := testCreateScanCycleWithPermutations(testDirectory1Entry, testDirectory1ContentMap, []string{"file/"}, SymlinkMode_SymlinkPortable, true); err != nil {
		t.Error("unexpected result when ignoring directory:", err)
	}
}

func TestScanSubfileNotIgnoredOnRootSpecification(t *testing.T) {
	if err := testCreateScanCycleWithPermutations(testDirectory1Entry, testDirectory1ContentMap, []string{"/subfile.exe"}, SymlinkMode_SymlinkPortable, true); err != nil {
		t.Error("unexpected result when ignoring directory:", err)
	}
}

func TestScanSymlinkRoot(t *testing.T) {

	parent, err := ioutil.TempDir("", "doppelganger_simulated")
	if err != nil {
		t.Fatal("unable to create temporary directory:", err)
	}
	defer os.RemoveAll(parent)

	root := filepath.Join(parent, "root")

	if err := os.Symlink("relative", root); err != nil {
		t.Fatal("unable to create symlink:", err)
	}

	if _, _, _, _, _, err := Scan(root, sha1.New(), nil, nil, nil, SymlinkMode_SymlinkPortable); err == nil {
		t.Error("scan of symlink root allowed")
	}
}

type rescanHashProxy struct {
	hash.Hash
	t *testing.T
}

func (p *rescanHashProxy) Sum(b []byte) []byte {
	p.t.Error("rehashing occurred")
	return p.Hash.Sum(b)
}

func TestEfficientRescan(t *testing.T) {

	root, parent, err := testTransitionCreate("", testDirectory1Entry, testDirectory1ContentMap, false)
	if err != nil {
		t.Fatal("unable to create test content on disk:", err)
	}
	defer os.RemoveAll(parent)

	hasher := newTestHasher()

	snapshot, preservesExecutability, _, cache, ignoreCache, err := Scan(root, hasher, nil, nil, nil, SymlinkMode_SymlinkPortable)
	if !preservesExecutability {
		snapshot = PropagateExecutability(nil, testDirectory1Entry, snapshot)
	}
	if err != nil {
		t.Fatal("unable to create snapshot:", err)
	} else if cache == nil {
		t.Fatal("nil cache returned")
	} else if ignoreCache == nil {
		t.Fatal("nil ignore cache returned")
	} else if !snapshot.Equal(testDirectory1Entry) {
		t.Error("snapshot did not match expected")
	}

	hasher = &rescanHashProxy{hasher, t}
	snapshot, preservesExecutability, _, cache, ignoreCache, err = Scan(root, hasher, cache, nil, nil, SymlinkMode_SymlinkPortable)
	if !preservesExecutability {
		snapshot = PropagateExecutability(nil, testDirectory1Entry, snapshot)
	}
	if err != nil {
		t.Fatal("unable to rescan:", err)
	} else if cache == nil {
		t.Fatal("nil second cache returned")
	} else if ignoreCache == nil {
		t.Fatal("nil second ignore cache returned")
	} else if !snapshot.Equal(testDirectory1Entry) {
		t.Error("second snapshot did not match expected")
	}
}

func TestScanCrossDeviceFail(t *testing.T) {

	fat32Subroot := os.Getenv("DOPPELGANGER_TEST_FAT32_SUBROOT")
	if fat32Subroot == "" {
		t.Skip()
	}

	parent := filepath.Dir(fat32Subroot)

	hasher := newTestHasher()

	if _, _, _, _, _, err := Scan(parent, hasher, nil, nil, nil, SymlinkMode_SymlinkPortable); err == nil {
		t.Error("scan across device boundary did not fail")
	}
}
