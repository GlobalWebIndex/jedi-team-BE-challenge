package model

type GetData struct {
	Name    string
	Age     int
	City    string
	Pincode int
}

type UserHistory struct {
	// ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SessionId string `bson:"sessionId" json:"sessionId"`
	// Question  string             `bson:"question" json:"question"`
	// Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
}

type MatchingAlgorithm string

const (
	MatchingAlgorithmWords  MatchingAlgorithm = "words"
	MatchingAlgorithmCosine MatchingAlgorithm = "cosine"
	MatchingAlgorithmFuzzy  MatchingAlgorithm = "fuzzy"
)

type SubmitQuestionRequest struct {
	SessionID         string            `json:"sessionId" binding:"required"`
	Query             string            `json:"query" binding:"required"`
	MatchingAlgorithm MatchingAlgorithm `json:"algorithm"`
}
