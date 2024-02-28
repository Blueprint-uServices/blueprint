package kubernetes

import "github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"

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
	nodes := ir.Filter[ir.ArtifactGenerator](n.Nodes)
	for _, node := range nodes {
		err := node.GenerateArtifacts(dir)
		if err != nil {
			return err
		}
	}
	return nil
}
