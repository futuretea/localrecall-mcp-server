package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	baseURL := "http://localhost:8080"
	apiKey := "test-api-key"

	client := NewClient(baseURL, apiKey)

	if client.baseURL != baseURL {
		t.Errorf("Expected baseURL %s, got %s", baseURL, client.baseURL)
	}
	if client.apiKey != apiKey {
		t.Errorf("Expected apiKey %s, got %s", apiKey, client.apiKey)
	}
	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", client.httpClient.Timeout)
	}
}

func TestSearch_Success(t *testing.T) {
	// Mock server that returns raw array (like actual LocalRecall)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/collections/test-collection/search" {
			t.Errorf("Expected path /api/collections/test-collection/search, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Expected Content-Type: application/json")
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("Expected Authorization header with Bearer token")
		}

		// Parse request body
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}
		if req["query"] != "test query" {
			t.Errorf("Expected query 'test query', got %v", req["query"])
		}
		if req["max_results"] != float64(5) {
			t.Errorf("Expected max_results 5, got %v", req["max_results"])
		}

		// Return raw array response (matching actual LocalRecall behavior)
		response := []map[string]interface{}{
			{
				"ID":         "1",
				"Content":    "Test content 1",
				"Similarity": 0.9,
				"Metadata": map[string]string{
					"source": "test.md",
				},
			},
			{
				"ID":         "2",
				"Content":    "Test content 2",
				"Similarity": 0.8,
				"Metadata": map[string]string{
					"source": "test2.md",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	result, err := client.Search(context.Background(), "test-collection", "test query", 5)

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if result.Query != "test query" {
		t.Errorf("Expected query 'test query', got %s", result.Query)
	}
	if result.MaxResults != 5 {
		t.Errorf("Expected max_results 5, got %d", result.MaxResults)
	}
	if result.Count != 2 {
		t.Errorf("Expected count 2, got %d", result.Count)
	}
	if len(result.Results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(result.Results))
	}
	if result.Results[0]["ID"] != "1" {
		t.Errorf("Expected first result ID '1', got %v", result.Results[0]["ID"])
	}
}

func TestSearch_DefaultMaxResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)

		// Verify default max_results is set to 5
		if req["max_results"] != float64(5) {
			t.Errorf("Expected default max_results 5, got %v", req["max_results"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	_, err := client.Search(context.Background(), "test", "query", 0)

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
}

func TestSearch_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Collection not found"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	_, err := client.Search(context.Background(), "nonexistent", "query", 5)

	if err == nil {
		t.Error("Expected error for 404 response")
	}
	if err.Error() != "search failed with status: 404" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestSearch_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	_, err := client.Search(context.Background(), "test", "query", 5)

	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestCreateCollection_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/collections" {
			t.Errorf("Expected path /api/collections, got %s", r.URL.Path)
		}

		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		if req["name"] != "new-collection" {
			t.Errorf("Expected name 'new-collection', got %v", req["name"])
		}

		// Return standard APIResponse
		response := APIResponse{
			Success: true,
			Message: "Collection created",
			Data: map[string]interface{}{
				"name":       "new-collection",
				"created_at": "2024-01-01T00:00:00Z",
			},
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	result, err := client.CreateCollection(context.Background(), "new-collection")

	if err != nil {
		t.Fatalf("CreateCollection failed: %v", err)
	}
	if result.Name != "new-collection" {
		t.Errorf("Expected name 'new-collection', got %s", result.Name)
	}
}

func TestCreateCollection_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "CONFLICT",
				Message: "Collection already exists",
			},
		}
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	_, err := client.CreateCollection(context.Background(), "existing")

	if err == nil {
		t.Error("Expected error for existing collection")
	}
}

func TestListCollections_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		response := APIResponse{
			Success: true,
			Data: map[string]interface{}{
				"collections": []string{"col1", "col2", "col3"},
				"count":       3,
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	result, err := client.ListCollections(context.Background())

	if err != nil {
		t.Fatalf("ListCollections failed: %v", err)
	}
	if len(result.Collections) != 3 {
		t.Errorf("Expected 3 collections, got %d", len(result.Collections))
	}
	if result.Count != 3 {
		t.Errorf("Expected count 3, got %d", result.Count)
	}
}

func TestListFiles_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/test/entries" {
			t.Errorf("Expected path /api/collections/test/entries, got %s", r.URL.Path)
		}

		response := APIResponse{
			Success: true,
			Data: map[string]interface{}{
				"collection": "test",
				"entries":    []string{"file1.txt", "file2.md"},
				"count":      2,
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	result, err := client.ListFiles(context.Background(), "test")

	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}
	if result.Collection != "test" {
		t.Errorf("Expected collection 'test', got %s", result.Collection)
	}
	if len(result.Entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(result.Entries))
	}
}

func TestAddDocument_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/collections/test/upload" {
			t.Errorf("Expected path /api/collections/test/upload, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") == "" || r.Header.Get("Content-Type")[:19] != "multipart/form-data" {
			t.Error("Expected multipart/form-data content type")
		}

		response := APIResponse{
			Success: true,
			Data: map[string]interface{}{
				"filename":   "test.txt",
				"collection": "test",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	result, err := client.AddDocument(context.Background(), "test", "test.txt", []byte("content"))

	if err != nil {
		t.Fatalf("AddDocument failed: %v", err)
	}
	if result.Filename != "test.txt" {
		t.Errorf("Expected filename 'test.txt', got %s", result.Filename)
	}
}

func TestResetCollection_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/collections/test/reset" {
			t.Errorf("Expected path /api/collections/test/reset, got %s", r.URL.Path)
		}

		response := APIResponse{
			Success: true,
			Data: map[string]interface{}{
				"collection": "test",
				"reset_at":   "2024-01-01T00:00:00Z",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	result, err := client.ResetCollection(context.Background(), "test")

	if err != nil {
		t.Fatalf("ResetCollection failed: %v", err)
	}
	if result.Name != "test" {
		t.Errorf("Expected collection 'test', got %s", result.Name)
	}
}

func TestDeleteEntry_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}

		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		if req["entry"] != "file.txt" {
			t.Errorf("Expected entry 'file.txt', got %v", req["entry"])
		}

		response := APIResponse{
			Success: true,
			Data: map[string]interface{}{
				"deleted_entry":     "file.txt",
				"remaining_entries": []string{"other.txt"},
				"entry_count":       1,
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	result, err := client.DeleteEntry(context.Background(), "test", "file.txt")

	if err != nil {
		t.Fatalf("DeleteEntry failed: %v", err)
	}
	if result.DeletedEntry != "file.txt" {
		t.Errorf("Expected deleted entry 'file.txt', got %s", result.DeletedEntry)
	}
	if result.EntryCount != 1 {
		t.Errorf("Expected entry count 1, got %d", result.EntryCount)
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	// Server that takes a long time to respond
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.Search(ctx, "test", "query", 5)

	if err == nil {
		t.Error("Expected error due to context cancellation")
	}
	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded error, got %v", ctx.Err())
	}
}

func TestClient_NetworkError(t *testing.T) {
	// Use invalid URL to simulate network error
	client := NewClient("http://invalid-host-that-does-not-exist:9999", "")

	_, err := client.Search(context.Background(), "test", "query", 5)

	if err == nil {
		t.Error("Expected network error")
	}
}
