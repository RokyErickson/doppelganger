package local

import (
	"context"
)

type endpointOptions struct {
	cachePathCallback   func(string, bool) (string, error)
	stagingRootCallback func(string, bool) (string, error)
	watchingMechanism   func(context.Context, string, chan<- struct{})
}

type EndpointOption interface {
	apply(*endpointOptions)
}

type functionEndpointOption struct {
	applier func(*endpointOptions)
}

func newFunctionEndpointOption(applier func(*endpointOptions)) EndpointOption {
	return &functionEndpointOption{applier}
}

func (o *functionEndpointOption) apply(options *endpointOptions) {
	o.applier(options)
}

func WithCachePathCallback(callback func(string, bool) (string, error)) EndpointOption {
	return newFunctionEndpointOption(func(options *endpointOptions) {
		options.cachePathCallback = callback
	})
}

func WithStagingRootCallback(callback func(string, bool) (string, error)) EndpointOption {
	return newFunctionEndpointOption(func(options *endpointOptions) {
		options.stagingRootCallback = callback
	})
}

func WithWatchingMechanism(callback func(context.Context, string, chan<- struct{})) EndpointOption {
	return newFunctionEndpointOption(func(options *endpointOptions) {
		options.watchingMechanism = callback
	})
}
