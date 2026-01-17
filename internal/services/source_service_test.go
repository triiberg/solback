package services

import (
	"context"
	"testing"

	"solback/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	return db
}

func createSourcesTable(t *testing.T, db *gorm.DB) {
	t.Helper()

	if err := db.Exec("CREATE TABLE sources (id TEXT PRIMARY KEY, url TEXT NOT NULL, comment TEXT)").Error; err != nil {
		t.Fatalf("create sources table: %v", err)
	}
}

func TestNewSourceServiceNilDB(t *testing.T) {
	if _, err := NewSourceService(nil); err == nil {
		t.Fatalf("NewSourceService nil db: expected error")
	}
}

func TestSourceServiceGetSources(t *testing.T) {
	db := openTestDB(t)
	createSourcesTable(t, db)

	comment := "Default source"
	if err := db.Create(&models.Source{
		ID:      "source-id",
		URL:     "https://example.com",
		Comment: &comment,
	}).Error; err != nil {
		t.Fatalf("insert source: %v", err)
	}

	service, err := NewSourceService(db)
	if err != nil {
		t.Fatalf("NewSourceService: %v", err)
	}

	sources, err := service.GetSources(context.Background())
	if err != nil {
		t.Fatalf("GetSources: %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("sources length = %d, want 1", len(sources))
	}
	if sources[0].URL != "https://example.com" {
		t.Fatalf("URL = %q, want %q", sources[0].URL, "https://example.com")
	}
	if sources[0].Comment == nil || *sources[0].Comment != "Default source" {
		t.Fatalf("Comment = %v, want %q", sources[0].Comment, "Default source")
	}
}

func TestSourceServiceNilReceiver(t *testing.T) {
	var service *SourceService
	if _, err := service.GetSources(context.Background()); err == nil {
		t.Fatalf("GetSources nil receiver: expected error")
	}
}
