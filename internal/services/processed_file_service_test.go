package services

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"
)

func createProcessedFilesTable(t *testing.T, db *gorm.DB) {
	t.Helper()

	query := "CREATE TABLE processed_files (id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))), zip_filename TEXT NOT NULL UNIQUE, processed_at DATETIME NOT NULL)"
	if err := db.Exec(query).Error; err != nil {
		t.Fatalf("create processed_files table: %v", err)
	}
}

func TestProcessedFileServiceMarkAndCheck(t *testing.T) {
	db := openTestDB(t)
	createProcessedFilesTable(t, db)

	service, err := NewProcessedFileService(db)
	if err != nil {
		t.Fatalf("NewProcessedFileService: %v", err)
	}

	processed, err := service.IsProcessed(context.Background(), "file.zip")
	if err != nil {
		t.Fatalf("IsProcessed: %v", err)
	}
	if processed {
		t.Fatalf("expected file not processed")
	}

	if err := service.MarkProcessed(context.Background(), "file.zip"); err != nil {
		t.Fatalf("MarkProcessed: %v", err)
	}

	processed, err = service.IsProcessed(context.Background(), "file.zip")
	if err != nil {
		t.Fatalf("IsProcessed: %v", err)
	}
	if !processed {
		t.Fatalf("expected file processed")
	}
}

func TestProcessedFileServiceMarkIdempotent(t *testing.T) {
	db := openTestDB(t)
	createProcessedFilesTable(t, db)

	service, err := NewProcessedFileService(db)
	if err != nil {
		t.Fatalf("NewProcessedFileService: %v", err)
	}

	if err := service.MarkProcessed(context.Background(), "file.zip"); err != nil {
		t.Fatalf("MarkProcessed: %v", err)
	}
	if err := service.MarkProcessed(context.Background(), "file.zip"); err != nil {
		t.Fatalf("MarkProcessed second: %v", err)
	}

	var count int64
	if err := db.Model(&struct {
		ZipFilename string
		ProcessedAt time.Time
	}{}).Table("processed_files").Where("zip_filename = ?", "file.zip").Count(&count).Error; err != nil {
		t.Fatalf("count processed files: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}
