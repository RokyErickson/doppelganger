// +build !windows

package daemon

import (
	"net"
	"os"
	"time"

	"github.com/pkg/errors"
)

const (
	socketName = "daemon.sock"
)

func DialTimeout(timeout time.Duration) (net.Conn, error) {

	socketPath, err := subpath(socketName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to compute socket path")
	}

	return net.DialTimeout("unix", socketPath, timeout)
}

func NewListener() (net.Listener, error) {

	socketPath, err := subpath(socketName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to compute socket path")
	}

	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrap(err, "unable to remove stale socket")
	}

	return net.Listen("unix", socketPath)
}
