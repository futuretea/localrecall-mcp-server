package mcp

import (
	"context"
	"maps"
	"net/http"
	"slices"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/futuretea/localrecall-mcp-server/pkg/client"
	"github.com/futuretea/localrecall-mcp-server/pkg/core/config"
	"github.com/futuretea/localrecall-mcp-server/pkg/core/logging"
	"github.com/futuretea/localrecall-mcp-server/pkg/core/version"
	"github.com/futuretea/localrecall-mcp-server/pkg/toolset"
	localrecallToolset "github.com/futuretea/localrecall-mcp-server/pkg/toolset/localrecall"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const authorizationKey contextKey = "Authorization"

// Configuration wraps the static configuration with additional runtime components
type Configuration struct {
	*config.StaticConfig
}

// Server represents the MCP server
type Server struct {
	configuration     *Configuration
	server            *server.MCPServer
	enabledTools      []string
	localRecallClient *client.Client
}

// NewServer creates a new MCP server with the given configuration
func NewServer(configuration Configuration) (*Server, error) {
	serverOptions := []server.ServerOption{
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithToolCapabilities(true),
		server.WithLogging(),
	}

	localRecallClient := client.NewClient(
		configuration.LocalRecallURL,
		configuration.LocalRecallAPIKey,
	)
	logging.Info("LocalRecall client initialized with URL: %s", configuration.LocalRecallURL)

	s := &Server{
		configuration:     &configuration,
		server:            server.NewMCPServer(version.BinaryName, version.Version, serverOptions...),
		localRecallClient: localRecallClient,
	}

	if err := s.registerTools(); err != nil {
		return nil, err
	}

	return s, nil
}

// registerTools registers all available tools based on configuration
func (s *Server) registerTools() error {
	localrecallTs := &localrecallToolset.Toolset{
		DefaultCollection: s.configuration.LocalRecallCollection,
	}

	wrappedClient := &toolset.LocalRecallClient{
		Client: s.localRecallClient,
	}

	for _, tool := range localrecallTs.GetTools(wrappedClient) {
		if !s.shouldEnableTool(tool.Tool.Name) {
			continue
		}
		s.registerTool(s.configureTool(tool, wrappedClient))
	}

	logging.Info("MCP server initialized with %d tools", len(s.enabledTools))
	return nil
}

// shouldEnableTool determines if a tool should be enabled based on configuration
func (s *Server) shouldEnableTool(toolName string) bool {
	if slices.Contains(s.configuration.DisabledTools, toolName) {
		return false
	}
	if len(s.configuration.EnabledTools) > 0 {
		return slices.Contains(s.configuration.EnabledTools, toolName)
	}
	return true
}

// configureTool creates a configured tool handler that uses server configuration
func (s *Server) configureTool(tool toolset.ServerTool, wrappedClient *toolset.LocalRecallClient) toolset.ServerTool {
	return toolset.ServerTool{
		Tool: tool.Tool,
		Handler: func(client interface{}, params map[string]interface{}) (string, error) {
			// Inject default output format if not specified
			if _, hasFormat := params["format"]; !hasFormat && s.configuration.ListOutput != "" {
				params["format"] = s.configuration.ListOutput
			}

			// Enforce collection isolation: always override collection_name
			if s.configuration.LocalRecallCollection != "" {
				params["collection_name"] = s.configuration.LocalRecallCollection
			}

			return tool.Handler(wrappedClient, params)
		},
	}
}

func contextFunc(ctx context.Context, r *http.Request) context.Context {
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		return context.WithValue(ctx, authorizationKey, authHeader)
	}
	return ctx
}

// registerTool registers a single tool with the MCP server
func (s *Server) registerTool(tool toolset.ServerTool) {
	s.server.AddTool(tool.Tool, server.ToolHandlerFunc(func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		logging.Debug("Tool %s called with params: %v", tool.Tool.Name, request.Params.Arguments)

		params := make(map[string]interface{})
		if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
			maps.Copy(params, args)
		}

		result, err := tool.Handler(nil, params)
		return NewTextResult(result, err), nil
	}))
	s.enabledTools = append(s.enabledTools, tool.Tool.Name)
	logging.Info("Registered tool: %s", tool.Tool.Name)
}

// ServeStdio starts the MCP server in stdio mode
func (s *Server) ServeStdio() error {
	logging.Info("Starting MCP server in stdio mode")
	return server.ServeStdio(s.server)
}

// ServeSse starts the MCP server in SSE mode
func (s *Server) ServeSse(baseURL string, httpServer *http.Server) *server.SSEServer {
	logging.Info("Starting MCP server in SSE mode")

	options := []server.SSEOption{
		server.WithHTTPServer(httpServer),
		server.WithSSEContextFunc(contextFunc),
	}
	if baseURL != "" {
		options = append(options, server.WithBaseURL(baseURL))
	}

	return server.NewSSEServer(s.server, options...)
}

// ServeHTTP starts the MCP server in HTTP mode
func (s *Server) ServeHTTP(httpServer *http.Server) *server.StreamableHTTPServer {
	logging.Info("Starting MCP server in HTTP mode")

	options := []server.StreamableHTTPOption{
		server.WithHTTPContextFunc(contextFunc),
		server.WithStreamableHTTPServer(httpServer),
		server.WithStateLess(true),
	}

	return server.NewStreamableHTTPServer(s.server, options...)
}

// GetEnabledTools returns the list of enabled tools
func (s *Server) GetEnabledTools() []string {
	return s.enabledTools
}

// Close cleans up the server resources
func (s *Server) Close() {
	logging.Info("Closing MCP server")
}

// NewTextResult creates a standardized text result for tool responses
func NewTextResult(content string, err error) *mcp.CallToolResult {
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: err.Error(),
				},
			},
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: content,
			},
		},
	}
}
