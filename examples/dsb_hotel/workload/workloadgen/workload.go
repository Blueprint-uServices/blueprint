package workloadgen

import (
	"context"
	"fmt"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow/hotelreservation"
)

type SimpleWorkload interface {
	ImplementsSimpleWorkload(ctx context.Context) error
}

type workloadGen struct {
	SimpleWorkload

	frontend hotelreservation.FrontEndService
}

func NewSimpleWorkload(ctx context.Context, frontend hotelreservation.FrontEndService) (SimpleWorkload, error) {
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
			hotels, err := s.frontend.SearchHandler(ctx, "Vaastav", "2015-04-09", "2015-04-10", 37.7835, -122.41, "en")
			if err != nil {
				return err
			}
			fmt.Println("Query found", len(hotels), "hotels!")
		}
	}
}

func (s *workloadGen) ImplementsSimpleWorkload(context.Context) error {
	return nil
}
