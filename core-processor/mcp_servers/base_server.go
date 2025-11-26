package mcp_servers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// BaseMCPServer provides base functionality for MCP servers
type BaseMCPServer interface {
	RegisterTools()
	Run() error
	Stop()
	AddTool(name, description string, handler func(args map[string]interface{}) (interface{}, error))
}

// MCPServerConfig holds configuration for MCP server
type MCPServerConfig struct {
	Name        string
	Version     string
	Transport   string // "stdio" or "tcp"
	Address     string // TCP address if transport is "tcp"
	Timeout     time.Duration
	MaxRetries  int
}

// BaseServerImpl provides a basic implementation of MCP server
type BaseServerImpl struct {
	Config     MCPServerConfig
	Tools      map[string]Tool
	listener   net.Listener
	running    bool
	stopCh     chan struct{}
	wg         sync.WaitGroup
	mu         sync.RWMutex
}

// Tool represents an MCP tool
type Tool struct {
	Name        string
	Description string
	Handler     func(args map[string]interface{}) (interface{}, error)
}

// MCPRequest represents an MCP request
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents an MCP response
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ToolInfo represents tool information for listing
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// NewBaseServer creates a new base server
func NewBaseServer(config MCPServerConfig) *BaseServerImpl {
	return &BaseServerImpl{
		Config: config,
		Tools:  make(map[string]Tool),
		stopCh: make(chan struct{}),
	}
}

// NewBaseServerWithDefaults creates a base server with default config
func NewBaseServerWithDefaults(name string) *BaseServerImpl {
	config := MCPServerConfig{
		Name:       name,
		Version:    "1.0.0",
		Transport:  "stdio",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}
	return &BaseServerImpl{
		Config: config,
		Tools:  make(map[string]Tool),
		stopCh: make(chan struct{}),
	}
}

// AddTool adds a tool to the server
func (s *BaseServerImpl) AddTool(name, description string, handler func(args map[string]interface{}) (interface{}, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Tools[name] = Tool{
		Name:        name,
		Description: description,
		Handler:     handler,
	}
	log.Printf("Registered tool: %s - %s", name, description)
}

// Run starts the MCP server
func (s *BaseServerImpl) Run() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server is already running")
	}
	s.running = true
	s.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start server based on transport type
	switch s.Config.Transport {
	case "stdio":
		return s.runStdio(ctx)
	case "tcp":
		return s.runTCP(ctx)
	default:
		return fmt.Errorf("unsupported transport: %s", s.Config.Transport)
	}
}

// runStdio runs the server over stdio
func (s *BaseServerImpl) runStdio(ctx context.Context) error {
	log.Printf("Starting %s MCP server on stdio", s.Config.Name)

	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.stopCh:
			log.Println("Server stop signal received")
			return nil
		default:
			// Read JSON-RPC request from stdin
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return fmt.Errorf("failed to read request: %w", err)
			}

			// Process request
			response := s.processRequest(line)

			// Write response to stdout
			responseJSON, err := json.Marshal(response)
			if err != nil {
				log.Printf("Failed to marshal response: %v", err)
				continue
			}

			if _, err := writer.Write(responseJSON); err != nil {
				return fmt.Errorf("failed to write response: %w", err)
			}
			if _, err := writer.Write([]byte("\n")); err != nil {
				return fmt.Errorf("failed to write newline: %w", err)
			}
			writer.Flush()
		}
	}
}

// runTCP runs the server over TCP
func (s *BaseServerImpl) runTCP(ctx context.Context) error {
	log.Printf("Starting %s MCP server on %s", s.Config.Name, s.Config.Address)

	var err error
	s.listener, err = net.Listen("tcp", s.Config.Address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.Config.Address, err)
	}
	defer s.listener.Close()

	// Accept connections in a goroutine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.stopCh:
				return
			default:
				conn, err := s.listener.Accept()
				if err != nil {
					if s.running {
						log.Printf("Failed to accept connection: %v", err)
					}
					continue
				}
				
				// Handle connection in goroutine
				s.wg.Add(1)
				go func(c net.Conn) {
					defer s.wg.Done()
					defer c.Close()
					s.handleConnection(ctx, c)
				}(conn)
			}
		}
	}()

	// Wait for stop signal
	<-ctx.Done()
	return ctx.Err()
}

// handleConnection handles a single TCP connection
func (s *BaseServerImpl) handleConnection(ctx context.Context, conn net.Conn) {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Set read timeout
			conn.SetReadDeadline(time.Now().Add(s.Config.Timeout))

			// Read request
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Printf("Failed to read from connection: %v", err)
				return
			}

			// Process request
			response := s.processRequest(line)

			// Write response
			responseJSON, err := json.Marshal(response)
			if err != nil {
				log.Printf("Failed to marshal response: %v", err)
				continue
			}

			if _, err := writer.Write(responseJSON); err != nil {
				log.Printf("Failed to write response: %v", err)
				return
			}
			if _, err := writer.Write([]byte("\n")); err != nil {
				log.Printf("Failed to write newline: %v", err)
				return
			}
			writer.Flush()
		}
	}
}

// processRequest processes a single JSON-RPC request
func (s *BaseServerImpl) processRequest(jsonData string) MCPResponse {
	var request MCPRequest
	if err := json.Unmarshal([]byte(jsonData), &request); err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error: &MCPError{
				Code:    -32700,
				Message: "Parse error",
				Data:    err.Error(),
			},
		}
	}

	// Handle the request based on method
	result, err := s.handleMethod(request.Method, request.Params)

	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &MCPError{
				Code:    -32000,
				Message: "Server error",
				Data:    err.Error(),
			},
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// handleMethod handles different MCP methods
func (s *BaseServerImpl) handleMethod(method string, params interface{}) (interface{}, error) {
	switch method {
	case "initialize":
		return s.handleInitialize(params)
	case "tools/list":
		return s.handleListTools(params)
	case "tools/call":
		return s.handleToolCall(params)
	default:
		return nil, fmt.Errorf("method not found: %s", method)
	}
}

// handleInitialize handles the initialize method
func (s *BaseServerImpl) handleInitialize(params interface{}) (interface{}, error) {
	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    s.Config.Name,
			"version": s.Config.Version,
		},
	}, nil
}

// handleListTools handles the tools/list method
func (s *BaseServerImpl) handleListTools(params interface{}) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]ToolInfo, 0, len(s.Tools))
	for name, tool := range s.Tools {
		tools = append(tools, ToolInfo{
			Name:        name,
			Description: tool.Description,
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{
						"type":        "string",
						"description": "Input text for the tool",
					},
				},
				"required": []string{"text"},
			},
		})
	}

	return map[string]interface{}{
		"tools": tools,
	}, nil
}

// handleToolCall handles the tools/call method
func (s *BaseServerImpl) handleToolCall(params interface{}) (interface{}, error) {
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid params format")
	}

	name, ok := paramsMap["name"].(string)
	if !ok {
		return nil, fmt.Errorf("tool name is required")
	}

	s.mu.RLock()
	tool, exists := s.Tools[name]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	arguments, _ := paramsMap["arguments"].(map[string]interface{})
	_, err := tool.Handler(arguments)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Tool %s executed successfully", name),
			},
		},
		"isError": false,
	}, nil
}

// Stop stops the server gracefully
func (s *BaseServerImpl) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopCh)

	if s.listener != nil {
		s.listener.Close()
	}

	// Wait for all goroutines to finish
	s.wg.Wait()

	log.Printf("%s MCP server stopped", s.Config.Name)
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
