# GWI Jedi Team - Backend Engineering Challenge

This repository contains the solution for the Jedi Team Backend Engineering Challenge at GWI. It is a Docker Compose project implementing a chatbot that helps GWI's clients answer questions based on market research data.

---

## Project Overview

The chatbot is designed to answer questions based on GWI market research data. The system persists chat history and allows users to continue conversations or open multiple chats. It integrates an LLM (Ollama), a Python service for retrieval-augmented generation (RAG), and a PostgreSQL database. 


### Architecture 🏗️

The project consists of six containers:

1. **Gateway (Go)** – main backend service handling API requests.
2. **LLM (Ollama)** – container running the language model.
3. **RAG Service (Python)** – performs retrieval-augmented generation using provided research data.
4. **PostgreSQL** – stores chat history.
5. **Tests** – runs integration tests testing basic chat flow.
6. **Grafana/k6** – performs benchmarks.

### Basic Chat Flow 🤖
```text
                                ┌──────────┐                          
                                │          │                          
                  ┌─────────────►PostgresDB│                          
                  │             │Container │                          
                  │       ┌┬────►          │                          
         (4) Save │       ││    └──────────┘                          
         Messages │       ││                                          
                  │       ││  (2) Query                               
                  │       ││  For Previous Chat                       
                  │       ││     Messages                             
                  │       ││                                                        
              ┌───┼───────┘▼┐                           ┌────────────┐
Req/Resp      │             ┼───────────────────────────►            │
  ───────────►│   Gateway   ◄───────────────────────────┐   Ollama   │
  ◄───────────│  Container  │   (3) Query Ollama        │ Container  │
              |             │    w/ augmented prompt    │            │
              └─────────┬─▲─┘    & chat context         └────────────┘
                        │ │                                               
                        │ │                                           
                        │ │(1) Retrieve and                           
                        │ │ Augment Prompt w/ topK                             
                        │ │ Relevant Sentences                                                                      
                        │ │ ┌────────────┐                            
                        │ │ │            │                            
                        │ └─┼    RAG     │                            
                        └───► Container  │                            
                            │            │                            
                            └────────────┘                            
```

## Getting Started 🚀

### Requirements 📦
1. Make sure you have `docker` & `docker-compose`:
    ```bash
    docker -v
    docker-compose -v
    ```
2. Make sure Ollama can run in a GPU-enabled container:
 - Check NVIDIA GPU (should list your graphics card):
    ```gash
    nvidia-smi
    ```
 - Check NVIDIA Container Toolkit:
    ```gash
    docker run --rm --gpus all nvidia/cuda:12.1-base nvidia-smi
    ```
    > If want to test without GPU-acceleration, remove the `deploy` section in the ollama container in docker compose and run a lightweight model like `gemma:2b`

### Setup & Run 🛠️
1. **Clone the repository**
2. **Create a `.env` file in the root directory of this project:**
    ```.env
    # DB CONFIG
    DB_HOST=postgres
    DB_PORT=5432
    POSTGRES_USER=myuser
    POSTGRES_PASSWORD=mypassword
    POSTGRES_DB=userdb
    DB_SSL_MODE=disable

    # OLLAMA CONFIG
    OLLAMA_MODEL=granite3-dense:8b # Choose a model based on capabilities. `granite3-dense:8b` is advised, use `gemma:2b` if running on CPU.
    OLLAMA_URL=http://ollama:11434/api/chat
    OLLAMA_STREAM=false

    # RAG CONFIG
    RAG_URL=http://rag:8000/retrieve
    RAG_TOPK=30 # Top_K Sentences to be picked by RAG container provided as context to the LLM
    ```
3. **Pull required Docker images:**
    ```bash
    docker compose pull
    ```
4. **Place market research data under `data/data.md`**

5. **Running the project:**
    ```bash
    docker compose up -d
    ```
6. **(Optional) Run Integration & Benchmark tests**
    ```bash
    docker compose up --profile tests
    ```

## API Documentation 📚

### Base URL
```text
http://localhost:8080
```

### Endpoints
1. Start a new chat
 - URL: `/chat`
 - Method: `POST`
 - Input:
    ```json
    {
        "user_id": "string",    // string: unique identifier of the user
        "message": "string"     // string: initial message from the user
    }
    ```
 - Output:
    ```json
    {
        "chat_id": 123,          // int: unique identifier of the chat
        "title": "string",       // string: auto-generated title for the chat
        "message": "string",     // string: user's message
        "response": "string",    // string: chatbot's response
        "created_at": "timestamp" // string: ISO 8601 timestamp of the message
    }
    ```
2. Continue an existing chat
 - URL: `/chat/:chat_id`
 - Method: `POST`
 - Input:
    ```json
    {
        "user_id": "string",    // string: unique identifier of the user
        "message": "string"     // string: new message from the user
    }
    ```
 - Output:
    ```json
    {
        "chat_id": 123,          // int: unique identifier of the chat
        "title": "string",       // string: chat title
        "message": "string",     // string: user's message
        "response": "string",    // string: chatbot's response
        "created_at": "timestamp" // string: ISO 8601 timestamp of the message
    }
    ```
3. Get chat history
 - URL: `/chat/:chat_id/history`
 - Method: `GET`
 - Output:
    ```json
    {
        "chat_id": 123,          // int: unique identifier of the chat
        "title": "string",       // string: chat title
        "messages": [
            {
                "message": "string",        // string: user's message
                "response": "string",       // string: chatbot's response
                "created_at": "timestamp"   // string: ISO 8601 timestamp
            }
        ]
    }
    ```
4. Get user's chat history
 - URL: `/chat/users/:user_id`
 - Method: `GET`
 - Output:
    ```json
    [
        {
            "id": 38,                         // int: chat ID
            "user_id": 5,                     // int: user ID
            "title": "string",                // string: chat title
            "last_updated": "timestamp"       // string: ISO 8601 timestamp
        },
        {
            "id": 39,
            "user_id": 5,
            "title": "string",
            "last_updated": "timestamp"
        },
        {
            "id": 40,
            "user_id": 5,
            "title": "string",
            "last_updated": "timestamp"
        }
    ]
    ```