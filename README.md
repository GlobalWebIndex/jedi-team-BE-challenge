# 🧠 Multi-Service Question Matching System

This project consists of three interconnected services:

1. **Question API (Go - Gin)** – Receives user questions and routes them through the Matching API, then stores the conversation in MongoDB.
2. **Matching API (Python - Flask)** – Matches user questions to the best possible replies using semantic similarity.
3. **MongoDB** – Stores all conversations, including questions and matched replies.


---

## 🚀 How to Deploy

All services are containerized using Docker. To bring everything up:

### ✅ Prerequisites

- Docker & Docker Compose installed
- Ports `5001`, `7004`, and `27017` available

---

### 🐳 Step 1: Run All Services

```bash
docker compose -f docker-compose.yml build --force-rm --no-cache && docker compose -f docker-compose.yml up
```

This command builds and starts:

chat-bot (Go) on port 7004

matching-api (Flask) on port 5001

mongo (MongoDB) on port 27017


### 🐳 Service Description
#### 1. Question API (Go)

Port: 8080

Purpose: Accepts user questions, calls the Matching API to get a reply, and saves both the question and reply to MongoDB. You can choose the comparison algorithm used by specifying the "algorithm" field in the body, which can be one of ["words", "cosine", "fuzzy"]. If not specified it defaults to cosine.

Endpoint:

POST /api/question
Content-Type: application/json

Request:
```json
{
    "sessionId": "myId",
    "query": "Who is that?",
    "algorithm": "cosine"
}
}
```

Response:
```json
{
    "matched": true,
    "reply": "Gen Z in Nashville are 106% more likely to find out about new brands and products through vlogs compared to the average person",
    "score": 4.0
}
```


#### 2. Matching API (Python + Flask)

Port: 5001

Purpose: Accepts a query and returns the best-matched response based on semantic similarity.

This api consists of 3 endpoints
POST /match-cosine
This compares the user's query with the replies given using the cosine comparison method for comparing sentences.

POST /match-words
This compares the number of common words in the users query with those of the sentences and returns the most common, given a threashold.

POST /match-fuzzy
This compares the user's query with the replies given using a fuzzing comparison method for comparing sentences. (THIS DOES NOT work as expected but there was no time to fix it)

Example Endpoint:

POST /match-cosine
Content-Type: application/json

Request:
```json
{
  "query": "How old are you?"
}
```

Response:
```json
{
    "matched": true,
    "reply": "Gen Z in Nashville are 106% more likely to find out about new brands and products through vlogs compared to the average person",
    "score": 4.0
}
```


#### 3. MongoDB

Port: 27017

Database: test

Collection: userhistory

You can access MongoDB locally (e.g., via MongoDB Compass) or from a script:

mongodb://root:password@localhost:27017/

Example stored document:

```json
  {
  "sessionId": "myID",
  "createdAt": {
    "$date": "2025-09-03T14:09:26.173Z"
  },
  "messages": [
    {
      "role": "question",
      "text": "Who are you?",
      "timestamp": {
        "$date": "2025-09-03T14:09:26.173Z"
      }
    },
    {
      "role": "reply",
      "text": "I am me",
      "timestamp": {
        "$date": "2025-09-03T14:10:07.144Z"
      }
    }
  ]
}
```

### ✅ To Do

 - [ ] Add authentication - if required

 - [ ] Tidy up and move hardcoded env variables to a file

 - [ ] Refine structure

 - [ ] WRITE TESTS: tesing performance and accuracy of each method and for different use cases

 - [ ] Add rate limiting - important since the endpoint is open for exploitation

 - [ ] Deploy to cloud (e.g., AWS/GCP/DigitalOcean) -  required

 - [ ] Make the db history writting a background job

 - [ ] Add chronjob that removes old conversations from mongodb

 - [ ] Create function that deletes mongo db entry

 - [ ] Fix fuzzy endpoint

 - [ ] Use preparatory LLM method for making the query more concise and comparing with replies - downloading the model is SLOW and using open LLMs is not an option since the data is the intellectual property of the company and user should be informed about their questions being processed by an Open LLM Model

 - [ ] Investigate why response is so slow


### 📝 **Notes:** 

Install Ollama:
https://ollama.com/download

ollama pull mistral

Use it in Python:
pip install ollama --> it was very time consuming to install the model so I omitted it, but the idea is that you can probably use an LLM to create a simpler question that will then will be able to find a reply int he set. https://ollama.com/download
