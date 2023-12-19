package specs

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/gotests"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/http"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/mongodb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

// A wiring spec that deploys each service into its own Docker container and uses http to communicate between services.
// The user service uses MongoDB instance to store their data.
var Docker = wiringcmd.SpecOption{
	Name:        "docker",
	Description: "Deploys each service in a separate container with http, and uses mongodb as NoSQL database backends",
	Build:       makeDockerSpec,
}

// Create a basic train ticket wiring spec.
// Returns the names of the nodes to instantiate or an error.
func makeDockerSpec(spec wiring.WiringSpec) ([]string, error) {
	var containers []string
	user_db := mongodb.Container(spec, "user_db")
	user_service := workflow.Service(spec, "user_service", "UserServiceImpl", user_db)
	user_cntr := applyDockerDefaults(spec, user_service, "user_proc", "user_container")
	allServices := []string{user_service}

	containers = append(containers, user_cntr)

	contacts_db := mongodb.Container(spec, "contacts_db")
	contacts_service := workflow.Service(spec, "contacts_service", "ContactsServiceImpl", contacts_db)
	contacts_cntr := applyDockerDefaults(spec, contacts_service, "contacts_proc", "contacts_container")
	allServices = append(allServices, contacts_service)
	containers = append(containers, contacts_cntr)

	price_db := mongodb.Container(spec, "price_db")
	price_service := workflow.Service(spec, "price_service", "PriceServiceImpl", price_db)
	price_cntr := applyDockerDefaults(spec, price_service, "price_proc", "price_container")
	allServices = append(allServices, price_service)
	containers = append(containers, price_cntr)

	tests := gotests.Test(spec, allServices...)
	containers = append(containers, tests)
	return containers, nil
}

func applyDockerDefaults(spec wiring.WiringSpec, serviceName, procName, ctrName string) string {
	http.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}
