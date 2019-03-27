package ssh

import (
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

	if url.Protocol != urlpkg.Protocol_SSH {
		panic("non-SSH URL dispatched to SSH protocol handler")
	}

	transport := &transport{url, prompter}

	return agent.Dial(transport, prompter, url.Path, session, version, configuration, alpha)
}

func init() {

	session.ProtocolHandlers[urlpkg.Protocol_SSH] = &protocolHandler{}
}
