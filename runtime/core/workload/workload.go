// Package workload provides an abstraction for a workload and a corresponding engine that
// executes the workload in an open-loop fashion.
//
// Sample Usage:
//
// import "github.com/Blueprint-uServices/blueprint/runtime/workload"
//
//	func (w *SimpleWorkload) Run(ctx context.Context) error {
//	   wld := workload.NewWorkload()
//	   wld.AddAPI("read", w.ReadRequestGenerator, 80)
//	   wld.AddAPI("write", w.WriteRequestGenerator, 20)
//	   engine, err := workload.NewEngine("stats.csv", 1000, "1m", wld)
//	   if err != nil { return err }
//	   engine.RunOpenLoop(ctx)
//	   return engine.PrintStats()
//	}
//
// For a complete example, refer to https://github.com/Blueprint-uServices/blueprint/tree/main/examples/dsb_hotel/cmplx_workload
package workload

import "context"

// RequestFunction wraps around the request execution and response processing.
// The function generates the arguments, makes a call to the desired service, and the processes the response.
// This function is provided by the user for each API.
type RequestFunction func(context.Context) Stat

// APIInfo encapsulates the execution information for each API.
type ApiInfo struct {
	Name       string
	Fn         RequestFunction
	Proportion int
}

// Workload represents a workload (list of APIs and their corresponding proportions) that is to be executed against the target application.
type Workload struct {
	ApiInfos []ApiInfo
}

// NewWorkload creates a new workload instance
func NewWorkload() *Workload {
	w := &Workload{}
	return w
}

// AddAPI adds a new API to be executed as part of the workload
//
// name is the name of the API
// fn is the RequestFunction that will generate the required parameters, execute the request, collect and report the stats for the request.
// proportion is the proportion of requests that need to be executed for this API.
//
// For a workload to be considered valid, the proportions of all the added APIs must sum to 100.
func (w *Workload) AddAPI(name string, fn RequestFunction, proportion int) error {
	info := ApiInfo{Name: name, Fn: fn, Proportion: proportion}
	w.ApiInfos = append(w.ApiInfos, info)
	return nil
}

// IsValid checks if the configured workload is valid or not.
// Validity Conditions:
//   - The proportions of all APIs should sum to 100
func (w *Workload) IsValid() bool {
	total := 0
	for _, api := range w.ApiInfos {
		total += api.Proportion
	}
	if total == 100 {
		return true
	}
	return false
}
