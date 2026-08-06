package kubernetes

import (
	"log/slog"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
)

// An IRNode representing a Kubernetes applicaiton deployment which is a collection of Kubernetes Pod + Service Deployment instances.
type Application struct {
	AppName string
	Nodes   []ir.IRNode
	Edges   []ir.IRNode
}

// Implements IRNode
func (n *Application) Name() string {
	return n.AppName
}

// Implements IRNode
func (n *Application) String() string {
	return ir.PrettyPrintNamespace(n.AppName, "KubeApp", n.Edges, n.Nodes)
}

// Implements ir.ArtifactGenerator
func (n *Application) GenerateArtifacts(dir string) error {
	slog.Info("Number of nodes in Kuberenets", "length", len(n.Nodes))
	nodes := ir.Filter[ir.ArtifactGenerator](n.Nodes)
	for _, node := range nodes {
		slog.Info("Generating Kubernetes artifact", "node", node)
		err := node.GenerateArtifacts(dir)
		if err != nil {
			return err
		}
	}
	for _, edge := range n.Edges {
		slog.Info("Edge to " + edge.Name())
	}
	return nil
}
