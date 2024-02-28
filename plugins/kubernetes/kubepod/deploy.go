package kubepod

import (
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/ioutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
	"golang.org/x/exp/slog"

	"github.com/blueprint-uservices/blueprint/plugins/kubernetes/kubepod/deploygen"
)

// A Kubernetes pod deployer. It generates the pod config files on the local filesystem.
type kubePodDeployment interface {
	ir.ArtifactGenerator
}

// A workspace used when deploying a set of containers as a Kubernetes Pod
//
// Implements docker.ContainerWorkspace defined in docker/ir.go
//
// This workspace generates Pod files at the root of the output directory.
type kubeDeploymentWorkspace struct {
	ir.VisitTrackerImpl

	info docker.ContainerWorkspaceInfo

	ImageDirs    map[string]string
	InstanceArgs map[string][]ir.IRNode // argnodes for each instance added to the workspace

	F *deploygen.KubeDeploymentFile
}

// Implements ir.ArtifactGenerator
func (node *PodDeployment) GenerateArtifacts(dir string) error {
	slog.Info(fmt.Sprintf("Generating container instances for Kubernetes Pod %s in %s", node.Name(), dir))
	workspace := NewKubePodWorkspace(node.Name(), dir)
	return node.generateArtifacts(workspace)
}

func (node *PodDeployment) generateArtifacts(workspace *kubeDeploymentWorkspace) error {
	// Add all locally-built container images
	for _, n := range ir.Filter[docker.ProvidesContainerImage](node.Nodes) {
		if err := n.AddContainerArtifacts(workspace); err != nil {
			return err
		}
	}

	// Add all pre-built container instances
	for _, n := range ir.Filter[docker.ProvidesContainerInstance](node.Nodes) {
		if err := n.AddContainerInstance(workspace); err != nil {
			return err
		}
	}

	// Build the Kubernetes pod config files
	if err := workspace.Finish(); err != nil {
		return err
	}
	return nil
}

func NewKubePodWorkspace(name string, dir string) *kubeDeploymentWorkspace {
	return &kubeDeploymentWorkspace{
		info: docker.ContainerWorkspaceInfo{
			Path:   filepath.Clean(dir),
			Target: "kubedeployment",
		},
		ImageDirs:    make(map[string]string),
		InstanceArgs: make(map[string][]ir.IRNode),
		F:            deploygen.NewKubeDeploymentFile(name, dir, name+"-deployment.yaml", name+"-service.yaml"),
	}
}

// Implements docker.ContainerWorkspace
func (p *kubeDeploymentWorkspace) Info() docker.ContainerWorkspaceInfo {
	return p.info
}

// Implements docker.ContainerWorkspace
func (p *kubeDeploymentWorkspace) CreateImageDir(imageName string) (string, error) {
	// Only alphanumeric and underscores are allowed in a proc name
	imageName = ir.CleanName(imageName)
	imageDir, err := ioutil.CreateNodeDir(p.info.Path, imageName)
	p.ImageDirs[imageName] = imageDir
	return imageDir, err
}

// Implements docker.ContainerWorkspace
func (p *kubeDeploymentWorkspace) DeclarePrebuiltInstance(instanceName string, image string, args ...ir.IRNode) error {
	p.InstanceArgs[instanceName] = args
	return p.F.AddImageInstance(instanceName, image)
}

// Implements docker.ContainerWorkspace
func (p *kubeDeploymentWorkspace) DeclareLocalImage(instanceName string, imageDir string, args ...ir.IRNode) error {
	slog.Info("Inside DeclareLocalImage")
	p.InstanceArgs[instanceName] = args
	// For now set image to instanceName
	image := instanceName
	return p.F.AddImageInstance(instanceName, image)
}

// Implements docker.ContainerWorkspace
func (p *kubeDeploymentWorkspace) SetEnvironmentVariable(instanceName string, key string, val string) error {
	return p.F.AddEnvVar(instanceName, key, val)
}

// Generates the pod config file
func (p *kubeDeploymentWorkspace) Finish() error {
	// We didn't set any arguments or environment variables while accumulating instances. Do so now.
	if err := p.processArgNodes(); err != nil {
		return err
	}

	return p.F.Generate()
}

func asMap[T any](s []*T) map[*T]struct{} {
	m := make(map[*T]struct{})
	for _, v := range s {
		m[v] = struct{}{}
	}
	return m
}

// Goes through each container's arg nodes, determining which need to be passed to the container
// as environment variables.
//
// Has special handling for addresses; containers that bind a server will have ports assigned,
// and containers that dial to a server within this namespace will have the dial address set.
//
// We don't pick external-facing ports for any addresses; these will be set by the caller or user.
func (p *kubeDeploymentWorkspace) processArgNodes() error {

	// (1) Assign ports to containers
	// Servers like backends will already be pre-bound to specific ports.  Other servers
	// like gRPC ones will need a port assigned.
	// The networking address space in a pod is shared between containers, so port assignments
	// must be unique across all containers in the pod.
	var allBinds []*address.BindConfig
	var assignedBinds map[*address.BindConfig]struct{}
	var localAddresses map[string]string
	{
		localAddresses = make(map[string]string)
		for _, instanceArgs := range p.InstanceArgs {
			allBinds = append(allBinds, ir.Filter[*address.BindConfig](instanceArgs)...)
		}

		_, assigned, err := address.AssignPorts(allBinds)
		if err != nil {
			return err
		}
		assignedBinds = asMap(assigned)

		localAddresses = make(map[string]string)
		for _, bind := range allBinds {
			localAddresses[bind.AddressName] = fmt.Sprintf("localhost:%v", bind.Port)
		}
	}

	// (2) Set environment variables for containers
	// (3) Expose container ports externally
	for instanceName, instanceArgs := range p.InstanceArgs {
		binds, dials, remaining := address.Split(instanceArgs)

		// Handle the instanceArgs that are regular config args and not address related
		for _, arg := range remaining {
			switch node := arg.(type) {
			case ir.IRConfig:
				if node.HasValue() {
					// Ignore if the value is already set, because it implies it's hard-coded
					// inside the container image
				} else {
					// TODO: if Kubernetes supports pass-through environment variables, then
					// implement this
					return blueprint.Errorf("kubernetes doesn't support runtime environment variable passthrough for %v", node.Name())
				}
			default:
				return blueprint.Errorf("container instance %v can only accept IRConfig nodes as arguments, but found %v of type %v", instanceName, arg, reflect.TypeOf(arg))
			}
		}

		// If we assigned ports for this container, then set the environment variables for them
		for _, bind := range binds {
			if _, isAssigned := assignedBinds[bind]; isAssigned {
				p.F.AddEnvVar(instanceName, bind.Name(), fmt.Sprintf("0.0.0.0:%v", bind.Port))
			}
		}

		// Expose all ports for this container
		for _, bind := range binds {
			p.F.ExposePort(instanceName, bind.AddressName, bind.Port)
		}

		// If a dial is local, then it just calls localhost:port.  If it's not local
		// then ??????????????????
		for _, dial := range dials {
			if addr, isLocalDial := localAddresses[dial.AddressName]; isLocalDial {
				p.F.AddEnvVar(instanceName, dial.Name(), addr)
			} else {
				// TODO: do we pass through an environment variable? how do we know the name to dial for
				// services that are outside this service?  maybe just use the
				// service_a.grpc.dial_addr name itself as the service lookup name???
			}
		}
	}

	return nil
}

func (p *kubeDeploymentWorkspace) ImplementsBuildContext()       {}
func (p *kubeDeploymentWorkspace) ImplementsContainerWorkspace() {}
