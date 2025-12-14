from typing import List, Optional, Dict, Literal
from pydantic import BaseModel, Field
import time

class ProcessorRule(BaseModel):
    id: str
    type: Literal["filter", "redact"]
    params: Dict[str, str] = Field(..., description="Configuration parameters for the processor")

    # Example params:
    # Filter: {"key": "level", "value": "DEBUG"}
    # Redact: {"pattern": "4111-xxxx", "replacement": "xxxx-xxxx"}

class PipelineConfig(BaseModel):
    name: str
    processors: List[ProcessorRule]

class Manifest(BaseModel):
    version: str = "1.0"
    timestamp: int = Field(default_factory=lambda: int(time.time()))
    pipelines: List[PipelineConfig]
