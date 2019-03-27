package filesystem

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestLockerFailOnDirectory(t *testing.T) {

	directory, err := ioutil.TempDir("", "doppelganger_filesystem_lock")
	if err != nil {
		t.Fatal("unable to create temporary directory:", err)
	}
	defer os.RemoveAll(directory)

	if _, err := NewLocker(directory, 0600); err == nil {
		t.Fatal("creating a locker on a directory path succeeded")
	}
}

func TestLockerCycle(t *testing.T) {

	lockfile, err := ioutil.TempFile("", "doppelganger_filesystem_lock")
	if err != nil {
		t.Fatal("unable to create temporary lock file:", err)
	} else if err = lockfile.Close(); err != nil {
		t.Error("unable to close temporary lock file:", err)
	}
	defer os.Remove(lockfile.Name())

	locker, err := NewLocker(lockfile.Name(), 0600)
	if err != nil {
		t.Fatal("unable to create locker:", err)
	}

	if err := locker.Lock(true); err != nil {
		t.Fatal("unable to acquire lock:", err)
	}

	if err := locker.Unlock(); err != nil {
		t.Fatal("unable to release lock:", err)
	}
}
