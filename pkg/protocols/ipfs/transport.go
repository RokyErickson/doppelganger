package ipfs

import (
	"github.com/RokyErickson/doppelganger/pkg/filesystem"
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

func (t *transport) Command(command string) gosh.Command {

	ipfsArguments := []string{"ipfs-exec.sh --output", filesystem.DoppelgangerDirectoryName}

	environment := setIpfsVariables(t.remote)

	ipfsArguments = append(ipfsArguments, t.remote.Path)

	ipfsArguments = append(ipfsArguments, command)

	ipfsCommand := gosh.Gosh(ipfsArguments,
		gosh.Opts{
			Env:      environment,
			Launcher: gosh.ExecCustomizingLauncher(process.DetachedProcessAttributes),
		},
	).Bake()

	return ipfsCommand
}

func (t *transport) Copy(localPath, destination string) error {

	if !filepath.IsAbs(localPath) {
		return errors.New("Source must be absoulute")
	}
	workingDirectory, sourceBase := filepath.Split(localPath)

	environment := setIpfsVariables(t.remote)

	ipfscpArguments := []string{"ipfs-exec.sh --output", filesystem.DoppelgangerDirectoryName, t.remote.Path}

	ipfscpArguments = append(ipfscpArguments, "cp -R", sourceBase, destination)

	ipfscpProcess := gosh.Gosh(ipfscpArguments,
		gosh.Opts{
			Cwd:      workingDirectory,
			Env:      environment,
			Launcher: gosh.ExecCustomizingLauncher(process.DetachedProcessAttributes),
		},
	).Bake()

	ipfscpProcess.Run()

	return nil
}
