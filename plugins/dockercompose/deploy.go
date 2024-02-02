package dockercompose

import (
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/ioutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
	"github.com/blueprint-uservices/blueprint/plugins/dockercompose/dockergen"
	"golang.org/x/exp/slog"
)

type (
	/*
		Docker compose is the default docker app deployer.  It simply
		generates a docker-compose.yml file on the local filesystem.
	*/
	dockerComposeDeployer interface {
		ir.ArtifactGenerator
	}

	/*
	   A workspace used when deploying a set of containers as a
	   docker-compose application

	   Implements docker.ContainerWorkspace defined in docker/ir.go

	   This workspace generates a docker-compose file at the root of the
	   output directory.  The docker-compose instantiates containers
	   that are either:
	    (a) pre-built images
	    (b) artifacts built using Dockerfiles in the output directory

	*/
	dockerComposeWorkspace struct {
		ir.VisitTrackerImpl

		info docker.ContainerWorkspaceInfo

		ImageDirs    map[string]string      // map from image name to directory
		InstanceArgs map[string][]ir.IRNode // argnodes for each instance added to the workspace

		DockerComposeFile *dockergen.DockerComposeFile
	}
)

// Implements ir.ArtifactGenerator
func (node *Deployment) GenerateArtifacts(dir string) error {
	slog.Info(fmt.Sprintf("Collecting container instances for deployment %s in %s", node.Name(), dir))
	workspace := NewDockerComposeWorkspace(node.Name(), dir)
	return node.generateArtifacts(workspace)
}

/*
The basic build process of a docker-compose deployment
*/
func (node *Deployment) generateArtifacts(workspace *dockerComposeWorkspace) error {

	// Add any locally-built container images
	for _, node := range ir.Filter[docker.ProvidesContainerImage](node.Nodes) {
		if err := node.AddContainerArtifacts(workspace); err != nil {
			return err
		}
	}

	// Collect all container instances
	for _, node := range ir.Filter[docker.ProvidesContainerInstance](node.Nodes) {
		if err := node.AddContainerInstance(workspace); err != nil {
			return err
		}
	}

	// Build the docker-compose file
	if err := workspace.Finish(); err != nil {
		return err
	}

	return nil
}

func NewDockerComposeWorkspace(name string, dir string) *dockerComposeWorkspace {
	return &dockerComposeWorkspace{
		info: docker.ContainerWorkspaceInfo{
			Path:   filepath.Clean(dir),
			Target: "docker-compose",
		},
		ImageDirs:         make(map[string]string),
		InstanceArgs:      make(map[string][]ir.IRNode),
		DockerComposeFile: dockergen.NewDockerComposeFile(name, dir, "docker-compose.yml"),
	}
}

// Implements docker.ContainerWorkspace
func (d *dockerComposeWorkspace) Info() docker.ContainerWorkspaceInfo {
	return d.info
}

// Implements docker.ContainerWorkspace
func (d *dockerComposeWorkspace) CreateImageDir(imageName string) (string, error) {
	// Only alphanumeric and underscores are allowed in an proc name
	imageName = ir.CleanName(imageName)
	imageDir, err := ioutil.CreateNodeDir(d.info.Path, imageName)
	d.ImageDirs[imageName] = imageDir
	return imageDir, err
}

// Implements docker.ContainerWorkspace
func (d *dockerComposeWorkspace) DeclarePrebuiltInstance(instanceName string, image string, args ...ir.IRNode) error {
	d.InstanceArgs[instanceName] = args
	return d.DockerComposeFile.AddImageInstance(instanceName, image)
}

// Implements docker.ContainerWorkspace
func (d *dockerComposeWorkspace) DeclareLocalImage(instanceName string, imageDir string, args ...ir.IRNode) error {
	d.InstanceArgs[instanceName] = args
	return d.DockerComposeFile.AddBuildInstance(instanceName, imageDir)
}

// Implements docker.ContainerWorkspace
func (d *dockerComposeWorkspace) SetEnvironmentVariable(instanceName string, key string, val string) error {
	return d.DockerComposeFile.AddEnvVar(instanceName, key, val)
}

// Generates the docker-compose file
func (d *dockerComposeWorkspace) Finish() error {
	// We didn't set any arguments or environment variables while accumulating instances. Do so now.
	if err := d.processArgNodes(); err != nil {
		return err
	}

	// Now that all images and instances have been declared, we can generate the docker-compose file
	return d.DockerComposeFile.Generate()
}

// Goes through each container's arg nodes, determining which need to be passed to the container
// as environment variables.
//
// Has special handling for addresses; containers that bind a server will have ports assigned,
// and containers that dial to a server within this namespace will have the dial address set.
//
// We don't pick external-facing ports for any addresses; these will be set by the caller or user.
func (d *dockerComposeWorkspace) processArgNodes() error {
	addresses := make(map[string]string)
	for instanceName, instanceArgs := range d.InstanceArgs {
		binds, _, remaining := address.Split(instanceArgs)

		// First handle the non-address arguments to the node, which will need to be passed
		// through as environment variables.
		// The only special handling is that if a config node already has a value set on it,
		// then we don't need to pass the value at all, because we can assume that the value
		// will be hard-coded inside the container.
		for _, arg := range remaining {
			switch node := arg.(type) {
			case ir.IRConfig:
				if !node.HasValue() {
					d.DockerComposeFile.PassthroughEnvVar(instanceName, node.Name(), node.Optional())
				}
			default:
				return blueprint.Errorf("container instance %v can only accept IRConfig nodes as arguments, but found %v of type %v", instanceName, arg, reflect.TypeOf(arg))
			}
		}

		// Some of the ports within this container might not yet be assigned; do so now.
		// Any ports that we assign will need to be passed into the container as environment
		// variables so that the server knows what port to bind to.
		_, assigned, err := address.AssignPorts(binds)
		if err != nil {
			return err
		}
		for _, bind := range assigned {
			d.DockerComposeFile.AddEnvVar(instanceName, bind.Name(), fmt.Sprintf("0.0.0.0:%v", bind.Port))
		}

		// All ports need to be exposed in the docker-compose file in order to be accessible
		// to other containers.  We then save the addresses so that other containers can dial
		// to them.
		for _, bind := range binds {
			hostname := ir.CleanName(instanceName)
			addresses[bind.AddressName] = fmt.Sprintf("%v:%v", hostname, bind.Port)
			d.DockerComposeFile.ExposePort(instanceName, bind.Port)
		}

		// The default logic for the docker-compose file is for the user to set environment
		// variables so that internal container ports are bound to external-facing ports on the
		// host machine.  These ports are chosen by the user at runtime and set by env vars.
		for _, bind := range binds {
			d.DockerComposeFile.MapPortToEnvVar(instanceName, bind.Port, bind.Name())
		}
		address.Clear(binds)
	}

	// Now that we know the local addresses of all servers bound within this workspace, set
	// all dials.  Dials to local servers can have the address set directly; dials to servers
	// that don't exist within this workspace will need to be passed through as an env var.
	for instanceName, instanceArgs := range d.InstanceArgs {
		_, dials, _ := address.Split(instanceArgs)
		for _, dial := range dials {
			if addr, isLocalDial := addresses[dial.AddressName]; isLocalDial {
				d.DockerComposeFile.AddEnvVar(instanceName, dial.Name(), addr)
			} else {
				d.DockerComposeFile.PassthroughEnvVar(instanceName, dial.Name(), false)
			}
		}
	}

	return nil
}

func (d *dockerComposeWorkspace) ImplementsBuildContext()       {}
func (d *dockerComposeWorkspace) ImplementsContainerWorkspace() {}
