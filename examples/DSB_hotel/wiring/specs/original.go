package specs

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
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
	user_db := mongodb.PrebuiltContainer(spec, "user_db")
	recommendations_db := mongodb.PrebuiltContainer(spec, "recomd_db")
	reserv_db := mongodb.PrebuiltContainer(spec, "reserv_db")
	geo_db := mongodb.PrebuiltContainer(spec, "geo_db")
	rate_db := mongodb.PrebuiltContainer(spec, "rate_db")
	profile_db := mongodb.PrebuiltContainer(spec, "profile_db")

	reserv_cache := memcached.PrebuiltContainer(spec, "reserv_cache")
	rate_cache := memcached.PrebuiltContainer(spec, "rate_cache")
	profile_cache := memcached.PrebuiltContainer(spec, "profile_cache")

	// Define internal services
	user_service := workflow.Define(spec, "user_service", "UserService", user_db)
	user_ctr := applyDefaults(spec, user_service, trace_collector)
	cntrs = append(cntrs, user_ctr)

	recomd_service := workflow.Define(spec, "recomd_service", "RecommendationService", recommendations_db)
	recomd_ctr := applyDefaults(spec, recomd_service, trace_collector)
	cntrs = append(cntrs, recomd_ctr)

	reserv_service := workflow.Define(spec, "reserv_service", "ReservationService", reserv_cache, reserv_db)
	reserv_ctr := applyDefaults(spec, reserv_service, trace_collector)
	cntrs = append(cntrs, reserv_ctr)

	geo_service := workflow.Define(spec, "geo_service", "GeoService", geo_db)
	geo_ctr := applyDefaults(spec, geo_service, trace_collector)
	cntrs = append(cntrs, geo_ctr)

	rate_service := workflow.Define(spec, "rate_service", "RateService", rate_cache, rate_db)
	rate_ctr := applyDefaults(spec, rate_service, trace_collector)
	cntrs = append(cntrs, rate_ctr)

	profile_service := workflow.Define(spec, "profile_service", "ProfileService", profile_cache, profile_db)
	profile_ctr := applyDefaults(spec, profile_service, trace_collector)
	cntrs = append(cntrs, profile_ctr)

	search_service := workflow.Define(spec, "search_service", "SearchService", geo_service, rate_service)
	search_ctr := applyDefaults(spec, search_service, trace_collector)
	cntrs = append(cntrs, search_ctr)

	// Define frontend service
	frontend_service := workflow.Define(spec, "frontend_service", "FrontEndService", search_service, profile_service, recomd_service, user_service, reserv_service)
	frontend_ctr := applyDefaults(spec, frontend_service, trace_collector)
	cntrs = append(cntrs, frontend_ctr)

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
