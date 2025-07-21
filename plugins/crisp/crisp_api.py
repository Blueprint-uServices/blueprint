from fastapi import FastAPI, HTTPException
import requests
import sys
import os

# Add CRISP repo to path
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "CRISP")))
from graph import Graph

app = FastAPI()

@app.get("/analyze/{trace_id}")
def analyze_trace(trace_id: str):
    # 1. Fetch trace from Jaeger
    # Get Jaeger base URL from environment variable, default to previous value
    jaeger_base_url = os.environ.get("JAEGER_URL", "http://jaeger_ctr:16686")
    jaeger_url = f"{jaeger_base_url}/api/traces/{trace_id}"
    resp = requests.get(jaeger_url)
    if resp.status_code != 200:
        raise HTTPException(status_code=404, detail="Trace not found in Jaeger")
    data = resp.json()
    if not data.get("data"):
        raise HTTPException(status_code=404, detail="No trace data found")

    trace_data = data["data"][0]

    # 2. Find root span and service/operation
    root_span = next(
        (span for span in trace_data["spans"] if not span.get("references")),
        trace_data["spans"][0]
    )
    process_id = root_span.get("processID", "unknown")
    service_name = trace_data.get("processes", {}).get(process_id, {}).get("serviceName", process_id)
    operation_name = root_span.get("operationName", "unknown")

    # 3. Prepare data for CRISP
    single_trace_data = {"data": [trace_data]}
    g = Graph(
        single_trace_data,
        service_name,
        operation_name,
        "from_jaeger_api",
        trace_id
    )

    critical_path_nodes = g.findCriticalPath()
    total_duration = sum(node.duration for node in critical_path_nodes)
    process_map = trace_data.get("processes", {})
    def get_service_name(pid):
        return process_map.get(pid, {}).get("serviceName", pid)
    path = [
        {
            "service": get_service_name(node.pid),
            "operation": node.opName,
            "duration": node.duration
        }
        for node in critical_path_nodes
    ]

    return {
        "trace_id": trace_id,
        "critical_path": path,
        "total_duration": total_duration
    } 