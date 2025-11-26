package mcpservers

import (
	"fmt"
	"log"
)

// BaseMCPServer provides base functionality for MCP servers
type BaseMCPServer interface {
	RegisterTools()
	Run()
	AddTool(name, description string, handler func(args map[string]interface{}) (interface{}, error))
}

// BaseServerImpl provides a basic implementation of MCP server
type BaseServerImpl struct {
	Name  string
	Tools map[string]Tool
}

// Tool represents an MCP tool
type Tool struct {
	Name        string
	Description string
	Handler     func(args map[string]interface{}) (interface{}, error)
}

// NewBaseServer creates a new base server
func NewBaseServer(name string) *BaseServerImpl {
	return &BaseServerImpl{
		Name:  name,
		Tools: make(map[string]Tool),
	}
}

// AddTool adds a tool to the server
func (s *BaseServerImpl) AddTool(name, description string, handler func(args map[string]interface{}) (interface{}, error)) {
	s.Tools[name] = Tool{
		Name:        name,
		Description: description,
		Handler:     handler,
	}
	log.Printf("Registered tool: %s - %s", name, description)
}

// Run starts the MCP server (placeholder implementation)
func (s *BaseServerImpl) Run() {
	log.Printf("Starting %s MCP server", s.Name)
	// In real implementation, this would start the MCP server loop
	// For now, just log that we're running
}

// RegisterTools should be implemented by concrete servers
func (s *BaseServerImpl) RegisterTools() {
	// This will be overridden by concrete implementations
}

// AIProcessingError represents an error in AI processing
type AIProcessingError struct {
	Message string
}

func (e AIProcessingError) Error() string {
	return fmt.Sprintf("AI processing error: %s", e.Message)
}

// ModelLoadError represents an error loading a model
type ModelLoadError struct {
	Message string
}

func (e ModelLoadError) Error() string {
	return fmt.Sprintf("Model load error: %s", e.Message)
}
