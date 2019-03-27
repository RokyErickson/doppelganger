package remote

import (
	"github.com/RokyErickson/doppelganger/pkg/protocols/local"
	"github.com/RokyErickson/doppelganger/pkg/session"
)

type EndpointConnectionValidator func(string, string, session.Version, *session.Configuration, bool) error

type endpointServerOptions struct {
	root                string
	configuration       *session.Configuration
	connectionValidator EndpointConnectionValidator
	endpointOptions     []local.EndpointOption
}

type EndpointServerOption interface {
	apply(*endpointServerOptions)
}

type functionEndpointServerOption struct {
	applier func(*endpointServerOptions)
}

func newFunctionEndpointServerOption(applier func(*endpointServerOptions)) EndpointServerOption {
	return &functionEndpointServerOption{applier}
}

func (o *functionEndpointServerOption) apply(options *endpointServerOptions) {
	o.applier(options)
}

func WithRoot(root string) EndpointServerOption {
	return newFunctionEndpointServerOption(func(options *endpointServerOptions) {
		options.root = root
	})
}

func WithConfiguration(configuration *session.Configuration) EndpointServerOption {
	return newFunctionEndpointServerOption(func(options *endpointServerOptions) {
		options.configuration = configuration
	})
}

func WithConnectionValidator(validator EndpointConnectionValidator) EndpointServerOption {
	return newFunctionEndpointServerOption(func(options *endpointServerOptions) {
		options.connectionValidator = validator
	})
}

func WithEndpointOption(option local.EndpointOption) EndpointServerOption {
	return newFunctionEndpointServerOption(func(options *endpointServerOptions) {
		options.endpointOptions = append(options.endpointOptions, option)
	})
}
