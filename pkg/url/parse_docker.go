package url

import (
	"strings"

	"github.com/pkg/errors"
)

const (
	dockerURLPrefix                    = "docker://"
	DockerHostEnvironmentVariable      = "DOCKER_HOST"
	DockerTLSVerifyEnvironmentVariable = "DOCKER_TLS_VERIFY"
	DockerCertPathEnvironmentVariable  = "DOCKER_CERT_PATH"
)

var DockerEnvironmentVariables = []string{
	DockerHostEnvironmentVariable,
	DockerTLSVerifyEnvironmentVariable,
	DockerCertPathEnvironmentVariable,
}

func isDockerURL(raw string) bool {
	return strings.HasPrefix(strings.ToLower(raw), dockerURLPrefix)
}

func parseDocker(raw string, alpha bool) (*URL, error) {
	raw = raw[len(dockerURLPrefix):]
	var username string
	for i, r := range raw {
		if r == '/' {
			break
		} else if r == '@' {
			username = raw[:i]
			raw = raw[i+1:]
			break
		}
	}
	var container, path string
	for i, r := range raw {
		if r == '/' {
			container = raw[:i]
			path = raw[i:]
			break
		}
	}
	if container == "" {
		return nil, errors.New("empty container name")
	} else if path == "" {
		return nil, errors.New("empty path")
	}

	if len(path) > 1 && path[1] == '~' {
		path = path[1:]
	}
	if isWindowsPath(path[1:]) {
		path = path[1:]
	}
	environment := make(map[string]string, len(DockerEnvironmentVariables))
	for _, variable := range DockerEnvironmentVariables {
		value, _ := getEnvironmentVariable(variable, alpha)
		environment[variable] = value
	}

	return &URL{
		Protocol:    Protocol_Docker,
		Username:    username,
		Hostname:    container,
		Path:        path,
		Environment: environment,
	}, nil
}
