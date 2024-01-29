package kubepod

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
)

// An IRNode representing a Kubernetes pod, which is simply a collection of container instances.
type PodDeployment struct {
	kubePodDeployment
	PodName string
	Nodes   []ir.IRNode
	Edges   []ir.IRNode
}

// Implements IRNode
func (node *PodDeployment) Name() string {
	return node.PodName
}

// Implements IRNode
func (node *PodDeployment) String() string {
	return ir.PrettyPrintNamespace(node.PodName, "KubernetesPod", node.Edges, node.Nodes)
}

// Implements [wiring.NamespaceHandler]
func (pod *PodDeployment) Accepts(nodeType any) bool {
	_, isDockerContainerNode := nodeType.(docker.Container)
	return isDockerContainerNode
}

// Implements [wiring.NamespaceHandler]
func (pod *PodDeployment) AddEdge(name string, edge ir.IRNode) error {
	pod.Edges = append(pod.Edges, edge)
	return nil
}

// Implements [wiring.NamespaceHandler]
func (pod *PodDeployment) AddNode(name string, node ir.IRNode) error {
	pod.Nodes = append(pod.Nodes, node)
	return nil
}
