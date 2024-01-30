// Package kubepod is a plugin for instantiating multiple container instances in a single Kubernetes pod deployment.
//
// # Wiring Spec Usage
//
// To use the kubepod plugin in your wiring spec, you can declare a Kubernetes Pod Deployment, giving it a name and specifying which container instances to include
//
//	kubepod.NewKubePod(spec, "my_pod", "my_container_1", "my_container_2")
//
// You can add containers to existing pods:
//
//	kubepod.AddContainerToPod(spec, "my_pod", "my_container_3")
//
// To deploy an application-level service in a Kubernetes Pod, make sure you first deploy the service to a process (with the [goproc] plugin) and to a container image (with the [linuxcontainer] plugin)
//
// # Artifacts Generated
//
// During compilation, the plugin generates a `podName-deployment.yaml` file that instantiates the pod as a Kubernetes deployment and a `podName-service.yaml` file that converts the deployed pod into a Kubernetes service.
//
// # Running Artifacts
//
// You need to have a working kubernetes cluster and `kubectl` installed.
// To deploy the pods to the cluster, use the following commands:
//
//	kubectl apply -f podName-deployment.yaml
//	kubectl apply -f podName-service.yaml
//
// [linuxcontainer]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/linuxcontainer
// [goproc]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/goproc
package kubepod

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/namespaceutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
)

// [AddContainerToPod] can be used by wiring specs to add more containers to a pod
func AddContainerToPod(spec wiring.WiringSpec, podName string, containerName string) {
	namespaceutil.AddNodeTo[PodDeployment](spec, podName, containerName)
}

// [NewKubePod] can be used by wiring specs to create a Kubernetes Pod that instantiates a single Kubernetes Pod consisting of multiple containers.
//
// Further containers can be added to the Pod by calling [AddContainerToPod].
//
// During compilation, generates the deployment.yaml and service.yaml files for the pod.
//
// Returns podName
func NewKubePod(spec wiring.WiringSpec, podName string, containers ...string) string {

	// If any children were provided in this call, add them to the pod via a property
	for _, containerName := range containers {
		AddContainerToPod(spec, podName, containerName)
	}

	spec.Define(podName, &PodDeployment{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		pod := &PodDeployment{PodName: podName}
		_, err := namespaceutil.InstantiateNamespace(ns, pod)
		return pod, err
	})

	return podName
}
