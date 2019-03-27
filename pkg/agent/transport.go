package agent

import (
	"github.com/polydawn/gosh"
)

type Transport interface {
	Copy(localPath, remoteName string) error

	Command(command string) gosh.Command
}

func run(transport Transport, command string) {

	process := transport.Command(command)

	process.Run()
}

func output(transport Transport, command string) []byte {
	process := transport.Command(command)
	return []byte(process.Output())
}
