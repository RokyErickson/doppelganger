package daemon

import (
	"context"
	"time"

	"github.com/RokyErickson/doppelganger/pkg/agent"
	"github.com/RokyErickson/doppelganger/pkg/doppelganger"
	"github.com/RokyErickson/doppelganger/pkg/protocols/local"
)

const (
	housekeepingInterval = 24 * time.Hour
)

func housekeep() {
	agent.Housekeep()
	local.HousekeepCaches()
	local.HousekeepStaging()
}

type Server struct {
	Termination chan struct{}
	context     context.Context
	shutdown    context.CancelFunc
}

func New() *Server {
	context, shutdown := context.WithCancel(context.Background())
	server := &Server{
		Termination: make(chan struct{}, 1),
		context:     context,
		shutdown:    shutdown,
	}

	go server.housekeep()

	return server
}

func (s *Server) housekeep() {
	housekeep()

	ticker := time.NewTicker(housekeepingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.context.Done():
			return
		case <-ticker.C:
			housekeep()
		}
	}
}

func (s *Server) Shutdown() {
	s.shutdown()
}

func (s *Server) Version(_ context.Context, _ *VersionRequest) (*VersionResponse, error) {

	return &VersionResponse{
		Major: doppelganger.VersionMajor,
		Minor: doppelganger.VersionMinor,
		Patch: doppelganger.VersionPatch,
	}, nil
}

func (s *Server) Terminate(_ context.Context, _ *TerminateRequest) (*TerminateResponse, error) {
	select {
	case s.Termination <- struct{}{}:
	default:
	}

	return &TerminateResponse{}, nil
}
