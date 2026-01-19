package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAiServiceExtractZipLinkSuccess(t *testing.T) {
	html := `<table><tr><td>GO 2024-2025 Global Results</td><td><a href="https://example.com/file.zip">zip</a></td></tr></table>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var req openAiChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if len(req.Messages) != 1 || !strings.Contains(req.Messages[0].Content, "GO 2024-2025 Global Results") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resp := openAiChatResponse{
			Choices: []openAiChoice{
				{Message: openAiResponseMessage{Content: `{"error":"","period":"2024-2025","description":"GO 2024-2025 Global Results","link":"https://example.com/file.zip"}`}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	logWriter := &stubLogWriter{}
	service, err := NewOpenAiService("test-key", logWriter, server.Client(), server.URL)
	if err != nil {
		t.Fatalf("NewOpenAiService: %v", err)
	}

	result, err := service.ExtractZipLink(context.Background(), html, nil)
	if err != nil {
		t.Fatalf("ExtractZipLink: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("result error = %q, want empty", result.Error)
	}
	if result.Link != "https://example.com/file.zip" {
		t.Fatalf("link = %q, want %q", result.Link, "https://example.com/file.zip")
	}

	if len(logWriter.entries) != 1 {
		t.Fatalf("log entries = %d, want 1", len(logWriter.entries))
	}
	if logWriter.entries[0].outcome != LogOutcomeSuccess {
		t.Fatalf("log outcome = %q, want %q", logWriter.entries[0].outcome, LogOutcomeSuccess)
	}
}

func TestOpenAiServiceExtractZipLinkEmptyHTML(t *testing.T) {
	logWriter := &stubLogWriter{}
	service, err := NewOpenAiService("test-key", logWriter, http.DefaultClient, "https://example.com")
	if err != nil {
		t.Fatalf("NewOpenAiService: %v", err)
	}

	result, err := service.ExtractZipLink(context.Background(), "", nil)
	if err != nil {
		t.Fatalf("ExtractZipLink: %v", err)
	}
	if result.Error != "EMPTY_HTML" {
		t.Fatalf("error = %q, want %q", result.Error, "EMPTY_HTML")
	}
	if len(logWriter.entries) != 1 {
		t.Fatalf("log entries = %d, want 1", len(logWriter.entries))
	}
	if logWriter.entries[0].outcome != LogOutcomeFail {
		t.Fatalf("log outcome = %q, want %q", logWriter.entries[0].outcome, LogOutcomeFail)
	}
}
