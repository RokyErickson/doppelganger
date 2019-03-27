package url

import (
	"github.com/pkg/errors"
)

func Parse(raw string, alpha bool) (*URL, error) {

	if raw == "" {
		return nil, errors.New("empty URL")
	}
	if isDockerURL(raw) {
		return parseDocker(raw, alpha)
	} else if isIpfsURL(raw) {
		return parseIpfs(raw, alpha)
	} else if isSCPSSHURL(raw) {
		return parseSCPSSH(raw)
	} else {
		return parseLocal(raw)
	}
}
