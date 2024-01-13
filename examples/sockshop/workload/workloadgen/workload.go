package workloadgen

import (
	"context"
	"fmt"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/frontend"
)

// The WorkloadGen interface, which the Blueprint compiler will treat as a
// Workflow service
type SimpleWorkload interface {
	ImplementsSimpleWorkload(context.Context) error
}

// workloadGen implementation
type workloadGen struct {
	SimpleWorkload

	frontend frontend.Frontend
}

func init() {
	// TODO: define cmd line flags
}

func NewSimpleWorkload(ctx context.Context, frontend frontend.Frontend) (SimpleWorkload, error) {
	return &workloadGen{frontend: frontend}, nil
}

func (s *workloadGen) Run(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return nil
		case t := <-ticker.C:
			fmt.Println("Tick at", t)
		}
	}
}

func (s *workloadGen) ImplementsSimpleWorkload(context.Context) error {
	return nil
}
