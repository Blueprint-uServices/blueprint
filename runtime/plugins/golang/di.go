package golang

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/exp/slog"
)

// This contains the runtime components used by plugin-generated golang code.

type Graph interface {
	Define(name string, build BuildFunc) error
}

type Container interface {
	Get(name string) (any, error)
	Context() context.Context   // In case the buildfunc wants to start background goroutines
	WaitGroup() *sync.WaitGroup // Waitgroup used by this container; plugins can call Add if they create goroutines
}

type BuildFunc func(ctr Container) (any, error)

/*
A simple dependency injection container used by generated go code

Create one with NewGraph() method
*/
type diImpl struct {
	Graph
	Container

	ctx        context.Context
	wg         *sync.WaitGroup
	buildFuncs map[string]BuildFunc
	built      map[string]any
}

func NewGraph(ctx context.Context) Graph {
	graph := &diImpl{}
	graph.buildFuncs = make(map[string]BuildFunc)
	graph.built = make(map[string]any)
	graph.ctx = ctx
	graph.wg = &sync.WaitGroup{}
	return graph
}

func (graph *diImpl) Define(name string, build BuildFunc) error {
	if _, exists := graph.buildFuncs[name]; exists {
		slog.Warn("redefining " + name + "; this might indicate a bad wiring spec")
	}
	graph.buildFuncs[name] = build
	return nil
}

func (graph *diImpl) Get(name string) (any, error) {
	if existing, exists := graph.built[name]; exists {
		return existing, nil
	}
	if build, exists := graph.buildFuncs[name]; exists {
		built, err := build(graph)
		if err != nil {
			return nil, err
		}
		graph.built[name] = built
		return built, nil
	} else {
		return nil, fmt.Errorf("unknown %v", name)
	}
}

func (graph *diImpl) Context() context.Context {
	return graph.ctx
}

func (graph *diImpl) WaitGroup() *sync.WaitGroup {
	return graph.wg
}
