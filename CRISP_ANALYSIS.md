# CRISP Critical Path Analysis Guide

This guide explains how to use the CRISP plugin for critical path analysis in the Blueprint microservices framework, including integration details, experiment setup, and trace analysis.

---

## What is CRISP?

**CRISP** is a Python-based tool for automated critical path analysis of distributed traces. It exposes a FastAPI service that can analyze Jaeger traces and return the critical path and its latency, helping you identify bottlenecks in microservice applications.

**Critical Path Analysis** is the process of finding the slowest (longest-latency) sequence of operations in a distributed trace, which determines the end-to-end response time for a request.

---

## Where and How is CRISP Integrated?

- The CRISP plugin is located in [`plugins/crisp/`](plugins/crisp/).
- It is integrated into the DeathStarBench Hotel Reservation example (`examples/dsb_hotel`) via the wiring spec files:
  - [`examples/dsb_hotel/wiring/specs/original.go`](examples/dsb_hotel/wiring/specs/original.go)
  - [`examples/dsb_hotel/wiring/specs/chain.go`](examples/dsb_hotel/wiring/specs/chain.go)
  - [`examples/dsb_hotel/wiring/specs/fanin.go`](examples/dsb_hotel/wiring/specs/fanin.go)
  - [`examples/dsb_hotel/wiring/specs/fanout.go`](examples/dsb_hotel/wiring/specs/fanout.go)
- In each wiring spec, a CRISP analysis container is added:
  ```go
  analysisContainer := crisp.Container(spec, "trace_analysis")
  // (add environment variable for Jaeger URL, see below)
  ```
- The CRISP container is included in the auto-generated `docker-compose.yml` for each experiment.

---

## How to Build and Run Experiments with CRISP

### 1. Build the CRISP Docker Image

```sh
cd plugins/crisp
docker build -t crisp:latest .
```

### 2. Compile the Blueprint Application with Desired Topology

```sh
cd examples/dsb_hotel/wiring
go run main.go -w <spec> -o build_<spec>
```
Where `<spec>` is one of: `original`, `chain`, `fanin`, `fanout`.

### 3. (Optional) Set Jaeger URL for CRISP

By default, CRISP will look for Jaeger at `http://jaeger_ctr:16686` (the default service name and port in the generated Docker Compose). If you want to override this, set the environment variable in the wiring spec (recommended for automation):

```go
import "github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
...
crispCtr := crisp.Container(spec, "trace_analysis")
linuxcontainer.SetEnv(crispCtr, "JAEGER_URL", "http://jaeger_ctr:16686")
```

Or, edit the generated `docker-compose.yml` to add:
```yaml
  environment:
    - JAEGER_URL=http://jaeger_ctr:16686
```

### 4. Run the Application

```sh
cd build_<spec>/docker
docker compose up
```

### 5. Generate Workload and Collect Traces

- Use the workload generator included in the build directory (e.g., `wlgen_proc`) to send requests to the frontend service.
- Traces will be collected by Jaeger.

### 6. Analyze Traces with CRISP

- Access Jaeger UI at [http://localhost:16686](http://localhost:16686) to find trace IDs.
- Use CRISP's API to analyze a trace:
  ```sh
  curl http://localhost:8000/analyze/<trace_id>
  ```
- The response will include the critical path and its total duration.

---

## How the Wiring Specs and Docker Compose are Set Up

- The wiring spec files define which services and containers are included in each experiment.
- When you build a spec, Blueprint generates a `docker-compose.yml` that includes all services, including Jaeger and CRISP.
- Go-based services are instrumented with tracing plugins and send traces to Jaeger via `JAEGER_DIAL_ADDR`.
- CRISP reads traces from Jaeger via the query API (`JAEGER_URL`).

---

## Example: Adding JAEGER_URL in the Wiring Spec

```go
import "github.com/blueprint-uservices/blueprint/plugins/crisp"
import "github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
...
crispCtr := crisp.Container(spec, "trace_analysis")
linuxcontainer.SetEnv(crispCtr, "JAEGER_URL", "http://jaeger_ctr:16686")
```

---

## Troubleshooting

- **CRISP can't connect to Jaeger:**
  - Make sure the Jaeger service name and port match what CRISP expects (`jaeger_ctr:16686` by default).
  - Set the `JAEGER_URL` environment variable if needed.
- **CRISP container not present in Docker Compose:**
  - Ensure the wiring spec includes the CRISP container.
- **Manual changes to docker-compose.yml are overwritten:**
  - Automate environment variable injection in the wiring spec for persistence.

---

## Further Reading

- [CRISP Plugin README](plugins/crisp/README.md)
- [Blueprint User Manual](docs/manual)
- [DeathStarBench Hotel Reservation Example](examples/dsb_hotel/README.md)

---

For questions or contributions, see the main [Blueprint README](README.md) or contact the maintainers via Slack or mailing list. 