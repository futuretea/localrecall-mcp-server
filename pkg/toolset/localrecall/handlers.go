package localrecall

import (
	"context"
	"fmt"
	"os"

	"github.com/futuretea/localrecall-mcp-server/pkg/toolset"
	"github.com/futuretea/localrecall-mcp-server/pkg/toolset/handler"
)

// SearchHandler handles search requests
func SearchHandler(clientInterface interface{}, params map[string]interface{}) (string, error) {
	client, ok := clientInterface.(*toolset.LocalRecallClient)
	if !ok {
		return "", fmt.Errorf("invalid client type")
	}

	// Extract parameters
	query := handler.GetStringParam(params, "query", "")
	if query == "" {
		return "", fmt.Errorf("query parameter is required")
	}

	collectionName := handler.GetStringParam(params, "collection_name", "")
	maxResults := handler.GetIntParam(params, "max_results", 5)
	format := handler.GetStringParam(params, "format", "json")

	// Call API
	ctx := context.Background()
	result, err := client.Client.Search(ctx, collectionName, query, maxResults)
	if err != nil {
		return "", fmt.Errorf("search failed: %w", err)
	}

	// Format output
	return handler.FormatOutput(result, format)
}

// CreateCollectionHandler handles create collection requests
func CreateCollectionHandler(clientInterface interface{}, params map[string]interface{}) (string, error) {
	client, ok := clientInterface.(*toolset.LocalRecallClient)
	if !ok {
		return "", fmt.Errorf("invalid client type")
	}

	// Extract parameters
	name, err := handler.RequireStringParam(params, "name")
	if err != nil {
		return "", err
	}

	format := handler.GetStringParam(params, "format", "json")

	// Call API
	ctx := context.Background()
	result, err := client.Client.CreateCollection(ctx, name)
	if err != nil {
		return "", fmt.Errorf("create collection failed: %w", err)
	}

	// Format output
	return handler.FormatOutput(result, format)
}

// ResetCollectionHandler handles reset collection requests
func ResetCollectionHandler(clientInterface interface{}, params map[string]interface{}) (string, error) {
	client, ok := clientInterface.(*toolset.LocalRecallClient)
	if !ok {
		return "", fmt.Errorf("invalid client type")
	}

	// Extract parameters
	name, err := handler.RequireStringParam(params, "name")
	if err != nil {
		return "", err
	}

	format := handler.GetStringParam(params, "format", "json")

	// Call API
	ctx := context.Background()
	result, err := client.Client.ResetCollection(ctx, name)
	if err != nil {
		return "", fmt.Errorf("reset collection failed: %w", err)
	}

	// Format output
	return handler.FormatOutput(result, format)
}

// AddDocumentHandler handles add document requests
func AddDocumentHandler(clientInterface interface{}, params map[string]interface{}) (string, error) {
	client, ok := clientInterface.(*toolset.LocalRecallClient)
	if !ok {
		return "", fmt.Errorf("invalid client type")
	}

	// Extract parameters
	collectionName := handler.GetStringParam(params, "collection_name", "")

	filename, err := handler.RequireStringParam(params, "filename")
	if err != nil {
		return "", err
	}

	filePath := handler.GetStringParam(params, "file_path", "")
	fileContent := handler.GetStringParam(params, "file_content", "")
	format := handler.GetStringParam(params, "format", "json")

	// Validate that either file_path or file_content is provided
	if filePath == "" && fileContent == "" {
		return "", fmt.Errorf("either file_path or file_content must be provided")
	}

	if filePath != "" && fileContent != "" {
		return "", fmt.Errorf("cannot specify both file_path and file_content")
	}

	// Read file content if file_path is provided
	var fileBytes []byte
	if filePath != "" {
		fileBytes, err = os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
	} else {
		fileBytes = []byte(fileContent)
	}

	// Call API
	ctx := context.Background()
	result, err := client.Client.AddDocument(ctx, collectionName, filename, fileBytes)
	if err != nil {
		return "", fmt.Errorf("add document failed: %w", err)
	}

	// Format output
	return handler.FormatOutput(result, format)
}

// ListCollectionsHandler handles list collections requests
func ListCollectionsHandler(clientInterface interface{}, params map[string]interface{}) (string, error) {
	client, ok := clientInterface.(*toolset.LocalRecallClient)
	if !ok {
		return "", fmt.Errorf("invalid client type")
	}

	format := handler.GetStringParam(params, "format", "json")

	// Call API
	ctx := context.Background()
	result, err := client.Client.ListCollections(ctx)
	if err != nil {
		return "", fmt.Errorf("list collections failed: %w", err)
	}

	// Format output
	return handler.FormatOutput(result, format)
}

// ListFilesHandler handles list files requests
func ListFilesHandler(clientInterface interface{}, params map[string]interface{}) (string, error) {
	client, ok := clientInterface.(*toolset.LocalRecallClient)
	if !ok {
		return "", fmt.Errorf("invalid client type")
	}

	// Extract parameters
	collectionName := handler.GetStringParam(params, "collection_name", "")
	format := handler.GetStringParam(params, "format", "json")

	// Call API
	ctx := context.Background()
	result, err := client.Client.ListFiles(ctx, collectionName)
	if err != nil {
		return "", fmt.Errorf("list files failed: %w", err)
	}

	// Format output
	return handler.FormatOutput(result, format)
}

// DeleteEntryHandler handles delete entry requests
func DeleteEntryHandler(clientInterface interface{}, params map[string]interface{}) (string, error) {
	client, ok := clientInterface.(*toolset.LocalRecallClient)
	if !ok {
		return "", fmt.Errorf("invalid client type")
	}

	// Extract parameters
	collectionName := handler.GetStringParam(params, "collection_name", "")

	entry, err := handler.RequireStringParam(params, "entry")
	if err != nil {
		return "", err
	}

	format := handler.GetStringParam(params, "format", "json")

	// Call API
	ctx := context.Background()
	result, err := client.Client.DeleteEntry(ctx, collectionName, entry)
	if err != nil {
		return "", fmt.Errorf("delete entry failed: %w", err)
	}

	// Format output
	return handler.FormatOutput(result, format)
}
