package workload

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/namespaceutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	"github.com/pkg/errors"
)

func Generator(spec wiring.WiringSpec, name string, workloadType string, workloadArgs ...string) string {

	serviceName := name + ".service"
	procName := name
	wlgenName := name + ".wlgen"

	// Define the service
	workflow.Service(spec, serviceName, workloadType, workloadArgs...)

	// Wrap the service in a process
	goproc.CreateProcess(spec, procName, serviceName)

	// Define the workload gen node
	spec.Define(wlgenName, &workloadGenerator{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		w := newWorkloadGenerator(wlgenName, procName)

		wlgenNamespace, err := namespaceutil.InstantiateNamespace(namespace, &wlgen{w})
		if err != nil {
			return nil, err
		}

		if err := wlgenNamespace.Get(procName, &w.ProcNode); err != nil {
			return nil, err
		}

		return w, nil
	})

	return wlgenName
}

type wlgen struct {
	*workloadGenerator
}

// Implements [wiring.NamespaceHandler]
func (w *wlgen) Accepts(nodeType any) bool {
	proc, isProcNode := nodeType.(*goproc.Process)
	return isProcNode && proc.InstanceName == w.ProcName
}

// Implements [wiring.NamespaceHandler]
func (w *wlgen) AddEdge(name string, edge ir.IRNode) error {
	// Ignore
	return nil
}

// Implements [wiring.NamespaceHandler]
func (w *wlgen) AddNode(name string, node ir.IRNode) error {
	proc, isProcNode := node.(*goproc.Process)
	if !isProcNode {
		return errors.Errorf("Unexpected node %v = %v", name, node)
	} else if proc.InstanceName != w.WorkloadName {
		return errors.Errorf("Unexpected process %v != %v", proc.InstanceName, w.WorkloadName)
	} else if w.ProcNode != nil {
		return errors.Errorf("Unexpected duplicate workload proc %v", node)
	} else {
		w.ProcNode = proc
		return nil
	}
}
