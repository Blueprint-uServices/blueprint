package goproc

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/namespaceutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
)

// Adds a child node to an existing process
func AddToProcess(spec wiring.WiringSpec, procName, childName string) {
	namespaceutil.AddNodeTo[Process](spec, procName, childName)
}

// Wraps serviceName with a modifier that deploys the service inside a Golang process
func Deploy(spec wiring.WiringSpec, serviceName string) string {
	procName := serviceName + "_proc"
	CreateProcess(spec, procName, serviceName)
	return serviceName
}

// Creates a process with a given name, and adds the provided nodes as children.  This method
// is only needed when creating processes with more than one child node; otherwise it is easier
// to use [Deploy]
func CreateProcess(spec wiring.WiringSpec, procName string, children ...string) string {
	// If any children were provided in this call, add them to the process via a property
	for _, childName := range children {
		AddToProcess(spec, procName, childName)
	}

	// Install default metric collector
	metric_collector := defineStdoutMetricCollector(spec, procName)
	SetMetricCollector(spec, procName, metric_collector)

	// Install default logger
	logger := defineStdoutLogger(spec, procName)
	SetLogger(spec, procName, logger)

	// The process node is simply a namespace that accepts [golang.Node] nodes
	spec.Define(procName, &Process{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		var metric_coll string
		err := spec.GetProperty(procName, "metricCollector", &metric_coll)
		if err != nil {
			return nil, err
		}
		var logger_name string
		err = spec.GetProperty(procName, "logger", &logger_name)
		if err != nil {
			return nil, err
		}
		proc := newGolangProcessNode(procName)

		procNamespace, err := namespaceutil.InstantiateNamespace(namespace, &GolangProcessNamespace{proc})
		if err != nil {
			return nil, err
		}
		err = procNamespace.Get(metric_coll, &proc.metricProvider)
		if err != nil {
			return nil, err
		}
		err = procNamespace.Get(logger_name, &proc.logger)
		if err != nil {
			return nil, err
		}
		return proc, err
	})

	return procName
}

// Creates a process that contains clients to the specified children.  This is for convenience in
// serving as a starting point to write a custom client
func CreateClientProcess(spec wiring.WiringSpec, procName string, children ...string) string {
	spec.Define(procName, &Process{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		proc := newGolangProcessNode(procName)
		procNamespace, err := namespace.DeriveNamespace(procName, &GolangProcessNamespace{proc})
		if err != nil {
			return nil, err
		}
		for _, child := range children {
			var childNode ir.IRNode
			if err := procNamespace.Get(child, &childNode); err != nil {
				return nil, err
			}
		}
		return proc, err
	})

	return procName
}

// Override the default metric collector for this process
func SetMetricCollector(spec wiring.WiringSpec, procName string, metricCollNodeName string) {
	spec.SetProperty(procName, "metricCollector", metricCollNodeName)
}

// Override the default logger for this process
func SetLogger(spec wiring.WiringSpec, procName string, loggerNodeName string) {
	spec.SetProperty(procName, "logger", loggerNodeName)
}

// Defines the default metric collector
func defineStdoutMetricCollector(spec wiring.WiringSpec, processName string) string {
	collector := processName + ".stdoutmetriccollector"
	spec.Define(collector, &stdoutMetricCollector{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		return newStdOutMetricCollector(collector)
	})
	return collector
}

// Defines the default logger
func defineStdoutLogger(spec wiring.WiringSpec, processName string) string {
	logger := processName + ".logger"
	spec.Define(logger, &stdoutLogger{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		return newStdoutLogger(logger)
	})
	return logger
}

// A [wiring.NamespaceHandler] used to build [Process] IRNodes
type GolangProcessNamespace struct {
	*Process
}

// Implements [wiring.NamespaceHandler]
func (proc *GolangProcessNamespace) Accepts(nodeType any) bool {
	_, isGolangNode := nodeType.(golang.Node)
	return isGolangNode
}

// Implements [wiring.NamespaceHandler]
func (proc *GolangProcessNamespace) AddEdge(name string, edge ir.IRNode) error {
	proc.Edges = append(proc.Edges, edge)
	return nil
}

// Implements [wiring.NamespaceHandler]
func (proc *GolangProcessNamespace) AddNode(name string, node ir.IRNode) error {
	proc.Nodes = append(proc.Nodes, node)
	return nil
}
