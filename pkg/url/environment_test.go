package url

import (
	"testing"
)

const (
	alphaSpecificDockerHostEnvironmentVariable     = "DOPPELGANGER_ALPHA_DOCKER_HOST"
	betaSpecificDockerTLSVerifyEnvironmentVariable = "DOPPELGANGER_BETA_DOCKER_TLS_VERIFY"
	defaultDockerHost                              = "unix:///default/docker.sock"
	alphaSpecificDockerHost                        = "unix:///alpha/docker.sock"
	defaultDockerTLSVerify                         = "sure!"
	betaSpecificDockerTLSVerify                    = "true"
)

var mockEnvironment = map[string]string{
	DockerHostEnvironmentVariable:                  defaultDockerHost,
	alphaSpecificDockerHostEnvironmentVariable:     alphaSpecificDockerHost,
	DockerTLSVerifyEnvironmentVariable:             defaultDockerTLSVerify,
	betaSpecificDockerTLSVerifyEnvironmentVariable: betaSpecificDockerTLSVerify,
}

func mockLookupEnv(name string) (string, bool) {
	value, ok := mockEnvironment[name]
	return value, ok
}

func init() {
	lookupEnv = mockLookupEnv
}

func TestAlphaLookupAlphaSpecificExists(t *testing.T) {
	if value, ok := getEnvironmentVariable(DockerHostEnvironmentVariable, true); !ok {
		t.Fatal("unable to find alpha-specific value")
	} else if value != alphaSpecificDockerHost {
		t.Fatal("alpha-specific value does not match expected")
	}
}

func TestAlphaLookupOnlyDefaultExists(t *testing.T) {
	if value, ok := getEnvironmentVariable(DockerTLSVerifyEnvironmentVariable, true); !ok {
		t.Fatal("unable to find non-endpoint-specific value for alpha")
	} else if value != defaultDockerTLSVerify {
		t.Fatal("non-endpoint-specific value does not match expected")
	}
}

func TestAlphaLookupNeitherExists(t *testing.T) {
	if _, ok := getEnvironmentVariable(DockerCertPathEnvironmentVariable, true); ok {
		t.Fatal("able to find unset environment variable")
	}
}

func TestBetaLookupBetaSpecificExists(t *testing.T) {
	if value, ok := getEnvironmentVariable(DockerTLSVerifyEnvironmentVariable, false); !ok {
		t.Fatal("unable to find beta-specific value")
	} else if value != betaSpecificDockerTLSVerify {
		t.Fatal("beta-specific value does not match expected")
	}
}

func TestBetaLookupOnlyDefaultExists(t *testing.T) {
	if value, ok := getEnvironmentVariable(DockerHostEnvironmentVariable, false); !ok {
		t.Fatal("unable to find non-endpoint-specific value for alpha")
	} else if value != defaultDockerHost {
		t.Fatal("non-endpoint-specific value does not match expected")
	}
}

func TestBetaLookupNeitherExists(t *testing.T) {
	if _, ok := getEnvironmentVariable(DockerCertPathEnvironmentVariable, true); ok {
		t.Fatal("able to find unset environment variable")
	}
}
