package localrecall

import (
	"context"
	"fmt"
	"os"

	lrclient "github.com/futuretea/localrecall-mcp-server/pkg/client"
	"github.com/futuretea/localrecall-mcp-server/pkg/toolset"
	"github.com/futuretea/localrecall-mcp-server/pkg/toolset/handler"
)

func getClient(clientInterface interface{}) (*toolset.LocalRecallClient, error) {
	c, ok := clientInterface.(*toolset.LocalRecallClient)
	if !ok {
		return nil, fmt.Errorf("invalid client type")
	}
	return c, nil
}

// SearchHandler handles search requests
func SearchHandler(clientInterface interface{}, params map[string]interface{}) (string, error) {
	client, err := getClient(clientInterface)
	if err != nil {
		return "", err
	}

	query := handler.GetStringParam(params, "query", "")
	if query == "" {
		return "", fmt.Errorf("query parameter is required")
	}

	collectionName := handler.GetStringParam(params, "collection_name", "")
	maxResults := handler.GetIntParam(params, "max_results", 5)
	format := handler.GetStringParam(params, "format", "json")

	var opts *lrclient.SearchOptions
	minSim := handler.GetFloat64Param(params, "min_similarity", 0)
	filters := handler.GetStringMapParam(params, "filters")
	if minSim > 0 || len(filters) > 0 {
		opts = &lrclient.SearchOptions{
			MinSimilarity: minSim,
			Filters:       filters,
		}
	}

	result, err := client.Client.SearchWithOptions(context.Background(), collectionName, query, maxResults, opts)
	if err != nil {
		return "", fmt.Errorf("search failed: %w", err)
	}

	return handler.FormatOutput(result, format)
}

// CreateCollectionHandler handles create collection requests
func CreateCollectionHandler(clientInterface interface{}, params map[string]interface{}) (string, error) {
	client, err := getClient(clientInterface)
	if err != nil {
		return "", err
	}

	name, err := handler.RequireStringParam(params, "name")
	if err != nil {
		return "", err
	}

	format := handler.GetStringParam(params, "format", "json")

	result, err := client.Client.CreateCollection(context.Background(), name)
	if err != nil {
		return "", fmt.Errorf("create collection failed: %w", err)
	}

	return handler.FormatOutput(result, format)
}

// ResetCollectionHandler handles reset collection requests
func ResetCollectionHandler(clientInterface interface{}, params map[string]interface{}) (string, error) {
	client, err := getClient(clientInterface)
	if err != nil {
		return "", err
	}

	name, err := handler.RequireStringParam(params, "name")
	if err != nil {
		return "", err
	}

	format := handler.GetStringParam(params, "format", "json")

	result, err := client.Client.ResetCollection(context.Background(), name)
	if err != nil {
		return "", fmt.Errorf("reset collection failed: %w", err)
	}

	return handler.FormatOutput(result, format)
}

// AddDocumentHandler handles add document requests
func AddDocumentHandler(clientInterface interface{}, params map[string]interface{}) (string, error) {
	client, err := getClient(clientInterface)
	if err != nil {
		return "", err
	}

	collectionName := handler.GetStringParam(params, "collection_name", "")

	filename, err := handler.RequireStringParam(params, "filename")
	if err != nil {
		return "", err
	}

	filePath := handler.GetStringParam(params, "file_path", "")
	fileContent := handler.GetStringParam(params, "file_content", "")
	format := handler.GetStringParam(params, "format", "json")

	if filePath == "" && fileContent == "" {
		return "", fmt.Errorf("either file_path or file_content must be provided")
	}
	if filePath != "" && fileContent != "" {
		return "", fmt.Errorf("cannot specify both file_path and file_content")
	}

	var fileBytes []byte
	if filePath != "" {
		fileBytes, err = os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
	} else {
		fileBytes = []byte(fileContent)
	}

	result, err := client.Client.AddDocument(context.Background(), collectionName, filename, fileBytes)
	if err != nil {
		return "", fmt.Errorf("add document failed: %w", err)
	}

	return handler.FormatOutput(result, format)
}

// ListCollectionsHandler handles list collections requests
func ListCollectionsHandler(clientInterface interface{}, params map[string]interface{}) (string, error) {
	client, err := getClient(clientInterface)
	if err != nil {
		return "", err
	}

	format := handler.GetStringParam(params, "format", "json")

	result, err := client.Client.ListCollections(context.Background())
	if err != nil {
		return "", fmt.Errorf("list collections failed: %w", err)
	}

	return handler.FormatOutput(result, format)
}

// ListFilesHandler handles list files requests
func ListFilesHandler(clientInterface interface{}, params map[string]interface{}) (string, error) {
	client, err := getClient(clientInterface)
	if err != nil {
		return "", err
	}

	collectionName := handler.GetStringParam(params, "collection_name", "")
	format := handler.GetStringParam(params, "format", "json")

	result, err := client.Client.ListFiles(context.Background(), collectionName)
	if err != nil {
		return "", fmt.Errorf("list files failed: %w", err)
	}

	return handler.FormatOutput(result, format)
}

// DeleteEntryHandler handles delete entry requests
func DeleteEntryHandler(clientInterface interface{}, params map[string]interface{}) (string, error) {
	client, err := getClient(clientInterface)
	if err != nil {
		return "", err
	}

	collectionName := handler.GetStringParam(params, "collection_name", "")

	entry, err := handler.RequireStringParam(params, "entry")
	if err != nil {
		return "", err
	}

	format := handler.GetStringParam(params, "format", "json")

	result, err := client.Client.DeleteEntry(context.Background(), collectionName, entry)
	if err != nil {
		return "", fmt.Errorf("delete entry failed: %w", err)
	}

	return handler.FormatOutput(result, format)
}

// GetEntryContentHandler handles get entry content requests
func GetEntryContentHandler(clientInterface interface{}, params map[string]interface{}) (string, error) {
	client, err := getClient(clientInterface)
	if err != nil {
		return "", err
	}

	collectionName := handler.GetStringParam(params, "collection_name", "")

	entry, err := handler.RequireStringParam(params, "entry")
	if err != nil {
		return "", err
	}

	format := handler.GetStringParam(params, "format", "json")

	result, err := client.Client.GetEntryContent(context.Background(), collectionName, entry)
	if err != nil {
		return "", fmt.Errorf("get entry content failed: %w", err)
	}

	return handler.FormatOutput(result, format)
}

// RegisterSourceHandler handles register external source requests
func RegisterSourceHandler(clientInterface interface{}, params map[string]interface{}) (string, error) {
	client, err := getClient(clientInterface)
	if err != nil {
		return "", err
	}

	collectionName := handler.GetStringParam(params, "collection_name", "")

	sourceURL, err := handler.RequireStringParam(params, "url")
	if err != nil {
		return "", err
	}

	updateInterval := handler.GetIntParam(params, "update_interval", 0)
	format := handler.GetStringParam(params, "format", "json")

	result, err := client.Client.RegisterSource(context.Background(), collectionName, sourceURL, updateInterval)
	if err != nil {
		return "", fmt.Errorf("register source failed: %w", err)
	}

	return handler.FormatOutput(result, format)
}

// RemoveSourceHandler handles remove external source requests
func RemoveSourceHandler(clientInterface interface{}, params map[string]interface{}) (string, error) {
	client, err := getClient(clientInterface)
	if err != nil {
		return "", err
	}

	collectionName := handler.GetStringParam(params, "collection_name", "")

	sourceURL, err := handler.RequireStringParam(params, "url")
	if err != nil {
		return "", err
	}

	if err := client.Client.RemoveSource(context.Background(), collectionName, sourceURL); err != nil {
		return "", fmt.Errorf("remove source failed: %w", err)
	}

	result := map[string]interface{}{
		"collection": collectionName,
		"url":        sourceURL,
		"removed":    true,
	}

	format := handler.GetStringParam(params, "format", "json")
	return handler.FormatOutput(result, format)
}

// ListSourcesHandler handles list external sources requests
func ListSourcesHandler(clientInterface interface{}, params map[string]interface{}) (string, error) {
	client, err := getClient(clientInterface)
	if err != nil {
		return "", err
	}

	collectionName := handler.GetStringParam(params, "collection_name", "")
	format := handler.GetStringParam(params, "format", "json")

	result, err := client.Client.ListSources(context.Background(), collectionName)
	if err != nil {
		return "", fmt.Errorf("list sources failed: %w", err)
	}

	return handler.FormatOutput(result, format)
}
