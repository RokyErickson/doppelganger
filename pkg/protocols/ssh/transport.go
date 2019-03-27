package ssh

import (
	"fmt"
	"github.com/RokyErickson/doppelganger/pkg/process"
	"github.com/RokyErickson/doppelganger/pkg/url"
	"github.com/pkg/errors"
	"github.com/polydawn/gosh"
	"path/filepath"
)

type transport struct {
	remote   *url.URL
	prompter string
}

func (t *transport) Copy(localPath, remoteName string) error {

	if !filepath.IsAbs(localPath) {
		return errors.New("scp source path must be absolute")
	}
	workingDirectory, sourceBase := filepath.Split(localPath)

	destinationURL := fmt.Sprintf("%s:%s", t.remote.Hostname, remoteName)
	if t.remote.Username != "" {
		destinationURL = fmt.Sprintf("%s@%s", t.remote.Username, destinationURL)
	}

	var scpArguments []string
	scpArguments = append(scpArguments, "-C")
	scpArguments = append(scpArguments, "-oConnectTimeout=5")
	if t.remote.Port != 0 {
		scpArguments = append(scpArguments, "-P", fmt.Sprintf("%d", t.remote.Port))
	}
	environment := setPrompterVariables(t.prompter)

	scpArguments = append(scpArguments, sourceBase, destinationURL)

	scpProcess := gosh.Gosh("scp", scpArguments,
		gosh.Opts{
			Cwd:      workingDirectory,
			Env:      environment,
			Launcher: gosh.ExecCustomizingLauncher(process.DetachedProcessAttributes),
		},
	).Bake()

	scpProcess.Run()

	return nil
}

func (t *transport) Command(command string) gosh.Command {

	target := t.remote.Hostname
	if t.remote.Username != "" {
		target = fmt.Sprintf("%s@%s", t.remote.Username, t.remote.Hostname)
	}

	var sshArguments []string
	sshArguments = append(sshArguments, "-oConnectTimeout=5")
	if t.remote.Port != 0 {
		sshArguments = append(sshArguments, "-p", fmt.Sprintf("%d", t.remote.Port))
	}

	environment := setPrompterVariables(t.prompter)

	sshArguments = append(sshArguments, target, command)

	sshProcess := gosh.Gosh("ssh", sshArguments,
		gosh.Opts{
			Env:      environment,
			Launcher: gosh.ExecCustomizingLauncher(process.DetachedProcessAttributes),
		},
	).Bake()

	return sshProcess
}
