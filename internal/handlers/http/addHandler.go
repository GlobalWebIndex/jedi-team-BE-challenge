package http

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/loukaspe/jedi-team-challenge/pkg/logger"
	"net/http"
)

type Payload struct {
	A float64 `json:"a"`
	B float64 `json:"b"`
}

type AddHandler struct {
	logger   logger.LoggerInterface
	upgrader *websocket.Upgrader
}

func NewAddHandler(
	upgrader *websocket.Upgrader,
	logger logger.LoggerInterface,
) *AddHandler {
	return &AddHandler{
		upgrader: upgrader,
		logger:   logger,
	}
}

func (handler AddHandler) AddController(w http.ResponseWriter, r *http.Request) {
	conn, err := handler.upgrader.Upgrade(w, r, nil)
	if err != nil {
		handler.logger.Error("Upgrade error:", map[string]interface{}{
			"errorMessage": err.Error(),
		})
		return
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Read error:", err)
			break
		}

		var data Payload
		if err := json.Unmarshal(msg, &data); err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("Invalid input"))
			continue
		}

		result := data.A + data.B
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Result: %.2f", result)))
	}
}
