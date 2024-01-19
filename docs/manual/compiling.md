# Compiling a Blueprint Application

After writing a wiring spec, we compile it by running it.  If you are making use of the [cmdbuilder](../../plugins/cmdbuilder) then it will automatically create a command line program with flags for selecting configuration options like output directory and wiring spec name.

Using [SockShop](../../examples/sockshop) as an example, we would run:

```
go run main.go -o build -w docker
```

This will compile the artifacts to a directory `build`, after which we can then run them.

# Compilation

Compilation follows three stages.  

In the first stage, the wiring spec is executed, which defines services and modifies them.  This corresponds to everything up to the `BuildIR` invocation in the wiring spec.  If using the cmdbuilder, Stage 1 will end by printing out the wiring spec definitions, e.g.

```
[15:40:15.399] [plugins\cmdbuilder\cmdbuilder.go:186] INFO: Constructed SockShop WiringSpec docker:
SockShop = WiringSpec {
  catalogue_proc.stdoutmetriccollector = stdoutMetricCollector()
  wlgen.service = workflowNode(ptr=[wlgen.service.client] -> [wlgen.service.wlgen.proc -> wlgen.service.dst])
  cart_service.grpc.bind_addr = ApplicationNode()
  shipping_service.grpc.dial_addr = ApplicationNode()
  user_service.grpc_client = golangClient()
  payment_service.clientpool = ClientPool()
  ...
}
```

The second stage corresponds to the `BuildIR` call, which constructs nodes representing different entities of the Blueprint application.  If there were any erroneous definitions, `BuildIR` can fail.  If using the cmdbuilder, Stage 2 will end by printing out the application IR, e.g.

```
[15:04:25.607] [plugins\cmdbuilder\cmdbuilder.go:190] INFO: SockShop docker IR:
SockShop = BlueprintApplication() {
  cart_db.addr
  cart_db.bind_addr = AddressConfig()
  cart_db.ctr = MongoDBProcess(cart_db.bind_addr)
  cart_db.dial_addr = AddressConfig()
  cart_service.grpc.addr
  cart_service.grpc.bind_addr = AddressConfig()
  cart_service.grpc.dial_addr = AddressConfig()
  cart_service.handler.visibility
  cart_service_ctr = LinuxContainer(cart_db.dial_addr, cart_service.grpc.bind_addr, zipkin.dial_addr) {
    cart_service_proc = GolangProcessNode(cart_db.dial_addr, cart_service.grpc.bind_addr, zipkin.dial_addr) {
      cart_db.client = MongoClient(cart_db.dial_addr)
      cart_service = CartService(cart_db.client)
      cart_service.grpc_server = GRPCServer(cart_service.server.ot, cart_service.grpc.bind_addr)
      cart_service.server.ot = OTServerWrapper(cart_service, zipkin.client)
      cart_service_proc.logger = SLogger()
      cart_service_proc.stdoutmetriccollector = StdoutMetricCollector()
      zipkin.client = ZipkinClient(zipkin.dial_addr)
    }
  }
  catalogue_db.addr
  ...
}
```

The third stage corresponds to the `GenerateArtifacts` call, which outputs code artifacts from the application's IR.  This can be the most time consuming stage of compilation, depending on the complexity of the application.

# Running

Artifacts will be generated to the specified output directory; in the case of the example above, `build`.

The exact method of running the artifacts will depend on the wiring spec.  For example, if the wiring spec only generates processes, then you will need to manually run all of the processes.  If the wiring spec creates docker container images or a docker-compose file, then you will need to use docker commands to run them.

Each plugin documents how to run its generated artifacts; you should consult this documentation depending on the plugins used in your wiring spec.