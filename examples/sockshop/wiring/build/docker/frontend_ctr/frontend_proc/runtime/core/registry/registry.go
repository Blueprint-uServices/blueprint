// Package registry provides a struct for registering different
// service constructors for use at runtime.
//
// This package is primarily used by plugins for testing
// and workload generators that don't know at development time
// the service instance to be used.
package registry

import (
	"context"
	"fmt"

	"golang.org/x/exp/slog"
)

// Used for registering constructors for different service clients
type ServiceRegistry[T any] struct {
	name             string
	defaultBuildFunc string
	registered       map[string]func(context.Context) (T, error)
	isBuilt          bool
	built            T
	buildError       error
}

func NewServiceRegistry[T any](name string) *ServiceRegistry[T] {
	return &ServiceRegistry[T]{
		name:       name,
		registered: make(map[string]func(context.Context) (T, error)),
		isBuilt:    false,
	}
}

func (r *ServiceRegistry[T]) SetDefault(name string) {
	r.defaultBuildFunc = name
}

func (r *ServiceRegistry[T]) Register(name string, build func(ctx context.Context) (T, error)) {
	slog.Info(fmt.Sprintf("ServiceRegistry \"%v\" added client \"%v\"", r.name, name))
	if len(r.registered) == 0 {
		r.defaultBuildFunc = name
	}
	r.registered[name] = build
}

// Returns the service instance or client.  If this is the first call to Get,
// then the client will be built.  The client and any build error is cached,
// and will be returned by any subsequent calls to Get
func (r *ServiceRegistry[T]) Get(ctx context.Context) (T, error) {
	if r.isBuilt {
		return r.built, r.buildError
	}
	r.isBuilt = true

	if len(r.registered) == 0 {
		r.buildError = fmt.Errorf("no clients registered for %v", r.name)
		return r.built, r.buildError
	}

	buildFunc, hasDefault := r.registered[r.defaultBuildFunc]
	if !hasDefault {
		r.buildError = fmt.Errorf("no client called \"%v\" known for %v", r.defaultBuildFunc, r.name)
		return r.built, r.buildError
	}

	slog.Info(fmt.Sprintf("ServiceRegistry \"%v\" building client \"%v\"", r.name, r.defaultBuildFunc))
	r.built, r.buildError = buildFunc(ctx)
	return r.built, r.buildError
}
