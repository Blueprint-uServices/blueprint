package workspace

import (
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer/linuxgen"
)

/*
Most of the time, processes will be packaged into a Docker container.

This is the Docker ipmlementation of the linux.ProcessWorkspace defined
in linux/ir.go.

This workspace performs the same actions as the BasicWorkspace but also
allows processes to optionally provide Dockerfile build commands in lieu
of adding a build script to the build.sh
*/

type DockerWorkspaceImpl struct {
	BasicWorkspace

	Dockerfile *linuxgen.Dockerfile
}

func NewDockerWorkspace(name string, dir string) *DockerWorkspaceImpl {
	ws := &DockerWorkspaceImpl{}
	ws.BasicWorkspace = *NewBasicWorkspace(name, dir)
	ws.Dockerfile = linuxgen.NewDockerfile(dir)
	return ws
}

func (ws *DockerWorkspaceImpl) AddDockerfileCommands(procName, commands string) error {
	return nil
}

/*
Generates the build.sh and run.sh as well as Dockerfile
*/
func (ws *DockerWorkspaceImpl) Finish() error {
	// Create the BasicWorkspace's build.sh and run.sh
	if err := ws.BasicWorkspace.Finish(); err != nil {
		return err
	}

	// Additionally generate the dockerfile
	return ws.Dockerfile.Generate()
}

func (ws *DockerWorkspaceImpl) ImplementsDockerProcessWorkspace() {}
