package memcached

import (
	"bytes"
	"text/template"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/backend"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/irutil"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"golang.org/x/exp/slog"
)

type MemcachedGoClient struct {
	golang.Service
	backend.Cache

	InstanceName string
	Addr         *MemcachedAddr
}

func newMemcachedGoClient(name string, addr *MemcachedAddr) (*MemcachedGoClient, error) {
	client := &MemcachedGoClient{}
	client.InstanceName = name
	client.Addr = addr
	return client, nil
}

func (n *MemcachedGoClient) String() string {
	return n.InstanceName + " = MemcachedClient(" + n.Addr.Name() + ")"
}

func (n *MemcachedGoClient) Name() string {
	return n.InstanceName
}

func (n *MemcachedGoClient) GetInterface(visitor irutil.BuildContext) service.ServiceInterface {
	// TODO: return memcached interface
	return nil
}

func (n *MemcachedGoClient) GetGoInterface(visitor irutil.BuildContext) *gocode.ServiceInterface {
	// TODO: return memcached interface
	return nil
}

var clientBuildFuncTemplate = `func(ctr golang.Container) (any, error) {

		// TODO: generated memcached client constructor

		return nil, nil

	}`

// Part of code generation compilation pass; provides instantiation snippet
func (node *MemcachedGoClient) AddInstantiation(builder golang.GraphBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	// TODO: generate the grpc stubs

	// Instantiate the code template
	t, err := template.New(node.InstanceName).Parse(clientBuildFuncTemplate)
	if err != nil {
		return err
	}

	// Generate the code
	buf := &bytes.Buffer{}
	err = t.Execute(buf, node)
	if err != nil {
		return err
	}

	slog.Info("instantiating memcached client")
	return builder.Declare(node.InstanceName, buf.String())
}

func (node *MemcachedGoClient) ImplementsGolangNode()    {}
func (node *MemcachedGoClient) ImplementsGolangService() {}
