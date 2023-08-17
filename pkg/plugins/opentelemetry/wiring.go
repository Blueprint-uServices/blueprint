package opentelemetry

var DefaultOpenTelemetryCollectorName = "opentelemetry.collector"

// func initOpenTelemetry(wiring *blueprint.WiringSpec, collectorName string) {
// 	collectorAddr := collectorName + ".server.addr"
// 	wiring.Define(collectorAddr, &blueprint.Address{ /* TODO visibility? */ }, func(scope blueprint.Scope) (blueprint.IRNode, error) {
// 		return newNetworkAddress()
// 	})

// 	collector := collectorName + ".server"
// 	wiring.Define(collector, &OpenTelemetryCollector{}, func(scope blueprint.Scope) (blueprint.IRNode, error) {
// 		addr := scope.Get(collectorAddr)
// 		return newOpenTelemetryCollector(collector, addr), nil
// 	})

// 	client := collectorName + ".client"
// 	wiring.Define(client, &OpenTelemetryClient{}, func(scope blueprint.Scope) (blueprint.IRNode, error) {
// 		addr := scope.Get(collectorAddr)
// 		return newOpenTelemetryClient(client, addr), nil
// 	})
// }

// func defineClientsideOT(wiring *blueprint.WiringSpec, serviceName, downstream string) {
// 	otClientName := "opentelemetry.collector.client"
// 	serviceWithOT := serviceName + ".client.opentelemetry"
// 	wiring.Define(serviceWithOT, &InstrumentedClientsideService, func(scope blueprint.Scope) (blueprint.IRNode, error) {
// 		otClient := scope.Get(otClientName)
// 		wrappedService := scope.Get(downstream)
// 		return newInstrumentedService(wrappedService, otClient)
// 	})
// }

// func defineServersideOT(wiring *blueprint.WiringSpec, serviceName, handlerName string) {
// 	wrappedHandlerName := serviceName + ".server.opentelemetry"
// 	wiring.Define(serviceWithOT, &InstrumentedServersideService, func(scope blueprint.Scope) (blueprint.IRNode, error) {
// 		otClient := scope.Get("opentelemetry.collector.client")
// 		handler := scope.Get(handlerName)
// 		return newInstrumentedService(wrappedHandlerName, otClient, handler)
// 	})
// 	return wrappedHandlerName
// }

// // Instruments a service with OpenTelemetry.  This will do the following:
// //   - Instantiate the OpenTelemetry collector process and define client libraries
// //   - On the server side, it will wrap the handler with an extended interface that creates spans and then proxies calls
// //   - On the client side, it will intercept calls to the server and create spans before proxying calls
// func Instrument(wiring *blueprint.WiringSpec, serviceName string) {
// 	// Define the OpenTelemetry collector process and client library
// 	initOpenTelemetry(wiring, DefaultOpenTelemetryCollectorName)

// 	addr, handler := wiring.GetPointer(serviceName)

// 	wrappedHandler := defineServersideOT(wiring, serviceName, handler)

// 	// explicitly define the addr between client and server, with some name
// 	// define the clientside with the addr as the downstream
// 	// redefine the previous addr as the clientside
// 	// readvertise the service pointer as pointing to the wrapped handler via the client and server handlers
// 	defineClientsideOT(wiring, serviceName, addr)

// 	wiring.MakePointer(serviceName, wrappedHandler)

// 	// Define the client-side and server-side wrapper classes
// 	defineClientsideOT(wiring, serviceName, addr)
// }
