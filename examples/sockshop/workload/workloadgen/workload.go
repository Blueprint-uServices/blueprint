package workloadgen

import (
	"context"
	"fmt"
	"time"

	"flag"

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

var myarg = flag.Int("myarg", 12345, "help message for myarg")

func NewSimpleWorkload(ctx context.Context, frontend frontend.Frontend) (SimpleWorkload, error) {
	return &workloadGen{frontend: frontend}, nil
}

func (s *workloadGen) Run(ctx context.Context) error {
	_, err := s.frontend.LoadCatalogue(ctx)
	if err != nil {
		fmt.Println("Failed to load catalogue")
		return err
	}
	fmt.Printf("myarg is %v\n", *myarg)
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return nil
		case t := <-ticker.C:
			fmt.Println("Tick at", t)
			items, err := s.frontend.ListItems(ctx, []string{}, "", 1, 100)
			if err != nil {
				return err
			}
			fmt.Println("Got", len(items), "items!")
		}
	}
}

func (s *workloadGen) ImplementsSimpleWorkload(context.Context) error {
	return nil
}
