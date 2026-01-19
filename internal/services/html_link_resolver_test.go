package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveZipLinks(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "example.html")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read example.html: %v", err)
	}

	baseURL := "https://www.eex.com/en/markets/energy-certificates/french-auctions-power"
	resolved, err := ResolveZipLinks(baseURL, string(data))
	if err != nil {
		t.Fatalf("ResolveZipLinks: %v", err)
	}

	if !strings.Contains(resolved, "https://www.eex.com/fileadmin/") {
		t.Fatalf("expected resolved zip links to be absolute")
	}
}
