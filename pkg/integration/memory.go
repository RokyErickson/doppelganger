package integration

import (
	"net"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/remote"
	"github.com/RokyErickson/doppelganger/pkg/session"
	urlpkg "github.com/RokyErickson/doppelganger/pkg/url"
)

const (
	inMemoryProtocol urlpkg.Protocol = -1
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

	if url.Protocol != inMemoryProtocol {
		panic("non-in-memory URL dispatched to in-memory protocol handler")
	}

	clientConnection, serverConnection := net.Pipe()

	go remote.ServeEndpoint(serverConnection)

	endpoint, err := remote.NewEndpointClient(
		clientConnection,
		url.Path,
		session,
		version,
		configuration,
		alpha,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create in-memory endpoint client")
	}

	return endpoint, nil
}

func init() {

	session.ProtocolHandlers[inMemoryProtocol] = &protocolHandler{}
}
