package local

import (
	"github.com/pkg/errors"

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

	if url.Protocol != urlpkg.Protocol_Local {
		panic("non-local URL dispatched to local protocol handler")
	}

	endpoint, err := NewEndpoint(url.Path, session, version, configuration, alpha)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create local endpoint")
	}

	return endpoint, nil
}

func init() {

	session.ProtocolHandlers[urlpkg.Protocol_Local] = &protocolHandler{}
}
