package session

import (
	"context"

	"github.com/pkg/errors"

	urlpkg "github.com/RokyErickson/doppelganger/pkg/url"
)

type ProtocolHandler interface {
	Dial(
		url *urlpkg.URL,
		prompter,
		session string,
		version Version,
		configuration *Configuration,
		alpha bool,
	) (Endpoint, error)
}

var ProtocolHandlers = map[urlpkg.Protocol]ProtocolHandler{}

func connect(
	url *urlpkg.URL,
	prompter,
	session string,
	version Version,
	configuration *Configuration,
	alpha bool,
) (Endpoint, error) {
	handler, ok := ProtocolHandlers[url.Protocol]
	if !ok {
		return nil, errors.Errorf("unknown protocol: %s", url.Protocol)
	} else if handler == nil {
		panic("nil protocol handler registered")
	}

	endpoint, err := handler.Dial(url, prompter, session, version, configuration, alpha)
	if err != nil {
		return nil, errors.Wrap(err, "unable to connect to endpoint")
	}

	return endpoint, nil
}

type asyncConnectResult struct {
	endpoint Endpoint
	error    error
}

func reconnect(
	ctx context.Context,
	url *urlpkg.URL,
	session string,
	version Version,
	configuration *Configuration,
	alpha bool,
) (Endpoint, error) {
	results := make(chan asyncConnectResult)

	go func() {
		endpoint, err := connect(url, "", session, version, configuration, alpha)

		select {
		case <-ctx.Done():
			if endpoint != nil {
				endpoint.Shutdown()
			}
		case results <- asyncConnectResult{endpoint, err}:
		}
	}()

	select {
	case <-ctx.Done():
		return nil, errors.New("reconnect cancelled")
	case result := <-results:
		return result.endpoint, result.error
	}
}
