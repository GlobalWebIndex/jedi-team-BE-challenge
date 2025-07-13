package mcp

import (
	"context"
	"errors"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) InitialiseSSEServer() *server.SSEServer {
	add := mcp.NewTool("add",
		mcp.WithDescription("Add two numbers"),
		mcp.WithNumber("a",
			mcp.Required(),
			mcp.Description("First number"),
		),
		mcp.WithNumber("b",
			mcp.Required(),
			mcp.Description("Second number"),
		),
	)

	// Add subtraction tool
	subtract := mcp.NewTool("subtract",
		mcp.WithDescription("Subtract second number from first number"),
		mcp.WithNumber("a",
			mcp.Required(),
			mcp.Description("First number"),
		),
		mcp.WithNumber("b",
			mcp.Required(),
			mcp.Description("Second number"),
		),
	)

	// Add multiplication tool
	multiply := mcp.NewTool("multiply",
		mcp.WithDescription("Multiply two numbers"),
		mcp.WithNumber("a",
			mcp.Required(),
			mcp.Description("First number"),
		),
		mcp.WithNumber("b",
			mcp.Required(),
			mcp.Description("Second number"),
		),
	)

	// Add division tool
	divide := mcp.NewTool("divide",
		mcp.WithDescription("Divide first number by second number"),
		mcp.WithNumber("a",
			mcp.Required(),
			mcp.Description("Numerator"),
		),
		mcp.WithNumber("b",
			mcp.Required(),
			mcp.Description("Denominator"),
		),
	)

	// Add percentage tool
	percentage := mcp.NewTool("percentage",
		mcp.WithDescription("Calculate what percentage the first number is of the second number"),
		mcp.WithNumber("a",
			mcp.Required(),
			mcp.Description("Part value"),
		),
		mcp.WithNumber("b",
			mcp.Required(),
			mcp.Description("Total value"),
		),
	)

	// Add tool handlers
	s.mcpServer.AddTool(add, addHandler)
	s.mcpServer.AddTool(subtract, subtractHandler)
	s.mcpServer.AddTool(multiply, multiplyHandler)
	s.mcpServer.AddTool(divide, divideHandler)
	s.mcpServer.AddTool(percentage, percentageHandler)

	return server.NewSSEServer(
		s.mcpServer,
		server.WithStaticBasePath("/"),
		server.WithSSEEndpoint("/sse"),
		server.WithMessageEndpoint("/message"),
	)
}

// addHandler processes addition requests, adding two numbers and returning the formatted result
// Returns an error if inputs are not valid numbers
func addHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	a, err := request.RequireFloat("a")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	b, err := request.RequireFloat("b")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := a + b
	return mcp.NewToolResultText(fmt.Sprintf("%.2f", result)), nil
}

// subtractHandler processes subtraction requests, subtracting the second number from the first
// Returns an error if inputs are not valid numbers
func subtractHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	a, err := request.RequireFloat("a")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	b, err := request.RequireFloat("b")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := a - b
	return mcp.NewToolResultText(fmt.Sprintf("%.2f", result)), nil
}

// multiplyHandler processes multiplication requests, multiplying two numbers together
// Returns an error if inputs are not valid numbers
func multiplyHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	a, err := request.RequireFloat("a")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	b, err := request.RequireFloat("b")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := a * b
	return mcp.NewToolResultText(fmt.Sprintf("%.2f", result)), nil
}

// divideHandler processes division requests, dividing the first number by the second
// Returns an error if inputs are not valid numbers or if dividing by zero
func divideHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	a, err := request.RequireFloat("a")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	b, err := request.RequireFloat("b")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if b == 0 {
		return nil, errors.New("cannot divide by zero")
	}

	result := a / b
	return mcp.NewToolResultText(fmt.Sprintf("%.2f", result)), nil
}

// percentageHandler calculates what percentage the first number is of the second number
// Returns an error if inputs are not valid numbers or if the total value is zero
func percentageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	a, err := request.RequireFloat("a")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	b, err := request.RequireFloat("b")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if b == 0 {
		return nil, errors.New("total value cannot be zero")
	}

	result := (a / b) * 100
	return mcp.NewToolResultText(fmt.Sprintf("%.2f%%", result)), nil
}
