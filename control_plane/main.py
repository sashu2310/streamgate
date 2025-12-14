from fastapi import FastAPI, HTTPException
import redis
import json
import os
from typing import List
from models import Manifest, PipelineConfig, ProcessorRule

app = FastAPI(title="StreamGate Control Plane")

# Configuration
REDIS_HOST = os.getenv("REDIS_HOST", "localhost")
REDIS_PORT = int(os.getenv("REDIS_PORT", 6379))
REDIS_CHANNEL = "streamgate_updates"
REDIS_KEY = "streamgate_config"

# Redis Client
r = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)

# In-memory store for the prototype (In real app, use DB)
# Default empty state
current_rules: List[ProcessorRule] = []

@app.get("/")
def health():
    return {"status": "ok", "service": "streamgate-control-plane"}

@app.get("/rules", response_model=List[ProcessorRule])
def get_rules():
    return current_rules

@app.post("/rules")
def add_rule(rule: ProcessorRule):
    # Basic validation: Check if ID exists
    for r in current_rules:
        if r.id == rule.id:
            raise HTTPException(status_code=400, detail=f"Rule ID {rule.id} already exists")
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

@app.post("/publish")
def publish_config():
    """
    Compiles the current rules into a Manifest and pushes to Redis.
    """
    # 1. Build Manifest
    pipeline = PipelineConfig(name="default_pipeline", processors=current_rules)
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
        "subscribers_notified": subscribers
    }
