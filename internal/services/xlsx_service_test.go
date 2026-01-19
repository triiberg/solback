package services

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestXlsxServiceExtractAuctionPayloads(t *testing.T) {
	zipPath := filepath.Join("..", "..", "docs", "20251119_GO_2024_2025_GLOBAL_Results.zip")
	zipBytes, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("read zip file: %v", err)
	}

	service, err := NewXlsxService()
	if err != nil {
		t.Fatalf("NewXlsxService: %v", err)
	}

	payloads, err := service.ExtractAuctionPayloads(context.Background(), zipBytes)
	if err != nil {
		t.Fatalf("ExtractAuctionPayloads: %v", err)
	}
	if len(payloads) == 0 {
		t.Fatalf("expected payloads, got 0")
	}

	var found bool
	for _, payload := range payloads {
		if strings.Contains(payload.SourceFile, "20251119_August_2025_83") {
			found = true
			if payload.Participants <= 0 {
				t.Fatalf("participants = %d, want positive", payload.Participants)
			}
			if len(payload.Headers) == 0 {
				t.Fatalf("headers are empty")
			}
			if len(payload.Rows) == 0 {
				t.Fatalf("rows are empty")
			}
			break
		}
	}
	if !found {
		t.Fatalf("expected August 2025 payload in zip")
	}
}
