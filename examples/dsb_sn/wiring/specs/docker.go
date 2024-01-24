package specs

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/examples/dsb_sn/workflow/socialnetwork"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/gotests"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
	"github.com/blueprint-uservices/blueprint/plugins/memcached"
	"github.com/blueprint-uservices/blueprint/plugins/mongodb"
	"github.com/blueprint-uservices/blueprint/plugins/thrift"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

// A wiring spec that deploys each service into its own Docker container and using thrift to communicate between services.
// All services except the Wrk2API service use thrift for communication; WRK2API service provides the http frontend.
// The user, socialgraph, urlshorten, and usertimeline services use MongoDB instances to store their data.
// The user, socialgraph, urlshorten, usertimeine, and hometimeline services use memcached instances as the cache data for faster responses.
var Docker = cmdbuilder.SpecOption{
	Name:        "docker",
	Description: "Deploys each service in a separate container with thrift, and uses mongodb as NoSQL database backends.",
	Build:       makeDockerSpec,
}

// Create a basic social network wiring spec.
// Returns the names of the nodes to instantiate or an error.
func makeDockerSpec(spec wiring.WiringSpec) ([]string, error) {
	var containers []string
	var allServices []string

	// Define backends
	user_cache := memcached.Container(spec, "user_cache")
	user_db := mongodb.Container(spec, "user_db")
	post_cache := memcached.Container(spec, "post_cache")
	post_db := mongodb.Container(spec, "post_db")
	social_cache := memcached.Container(spec, "social_cache")
	social_db := mongodb.Container(spec, "social_db")
	urlshorten_db := mongodb.Container(spec, "urlshorten_db")
	usertimeline_cache := memcached.Container(spec, "usertimeline_cache")
	usertimeline_db := mongodb.Container(spec, "usertimeline_db")
	hometimeline_cache := memcached.Container(spec, "hometimeline_cache")

	// Add backends to services list so that their client libraries are used in the generated tests!
	allServices = append(allServices, user_cache)
	allServices = append(allServices, user_db)
	allServices = append(allServices, post_cache)
	allServices = append(allServices, post_db)
	allServices = append(allServices, social_cache)
	allServices = append(allServices, social_db)
	allServices = append(allServices, usertimeline_cache)
	allServices = append(allServices, usertimeline_db)
	allServices = append(allServices, hometimeline_cache)

	// Define url_shorten service
	urlshorten_service := workflow.Service[socialnetwork.UrlShortenService](spec, "urlshorten_service", urlshorten_db)
	urlshorten_ctr := applyDockerDefaults(spec, urlshorten_service, "urlshorten_proc", "urlshorten_container")
	containers = append(containers, urlshorten_ctr)
	allServices = append(allServices, "urlshorten_service")

	// Define user_mention service
	usermention_service := workflow.Service[socialnetwork.UserMentionService](spec, "usermention_service", user_cache, user_db)
	usermention_ctr := applyDockerDefaults(spec, usermention_service, "usermention_proc", "usermention_container")
	containers = append(containers, usermention_ctr)
	allServices = append(allServices, "usermention_service")

	// Define post_storage service
	post_storage_service := workflow.Service[socialnetwork.PostStorageService](spec, "post_storage_service", post_cache, post_db)
	post_storage_ctr := applyDockerDefaults(spec, post_storage_service, "post_storage_proc", "post_storage_container")
	containers = append(containers, post_storage_ctr)
	allServices = append(allServices, "post_storage_service")

	// Define media service
	media_service := workflow.Service[socialnetwork.MediaService](spec, "media_service")
	media_ctr := applyDockerDefaults(spec, media_service, "media_proc", "media_container")
	containers = append(containers, media_ctr)
	allServices = append(allServices, "media_service")

	// Define uniqueid service
	uniqueId_service := workflow.Service[socialnetwork.UniqueIdService](spec, "uniqueid_service")
	uniqueId_ctr := applyDockerDefaults(spec, uniqueId_service, "uniqueid_proc", "uniqueid_container")
	containers = append(containers, uniqueId_ctr)
	allServices = append(allServices, "uniqueid_service")

	// Define user_id service
	userid_service := workflow.Service[socialnetwork.UserIDService](spec, "userid_service", user_cache, user_db)
	userid_ctr := applyDockerDefaults(spec, userid_service, "userid_proc", "userid_container")
	containers = append(containers, userid_ctr)
	allServices = append(allServices, "userid_service")

	// Define social_graph service
	socialgraph_service := workflow.Service[socialnetwork.SocialGraphService](spec, "socialgraph_service", social_cache, social_db, userid_service)
	socialgraph_ctr := applyDockerDefaults(spec, socialgraph_service, "socailgraph_proc", "socialgraph_container")
	containers = append(containers, socialgraph_ctr)
	allServices = append(allServices, "socialgraph_service")

	// Define home_timeline service
	hometimeline_service := workflow.Service[socialnetwork.HomeTimelineService](spec, "hometimeline_service", hometimeline_cache, post_storage_service, socialgraph_service)
	hometimeline_ctr := applyDockerDefaults(spec, hometimeline_service, "hometimeline_proc", "hometimeline_container")
	containers = append(containers, hometimeline_ctr)
	allServices = append(allServices, "hometimeline_service")

	// Define user service
	user_service := workflow.Service[socialnetwork.UserService](spec, "user_service", user_cache, user_db, socialgraph_service, "secret")
	user_ctr := applyDockerDefaults(spec, user_service, "user_proc", "user_container")
	containers = append(containers, user_ctr)
	allServices = append(allServices, "user_service")

	// Define text service
	text_service := workflow.Service[socialnetwork.TextService](spec, "text_service", urlshorten_service, usermention_service)
	text_ctr := applyDockerDefaults(spec, text_service, "text_proc", "text_container")
	containers = append(containers, text_ctr)
	allServices = append(allServices, "text_service")

	// Define user_timeline service
	usertimeline_service := workflow.Service[socialnetwork.UserTimelineService](spec, "usertimeline_service", usertimeline_cache, usertimeline_db, post_storage_service)
	usertimeline_ctr := applyDockerDefaults(spec, usertimeline_service, "usertimeline_proc", "usertimeline_container")
	containers = append(containers, usertimeline_ctr)
	allServices = append(allServices, "usertimeline_service")

	// Define compose post service
	composepost_service := workflow.Service[socialnetwork.ComposePostService](spec, "composepost_service", post_storage_service, usertimeline_service, user_service, uniqueId_service, media_service, text_service, hometimeline_service)
	compose_ctr := applyDockerDefaults(spec, composepost_service, "compose_proc", "compose_container")
	containers = append(containers, compose_ctr)
	allServices = append(allServices, "composepost_service")

	// Define frontend service
	wrk2api_service := workflow.Service[socialnetwork.Wrk2APIService](spec, "wrk2api_service", user_service, composepost_service, usertimeline_service, hometimeline_service, socialgraph_service)
	wrk2api_ctr := applyHTTPDefaults(spec, wrk2api_service, "wrk2api_proc", "wrk2api_container")
	containers = append(containers, wrk2api_ctr)
	allServices = append(allServices, "wrk2api_service")

	tests := gotests.Test(spec, allServices...)
	containers = append(containers, tests)

	return containers, nil
}

func applyDockerDefaults(spec wiring.WiringSpec, serviceName, procName, ctrName string) string {
	thrift.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}

func applyHTTPDefaults(spec wiring.WiringSpec, serviceName, procName, ctrName string) string {
	http.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}
