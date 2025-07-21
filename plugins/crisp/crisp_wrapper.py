import sys
import json
import os

# Add CRISP repo to path
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "CRISP")))

from graph import Graph

if len(sys.argv) != 3:
    print("Usage: python crisp_wrapper.py <traces.json> <trace_id>")
    sys.exit(1)

trace_file = sys.argv[1]
trace_id = sys.argv[2]

with open(trace_file, 'r') as f:
    data = json.load(f)

traces = data["data"]
trace_data = next((t for t in traces if t.get("traceID") == trace_id), None)
if not trace_data:
    print(f"Trace {trace_id} not found.")
    sys.exit(1)

# Find the root span (no parent)
root_span = next(
    (span for span in trace_data["spans"] if not span.get("references")),
    trace_data["spans"][0]
)
process_id = root_span.get("processID", "unknown")
service_name = trace_data.get("processes", {}).get(process_id, {}).get("serviceName", process_id)
operation_name = root_span.get("operationName", "unknown")

single_trace_data = {"data": [trace_data]}
g = Graph(
    single_trace_data,
    service_name,
    operation_name,
    trace_file,
    trace_id
)

critical_path_nodes = g.findCriticalPath()
print("Critical Path Spans:")
total_duration = 0
for node in critical_path_nodes:
    print(f"  {node.opName} ({node.duration})")
    total_duration += node.duration
print("Critical Path Duration:", total_duration)

# Print the critical path as a sequence with service names
process_map = trace_data.get("processes", {})
def get_service_name(pid):
    return process_map.get(pid, {}).get("serviceName", pid)

path_str = " -> ".join(f"[{get_service_name(node.pid)}] {node.opName}" for node in critical_path_nodes)
print("Critical Path Sequence (with services):")
print(path_str) 