package workload

// Stat captures basic statistics about a request
type Stat struct {
	Start    int64
	Duration int64
	IsError  bool
}
