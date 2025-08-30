from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from .rag import load_data_and_index, search_similar

app = FastAPI(title="Local RAG Service", description="A local RAG system .")

class QueryRequest(BaseModel):
    message: str
    k: int = 5

@app.on_event("startup")
def startup_event():
    load_data_and_index()

@app.post("/retrieve", response_model=list)
def retrieve_similar(request: QueryRequest):
    if not request.message.strip():
        raise HTTPException(status_code=400, detail="Empty message")
    try:
        results = search_similar(request.message, k=request.k)
        return results
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Search failed: {str(e)}")

@app.get("/healthz")
def health_check():
    return {"status": "ok", "message": "RAG service is healthy"}