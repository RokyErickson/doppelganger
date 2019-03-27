package agent

import (
	//"bytes"
	"fmt"
	"strings"
	"time"
	//"unicode/utf8"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/doppelganger"
	"github.com/RokyErickson/doppelganger/pkg/filesystem"
	"github.com/RokyErickson/doppelganger/pkg/process"
	"github.com/RokyErickson/doppelganger/pkg/prompt"
	"github.com/RokyErickson/doppelganger/pkg/remote"
	"github.com/RokyErickson/doppelganger/pkg/session"
)

const (
	agentKillDelay = 5 * time.Second
)

func connect(
	transport Transport,
	prompter string,
	cmdExe bool,
	root,
	session string,
	version session.Version,
	configuration *session.Configuration,
	alpha bool,
) (session.Endpoint, bool, bool, error) {
	pathSeparator := "/"
	if cmdExe {
		pathSeparator = "\\"
	}
	agentInvocationPath := strings.Join([]string{
		filesystem.DoppelgangerDirectoryName,
		agentsDirectoryName,
		doppelganger.Version,
		agentBaseName,
	}, pathSeparator)

	command := fmt.Sprintf("%s %s", agentInvocationPath, ModeEndpoint)

	message := "Connecting to agent (POSIX)..."
	if cmdExe {
		message = "Connecting to agent (Windows)..."
	}
	if err := prompt.Message(prompter, message); err != nil {
		return nil, false, false, errors.Wrap(err, "unable to message prompter")
	}
	agentProcess := transport.Command(command)

	connection, err := process.NewConnection(agentProcess, agentKillDelay)
	if err != nil {
		return nil, false, false, errors.Wrap(err, "unable to create agent process connection")
	}

	endpoint, err := remote.NewEndpointClient(connection, root, session, version, configuration, alpha)

	connection.SetKillDelay(time.Duration(0))

	return endpoint, false, false, nil
}

func Dial(
	transport Transport,
	prompter,
	root,
	session string,
	version session.Version,
	configuration *session.Configuration,
	alpha bool,
) (session.Endpoint, error) {

	endpoint, tryInstall, cmdExe, err :=
		connect(transport, prompter, false, root, session, version, configuration, alpha)
	if err == nil {
		return endpoint, nil
	} else if cmdExe {
		endpoint, tryInstall, cmdExe, err =
			connect(transport, prompter, true, root, session, version, configuration, alpha)
		if err == nil {
			return endpoint, nil
		}
	}

	if !tryInstall {
		return nil, err
	}

	if err := install(transport, prompter); err != nil {
		return nil, errors.Wrap(err, "unable to install agent")
	}

	endpoint, _, _, err = connect(transport, prompter, cmdExe, root, session, version, configuration, alpha)
	if err != nil {
		return nil, err
	}
	return endpoint, nil
}
