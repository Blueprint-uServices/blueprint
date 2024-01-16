// Package workload is a plugin for creating executable workload generators.
//
// Most out-of-the-box applications will come with a workload generator that can be included with the
// compiled application using this plugin.  Typically the workload generator is implemented in a
// directory called "workload" next to the application's workflow, wiring, and tests.
//
// For details on how to write custom workload generators, see the Writing Workload Generators section
// below.
//
// # Wiring Spec Usage
//
// To include a workload generator in your wiring spec, specify a name for the workload generator and
// point it at the implementation.  A typical workload generator will also have some arguments that are
// service clients.
//
//	workload.Generator(spec, "my_workload_gen", "WorkloadImpl", "my_frontend_service")
//
// The workload plugin will search the workflow spec modules for a valid workflow service called "WorkloadImpl".
// It will create and compile a process that runs the service.
//
// Workload generators are typically implemented in a separate module from the workflow logic, so you will
// probably need to make sure that the workload generator module of your application is on the workflow spec
// search path.  See for example the [SockShop Workload Generator].
//
//	workflow.Init("../workflow", "../tests", "../workload")
//
// # Artifacts Generated
//
// The plugin creates a subdirectory in the output containing the workload generator source.
// The plugin also compiles the workload generator source into an executable that resides in the subdirectory.
//
// # Running the Workload Generator
//
// The workload generator plugin automatically compiles an executable binary.  Navigate to the output
// directory and run the binary, e.g.
//
//	chmod +x ./build/my_workload_gen/my_workload_gen_proc
//	./build/my_workload_gen/my_workload_gen_proc
//
// Your application will probably require some addresses to be provided as arguments, and will complain if
// they are absent.
//
// For convenience, the generated source for the workload generator is also included in the build directory.
//
// # Writing Workload Generators
//
// You can implement your workload generator within the same package/module as your workflow logic or in a
// separate package/module.  For convention, we recommend implementing your workload generator in a separate
// module called "workload" alongside your workflow, wiring, and tests directories.
//
// Workload generators are implemented in an identical manner to workflow services.  They receive service
// clients as arguments.  They can define additional flags for command-line arguments to configure the
// workload (e.g. number of threads, request rate, etc.).
//
// The logic of the workload generator should reside in the Run method.  See [SockShop Workload Generator]
// for an example.
//
// # Example:
//
//	// The WorkloadGen interface, to use as the workloadType in the call to workload.Generator
//	type SimpleWorkload interface {
//		ImplementsSimpleWorkload(context.Context) error
//	}
//
//	// WorkloadGen implementation
//	type workloadGen struct {
//		SimpleWorkload
//
//		frontend frontend.Frontend 		// Application client
//	}
//
//	var myarg = flag.Int("myarg", 12345, "help message for myarg")
//
//	func NewSimpleWorkload(ctx context.Context, frontend frontend.Frontend) (SimpleWorkload, error) {
//		return &workloadGen{frontend: frontend}, nil
//	}
//
//	func (s *workloadGen) Run(ctx context.Context) error {
//		fmt.Printf("myarg is %v\n", *myarg)
//		ticker := time.NewTicker(1 * time.Second)
//		for {
//			// Workload runs here
//			select {
//			case <-ctx.Done():
//				return nil
//			case t := <-ticker.C:
//				fmt.Println("Tick at", t)
//			}
//		}
//	}
//
//	func (s *workloadGen) ImplementsSimpleWorkload(context.Context) error {
//		return nil
//	}
//
// [SockShop Workload Generator]: https://github.com/blueprint-uservices/blueprint/tree/main/examples/sockshop/workload/workloadgen/workload.go
package workload

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/namespaceutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	"github.com/pkg/errors"
)

// [Generator] can be used by wiring specs to build an executable workload generator.
//
// workloadType should correspond to a workload generator implementation
//
// workloadArgs should correspond to arguments used by the workload generator implementation
func Generator(spec wiring.WiringSpec, name string, workloadType string, workloadArgs ...string) string {

	serviceName := name + ".service"
	procName := name + ".proc"
	wlgenName := name

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
