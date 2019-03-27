// +build !windows

package filesystem

import (
	"testing"

	"golang.org/x/sys/unix"
)

func TestModePermissionsMaskMatchesOS(t *testing.T) {

	if ModePermissionsMask != Mode(unix.S_IRWXU|unix.S_IRWXG|unix.S_IRWXO) {
		t.Error("ModePermissionsMask does not match expected value")
	}

	if ModePermissionUserRead != Mode(unix.S_IRUSR) {
		t.Error("ModePermissionUserRead does not match expected")
	}

	if ModePermissionUserWrite != Mode(unix.S_IWUSR) {
		t.Error("ModePermissionUserWrite does not match expected")
	}

	if ModePermissionUserExecute != Mode(unix.S_IXUSR) {
		t.Error("ModePermissionUserExecute does not match expected")
	}

	if ModePermissionGroupRead != Mode(unix.S_IRGRP) {
		t.Error("ModePermissionGroupRead does not match expected")
	}

	if ModePermissionGroupWrite != Mode(unix.S_IWGRP) {
		t.Error("ModePermissionGroupWrite does not match expected")
	}

	if ModePermissionGroupExecute != Mode(unix.S_IXGRP) {
		t.Error("ModePermissionGroupExecute does not match expected")
	}

	if ModePermissionOthersRead != Mode(unix.S_IROTH) {
		t.Error("ModePermissionOthersRead does not match expected")
	}

	if ModePermissionOthersWrite != Mode(unix.S_IWOTH) {
		t.Error("ModePermissionOthersWrite does not match expected")
	}

	if ModePermissionOthersExecute != Mode(unix.S_IXOTH) {
		t.Error("ModePermissionOthersExecute does not match expected")
	}
}
