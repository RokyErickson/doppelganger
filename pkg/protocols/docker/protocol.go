package docker

import (
	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/agent"
	"github.com/RokyErickson/doppelganger/pkg/session"
	urlpkg "github.com/RokyErickson/doppelganger/pkg/url"
)

type protocolHandler struct{}

func (h *protocolHandler) Dial(
	url *urlpkg.URL,
	prompter,
	session string,
	version session.Version,
	configuration *session.Configuration,
	alpha bool,
) (session.Endpoint, error) {

	if url.Protocol != urlpkg.Protocol_Docker {
		panic("non-Docker URL dispatched to Docker protocol handler")
	}

	transport, err := newTransport(url, prompter)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create Docker transport")
	}

	return agent.Dial(transport, prompter, url.Path, session, version, configuration, alpha)
}

func init() {

	session.ProtocolHandlers[urlpkg.Protocol_Docker] = &protocolHandler{}
}
