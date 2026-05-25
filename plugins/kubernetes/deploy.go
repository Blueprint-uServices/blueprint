package kubernetes

import (
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/ioutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
	"github.com/blueprint-uservices/blueprint/plugins/kubernetes/kubernetesgen"
	"golang.org/x/exp/slog"
)

type (
	// kubernetesDeployer is the deployer interface for Kubernetes deployments
	kubernetesDeployer interface {
		ir.ArtifactGenerator
	}

	// kubernetesWorkspace is a workspace used when deploying a set of containers
	// to a Kubernetes cluster. It implements docker.ContainerWorkspace.
	kubernetesWorkspace struct {
		ir.VisitTrackerImpl

		info docker.ContainerWorkspaceInfo

		ImageDirs    map[string]string      // map from image name to directory
		InstanceArgs map[string][]ir.IRNode // argnodes for each instance added to the workspace

		ManifestBuilder *kubernetesgen.ManifestBuilder
		Deployment      *KubernetesDeployment
	}
)

// Implements ir.ArtifactGenerator
func (node *KubernetesDeployment) GenerateArtifacts(dir string) error {
	slog.Info(fmt.Sprintf("Collecting container instances for Kubernetes deployment %s in %s", node.Name(), dir))
	workspace := NewKubernetesWorkspace(node, dir)
	return node.generateArtifacts(workspace)
}

// The basic build process of a Kubernetes deployment
func (node *KubernetesDeployment) generateArtifacts(workspace *kubernetesWorkspace) error {
	// Add any locally-built container images
	for _, containerNode := range ir.Filter[docker.ProvidesContainerImage](node.Nodes) {
		if err := containerNode.AddContainerArtifacts(workspace); err != nil {
			return err
		}
	}

	// Collect all container instances
	for _, containerNode := range ir.Filter[docker.ProvidesContainerInstance](node.Nodes) {
		if err := containerNode.AddContainerInstance(workspace); err != nil {
			return err
		}
	}

	// Build the Kubernetes manifests
	if err := workspace.Finish(); err != nil {
		return err
	}

	return nil
}

func NewKubernetesWorkspace(deployment *KubernetesDeployment, dir string) *kubernetesWorkspace {
	return &kubernetesWorkspace{
		info: docker.ContainerWorkspaceInfo{
			Path:   filepath.Clean(dir),
			Target: "kubernetes",
		},
		ImageDirs:    make(map[string]string),
		InstanceArgs: make(map[string][]ir.IRNode),
		ManifestBuilder: kubernetesgen.NewManifestBuilder(
			deployment.DeploymentName,
			deployment.Namespace,
			deployment.Replicas,
			dir,
		),
		Deployment: deployment,
	}
}

// Implements docker.ContainerWorkspace
func (w *kubernetesWorkspace) Info() docker.ContainerWorkspaceInfo {
	return w.info
}

// Implements docker.ContainerWorkspace
func (w *kubernetesWorkspace) CreateImageDir(imageName string) (string, error) {
	// Only alphanumeric and underscores are allowed in an image name
	imageName = ir.CleanName(imageName)
	imageDir, err := ioutil.CreateNodeDir(w.info.Path, imageName)
	w.ImageDirs[imageName] = imageDir
	return imageDir, err
}

// Implements docker.ContainerWorkspace
func (w *kubernetesWorkspace) DeclarePrebuiltInstance(instanceName string, image string, args ...ir.IRNode) error {
	w.InstanceArgs[instanceName] = args
	return w.ManifestBuilder.AddContainer(instanceName, image, false)
}

// Implements docker.ContainerWorkspace
func (w *kubernetesWorkspace) DeclareLocalImage(instanceName string, imageDir string, args ...ir.IRNode) error {
	w.InstanceArgs[instanceName] = args
	// For local images, we'll need to build and push them to a registry
	// For now, we'll use the imageDir as the image name
	return w.ManifestBuilder.AddContainer(instanceName, imageDir, true)
}

// Implements docker.ContainerWorkspace
func (w *kubernetesWorkspace) SetEnvironmentVariable(instanceName string, key string, val string) error {
	return w.ManifestBuilder.AddEnvVar(instanceName, key, val)
}

// Generates the Kubernetes manifests
func (w *kubernetesWorkspace) Finish() error {
	// Process arg nodes for environment variables and networking
	if err := w.processArgNodes(); err != nil {
		return err
	}

	// Generate all Kubernetes manifests
	return w.ManifestBuilder.Generate()
}

// processArgNodes processes each container's arg nodes, determining which need to be passed
// as environment variables, and handling service discovery through Kubernetes Services.
func (w *kubernetesWorkspace) processArgNodes() error {
	addresses := make(map[string]string)

	for instanceName, instanceArgs := range w.InstanceArgs {
		binds, dials, remaining := address.Split(instanceArgs)

		// Handle non-address arguments (config nodes)
		for _, arg := range remaining {
			switch node := arg.(type) {
			case ir.IRConfig:
				if !node.HasValue() {
					// Pass through as environment variable
					w.ManifestBuilder.PassthroughEnvVar(instanceName, node.Name(), node.Optional())
				}
			default:
				return blueprint.Errorf("container instance %v can only accept IRConfig nodes as arguments, but found %v of type %v",
					instanceName, arg, reflect.TypeOf(arg))
			}
		}

		// Assign ports and create services for bound addresses
		_, assigned, err := address.AssignPorts(binds)
		if err != nil {
			return err
		}

		// Add assigned ports as environment variables and create Kubernetes services
		for _, bind := range assigned {
			w.ManifestBuilder.AddEnvVar(instanceName, bind.Name(), fmt.Sprintf("0.0.0.0:%v", bind.Port))
		}

		// Create Kubernetes services for all bound ports
		for _, bind := range binds {
			serviceName := ir.CleanName(instanceName)
			// In Kubernetes, services are accessed by their DNS name
			addresses[bind.AddressName] = fmt.Sprintf("%v:%v", serviceName, bind.Port)
			w.ManifestBuilder.ExposePort(instanceName, bind.Port, bind.Name())
		}

		address.Clear(binds)
	}

	// Set dial addresses for inter-service communication
	for instanceName, instanceArgs := range w.InstanceArgs {
		_, dials, _ := address.Split(instanceArgs)
		for _, dial := range dials {
			if addr, isLocalDial := addresses[dial.AddressName]; isLocalDial {
				// Use Kubernetes service DNS name for local services
				w.ManifestBuilder.AddEnvVar(instanceName, dial.Name(), addr)
			} else {
				// External service - pass through as environment variable
				w.ManifestBuilder.PassthroughEnvVar(instanceName, dial.Name(), false)
			}
		}
	}

	return nil
}

func (w *kubernetesWorkspace) ImplementsBuildContext()       {}
func (w *kubernetesWorkspace) ImplementsContainerWorkspace() {}
