# StreamGate: High-Performance Observability Governance Proxy
## System Design Document (v1.0)

### 1. Executive Summary
StreamGate is a high-throughput, low-latency middleware proxy designed to intercept observability data (logs, metrics) between application services and upstream vendors (Datadog, CloudWatch).
**Goal:** Reduce observability costs by 40-60% through intelligent filtering, deduplication, and dynamic sampling before data leaves the network.

### 2. High-Level Architecture
The system follows a **Split-Plane Architecture**:

#### A. Data Plane (The "Muscle") - Language: Go (Golang)
* **Role:** Stateless, high-performance proxy.
* **Responsibilities:**
    * Ingest logs via TCP/UDP (Syslog/StatsD compatible) and HTTP.
    * Apply transformation rules (Redact, Drop, Sample) in real-time.
    * Forward processed data to upstream APIs.
    * **Critical Requirement:** "Fail-Open" design. Uses a Circuit Breaker pattern. If the processing pipeline stalls or errors spike, traffic bypasses the filter to prevent application outages.
* **Tech Stack:** Go 1.23+, `fasthttp` (for low allocation), Channels for async processing.

#### B. Control Plane (The "Brain") - Language: Python (FastAPI)
* **Role:** Configuration management and analytics.
* **Responsibilities:**
    * REST API for defining rules (e.g., "Drop all DEBUG logs from Service A").
    * Dashboard for visualizing "Ingestion vs. Output" volume.
    *   Pushes configuration updates to the Data Plane via Redis Pub/Sub (Signaling only).
* **Tech Stack:** Python 3.12, FastAPI, Pydantic, Redis.

#### C. System Diagram

```mermaid
graph TD
    user((User/Dev)) -->|Configures Rules| CP_UI[Control Plane Dashboard]
    
    subgraph "Control Plane (Python)"
        CP_API[FastAPI Server]
        CP_UI --> CP_API
        CP_API -->|Publish Config| Redis[(Redis)]
    end

    subgraph "Data Plane (Go)"
        direction TB
        
        subgraph "StreamGate Instance"
            Ingest[Ingest Listener<br>TCP/UDP/HTTP]
            
            subgraph "Processing Core"
                LocalCache[Local Config Cache]
                FilterEngine[Filter & Transform Engine]
                Bypass[Fail-Open Bypass]
            end
            
            Buffer[Ring Buffer]
            Sender[Batch Sender]
            
            CircuitBreaker{Circuit<br>Breaker}
        end
        
        Ingest --> CircuitBreaker
        
        CircuitBreaker -->|Closed (Normal)| FilterEngine
        CircuitBreaker -->|Open (High Load)| Bypass
        
        Redis -.->|Async Update| LocalCache
        LocalCache --> FilterEngine
        
        FilterEngine -->|Processed| Buffer
        Bypass -->|Raw| Buffer
        Buffer --> Sender
    end

    subgraph "Upstream Destinations"
        DD[Datadog]
        CW[CloudWatch]
    end

    Sender -->|HTTPS Batch| DD
    Sender -->|HTTPS Batch| CW

    Sources[Microservices] -->|Logs/Metrics| Ingest
```

---

### 3. Core Component Design

#### 3.1 The Processing Pipeline (Go)
The Data Plane implements a "Chain of Responsibility" pattern. Every incoming log entry passes through a configured list of processors.

**Interface Definition:**
```go
type Processor interface {
    // Process takes a log entry buffer and returns (modified_entry, should_drop, error)
    // Implementations must assume zero-allocation hot paths.
    Process(context *ProcessingContext, entry []byte) ([]byte, bool, error)
}
