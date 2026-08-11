package specs

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/examples/dsb_mm/workflow/media"
	"github.com/blueprint-uservices/blueprint/examples/dsb_mm/workload/workloadgen"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/jaeger"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
	"github.com/blueprint-uservices/blueprint/plugins/memcached"
	"github.com/blueprint-uservices/blueprint/plugins/mongodb"
	"github.com/blueprint-uservices/blueprint/plugins/opentelemetry"
	"github.com/blueprint-uservices/blueprint/plugins/redis"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	"github.com/blueprint-uservices/blueprint/plugins/workload"
)

var Default = cmdbuilder.SpecOption{
	Name:        "default",
	Description: "Deploys the system as HTTP servers in docker containers with Jaeger tracing enabled.",
	Build:       makeDefaultSpec,
}

func makeDefaultSpec(spec wiring.WiringSpec) ([]string, error) {
	var cntrs []string

	// Define backends
	trace_collector := jaeger.Collector(spec, "jaeger")

	castinfo_cache := memcached.Container(spec, "castinfo_cache")
	movieid_cache := memcached.Container(spec, "movieid_cache")
	user_cache := memcached.Container(spec, "user_cache")
	user_review_cache := redis.Container(spec, "userreview_cache")
	movie_review_cache := redis.Container(spec, "moviereview_cache")
	review_storage_cache := memcached.Container(spec, "reviewstorage_cache")
	rating_cache := redis.Container(spec, "rating_cache")
	compose_review_cache := memcached.Container(spec, "composereview_cache")
	plot_cache := memcached.Container(spec, "plot_cache")
	movieinfo_cache := memcached.Container(spec, "movieinfo_cache")

	castinfo_db := mongodb.Container(spec, "castinfo_db")
	movieid_db := mongodb.Container(spec, "movieid_db")
	user_db := mongodb.Container(spec, "user_db")
	user_review_db := mongodb.Container(spec, "userreview_db")
	movie_review_db := mongodb.Container(spec, "moviereview_db")
	review_storage_db := mongodb.Container(spec, "reviewstorage_db")
	plot_db := mongodb.Container(spec, "plot_db")
	movieinfo_db := mongodb.Container(spec, "movieinfo_db")

	// Define internal services

	castinfo_service := workflow.Service[media.CastInfoService](spec, "castinfo_service", castinfo_cache, castinfo_db)
	castinfo_ctr := applyHTTPDefaults(spec, castinfo_service, trace_collector)
	cntrs = append(cntrs, castinfo_ctr)

	reviewstorage_service := workflow.Service[media.ReviewStorageService](spec, review_storage_cache, review_storage_db)
	reviewstorage_ctr := applyHTTPDefaults(spec, reviewstorage_service, trace_collector)
	cntrs = append(cntrs, reviewstorage_ctr)

	userreview_service := workflow.Service[media.UserReviewService](spec, "userreview_service", reviewstorage_service, user_review_db, user_review_cache)
	userreview_ctr := applyHTTPDefaults(spec, userreview_service, trace_collector)
	cntrs = append(cntrs, userreview_ctr)

	moviereview_service := workflow.Service[media.MovieReviewService](spec, "moviereview_service", reviewstorage_service, movie_review_db, movie_review_cache)
	moviereview_ctr := applyHTTPDefaults(spec, moviereview_service, trace_collector)
	cntrs = append(cntrs, moviereview_ctr)

	composereview_service := workflow.Service[media.ComposeReviewService](spec, "composereview_service", compose_review_cache, reviewstorage_service, userreview_service, moviereview_service)
	composereview_ctr := applyHTTPDefaults(spec, composereview_service, trace_collector)
	cntrs = append(cntrs, composereview_ctr)

	rating_service := workflow.Service[media.RatingService](spec, "rating_service", composereview_service, rating_cache)
	rating_ctr := applyHTTPDefaults(spec, rating_service, trace_collector)
	cntrs = append(cntrs, rating_ctr)

	text_service := workflow.Service[media.TextService](spec, "text_service", composereview_service)
	text_ctr := applyHTTPDefaults(spec, text_service, trace_collector)
	cntrs = append(cntrs, text_ctr)

	user_service := workflow.Service[media.UserService](spec, "user_service", user_cache, user_db, composereview_service, "secret")
	user_ctr := applyHTTPDefaults(spec, user_service, trace_collector)
	cntrs = append(cntrs, user_ctr)

	plot_service := workflow.Service[media.PlotService](spec, "plot_service", plot_cache, plot_db)
	plot_ctr := applyHTTPDefaults(spec, plot_service, trace_collector)
	cntrs = append(cntrs, plot_ctr)

	movieid_service := workflow.Service[media.MovieIdService](spec, "movieid_service", movieid_cache, movieid_db, rating_service, composereview_service)
	movieid_ctr := applyHTTPDefaults(spec, movieid_service, trace_collector)
	cntrs = append(cntrs, movieid_ctr)

	movieinfo_service := workflow.Service[media.MovieInfoService](spec, "movieinfo_service", movieinfo_cache, movieinfo_db)
	movieinfo_ctr := applyHTTPDefaults(spec, movieinfo_service, trace_collector)
	cntrs = append(cntrs, movieinfo_ctr)

	page_service := workflow.Service[media.PageService](spec, "page_service", movieinfo_service, moviereview_service, castinfo_service, plot_service)
	page_ctr := applyHTTPDefaults(spec, page_service, trace_collector)
	cntrs = append(cntrs, page_ctr)

	uniqueid_service := workflow.Service[media.UniqueIdService](spec, "uniqueid_service", composereview_service)
	uniqueid_ctr := applyHTTPDefaults(spec, uniqueid_service, trace_collector)
	cntrs = append(cntrs, uniqueid_ctr)

	wrk2api_service := workflow.Service[media.Wrk2APIService](spec, "wrk2api_service", user_service, castinfo_service, text_service, plot_service, movieid_service, movieinfo_service, uniqueid_service)
	wrk2api_ctr := applyHTTPDefaults(spec, wrk2api_service, trace_collector)
	cntrs = append(cntrs, wrk2api_ctr)

	wlgen := workload.Generator[workloadgen.MediaWorkload](spec, "wlgen", wrk2api_service)
	cntrs = append(cntrs, wlgen)

	return cntrs, nil
}

func applyHTTPDefaults(spec wiring.WiringSpec, serviceName string, collectorName string) string {
	procName := fmt.Sprintf("%s_process", serviceName)
	ctrName := fmt.Sprintf("%s_container", serviceName)
	opentelemetry.InstrumentWithoutClientSpans(spec, serviceName, collectorName)
	http.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}
