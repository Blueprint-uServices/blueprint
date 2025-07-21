package specs

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/jaeger"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	"github.com/blueprint-uservices/blueprint/plugins/memcached"
	"github.com/blueprint-uservices/blueprint/plugins/mongodb"
	"github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow/hotelreservation"
	"github.com/blueprint-uservices/blueprint/examples/dsb_hotel/cmplx_workload/workloadgen"
	"github.com/blueprint-uservices/blueprint/plugins/workload"
	"github.com/blueprint-uservices/blueprint/plugins/crisp"
)

var Chain = cmdbuilder.SpecOption{
	Name:        "chain",
	Description: "Chain topology: Frontend -> Search -> Geo (other dependencies are dummies)",
	Build:       makeChainSpec,
}

func makeChainSpec(spec wiring.WiringSpec) ([]string, error) {
	var cntrs []string

	// Add critical path analysis service as a container
	analysisContainer := crisp.Container(spec, "trace_analysis")
	cntrs = append(cntrs, analysisContainer)

	trace_collector := jaeger.Collector(spec, "jaeger")
	user_db := mongodb.Container(spec, "user_db")
	recommendation_db := mongodb.Container(spec, "recommendation_db")
	reservation_db := mongodb.Container(spec, "reservation_db")
	geo_db := mongodb.Container(spec, "geo_db")
	profile_db := mongodb.Container(spec, "profile_db")
	profile_cache := memcached.Container(spec, "profile_cache")
	reservation_cache := memcached.Container(spec, "reservation_cache")
	rate_db := mongodb.Container(spec, "rate_db")
	rate_cache := memcached.Container(spec, "rate_cache")

	// Geo Service
	geo_service := workflow.Service[hotelreservation.GeoService](spec, "geo_service", geo_db)
	geo_ctr := applyDefaults(spec, geo_service, trace_collector)
	cntrs = append(cntrs, geo_ctr)

	// Rate Service (dummy for dependency)
	rate_service := workflow.Service[hotelreservation.RateService](spec, "rate_service", rate_cache, rate_db)
	rate_ctr := applyDefaults(spec, rate_service, trace_collector)
	cntrs = append(cntrs, rate_ctr)

	// Search Service (depends on Geo and Rate)
	search_service := workflow.Service[hotelreservation.SearchService](spec, "search_service", geo_service, rate_service)
	search_ctr := applyDefaults(spec, search_service, trace_collector)
	cntrs = append(cntrs, search_ctr)

	// Dummy services for required dependencies
	profile_service := workflow.Service[hotelreservation.ProfileService](spec, "profile_service", profile_cache, profile_db)
	profile_ctr := applyDefaults(spec, profile_service, trace_collector)
	cntrs = append(cntrs, profile_ctr)

	recommendation_service := workflow.Service[hotelreservation.RecommendationService](spec, "recommendation_service", recommendation_db)
	recommendation_ctr := applyDefaults(spec, recommendation_service, trace_collector)
	cntrs = append(cntrs, recommendation_ctr)

	user_service := workflow.Service[hotelreservation.UserService](spec, "user_service", user_db)
	user_ctr := applyDefaults(spec, user_service, trace_collector)
	cntrs = append(cntrs, user_ctr)

	reservation_service := workflow.Service[hotelreservation.ReservationService](spec, "reservation_service", reservation_cache, reservation_db)
	reservation_ctr := applyDefaults(spec, reservation_service, trace_collector)
	cntrs = append(cntrs, reservation_ctr)

	// Frontend Service (all dependencies)
	frontend_service := workflow.Service[hotelreservation.FrontEndService](
		spec, "frontend_service",
		search_service, profile_service, recommendation_service, user_service, reservation_service,
	)
	frontend_ctr := applyHTTPDefaults(spec, frontend_service, trace_collector)
	cntrs = append(cntrs, frontend_ctr)

	// Add workload generator
	wlgen := workload.Generator[workloadgen.ComplexWorkload](spec, "wlgen", frontend_service)
	cntrs = append(cntrs, wlgen)

	return cntrs, nil
} 