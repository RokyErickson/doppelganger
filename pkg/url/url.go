package url

import (
	"github.com/pkg/errors"
)

func (u *URL) EnsureValid() error {
	if u == nil {
		return errors.New("nil URL")
	}

	if u.Protocol == Protocol_Local {
		if u.Username != "" {
			return errors.New("local URL with non-empty username")
		} else if u.Hostname != "" {
			return errors.New("local URL with non-empty hostname")
		} else if u.Port != 0 {
			return errors.New("local URL with non-zero port")
		} else if u.Path == "" {
			return errors.New("local URL with empty path")
		} else if len(u.Environment) != 0 {
			return errors.New("local URL with environment variables")
		}
	} else if u.Protocol == Protocol_Ipfs {
		if u.Username != "" {
			return errors.New("Ipfs URL with non-empty username")
		} else if u.Hostname != "" {
			return errors.New("Ipfs URL with non-empty hostname")
		} else if u.Port != 0 {
			return errors.New("Ipfs URL with non-zero port")
		} else if u.Path == "" {
			return errors.New("Ipfs URL with empty path")
		}
	} else if u.Protocol == Protocol_SSH {
		if u.Hostname == "" {
			return errors.New("SSH URL with empty hostname")
		} else if u.Path == "" {
			return errors.New("SSH URL with empty path")
		} else if len(u.Environment) != 0 {
			return errors.New("SSH URL with environment variables")
		}
	} else if u.Protocol == Protocol_Docker {
		if u.Hostname == "" {
			return errors.New("Docker URL with empty container identifier")
		} else if u.Port != 0 {
			return errors.New("Docker URL with non-zero port")
		} else if u.Path == "" {
			return errors.New("Docker URL with empty path")
		} else if !(u.Path[0] == '/' || u.Path[0] == '~' || isWindowsPath(u.Path)) {
			return errors.New("Docker URL with incorrect first path character")
		}
	} else {
		return errors.New("unknown or unsupported protocol")
	}

	return nil
}
