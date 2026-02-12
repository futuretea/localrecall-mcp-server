package localrecall

import (
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/futuretea/localrecall-mcp-server/pkg/toolset"
)

// Toolset implements the toolset.Toolset interface for LocalRecall
type Toolset struct {
	DefaultCollection string
}

// GetName returns the name of the toolset
func (t *Toolset) GetName() string {
	return "localrecall"
}

// GetDescription returns the description of the toolset
func (t *Toolset) GetDescription() string {
	return "LocalRecall knowledge base management tools"
}

// GetTools returns all LocalRecall tools
func (t *Toolset) GetTools(client interface{}) []toolset.ServerTool {
	tools := []toolset.ServerTool{}

	// Determine if we should use simplified tools (with default collection)
	useDefaultCollection := t.DefaultCollection != ""

	if useDefaultCollection {
		// Tools with default collection - collection_name is optional
		tools = append(tools,
			toolset.ServerTool{
				Tool: mcp.Tool{
					Name:        "search",
					Description: "Search content in LocalRecall collection '" + t.DefaultCollection + "'",
					InputSchema: mcp.ToolInputSchema{
						Type: "object",
						Properties: map[string]interface{}{
							"query": map[string]interface{}{
								"type":        "string",
								"description": "The search query",
							},
							"max_results": map[string]interface{}{
								"type":        "number",
								"description": "Maximum number of results to return (default: 5)",
							},
							"collection_name": map[string]interface{}{
								"type":        "string",
								"description": "Optional: override the default collection",
							},
						},
						Required: []string{"query"},
					},
				},
				Handler: SearchHandler,
			},
			toolset.ServerTool{
				Tool: mcp.Tool{
					Name:        "add_document",
					Description: "Add a document to LocalRecall collection '" + t.DefaultCollection + "'",
					InputSchema: mcp.ToolInputSchema{
						Type: "object",
						Properties: map[string]interface{}{
							"filename": map[string]interface{}{
								"type":        "string",
								"description": "The filename for the document",
							},
							"file_path": map[string]interface{}{
								"type":        "string",
								"description": "Path to the file to upload (mutually exclusive with file_content)",
							},
							"file_content": map[string]interface{}{
								"type":        "string",
								"description": "File content as string (mutually exclusive with file_path)",
							},
							"collection_name": map[string]interface{}{
								"type":        "string",
								"description": "Optional: override the default collection",
							},
						},
						Required: []string{"filename"},
					},
				},
				Handler: AddDocumentHandler,
			},
			toolset.ServerTool{
				Tool: mcp.Tool{
					Name:        "list_files",
					Description: "List files in LocalRecall collection '" + t.DefaultCollection + "'",
					InputSchema: mcp.ToolInputSchema{
						Type: "object",
						Properties: map[string]interface{}{
							"collection_name": map[string]interface{}{
								"type":        "string",
								"description": "Optional: override the default collection",
							},
						},
					},
				},
				Handler: ListFilesHandler,
			},
			toolset.ServerTool{
				Tool: mcp.Tool{
					Name:        "delete_entry",
					Description: "Delete an entry from LocalRecall collection '" + t.DefaultCollection + "'",
					InputSchema: mcp.ToolInputSchema{
						Type: "object",
						Properties: map[string]interface{}{
							"entry": map[string]interface{}{
								"type":        "string",
								"description": "The filename of the entry to delete",
							},
							"collection_name": map[string]interface{}{
								"type":        "string",
								"description": "Optional: override the default collection",
							},
						},
						Required: []string{"entry"},
					},
				},
				Handler: DeleteEntryHandler,
			},
			toolset.ServerTool{
				Tool: mcp.Tool{
					Name:        "get_entry_content",
					Description: "Get the content of a specific entry in LocalRecall collection '" + t.DefaultCollection + "'",
					InputSchema: mcp.ToolInputSchema{
						Type: "object",
						Properties: map[string]interface{}{
							"entry": map[string]interface{}{
								"type":        "string",
								"description": "The filename of the entry to retrieve",
							},
							"collection_name": map[string]interface{}{
								"type":        "string",
								"description": "Optional: override the default collection",
							},
						},
						Required: []string{"entry"},
					},
				},
				Handler: GetEntryContentHandler,
			},
			toolset.ServerTool{
				Tool: mcp.Tool{
					Name:        "register_source",
					Description: "Register an external source for LocalRecall collection '" + t.DefaultCollection + "'",
					InputSchema: mcp.ToolInputSchema{
						Type: "object",
						Properties: map[string]interface{}{
							"url": map[string]interface{}{
								"type":        "string",
								"description": "The URL of the external source",
							},
							"update_interval": map[string]interface{}{
								"type":        "number",
								"description": "Update interval in seconds (0 or omit for no auto-update)",
							},
							"collection_name": map[string]interface{}{
								"type":        "string",
								"description": "Optional: override the default collection",
							},
						},
						Required: []string{"url"},
					},
				},
				Handler: RegisterSourceHandler,
			},
			toolset.ServerTool{
				Tool: mcp.Tool{
					Name:        "remove_source",
					Description: "Remove an external source from LocalRecall collection '" + t.DefaultCollection + "'",
					InputSchema: mcp.ToolInputSchema{
						Type: "object",
						Properties: map[string]interface{}{
							"url": map[string]interface{}{
								"type":        "string",
								"description": "The URL of the external source to remove",
							},
							"collection_name": map[string]interface{}{
								"type":        "string",
								"description": "Optional: override the default collection",
							},
						},
						Required: []string{"url"},
					},
				},
				Handler: RemoveSourceHandler,
			},
			toolset.ServerTool{
				Tool: mcp.Tool{
					Name:        "list_sources",
					Description: "List external sources for LocalRecall collection '" + t.DefaultCollection + "'",
					InputSchema: mcp.ToolInputSchema{
						Type: "object",
						Properties: map[string]interface{}{
							"collection_name": map[string]interface{}{
								"type":        "string",
								"description": "Optional: override the default collection",
							},
						},
					},
				},
				Handler: ListSourcesHandler,
			},
		)
	} else {
		// Tools without default collection - collection_name is required
		tools = append(tools,
			toolset.ServerTool{
				Tool: mcp.Tool{
					Name:        "search",
					Description: "Search content in a LocalRecall collection",
					InputSchema: mcp.ToolInputSchema{
						Type: "object",
						Properties: map[string]interface{}{
							"collection_name": map[string]interface{}{
								"type":        "string",
								"description": "The name of the collection to search",
							},
							"query": map[string]interface{}{
								"type":        "string",
								"description": "The search query",
							},
							"max_results": map[string]interface{}{
								"type":        "number",
								"description": "Maximum number of results to return (default: 5)",
							},
						},
						Required: []string{"collection_name", "query"},
					},
				},
				Handler: SearchHandler,
			},
			toolset.ServerTool{
				Tool: mcp.Tool{
					Name:        "add_document",
					Description: "Add a document to a LocalRecall collection",
					InputSchema: mcp.ToolInputSchema{
						Type: "object",
						Properties: map[string]interface{}{
							"collection_name": map[string]interface{}{
								"type":        "string",
								"description": "The name of the collection",
							},
							"filename": map[string]interface{}{
								"type":        "string",
								"description": "The filename for the document",
							},
							"file_path": map[string]interface{}{
								"type":        "string",
								"description": "Path to the file to upload (mutually exclusive with file_content)",
							},
							"file_content": map[string]interface{}{
								"type":        "string",
								"description": "File content as string (mutually exclusive with file_path)",
							},
						},
						Required: []string{"collection_name", "filename"},
					},
				},
				Handler: AddDocumentHandler,
			},
			toolset.ServerTool{
				Tool: mcp.Tool{
					Name:        "list_files",
					Description: "List files in a LocalRecall collection",
					InputSchema: mcp.ToolInputSchema{
						Type: "object",
						Properties: map[string]interface{}{
							"collection_name": map[string]interface{}{
								"type":        "string",
								"description": "The name of the collection",
							},
						},
						Required: []string{"collection_name"},
					},
				},
				Handler: ListFilesHandler,
			},
			toolset.ServerTool{
				Tool: mcp.Tool{
					Name:        "delete_entry",
					Description: "Delete an entry from a LocalRecall collection",
					InputSchema: mcp.ToolInputSchema{
						Type: "object",
						Properties: map[string]interface{}{
							"collection_name": map[string]interface{}{
								"type":        "string",
								"description": "The name of the collection",
							},
							"entry": map[string]interface{}{
								"type":        "string",
								"description": "The filename of the entry to delete",
							},
						},
						Required: []string{"collection_name", "entry"},
					},
				},
				Handler: DeleteEntryHandler,
			},
			toolset.ServerTool{
				Tool: mcp.Tool{
					Name:        "get_entry_content",
					Description: "Get the content of a specific entry in a LocalRecall collection",
					InputSchema: mcp.ToolInputSchema{
						Type: "object",
						Properties: map[string]interface{}{
							"collection_name": map[string]interface{}{
								"type":        "string",
								"description": "The name of the collection",
							},
							"entry": map[string]interface{}{
								"type":        "string",
								"description": "The filename of the entry to retrieve",
							},
						},
						Required: []string{"collection_name", "entry"},
					},
				},
				Handler: GetEntryContentHandler,
			},
			toolset.ServerTool{
				Tool: mcp.Tool{
					Name:        "register_source",
					Description: "Register an external source for a LocalRecall collection",
					InputSchema: mcp.ToolInputSchema{
						Type: "object",
						Properties: map[string]interface{}{
							"collection_name": map[string]interface{}{
								"type":        "string",
								"description": "The name of the collection",
							},
							"url": map[string]interface{}{
								"type":        "string",
								"description": "The URL of the external source",
							},
							"update_interval": map[string]interface{}{
								"type":        "number",
								"description": "Update interval in seconds (0 or omit for no auto-update)",
							},
						},
						Required: []string{"collection_name", "url"},
					},
				},
				Handler: RegisterSourceHandler,
			},
			toolset.ServerTool{
				Tool: mcp.Tool{
					Name:        "remove_source",
					Description: "Remove an external source from a LocalRecall collection",
					InputSchema: mcp.ToolInputSchema{
						Type: "object",
						Properties: map[string]interface{}{
							"collection_name": map[string]interface{}{
								"type":        "string",
								"description": "The name of the collection",
							},
							"url": map[string]interface{}{
								"type":        "string",
								"description": "The URL of the external source to remove",
							},
						},
						Required: []string{"collection_name", "url"},
					},
				},
				Handler: RemoveSourceHandler,
			},
			toolset.ServerTool{
				Tool: mcp.Tool{
					Name:        "list_sources",
					Description: "List external sources for a LocalRecall collection",
					InputSchema: mcp.ToolInputSchema{
						Type: "object",
						Properties: map[string]interface{}{
							"collection_name": map[string]interface{}{
								"type":        "string",
								"description": "The name of the collection",
							},
						},
						Required: []string{"collection_name"},
					},
				},
				Handler: ListSourcesHandler,
			},
		)
	}

	// These tools are always the same regardless of default collection
	tools = append(tools,
		toolset.ServerTool{
			Tool: mcp.Tool{
				Name:        "create_collection",
				Description: "Create a new collection in LocalRecall",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "The name of the collection to create",
						},
					},
					Required: []string{"name"},
				},
			},
			Handler: CreateCollectionHandler,
		},
		toolset.ServerTool{
			Tool: mcp.Tool{
				Name:        "reset_collection",
				Description: "Reset (clear) a collection in LocalRecall",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "The name of the collection to reset",
						},
					},
					Required: []string{"name"},
				},
			},
			Handler: ResetCollectionHandler,
		},
		toolset.ServerTool{
			Tool: mcp.Tool{
				Name:        "list_collections",
				Description: "List all collections in LocalRecall",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
			Handler: ListCollectionsHandler,
		},
	)

	return tools
}
