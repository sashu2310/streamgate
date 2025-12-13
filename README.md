# StreamGate

**StreamGate** is a high-performance, fail-open observability governance proxy. It acts as a middleware layer between your microservices and upstream observability vendors (like Datadog, New Relic, or CloudWatch).

## Purpose
StreamGate is designed to reduce observability costs and improve system stability by:
*   **Filtering:** Dropping low-value logs (e.g., `DEBUG` noise) at the source.
*   **Redaction:** Stripping PII (Personally Identifiable Information) before it leaves your network.
*   **Sampling:** Dynamically sampling high-volume streams.
*   **Fail-Open Design:** Ensuring that governance never becomes a bottleneck; if StreamGate is under pressure, it bypasses processing to prioritize throughput.

## Architecture
StreamGate uses a **Split-Plane Architecture**:
*   **Data Plane (Go):** Stateless, high-throughput proxy handling the hot path (Ingestion -> Buffer -> Process -> Output). Optimized for zero specific allocations and 100k+ events/sec.
*   **Control Plane (Python):** (Planned) Manages configuration rules and provides a dashboard for visibility.

## Current Status
*   [x] Project Skeleton & Interfaces
*   [x] Ingestion Layer (TCP/UDP)
*   [x] Buffering & Processing Engine
*   [x] Output Providers (Console)

## Quick Start

### Prerequisites
*   Go 1.23+

### Building
```bash
go build -o streamgate ./cmd/streamgate
```

### Running
Start the server:
```bash
./streamgate
```

### Testing (Manual)
Send a test log via TCP (in a separate terminal):
```bash
# Normal Log
echo "INFO: Hello StreamGate" | nc localhost 8081

# Sensitive Log (will be redacted)
echo "User 4111-1234 checkout" | nc localhost 8081

# Debug Log (will be dropped)
echo "DEBUG: this should vanish" | nc localhost 8081
```
