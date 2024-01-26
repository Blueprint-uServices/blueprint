package kubepod

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
)

// An IRNode representing a Kubernetes pod, which is simply a collection of container instances.
type Pod struct {
	kubePod
	PodName string
	Nodes   []ir.IRNode
	Edges   []ir.IRNode
}

// Implements IRNode
func (node *Pod) Name() string {
	return node.PodName
}

// Implements IRNode
func (node *Pod) String() string {
	return ir.PrettyPrintNamespace(node.PodName, "KubernetesPod", node.Edges, node.Nodes)
}

// Implements [wiring.NamespaceHandler]
func (pod *Pod) Accepts(nodeType any) bool {
	_, isDockerContainerNode := nodeType.(docker.Container)
	return isDockerContainerNode
}

// Implements [wiring.NamespaceHandler]
func (pod *Pod) AddEdge(name string, edge ir.IRNode) error {
	pod.Edges = append(pod.Edges, edge)
	return nil
}

// Implements [wiring.NamespaceHandler]
func (pod *Pod) AddNode(name string, node ir.IRNode) error {
	pod.Nodes = append(pod.Nodes, node)
	return nil
}
