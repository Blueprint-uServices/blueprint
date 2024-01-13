package workloadgen

import (
	"context"
	"fmt"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/frontend"
)

// The WorkloadGen interface, which the Blueprint compiler will treat as a
// Workflow service
type WorkloadGen interface {
	ImplementsWorkloadGen()
}

// workloadGen implementation
type workloadGen struct {
	WorkloadGen

	frontend frontend.Frontend
}

func NewWorkloadGen(ctx context.Context, frontend frontend.Frontend) (WorkloadGen, error) {
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
