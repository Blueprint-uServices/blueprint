package linuxcontainer

import (
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer/dockergen"
)

/*
A collection of process can be combined into a Docker image.

To do this requires, beyond the regular things, generating a Dockerfile
*/

type (
	/*
	   The docker deployer for linux containers extends the default deployer,
	   in that it collects and packages the process artifacts for the container.
	   In addition, it then creates a Dockerfile for the container.

	   The Dockerfile will copy all process artifacts into the container.
	   By default, the Dockerfile will then call the build.sh from within the
	   container.

	   Each process within the container can optionally provide additional
	   commands to add to the Dockerfile, if implemented.
	*/

	dockerDeployer interface {
		docker.Container
		docker.ProvidesContainerImage
		docker.ProvidesContainerInstance
	}

	/*
	   Most of the time, processes will be packaged into a Docker container.

	   This is the Docker ipmlementation of the linux.ProcessWorkspace defined
	   in linux/ir.go.

	   This workspace performs the same actions as the BasicWorkspace but also
	   allows processes to optionally provide Dockerfile build commands in lieu
	   of adding a build script to the build.sh
	*/

	dockerWorkspaceImpl struct {
		filesystemWorkspace

		Dockerfile *dockergen.Dockerfile
	}
)

func (node *Container) AddContainerArtifacts(target docker.ContainerWorkspace) error {
	// The image only needs to be created in the output directory once
	if target.Visited(node.ImageName) {
		return nil
	}

	// Create a new subdirectory to construct the image
	dir, err := target.CreateImageDir(node.ImageName)
	if err != nil {
		return err
	}

	// Generate artifacts to the image directory in the normal way.
	// By providing a docker workspace, processes will be able to
	// add dockerfile commands
	// The docker workspace extends the Finish() implementation
	// to also generate the Dockerfile
	workspace := NewDockerWorkspace(node.Name(), dir)
	if err := node.generateArtifacts(workspace); err != nil {
		return err
	}
	return nil
}

func (node *Container) AddContainerInstance(target docker.ContainerWorkspace) error {
	// The instance only needs to be added to the output directory once
	if target.Visited(node.InstanceName) {
		return nil
	}

	// TODO: all address and port related shenanigans will need to go here

	return target.DeclareLocalImage(node.InstanceName, node.ImageName)
}

func NewDockerWorkspace(name string, dir string) *dockerWorkspaceImpl {
	ws := &dockerWorkspaceImpl{}
	ws.info.Target = "docker"
	ws.filesystemWorkspace = *NewBasicWorkspace(name, dir)
	ws.Dockerfile = dockergen.NewDockerfile(name, dir)
	return ws
}

func (ws *dockerWorkspaceImpl) AddDockerfileCommands(procName, commands string) {
	ws.Dockerfile.AddCustomCommands(procName, commands)
}

/*
Generates the build.sh and run.sh as well as Dockerfile
*/
func (ws *dockerWorkspaceImpl) Finish() error {
	// Create the BasicWorkspace's build.sh and run.sh
	if err := ws.filesystemWorkspace.Finish(); err != nil {
		return err
	}

	// Additionally generate the dockerfile
	return ws.Dockerfile.Generate(ws.ProcDirs)
}

func (ws *dockerWorkspaceImpl) ImplementsDockerProcessWorkspace() {}
func (node *Container) ImplementsDockerContainer()                {}
