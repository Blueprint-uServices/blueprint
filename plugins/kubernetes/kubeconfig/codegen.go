package kubeconfig

import (
	"log/slog"
	"path/filepath"

	"github.com/blueprint-uservices/blueprint/plugins/kubernetes/kubetemplate"
)

type KubeConfigFile struct {
	Name     string
	Filename string
	FilePath string
	Data     map[string]string
}

func NewKubeConfigFile(configName string, filename string, workspaceDir string) *KubeConfigFile {
	return &KubeConfigFile{
		Name:     configName,
		Filename: filename,
		FilePath: filepath.Join(workspaceDir, filename),
		Data:     make(map[string]string),
	}
}

// AddData adds a key-value pair to the config map manifest
// If the key already exists, AddData overwrites the value for the existing key with the newly provided value.
func (k *KubeConfigFile) AddData(key string, val string) {
	k.Data[key] = val
}

func (k *KubeConfigFile) Generate() error {
	slog.Info("Generating %v/%v", k.Name, k.Filename)
	return kubetemplate.ExecuteTemplateToFile("kubeconfig", kubeconfigTemplate, k, k.FilePath)
}

var kubeconfigTemplate = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{.Name}}
data:
  {{- range $k, $v := .Data}}
  {{$k}}: {{$v}}
  {{- end -}}
`
