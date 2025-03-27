package deploygen

import (
	"fmt"
	"path/filepath"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/kubernetes/kubetemplate"
	"github.com/blueprint-uservices/blueprint/plugins/linux"
	"golang.org/x/exp/slog"
)

type KubeDeploymentFile struct {
	Name            string
	WorkspaceDir    string
	FileName        string
	ServiceFilename string
	FilePath        string
	NumReplicas     int64
	Instances       map[string]*instance
}

type instance struct {
	InstanceName string
	Image        string
	Ports        map[string]uint16
	Config       map[string]string
}

func NewKubeDeploymentFile(workspaceName string, workspaceDir string, filename string, serviceFilename string) *KubeDeploymentFile {
	return &KubeDeploymentFile{
		Name:            workspaceName,
		WorkspaceDir:    workspaceDir,
		FileName:        filename,
		ServiceFilename: serviceFilename,
		FilePath:        filepath.Join(workspaceDir, filename),
		Instances:       make(map[string]*instance),
		// For now NumReplicas is fixed.
		NumReplicas: 1,
	}
}

func (k *KubeDeploymentFile) Generate() error {
	slog.Info(fmt.Sprintf("Generating %v/%v", k.Name, k.FileName))
	err := kubetemplate.ExecuteTemplateToFile("kubedeployment", kubernetesTemplate, k, k.FilePath)
	if err != nil {
		return err
	}
	slog.Info("NUmber of instances: ", "num", len(k.Instances))
	serviceFilePath := filepath.Join(k.WorkspaceDir, k.ServiceFilename)
	slog.Info(fmt.Sprintf("Generating %v/%v", k.Name, k.ServiceFilename))
	return kubetemplate.ExecuteTemplateToFile("kubedeployment", kubernetesServiceTemplate, k, serviceFilePath)
}

func (k *KubeDeploymentFile) AddImageInstance(instanceName string, image string) error {
	return k.addInstance(instanceName, image)
}

func (k *KubeDeploymentFile) getInstance(instanceName string) (*instance, error) {
	instanceName = ir.CleanName(instanceName)
	if i, exists := k.Instances[instanceName]; exists {
		return i, nil
	} else {
		return nil, blueprint.Errorf("container instance with name %v not found", instanceName)
	}
}

func (k *KubeDeploymentFile) AddEnvVar(instanceName string, key string, val string) error {
	key = linux.EnvVar(key)
	instance, err := k.getInstance(instanceName)
	if err != nil {
		return err
	} else {
		instance.Config[key] = val
		return nil
	}
}

func (k *KubeDeploymentFile) ExposePort(instanceName string, portName string, port uint16) error {
	instance, err := k.getInstance(instanceName)
	if err != nil {
		return err
	} else {
		instance.Ports[portName] = port
		return nil
	}
}

func (k *KubeDeploymentFile) addInstance(instanceName string, image string) error {
	instanceName = ir.CleanName(instanceName)
	if _, exists := k.Instances[instanceName]; exists {
		return blueprint.Errorf("re-declaration of container instance %v of image %v", instanceName, image)
	}
	instance := &instance{
		InstanceName: instanceName,
		Image:        image,
		Ports:        make(map[string]uint16),
		Config:       make(map[string]string),
	}
	k.Instances[instanceName] = instance
	return nil
}

var kubernetesTemplate = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Name}}
  labels:
    blueprint.service: {{.Name}}
spec:
  replicas: {{.NumReplicas}}
  selector:
    matchLabels:
      blueprint.service: {{.Name}}
  template:
    metadata:
      name: {{.Name}}
      labels:
        blueprint.service: {{.Name}}
    spec:
      containers:
      {{range $_, $decl := .Instances}}
        - name: {{.InstanceName}}
          image: {{.Image}}
          {{- if .Config}}
          env:
          {{- range $name, $value := .Config}}
            - name: {{$name}}
              value: "{{$value}}"
          {{- end}}
          {{- end}}
          {{- if .Ports}}
          ports:
          {{- range $name, $port := .Ports}}
            - containerPort: {{$port}}
          {{- end}}
		  {{- end}}
      {{- end}}
      restartPolicy: Always
      hostname: {{.Name}}
`

var kubernetesServiceTemplate = `
apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}
spec:
  selector:
    blueprint.service: {{.Name}}
  ports:
  {{range $_, $decl := .Instances}}
  {{- range $name, $port := .Ports}}
    - name: {{$name}}
	  port: {{$port}}
      targetPort: {{$port}}
  {{- end}}
  {{- end}}
`
