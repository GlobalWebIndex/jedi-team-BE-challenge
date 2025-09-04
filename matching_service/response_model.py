from pydantic import BaseModel, Field
from typing import Optional

class MatchResponse(BaseModel):
    matched: bool = Field(..., description="Whether a match was found")
    reply: Optional[str] = Field(None, description="The matched response")
    score: Optional[float] = Field(None, description="Confidence score (e.g., similarity)")
    