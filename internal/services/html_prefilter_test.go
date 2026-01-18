package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractZipTablesFromExample(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "example.html")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read example.html: %v", err)
	}

	tables, err := ExtractZipTables(string(data))
	if err != nil {
		t.Fatalf("ExtractZipTables: %v", err)
	}
	if len(tables) == 0 {
		t.Fatalf("expected tables with zip links, got 0")
	}

	combined := strings.Join(tables, "\n")
	if !strings.Contains(combined, "GO 2024-2025 Global Results") {
		t.Fatalf("expected GO 2024-2025 Global Results in tables")
	}
	if !strings.Contains(strings.ToLower(combined), ".zip") {
		t.Fatalf("expected .zip link in tables")
	}
}
