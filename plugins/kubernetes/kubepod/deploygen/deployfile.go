package deploygen

import (
	"fmt"
	"path/filepath"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/kubernetes/kubetemplate"
	"github.com/blueprint-uservices/blueprint/plugins/linux"
	"golang.org/x/exp/slog"
)

type KubeDeploymentFile struct {
	WorkspaceName   string
	WorkspaceDir    string
	FileName        string
	ServiceFilename string
	FilePath        string
	Instances       map[string]instance
	localServers    map[string]*address.BindConfig
	localDials      map[string]*address.DialConfig
}

type instance struct {
	InstanceName string
	Image        string
	Ports        map[string]uint16
	Config       map[string]string
	Passthrough  map[string]struct{}
}

func NewKubeDeploymentFile(workspaceName string, workspaceDir string, filename string, serviceFilename string) *KubeDeploymentFile {
	return &KubeDeploymentFile{
		WorkspaceName:   workspaceName,
		WorkspaceDir:    workspaceDir,
		FileName:        filename,
		ServiceFilename: serviceFilename,
		FilePath:        filepath.Join(workspaceDir, filename),
		Instances:       make(map[string]instance),
	}
}

func (k *KubeDeploymentFile) Generate() error {
	k.ResolveLocalDials()
	slog.Info(fmt.Sprintf("Generating %v/%v", k.WorkspaceName, k.FileName))
	err := kubetemplate.ExecuteTemplateToFile("kubedeployment", kubernetesTemplate, k, k.FilePath)
	if err != nil {
		return err
	}
	serviceFilePath := filepath.Join(k.WorkspaceDir, k.ServiceFilename)
	return kubetemplate.ExecuteTemplateToFile("kubedeployment", kubernetesServiceTemplate, k, serviceFilePath)
}

func (k *KubeDeploymentFile) AddImageInstance(instanceName string, image string, args ...ir.IRNode) error {
	return k.addInstance(instanceName, image, args...)
}

func (k *KubeDeploymentFile) AddEnvVar(instanceName string, key string, val string) error {
	instanceName = ir.CleanName(instanceName)
	if i, exists := k.Instances[instanceName]; !exists {
		return blueprint.Errorf("container instance with name %v not found", instanceName)
	} else {
		i.Config[key] = val
		k.Instances[instanceName] = i
	}
	return nil
}

func (k *KubeDeploymentFile) addInstance(instanceName string, image string, args ...ir.IRNode) error {
	instanceName = ir.CleanName(instanceName)
	if _, exists := k.Instances[instanceName]; exists {
		return blueprint.Errorf("re-declaration of container instance %v of image %v", instanceName, image)
	}
	instance := instance{
		InstanceName: instanceName,
		Image:        image,
		Ports:        make(map[string]uint16),
		Config:       make(map[string]string),
	}
	for _, node := range args {
		varname := linux.EnvVar(node.Name())

		if bind, isBindConf := node.(*address.BindConfig); isBindConf {
			if bind.Port == 0 {
				return blueprint.Errorf("cannot add container instance %v due to unbound server port %v", instanceName, bind.Name())
			}
			instance.Ports[requiredEnvVar(node)] = bind.Port
		}

		if conf, isConfig := node.(ir.IRConfig); isConfig {
			if conf.HasValue() {
				instance.Config[varname] = conf.Value()
				continue
			} else if conf.Optional() {
				instance.Passthrough[varname] = struct{}{}
				continue
			}
		}
		instance.Config[varname] = requiredEnvVar(node)
	}

	k.Instances[instanceName] = instance

	k.checkForAddrs(args)

	return nil
}

func (d *KubeDeploymentFile) checkForAddrs(nodes []ir.IRNode) {
	for _, node := range nodes {
		switch c := node.(type) {
		case *address.BindConfig:
			d.localServers[c.AddressName] = c
		case *address.DialConfig:
			d.localDials[c.AddressName] = c
		}
	}
}

func (d *KubeDeploymentFile) ResolveLocalDials() error {
	for name, bind := range d.localServers {
		dial, hasLocalDial := d.localDials[name]
		if !hasLocalDial {
			continue
		}

		// Update the configured value for any instance that uses this dial addr
		// to point it directly towards the local server
		dialVarname := linux.EnvVar(dial.Name())
		for _, instance := range d.Instances {
			if _, hasConfig := instance.Config[dialVarname]; hasConfig {
				instance.Config[dialVarname] = bind.Value()
			}
		}
	}
	return nil
}

func requiredEnvVar(node ir.IRNode) string {
	return fmt.Sprintf("${%v?%v must be set by the calling environment}", linux.EnvVar(node.Name()), node.Name())
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
          {{- range $external, $internal := .Ports}}
            - containerPort: {{$internal}}
              hostIP: 0.0.0.0
              hostPort: {{$external}}
          {{- end}}
      {{-end}}
      restartPolicy: Always
      hostname: {{.InstanceName}}
`

var kubernetesServiceTemplate = `
apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}-service
spec:
  selector:
    blueprint.service: {{.Name}}
  ports:
  {{- range $external, $internal := .Ports}}
    - port: {{$internal}}
      targetPort: {{$external}}
  {{- end}}
`
