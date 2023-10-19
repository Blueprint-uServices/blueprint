package dockergen

import (
	"fmt"
	"path/filepath"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linux"
	"golang.org/x/exp/slog"
)

/*
Generates the docker-compose file of a docker app
*/

type DockerComposeFile struct {
	WorkspaceName string
	WorkspaceDir  string
	FileName      string
	FilePath      string
	Instances     map[string]instance // Container instance declarations
}

type instance struct {
	InstanceName      string
	ContainerTemplate string            // only used if built; empty if not
	Image             string            // only used by prebuilt; empty if not
	Ports             map[int]int       // Map from external port to internal port
	Config            map[string]string // Map from environment variable name to value
}

func NewDockerComposeFile(workspaceName, workspaceDir, fileName string) *DockerComposeFile {
	return &DockerComposeFile{
		WorkspaceName: workspaceName,
		WorkspaceDir:  workspaceDir,
		FileName:      fileName,
		FilePath:      filepath.Join(workspaceDir, fileName),
		Instances:     make(map[string]instance),
	}
}

func (d *DockerComposeFile) Generate() error {
	slog.Info(fmt.Sprintf("Generating %v/%v", d.WorkspaceName, d.FileName))
	return ExecuteTemplateToFile("docker-compose", dockercomposeTemplate, d, d.FilePath)

}

func (d *DockerComposeFile) AddImageInstance(instanceName string, image string, args ...blueprint.IRNode) error {
	return d.addInstance(instanceName, image, "", args...)
}

func (d *DockerComposeFile) AddBuildInstance(instanceName string, containerTemplateName string, args ...blueprint.IRNode) error {
	return d.addInstance(instanceName, "", containerTemplateName, args...)
}

func (d *DockerComposeFile) addInstance(instanceName string, image string, containerTemplateName string, args ...blueprint.IRNode) error {
	if _, exists := d.Instances[instanceName]; exists {
		return blueprint.Errorf("re-declaration of container instance %v of image %v", instanceName, image)
	}
	instance := instance{
		InstanceName:      instanceName,
		ContainerTemplate: containerTemplateName,
		Image:             image,
		Ports:             make(map[int]int),
		Config:            make(map[string]string),
	}
	// TODO: assign ports etc properly
	//   for now just pull from environment
	for _, node := range args {
		varname := linux.EnvVar(node.Name())
		instance.Config[varname] = fmt.Sprintf("${%v}", varname)
	}
	d.Instances[instanceName] = instance
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
    restart: always
    {{- end}}
{{end}}
`
