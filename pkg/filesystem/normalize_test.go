package filesystem

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

func TestTildeNotPathSeparator(t *testing.T) {
	if os.IsPathSeparator('~') {
		t.Fatal("tilde considered path separator")
	}
}

func TestTildeExpandHome(t *testing.T) {

	expanded, err := tildeExpand("~")
	if err != nil {
		t.Fatal("tilde expansion failed:", err)
	}

	if expanded != HomeDirectory {
		t.Error("tilde-expanded path does not match expected")
	}
}

func TestTildeExpandHomeSlash(t *testing.T) {

	expanded, err := tildeExpand("~/")
	if err != nil {
		t.Fatal("tilde expansion failed:", err)
	}

	if expanded != HomeDirectory {
		t.Error("tilde-expanded path does not match expected")
	}
}

func TestTildeExpandHomeBackslash(t *testing.T) {

	expectFailure := runtime.GOOS != "windows"

	expanded, err := tildeExpand("~\\")
	if expectFailure && err == nil {
		t.Error("tilde expansion succeeded unexpectedly")
	} else if !expectFailure && err != nil {
		t.Fatal("tilde expansion failed:", err)
	}

	if expectFailure {
		return
	}

	if expanded != HomeDirectory {
		t.Error("tilde-expanded path does not match expected")
	}
}

func currentUsername() (string, error) {

	user, err := user.Current()
	if err != nil {
		return "", errors.Wrap(err, "unable to get current user")
	}

	if runtime.GOOS != "windows" {
		return user.Username, nil
	}

	if index := strings.IndexByte(user.Username, '\\'); index >= 0 {
		if index == len(user.Username) {
			return "", errors.New("domain extends to end of username")
		}
		return user.Username[index+1:], nil
	}
	return user.Username, nil
}

func TestTildeExpandLookup(t *testing.T) {

	username, err := currentUsername()
	if err != nil {
		t.Fatal("unable to look up current username:", err)
	}

	expanded, err := tildeExpand("~" + username)
	if err != nil {
		t.Fatal("tilde expansion failed:", err)
	}

	if expanded != HomeDirectory {
		t.Error("tilde-expanded path does not match expected")
	}
}

func TestTildeExpandLookupSlash(t *testing.T) {

	username, err := currentUsername()
	if err != nil {
		t.Fatal("unable to look up current username:", err)
	}

	expanded, err := tildeExpand(fmt.Sprintf("~%s/", username))
	if err != nil {
		t.Fatal("tilde expansion failed:", err)
	}

	if expanded != HomeDirectory {
		t.Error("tilde-expanded path does not match expected")
	}
}

func TestTildeExpandLookupBackslash(t *testing.T) {

	expectFailure := runtime.GOOS != "windows"

	username, err := currentUsername()
	if err != nil {
		t.Fatal("unable to look up current username:", err)
	}

	expanded, err := tildeExpand(fmt.Sprintf("~%s\\", username))
	if expectFailure && err == nil {
		t.Error("tilde expansion succeeded unexpectedly")
	} else if !expectFailure && err != nil {
		t.Fatal("tilde expansion failed:", err)
	}

	if expectFailure {
		return
	}

	if expanded != HomeDirectory {
		t.Error("tilde-expanded path does not match expected")
	}
}

func TestNormalizeHome(t *testing.T) {

	normalized, err := Normalize("~/somepath")
	if err != nil {
		t.Fatal("unable to normalize path:", err)
	}

	if normalized != filepath.Join(HomeDirectory, "somepath") {
		t.Error("normalized path does not match expected")
	}
}
