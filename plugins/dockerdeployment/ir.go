package dockerdeployment

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/stringutil"
)

/* A deployment is a collection of containers */
type Deployment struct {
	/* The implemented build targets for dockercompose.DockerCompose nodes */
	dockerComposeDeployer /* Can be deployed as a docker-compose file; implemented in deploydockercompose.go */

	DeploymentName string
	ArgNodes       []blueprint.IRNode
	ContainedNodes []blueprint.IRNode
}

func newContainerDeployment(name string) *Deployment {
	return &Deployment{DeploymentName: name}
}

func (node *Deployment) Name() string {
	return node.DeploymentName
}

func (node *Deployment) String() string {
	var b strings.Builder
	b.WriteString(node.DeploymentName)
	b.WriteString(" = DockerApp(")
	var args []string
	for _, arg := range node.ArgNodes {
		args = append(args, arg.Name())
	}
	b.WriteString(strings.Join(args, ", "))
	b.WriteString(") {\n")
	var children []string
	for _, child := range node.ContainedNodes {
		children = append(children, child.String())
	}
	b.WriteString(stringutil.Indent(strings.Join(children, "\n"), 2))
	b.WriteString("\n}")
	return b.String()
}

func (node *Deployment) AddArg(argnode blueprint.IRNode) {
	node.ArgNodes = append(node.ArgNodes, argnode)
}

func (node *Deployment) AddChild(child blueprint.IRNode) error {
	node.ContainedNodes = append(node.ContainedNodes, child)
	return nil
}
