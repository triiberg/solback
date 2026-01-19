package services

import (
	"context"
	"testing"
	"time"

	"solback/internal/models"

	"gorm.io/gorm"
)

func createLogsTable(t *testing.T, db *gorm.DB) {
	t.Helper()

	query := "CREATE TABLE logs (id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))), event_id TEXT, datetime DATETIME NOT NULL, action TEXT NOT NULL, outcome TEXT NOT NULL, message TEXT)"
	if err := db.Exec(query).Error; err != nil {
		t.Fatalf("create logs table: %v", err)
	}
}

func TestNewLogServiceNilDB(t *testing.T) {
	if _, err := NewLogService(nil); err == nil {
		t.Fatalf("NewLogService nil db: expected error")
	}
}

func TestLogServiceCreateLog(t *testing.T) {
	db := openTestDB(t)
	createLogsTable(t, db)

	service, err := NewLogService(db)
	if err != nil {
		t.Fatalf("NewLogService: %v", err)
	}

	message := "started"
	eventID := "event-1"
	if err := service.CreateLog(context.Background(), &eventID, "DATA_RETRIVAL", "SUCCESS", &message); err != nil {
		t.Fatalf("CreateLog: %v", err)
	}

	var logs []models.Log
	if err := db.Find(&logs).Error; err != nil {
		t.Fatalf("select logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("logs length = %d, want 1", len(logs))
	}
	if logs[0].ID == "" {
		t.Fatalf("log id is empty")
	}
	if logs[0].Action != "DATA_RETRIVAL" {
		t.Fatalf("Action = %q, want %q", logs[0].Action, "DATA_RETRIVAL")
	}
	if logs[0].Outcome != "SUCCESS" {
		t.Fatalf("Outcome = %q, want %q", logs[0].Outcome, "SUCCESS")
	}
	if logs[0].Message == nil || *logs[0].Message != "started" {
		t.Fatalf("Message = %v, want %q", logs[0].Message, "started")
	}
	if logs[0].EventID == nil || *logs[0].EventID != eventID {
		t.Fatalf("EventID = %v, want %q", logs[0].EventID, eventID)
	}
	if logs[0].Datetime.IsZero() {
		t.Fatalf("Datetime is zero")
	}
}

func TestLogServiceGetLogs(t *testing.T) {
	db := openTestDB(t)
	createLogsTable(t, db)

	now := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	logs := []models.Log{
		{ID: "log-1", Datetime: now.Add(-time.Hour), Action: "DATA_RETRIVAL", Outcome: "SUCCESS"},
		{ID: "log-2", Datetime: now, Action: "DATA_RETRIVAL", Outcome: "FAIL"},
	}
	if err := db.Create(&logs).Error; err != nil {
		t.Fatalf("insert logs: %v", err)
	}

	service, err := NewLogService(db)
	if err != nil {
		t.Fatalf("NewLogService: %v", err)
	}

	latest, err := service.GetLogs(context.Background(), 1, "")
	if err != nil {
		t.Fatalf("GetLogs: %v", err)
	}
	if len(latest) != 1 {
		t.Fatalf("logs length = %d, want 1", len(latest))
	}
	if latest[0].ID != "log-2" {
		t.Fatalf("latest id = %q, want %q", latest[0].ID, "log-2")
	}
}

func TestLogServiceGetLogsEventID(t *testing.T) {
	db := openTestDB(t)
	createLogsTable(t, db)

	eventA := "event-a"
	eventB := "event-b"
	logs := []models.Log{
		{ID: "log-1", EventID: &eventA, Datetime: time.Now().Add(-time.Hour), Action: "DATA_RETRIVAL", Outcome: "SUCCESS"},
		{ID: "log-2", EventID: &eventB, Datetime: time.Now(), Action: "DATA_RETRIVAL", Outcome: "FAIL"},
	}
	if err := db.Create(&logs).Error; err != nil {
		t.Fatalf("insert logs: %v", err)
	}

	service, err := NewLogService(db)
	if err != nil {
		t.Fatalf("NewLogService: %v", err)
	}

	filtered, err := service.GetLogs(context.Background(), 10, eventA)
	if err != nil {
		t.Fatalf("GetLogs: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("logs length = %d, want 1", len(filtered))
	}
	if filtered[0].ID != "log-1" {
		t.Fatalf("log id = %q, want %q", filtered[0].ID, "log-1")
	}
}

func TestLogServiceTruncateLogs(t *testing.T) {
	db := openTestDB(t)
	createLogsTable(t, db)

	logs := []models.Log{
		{ID: "log-1", Datetime: time.Now(), Action: "DATA_RETRIVAL", Outcome: "SUCCESS"},
		{ID: "log-2", Datetime: time.Now(), Action: "DATA_RETRIVAL", Outcome: "FAIL"},
	}
	if err := db.Create(&logs).Error; err != nil {
		t.Fatalf("insert logs: %v", err)
	}

	service, err := NewLogService(db)
	if err != nil {
		t.Fatalf("NewLogService: %v", err)
	}

	deleted, err := service.TruncateLogs(context.Background())
	if err != nil {
		t.Fatalf("TruncateLogs: %v", err)
	}
	if deleted != 2 {
		t.Fatalf("deleted = %d, want 2", deleted)
	}

	var remaining []models.Log
	if err := db.Find(&remaining).Error; err != nil {
		t.Fatalf("select logs: %v", err)
	}
	if len(remaining) != 0 {
		t.Fatalf("remaining logs = %d, want 0", len(remaining))
	}
}

func TestLogServiceNilReceiver(t *testing.T) {
	var service *LogService
	if err := service.CreateLog(context.Background(), nil, "DATA_RETRIVAL", "SUCCESS", nil); err == nil {
		t.Fatalf("CreateLog nil receiver: expected error")
	}
	if _, err := service.GetLogs(context.Background(), 1, ""); err == nil {
		t.Fatalf("GetLogs nil receiver: expected error")
	}
}
