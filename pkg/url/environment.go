package url

import (
	"os"
)

const (
	alphaSpecificEnvironmentVariablePrefix = "DOPPELGANGER_ALPHA_"
	betaSpecificEnvironmentVariablePrefix  = "DOPPELGANGER_BETA_"
)

var lookupEnv = os.LookupEnv

func getEnvironmentVariable(name string, alpha bool) (string, bool) {

	if name == "" {
		return "", false
	}

	var endpointSpecificName string
	if alpha {
		endpointSpecificName = alphaSpecificEnvironmentVariablePrefix + name
	} else {
		endpointSpecificName = betaSpecificEnvironmentVariablePrefix + name
	}
	if value, ok := lookupEnv(endpointSpecificName); ok {
		return value, true
	}

	return lookupEnv(name)
}
