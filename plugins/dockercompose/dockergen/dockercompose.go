package dockergen

import (
	"fmt"
	"path/filepath"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/linux"
	"golang.org/x/exp/slog"
)

/*
Used for generating the docker-compose file of a docker app
*/
type DockerComposeFile struct {
	WorkspaceName string
	WorkspaceDir  string
	FileName      string
	FilePath      string
	Instances     map[string]*instance           // Container instance declarations
	localServers  map[string]*address.BindConfig // Servers that have been defined within this docker-compose file
	localDials    map[string]*address.DialConfig // All servers that will be dialed from within this docker-compose file
}

type instance struct {
	InstanceName      string
	ContainerTemplate string              // only used if built; empty if not
	Image             string              // only used by prebuilt; empty if not
	Ports             map[string]uint16   // Map from bindconfig name to internal port
	Expose            map[uint16]struct{} // Ports exposed with expose directive
	Config            map[string]string   // Map from environment variable name to value
	Passthrough       map[string]struct{} // Environment variables that just get passed through to the container
}

func NewDockerComposeFile(workspaceName, workspaceDir, fileName string) *DockerComposeFile {
	return &DockerComposeFile{
		WorkspaceName: workspaceName,
		WorkspaceDir:  workspaceDir,
		FileName:      fileName,
		FilePath:      filepath.Join(workspaceDir, fileName),
		Instances:     make(map[string]*instance),
		localServers:  make(map[string]*address.BindConfig),
		localDials:    make(map[string]*address.DialConfig),
	}
}

func (d *DockerComposeFile) Generate() error {
	slog.Info(fmt.Sprintf("Generating %v/%v", d.WorkspaceName, d.FileName))
	return ExecuteTemplateToFile("docker-compose", dockercomposeTemplate, d, d.FilePath)

}

// Adds an instance to the docker-compose file, that will use an off-the-shelf image.
//
// The instanceName is chosen by the user; it can subsequently be passed in methods such as [AddEnvVar],
// [PassthroughEnvVar], [ExposePort], [MapPort], and [MapPortToEnvVar].
func (d *DockerComposeFile) AddImageInstance(instanceName string, image string) error {
	return d.addInstance(instanceName, image, "")
}

// Adds an instance to the docker-compose file, that will be built from a container template
// on the local filesystem
//
// The instanceName is chosen by the user; it can subsequently be passed in methods such as [AddEnvVar],
// [PassthroughEnvVar], [ExposePort], [MapPort], and [MapPortToEnvVar].
func (d *DockerComposeFile) AddBuildInstance(instanceName string, containerTemplateName string) error {
	return d.addInstance(instanceName, "", containerTemplateName)
}

func (d *DockerComposeFile) getInstance(instanceName string) (*instance, error) {
	instanceName = ir.CleanName(instanceName)
	if i, exists := d.Instances[instanceName]; exists {
		return i, nil
	}
	return nil, blueprint.Errorf("container instance with name %v not found", instanceName)
}

// Sets an environment variable key to the specified val for instanceName
func (d *DockerComposeFile) AddEnvVar(instanceName string, key string, val string) error {
	instance, err := d.getInstance(instanceName)
	if err != nil {
		return err
	}
	key = linux.EnvVar(key)
	instance.Config[key] = val
	return nil
}

// Pass through the specified environment variable key from the calling environment
func (d *DockerComposeFile) PassthroughEnvVar(instanceName string, key string, optional bool) error {
	var passthroughValue string
	if optional {
		passthroughValue = fmt.Sprintf("${%v:-}", linux.EnvVar(key))
	} else {
		passthroughValue = fmt.Sprintf("${%v?%v must be set by the calling environment}", linux.EnvVar(key), key)
	}
	return d.AddEnvVar(instanceName, key, passthroughValue)
}

// Exposes a container-internal port for use by other containers within the docker-compose file
func (d *DockerComposeFile) ExposePort(instanceName string, internalPort uint16) error {
	instance, err := d.getInstance(instanceName)
	if err != nil {
		return err
	}
	instance.Expose[internalPort] = struct{}{}
	return nil
}

// Further to [ExposePort], adds a Port directive so that the host machine can access the internalPort
// of the container via the externalAddress.  Typically externalAddress will be a localhost or 0.0.0.0
// address
func (d *DockerComposeFile) MapPort(instanceName string, internalPort uint16, externalAddress string) error {
	instance, err := d.getInstance(instanceName)
	if err != nil {
		return err
	}
	instance.Ports[externalAddress] = internalPort
	return nil
}

// Further to [ExposePort], adds a Port directive so that the host machine can access the internalPort
// of the container, using a runtime substitution of envVarName as the externalAddress
func (d *DockerComposeFile) MapPortToEnvVar(instanceName string, internalPort uint16, envVarName string) error {
	externalAddress := fmt.Sprintf("${%v?%v must be set by the calling environment}", linux.EnvVar(envVarName), envVarName)
	return d.MapPort(instanceName, internalPort, externalAddress)
}

func (d *DockerComposeFile) addInstance(instanceName string, image string, containerTemplateName string) error {
	instanceName = ir.CleanName(instanceName)
	if _, exists := d.Instances[instanceName]; exists {
		return blueprint.Errorf("re-declaration of container instance %v of image %v", instanceName, image)
	}
	instance := instance{
		InstanceName:      instanceName,
		ContainerTemplate: containerTemplateName,
		Image:             image,
		Expose:            make(map[uint16]struct{}),
		Ports:             make(map[string]uint16),
		Config:            make(map[string]string),
		Passthrough:       make(map[string]struct{}),
	}
	d.Instances[instanceName] = &instance
	return nil
}

var dockercomposeTemplate = `
version: '3'
services:
{{range $_, $decl := .Instances}}
  {{.InstanceName}}:
    {{if .Image -}}
    image: {{.Image}}
    {{- else if .ContainerTemplate -}}
    build:
      context: {{.ContainerTemplate}}
      dockerfile: ./Dockerfile
    {{- end}}
    hostname: {{.InstanceName}}
    {{- if .Ports}}
    expose:
    {{- range $internal, $_ := .Expose}}
     - "{{$internal}}"
    {{- end}}
    ports:
    {{- range $external, $internal := .Ports}}
     - "{{$external}}:{{$internal}}"
    {{- end}}
    {{- end}}
    {{- if .Config}}
    environment:
    {{- range $name, $value := .Config}}
     - {{$name}}={{$value}}
    {{- end}}
    {{- end}}
    restart: always
{{end}}
`
