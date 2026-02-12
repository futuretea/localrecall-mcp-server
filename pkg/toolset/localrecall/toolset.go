package localrecall

import (
	"maps"

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

// prop creates a JSON schema property definition.
func prop(typ, desc string) map[string]interface{} {
	return map[string]interface{}{
		"type":        typ,
		"description": desc,
	}
}

// collectionToolDef describes a tool that operates on a specific collection.
// The collection_name parameter is automatically added based on DefaultCollection.
type collectionToolDef struct {
	name        string
	descDefault string // description prefix when default collection is set (collection name appended)
	descGeneric string // description when no default collection
	handler     toolset.ToolHandler
	props       map[string]interface{} // properties excluding collection_name
	required    []string               // required params excluding collection_name
}

// buildCollectionTool creates a ServerTool from a collectionToolDef,
// handling default/non-default collection_name automatically.
func (t *Toolset) buildCollectionTool(def collectionToolDef) toolset.ServerTool {
	props := make(map[string]interface{})
	maps.Copy(props, def.props)

	var desc string
	required := make([]string, len(def.required))
	copy(required, def.required)

	if t.DefaultCollection != "" {
		desc = def.descDefault + " '" + t.DefaultCollection + "'"
		// Enforced isolation: do NOT expose collection_name parameter
	} else {
		desc = def.descGeneric
		props["collection_name"] = prop("string", "The name of the collection")
		required = append([]string{"collection_name"}, required...)
	}

	return toolset.ServerTool{
		Tool: mcp.Tool{
			Name:        def.name,
			Description: desc,
			InputSchema: mcp.ToolInputSchema{
				Type:       "object",
				Properties: props,
				Required:   required,
			},
		},
		Handler: def.handler,
	}
}

// GetTools returns all LocalRecall tools
func (t *Toolset) GetTools(client interface{}) []toolset.ServerTool {
	// Collection-scoped tools: collection_name is optional (with default) or required (without)
	collectionTools := []collectionToolDef{
		{
			name:        "search",
			descDefault: "Search content in LocalRecall collection",
			descGeneric: "Search content in a LocalRecall collection",
			handler:     SearchHandler,
			props: map[string]interface{}{
				"query":          prop("string", "The search query"),
				"max_results":    prop("number", "Maximum number of results to return (default: 5)"),
				"min_similarity": prop("number", "Minimum cosine similarity threshold (0-1). Results below this score are filtered out. 0 or omit to disable."),
				"filters": map[string]interface{}{
					"type":        "object",
					"description": "Metadata key-value filters. Only results whose metadata contains all specified key-value pairs are returned.",
					"additionalProperties": map[string]interface{}{
						"type": "string",
					},
				},
			},
			required: []string{"query"},
		},
		{
			name:        "add_document",
			descDefault: "Add a document to LocalRecall collection",
			descGeneric: "Add a document to a LocalRecall collection",
			handler:     AddDocumentHandler,
			props: map[string]interface{}{
				"filename":     prop("string", "The filename for the document"),
				"file_path":    prop("string", "Path to the file to upload (mutually exclusive with file_content)"),
				"file_content": prop("string", "File content as string (mutually exclusive with file_path)"),
			},
			required: []string{"filename"},
		},
		{
			name:        "list_files",
			descDefault: "List files in LocalRecall collection",
			descGeneric: "List files in a LocalRecall collection",
			handler:     ListFilesHandler,
		},
		{
			name:        "delete_entry",
			descDefault: "Delete an entry from LocalRecall collection",
			descGeneric: "Delete an entry from a LocalRecall collection",
			handler:     DeleteEntryHandler,
			props: map[string]interface{}{
				"entry": prop("string", "The filename of the entry to delete"),
			},
			required: []string{"entry"},
		},
		{
			name:        "get_entry_content",
			descDefault: "Get the content of a specific entry in LocalRecall collection",
			descGeneric: "Get the content of a specific entry in a LocalRecall collection",
			handler:     GetEntryContentHandler,
			props: map[string]interface{}{
				"entry": prop("string", "The filename of the entry to retrieve"),
			},
			required: []string{"entry"},
		},
		{
			name:        "register_source",
			descDefault: "Register an external source for LocalRecall collection",
			descGeneric: "Register an external source for a LocalRecall collection",
			handler:     RegisterSourceHandler,
			props: map[string]interface{}{
				"url":             prop("string", "The URL of the external source"),
				"update_interval": prop("number", "Update interval in seconds (0 or omit for no auto-update)"),
			},
			required: []string{"url"},
		},
		{
			name:        "remove_source",
			descDefault: "Remove an external source from LocalRecall collection",
			descGeneric: "Remove an external source from a LocalRecall collection",
			handler:     RemoveSourceHandler,
			props: map[string]interface{}{
				"url": prop("string", "The URL of the external source to remove"),
			},
			required: []string{"url"},
		},
		{
			name:        "list_sources",
			descDefault: "List external sources for LocalRecall collection",
			descGeneric: "List external sources for a LocalRecall collection",
			handler:     ListSourcesHandler,
		},
		{
			name:        "reindex",
			descDefault: "Re-chunk and re-index all documents in LocalRecall collection",
			descGeneric: "Re-chunk and re-index all documents in a LocalRecall collection using the current chunking strategy",
			handler:     ReindexHandler,
		},
	}

	tools := make([]toolset.ServerTool, 0, len(collectionTools))
	for _, def := range collectionTools {
		tools = append(tools, t.buildCollectionTool(def))
	}

	// Collection-independent tools: only available when no collection isolation is configured
	if t.DefaultCollection == "" {
		tools = append(tools,
			toolset.ServerTool{
				Tool: mcp.Tool{
					Name:        "create_collection",
					Description: "Create a new collection in LocalRecall",
					InputSchema: mcp.ToolInputSchema{
						Type: "object",
						Properties: map[string]interface{}{
							"name": prop("string", "The name of the collection to create"),
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
							"name": prop("string", "The name of the collection to reset"),
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
	}

	return tools
}
