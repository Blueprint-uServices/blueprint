package opentelemetry

import (
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

// Interface that indicates if an IRNode implements the OTCollector interface
// All custom collector clients **must** implement this interface
type OpenTelemetryCollectorInterface interface {
	golang.Node
	golang.Instantiable
	ImplementsOTCollectorClient()
}
