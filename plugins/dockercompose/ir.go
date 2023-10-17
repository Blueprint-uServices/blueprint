package dockercompose

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
)

type ContainerDeployment interface {
	blueprint.IRNode
}

/* A collection of containers combined into a docker-compose file */
type DockerCompose struct {
	ContainerDeployment

	DeploymentName string
	ArgNodes       []blueprint.IRNode
	ContainedNodes []blueprint.IRNode
}

type DockerComposeBuilder interface {
	AddImage()
	AddDockerfile()
	ConfigureContainer(name string)
}

type DockerContainerBuilder interface {
	ExposePort(internalPort string) string
	SetConfig(name string, value string)
}

func newDockerComposeDeployment(name string) *DockerCompose {
	return &DockerCompose{DeploymentName: name}
}

func (node *DockerCompose) Name() string {
	return node.DeploymentName
}

func (node *DockerCompose) String() string {
	var b strings.Builder
	b.WriteString(node.DeploymentName)
	b.WriteString(" = DockerCompose(")
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
	b.WriteString(blueprint.Indent(strings.Join(children, "\n"), 2))
	b.WriteString("\n}")
	return b.String()
}

func (node *DockerCompose) AddArg(argnode blueprint.IRNode) {
	node.ArgNodes = append(node.ArgNodes, argnode)
}

func (node *DockerCompose) AddChild(child blueprint.IRNode) error {
	node.ContainedNodes = append(node.ContainedNodes, child)
	return nil
}

func (node *DockerCompose) GenerateArtifacts(dir string) error {
	// builder, err := dockergen.NewDockerComposeBuilder(dir)
	// if err != nil {
	// 	return err
	// }

	// for _, node := range node.ContainedNodes {
	// 	if n, valid := node.(docker.BuildsContainerImage); valid {
	// 		if err := n.PrepareDockerfile(builder); err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	// for _, node := range node.ContainedNodes {
	// 	if n, valid := node.(docker.Container); valid {
	// 		if err := n.AddToDockerCompose(builder); err != nil {
	// 			return err
	// 		}
	// 	}
	// }
	return nil
}
