package url

import (
	"runtime"
	"strconv"

	"github.com/pkg/errors"
)

func isSCPSSHURL(raw string) bool {

	if runtime.GOOS == "windows" && isWindowsPath(raw) {
		return false
	}

	for _, c := range raw {
		if c == ':' {
			return true
		} else if c == '/' {
			break
		}
	}

	return false
}

func parseSCPSSH(raw string) (*URL, error) {
	var username string
	for i, r := range raw {
		if r == ':' {
			break
		} else if r == '@' {
			if i == 0 {
				return nil, errors.New("empty username specified")
			}
			username = raw[:i]
			raw = raw[i+1:]
			break
		}
	}

	var hostname string
	for i, r := range raw {
		if r == ':' {
			if i == 0 {
				return nil, errors.New("empty hostname")
			}
			hostname = raw[:i]
			raw = raw[i+1:]
			break
		}
	}
	if hostname == "" {
		return nil, errors.New("no hostname present")
	}

	var port uint32
	for i, r := range raw {

		if '0' <= r && r <= '9' {
			continue
		}

		if r == ':' {
			if port64, err := strconv.ParseUint(raw[:i], 10, 16); err != nil {
				return nil, errors.New("invalid port value specified")
			} else {
				port = uint32(port64)
				raw = raw[i+1:]
			}
		}

		break
	}

	path := raw
	if path == "" {
		return nil, errors.New("empty path")
	}

	return &URL{
		Protocol: Protocol_SSH,
		Username: username,
		Hostname: hostname,
		Port:     port,
		Path:     path,
	}, nil
}
