package repo

import (
	"os"
	"path/filepath"
	"testing"

	"solback/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	return db
}

func createSourcesTableWithDefault(t *testing.T, db *gorm.DB) {
	t.Helper()

	query := "CREATE TABLE sources (id TEXT PRIMARY KEY DEFAULT 'test-id', url TEXT NOT NULL, comment TEXT)"
	if err := db.Exec(query).Error; err != nil {
		t.Fatalf("create sources table: %v", err)
	}
}

func createSourcesTable(t *testing.T, db *gorm.DB) {
	t.Helper()

	query := "CREATE TABLE sources (id TEXT PRIMARY KEY, url TEXT NOT NULL, comment TEXT)"
	if err := db.Exec(query).Error; err != nil {
		t.Fatalf("create sources table: %v", err)
	}
}

func writeRepoTempFile(t *testing.T, dir string, name string, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	return path
}

func TestConnectEmptyDSN(t *testing.T) {
	if _, err := Connect(""); err == nil {
		t.Fatalf("Connect empty dsn: expected error")
	}
}

func TestMigrateNilDB(t *testing.T) {
	if err := Migrate(nil); err == nil {
		t.Fatalf("Migrate nil db: expected error")
	}
}

func TestEnsureDefaultSourceInsertsWhenEmpty(t *testing.T) {
	db := openRepoTestDB(t)
	createSourcesTableWithDefault(t, db)

	dir := t.TempDir()
	writeRepoTempFile(t, dir, "config.json", `{"source":{"url":"https://example.com","comment":"Default source"}}`)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working dir: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("change working dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Errorf("restore working dir: %v", err)
		}
	})

	if err := ensureDefaultSource(db); err != nil {
		t.Fatalf("ensureDefaultSource: %v", err)
	}

	var count int64
	if err := db.Model(&models.Source{}).Count(&count).Error; err != nil {
		t.Fatalf("count sources: %v", err)
	}
	if count != 1 {
		t.Fatalf("sources count = %d, want 1", count)
	}

	var source models.Source
	if err := db.First(&source).Error; err != nil {
		t.Fatalf("select source: %v", err)
	}
	if source.URL != "https://example.com" {
		t.Fatalf("URL = %q, want %q", source.URL, "https://example.com")
	}
	if source.Comment == nil || *source.Comment != "Default source" {
		t.Fatalf("Comment = %v, want %q", source.Comment, "Default source")
	}
}

func TestEnsureDefaultSourceSkipsWhenNotEmpty(t *testing.T) {
	db := openRepoTestDB(t)
	createSourcesTable(t, db)

	insert := "INSERT INTO sources (id, url, comment) VALUES ('existing-id', 'https://example.com', 'existing')"
	if err := db.Exec(insert).Error; err != nil {
		t.Fatalf("insert source: %v", err)
	}

	if err := ensureDefaultSource(db); err != nil {
		t.Fatalf("ensureDefaultSource: %v", err)
	}

	var count int64
	if err := db.Model(&models.Source{}).Count(&count).Error; err != nil {
		t.Fatalf("count sources: %v", err)
	}
	if count != 1 {
		t.Fatalf("sources count = %d, want 1", count)
	}
}
