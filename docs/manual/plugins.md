# 📝 Wiring Spec Plugins

Here we provide a non-exhaustive list of some of the more important Blueprint plugins along with brief examples of their use.  Full wiring spec examples can be found in the [example applications](../../examples).

Full details of all plugins can be found by browsing the [/plugins](../../plugins) directory of the repository.

## Workflow Services

### ✏️[workflow](../../plugins/workflow)
Creates named application-level instances of services defined in the application's workflow spec.
```
payment_service := workflow.Service[payment.PaymentService](spec, "payment_service")
```

## Workflow Backends

### ✏️[simple](../../plugins/simple)
Creates basic in-memory instances of backends that are only accessible within the same process.  Provides `backend.NoSQLDatabase`, `backend.RelationalDB`, `backend.Queue`, and `backend.Cache` instances.
```
cart_db := simple.NoSQLDB(spec, "cart_db")
catalogue_db := simple.RelationalDB(spec, "catalogue_db")
shipqueue := simple.Queue(spec, "shipping_queue")
user_cache := simple.Cache(spec, "user_cache")
```

### ✏️[memcached](../../plugins/memcached)
Creates container-level instances of `backend.Cache` using memcached.  
```
user_cache := memcached.Container(spec, "user_cache")
```
See also ✏️[redis](../../plugins/redis) to create `backend.Cache` instances using Redis.

### ✏️[mongodb](../../plugins/mongodb)
Creates container-level instances of `backend.NoSQLDatabase` using MongoDB
```
user_cache := memcached.Container(spec, "user_cache")
```

### ✏️[mysql](../../plugins/mysql)
Creates container-level instances of `backend.RelationalDB` using MySQL
```
user_cache := memcached.Container(spec, "user_cache")
```

### ✏️[rabbitmq](../../plugins/rabbitmq)
Creates container-level instances of `backend.Queue` using RabbitMQ
```
user_cache := memcached.Container(spec, "user_cache")
```

### ✏️[jaeger](../../plugins/jaeger)
Creates a Jaeger container instance, for use as a collector in conjunction with the opentelemetry plugin.
```
trace_collector := jaeger.Collector(spec, "trace_collector")
```
See also ✏️[zipkin](../../plugins/zipkin) to use Zipkin as the trace collector.


## Service Modifiers

### ✏️[retries](../../plugins/retries)
Modifies an application-level service to add retries and timeouts to clients that call this service.
```
retries.AddRetriesWithTimeouts(spec, "payment_service", 3, "1s")
```


### ✏️[opentelemetry](../../plugins/opentelemetry)
Modifies an application-level service to create OpenTelemetry trace spans on both the client and server side.
References a trace collector that was defined using a plugin such as the jaeger or zipkin plugin.
```
opentelemetry.Instrument(spec, "payment_service", "trace_collector")
```
See also ✏️[plugins/xtrace](../../plugins/xtrace) to trace applications using X-Trace.


### ✏️[grpc](../../plugins/grpc)
Deploys an application-level service instance over RPC using gRPC, enabling its use by other services running in other processes.
```
grpc.Deploy(spec, "payment_service")
```
See also ✏️[plugins/thrift](../../plugins/thrift) to use Thrift as the RPC framework.

## Namespaces

### ✏️[goproc](../../plugins/goproc)
Combines application-level instances into a process-level instance
```
goproc.Deploy(spec, "payment_service")
```

### ✏️[linuxcontainer](../../plugins/linuxcontainer)
Combines process-level instances into a container-level instance
```
linuxcontainer.Deploy(spec, "payment_service")
```
