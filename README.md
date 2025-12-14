# StreamGate

**StreamGate** is a high-performance, fail-open observability governance proxy. It acts as a middleware layer between your microservices and upstream observability vendors (like Datadog, New Relic, or CloudWatch).

## Purpose
StreamGate is designed to reduce observability costs and improve system stability by:
*   **Filtering:** Dropping low-value logs (e.g., `DEBUG` noise) at the source.
*   **Redaction:** Stripping PII (Personally Identifiable Information) before it leaves your network.
*   **Sampling:** Dynamically sampling high-volume streams.
*   **Fail-Open Design:** Ensuring that governance never becomes a bottleneck; if StreamGate is under pressure, it bypasses processing to prioritize throughput.

## Features
- **Ingestion**: TCP & UDP (High Performance)
- **Processing**:
    - JSON-based Rules (Redis)
    - Hot-swappable Filter & Redaction processors
    - Dynamic Batch Size configuration
- **Output**:
    - Console (Stdout)
    - HTTP (Webhooks, External APIs)
    - Fan-out (Multiple outputs simultaneously)
- **Architecture**:
    - Separate Control Plane (Python) & Data Plane (Go)
    - Redis-based state management
    - Dockerized & Ready for Orchestration

## Architecture
StreamGate uses a **Split-Plane Architecture**:
*   **Data Plane (Go):** Stateless, high-throughput proxy handling the hot path (Ingestion -> Buffer -> Process -> Output). Optimized for zero specific allocations and 100k+ events/sec.
*   **Control Plane (Python):** REST API for managing configuration rules (Filters, Redactions).
*   **Sync Layer (Redis):** Pub/Sub mechanism ensuring real-time configuration hot-reloading without service restarts.

## Quick Start (Docker)

The easiest way to run StreamGate is using Docker Compose. This starts Redis, the Control Plane, and the Data Plane automatically.

1.  **Clone the repository**:
    ```bash
    git clone https://github.com/sashu2310/streamgate.git
    cd streamgate
    ```

2.  **Start Services**:
    ```bash
    docker-compose up --build
    ```

3.  **Verify**:
    *   **Control Plane (API):** `http://localhost:8000/docs`
    *   **Data Plane (Ingest):** Sending logs to `localhost:8081`

---

## Manual Setup (Development)

### Prerequisites
*   Go 1.23+
*   Python 3.8+
*   Redis (running on localhost:6379)

### 1. Start Redis
```bash
docker run -p 6379:6379 -d redis
# OR
redis-server
```

### 2. Start Control Plane (Python)
```bash
cd control_plane
pip install -r requirements.txt
uvicorn main:app --reload --port 8000
```
*API is now running at `http://localhost:8000`*

### 3. Start Data Plane (Go)
Open a new terminal:
```bash
go run cmd/streamgate/main.go
```
*StreamGate is now listening on TCP :8081 and UDP :8082*

## Usage & Verification

### Send Logs (Data Plane)
StreamGate starts with an **Empty Pipeline** (Pass-through).
```bash
echo "DEBUG: test log" | nc localhost 8081
# Output: "DEBUG: test log"
```

### Add a Rule (Control Plane)
Add a rule to filter out "DEBUG" logs dynamically.
```bash
# 1. Create Rule
curl -X POST "http://localhost:8000/rules" \
     -H "Content-Type: application/json" \
     -d '{"id": "drop_debug", "type": "filter", "params": {"value": "DEBUG"}}'

# 2. Publish Config (Hot Reload)
curl -X POST "http://localhost:8000/publish" -d ''
```

### Verify Logic Change
Send the same log again. It should now be dropped only because of the rule update.
```bash
echo "DEBUG: test log" | nc localhost 8081
# Output: (Silence)
```

## Configuration
*   **TCP Port**: 8081
*   **UDP Port**: 8082
*   **Redis**: localhost:6379
*   **Control Plane API**: localhost:8000

## Current Status
*   [x] Project Skeleton & Interfaces
*   [x] Ingestion Layer (TCP/UDP)
*   [x] High-Performance Ring Buffer
*   [x] Processing Engine (Filter, Redact)
*   [x] Control Plane API (Python/FastAPI)
*   [x] End-to-End Hot Reloading
*   [x] HTTP Output Provider
*   [x] Dynamic Batch Size Configuration
