package controller

import (
	"challenge/api/model"
	"challenge/database/handler"
	"encoding/json"
	"log"
	"net/http"

	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

type RequestPayload struct {
	Query     string  `json:"query"`
	Threshold float64 `json:"threshold"`
}

type ResponsePayload struct {
	Matched bool    `json:"matched"`
	Reply   string  `json:"reply,omitempty"`
	Score   float64 `json:"score,omitempty"`
}

type MatchResponse struct {
	Matched bool    `json:"matched" binding:"required"`
	Reply   string  `json:"reply"`
	Score   float64 `json:"score"`
}

func RemoveConversation(c *gin.Context) {
	sessionID := c.Query("sessionID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionID is required"})
		return
	}

	err := handler.RemoveSessionData(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessionID)
}

func SubmitQuestion(c *gin.Context) {

	// need the sessionID - required field
	var req model.SubmitQuestionRequest

	// Bind and validate the JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId and question are required"})
		return
	}

	// add question to mongo for the sessionId, if the session id doesnt exist it will create a new entry in mongo
	err := handler.AddMessage(req.SessionID, "question", req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add question to db"})
		return
	}

	// submit question to the matching api and get answer
	matchingAlgorithm := req.MatchingAlgorithm

	// Set default if it's empty
	if matchingAlgorithm == "" {
		matchingAlgorithm = model.MatchingAlgorithmCosine
	}

	// url := "http://localhost:5001/match-" + string(matchingAlgorithm)
	url := "http://matching-api:5001/match-" + string(matchingAlgorithm) // TODO - env variable

	response, err := PostJSON[RequestPayload, model.SubmitQuestionRequest](url, req)
	if err != nil {
		log.Fatalf("Fatal error calling matching API: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call matching API"})
		return
	}

	fmt.Println(response.Reply)

	if response.Reply == "" {
		response.Reply = "Sorry we are unable to answer your question."
	}

	// add reply to mongo for the sessionId
	err = handler.AddMessage(req.SessionID, "reply", response.Reply)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add question to db"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"matched": response.Matched,
		"reply":   response.Reply,
		"score":   response.Score,
	})

	// fmt.Printf("Matched: %v\nResponse: %s\nScore: %.2f\n", response.Matched, response.Reply, response.Score)
}

// PostJSON sends a POST request with JSON-encoded data and decodes the JSON response.
func PostJSON[T any, R any](url string, payload model.SubmitQuestionRequest) (*ResponsePayload, error) {
	// Marshal request payload to JSON
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{Timeout: 20 * time.Second}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Decode response
	var response ResponsePayload
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w\nRaw body: %s", err, string(body))
	}

	return &response, nil
}
