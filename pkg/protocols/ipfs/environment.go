package ipfs

import (
	"github.com/RokyErickson/doppelganger/pkg/url"
)

func setIpfsVariables(remote *url.URL) map[string]string {
	environment := make(map[string]string)

	for _, variable := range url.IpfsEnvironmentVariables {
		environment[variable] = remote.Environment[variable]
	}

	return environment
}
