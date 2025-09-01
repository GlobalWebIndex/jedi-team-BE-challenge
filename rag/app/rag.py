# app/rag.py
import os
import numpy as np
from sentence_transformers import SentenceTransformer
import faiss
import markdown

MODEL_NAME = "all-MiniLM-L6-v2"
EMBEDDING_DIM = 384  # Output dim for all-MiniLM-L6-v2
INDEX_FILE = "vector.index"
DATA_FILE = "app/data/data.md"
K_DEFAULT = 5

model = SentenceTransformer(MODEL_NAME)

index = None
sentences = []

def markdown_to_sentences(text):
    """
    Extract sentences from a Markdown table
    """
    lines = text.strip().splitlines()
    sentences = []

    for line in lines[2:]:
        line = line.strip()
        if line.startswith('|') and line.endswith('|'):
            content = line[1:-1].strip()
            if content and not content.startswith(':') and not content.lower() == 'text': # redudant but make sure we arent checking first lines
                sentences.append(content)

    return sentences

def load_data_and_index():
    """
        Load data.md, split into sentences, create embeddings, and build FAISS index.
    """
    global index, sentences

    if os.path.exists(INDEX_FILE):
        print("Loading existing FAISS index...")
        index = faiss.read_index(INDEX_FILE)
        with open("sentences.npy", "rb") as f:
            sentences = np.load(f, allow_pickle=True).tolist()
        return

    if not os.path.exists(DATA_FILE):
        print(f"Data file {DATA_FILE} not found. Failing...")
        raise FileNotFoundError(f"The file '{DATA_FILE}' does not exist.")
    else:
        with open(DATA_FILE, "r", encoding="utf-8") as f:
            content = f.read()
        raw_sentences = markdown_to_sentences(content)
        sentences = [s.strip() for s in raw_sentences if len(s.strip()) > 10] # redundant but skip irrelevant sentences

    print(f"Found {len(sentences)} sentences. Creating embeddings...")
    embeddings = model.encode(sentences, convert_to_numpy=True)

    # Normalize embeddings
    faiss.normalize_L2(embeddings)

    # Build FAISS index
    index = faiss.IndexFlatIP(EMBEDDING_DIM) 
    index.add(embeddings)

    faiss.write_index(index, INDEX_FILE)
    with open("sentences.npy", "wb") as f:
        np.save(f, np.array(sentences))
    print("Index built and saved.")

def search_similar(query: str, k: int = K_DEFAULT) -> list:
    """Search for top-k most similar sentences."""
    query_vec = model.encode([query])
    faiss.normalize_L2(query_vec)
    scores, indices = index.search(query_vec, k)
    results = [
        {"sentence": sentences[i], "score": float(score)}
        for score, i in zip(scores[0], indices[0]) if i != -1
    ]
    return results