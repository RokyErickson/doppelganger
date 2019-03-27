package session

import (
	contextpkg "context"
	"fmt"
	"os"
	syncpkg "sync"
	"time"

	"github.com/pkg/errors"

	"github.com/google/uuid"

	"github.com/golang/protobuf/ptypes"

	"github.com/RokyErickson/doppelganger/pkg/doppelganger"
	"github.com/RokyErickson/doppelganger/pkg/encoding"
	"github.com/RokyErickson/doppelganger/pkg/filesystem"
	"github.com/RokyErickson/doppelganger/pkg/prompt"
	"github.com/RokyErickson/doppelganger/pkg/rsync"
	"github.com/RokyErickson/doppelganger/pkg/state"
	"github.com/RokyErickson/doppelganger/pkg/sync"
	"github.com/RokyErickson/doppelganger/pkg/url"
)

const (
	autoReconnectInterval = 30 * time.Second
	rescanWaitDuration    = 5 * time.Second
)

type controller struct {
	sessionPath              string
	archivePath              string
	stateLock                *state.TrackingLock
	session                  *Session
	mergedAlphaConfiguration *Configuration
	mergedBetaConfiguration  *Configuration
	state                    *State
	lifecycleLock            syncpkg.Mutex
	disabled                 bool
	cancel                   contextpkg.CancelFunc
	flushRequests            chan chan error
	done                     chan struct{}
}

func newSession(
	tracker *state.Tracker,
	alpha, beta *url.URL,
	configuration, configurationAlpha, configurationBeta *Configuration,
	prompter string,
) (*controller, error) {
	prompt.Message(prompter, "Creating session...")

	globalConfiguration, err := snapshotGlobalConfiguration()
	if err != nil {
		return nil, errors.Wrap(err, "unable to snapshot global configuration")
	}

	mergedConfiguration := MergeConfigurations(globalConfiguration, configuration)

	mergedAlphaConfiguration := MergeConfigurations(mergedConfiguration, configurationAlpha)
	mergedBetaConfiguration := MergeConfigurations(mergedConfiguration, configurationBeta)

	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, errors.Wrap(err, "unable to generate UUID for session")
	}
	identifier := randomUUID.String()

	version := Version_Version1

	creationTime := time.Now()
	creationTimeProto, err := ptypes.TimestampProto(creationTime)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert creation time format")
	}

	alphaEndpoint, err := connect(alpha, prompter, identifier, version, mergedAlphaConfiguration, true)
	if err != nil {
		return nil, errors.Wrap(err, "unable to connect to alpha")
	}
	betaEndpoint, err := connect(beta, prompter, identifier, version, mergedBetaConfiguration, false)
	if err != nil {
		alphaEndpoint.Shutdown()
		return nil, errors.Wrap(err, "unable to connect to beta")
	}

	session := &Session{
		Identifier:           identifier,
		Version:              version,
		CreationTime:         creationTimeProto,
		CreatingVersionMajor: doppelganger.VersionMajor,
		CreatingVersionMinor: doppelganger.VersionMinor,
		CreatingVersionPatch: doppelganger.VersionPatch,
		Alpha:                alpha,
		Beta:                 beta,
		Configuration:        mergedConfiguration,
		ConfigurationAlpha:   configurationAlpha,
		ConfigurationBeta:    configurationBeta,
	}
	archive := &sync.Archive{}

	sessionPath, err := pathForSession(session.Identifier)
	if err != nil {
		alphaEndpoint.Shutdown()
		betaEndpoint.Shutdown()
		return nil, errors.Wrap(err, "unable to compute session path")
	}
	archivePath, err := pathForArchive(session.Identifier)
	if err != nil {
		alphaEndpoint.Shutdown()
		betaEndpoint.Shutdown()
		return nil, errors.Wrap(err, "unable to compute archive path")
	}

	if err := encoding.MarshalAndSaveProtobuf(sessionPath, session); err != nil {
		alphaEndpoint.Shutdown()
		betaEndpoint.Shutdown()
		return nil, errors.Wrap(err, "unable to save session")
	}
	if err := encoding.MarshalAndSaveProtobuf(archivePath, archive); err != nil {
		os.Remove(sessionPath)
		alphaEndpoint.Shutdown()
		betaEndpoint.Shutdown()
		return nil, errors.Wrap(err, "unable to save archive")
	}
	controller := &controller{
		sessionPath:              sessionPath,
		archivePath:              archivePath,
		stateLock:                state.NewTrackingLock(tracker),
		session:                  session,
		mergedAlphaConfiguration: mergedAlphaConfiguration,
		mergedBetaConfiguration:  mergedBetaConfiguration,
		state: &State{
			Session: session,
		},
	}

	context, cancel := contextpkg.WithCancel(contextpkg.Background())
	controller.cancel = cancel
	controller.flushRequests = make(chan chan error, 1)
	controller.done = make(chan struct{})
	go controller.run(context, alphaEndpoint, betaEndpoint)

	return controller, nil
}

func loadSession(tracker *state.Tracker, identifier string) (*controller, error) {
	sessionPath, err := pathForSession(identifier)
	if err != nil {
		return nil, errors.Wrap(err, "unable to compute session path")
	}
	archivePath, err := pathForArchive(identifier)
	if err != nil {
		return nil, errors.Wrap(err, "unable to compute archive path")
	}

	session := &Session{}
	if err := encoding.LoadAndUnmarshalProtobuf(sessionPath, session); err != nil {
		return nil, errors.Wrap(err, "unable to load session configuration")
	}
	if session.ConfigurationAlpha == nil {
		session.ConfigurationAlpha = &Configuration{}
	}
	if session.ConfigurationBeta == nil {
		session.ConfigurationBeta = &Configuration{}
	}
	if err := session.EnsureValid(); err != nil {
		return nil, errors.Wrap(err, "invalid session found on disk")
	}

	controller := &controller{
		sessionPath: sessionPath,
		archivePath: archivePath,
		stateLock:   state.NewTrackingLock(tracker),
		session:     session,
		mergedAlphaConfiguration: MergeConfigurations(
			session.Configuration,
			session.ConfigurationAlpha,
		),
		mergedBetaConfiguration: MergeConfigurations(
			session.Configuration,
			session.ConfigurationBeta,
		),
		state: &State{
			Session: session,
		},
	}

	if !session.Paused {
		context, cancel := contextpkg.WithCancel(contextpkg.Background())
		controller.cancel = cancel
		controller.flushRequests = make(chan chan error, 1)
		controller.done = make(chan struct{})
		go controller.run(context, nil, nil)
	}

	return controller, nil
}

func (c *controller) currentState() *State {
	c.stateLock.Lock()
	defer c.stateLock.UnlockWithoutNotify()

	return c.state.Copy()
}

func (c *controller) flush(prompter string, skipWait bool, context contextpkg.Context) error {
	prompt.Message(prompter, fmt.Sprintf("Forcing synchronization cycle for session %s...", c.session.Identifier))

	c.lifecycleLock.Lock()
	defer c.lifecycleLock.Unlock()

	if c.disabled {
		return errors.New("controller disabled")
	}

	if c.cancel == nil {
		return errors.New("session is paused")
	}

	request := make(chan error, 1)

	if skipWait {
		select {
		case c.flushRequests <- request:
		default:
		}

		return nil
	}

	select {
	case c.flushRequests <- request:
	case <-context.Done():
		return errors.New("flush cancelled before request could be sent")
	}

	select {
	case err := <-request:
		if err != nil {
			return err
		}
	case <-context.Done():
		return errors.New("flush cancelled while waiting for synchronization cycle")
	}

	return nil
}

func (c *controller) resume(prompter string) error {

	prompt.Message(prompter, fmt.Sprintf("Resuming session %s...", c.session.Identifier))

	c.lifecycleLock.Lock()
	defer c.lifecycleLock.Unlock()

	if c.disabled {
		return errors.New("controller disabled")
	}

	if c.cancel != nil {

		c.stateLock.Lock()
		connected := c.state.Status >= Status_Watching
		c.stateLock.UnlockWithoutNotify()

		if connected {
			return nil
		}

		c.cancel()
		<-c.done

		c.cancel = nil
		c.flushRequests = nil
		c.done = nil
	}

	c.stateLock.Lock()
	c.session.Paused = false
	saveErr := encoding.MarshalAndSaveProtobuf(c.sessionPath, c.session)
	c.stateLock.Unlock()

	c.stateLock.Lock()
	c.state.Status = Status_ConnectingAlpha
	c.stateLock.Unlock()
	alpha, alphaConnectErr := connect(
		c.session.Alpha,
		prompter,
		c.session.Identifier,
		c.session.Version,
		c.mergedAlphaConfiguration,
		true,
	)
	c.stateLock.Lock()
	c.state.AlphaConnected = (alpha != nil)
	c.stateLock.Unlock()

	c.stateLock.Lock()
	c.state.Status = Status_ConnectingBeta
	c.stateLock.Unlock()
	beta, betaConnectErr := connect(
		c.session.Beta,
		prompter,
		c.session.Identifier,
		c.session.Version,
		c.mergedBetaConfiguration,
		false,
	)
	c.stateLock.Lock()
	c.state.BetaConnected = (beta != nil)
	c.stateLock.Unlock()

	context, cancel := contextpkg.WithCancel(contextpkg.Background())
	c.cancel = cancel
	c.flushRequests = make(chan chan error, 1)
	c.done = make(chan struct{})
	go c.run(context, alpha, beta)

	if saveErr != nil {
		return errors.Wrap(saveErr, "unable to save session configuration")
	} else if alphaConnectErr != nil {
		return errors.Wrap(alphaConnectErr, "unable to connect to alpha")
	} else if betaConnectErr != nil {
		return errors.Wrap(betaConnectErr, "unable to connect to beta")
	}

	return nil
}

type haltMode uint8

const (
	haltModePause haltMode = iota
	haltModeShutdown
	haltModeTerminate
)

func (m haltMode) description() string {
	switch m {
	case haltModePause:
		return "Pausing"
	case haltModeShutdown:
		return "Shutting down"
	case haltModeTerminate:
		return "Terminating"
	default:
		panic("unhandled halt mode")
	}
}

func (c *controller) halt(mode haltMode, prompter string) error {
	prompt.Message(prompter, fmt.Sprintf("%s session %s...", mode.description(), c.session.Identifier))
	c.lifecycleLock.Lock()
	defer c.lifecycleLock.Unlock()
	if c.disabled {
		return errors.New("controller disabled")
	}
	if c.cancel != nil {
		c.cancel()
		<-c.done
		c.cancel = nil
		c.flushRequests = nil
		c.done = nil
	}

	if mode == haltModePause {
		c.stateLock.Lock()
		c.session.Paused = true
		err := encoding.MarshalAndSaveProtobuf(c.sessionPath, c.session)
		c.stateLock.Unlock()
		if err != nil {
			return errors.Wrap(err, "unable to save session state")
		}
	} else if mode == haltModeShutdown {
		c.disabled = true
	} else if mode == haltModeTerminate {
		c.disabled = true
		sessionRemoveErr := os.Remove(c.sessionPath)
		archiveRemoveErr := os.Remove(c.archivePath)
		if sessionRemoveErr != nil {
			return errors.Wrap(sessionRemoveErr, "unable to remove session from disk")
		} else if archiveRemoveErr != nil {
			return errors.Wrap(archiveRemoveErr, "unable to remove archive from disk")
		}
	} else {
		panic("invalid halt mode specified")
	}
	return nil
}

func (c *controller) run(context contextpkg.Context, alpha, beta Endpoint) {
	defer func() {
		if alpha != nil {
			alpha.Shutdown()
		}
		if beta != nil {
			beta.Shutdown()
		}

		c.stateLock.Lock()
		c.state = &State{
			Session: c.session,
		}
		c.stateLock.Unlock()

		close(c.done)
	}()

	for {

		for {
			if alpha == nil {
				c.stateLock.Lock()
				c.state.Status = Status_ConnectingAlpha
				c.stateLock.Unlock()
				alpha, _ = reconnect(
					context,
					c.session.Alpha,
					c.session.Identifier,
					c.session.Version,
					c.mergedAlphaConfiguration,
					true,
				)
			}
			c.stateLock.Lock()
			c.state.AlphaConnected = (alpha != nil)
			c.stateLock.Unlock()

			select {
			case <-context.Done():
				return
			default:
			}

			if beta == nil {
				c.stateLock.Lock()
				c.state.Status = Status_ConnectingBeta
				c.stateLock.Unlock()
				beta, _ = reconnect(
					context,
					c.session.Beta,
					c.session.Identifier,
					c.session.Version,
					c.mergedBetaConfiguration,
					false,
				)
			}
			c.stateLock.Lock()
			c.state.BetaConnected = (beta != nil)
			c.stateLock.Unlock()

			if alpha != nil && beta != nil {
				break
			}

			select {
			case <-context.Done():
				return
			case <-time.After(autoReconnectInterval):
			}
		}

		err := c.synchronize(context, alpha, beta)

		alpha.Shutdown()
		alpha = nil
		beta.Shutdown()
		beta = nil

		c.stateLock.Lock()
		c.state = &State{
			Session:   c.session,
			LastError: err.Error(),
		}
		c.stateLock.Unlock()

		select {
		case <-context.Done():
			return
		case <-time.After(autoReconnectInterval):
		}
	}
}

func (c *controller) synchronize(context contextpkg.Context, alpha, beta Endpoint) error {
	c.stateLock.Lock()
	if c.state.LastError != "" {
		c.state.LastError = ""
		c.stateLock.Unlock()
	} else {
		c.stateLock.UnlockWithoutNotify()
	}

	var flushRequest chan error
	defer func() {
		if flushRequest != nil {
			flushRequest <- errors.New("synchronization cycle failed")
			flushRequest = nil
		}
	}()

	archive := &sync.Archive{}
	if err := encoding.LoadAndUnmarshalProtobuf(c.archivePath, archive); err != nil {
		return errors.Wrap(err, "unable to load archive")
	} else if err = archive.Root.EnsureValid(); err != nil {
		return errors.Wrap(err, "invalid archive found on disk")
	}
	ancestor := archive.Root

	synchronizationMode := c.session.Configuration.SynchronizationMode
	if synchronizationMode.IsDefault() {
		synchronizationMode = c.session.Version.DefaultSynchronizationMode()
	}

	αWatchMode := c.mergedAlphaConfiguration.WatchMode
	βWatchMode := c.mergedBetaConfiguration.WatchMode
	if αWatchMode.IsDefault() {
		αWatchMode = c.session.Version.DefaultWatchMode()
	}
	if βWatchMode.IsDefault() {
		βWatchMode = c.session.Version.DefaultWatchMode()
	}
	αDisablePolling := (αWatchMode == filesystem.WatchMode_WatchModeNoWatch)
	βDisablePolling := (βWatchMode == filesystem.WatchMode_WatchModeNoWatch)

	skipPolling := (!αDisablePolling || !βDisablePolling)

	for {
		if !skipPolling {
			c.stateLock.Lock()
			c.state.Status = Status_Watching
			c.stateLock.Unlock()

			pollContext, pollCancel := contextpkg.WithCancel(contextpkg.Background())
			αPollResults := make(chan error, 1)
			go func() {
				if αDisablePolling {
					<-pollContext.Done()
					αPollResults <- nil
				} else {
					αPollResults <- alpha.Poll(pollContext)
				}
			}()

			βPollResults := make(chan error, 1)
			go func() {
				if βDisablePolling {
					<-pollContext.Done()
					βPollResults <- nil
				} else {
					βPollResults <- beta.Poll(pollContext)
				}
			}()

			var αPollErr, βPollErr error
			cancelled := false
			select {
			case αPollErr = <-αPollResults:
				pollCancel()
				βPollErr = <-βPollResults
			case βPollErr = <-βPollResults:
				pollCancel()
				αPollErr = <-αPollResults
			case flushRequest = <-c.flushRequests:
				if cap(flushRequest) < 1 {
					panic("unbuffered flush request")
				}
				pollCancel()
				αPollErr = <-αPollResults
				βPollErr = <-βPollResults
			case <-context.Done():
				cancelled = true
				pollCancel()
				αPollErr = <-αPollResults
				βPollErr = <-βPollResults
			}

			if cancelled {
				return errors.New("cancelled during polling")
			} else if αPollErr != nil {
				return errors.Wrap(αPollErr, "alpha polling error")
			} else if βPollErr != nil {
				return errors.Wrap(βPollErr, "beta polling error")
			}
		} else {
			skipPolling = false
		}

		c.stateLock.Lock()
		c.state.Status = Status_Scanning
		c.stateLock.Unlock()
		var αSnapshot, βSnapshot *sync.Entry
		var αPreservesExecutability, βPreservesExecutability bool
		var αScanErr, βScanErr error
		var αTryAgain, βTryAgain bool
		scanDone := &syncpkg.WaitGroup{}
		scanDone.Add(2)
		go func() {
			αSnapshot, αPreservesExecutability, αScanErr, αTryAgain = alpha.Scan(ancestor)
			scanDone.Done()
		}()
		go func() {
			βSnapshot, βPreservesExecutability, βScanErr, βTryAgain = beta.Scan(ancestor)
			scanDone.Done()
		}()
		scanDone.Wait()
		if αScanErr != nil {
			αScanErr = errors.Wrap(αScanErr, "alpha scan error")
			if !αTryAgain {
				return αScanErr
			} else {
				c.stateLock.Lock()
				c.state.LastError = αScanErr.Error()
				c.stateLock.Unlock()
			}
		}
		if βScanErr != nil {
			βScanErr = errors.Wrap(βScanErr, "beta scan error")
			if !βTryAgain {
				return βScanErr
			} else {
				c.stateLock.Lock()
				c.state.LastError = βScanErr.Error()
				c.stateLock.Unlock()
			}
		}

		if αTryAgain || βTryAgain {
			c.stateLock.Lock()
			c.state.Status = Status_WaitingForRescan
			c.stateLock.Unlock()

			select {
			case <-time.After(rescanWaitDuration):
			case <-context.Done():
				return errors.New("cancelled during rescan wait")
			}

			skipPolling = true
			continue
		}

		c.stateLock.Lock()
		if c.state.LastError != "" {
			c.state.LastError = ""
			c.stateLock.Unlock()
		} else {
			c.stateLock.UnlockWithoutNotify()
		}

		if αPreservesExecutability && !βPreservesExecutability {
			βSnapshot = sync.PropagateExecutability(ancestor, αSnapshot, βSnapshot)
		} else if βPreservesExecutability && !αPreservesExecutability {
			αSnapshot = sync.PropagateExecutability(ancestor, βSnapshot, αSnapshot)
		}

		c.stateLock.Lock()
		c.state.Status = Status_Reconciling
		c.stateLock.Unlock()

		ancestorChanges, αTransitions, βTransitions, conflicts := sync.Reconcile(
			ancestor,
			αSnapshot,
			βSnapshot,
			synchronizationMode,
		)

		var slimConflicts []*sync.Conflict
		if len(conflicts) > 0 {
			slimConflicts = make([]*sync.Conflict, len(conflicts))
			for c, conflict := range conflicts {
				slimConflicts[c] = conflict.CopySlim()
			}
		}
		c.stateLock.Lock()
		c.state.Conflicts = slimConflicts
		c.stateLock.Unlock()

		rootDeletion := false
		for _, t := range αTransitions {
			if isRootDeletion(t) {
				rootDeletion = true
				break
			}
		}
		if !rootDeletion {
			for _, t := range βTransitions {
				if isRootDeletion(t) {
					rootDeletion = true
					break
				}
			}
		}
		if rootDeletion {
			c.stateLock.Lock()
			c.state.Status = Status_HaltedOnRootDeletion
			c.stateLock.Unlock()
			<-context.Done()
			return errors.New("cancelled while halted on root deletion")
		}

		rootTypeChange := false
		for _, t := range αTransitions {
			if isRootTypeChange(t) {
				rootTypeChange = true
				break
			}
		}
		if !rootTypeChange {
			for _, t := range βTransitions {
				if isRootTypeChange(t) {
					rootTypeChange = true
					break
				}
			}
		}
		if rootTypeChange {
			c.stateLock.Lock()
			c.state.Status = Status_HaltedOnRootTypeChange
			c.stateLock.Unlock()
			<-context.Done()
			return errors.New("cancelled while halted on root type change")
		}

		monitor := func(status *rsync.ReceiverStatus) error {
			c.stateLock.Lock()
			c.state.StagingStatus = status
			c.stateLock.Unlock()
			return nil
		}

		c.stateLock.Lock()
		c.state.Status = Status_StagingAlpha
		c.stateLock.Unlock()
		if paths, digests, err := sync.TransitionDependencies(αTransitions); err != nil {
			return errors.Wrap(err, "unable to determine paths for staging on alpha")
		} else if len(paths) > 0 {
			filteredPaths, signatures, receiver, err := alpha.Stage(paths, digests)
			if err != nil {
				return errors.Wrap(err, "unable to begin staging on alpha")
			}
			if !filteredPathsAreSubset(filteredPaths, paths) {
				return errors.New("alpha returned incorrect subset of staging paths")
			}
			if len(filteredPaths) > 0 {
				receiver = rsync.NewMonitoringReceiver(receiver, filteredPaths, monitor)
				receiver = rsync.NewPreemptableReceiver(receiver, context)
				if err = beta.Supply(filteredPaths, signatures, receiver); err != nil {
					return errors.Wrap(err, "unable to stage files on alpha")
				}
			}
		}

		c.stateLock.Lock()
		c.state.Status = Status_StagingBeta
		c.stateLock.Unlock()
		if paths, digests, err := sync.TransitionDependencies(βTransitions); err != nil {
			return errors.Wrap(err, "unable to determine paths for staging on beta")
		} else if len(paths) > 0 {
			filteredPaths, signatures, receiver, err := beta.Stage(paths, digests)
			if err != nil {
				return errors.Wrap(err, "unable to begin staging on beta")
			}
			if !filteredPathsAreSubset(filteredPaths, paths) {
				return errors.New("beta returned incorrect subset of staging paths")
			}
			if len(filteredPaths) > 0 {
				receiver = rsync.NewMonitoringReceiver(receiver, filteredPaths, monitor)
				receiver = rsync.NewPreemptableReceiver(receiver, context)
				if err = alpha.Supply(filteredPaths, signatures, receiver); err != nil {
					return errors.Wrap(err, "unable to stage files on beta")
				}
			}
		}

		c.stateLock.Lock()
		c.state.Status = Status_Transitioning
		c.stateLock.Unlock()
		var αResults, βResults []*sync.Entry
		var αProblems, βProblems []*sync.Problem
		var αTransitionErr, βTransitionErr error
		var αChanges, βChanges []*sync.Change
		transitionDone := &syncpkg.WaitGroup{}
		if len(αTransitions) > 0 {
			transitionDone.Add(1)
		}
		if len(βTransitions) > 0 {
			transitionDone.Add(1)
		}
		if len(αTransitions) > 0 {
			go func() {
				αResults, αProblems, αTransitionErr = alpha.Transition(αTransitions)
				if αTransitionErr == nil {
					for t, transition := range αTransitions {
						αChanges = append(αChanges, &sync.Change{Path: transition.Path, New: αResults[t]})
					}
				}
				transitionDone.Done()
			}()
		}
		if len(βTransitions) > 0 {
			go func() {
				βResults, βProblems, βTransitionErr = beta.Transition(βTransitions)
				if βTransitionErr == nil {
					for t, transition := range βTransitions {
						βChanges = append(βChanges, &sync.Change{Path: transition.Path, New: βResults[t]})
					}
				}
				transitionDone.Done()
			}()
		}
		transitionDone.Wait()

		c.stateLock.Lock()
		c.state.Status = Status_Saving
		c.state.AlphaProblems = αProblems
		c.state.BetaProblems = βProblems
		c.stateLock.Unlock()
		ancestorChanges = append(ancestorChanges, αChanges...)
		ancestorChanges = append(ancestorChanges, βChanges...)
		if newAncestor, err := sync.Apply(ancestor, ancestorChanges); err != nil {
			return errors.Wrap(err, "unable to propagate changes to ancestor")
		} else {
			ancestor = newAncestor
		}

		if err := ancestor.EnsureValid(); err != nil {
			return errors.Wrap(err, "new ancestor is invalid")
		}

		archive.Root = ancestor
		if err := encoding.MarshalAndSaveProtobuf(c.archivePath, archive); err != nil {
			return errors.Wrap(err, "unable to save ancestor")
		}

		if αTransitionErr != nil {
			return errors.Wrap(αTransitionErr, "unable to apply changes to alpha")
		} else if βTransitionErr != nil {
			return errors.Wrap(βTransitionErr, "unable to apply changes to beta")
		}

		c.stateLock.Lock()
		c.state.SuccessfulSynchronizationCycles++
		c.stateLock.Unlock()

		if flushRequest != nil {
			flushRequest <- nil
			flushRequest = nil
		}
	}
}
