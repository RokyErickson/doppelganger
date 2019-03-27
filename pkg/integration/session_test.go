package integration

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/pkg/errors"

	"github.com/google/uuid"

	"github.com/RokyErickson/doppelganger/pkg/agent"
	"github.com/RokyErickson/doppelganger/pkg/daemon"
	"github.com/RokyErickson/doppelganger/pkg/prompt"
	"github.com/RokyErickson/doppelganger/pkg/protocols/local"
	"github.com/RokyErickson/doppelganger/pkg/session"
	"github.com/RokyErickson/doppelganger/pkg/url"

	_ "github.com/RokyErickson/doppelganger/pkg/protocols/docker"
	_ "github.com/RokyErickson/doppelganger/pkg/protocols/ipfs"
	_ "github.com/RokyErickson/doppelganger/pkg/protocols/local"
	_ "github.com/RokyErickson/doppelganger/pkg/protocols/ssh"
)

var daemonLock *daemon.Lock

var sessionManager *session.Manager

func init() {

	if err := agent.CopyBundleForTesting(); err != nil {
		panic(errors.Wrap(err, "unable to copy agent bundle for testing"))
	}

	if l, err := daemon.AcquireLock(); err != nil {
		panic(errors.Wrap(err, "unable to acquire daemon lock"))
	} else {
		daemonLock = l
	}

	if m, err := session.NewManager(); err != nil {
		panic(errors.Wrap(err, "unable to create session manager"))
	} else {
		sessionManager = m
	}

	agent.Housekeep()
	local.HousekeepCaches()
	local.HousekeepStaging()
}

func waitForSuccessfulSynchronizationCycle(sessionId string, allowConflicts, allowProblems bool) error {

	specification := []string{sessionId}

	var previousStateIndex uint64
	var states []*session.State
	var err error
	for {
		previousStateIndex, states, err = sessionManager.List(previousStateIndex, specification)
		if err != nil {
			return errors.Wrap(err, "unable to list session states")
		} else if len(states) != 1 {
			return errors.New("invalid number of session states returned")
		} else if states[0].SuccessfulSynchronizationCycles > 0 {
			if !allowProblems && (len(states[0].AlphaProblems) > 0 || len(states[0].BetaProblems) > 0) {
				return errors.New("problems detected (and disallowed)")
			} else if !allowConflicts && len(states[0].Conflicts) > 0 {
				return errors.New("conflicts detected (and disallowed)")
			}
			return nil
		}
	}
}

func testSessionLifecycle(prompter string, alpha, beta *url.URL, configuration *session.Configuration, allowConflicts, allowProblems bool) error {

	sessionId, err := sessionManager.Create(
		alpha, beta,
		configuration, &session.Configuration{}, &session.Configuration{},
		prompter,
	)
	if err != nil {
		return errors.Wrap(err, "unable to create session")
	}

	specification := []string{sessionId}

	if err := waitForSuccessfulSynchronizationCycle(sessionId, allowConflicts, allowProblems); err != nil {
		return errors.Wrap(err, "unable to wait for successful synchronization")
	}

	if err := sessionManager.Pause(specification, ""); err != nil {
		return errors.Wrap(err, "unable to pause session")
	}

	if err := sessionManager.Resume(specification, ""); err != nil {
		return errors.Wrap(err, "unable to resume session")
	}

	if err := waitForSuccessfulSynchronizationCycle(sessionId, allowConflicts, allowProblems); err != nil {
		return errors.Wrap(err, "unable to wait for additional synchronization")
	}

	if err := sessionManager.Resume(specification, ""); err != nil {
		return errors.Wrap(err, "unable to perform additional resume")
	}

	if err := sessionManager.Terminate(specification, ""); err != nil {
		return errors.Wrap(err, "unable to terminate session")
	}

	return nil
}

func TestSessionBothRootsNil(t *testing.T) {

	t.Parallel()

	directory, err := ioutil.TempDir("", "doppelganger_end_to_end")
	if err != nil {
		t.Fatal("unable to create temporary directory:", err)
	}
	defer os.RemoveAll(directory)

	alphaRoot := filepath.Join(directory, "alpha")
	betaRoot := filepath.Join(directory, "beta")

	alphaURL := &url.URL{Path: alphaRoot}
	betaURL := &url.URL{Path: betaRoot}

	configuration := &session.Configuration{}

	if err := testSessionLifecycle("", alphaURL, betaURL, configuration, false, false); err != nil {
		t.Fatal("session lifecycle test failed:", err)
	}
}

func TestSessionGOROOTSrcToBeta(t *testing.T) {

	endToEndTestMode := os.Getenv("DOPPELGANGER_TEST_END_TO_END")
	var sourceRoot string
	if endToEndTestMode == "" {
		t.Skip()
	} else if endToEndTestMode == "full" {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src")
	} else if endToEndTestMode == "slim" {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src", "bufio")
	} else {
		t.Fatal("unknown end-to-end test mode specified:", endToEndTestMode)
	}

	t.Parallel()

	directory, err := ioutil.TempDir("", "doppelganger_end_to_end")
	if err != nil {
		t.Fatal("unable to create temporary directory:", err)
	}
	defer os.RemoveAll(directory)

	alphaRoot := sourceRoot
	betaRoot := filepath.Join(directory, "beta")

	alphaURL := &url.URL{Path: alphaRoot}
	betaURL := &url.URL{Path: betaRoot}

	configuration := &session.Configuration{}

	if err := testSessionLifecycle("", alphaURL, betaURL, configuration, false, false); err != nil {
		t.Fatal("session lifecycle test failed:", err)
	}
}

func TestSessionGOROOTSrcToAlpha(t *testing.T) {

	endToEndTestMode := os.Getenv("DOPPELGANGER_TEST_END_TO_END")
	var sourceRoot string
	if endToEndTestMode == "" {
		t.Skip()
	} else if endToEndTestMode == "full" {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src")
	} else if endToEndTestMode == "slim" {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src", "bufio")
	} else {
		t.Fatal("unknown end-to-end test mode specified:", endToEndTestMode)
	}

	t.Parallel()

	directory, err := ioutil.TempDir("", "doppelganger_end_to_end")
	if err != nil {
		t.Fatal("unable to create temporary directory:", err)
	}
	defer os.RemoveAll(directory)

	alphaRoot := filepath.Join(directory, "alpha")
	betaRoot := sourceRoot

	alphaURL := &url.URL{Path: alphaRoot}
	betaURL := &url.URL{Path: betaRoot}

	configuration := &session.Configuration{}

	if err := testSessionLifecycle("", alphaURL, betaURL, configuration, false, false); err != nil {
		t.Fatal("session lifecycle test failed:", err)
	}
}

func TestSessionGOROOTSrcToBetaInMemory(t *testing.T) {

	endToEndTestMode := os.Getenv("DOPPELGANGER_TEST_END_TO_END")
	var sourceRoot string
	if endToEndTestMode == "" {
		t.Skip()
	} else if endToEndTestMode == "full" {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src")
	} else if endToEndTestMode == "slim" {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src", "bufio")
	} else {
		t.Fatal("unknown end-to-end test mode specified:", endToEndTestMode)
	}

	t.Parallel()

	directory, err := ioutil.TempDir("", "doppelganger_end_to_end")
	if err != nil {
		t.Fatal("unable to create temporary directory:", err)
	}
	defer os.RemoveAll(directory)

	alphaRoot := sourceRoot
	betaRoot := filepath.Join(directory, "beta")

	alphaURL := &url.URL{Path: alphaRoot}
	betaURL := &url.URL{
		Protocol: inMemoryProtocol,
		Path:     betaRoot,
	}

	configuration := &session.Configuration{}

	if err := testSessionLifecycle("", alphaURL, betaURL, configuration, false, false); err != nil {
		t.Fatal("session lifecycle test failed:", err)
	}
}

func TestSessionGOROOTSrcToBetaOverSSH(t *testing.T) {

	if os.Getenv("DOPPELGANGER_TEST_SSH") != "true" {
		t.Skip()
	}

	endToEndTestMode := os.Getenv("DOPPELGANGER_TEST_END_TO_END")
	var sourceRoot string
	if endToEndTestMode == "" {
		t.Skip()
	} else if endToEndTestMode == "full" {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src")
	} else if endToEndTestMode == "slim" {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src", "bufio")
	} else {
		t.Fatal("unknown end-to-end test mode specified:", endToEndTestMode)
	}

	t.Parallel()

	directory, err := ioutil.TempDir("", "doppelganger_end_to_end")
	if err != nil {
		t.Fatal("unable to create temporary directory:", err)
	}
	defer os.RemoveAll(directory)

	alphaRoot := sourceRoot
	betaRoot := filepath.Join(directory, "beta")

	alphaURL := &url.URL{Path: alphaRoot}
	betaURL := &url.URL{
		Protocol: url.Protocol_SSH,
		Hostname: "localhost",
		Path:     betaRoot,
	}

	configuration := &session.Configuration{}

	if err := testSessionLifecycle("", alphaURL, betaURL, configuration, false, false); err != nil {
		t.Fatal("session lifecycle test failed:", err)
	}
}

type testWindowsDockerTransportPrompter struct{}

func (t *testWindowsDockerTransportPrompter) Message(_ string) error {
	return nil
}

func (t *testWindowsDockerTransportPrompter) Prompt(_ string) (string, error) {
	return "yes", nil
}

func TestSessionGOROOTSrcToBetaOverDocker(t *testing.T) {

	if os.Getenv("DOPPELGANGER_TEST_DOCKER") != "true" {
		t.Skip()
	}

	endToEndTestMode := os.Getenv("DOPPELGANGER_TEST_END_TO_END")
	var sourceRoot string
	if endToEndTestMode == "" {
		t.Skip()
	} else if endToEndTestMode == "full" {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src")
	} else if endToEndTestMode == "slim" {
		sourceRoot = filepath.Join(runtime.GOROOT(), "src", "bufio")
	} else {
		t.Fatal("unknown end-to-end test mode specified:", endToEndTestMode)
	}

	t.Parallel()

	var prompter string
	if runtime.GOOS == "windows" {
		if p, err := prompt.RegisterPrompter(&testWindowsDockerTransportPrompter{}); err != nil {
			t.Fatal("unable to register prompter:", err)
		} else {
			prompter = p
			defer prompt.UnregisterPrompter(prompter)
		}
	}

	randomUUID, err := uuid.NewRandom()
	if err != nil {
		t.Fatal("unable to create random directory UUID:", err)
	}

	alphaRoot := sourceRoot
	betaRoot := "~/" + randomUUID.String()

	environment := make(map[string]string, len(url.DockerEnvironmentVariables))
	for _, variable := range url.DockerEnvironmentVariables {
		environment[variable] = os.Getenv(variable)
	}

	alphaURL := &url.URL{Path: alphaRoot}
	betaURL := &url.URL{
		Protocol:    url.Protocol_Docker,
		Username:    os.Getenv("DOPPELGANGER_TEST_DOCKER_USERNAME"),
		Hostname:    os.Getenv("DOPPELGANGER_TEST_DOCKER_CONTAINER_NAME"),
		Path:        betaRoot,
		Environment: environment,
	}

	if err := betaURL.EnsureValid(); err != nil {
		t.Fatal("beta URL is invalid:", err)
	}

	configuration := &session.Configuration{}

	if err := testSessionLifecycle(prompter, alphaURL, betaURL, configuration, false, false); err != nil {
		t.Fatal("session lifecycle test failed:", err)
	}
}
