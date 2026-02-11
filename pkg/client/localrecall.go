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

// parseAPIResponse parses and validates APIResponse from response body
func parseAPIResponse(respBody []byte, statusCode int) (*APIResponse, error) {
	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response (status: %d, body length: %d): %w",
			statusCode, len(respBody), err)
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

// getDataMap extracts data as map[string]interface{} from APIResponse
func getDataMap(resp *APIResponse) (map[string]interface{}, error) {
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response data format")
	}
	return data, nil
}

// getStringField extracts a string field from data map
func getStringField(data map[string]interface{}, field string) string {
	if val, ok := data[field].(string); ok {
		return val
	}
	return ""
}

// getStringArray extracts a string array from data map
func getStringArray(data map[string]interface{}, field string) []string {
	result := []string{}
	if arr, ok := data[field].([]interface{}); ok {
		for _, item := range arr {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
	}
	return result
}

// getIntField extracts an int field from data map
func getIntField(data map[string]interface{}, field string) int {
	if val, ok := data[field].(float64); ok {
		return int(val)
	}
	return 0
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

	// Debug logging can be enabled via environment variable DEBUG_API_RESPONSES=true
	// Note: Avoid using fmt.Printf in production code
	// if os.Getenv("DEBUG_API_RESPONSES") == "true" {
	// 	log.Printf("API Response (status %d, length %d): %s", resp.StatusCode, len(respBody), string(respBody))
	// }

	return parseAPIResponse(respBody, resp.StatusCode)
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

	// Debug logging can be enabled via environment variable DEBUG_API_RESPONSES=true
	// Note: Avoid using fmt.Printf in production code
	// if os.Getenv("DEBUG_API_RESPONSES") == "true" {
	// 	log.Printf("API Response (status %d, length %d): %s", resp.StatusCode, len(respBody), string(respBody))
	// }

	return parseAPIResponse(respBody, resp.StatusCode)
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

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+fmt.Sprintf("/api/collections/%s/search", collectionName), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed with status: %d", resp.StatusCode)
	}

	// LocalRecall returns raw array of results directly (not wrapped in APIResponse)
	// This matches the behavior of LocalRecall's official Go client
	var rawResults []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawResults); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &SearchResult{
		Query:      query,
		MaxResults: maxResults,
		Results:    rawResults,
		Count:      len(rawResults),
	}, nil
}

// CreateCollection creates a new collection
func (c *Client) CreateCollection(ctx context.Context, name string) (*CollectionInfo, error) {
	resp, err := c.makeRequest(ctx, "POST", "/api/collections", map[string]interface{}{"name": name})
	if err != nil {
		return nil, err
	}

	data, err := getDataMap(resp)
	if err != nil {
		return nil, err
	}

	return &CollectionInfo{
		Name:      name,
		CreatedAt: getStringField(data, "created_at"),
	}, nil
}

// ResetCollection resets (clears) a collection
func (c *Client) ResetCollection(ctx context.Context, name string) (*CollectionInfo, error) {
	resp, err := c.makeRequest(ctx, "POST", fmt.Sprintf("/api/collections/%s/reset", name), nil)
	if err != nil {
		return nil, err
	}

	data, err := getDataMap(resp)
	if err != nil {
		return nil, err
	}

	return &CollectionInfo{
		Name:    name,
		ResetAt: getStringField(data, "reset_at"),
	}, nil
}

// AddDocument adds a document to a collection
func (c *Client) AddDocument(ctx context.Context, collectionName, filename string, fileContent []byte) (*DocumentInfo, error) {
	resp, err := c.makeMultipartRequest(ctx, fmt.Sprintf("/api/collections/%s/upload", collectionName), filename, fileContent)
	if err != nil {
		return nil, err
	}

	data, err := getDataMap(resp)
	if err != nil {
		return nil, err
	}

	return &DocumentInfo{
		Filename:   filename,
		Collection: collectionName,
		UploadedAt: getStringField(data, "uploaded_at"),
	}, nil
}

// ListCollections lists all collections
func (c *Client) ListCollections(ctx context.Context) (*CollectionsList, error) {
	resp, err := c.makeRequest(ctx, "GET", "/api/collections", nil)
	if err != nil {
		return nil, err
	}

	data, err := getDataMap(resp)
	if err != nil {
		return nil, err
	}

	return &CollectionsList{
		Collections: getStringArray(data, "collections"),
		Count:       getIntField(data, "count"),
	}, nil
}

// ListFiles lists files in a collection
func (c *Client) ListFiles(ctx context.Context, collectionName string) (*FilesList, error) {
	resp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/api/collections/%s/entries", collectionName), nil)
	if err != nil {
		return nil, err
	}

	data, err := getDataMap(resp)
	if err != nil {
		return nil, err
	}

	return &FilesList{
		Collection: collectionName,
		Entries:    getStringArray(data, "entries"),
		Count:      getIntField(data, "count"),
	}, nil
}

// DeleteEntry deletes an entry from a collection
func (c *Client) DeleteEntry(ctx context.Context, collectionName, entry string) (*DeleteResult, error) {
	resp, err := c.makeRequest(ctx, "DELETE", fmt.Sprintf("/api/collections/%s/entry/delete", collectionName), map[string]interface{}{"entry": entry})
	if err != nil {
		return nil, err
	}

	data, err := getDataMap(resp)
	if err != nil {
		return nil, err
	}

	return &DeleteResult{
		DeletedEntry:     entry,
		RemainingEntries: getStringArray(data, "remaining_entries"),
		EntryCount:       getIntField(data, "entry_count"),
	}, nil
}
