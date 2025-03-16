package workloadgen

import (
	"context"
	"flag"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow/hotelreservation"
	"github.com/blueprint-uservices/blueprint/runtime/core/workload"
)

// Workload specific flags
var outfile = flag.String("outfile", "stats.csv", "Outfile where individual request information will be stored")
var duration = flag.String("duration", "1m", "Duration for which the workload should be run")
var tput = flag.Int64("tput", 1000, "Desired throughput")

type ComplexWorkload interface {
	ImplementsComplexWorkload(ctx context.Context) error
}

type complexWldGen struct {
	ComplexWorkload

	frontend hotelreservation.FrontEndService
}

func NewComplexWorkload(ctx context.Context, frontend hotelreservation.FrontEndService) (ComplexWorkload, error) {
	w := &complexWldGen{frontend: frontend}
	return w, nil
}

type FnType func() error

func statWrapper(fn FnType) workload.Stat {
	start := time.Now()
	err := fn()
	duration := time.Since(start)
	s := workload.Stat{}
	s.Start = start.UnixNano()
	s.Duration = duration.Nanoseconds()
	s.IsError = (err != nil)
	return s
}

func (w *complexWldGen) RunSearchHandler(ctx context.Context) workload.Stat {
	lat, lon, indate, outdate := GenSearchHandler()
	customerName := "Blueprint User"
	locale := "en"
	return statWrapper(func() error {
		_, err := w.frontend.SearchHandler(ctx, customerName, indate, outdate, lat, lon, locale)

		return err
	})
}

func (w *complexWldGen) RunUserHandler(ctx context.Context) workload.Stat {
	username, password := GenUserHandler()
	return statWrapper(func() error {
		_, err := w.frontend.UserHandler(ctx, username, password)

		return err
	})
}

func (w *complexWldGen) RunRecommendHandler(ctx context.Context) workload.Stat {
	lat, lon, req := GenRecommendHandler()
	return statWrapper(func() error {
		_, err := w.frontend.RecommendHandler(ctx, lat, lon, req, "en")

		return err
	})
}

func (w *complexWldGen) RunReservationHandler(ctx context.Context) workload.Stat {
	indate, outdate, hotelid, username, password, customername, roomnumber := GenReservationHandler()
	return statWrapper(func() error {
		_, err := w.frontend.ReservationHandler(ctx, indate, outdate, hotelid, customername, username, password, roomnumber)

		return err
	})
}

func (w *complexWldGen) Run(ctx context.Context) error {
	wrk := workload.NewWorkload()
	// Configure the workload with the client side generators for the various APIs and their respective proportions
	wrk.AddAPI("SearchHandler", w.RunSearchHandler, 60)
	wrk.AddAPI("RecommendHandler", w.RunRecommendHandler, 38)
	wrk.AddAPI("UserHandler", w.RunUserHandler, 1)
	wrk.AddAPI("ReservationHandler", w.RunReservationHandler, 1)
	// Initialize the engine
	engine, err := workload.NewEngine(*outfile, *tput, *duration, wrk)
	if err != nil {
		return err
	}
	// Run the workload
	engine.RunOpenLoop(ctx)
	// Print statistics from the workload
	return engine.PrintStats()
}

func (w *complexWldGen) ImplementsComplexWorkload(context.Context) error {
	return nil
}
