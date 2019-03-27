package main

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/dustin/go-humanize"

	"github.com/RokyErickson/doppelganger/cmd"
	fs "github.com/RokyErickson/doppelganger/pkg/filesystem"
	promptpkg "github.com/RokyErickson/doppelganger/pkg/prompt"
	sessionsvcpkg "github.com/RokyErickson/doppelganger/pkg/service/session"
	sessionpkg "github.com/RokyErickson/doppelganger/pkg/session"
	"github.com/RokyErickson/doppelganger/pkg/sync"
	"github.com/RokyErickson/doppelganger/pkg/url"
)

func createMain(command *cobra.Command, arguments []string) error {

	if len(arguments) != 2 {
		return errors.New("invalid number of endpoint URLs provided")
	}
	alpha, err := url.Parse(arguments[0], true)
	if err != nil {
		return errors.Wrap(err, "unable to parse alpha URL")
	}
	beta, err := url.Parse(arguments[1], false)
	if err != nil {
		return errors.Wrap(err, "unable to parse beta URL")
	}

	if alpha.Protocol == url.Protocol_Local {
		if alphaPath, err := fs.Normalize(alpha.Path); err != nil {
			return errors.Wrap(err, "unable to normalize alpha path")
		} else {
			alpha.Path = alphaPath
		}
	}
	if beta.Protocol == url.Protocol_Local {
		if betaPath, err := fs.Normalize(beta.Path); err != nil {
			return errors.Wrap(err, "unable to normalize beta path")
		} else {
			beta.Path = betaPath
		}
	}

	var synchronizationMode sync.SynchronizationMode
	if createConfiguration.synchronizationMode != "" {
		if err := synchronizationMode.UnmarshalText([]byte(createConfiguration.synchronizationMode)); err != nil {
			return errors.Wrap(err, "unable to parse synchronization mode")
		}
	}

	var maximumStagingFileSize uint64
	if createConfiguration.maximumStagingFileSize != "" {
		if s, err := humanize.ParseBytes(createConfiguration.maximumStagingFileSize); err != nil {
			return errors.Wrap(err, "unable to parse maximum staging file size")
		} else {
			maximumStagingFileSize = s
		}
	}

	var symbolicLinkMode sync.SymlinkMode
	if createConfiguration.symbolicLinkMode != "" {
		if err := symbolicLinkMode.UnmarshalText([]byte(createConfiguration.symbolicLinkMode)); err != nil {
			return errors.Wrap(err, "unable to parse symbolic link mode")
		}
	}

	var watchMode, watchModeAlpha, watchModeBeta fs.WatchMode
	if createConfiguration.watchMode != "" {
		if err := watchMode.UnmarshalText([]byte(createConfiguration.watchMode)); err != nil {
			return errors.Wrap(err, "unable to parse watch mode")
		}
	}
	if createConfiguration.watchModeAlpha != "" {
		if err := watchModeAlpha.UnmarshalText([]byte(createConfiguration.watchModeAlpha)); err != nil {
			return errors.Wrap(err, "unable to parse watch mode for alpha")
		}
	}
	if createConfiguration.watchModeBeta != "" {
		if err := watchModeBeta.UnmarshalText([]byte(createConfiguration.watchModeBeta)); err != nil {
			return errors.Wrap(err, "unable to parse watch mode for beta")
		}
	}

	for _, ignore := range createConfiguration.ignores {
		if !sync.ValidIgnorePattern(ignore) {
			return errors.Errorf("invalid ignore pattern: %s", ignore)
		}
	}

	var ignoreVCSMode sync.IgnoreVCSMode
	if createConfiguration.ignoreVCS && createConfiguration.noIgnoreVCS {
		return errors.New("conflicting VCS ignore behavior specified")
	} else if createConfiguration.ignoreVCS {
		ignoreVCSMode = sync.IgnoreVCSMode_IgnoreVCS
	} else if createConfiguration.noIgnoreVCS {
		ignoreVCSMode = sync.IgnoreVCSMode_PropagateVCS
	}

	var defaultFileMode, defaultFileModeAlpha, defaultFileModeBeta uint32
	if createConfiguration.defaultFileMode != "" {
		if m, err := fs.ParseMode(createConfiguration.defaultFileMode, fs.ModePermissionsMask); err != nil {
			return errors.Wrap(err, "unable to parse default file mode")
		} else if err = sync.EnsureDefaultFileModeValid(m); err != nil {
			return errors.Wrap(err, "invalid default file mode")
		} else {
			defaultFileMode = uint32(m)
		}
	}
	if createConfiguration.defaultFileModeAlpha != "" {
		if m, err := fs.ParseMode(createConfiguration.defaultFileModeAlpha, fs.ModePermissionsMask); err != nil {
			return errors.Wrap(err, "unable to parse default file mode for alpha")
		} else if err = sync.EnsureDefaultFileModeValid(m); err != nil {
			return errors.Wrap(err, "invalid default file mode for alpha")
		} else {
			defaultFileModeAlpha = uint32(m)
		}
	}
	if createConfiguration.defaultFileModeBeta != "" {
		if m, err := fs.ParseMode(createConfiguration.defaultFileModeBeta, fs.ModePermissionsMask); err != nil {
			return errors.Wrap(err, "unable to parse default file mode for beta")
		} else if err = sync.EnsureDefaultFileModeValid(m); err != nil {
			return errors.Wrap(err, "invalid default file mode for beta")
		} else {
			defaultFileModeBeta = uint32(m)
		}
	}

	var defaultDirectoryMode, defaultDirectoryModeAlpha, defaultDirectoryModeBeta uint32
	if createConfiguration.defaultDirectoryMode != "" {
		if m, err := fs.ParseMode(createConfiguration.defaultDirectoryMode, fs.ModePermissionsMask); err != nil {
			return errors.Wrap(err, "unable to parse default directory mode")
		} else if err = sync.EnsureDefaultDirectoryModeValid(m); err != nil {
			return errors.Wrap(err, "invalid default directory mode")
		} else {
			defaultDirectoryMode = uint32(m)
		}
	}
	if createConfiguration.defaultDirectoryModeAlpha != "" {
		if m, err := fs.ParseMode(createConfiguration.defaultDirectoryModeAlpha, fs.ModePermissionsMask); err != nil {
			return errors.Wrap(err, "unable to parse default directory mode for alpha")
		} else if err = sync.EnsureDefaultDirectoryModeValid(m); err != nil {
			return errors.Wrap(err, "invalid default directory mode for alpha")
		} else {
			defaultDirectoryModeAlpha = uint32(m)
		}
	}
	if createConfiguration.defaultDirectoryModeBeta != "" {
		if m, err := fs.ParseMode(createConfiguration.defaultDirectoryModeBeta, fs.ModePermissionsMask); err != nil {
			return errors.Wrap(err, "unable to parse default directory mode for beta")
		} else if err = sync.EnsureDefaultDirectoryModeValid(m); err != nil {
			return errors.Wrap(err, "invalid default directory mode for beta")
		} else {
			defaultDirectoryModeBeta = uint32(m)
		}
	}

	if createConfiguration.defaultOwner != "" {
		if kind, _ := fs.ParseOwnershipIdentifier(createConfiguration.defaultOwner); kind == fs.OwnershipIdentifierKindInvalid {
			return errors.New("invalid ownership specification")
		}
	}
	if createConfiguration.defaultOwnerAlpha != "" {
		if kind, _ := fs.ParseOwnershipIdentifier(createConfiguration.defaultOwnerAlpha); kind == fs.OwnershipIdentifierKindInvalid {
			return errors.New("invalid ownership specification for alpha")
		}
	}
	if createConfiguration.defaultOwnerBeta != "" {
		if kind, _ := fs.ParseOwnershipIdentifier(createConfiguration.defaultOwnerBeta); kind == fs.OwnershipIdentifierKindInvalid {
			return errors.New("invalid ownership specification for beta")
		}
	}

	if createConfiguration.defaultGroup != "" {
		if kind, _ := fs.ParseOwnershipIdentifier(createConfiguration.defaultGroup); kind == fs.OwnershipIdentifierKindInvalid {
			return errors.New("invalid group ownership specification")
		}
	}
	if createConfiguration.defaultGroupAlpha != "" {
		if kind, _ := fs.ParseOwnershipIdentifier(createConfiguration.defaultGroupAlpha); kind == fs.OwnershipIdentifierKindInvalid {
			return errors.New("invalid group ownership specification for alpha")
		}
	}
	if createConfiguration.defaultGroupBeta != "" {
		if kind, _ := fs.ParseOwnershipIdentifier(createConfiguration.defaultGroupBeta); kind == fs.OwnershipIdentifierKindInvalid {
			return errors.New("invalid group ownership specification for beta")
		}
	}

	daemonConnection, err := createDaemonClientConnection()
	if err != nil {
		return errors.Wrap(err, "unable to connect to daemon")
	}
	defer daemonConnection.Close()

	sessionService := sessionsvcpkg.NewSessionsClient(daemonConnection)

	createContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := sessionService.Create(createContext)
	if err != nil {
		return errors.Wrap(peelAwayRPCErrorLayer(err), "unable to invoke create")
	}

	request := &sessionsvcpkg.CreateRequest{
		Alpha: alpha,
		Beta:  beta,
		Configuration: &sessionpkg.Configuration{
			SynchronizationMode:    synchronizationMode,
			MaximumEntryCount:      createConfiguration.maximumEntryCount,
			MaximumStagingFileSize: maximumStagingFileSize,
			SymlinkMode:            symbolicLinkMode,
			WatchMode:              watchMode,
			WatchPollingInterval:   createConfiguration.watchPollingInterval,
			Ignores:                createConfiguration.ignores,
			IgnoreVCSMode:          ignoreVCSMode,
			DefaultFileMode:        defaultFileMode,
			DefaultDirectoryMode:   defaultDirectoryMode,
			DefaultOwner:           createConfiguration.defaultOwner,
			DefaultGroup:           createConfiguration.defaultGroup,
		},
		ConfigurationAlpha: &sessionpkg.Configuration{
			WatchMode:            watchModeAlpha,
			WatchPollingInterval: createConfiguration.watchPollingIntervalAlpha,
			DefaultFileMode:      defaultFileModeAlpha,
			DefaultDirectoryMode: defaultDirectoryModeAlpha,
			DefaultOwner:         createConfiguration.defaultOwnerAlpha,
			DefaultGroup:         createConfiguration.defaultGroupAlpha,
		},
		ConfigurationBeta: &sessionpkg.Configuration{
			WatchMode:            watchModeBeta,
			WatchPollingInterval: createConfiguration.watchPollingIntervalBeta,
			DefaultFileMode:      defaultFileModeBeta,
			DefaultDirectoryMode: defaultDirectoryModeBeta,
			DefaultOwner:         createConfiguration.defaultOwnerBeta,
			DefaultGroup:         createConfiguration.defaultGroupBeta,
		},
	}
	if err := stream.Send(request); err != nil {
		return errors.Wrap(peelAwayRPCErrorLayer(err), "unable to send create request")
	}

	statusLinePrinter := &cmd.StatusLinePrinter{}
	defer statusLinePrinter.BreakIfNonEmpty()

	for {
		if response, err := stream.Recv(); err != nil {
			return errors.Wrap(peelAwayRPCErrorLayer(err), "create failed")
		} else if err = response.EnsureValid(); err != nil {
			return errors.Wrap(err, "invalid create response received")
		} else if response.Session != "" {
			statusLinePrinter.Print(fmt.Sprintf("Created session %s", response.Session))
			return nil
		} else if response.Message != "" {
			statusLinePrinter.Print(response.Message)
			if err := stream.Send(&sessionsvcpkg.CreateRequest{}); err != nil {
				return errors.Wrap(peelAwayRPCErrorLayer(err), "unable to send message response")
			}
		} else if response.Prompt != "" {
			statusLinePrinter.BreakIfNonEmpty()
			if response, err := promptpkg.PromptCommandLine(response.Prompt); err != nil {
				return errors.Wrap(err, "unable to perform prompting")
			} else if err = stream.Send(&sessionsvcpkg.CreateRequest{Response: response}); err != nil {
				return errors.Wrap(peelAwayRPCErrorLayer(err), "unable to send prompt response")
			}
		}
	}
}

var createCommand = &cobra.Command{
	Use:   "create <alpha> <beta>",
	Short: "Creates and starts a new synchronization session",
	Run:   cmd.Mainify(createMain),
}

var createConfiguration struct {
	help                      bool
	synchronizationMode       string
	maximumEntryCount         uint64
	maximumStagingFileSize    string
	symbolicLinkMode          string
	watchMode                 string
	watchModeAlpha            string
	watchModeBeta             string
	watchPollingInterval      uint32
	watchPollingIntervalAlpha uint32
	watchPollingIntervalBeta  uint32
	ignores                   []string
	ignoreVCS                 bool
	noIgnoreVCS               bool
	defaultFileMode           string
	defaultFileModeAlpha      string
	defaultFileModeBeta       string
	defaultDirectoryMode      string
	defaultDirectoryModeAlpha string
	defaultDirectoryModeBeta  string
	defaultOwner              string
	defaultOwnerAlpha         string
	defaultOwnerBeta          string
	defaultGroup              string
	defaultGroupAlpha         string
	defaultGroupBeta          string
}

func init() {
	flags := createCommand.Flags()

	flags.BoolVarP(&createConfiguration.help, "help", "h", false, "Show help information")

	flags.StringVarP(&createConfiguration.synchronizationMode, "sync-mode", "m", "", "Specify synchronization mode (two-way-safe|two-way-resolved|one-way-safe|one-way-replica)")
	flags.Uint64Var(&createConfiguration.maximumEntryCount, "max-entry-count", 0, "Specify the maximum number of entries that endpoints will manage")
	flags.StringVar(&createConfiguration.maximumStagingFileSize, "max-staging-file-size", "", "Specify the maximum (individual) file size that endpoints will stage")

	flags.StringVar(&createConfiguration.symbolicLinkMode, "symlink-mode", "", "Specify symlink mode (ignore|portable|posix-raw)")

	flags.StringVar(&createConfiguration.watchMode, "watch-mode", "", "Specify watch mode (portable|force-poll|no-watch)")
	flags.StringVar(&createConfiguration.watchModeAlpha, "watch-mode-alpha", "", "Specify watch mode for alpha (portable|force-poll|no-watch)")
	flags.StringVar(&createConfiguration.watchModeBeta, "watch-mode-beta", "", "Specify watch mode for alpha (portable|force-poll|no-watch)")
	flags.Uint32Var(&createConfiguration.watchPollingInterval, "watch-polling-interval", 0, "Specify watch polling interval in seconds")
	flags.Uint32Var(&createConfiguration.watchPollingIntervalAlpha, "watch-polling-interval-alpha", 0, "Specify watch polling interval in seconds for alpha")
	flags.Uint32Var(&createConfiguration.watchPollingIntervalBeta, "watch-polling-interval-beta", 0, "Specify watch polling interval in seconds for beta")

	flags.StringSliceVarP(&createConfiguration.ignores, "ignore", "i", nil, "Specify ignore paths")
	flags.BoolVar(&createConfiguration.ignoreVCS, "ignore-vcs", false, "Ignore VCS directories")
	flags.BoolVar(&createConfiguration.noIgnoreVCS, "no-ignore-vcs", false, "Propagate VCS directories")

	flags.StringVar(&createConfiguration.defaultFileMode, "default-file-mode", "", "Specify default file permission mode")
	flags.StringVar(&createConfiguration.defaultFileModeAlpha, "default-file-mode-alpha", "", "Specify default file permission mode for alpha")
	flags.StringVar(&createConfiguration.defaultFileModeBeta, "default-file-mode-beta", "", "Specify default file permission mode for beta")
	flags.StringVar(&createConfiguration.defaultDirectoryMode, "default-directory-mode", "", "Specify default directory permission mode")
	flags.StringVar(&createConfiguration.defaultDirectoryModeAlpha, "default-directory-mode-alpha", "", "Specify default directory permission mode for alpha")
	flags.StringVar(&createConfiguration.defaultDirectoryModeBeta, "default-directory-mode-beta", "", "Specify default directory permission mode for beta")
	flags.StringVar(&createConfiguration.defaultOwner, "default-owner", "", "Specify default file/directory owner")
	flags.StringVar(&createConfiguration.defaultOwnerAlpha, "default-owner-alpha", "", "Specify default file/directory owner for alpha")
	flags.StringVar(&createConfiguration.defaultOwnerBeta, "default-owner-beta", "", "Specify default file/directory owner for beta")
	flags.StringVar(&createConfiguration.defaultGroup, "default-group", "", "Specify default file/directory group")
	flags.StringVar(&createConfiguration.defaultGroupAlpha, "default-group-alpha", "", "Specify default file/directory group for alpha")
	flags.StringVar(&createConfiguration.defaultGroupBeta, "default-group-beta", "", "Specify default file/directory group for beta")
}
