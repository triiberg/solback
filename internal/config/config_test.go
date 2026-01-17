package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, dir string, name string, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	return path
}

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "secrets.json", `{"db_dsn":"dsn","openai_api_key":"key"}`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DBDSN != "dsn" {
		t.Fatalf("DBDSN = %q, want %q", cfg.DBDSN, "dsn")
	}
	if cfg.OpenAIAPIKey != "key" {
		t.Fatalf("OpenAIAPIKey = %q, want %q", cfg.OpenAIAPIKey, "key")
	}
}

func TestLoadConfigErrors(t *testing.T) {
	if _, err := Load(""); err == nil {
		t.Fatalf("Load empty path: expected error")
	}

	dir := t.TempDir()
	missingDB := writeTempFile(t, dir, "missing_db.json", `{"openai_api_key":"key"}`)
	if _, err := Load(missingDB); err == nil {
		t.Fatalf("Load missing db_dsn: expected error")
	}

	invalid := writeTempFile(t, dir, "invalid.json", "{")
	if _, err := Load(invalid); err == nil {
		t.Fatalf("Load invalid json: expected error")
	}
}

func TestLoadSourceConfig(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "config.json", `{"source":{"url":"https://example.com","comment":"Default source"}}`)

	cfg, err := LoadSourceConfig(path)
	if err != nil {
		t.Fatalf("LoadSourceConfig: %v", err)
	}
	if cfg.Source.URL != "https://example.com" {
		t.Fatalf("URL = %q, want %q", cfg.Source.URL, "https://example.com")
	}
	if cfg.Source.Comment != "Default source" {
		t.Fatalf("Comment = %q, want %q", cfg.Source.Comment, "Default source")
	}
}

func TestLoadSourceConfigErrors(t *testing.T) {
	if _, err := LoadSourceConfig(""); err == nil {
		t.Fatalf("LoadSourceConfig empty path: expected error")
	}

	dir := t.TempDir()
	missingURL := writeTempFile(t, dir, "missing_url.json", `{"source":{"comment":"Default source"}}`)
	if _, err := LoadSourceConfig(missingURL); err == nil {
		t.Fatalf("LoadSourceConfig missing url: expected error")
	}

	missingComment := writeTempFile(t, dir, "missing_comment.json", `{"source":{"url":"https://example.com"}}`)
	if _, err := LoadSourceConfig(missingComment); err == nil {
		t.Fatalf("LoadSourceConfig missing comment: expected error")
	}
}
