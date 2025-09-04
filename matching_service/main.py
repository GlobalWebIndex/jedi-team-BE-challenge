from request_model import MatchRequest
from response_model import MatchResponse
from flask import Flask, request, jsonify
from sentence_transformers import SentenceTransformer
import faiss
from utils.matching_algorithms import most_common_word_match, best_fuzzy_match



app = Flask(__name__)


# Load replies
with open("data.md", "r") as f:
    sentences = [line.strip() for line in f if line.strip()]
    (len(sentences))
    # sentences_without_percentages = [
    #     re.sub(r'\b\d+(\.\d+)?%\b|[|]', '', line.strip())
    #     for line in sentences if line.strip()
    # ]
    # sentences_without_percentages = [
    # re.sub(r'\b\d+(\.\d+)?%\b', '', line.replace('|', '')).strip()
    # for line in sentences
    # ]

    sentences_without_fullstops = [line.replace(".","") for line in sentences]




@app.route("/match-cosine", methods=["POST"]) 
def match():
    data = request.get_json()
    req_model = MatchRequest.model_validate(data) 
    query = req_model.query
    threshold = req_model.threshold if req_model.threshold != 0 else 0.70  # cosine similarity in L2 space

    # Load model
    model = SentenceTransformer("all-MiniLM-L6-v2")

    # Encode replies
    embeddings = model.encode(sentences, convert_to_numpy=True)

    # Store in FAISS index
    index = faiss.IndexFlatL2(embeddings.shape[1])
    index.add(embeddings)

    query_embedding = model.encode([query], convert_to_numpy=True)
    D, I = index.search(query_embedding, k=1)
    score = 1 - (D[0][0] / 2)  # Convert L2 to approximate cosine similarity

    if score >= threshold:
        response = MatchResponse(
            matched=True,
            reply=sentences[I[0][0]],
            score=float(score)
        )
    else:
        response = MatchResponse(
            matched=False,
            reply=None,
            score=float(score)
        )
    return jsonify(response.model_dump())


@app.route("/match-words", methods=["POST"])
def match_words():
    data = request.get_json()
    req_model = MatchRequest.model_validate(data) 
    query = req_model.query

    match, score = most_common_word_match(query.replace('?',''), sentences_without_fullstops)


    if score != 0:
        response = MatchResponse(
            matched=True,
            reply=match,
            score=float(score)
        )
    else:
        response = MatchResponse(
            matched=False,
            reply=None,
            score=float(score)
        )
    return jsonify(response.model_dump())


@app.route("/match-fuzzy", methods=["POST"])
def match_fuzzer():
    data = request.get_json()
    req_model = MatchRequest.model_validate(data) 
    query = req_model.query
    threshold = req_model.threshold if req_model.threshold != 0 else 75


    match, score = best_fuzzy_match(query.replace('?',''), sentences, threshold=threshold) 
    # 85 is ok
    
    if score != 0:
        response = MatchResponse(
            matched=True,
            reply=match,
            score=float(score)
        )
    else:
        response = MatchResponse(
            matched=False,
            reply=None,
            score=float(score)
        )
    return jsonify(response.model_dump())


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5001)
