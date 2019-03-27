package docker

import (
	"fmt"
	"github.com/RokyErickson/doppelganger/pkg/process"
	"github.com/RokyErickson/doppelganger/pkg/prompt"
	"github.com/RokyErickson/doppelganger/pkg/url"
	"github.com/pkg/errors"
	"github.com/polydawn/gosh"
)

const windowsContainerCopyNotification = `!!! ATTENTION !!!
In order to install its agent binary inside a Windows container, Doppelganger will
need to stop and re-start the associated container. This is necessary because
Hyper-V doesn't support copying files into running containers.

Would you like to continue? (yes/no)? `

type transport struct {
	remote                 *url.URL
	prompter               string
	containerProbed        bool
	containerIsWindows     bool
	containerHomeDirectory string
	containerUsername      string
	containerUserGroup     string
	containerProbeError    error
}

func newTransport(remote *url.URL, prompter string) (*transport, error) {

	return &transport{
		remote:   remote,
		prompter: prompter,
	}, nil
}

func (t *transport) command(command, workingDirectory, user string) gosh.Command {

	dockerArguments := []string{"docker", "exec", "--interactive"}

	if user != "" {
		dockerArguments = append(dockerArguments, "--user", user)
	} else if t.remote.Username != "" {
		dockerArguments = append(dockerArguments, "--user", t.remote.Username)
	}
	if workingDirectory != "" {
		dockerArguments = append(dockerArguments, "--workdir", workingDirectory)
	}

	environment := setDockerVariables(t.remote)

	dockerArguments = append(dockerArguments, t.remote.Hostname)

	dockerArguments = append(dockerArguments, command)

	dockerCommand := gosh.Gosh(dockerArguments,
		gosh.Opts{
			Env:      environment,
			Launcher: gosh.ExecCustomizingLauncher(process.DetachedProcessAttributes),
		},
	).Bake()

	return dockerCommand
}

func (t *transport) probeContainer() error {

	if t.containerProbeError != nil {
		return errors.Wrap(t.containerProbeError, "previous container probing failed")
	}

	if t.containerProbed {
		return nil
	}
	t.containerProbed = true

	var windows bool
	var home string
	var posixErr, windowsErr error

	if home == "" {
		t.containerProbeError = errors.Errorf(
			"container probing failed under POSIX hypothesis (%s) and Windows hypothesis (%s)",
			posixErr.Error(),
			windowsErr.Error(),
		)
		return t.containerProbeError
	}

	var username, group string
	t.containerIsWindows = windows
	t.containerHomeDirectory = home
	t.containerUsername = username
	t.containerUserGroup = group

	return nil
}

func (t *transport) changeContainerStatus(stop bool) gosh.Proc {

	operation := "start"
	if stop {
		operation = "stop"
	}
	environment := setDockerVariables(t.remote)

	dockerCommand := gosh.Gosh("docker", operation, t.remote.Hostname,
		gosh.Opts{
			Env:      environment,
			Launcher: gosh.ExecCustomizingLauncher(process.DetachedProcessAttributes),
		},
	).Bake()

	return dockerCommand.Run()
}

func (t *transport) Copy(localPath, remoteName string) error {

	if err := t.probeContainer(); err != nil {
		return errors.Wrap(err, "unable to probe container")
	}

	if t.containerIsWindows {
		if t.prompter == "" {
			return errors.New("no prompter for Docker copy behavior confirmation")
		}
		for {
			if response, err := prompt.Prompt(t.prompter, windowsContainerCopyNotification); err != nil {
				return errors.Wrap(err, "unable to prompt for Docker copy behavior confirmation")
			} else if response == "no" {
				return errors.New("user cancelled copy operation")
			} else if response == "yes" {
				break
			}
		}
		t.changeContainerStatus(true)
	}

	var containerPath string
	if t.containerIsWindows {
		containerPath = fmt.Sprintf("%s:%s\\%s",
			t.remote.Hostname,
			t.containerHomeDirectory,
			remoteName,
		)
	} else {
		containerPath = fmt.Sprintf("%s:%s/%s",
			t.remote.Hostname,
			t.containerHomeDirectory,
			remoteName,
		)
	}

	environment := setDockerVariables(t.remote)

	dockerCommand := gosh.Gosh("docker", "cp", localPath, containerPath,
		gosh.Opts{
			Env:      environment,
			Launcher: gosh.ExecCustomizingLauncher(process.DetachedProcessAttributes),
		},
	).Bake()

	dockerCommand.Run()

	if !t.containerIsWindows {
		chownCommand := fmt.Sprintf(
			"chown %s:%s %s",
			t.containerUsername,
			t.containerUserGroup,
			remoteName,
		)
		t.command(chownCommand, t.containerHomeDirectory, "root").Run()
	}

	if t.containerIsWindows {
		t.changeContainerStatus(false)
	}

	return nil
}

func (t *transport) Command(command string) gosh.Command {
	if err := t.probeContainer(); err != nil {
		panic("container not probed")
	}

	return t.command(command, t.containerHomeDirectory, "")
}
