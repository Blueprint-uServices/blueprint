package specs

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/examples/dsb_hotel/cmplx_workload/workloadgen"
	"github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow/hotelreservation"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/gotests"
	"github.com/blueprint-uservices/blueprint/plugins/grpc"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/jaeger"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
	"github.com/blueprint-uservices/blueprint/plugins/memcached"
	"github.com/blueprint-uservices/blueprint/plugins/mongodb"
	"github.com/blueprint-uservices/blueprint/plugins/opentelemetry"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	"github.com/blueprint-uservices/blueprint/plugins/workload"
)

// Wiring spec that represents the original configuration of the HotelReservation application.
// Each service is deployed in a separate container with all inter-service communication happening via GRPC.
// FrontEnd service provides a http frontend for making requests.
// All services are instrumented with opentelemetry tracing with spans being exported to a central Jaeger collector.
var Original = cmdbuilder.SpecOption{
	Name:        "original",
	Description: "Deploys the original configuration of the DeathStarBench application.",
	Build:       makeOriginalSpec,
}

func makeOriginalSpec(spec wiring.WiringSpec) ([]string, error) {
	var cntrs []string

	var allServices []string
	// Define backends
	trace_collector := jaeger.Collector(spec, "jaeger")
	user_db := mongodb.Container(spec, "user_db")
	recommendations_db := mongodb.Container(spec, "recomd_db")
	reserv_db := mongodb.Container(spec, "reserv_db")
	geo_db := mongodb.Container(spec, "geo_db")
	rate_db := mongodb.Container(spec, "rate_db")
	profile_db := mongodb.Container(spec, "profile_db")

	reserv_cache := memcached.Container(spec, "reserv_cache")
	rate_cache := memcached.Container(spec, "rate_cache")
	profile_cache := memcached.Container(spec, "profile_cache")

	// Define internal services
	user_service := workflow.Service[hotelreservation.UserService](spec, "user_service", user_db)
	user_ctr := applyDefaults(spec, user_service, trace_collector)
	cntrs = append(cntrs, user_ctr)
	allServices = append(allServices, "user_service")

	recomd_service := workflow.Service[hotelreservation.RecommendationService](spec, "recomd_service", recommendations_db)
	recomd_ctr := applyDefaults(spec, recomd_service, trace_collector)
	cntrs = append(cntrs, recomd_ctr)
	allServices = append(allServices, "recomd_service")

	reserv_service := workflow.Service[hotelreservation.ReservationService](spec, "reserv_service", reserv_cache, reserv_db)
	reserv_ctr := applyDefaults(spec, reserv_service, trace_collector)
	cntrs = append(cntrs, reserv_ctr)
	allServices = append(allServices, "reserv_service")

	geo_service := workflow.Service[hotelreservation.GeoService](spec, "geo_service", geo_db)
	geo_ctr := applyDefaults(spec, geo_service, trace_collector)
	cntrs = append(cntrs, geo_ctr)
	allServices = append(allServices, "geo_service")

	rate_service := workflow.Service[hotelreservation.RateService](spec, "rate_service", rate_cache, rate_db)
	rate_ctr := applyDefaults(spec, rate_service, trace_collector)
	cntrs = append(cntrs, rate_ctr)
	allServices = append(allServices, "rate_service")

	profile_service := workflow.Service[hotelreservation.ProfileService](spec, "profile_service", profile_cache, profile_db)
	profile_ctr := applyDefaults(spec, profile_service, trace_collector)
	cntrs = append(cntrs, profile_ctr)
	allServices = append(allServices, "profile_service")

	search_service := workflow.Service[hotelreservation.SearchService](spec, "search_service", geo_service, rate_service)
	search_ctr := applyDefaults(spec, search_service, trace_collector)
	cntrs = append(cntrs, search_ctr)
	allServices = append(allServices, "search_service")

	// Define frontend service
	frontend_service := workflow.Service[hotelreservation.FrontEndService](spec, "frontend_service", search_service, profile_service, recomd_service, user_service, reserv_service)
	frontend_ctr := applyHTTPDefaults(spec, frontend_service, trace_collector)
	cntrs = append(cntrs, frontend_ctr)
	allServices = append(allServices, "frontend_service")

	wlgen := workload.Generator[workloadgen.ComplexWorkload](spec, "wlgen", frontend_service)
	cntrs = append(cntrs, wlgen)

	tests := gotests.Test(spec, allServices...)
	cntrs = append(cntrs, tests)

	return cntrs, nil
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
