package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// Client represents a LocalRecall API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// APIResponse represents the standard LocalRecall API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents an error in the API response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SearchResult represents a search result
type SearchResult struct {
	Query      string                   `json:"query"`
	MaxResults int                      `json:"max_results"`
	Results    []map[string]interface{} `json:"results"`
	Count      int                      `json:"count"`
}

// CollectionInfo represents collection information
type CollectionInfo struct {
	Name      string `json:"name"`
	CreatedAt string `json:"created_at,omitempty"`
	ResetAt   string `json:"reset_at,omitempty"`
}

// DocumentInfo represents document upload information
type DocumentInfo struct {
	Filename   string `json:"filename"`
	Collection string `json:"collection"`
	UploadedAt string `json:"uploaded_at,omitempty"`
}

// CollectionsList represents a list of collections
type CollectionsList struct {
	Collections []string `json:"collections"`
	Count       int      `json:"count"`
}

// FilesList represents a list of files in a collection
type FilesList struct {
	Collection string   `json:"collection"`
	Entries    []string `json:"entries"`
	Count      int      `json:"count"`
}

// DeleteResult represents the result of a delete operation
type DeleteResult struct {
	DeletedEntry     string   `json:"deleted_entry"`
	RemainingEntries []string `json:"remaining_entries"`
	EntryCount       int      `json:"entry_count"`
}

// NewClient creates a new LocalRecall API client
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// makeRequest makes an HTTP request to the LocalRecall API
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body interface{}) (*APIResponse, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Add debug logging for response (helpful for troubleshooting)
	if len(respBody) > 1000 {
		fmt.Printf("DEBUG: Response body (truncated): %s...\n", string(respBody[:1000]))
	} else {
		fmt.Printf("DEBUG: Response body: %s\n", string(respBody))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		// Provide more context in error message
		return nil, fmt.Errorf("failed to parse response (status: %d, body length: %d): %w",
			resp.StatusCode, len(respBody), err)
	}

	// Check if the response is valid
	if !apiResp.Success && apiResp.Error == nil && apiResp.Data == nil {
		return nil, fmt.Errorf("invalid API response: missing success, error, and data fields")
	}

	if !apiResp.Success {
		errorMsg := "unknown error"
		if apiResp.Error != nil {
			errorMsg = fmt.Sprintf("%s: %s", apiResp.Error.Code, apiResp.Error.Message)
			if apiResp.Error.Details != "" {
				errorMsg += " - " + apiResp.Error.Details
			}
		}
		return nil, fmt.Errorf("API error: %s", errorMsg)
	}

	return &apiResp, nil
}

// makeMultipartRequest makes a multipart form request for file uploads
func (c *Client) makeMultipartRequest(ctx context.Context, endpoint, filename string, fileContent []byte) (*APIResponse, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(fileContent); err != nil {
		return nil, fmt.Errorf("failed to write file content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+endpoint, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Add debug logging for response (helpful for troubleshooting)
	if len(respBody) > 1000 {
		fmt.Printf("DEBUG: Response body (truncated): %s...\n", string(respBody[:1000]))
	} else {
		fmt.Printf("DEBUG: Response body: %s\n", string(respBody))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		// Provide more context in error message
		return nil, fmt.Errorf("failed to parse response (status: %d, body length: %d): %w",
			resp.StatusCode, len(respBody), err)
	}

	// Check if the response is valid
	if !apiResp.Success && apiResp.Error == nil && apiResp.Data == nil {
		return nil, fmt.Errorf("invalid API response: missing success, error, and data fields")
	}

	if !apiResp.Success {
		errorMsg := "unknown error"
		if apiResp.Error != nil {
			errorMsg = fmt.Sprintf("%s: %s", apiResp.Error.Code, apiResp.Error.Message)
			if apiResp.Error.Details != "" {
				errorMsg += " - " + apiResp.Error.Details
			}
		}
		return nil, fmt.Errorf("API error: %s", errorMsg)
	}

	return &apiResp, nil
}

// Search searches content in a LocalRecall collection
func (c *Client) Search(ctx context.Context, collectionName, query string, maxResults int) (*SearchResult, error) {
	if maxResults == 0 {
		maxResults = 5
	}

	requestBody := map[string]interface{}{
		"query":       query,
		"max_results": maxResults,
	}

	apiResp, err := c.makeRequest(ctx, "POST", fmt.Sprintf("/api/collections/%s/search", collectionName), requestBody)
	if err != nil {
		return nil, err
	}

	// Extract data from response
	data, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response data format")
	}

	results := []map[string]interface{}{}
	if resultsData, ok := data["results"].([]interface{}); ok {
		for _, r := range resultsData {
			if resultMap, ok := r.(map[string]interface{}); ok {
				results = append(results, resultMap)
			}
		}
	}

	count := 0
	if countVal, ok := data["count"].(float64); ok {
		count = int(countVal)
	}

	return &SearchResult{
		Query:      query,
		MaxResults: maxResults,
		Results:    results,
		Count:      count,
	}, nil
}

// CreateCollection creates a new collection
func (c *Client) CreateCollection(ctx context.Context, name string) (*CollectionInfo, error) {
	requestBody := map[string]interface{}{
		"name": name,
	}

	apiResp, err := c.makeRequest(ctx, "POST", "/api/collections", requestBody)
	if err != nil {
		return nil, err
	}

	// Extract data from response
	data, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response data format")
	}

	createdAt := ""
	if createdAtVal, ok := data["created_at"].(string); ok {
		createdAt = createdAtVal
	}

	return &CollectionInfo{
		Name:      name,
		CreatedAt: createdAt,
	}, nil
}

// ResetCollection resets (clears) a collection
func (c *Client) ResetCollection(ctx context.Context, name string) (*CollectionInfo, error) {
	apiResp, err := c.makeRequest(ctx, "POST", fmt.Sprintf("/api/collections/%s/reset", name), nil)
	if err != nil {
		return nil, err
	}

	// Extract data from response
	data, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response data format")
	}

	resetAt := ""
	if resetAtVal, ok := data["reset_at"].(string); ok {
		resetAt = resetAtVal
	}

	return &CollectionInfo{
		Name:    name,
		ResetAt: resetAt,
	}, nil
}

// AddDocument adds a document to a collection
func (c *Client) AddDocument(ctx context.Context, collectionName, filename string, fileContent []byte) (*DocumentInfo, error) {
	apiResp, err := c.makeMultipartRequest(ctx, fmt.Sprintf("/api/collections/%s/upload", collectionName), filename, fileContent)
	if err != nil {
		return nil, err
	}

	// Extract data from response
	data, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response data format")
	}

	uploadedAt := ""
	if uploadedAtVal, ok := data["uploaded_at"].(string); ok {
		uploadedAt = uploadedAtVal
	}

	return &DocumentInfo{
		Filename:   filename,
		Collection: collectionName,
		UploadedAt: uploadedAt,
	}, nil
}

// ListCollections lists all collections
func (c *Client) ListCollections(ctx context.Context) (*CollectionsList, error) {
	apiResp, err := c.makeRequest(ctx, "GET", "/api/collections", nil)
	if err != nil {
		return nil, err
	}

	// Extract data from response
	data, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response data format")
	}

	collections := []string{}
	if collectionsData, ok := data["collections"].([]interface{}); ok {
		for _, c := range collectionsData {
			if col, ok := c.(string); ok {
				collections = append(collections, col)
			}
		}
	}

	count := 0
	if countVal, ok := data["count"].(float64); ok {
		count = int(countVal)
	}

	return &CollectionsList{
		Collections: collections,
		Count:       count,
	}, nil
}

// ListFiles lists files in a collection
func (c *Client) ListFiles(ctx context.Context, collectionName string) (*FilesList, error) {
	apiResp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/api/collections/%s/entries", collectionName), nil)
	if err != nil {
		return nil, err
	}

	// Extract data from response
	data, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response data format")
	}

	entries := []string{}
	if entriesData, ok := data["entries"].([]interface{}); ok {
		for _, e := range entriesData {
			if entry, ok := e.(string); ok {
				entries = append(entries, entry)
			}
		}
	}

	count := 0
	if countVal, ok := data["count"].(float64); ok {
		count = int(countVal)
	}

	return &FilesList{
		Collection: collectionName,
		Entries:    entries,
		Count:      count,
	}, nil
}

// DeleteEntry deletes an entry from a collection
func (c *Client) DeleteEntry(ctx context.Context, collectionName, entry string) (*DeleteResult, error) {
	requestBody := map[string]interface{}{
		"entry": entry,
	}

	apiResp, err := c.makeRequest(ctx, "DELETE", fmt.Sprintf("/api/collections/%s/entry/delete", collectionName), requestBody)
	if err != nil {
		return nil, err
	}

	// Extract data from response
	data, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response data format")
	}

	remainingEntries := []string{}
	entryCount := 0

	if remainingData, ok := data["remaining_entries"].([]interface{}); ok {
		for _, e := range remainingData {
			if entryStr, ok := e.(string); ok {
				remainingEntries = append(remainingEntries, entryStr)
			}
		}
		entryCount = len(remainingEntries)
	}

	if countVal, ok := data["entry_count"].(float64); ok {
		entryCount = int(countVal)
	}

	return &DeleteResult{
		DeletedEntry:     entry,
		RemainingEntries: remainingEntries,
		EntryCount:       entryCount,
	}, nil
}
