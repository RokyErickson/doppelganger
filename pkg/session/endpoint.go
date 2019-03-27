package session

import (
	"context"

	"github.com/RokyErickson/doppelganger/pkg/rsync"
	"github.com/RokyErickson/doppelganger/pkg/sync"
)

type Endpoint interface {
	Poll(context context.Context) error

	Scan(ancestor *sync.Entry) (*sync.Entry, bool, error, bool)

	Stage(paths []string, digests [][]byte) ([]string, []*rsync.Signature, rsync.Receiver, error)

	Supply(paths []string, signatures []*rsync.Signature, receiver rsync.Receiver) error

	Transition(transitions []*sync.Change) ([]*sync.Entry, []*sync.Problem, error)

	Shutdown() error
}
