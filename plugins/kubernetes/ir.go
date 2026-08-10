package kubernetes

import (
	"errors"
	"fmt"
	"log"
	"log/slog"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/kubernetes/kubeconfig"
	"github.com/blueprint-uservices/blueprint/plugins/kubernetes/kubepod"
	"github.com/blueprint-uservices/blueprint/plugins/linux"
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

type kubeAddrInfo struct {
	AddrName string
	PodName  string
	Bind     *address.BindConfig
	Dial     *address.DialConfig
}

// Implements ir.ArtifactGenerator
func (n *Application) GenerateArtifacts(dir string) error {
	slog.Info("Number of nodes in Kubernetes", "length", len(n.Nodes))
	addrs := make(map[string]*kubeAddrInfo)
	var dials []*address.DialConfig
	nodes := ir.Filter[*kubepod.PodDeployment](n.Nodes)
	for _, node := range nodes {
		err := node.GenerateArtifacts(dir)
		if err != nil {
			return err
		}
		dials = append(dials, node.AllDials...)
		log.Println("Pod Name:", kubepod.KubernetesName(node.PodName))
		for _, bind := range node.AllBinds {
			addrs[bind.AddressName] = &kubeAddrInfo{AddrName: bind.AddressName, PodName: kubepod.KubernetesName(node.PodName), Bind: bind, Dial: nil}
		}
	}

	config_file := kubeconfig.NewKubeConfigFile("app-config", "app_config.yaml", dir)

	for _, dial := range dials {
		if info, ok := addrs[dial.AddressName]; ok {
			config_file.AddData(linux.EnvVar(dial.Name()), fmt.Sprintf("%s:%v", info.PodName, info.Bind.Port))
		} else {
			return errors.New("Kubernetes Config: Failed to find bind information for address:" + dial.AddressName)
		}
	}

	return config_file.Generate()
}
