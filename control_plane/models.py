from typing import List, Optional, Dict, Literal
from pydantic import BaseModel, Field
import time

class ProcessorRule(BaseModel):
    id: str
    type: Literal["filter", "redact", "attribute_filter"]
    params: Dict[str, str] = Field(..., description="Configuration parameters for the processor")

    # Example params:
    # Filter: {"value": "DEBUG"}
    # Redact: {"pattern": "\\d{3}-\\d{2}-\\d{4}", "replacement": "XXX-XX-XXXX"}
    # AttributeFilter (well-known OTel attribute, auto-search):
    #   {"attribute": "service.name", "operator": "equals", "value": "test-service"}
    # AttributeFilter (explicit path):
    #   {"path": "resource/attributes/custom.field", "operator": "contains", "value": "debug"}

class OutputTarget(BaseModel):
    type: Literal["console", "http"]
    url: Optional[str] = None
    headers: Optional[Dict[str, str]] = None

class PipelineConfig(BaseModel):
    name: str
    processors: List[ProcessorRule]
    outputs: List[OutputTarget] = Field(default_factory=list)
    batch_size: int = Field(default=100, ge=1, le=10000)

class Manifest(BaseModel):
    version: str = "1.0"
    timestamp: int = Field(default_factory=lambda: int(time.time()))
    pipelines: List[PipelineConfig]
