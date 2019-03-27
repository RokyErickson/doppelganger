package url

import (
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

const moshURLPrefix = "mosh:"

func isSHOOPMOSHURL(raw string) bool {
	return strings.HasPrefix(strings.ToLower(raw), moshURLPrefix)
}

func parseSHOOPMOSH(raw string) (*URL, error) {
	raw = raw[len(moshURLPrefix):]
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
		Protocol: Protocol_MOSH,
		Username: username,
		Hostname: hostname,
		Port:     port,
		Path:     path,
	}, nil
}
