package docker

import (
	"strings"

	"github.com/RokyErickson/doppelganger/pkg/url"
)

func setDockerVariables(remote *url.URL) map[string]string {

	environment := make(map[string]string)

	for _, variable := range url.DockerEnvironmentVariables {
		environment[variable] = remote.Environment[variable]
	}

	return environment
}

func findEnviromentVariable(outputBlock, variable string) (string, bool) {

	outputBlock = strings.Replace(outputBlock, "\r\n", "\n", -1)
	outputBlock = strings.TrimSpace(outputBlock)
	environment := strings.Split(outputBlock, "\n")

	for _, line := range environment {
		if strings.HasPrefix(line, variable+"=") {
			return line[len(variable)+1:], true
		}
	}

	return "", false
}
