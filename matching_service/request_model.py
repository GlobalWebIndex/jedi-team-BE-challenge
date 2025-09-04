from pydantic import BaseModel, Field, validator
from typing import Optional
import re

class MatchRequest(BaseModel):
    query: str = Field(..., description="User query to match against")
    threshold: Optional[float] = Field(0, description="Minimum score threshold")

    @validator("query")
    def sanitize_query(cls, v):
        # Remove suspicious characters that could be used in path injection
        if not v.strip():
            raise ValueError("Query must not be empty")

        # Basic sanitization: remove dangerous characters
        sanitized = re.sub(r'[<>;"\'|&$`]', '', v)

        # You can also add limits here if needed
        if len(sanitized) > 1000:
            raise ValueError("Query is too long")

        return sanitized