package dockergen

import (
	"fmt"
	"path/filepath"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linux"
)

/*
Generates the docker-compose file of a docker app
*/

type DockerComposeFile struct {
	WorkspaceName string
	WorkspaceDir  string
	FileName      string
	FilePath      string
	Instances     map[string]string // Container instance declarations
}

func NewDockerComposeFile(workspaceName, workspaceDir, fileName string) *DockerComposeFile {
	return &DockerComposeFile{
		WorkspaceName: workspaceName,
		WorkspaceDir:  workspaceDir,
		FileName:      fileName,
		FilePath:      filepath.Join(workspaceDir, fileName),
	}
}

func (d *DockerComposeFile) Generate() error {
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
	templateArgs := instanceTemplateArgs{
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
		templateArgs.Config[varname] = fmt.Sprintf("${%v}", varname)
	}
	decl, err := ExecuteTemplate("declareDockerComposeInstance", instanceTemplate, templateArgs)
	d.Instances[instanceName] = decl
	return err
}

type instanceTemplateArgs struct {
	InstanceName      string
	ContainerTemplate string            // only used if built; empty if not
	Image             string            // only used by prebuilt; empty if not
	Ports             map[int]int       // Map from external port to internal port
	Config            map[string]string // Map from environment variable name to value
}

var instanceTemplate = `
  {{.InstanceName}}:
    {{if .Image -}}
    image: {{.Image}}
	{{- else if .ContainerTemplate -}}
    build:
      context: ./{{.ContainerTemplate}}/
      dockerfile: {{.ContainerTemplate}}/Dockerfile
	{{- end}}
	hostname: {{.InstanceName}}
	ports:
	  {{- range $external, $internal := .Ports -}}
      - "{{$external}}:{{$internal}}"
	  {{end -}}
	environment:
	  {{- range $name, $value := .Config }}
	  - {{$name}}={{$value}}
	  {{end -}}
	restart: always`

var dockercomposeTemplate = `
version: '3'
services:
{{range .Instances}}
{{.}}
{{end}}
`
