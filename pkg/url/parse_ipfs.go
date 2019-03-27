package url

import "strings"

const (
	ipfsURLPrefix               = "ipfs:"
	IpfsPathEnvironmentVariable = "IPFS_PATH"
)

var IpfsEnvironmentVariables = []string{
	IpfsPathEnvironmentVariable,
}

func isIpfsURL(raw string) bool {
	return strings.HasPrefix(strings.ToLower(raw), ipfsURLPrefix)
}

func parseIpfs(raw string, alpha bool) (*URL, error) {
	path := raw[len(ipfsURLPrefix):]

	environment := make(map[string]string, len(IpfsEnvironmentVariables))
	for _, variable := range IpfsEnvironmentVariables {
		value, _ := getEnvironmentVariable(variable, alpha)
		environment[variable] = value
	}

	return &URL{
		Protocol:    Protocol_Ipfs,
		Path:        path,
		Environment: environment,
	}, nil
}
