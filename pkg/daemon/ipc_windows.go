package daemon

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"time"

	"github.com/pkg/errors"

	"github.com/google/uuid"

	"github.com/Microsoft/go-winio"

	"github.com/RokyErickson/doppelganger/pkg/filesystem"
)

const (
	pipeNameRecordName = "daemon.pipe"
)

func DialTimeout(timeout time.Duration) (net.Conn, error) {

	pipeNameRecordPath, err := subpath(pipeNameRecordName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to compute pipe name record path")
	}

	pipeNameBytes, err := ioutil.ReadFile(pipeNameRecordPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read pipe name")
	}
	pipeName := string(pipeNameBytes)

	var timeoutPointer *time.Duration
	if timeout != 0 {
		timeoutPointer = &timeout
	}

	return winio.DialPipe(pipeName, timeoutPointer)
}

type daemonListener struct {
	net.Listener

	pipeNameRecordPath string
}

func (l *daemonListener) Close() error {

	if l.pipeNameRecordPath != "" {
		os.Remove(l.pipeNameRecordPath)
	}

	return l.Listener.Close()
}

func NewListener() (net.Listener, error) {

	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, errors.Wrap(err, "unable to generate UUID for named pipe")
	}
	pipeName := fmt.Sprintf(`\\.\pipe\doppelganger-%s`, randomUUID.String())

	pipeNameRecordPath, err := subpath(pipeNameRecordName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to compute pipe name record path")
	}

	user, err := user.Current()
	if err != nil {
		return nil, errors.Wrap(err, "unable to look up current user")
	}
	sid := user.Uid

	securityDescriptor := fmt.Sprintf("D:P(A;;GA;;;%s)", sid)

	configuration := &winio.PipeConfig{
		SecurityDescriptor: securityDescriptor,
	}

	rawListener, err := winio.ListenPipe(pipeName, configuration)
	if err != nil {
		return nil, err
	}
	listener := &daemonListener{rawListener, ""}

	if err = filesystem.WriteFileAtomic(pipeNameRecordPath, []byte(pipeName), 0600); err != nil {
		listener.Close()
		return nil, errors.Wrap(err, "unable to record pipe name")
	}
	listener.pipeNameRecordPath = pipeNameRecordPath

	return listener, nil
}
