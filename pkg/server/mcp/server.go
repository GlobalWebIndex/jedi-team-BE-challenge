package mcp

import (
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	mcpServer *server.MCPServer
}

func NewServer(
	mcpServer *server.MCPServer,
) *Server {
	return &Server{
		mcpServer: mcpServer,
	}
}
