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

## Getting Started

### Prerequisites
*   Go 1.23+

### Building and Running
1.  **Clone the repository**:
    ```bash
    git clone https://github.com/sashu2310/streamgate.git
    cd streamgate
    ```

2.  **Run the application**:
    ```bash
    go run cmd/streamgate/main.go
    ```
    *Alternatively, build and run binary:*
    ```bash
    go build -o streamgate cmd/streamgate/main.go
    ./streamgate
    ```

3.  The server starts with default settings:
    *   **TCP Listener**: Port 8081
    *   **UDP Listener**: Port 8082
    *   **HTTP API**: Port 8080 (Placeholder)

## Usage Examples

Once StreamGate is running, you can send logs to it using `netcat` (nc) or any logging client.

### 1. Send Logs via TCP
```bash
echo "Info: User logged in" | nc localhost 8081
```

### 2. Send Logs via UDP
```bash
echo "Info: Metric received" | nc -u -w1 localhost 8082
```

### 3. Test Filtering (Drop DEBUG logs)
StreamGate is configured to drop logs containing "DEBUG".
```bash
# This will be dropped and NOT shown in the console output
echo "DEBUG: Detailed trace info" | nc localhost 8081
```

### 4. Test Redaction (PII)
StreamGate is configured to redact Credit Card numbers (pattern `4111-1234`).
```bash
# Input
echo "Payment processed for 4111-1234" | nc localhost 8081

# Output in Console
# Payment processed for xxxx-xxxx
```

## Configuration

Currently, configuration is statically defined in `cmd/streamgate/main.go` for the prototype:
*   **Buffer Size**: 65536 slots
*   **Processors**:
    *   Filter: Drops lines with "DEBUG"
    *   Redactor: Masks "4111-1234"

## Current Status
*   [x] Project Skeleton & Interfaces
*   [x] Ingestion Layer (TCP/UDP)
*   [x] High-Performance Ring Buffer
*   [x] Processing Engine (Filter, Redact)
*   [x] Console Output
*   [ ] Control Plane Integration
