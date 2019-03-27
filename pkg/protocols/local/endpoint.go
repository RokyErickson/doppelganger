package local

import (
	"context"
	"hash"
	"io"
	syncpkg "sync"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/encoding"
	"github.com/RokyErickson/doppelganger/pkg/filesystem"
	"github.com/RokyErickson/doppelganger/pkg/rsync"
	"github.com/RokyErickson/doppelganger/pkg/session"
	"github.com/RokyErickson/doppelganger/pkg/sync"
)

type endpoint struct {
	root                           string
	readOnly                       bool
	maximumEntryCount              uint64
	watchCancel                    context.CancelFunc
	watchEvents                    chan struct{}
	symlinkMode                    sync.SymlinkMode
	ignores                        []string
	defaultFileMode                filesystem.Mode
	defaultDirectoryMode           filesystem.Mode
	defaultOwnership               *filesystem.OwnershipSpecification
	cachePath                      string
	cacheLock                      syncpkg.Mutex
	cacheWriteError                error
	cache                          *sync.Cache
	ignoreCache                    sync.IgnoreCache
	recomposeUnicode               bool
	scanHasher                     hash.Hash
	stager                         *stager
	lastScanCount                  uint64
	scannedSinceLastStageCall      bool
	scannedSinceLastTransitionCall bool
}

func NewEndpoint(
	root,
	sessionIdentifier string,
	version session.Version,
	configuration *session.Configuration,
	alpha bool,
	options ...EndpointOption,
) (session.Endpoint, error) {
	root, err := filesystem.Normalize(root)
	if err != nil {
		return nil, errors.Wrap(err, "unable to normalize root path")
	}

	endpointOptions := &endpointOptions{}
	for _, o := range options {
		o.apply(endpointOptions)
	}

	synchronizationMode := configuration.SynchronizationMode
	if synchronizationMode.IsDefault() {
		synchronizationMode = version.DefaultSynchronizationMode()
	}
	unidirectional := synchronizationMode == sync.SynchronizationMode_SynchronizationModeOneWaySafe ||
		synchronizationMode == sync.SynchronizationMode_SynchronizationModeOneWayReplica
	readOnly := alpha && unidirectional

	symlinkMode := configuration.SymlinkMode
	if symlinkMode.IsDefault() {
		symlinkMode = version.DefaultSymlinkMode()
	}

	watchMode := configuration.WatchMode
	if watchMode.IsDefault() {
		watchMode = version.DefaultWatchMode()
	}

	watchPollingInterval := configuration.WatchPollingInterval
	if watchPollingInterval == 0 {
		watchPollingInterval = version.DefaultWatchPollingInterval()
	}

	ignoreVCSMode := configuration.IgnoreVCSMode
	if ignoreVCSMode.IsDefault() {
		ignoreVCSMode = version.DefaultIgnoreVCSMode()
	}

	defaultFileMode := filesystem.Mode(configuration.DefaultFileMode)
	if defaultFileMode == 0 {
		defaultFileMode = version.DefaultFileMode()
	}

	defaultDirectoryMode := filesystem.Mode(configuration.DefaultDirectoryMode)
	if defaultDirectoryMode == 0 {
		defaultDirectoryMode = version.DefaultDirectoryMode()
	}

	defaultOwnerSpecification := configuration.DefaultOwner
	if defaultOwnerSpecification == "" {
		defaultOwnerSpecification = version.DefaultOwnerSpecification()
	}

	defaultGroupSpecification := configuration.DefaultGroup
	if defaultGroupSpecification == "" {
		defaultGroupSpecification = version.DefaultGroupSpecification()
	}

	defaultOwnership, err := filesystem.NewOwnershipSpecification(
		defaultOwnerSpecification,
		defaultGroupSpecification,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create ownership specification")
	}

	var ignores []string
	if ignoreVCSMode == sync.IgnoreVCSMode_IgnoreVCS {
		ignores = append(ignores, sync.DefaultVCSIgnores...)
	}
	ignores = append(ignores, configuration.DefaultIgnores...)
	ignores = append(ignores, configuration.Ignores...)

	watchContext, watchCancel := context.WithCancel(context.Background())
	watchEvents := make(chan struct{}, 1)
	if endpointOptions.watchingMechanism != nil {
		go endpointOptions.watchingMechanism(watchContext, root, watchEvents)
	} else {
		go filesystem.Watch(
			watchContext,
			root,
			watchEvents,
			watchMode,
			watchPollingInterval,
		)
	}

	var cachePath string
	if endpointOptions.cachePathCallback != nil {
		cachePath, err = endpointOptions.cachePathCallback(sessionIdentifier, alpha)
	} else {
		cachePath, err = pathForCache(sessionIdentifier, alpha)
	}
	if err != nil {
		watchCancel()
		return nil, errors.Wrap(err, "unable to compute/create cache path")
	}

	cache := &sync.Cache{}
	if encoding.LoadAndUnmarshalProtobuf(cachePath, cache) != nil {
		cache = &sync.Cache{}
	} else if cache.EnsureValid() != nil {
		cache = &sync.Cache{}
	}

	var stagingRoot string
	if endpointOptions.stagingRootCallback != nil {
		stagingRoot, err = endpointOptions.stagingRootCallback(sessionIdentifier, alpha)
	} else {
		stagingRoot, err = pathForStagingRoot(sessionIdentifier, alpha)
	}
	if err != nil {
		watchCancel()
		return nil, errors.Wrap(err, "unable to compute staging root")
	}

	return &endpoint{
		root:                 root,
		readOnly:             readOnly,
		maximumEntryCount:    configuration.MaximumEntryCount,
		watchCancel:          watchCancel,
		watchEvents:          watchEvents,
		symlinkMode:          symlinkMode,
		ignores:              ignores,
		defaultFileMode:      defaultFileMode,
		defaultDirectoryMode: defaultDirectoryMode,
		defaultOwnership:     defaultOwnership,
		cachePath:            cachePath,
		cache:                cache,
		scanHasher:           version.Hasher(),
		stager:               newStager(version, stagingRoot, configuration.MaximumStagingFileSize),
	}, nil
}

func (e *endpoint) Poll(context context.Context) error {

	select {
	case _, ok := <-e.watchEvents:
		if !ok {
			return errors.New("endpoint watcher terminated")
		}
	case <-context.Done():
	}

	// Done.
	return nil
}

func (e *endpoint) Scan(_ *sync.Entry) (*sync.Entry, bool, error, bool) {

	e.cacheLock.Lock()

	if e.cacheWriteError != nil {
		defer e.cacheLock.Unlock()
		return nil, false, errors.Wrap(e.cacheWriteError, "unable to save cache to disk"), false
	}

	result, preservesExecutability, recomposeUnicode, newCache, newIgnoreCache, err := sync.Scan(
		e.root, e.scanHasher, e.cache, e.ignores, e.ignoreCache, e.symlinkMode,
	)
	if err != nil {
		e.cacheLock.Unlock()
		return nil, false, err, true
	}

	e.lastScanCount = result.Count()

	e.scannedSinceLastStageCall = true
	e.scannedSinceLastTransitionCall = true

	if e.maximumEntryCount != 0 && e.lastScanCount > e.maximumEntryCount {
		e.cacheLock.Unlock()
		return nil, false, errors.New("exceeded allowed entry count"), true
	}

	e.cache = newCache
	e.ignoreCache = newIgnoreCache
	e.recomposeUnicode = recomposeUnicode

	go func() {
		if err := encoding.MarshalAndSaveProtobuf(e.cachePath, e.cache); err != nil {
			e.cacheWriteError = err
		}
		e.cacheLock.Unlock()
	}()

	return result, preservesExecutability, nil, false
}

func (e *endpoint) stageFromRoot(
	path string,
	digest []byte,
	reverseLookupMap *sync.ReverseLookupMap,
	opener *filesystem.Opener,
) bool {

	sourcePath, sourcePathOk := reverseLookupMap.Lookup(digest)
	if !sourcePathOk {
		return false
	}

	source, err := opener.Open(sourcePath)
	if err != nil {
		return false
	}
	defer source.Close()

	sink, err := e.stager.Sink(path)
	if err != nil {
		return false
	}

	_, err = io.Copy(sink, source)
	sink.Close()
	if err != nil {
		return false
	}

	_, err = e.stager.Provide(path, digest)
	return err == nil
}

func (e *endpoint) Stage(paths []string, digests [][]byte) ([]string, []*rsync.Signature, rsync.Receiver, error) {

	if e.readOnly {
		return nil, nil, nil, errors.New("endpoint is in read-only mode")
	}

	if !e.scannedSinceLastStageCall {
		return nil, nil, nil, errors.New("multiple staging operations performed without scan")
	}
	e.scannedSinceLastStageCall = false

	if e.maximumEntryCount != 0 && (e.maximumEntryCount-e.lastScanCount) < uint64(len(paths)) {
		return nil, nil, nil, errors.New("staging would exceeded allowed entry count")
	}

	e.cacheLock.Lock()
	reverseLookupMap, err := e.cache.GenerateReverseLookupMap()
	e.cacheLock.Unlock()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "unable to generate reverse lookup map")
	}

	opener := filesystem.NewOpener(e.root)
	defer opener.Close()

	filteredPaths := paths[:0]
	for p, path := range paths {
		digest := digests[p]
		if _, err := e.stager.Provide(path, digest); err == nil {
			continue
		} else if e.stageFromRoot(path, digest, reverseLookupMap, opener) {
			continue
		} else {
			filteredPaths = append(filteredPaths, path)
		}
	}
	if len(filteredPaths) == 0 {
		return nil, nil, nil, nil
	}

	engine := rsync.NewEngine()

	signatures := make([]*rsync.Signature, len(filteredPaths))
	for p, path := range filteredPaths {
		if base, err := opener.Open(path); err != nil {
			signatures[p] = &rsync.Signature{}
			continue
		} else if signature, err := engine.Signature(base, 0); err != nil {
			base.Close()
			signatures[p] = &rsync.Signature{}
			continue
		} else {
			base.Close()
			signatures[p] = signature
		}
	}

	receiver, err := rsync.NewReceiver(e.root, filteredPaths, signatures, e.stager)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "unable to create rsync receiver")
	}

	return filteredPaths, signatures, receiver, nil
}

func (e *endpoint) Supply(paths []string, signatures []*rsync.Signature, receiver rsync.Receiver) error {
	return rsync.Transmit(e.root, paths, signatures, receiver)
}

func (e *endpoint) Transition(transitions []*sync.Change) ([]*sync.Entry, []*sync.Problem, error) {

	if e.readOnly {
		return nil, nil, errors.New("endpoint is in read-only mode")
	}

	if !e.scannedSinceLastTransitionCall {
		return nil, nil, errors.New("multiple transition operations performed without scan")
	}
	e.scannedSinceLastTransitionCall = false

	if e.maximumEntryCount != 0 {

		resultingEntryCount := e.lastScanCount
		for _, transition := range transitions {
			if removed := transition.Old.Count(); removed > resultingEntryCount {
				return nil, nil, errors.New("transition requires removing more entries than exist")
			} else {
				resultingEntryCount -= removed
			}
			resultingEntryCount += transition.New.Count()
		}

		results := make([]*sync.Entry, len(transitions))
		for t, transition := range transitions {
			results[t] = transition.Old
		}
		problems := []*sync.Problem{{Error: "transitioning would exceeded allowed entry count"}}
		if e.maximumEntryCount < resultingEntryCount {
			return results, problems, nil
		}
	}

	e.cacheLock.Lock()
	defer e.cacheLock.Unlock()

	results, problems := sync.Transition(
		e.root,
		transitions,
		e.cache,
		e.symlinkMode,
		e.defaultFileMode,
		e.defaultDirectoryMode,
		e.defaultOwnership,
		e.recomposeUnicode,
		e.stager,
	)

	e.stager.wipe()

	return results, problems, nil
}

func (e *endpoint) Shutdown() error {

	e.watchCancel()

	return nil
}
