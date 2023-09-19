package workload

import (
	"fmt"
	"os"
	"path/filepath"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
)

/*
Golang-level client that will make calls to a service
*/
type WorkloadgenClient struct {
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Wrapped      golang.Service

	outputPackage string
}

func NewWorkloadGenerator(name string, node blueprint.IRNode) (*WorkloadgenClient, error) {
	service, isService := node.(golang.Service)
	if !isService {
		return nil, fmt.Errorf("cannot create workload generator for non-service %v", node)
	}

	workload := &WorkloadgenClient{}
	workload.InstanceName = name
	workload.Wrapped = service
	workload.outputPackage = "workloadgen"
	return workload, nil
}
func (node *WorkloadgenClient) GenerateFuncs(builder golang.ModuleBuilder) error {
	// Only generate the workload code for this instance once
	if builder.Visited(node.Wrapped.GetInterface().GetName() + ".workloadgen") {
		return nil
	}

	// Create the package output dir
	outputDir := filepath.Join(builder.Info().Path, node.outputPackage)
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return fmt.Errorf("unable to create workloadgen output dir %v due to %v", outputDir, err.Error())
	}

	// The workload generator can only call methods on a gocode.ServiceInterface
	service, valid := node.Wrapped.GetInterface().(*gocode.ServiceInterface)
	if !valid {
		return fmt.Errorf("expected %v to have a gocode.ServiceInterface but got %v",
			node.Name(), node.Wrapped.GetInterface())
	}

	// Generate the code
	return GenerateWorkloadgenCode(builder, service, "workloadgen")
}

// Provides the golang code to instantiate the workloadgen client
func (node *WorkloadgenClient) AddInstantiation(builder golang.GraphBuilder) error {
	// Only add instantiation code for this specific client once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	// The workload generator can only call methods on a gocode.ServiceInterface
	service, valid := node.Wrapped.GetInterface().(*gocode.ServiceInterface)
	if !valid {
		return fmt.Errorf("expected %v to have a gocode.ServiceInterface but got %v",
			node.Name(), node.Wrapped.GetInterface())
	}

	constructor := &gocode.Constructor{
		Package: builder.Module().Info().Name + "/" + node.outputPackage,
		Func: gocode.Func{
			Name:      fmt.Sprintf("New_%v_WorkloadGenerator", service.Name),
			Arguments: []gocode.Variable{{Name: "service", Type: service}},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []blueprint.IRNode{node.Wrapped})
}

func (workloadgen *WorkloadgenClient) Name() string {
	return workloadgen.InstanceName
}

func (workloadgen *WorkloadgenClient) String() string {
	return fmt.Sprintf("%v = WorkloadGenerator(%v)", workloadgen.Name(), workloadgen.Wrapped.Name())
}

func (node *WorkloadgenClient) ImplementsGolangNode() {}
