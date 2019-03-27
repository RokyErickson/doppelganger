package main

import (
	"os"
	"os/signal"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"google.golang.org/grpc"

	"github.com/RokyErickson/doppelganger/cmd"
	"github.com/RokyErickson/doppelganger/pkg/daemon"
	mgrpc "github.com/RokyErickson/doppelganger/pkg/grpc"
	daemonsvc "github.com/RokyErickson/doppelganger/pkg/service/daemon"
	promptsvc "github.com/RokyErickson/doppelganger/pkg/service/prompt"
	sessionsvc "github.com/RokyErickson/doppelganger/pkg/service/session"
)

func daemonRunMain(command *cobra.Command, arguments []string) error {
	if len(arguments) != 0 {
		return errors.New("unexpected arguments provided")
	}

	lock, err := daemon.AcquireLock()
	if err != nil {
		return errors.Wrap(err, "unable to acquire daemon lock")
	}
	defer lock.Unlock()

	signalTermination := make(chan os.Signal, 1)
	signal.Notify(signalTermination, cmd.TerminationSignals...)

	server := grpc.NewServer(
		grpc.MaxSendMsgSize(mgrpc.MaximumIPCMessageSize),
		grpc.MaxRecvMsgSize(mgrpc.MaximumIPCMessageSize),
	)

	daemonServer := daemonsvc.New()
	daemonsvc.RegisterDaemonServer(server, daemonServer)
	defer daemonServer.Shutdown()

	promptsvc.RegisterPromptingServer(server, promptsvc.New())

	sessionsServer, err := sessionsvc.New()
	if err != nil {
		return errors.Wrap(err, "unable to create sessions service")
	}
	sessionsvc.RegisterSessionsServer(server, sessionsServer)
	defer sessionsServer.Shutdown()

	listener, err := daemon.NewListener()
	if err != nil {
		return errors.Wrap(err, "unable to create daemon listener")
	}
	defer listener.Close()

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.Serve(listener)
	}()

	select {
	case sig := <-signalTermination:
		return errors.Errorf("terminated by signal: %s", sig)
	case <-daemonServer.Termination:
		return nil
	case err = <-serverErrors:
		return errors.Wrap(err, "premature server termination")
	}
}

var daemonRunCommand = &cobra.Command{
	Use:    "run",
	Short:  "Runs the Doppelganger daemon",
	Run:    cmd.Mainify(daemonRunMain),
	Hidden: true,
}

var daemonRunConfiguration struct {
	help bool
}

func init() {
	flags := daemonRunCommand.Flags()
	flags.BoolVarP(&daemonRunConfiguration.help, "help", "h", false, "Show help information")
}
