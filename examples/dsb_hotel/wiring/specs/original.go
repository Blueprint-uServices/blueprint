package specs

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/gotests"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/http"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/jaeger"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/memcached"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/mongodb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/opentelemetry"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

var Original = wiringcmd.SpecOption{
	Name:        "original",
	Description: "Deploys the original configuration of the DeathStarBench application.",
	Build:       makeOriginalSpec,
}

func makeOriginalSpec(spec wiring.WiringSpec) ([]string, error) {
	var cntrs []string
	// Define backends
	trace_collector := jaeger.DefineJaegerCollector(spec, "jaeger")
	user_db := mongodb.Container(spec, "user_db")
	recommendations_db := mongodb.Container(spec, "recomd_db")
	reserv_db := mongodb.Container(spec, "reserv_db")
	geo_db := mongodb.Container(spec, "geo_db")
	rate_db := mongodb.Container(spec, "rate_db")
	profile_db := mongodb.Container(spec, "profile_db")

	reserv_cache := memcached.PrebuiltContainer(spec, "reserv_cache")
	rate_cache := memcached.PrebuiltContainer(spec, "rate_cache")
	profile_cache := memcached.PrebuiltContainer(spec, "profile_cache")

	// Define internal services
	user_service := workflow.Service(spec, "user_service", "UserService", user_db)
	user_ctr := applyDefaults(spec, user_service, trace_collector)
	cntrs = append(cntrs, user_ctr)

	recomd_service := workflow.Service(spec, "recomd_service", "RecommendationService", recommendations_db)
	recomd_ctr := applyDefaults(spec, recomd_service, trace_collector)
	cntrs = append(cntrs, recomd_ctr)

	reserv_service := workflow.Service(spec, "reserv_service", "ReservationService", reserv_cache, reserv_db)
	reserv_ctr := applyDefaults(spec, reserv_service, trace_collector)
	cntrs = append(cntrs, reserv_ctr)

	geo_service := workflow.Service(spec, "geo_service", "GeoService", geo_db)
	geo_ctr := applyDefaults(spec, geo_service, trace_collector)
	cntrs = append(cntrs, geo_ctr)

	rate_service := workflow.Service(spec, "rate_service", "RateService", rate_cache, rate_db)
	rate_ctr := applyDefaults(spec, rate_service, trace_collector)
	cntrs = append(cntrs, rate_ctr)

	profile_service := workflow.Service(spec, "profile_service", "ProfileService", profile_cache, profile_db)
	profile_ctr := applyDefaults(spec, profile_service, trace_collector)
	cntrs = append(cntrs, profile_ctr)

	search_service := workflow.Service(spec, "search_service", "SearchService", geo_service, rate_service)
	search_ctr := applyDefaults(spec, search_service, trace_collector)
	cntrs = append(cntrs, search_ctr)

	// Define frontend service
	frontend_service := workflow.Service(spec, "frontend_service", "FrontEndService", search_service, profile_service, recomd_service, user_service, reserv_service)
	frontend_ctr := applyHTTPDefaults(spec, frontend_service, trace_collector)
	cntrs = append(cntrs, frontend_ctr)

	tests := gotests.Test(spec, frontend_service)
	cntrs = append(cntrs, tests)

	return cntrs, nil
}

func applyDefaults(spec wiring.WiringSpec, serviceName string, collectorName string) string {
	procName := fmt.Sprintf("%s_process", serviceName)
	ctrName := fmt.Sprintf("%s_container", serviceName)
	opentelemetry.InstrumentUsingCustomCollector(spec, serviceName, collectorName)
	grpc.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}

func applyHTTPDefaults(spec wiring.WiringSpec, serviceName string, collectorName string) string {
	procName := fmt.Sprintf("%s_process", serviceName)
	ctrName := fmt.Sprintf("%s_container", serviceName)
	opentelemetry.InstrumentUsingCustomCollector(spec, serviceName, collectorName)
	http.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}
