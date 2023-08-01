package golang_workflow

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/service"
)

// Code location and interfaces of a service
type GolangServiceDetails struct {
	Interface service.ServiceInterface
	Files     []string
	Package   string
}

func (d GolangServiceDetails) String() string {
	var b strings.Builder
	b.WriteString(d.Interface.Name)

	var constructorArgs []string
	for _, arg := range d.Interface.ConstructorArgs {
		constructorArgs = append(constructorArgs, arg.Type)
	}
	b.WriteString("(")
	b.WriteString(strings.Join(constructorArgs, ", "))
	b.WriteString(")")

	return b.String()
}

// This Node represents a Golang Workflow spec service in the Blueprint IR.
type GolangWorkflowSpecServiceNode struct {
	blueprint.IRNode
	service.ServiceNode

	InstanceName   string
	ServiceDetails *GolangServiceDetails
	Args           []blueprint.IRNode
}

func (n GolangWorkflowSpecServiceNode) String() string {
	var b strings.Builder
	b.WriteString("GolangWorkflowSpecServiceNode ")
	b.WriteString(n.InstanceName)
	b.WriteString(" = ")
	b.WriteString(n.ServiceDetails.Interface.Name)

	var args []string
	for _, arg := range n.Args {
		args = append(args, arg.Name())
	}

	b.WriteString("(")
	b.WriteString(strings.Join(args, ", "))
	b.WriteString(")")

	return b.String()
}

func newGolangWorkflowSpecServiceNode(name string, details *GolangServiceDetails, args []blueprint.IRNode) *GolangWorkflowSpecServiceNode {
	node := GolangWorkflowSpecServiceNode{}

	node.InstanceName = name
	node.ServiceDetails = details
	node.Args = args
	return &node
}

func (node *GolangWorkflowSpecServiceNode) Name() string {
	return node.InstanceName
}

func (node *GolangWorkflowSpecServiceNode) GetInterface() *service.ServiceInterface {
	return &node.ServiceDetails.Interface
}

func (node *GolangWorkflowSpecServiceNode) GenerateInstantiationCode() string {
	return `
		di.Add(serviceName, scope, func(ctr) {
			first = ctr.Get(arg0)
			second = ctr.Get(arg1)
			return new ServiceName(first, second)
		})`
}
