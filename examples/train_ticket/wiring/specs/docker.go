package specs

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/assurance"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/config"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/consignprice"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/contacts"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/delivery"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/news"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/payment"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/price"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/route"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/station"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/stationfood"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/train"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/trainfood"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/user"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/gotests"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
	"github.com/blueprint-uservices/blueprint/plugins/mongodb"
	"github.com/blueprint-uservices/blueprint/plugins/rabbitmq"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

// A wiring spec that deploys each service into its own Docker container and uses http to communicate between services.
// The user service uses MongoDB instance to store their data.
var Docker = cmdbuilder.SpecOption{
	Name:        "docker",
	Description: "Deploys each service in a separate container with http, and uses mongodb as NoSQL database backends",
	Build:       makeDockerSpec,
}

// Create a basic train ticket wiring spec.
// Returns the names of the nodes to instantiate or an error.
func makeDockerSpec(spec wiring.WiringSpec) ([]string, error) {
	var containers []string
	var allServices []string
	applyDockerDefaults := func(serviceName, procName, ctrName string) {
		http.Deploy(spec, serviceName)
		goproc.CreateProcess(spec, procName, serviceName)
		linuxcontainer.CreateContainer(spec, ctrName, procName)
		allServices = append(allServices, serviceName)
		containers = append(containers, ctrName)
	}
	user_db := mongodb.Container(spec, "user_db")
	user_service := workflow.Service[*user.UserServiceImpl](spec, "user_service", user_db)
	applyDockerDefaults(user_service, "user_proc", "user_container")

	contacts_db := mongodb.Container(spec, "contacts_db")
	contacts_service := workflow.Service[*contacts.ContactsServiceImpl](spec, "contacts_service", contacts_db)
	applyDockerDefaults(contacts_service, "contacts_proc", "contacts_container")

	price_db := mongodb.Container(spec, "price_db")
	price_service := workflow.Service[*price.PriceServiceImpl](spec, "price_service", price_db)
	applyDockerDefaults(price_service, "price_proc", "price_container")

	station_db := mongodb.Container(spec, "station_db")
	station_service := workflow.Service[*station.StationServiceImpl](spec, "station_service", station_db)
	applyDockerDefaults(station_service, "station_proc", "station_container")

	news_service := workflow.Service[*news.NewsServiceImpl](spec, "news_service")
	applyDockerDefaults(news_service, "news_proc", "news_container")

	assurance_db := mongodb.Container(spec, "assurance_db")
	assurance_service := workflow.Service[*assurance.AssuranceServiceImpl](spec, "assurance_service", assurance_db)
	applyDockerDefaults(assurance_service, "assurance_proc", "assurance_container")

	config_db := mongodb.Container(spec, "config_db")
	config_service := workflow.Service[*config.ConfigServiceImpl](spec, "config_service", config_db)
	applyDockerDefaults(config_service, "config_proc", "config_container")

	consignprice_db := mongodb.Container(spec, "consignprice_db")
	consignprice_service := workflow.Service[*consignprice.ConsignPriceServiceImpl](spec, "consignprice_service", consignprice_db)
	applyDockerDefaults(consignprice_service, "consignprice_proc", "consignprice_container")

	payments_db := mongodb.Container(spec, "payments_db")
	money_db := mongodb.Container(spec, "money_db")
	payments_service := workflow.Service[*payment.PaymentServiceImpl](spec, "payments_service", payments_db, money_db)
	applyDockerDefaults(payments_service, "payments_proc", "payments_container")

	route_db := mongodb.Container(spec, "route_db")
	route_service := workflow.Service[*route.RouteServiceImpl](spec, "route_service", route_db)
	applyDockerDefaults(route_service, "route_proc", "route_container")

	stationfood_db := mongodb.Container(spec, "stationfood_db")
	stationfood_service := workflow.Service[*stationfood.StationFoodServiceImpl](spec, "stationfood_service", stationfood_db)
	applyDockerDefaults(stationfood_service, "stationfood_proc", "stationfood_container")

	trainfood_db := mongodb.Container(spec, "trainfood_db")
	trainfood_service := workflow.Service[*trainfood.TrainFoodServiceImpl](spec, "trainfood_service", trainfood_db)
	applyDockerDefaults(trainfood_service, "trainfood_proc", "trainfood_container")

	train_db := mongodb.Container(spec, "train_db")
	train_service := workflow.Service[*train.TrainServiceImpl](spec, "train_service", train_db)
	applyDockerDefaults(train_service, "train_proc", "train_container")

	delivery_queue := rabbitmq.Container(spec, "delivery_q", "delivery_q")
	delivery_db := mongodb.Container(spec, "delivery_db")
	delivery_service := workflow.Service[*delivery.DeliveryServiceImpl](spec, "delivery_service", delivery_queue, delivery_db)
	goproc.CreateProcess(spec, "delivery_proc", delivery_service)
	linuxcontainer.CreateContainer(spec, "delivery_container", "delivery_proc")
	containers = append(containers, "delivery_container")

	tests := gotests.Test(spec, allServices...)
	containers = append(containers, tests)
	return containers, nil
}
