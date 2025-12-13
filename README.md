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
*   [ ] Ingestion Layer (TCP/UDP)
*   [ ] Buffering & Processing Engine
*   [ ] Output Providers
