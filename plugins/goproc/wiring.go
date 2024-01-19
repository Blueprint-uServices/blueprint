// Package goproc is a plugin for instantiating golang application-level instances within a single golang process.
//
// # Wiring Spec Usage
//
// To use the goproc plugin in your wiring spec, you can declare a process, giving it a name and specifying which
// golang instances to include.
//
//	goproc.CreateProcess(spec, "my_process", "user_service", "payment_service", "cart_service")
//
// If you are only deploying a single service within the process, you can use the shorter [Deploy]:
//
//	goproc.Deploy(spec, "user_service")
//
// When a service is added to a process, the goproc plugin also adds a modifier to the service.  Thus, any
// application-level modifiers should be applied to the service *before* deploying it into a process.
//
// If you expect services to be reachable from outside the process, then make sure they have been deployed
// using e.g. the gRPC plugin, prior to deploying the service in a process.
//
// # Default Builder
//
// The goproc plugin is the default builder for "floating" golang instances (ie, golang instances that haven't
// been added to any process).
//
// # Artifacts Generated
//
// During compilation, the plugin creates a golang workspace and pulls in all module dependencies.  Within the
// workspace, the plugin creates a module and instructs all golang instances to generate their code into that
// module.  Finally, the plugin generates a main.go with a main method that instantiates the golang instances.
//
// # Running artifacts
//
// Optionally you can build an executable for the generated goproc by running
//
//	go build {{.procName}}
//
// A generated goproc can be run from the workspace directory containing main.go
//
//	cd {{.procName}}
//	go run . -h
//
// The goproc may require additional command line arguments (e.g. bind or dial addresses) in order to run; if so,
// running the goproc will report any missing variables.
//
// # Internals
//
// Internally, the goproc plugin makes use of interfaces defined in the [golang] plugin.  It can combine any
// golang.Node IRNodes.  The plugin uses the WorkspaceBuilder, ModuleBuilder, and NamespaceBuilder defined by
// the golang plugin to accumulate and generate code.
//
// [golang]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/golang
package goproc

import (
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/namespaceutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
)

// AddToProcess can be used by wiring specs to add a golang instance to an existing golang process.
func AddToProcess(spec wiring.WiringSpec, procName, childName string) {
	namespaceutil.AddNodeTo[Process](spec, procName, childName)
}

// Deploy can be used by wiring specs to deploy a golang service in a golang process.
//
// Adds a modifier to the service that will create the golang process if not already created.
//
// The name of the process created is determined by attempting to replace a "_service" suffix
// with "_proc", or adding "_proc" if serviceName doesn't end with "_service", e.g.
//
//	user_service => user_proc
//	user => user_proc
//	user_srv => user_srv_proc
//
// After calling [Deploy], serviceName will be a process-level service.
//
// Returns the name of the created process.
func Deploy(spec wiring.WiringSpec, serviceName string) string {
	servicePrefix, _ := strings.CutSuffix(serviceName, "_service")
	procName := servicePrefix + "_proc"
	CreateProcess(spec, procName, serviceName)
	return procName
}

// CreateProcess can be used by wiring specs to define a process called procName and to deploy
// the golang services children.  CreateProcess only needs to be used when more than one children
// are being added to the process; otherwise it is more convenient to use [Deploy].
//
// After calling CreateProcess, other golang services can still be added to the process by calling
// [AddToProcess] using the same procName.
//
// After calling CreateProcess, any children that are services will be process-level services
// that can now have process-level modifiers applied to them, or can be deployed to containers.
//
// procName is configured with a logger that prints to stdout.  To change the logger,
// call [SetLogger].
//
// procName is configured with a metric collector that prints to stdout.  To change the metric
// collector, call [SetMetricCollector]
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
	nodeType := newGolangProcessNode(procName)
	spec.Define(procName, nodeType, func(namespace wiring.Namespace) (ir.IRNode, error) {
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

		procNamespace, err := namespaceutil.InstantiateNamespace(namespace, &golangProcessNamespace{proc})
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

// CreateClientProcess can be used by wiring specs to create a process that contains only clients
// of the specified children.  This is for convenience in serving as a starting point to write a custom client
func CreateClientProcess(spec wiring.WiringSpec, procName string, children ...string) string {
	spec.Define(procName, &Process{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		proc := newGolangProcessNode(procName)
		procNamespace, err := namespace.DeriveNamespace(procName, &golangProcessNamespace{proc})
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

// SetMetricCollector is not used directly by wiring specs; instead it is used by other plugins such as
// [opentelemetry] to install custom metric collectors.
//
// [opentelemetry]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/opentelemetry
func SetMetricCollector(spec wiring.WiringSpec, procName string, metricCollNodeName string) {
	spec.SetProperty(procName, "metricCollector", metricCollNodeName)
}

// SetLogger is not used directly by wiring specs; instead it is used by other plugins such as
// [opentelemetry] to install custom loggers.
//
// [opentelemetry]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/opentelemetry
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
type golangProcessNamespace struct {
	*Process
}

// Implements [wiring.NamespaceHandler]
func (proc *golangProcessNamespace) Accepts(nodeType any) bool {
	_, isGolangNode := nodeType.(golang.Node)
	return isGolangNode
}

// Implements [wiring.NamespaceHandler]
func (proc *golangProcessNamespace) AddEdge(name string, edge ir.IRNode) error {
	proc.Edges = append(proc.Edges, edge)
	return nil
}

// Implements [wiring.NamespaceHandler]
func (proc *golangProcessNamespace) AddNode(name string, node ir.IRNode) error {
	proc.Nodes = append(proc.Nodes, node)
	return nil
}
