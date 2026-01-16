from fastapi import FastAPI, HTTPException
import redis
import os
from typing import List
from models import Manifest, PipelineConfig, ProcessorRule, OutputTarget

app = FastAPI(title="StreamGate Control Plane")

# Configuration
REDIS_HOST = os.getenv("REDIS_HOST", "localhost")
REDIS_PORT = int(os.getenv("REDIS_PORT", 6379))
REDIS_CHANNEL = "streamgate_updates"
REDIS_KEY = "streamgate_config"

# Redis Client
r = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)

# In-memory store
current_rules: List[ProcessorRule] = []
current_outputs: List[OutputTarget] = []
current_batch_size: int = 100


@app.get("/")
def health():
    return {"status": "ok", "service": "streamgate-control-plane"}


# --- Rules ---
@app.get("/rules", response_model=List[ProcessorRule])
def get_rules():
    return current_rules


@app.post("/rules")
def add_rule(rule: ProcessorRule):
    for r in current_rules:
        if r.id == rule.id:
            raise HTTPException(
                status_code=400, detail=f"Rule ID {rule.id} already exists"
            )
    current_rules.append(rule)
    return {"status": "added", "rule": rule}


@app.delete("/rules/{rule_id}")
def delete_rule(rule_id: str):
    global current_rules
    initial_len = len(current_rules)
    current_rules = [r for r in current_rules if r.id != rule_id]
    if len(current_rules) == initial_len:
        raise HTTPException(status_code=404, detail="Rule not found")
    return {"status": "deleted", "id": rule_id}


# --- Outputs ---
@app.get("/outputs", response_model=List[OutputTarget])
def get_outputs():
    return current_outputs


@app.post("/outputs")
def add_output(output: OutputTarget):
    # Basic check to avoid duplicate URLs for http
    if output.type == "http" and output.url:
        for o in current_outputs:
            if o.type == "http" and o.url == output.url:
                raise HTTPException(status_code=400, detail="Output URL already exists")
    current_outputs.append(output)
    return {"status": "added", "output": output}


@app.delete("/outputs")
def clear_outputs():
    """Clear all outputs (reset to empty)"""
    global current_outputs
    current_outputs = []
    return {"status": "cleared"}


# --- Settings ---
@app.get("/config/batch_size")
def get_batch_size():
    return {"batch_size": current_batch_size}


@app.post("/config/batch_size")
def set_batch_size(size: int):
    global current_batch_size
    if size < 1 or size > 10000:
        raise HTTPException(
            status_code=400, detail="Batch size must be between 1 and 10000"
        )
    current_batch_size = size
    return {"status": "updated", "batch_size": size}


# --- Publish ---
@app.post("/publish")
def publish_config():
    """
    Compiles Rules + Outputs into a Manifest and pushes to Redis.
    """
    # 1. Build Manifest
    pipeline = PipelineConfig(
        name="default_pipeline",
        processors=current_rules,
        outputs=current_outputs,
        batch_size=current_batch_size,
    )
    manifest = Manifest(pipelines=[pipeline])

    # 2. Serialize
    data = manifest.json()

    # 3. Save to Redis (Persist)
    r.set(REDIS_KEY, data)

    # 4. Notify Listeners (Hot Reload)
    subscribers = r.publish(REDIS_CHANNEL, "RELOAD")

    return {
        "status": "published",
        "manifest": manifest,
        "subscribers_notified": subscribers,
    }
