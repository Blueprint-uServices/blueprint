package kubepod

import (
	"fmt"
	"path/filepath"

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

	ImageDirs map[string]string

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

	// Reset any port assignments for externally-visible servers
	address.ResetPorts(node.Edges)
	return nil
}

func NewKubePodWorkspace(name string, dir string) *kubeDeploymentWorkspace {
	return &kubeDeploymentWorkspace{
		info: docker.ContainerWorkspaceInfo{
			Path:   filepath.Clean(dir),
			Target: "kubedeployment",
		},
		ImageDirs: make(map[string]string),
		F:         deploygen.NewKubeDeploymentFile(name, dir, name+"-deployment.yaml", name+"-service.yaml"),
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
	if err := address.CheckPorts(args); err != nil {
		return blueprint.Errorf("unable to add docker instance %v due to %v", instanceName, err.Error())
	}

	return p.F.AddImageInstance(instanceName, image, args...)
}

// Implements docker.ContainerWorkspace
func (p *kubeDeploymentWorkspace) DeclareLocalImage(instanceName string, imageDir string, args ...ir.IRNode) error {
	// Docker containers should assign all internal server ports (typically using address.AssignPorts) before adding an instance
	if err := address.CheckPorts(args); err != nil {
		return blueprint.Errorf("unable to add docker instance %v due to %v", instanceName, err.Error())
	}
	// For now set image to instanceName
	image := instanceName
	return p.F.AddImageInstance(instanceName, image, args...)
}

// Implements docker.ContainerWorkspace
func (p *kubeDeploymentWorkspace) SetEnvironmentVariable(instanceName string, key string, val string) error {
	return p.F.AddEnvVar(instanceName, key, val)
}

// Generates the pod config file
func (p *kubeDeploymentWorkspace) Finish() error {
	return p.F.Generate()
}

func (p *kubeDeploymentWorkspace) ImplementsBuildContext()       {}
func (p *kubeDeploymentWorkspace) ImplementsContainerWorkspace() {}
