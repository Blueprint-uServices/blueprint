package workspace

import (
	"path/filepath"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ioutil"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/dockerapp/dockergen"
)

/*
Deploys a set of containers as a docker-compose application.

Implements docker.ContainerWorkspace defined in docker/ir.go

This workspace generates a docker-compose file at the root of the
output directory.  The docker-compose instantiates containers
that are either:
 (a) pre-built images
 (b) artifacts built using Dockerfiles in the output directory

*/

type DockerComposeWorkspace struct {
	blueprint.VisitTrackerImpl

	info docker.ContainerWorkspaceInfo

	ImageDirs map[string]string // map from image name to directory

	Compose *dockergen.DockerComposeFile
}

func NewDockerComposeWorkspace(name string, dir string) *DockerComposeWorkspace {
	return &DockerComposeWorkspace{
		info: docker.ContainerWorkspaceInfo{
			Path:   filepath.Clean(dir),
			Target: "docker-compose",
		},
		ImageDirs: make(map[string]string),
		Compose:   dockergen.NewDockerComposeFile(name, dir, "docker-compose.yml"),
	}
}

func (d *DockerComposeWorkspace) Info() docker.ContainerWorkspaceInfo {
	return d.info
}

func (d *DockerComposeWorkspace) CreateImageDir(imageName string) (string, error) {
	// Only alphanumeric and underscores are allowed in an proc name
	imageName = blueprint.CleanName(imageName)

	// Can't redefine an image that already exists
	if _, exists := d.ImageDirs[imageName]; exists {
		return "", blueprint.Errorf("image dir %v already exists in output container workspace %v", imageName, d.info.Path)
	}

	// Create the dir
	imageDir := filepath.Join(d.info.Path, imageName)
	if err := ioutil.CheckDir(imageDir, true); err != nil {
		return "", blueprint.Errorf("cannot generate image to output workspace %v due to %v", imageName, err.Error())
	}
	d.ImageDirs[imageName] = imageDir
	return imageDir, nil
}

func (d *DockerComposeWorkspace) DeclarePrebuiltInstance(instanceName string, image string) error {
	return nil
}

func (d *DockerComposeWorkspace) DeclareLocalImage(instanceName string, imageName string) error {
	return nil
}

func (d *DockerComposeWorkspace) Finish() error {
	// Now that all images and instances have been declared, we can generate the docker-compose file
	return d.Compose.Generate()
}

func (d *DockerComposeWorkspace) ImplementsBuildContext()       {}
func (d *DockerComposeWorkspace) ImplementsContainerWorkspace() {}
