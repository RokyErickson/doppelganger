package session

import (
	contextpkg "context"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/filesystem"
	"github.com/RokyErickson/doppelganger/pkg/state"
	"github.com/RokyErickson/doppelganger/pkg/url"
)

type Manager struct {
	tracker      *state.Tracker
	sessionsLock *state.TrackingLock
	sessions     map[string]*controller
}

func NewManager() (*Manager, error) {

	tracker := state.NewTracker()
	sessionsLock := state.NewTrackingLock(tracker)
	sessions := make(map[string]*controller)

	sessionsDirectory, err := pathForSession("")
	if err != nil {
		return nil, errors.Wrap(err, "unable to compute sessions directory")
	}
	sessionsDirectoryContents, err := filesystem.DirectoryContentsByPath(sessionsDirectory)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read contents of sessions directory")
	}
	for _, c := range sessionsDirectoryContents {
		identifier := c.Name()
		if controller, err := loadSession(tracker, identifier); err != nil {
			continue
		} else {
			sessions[identifier] = controller
		}
	}

	return &Manager{
		tracker:      tracker,
		sessionsLock: sessionsLock,
		sessions:     sessions,
	}, nil
}

func (m *Manager) allControllers() []*controller {
	m.sessionsLock.Lock()
	defer m.sessionsLock.UnlockWithoutNotify()

	controllers := make([]*controller, 0, len(m.sessions))
	for _, controller := range m.sessions {
		controllers = append(controllers, controller)
	}

	return controllers
}

const (
	minimumSessionSpecificationLength = 5
)

func (m *Manager) findControllers(specifications []string) ([]*controller, error) {
	m.sessionsLock.Lock()
	defer m.sessionsLock.UnlockWithoutNotify()

	controllers := make([]*controller, 0, len(specifications))
	for _, specification := range specifications {
		if specification == "" {
			return nil, errors.New("empty session specification is invalid")
		} else if len(specification) < minimumSessionSpecificationLength {
			return nil, errors.Errorf(
				"session specification must be at least %d characters",
				minimumSessionSpecificationLength,
			)
		}

		var match *controller
		for _, controller := range m.sessions {
			if controller.session.Identifier == specification {
				match = controller
				break
			}
			fuzzy := strings.HasPrefix(controller.session.Identifier, specification) ||
				strings.Contains(controller.session.Alpha.Path, specification) ||
				strings.Contains(controller.session.Beta.Path, specification) ||
				strings.Contains(controller.session.Alpha.Hostname, specification) ||
				strings.Contains(controller.session.Beta.Hostname, specification)
			if fuzzy {
				if match != nil {
					return nil, errors.Errorf("specification \"%s\" matches multiple sessions", specification)
				}
				match = controller
			}
		}
		if match == nil {
			return nil, errors.Errorf("specification \"%s\" doesn't match any sessions", specification)
		}
		controllers = append(controllers, match)
	}
	return controllers, nil
}

func (m *Manager) Shutdown() {

	m.tracker.Poison()
	m.sessionsLock.Lock()
	defer m.sessionsLock.UnlockWithoutNotify()
	for _, controller := range m.sessions {
		if err := controller.halt(haltModeShutdown, ""); err != nil {

		}
	}
}

func (m *Manager) Create(
	alpha, beta *url.URL,
	configuration, configurationAlpha, configurationBeta *Configuration,
	prompter string,
) (string, error) {
	controller, err := newSession(
		m.tracker,
		alpha, beta,
		configuration, configurationAlpha, configurationBeta,
		prompter,
	)
	if err != nil {
		return "", err
	}

	m.sessionsLock.Lock()
	m.sessions[controller.session.Identifier] = controller
	m.sessionsLock.Unlock()

	return controller.session.Identifier, nil
}

func (m *Manager) List(previousStateIndex uint64, specifications []string) (uint64, []*State, error) {
	stateIndex, poisoned := m.tracker.WaitForChange(previousStateIndex)
	if poisoned {
		return 0, nil, errors.New("state tracking terminated")
	}

	var controllers []*controller
	if len(specifications) == 0 {
		controllers = m.allControllers()
	} else if cs, err := m.findControllers(specifications); err != nil {
		return 0, nil, errors.Wrap(err, "unable to locate requested sessions")
	} else {
		controllers = cs
	}

	states := make([]*State, len(controllers))
	for i, controller := range controllers {
		states[i] = controller.currentState()
	}

	sort.Slice(states, func(i, j int) bool {
		iTime := states[i].Session.CreationTime
		jTime := states[j].Session.CreationTime
		return iTime.Seconds < jTime.Seconds ||
			(iTime.Seconds == jTime.Seconds && iTime.Nanos < jTime.Nanos)
	})

	return stateIndex, states, nil
}

func (m *Manager) Flush(specifications []string, prompter string, skipWait bool, context contextpkg.Context) error {

	var controllers []*controller
	if len(specifications) == 0 {
		controllers = m.allControllers()
	} else if cs, err := m.findControllers(specifications); err != nil {
		return errors.Wrap(err, "unable to locate requested sessions")
	} else {
		controllers = cs
	}

	for _, controller := range controllers {
		if err := controller.flush(prompter, skipWait, context); err != nil {
			return errors.Wrap(err, "unable to flush session")
		}
	}

	return nil
}

func (m *Manager) Pause(specifications []string, prompter string) error {

	var controllers []*controller
	if len(specifications) == 0 {
		controllers = m.allControllers()
	} else if cs, err := m.findControllers(specifications); err != nil {
		return errors.Wrap(err, "unable to locate requested sessions")
	} else {
		controllers = cs
	}

	for _, controller := range controllers {
		if err := controller.halt(haltModePause, prompter); err != nil {
			return errors.Wrap(err, "unable to pause session")
		}
	}

	return nil
}

func (m *Manager) Resume(specifications []string, prompter string) error {

	var controllers []*controller
	if len(specifications) == 0 {
		controllers = m.allControllers()
	} else if cs, err := m.findControllers(specifications); err != nil {
		return errors.Wrap(err, "unable to locate requested sessions")
	} else {
		controllers = cs
	}

	for _, controller := range controllers {
		if err := controller.resume(prompter); err != nil {
			return errors.Wrap(err, "unable to resume session")
		}
	}

	return nil
}

func (m *Manager) Terminate(specifications []string, prompter string) error {

	var controllers []*controller
	if len(specifications) == 0 {
		controllers = m.allControllers()
	} else if cs, err := m.findControllers(specifications); err != nil {
		return errors.Wrap(err, "unable to locate requested sessions")
	} else {
		controllers = cs
	}

	for _, controller := range controllers {
		if err := controller.halt(haltModeTerminate, prompter); err != nil {
			return errors.Wrap(err, "unable to terminate session")
		}
		m.sessionsLock.Lock()
		delete(m.sessions, controller.session.Identifier)
		m.sessionsLock.Unlock()
	}

	return nil
}
