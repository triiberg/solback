package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestZipServiceDownloadRelativeLink(t *testing.T) {
	zipPath := filepath.Join("..", "..", "docs", "20251119_GO_2024_2025_GLOBAL_Results.zip")
	zipBytes, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("read zip file: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/file.zip" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(zipBytes)
	}))
	defer server.Close()

	logWriter := &stubLogWriter{}
	service, err := NewZipService(logWriter, server.Client())
	if err != nil {
		t.Fatalf("NewZipService: %v", err)
	}

	result, err := service.Download(context.Background(), "/file.zip", server.URL+"/page", nil)
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if len(result.Bytes) != len(zipBytes) {
		t.Fatalf("zip bytes length = %d, want %d", len(result.Bytes), len(zipBytes))
	}
	if !strings.HasPrefix(result.URL, server.URL) {
		t.Fatalf("resolved url = %q, want prefix %q", result.URL, server.URL)
	}
	if len(logWriter.entries) == 0 {
		t.Fatalf("expected log entries")
	}
	if logWriter.entries[len(logWriter.entries)-1].outcome != LogOutcomeSuccess {
		t.Fatalf("log outcome = %q, want %q", logWriter.entries[len(logWriter.entries)-1].outcome, LogOutcomeSuccess)
	}
}

func TestZipServiceRejectsNonZip(t *testing.T) {
	logWriter := &stubLogWriter{}
	service, err := NewZipService(logWriter, http.DefaultClient)
	if err != nil {
		t.Fatalf("NewZipService: %v", err)
	}

	if _, err := service.Download(context.Background(), "https://example.com/file.pdf", "", nil); err == nil {
		t.Fatalf("expected error for non-zip link")
	}
	if len(logWriter.entries) == 0 {
		t.Fatalf("expected log entries")
	}
}
