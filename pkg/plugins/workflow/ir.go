package workflow

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/golang"
)

// This Node represents a Golang Workflow spec service in the Blueprint IR.
type WorkflowService struct {
	golang.Node
	golang.Service
	service.ServiceNode
	golang.ArtifactGenerator
	golang.CodeGenerator

	InstanceName   string
	ServiceDetails *golang.GolangServiceDetails
	Args           []blueprint.IRNode
}

func (n WorkflowService) String() string {
	var b strings.Builder
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

func newWorkflowService(name string, details *golang.GolangServiceDetails, args []blueprint.IRNode) *WorkflowService {
	node := WorkflowService{}

	node.InstanceName = name
	node.ServiceDetails = details
	node.Args = args
	return &node
}

func (node *WorkflowService) Name() string {
	return node.InstanceName
}

func (node *WorkflowService) GetInterface() *service.ServiceInterface {
	return &node.ServiceDetails.Interface
}

func (node *WorkflowService) GenerateInstantiationCode(g *golang.GolangCodeGenerator) error {
	code := `
		di.Add(serviceName, scope, func(ctr) {
			first = ctr.Get(arg0)
			second = ctr.Get(arg1)
			return new ServiceName(first, second)
		})`
	g.Def(node.InstanceName, code)
	g.Import(node.ServiceDetails.Package)
	return nil
}

func (node *WorkflowService) CollectArtifacts(g *golang.GolangArtifactGenerator) error {
	return g.AddFiles(node.ServiceDetails.Files)
}

func (node *WorkflowService) ImplementsGolangNode()    {}
func (node *WorkflowService) ImplementsGolangService() {}
