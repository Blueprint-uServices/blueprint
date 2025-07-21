package specs

import (
	"fmt"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/grpc"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
	"github.com/blueprint-uservices/blueprint/plugins/opentelemetry"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
)

// Helper functions for minimal fake DB and service wiring
func fakeDB(spec wiring.WiringSpec, name string) string {
	return name // placeholder for a DB container if needed
}

func workflowServiceGeo(spec wiring.WiringSpec, geo_db string, collector string) string {
	serviceName := "geo_service"
	procName := serviceName + "_proc"
	ctrName := serviceName + "_ctr"
	opentelemetry.Instrument(spec, serviceName, collector)
	grpc.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}

func workflowServiceSearch(spec wiring.WiringSpec, geo_service string, collector string) string {
	serviceName := "search_service"
	procName := serviceName + "_proc"
	ctrName := serviceName + "_ctr"
	opentelemetry.Instrument(spec, serviceName, collector)
	grpc.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}

func workflowServiceProfile(spec wiring.WiringSpec, profile_cache, profile_db, collector string) string {
	serviceName := "profile_service"
	procName := serviceName + "_proc"
	ctrName := serviceName + "_ctr"
	opentelemetry.Instrument(spec, serviceName, collector)
	grpc.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}

func workflowServiceProfileWithGeo(spec wiring.WiringSpec, geo_service, profile_cache, profile_db, collector string) string {
	// For fanin, profile_service depends on geo_service
	return workflowServiceProfile(spec, profile_cache, profile_db, collector)
}

func workflowServiceFrontendChain(spec wiring.WiringSpec, search_service string, collector string) string {
	serviceName := "frontend_service"
	procName := serviceName + "_proc"
	ctrName := serviceName + "_ctr"
	opentelemetry.Instrument(spec, serviceName, collector)
	http.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}

func workflowServiceFrontendFanout(spec wiring.WiringSpec, search_service, profile_service, collector string) string {
	serviceName := "frontend_service"
	procName := serviceName + "_proc"
	ctrName := serviceName + "_ctr"
	opentelemetry.Instrument(spec, serviceName, collector)
	http.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}

func workflowServiceFrontendFanin(spec wiring.WiringSpec, search_service, profile_service, collector string) string {
	serviceName := "frontend_service"
	procName := serviceName + "_proc"
	ctrName := serviceName + "_ctr"
	opentelemetry.Instrument(spec, serviceName, collector)
	http.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}

func applyDefaults(spec wiring.WiringSpec, serviceName string, collectorName string) string {
	procName := fmt.Sprintf("%s_process", serviceName)
	ctrName := fmt.Sprintf("%s_container", serviceName)
	opentelemetry.Instrument(spec, serviceName, collectorName)
	grpc.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}

func applyHTTPDefaults(spec wiring.WiringSpec, serviceName string, collectorName string) string {
	procName := fmt.Sprintf("%s_process", serviceName)
	ctrName := fmt.Sprintf("%s_container", serviceName)
	opentelemetry.Instrument(spec, serviceName, collectorName)
	http.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
} 