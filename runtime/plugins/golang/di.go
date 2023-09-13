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

/* For nodes that want to run background goroutines */
type Runnable interface {
	Run(ctx context.Context) error
}

type BuildFunc func(ctr Container) (any, error)

/*
A simple dependency injection container used by generated go code

Create one with NewGraph() method
*/
type diImpl struct {
	Graph
	Container

	ctx          context.Context
	cancel       context.CancelFunc
	cancelParent context.CancelFunc
	wg           *sync.WaitGroup
	buildFuncs   map[string]BuildFunc
	built        map[string]any
}

func NewGraph(ctx context.Context, cancel context.CancelFunc) Graph {
	graph := &diImpl{}
	graph.buildFuncs = make(map[string]BuildFunc)
	graph.built = make(map[string]any)
	graph.ctx, graph.cancel = context.WithCancel(ctx)
	graph.cancelParent = cancel
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

		if runnable, isRunnable := built.(Runnable); isRunnable {
			graph.wg.Add(1)
			go func() {
				err := runnable.Run(graph.ctx)
				if err != nil {
					slog.Error("error running node " + name)
					graph.cancel()
					graph.cancelParent()
				}
				graph.wg.Done()
			}()
		}

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
