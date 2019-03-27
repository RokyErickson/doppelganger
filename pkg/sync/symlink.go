package sync

import (
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

func (m SymlinkMode) IsDefault() bool {
	return m == SymlinkMode_SymlinkDefault
}

func (m *SymlinkMode) UnmarshalText(textBytes []byte) error {
	text := string(textBytes)

	switch text {
	case "ignore":
		*m = SymlinkMode_SymlinkIgnore
	case "portable":
		*m = SymlinkMode_SymlinkPortable
	case "posix-raw":
		*m = SymlinkMode_SymlinkPOSIXRaw
	default:
		return errors.Errorf("unknown symlink mode specification: %s", text)
	}

	return nil
}

func (m SymlinkMode) Supported() bool {
	switch m {
	case SymlinkMode_SymlinkIgnore:
		return true
	case SymlinkMode_SymlinkPortable:
		return true
	case SymlinkMode_SymlinkPOSIXRaw:
		return true
	default:
		return false
	}
}

func (m SymlinkMode) Description() string {
	switch m {
	case SymlinkMode_SymlinkDefault:
		return "Default"
	case SymlinkMode_SymlinkIgnore:
		return "Ignore"
	case SymlinkMode_SymlinkPortable:
		return "Portable"
	case SymlinkMode_SymlinkPOSIXRaw:
		return "POSIX Raw"
	default:
		return "Unknown"
	}
}

const (
	maximumPortableSymlinkTargetLength = 247
)

func normalizeSymlinkAndEnsurePortable(path, target string) (string, error) {
	if target == "" {
		return "", errors.New("target empty")
	}

	if len(target) > maximumPortableSymlinkTargetLength {
		return "", errors.New("target too long")
	}

	if strings.Index(target, ":") != -1 {
		return "", errors.New("colon in target (absolute or unsupported path)")
	}

	if runtime.GOOS == "windows" {
		target = strings.Replace(target, "\\", "/", -1)
	} else if strings.Index(target, "\\") != -1 {
		return "", errors.New("backslash in target")
	}
	if target[0] == '/' {
		return "", errors.New("target is absolute")
	}

	pathDepth := strings.Count(path, "/")
	for _, component := range strings.Split(target, "/") {

		if component == "." {

		} else if component == ".." {
			pathDepth--
		} else {
			pathDepth++
		}

		if pathDepth < 0 {
			return "", errors.New("target references location outside synchronization root")
		}
	}

	return target, nil
}
