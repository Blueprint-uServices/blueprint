package workloadgen

// The WorkloadGen interface, which the Blueprint compiler will treat as a
// Workflow service
type WorkloadGen interface {
	ImplementsWorkloadGen()
}

// workloadGen implementation
type workloadGen struct {
	WorkloadGen
}
