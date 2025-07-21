package crisp

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
)

// CRISPContainer represents the CRISP Python FastAPI service container
type CRISPContainer struct {
	docker.Container
	docker.ProvidesContainerInstance

	InstanceName string
	BindAddr     *address.BindConfig
}

func newCRISPContainer(name string) *CRISPContainer {
	return &CRISPContainer{
		InstanceName: name,
		BindAddr:     &address.BindConfig{},
	}
}

// Implements ir.IRNode
func (c *CRISPContainer) Name() string {
	return c.InstanceName
}

// Implements ir.IRNode
func (c *CRISPContainer) String() string {
	return c.InstanceName + " = CRISPContainer(" + c.BindAddr.Name() + ")"
}

// Implements docker.ProvidesContainerInstance
func (node *CRISPContainer) AddContainerInstance(target docker.ContainerWorkspace) error {
	node.BindAddr.Port = 8000
	node.BindAddr.Key = node.InstanceName + ".bind_addr"
	return target.DeclarePrebuiltInstance(node.InstanceName, "crisp:latest", node.BindAddr)
}

// Container wires up the CRISP Python FastAPI service as a container
func Container(spec wiring.WiringSpec, name string) string {
	spec.Define(name, &CRISPContainer{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		return newCRISPContainer(name), nil
	})
	return name
} 