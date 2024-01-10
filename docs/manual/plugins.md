# 📝 Wiring Spec Plugins

Blueprint comes with a number of plugins out-of-the-box.  

Full details of all plugins can be found by browsing the [/plugins](../../plugins) directory of the repository.

Here we provide a non-exhaustive list of some of the more important plugins along with brief examples of their use.  Full wiring spec examples can be found in the [example applications](../../examples).

For full details of plugin usage, please click the link to the plugin.


## Workflow Services

✏️[plugins/workflow](../../plugins/workflow)\
Creates named application-level instances of services defined in the application's workflow spec.
```
payment_service := workflow.Service(spec, "payment_service", "PaymentService")
```

## Workflow Backends

✏️[plugins/simple](../../plugins/simple)\
Creates simple, in-memory instances of backends that are only accessible by services that run within the same process.  Provides `backend.NoSQLDatabase`, `backend.RelationalDB`, `backend.Queue`, and `backend.Cache` instances.
```
cart_db := simple.NoSQLDB(spec, "cart_db")
catalogue_db := simple.RelationalDB(spec, "catalogue_db")
shipqueue := simple.Queue(spec, "shipping_queue")
user_cache := simple.Cache(spec, "user_cache")
```

### ✏️[plugins/memcached](../../plugins/memcached)
Creates container-level instances of `backend.Cache` using memcached.  
```
user_cache := memcached.Container(spec, "user_cache")
```
See also ✏️[plugins/redis](../../plugins/redis)

### ✏️[plugins/mongodb](../../plugins/mongodb)
Creates container-level instances of `backend.NoSQLDatabase` using MongoDB
```
user_cache := memcached.Container(spec, "user_cache")
```

✏️[plugins/mysql](../../plugins/mysql)\
Creates container-level instances of `backend.RelationalDB` using MySQL
```
user_cache := memcached.Container(spec, "user_cache")
```

✏️[plugins/rabbitmq](../../plugins/rabbitmq)\
Creates container-level instances of `backend.Queue` using RabbitMQ
```
user_cache := memcached.Container(spec, "user_cache")
```

✏️[plugins/jaeger](../../plugins/jaeger)\
Creates a Jaeger container instance, for use as a collector in conjunction with the opentelemetry plugin.
```
trace_collector := jaeger.Collector(spec, "trace_collector")
```
See also ✏️[plugins/zipkin](../../plugins/zipkin)


## Service Modifiers

✏️[plugins/retries](../../plugins/retries)\
Adds retries and timeouts to clients of this service.
```
retries.AddRetriesWithTimeouts(spec, "payment_service", 3, "1s")
```


✏️[plugins/opentelemetry](../../plugins/opentelemetry)\
Instruments all calls with OpenTelemetry to create trace spans.
References a trace collector that was defined using a plugin such as the jaeger or zipkin plugin.
```
opentelemetry.Instrument(spec, "payment_service", "trace_collector")
```


✏️[plugins/zipkin](../../plugins/jaeger)\

✏️[plugins/retries](../../plugins/retries)\
Adds retries and timeouts to clients of this service.
```
retries.AddRetriesWithTimeouts(spec, "payment_service", 3, "1s")
```

✏️[plugins/grpc](../../plugins/grpc)\
Deploys an application-level service instance over RPC using gRPC, enabling its use by other services running in other processes.
```
grpc.Deploy(spec, "payment_service")
```

✏️[plugins/goproc](../../plugins/goproc)\
Creates a processes that contain application-level instances
```
goproc.Deploy(spec, "payment_service")
```

✏️[plugins/linuxcontainer](../../plugins/linuxcontainer)\
Creates container images that contain processes
```
linuxcontainer.Deploy(spec, "payment_service")
```
