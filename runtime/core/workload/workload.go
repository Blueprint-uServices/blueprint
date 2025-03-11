package workload

import "context"

// RequestFunction wraps around the request execution and response processing.
// This function is user provided for each API.
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
