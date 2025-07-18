package http

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/loukaspe/jedi-team-challenge/pkg/logger"
	"net/http"
)

type SubtractHandler struct {
	logger   logger.LoggerInterface
	upgrader *websocket.Upgrader
}

func NewSubtractHandler(
	upgrader *websocket.Upgrader,
	logger logger.LoggerInterface,
) *SubtractHandler {
	return &SubtractHandler{
		upgrader: upgrader,
		logger:   logger,
	}
}

func (handler SubtractHandler) SubtractController(w http.ResponseWriter, r *http.Request) {
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

		result := data.A - data.B
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Result: %.2f", result)))
	}
}
