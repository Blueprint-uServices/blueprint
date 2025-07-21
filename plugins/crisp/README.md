# CRISP Plugin for Blueprint

This plugin allows you to run critical path analysis on Jaeger traces using the CRISP tool, exposed as a Python FastAPI service in a Docker container.

## Setup

1. Clone the CRISP repo into this directory:
   ```
   git clone https://github.com/uber-research/CRISP.git
   ```
2. Ensure you have Python 3 installed.
3. Install any required Python packages for CRISP (most are standard libraries).

## Building the Docker image

Before running the Blueprint application with critical path analysis, you must build the CRISP Docker image:

```
docker build -t crisp:latest .
```

**Note:**
- If you make changes to the CRISP plugin or its dependencies, rebuild the Docker image using the above command.
- See the main README in the example application for integration and usage steps.

## Environment Variables

- `JAEGER_URL`: The base URL for the Jaeger query service. The CRISP service will use this to fetch traces. Default: `http://jaeger_ctr:16686`

You can set this variable in your `docker-compose.yml` or Dockerfile:

**docker-compose.yml**
```yaml
services:
  crisp:
    image: crisp:latest
    environment:
      - JAEGER_URL=http://jaeger:16686
```

**Dockerfile**
```dockerfile
ENV JAEGER_URL=http://jaeger:16686
```

If you use a `.env` file, Docker Compose will automatically load it if present.

## Usage

### As a Service:
- The CRISP service is exposed via FastAPI in a Docker container.
- The container exposes an endpoint:
  - `GET /analyze/{trace_id}`: Runs critical path analysis for the given Jaeger trace ID.

### Example HTTP Request:
```sh
curl http://localhost:8000/analyze/<trace_id>
```

### Output
- The endpoint returns a JSON object with the critical path and total duration for the given trace.

---

**Note:**
- The Go plugin wires up the CRISP container in your Blueprint deployment. You can interact with it from any language via HTTP. 