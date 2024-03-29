package main

import (
	"context"
	"net"
	"time"

	"github.com/pkg/errors"

	"google.golang.org/grpc"
	grpcstatus "google.golang.org/grpc/status"

	"github.com/RokyErickson/doppelganger/pkg/daemon"
	mgrpc "github.com/RokyErickson/doppelganger/pkg/grpc"
)

func daemonDialer(_ string, timeout time.Duration) (net.Conn, error) {
	return daemon.DialTimeout(timeout)
}

func createDaemonClientConnection() (*grpc.ClientConn, error) {

	dialContext, cancel := context.WithTimeout(
		context.Background(),
		daemon.RecommendedDialTimeout,
	)
	defer cancel()

	return grpc.DialContext(
		dialContext,
		"",
		grpc.WithInsecure(),
		grpc.WithDialer(daemonDialer),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(mgrpc.MaximumIPCMessageSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(mgrpc.MaximumIPCMessageSize)),
	)
}

func peelAwayRPCErrorLayer(err error) error {
	if status, ok := grpcstatus.FromError(err); !ok {
		return err
	} else {
		return errors.New(status.Message())
	}
}
