package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"solback/internal/models"

	"gorm.io/gorm"
)

type LogService struct {
	db *gorm.DB
}

func NewLogService(db *gorm.DB) (*LogService, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	return &LogService{db: db}, nil
}

func (s *LogService) CreateLog(ctx context.Context, eventID *string, action string, outcome string, message *string) error {
	if s == nil {
		return errors.New("log service is nil")
	}
	if s.db == nil {
		return errors.New("db is nil")
	}
	if action == "" {
		return errors.New("action is empty")
	}
	if outcome == "" {
		return errors.New("outcome is empty")
	}

	entry := models.Log{
		EventID:  eventID,
		Datetime: time.Now().UTC(),
		Action:   action,
		Outcome:  outcome,
		Message:  message,
	}
	if err := s.db.WithContext(ctx).Create(&entry).Error; err != nil {
		return fmt.Errorf("create log: %w", err)
	}

	return nil
}

func (s *LogService) GetLogs(ctx context.Context, limit int, eventID string) ([]models.Log, error) {
	if s == nil {
		return nil, errors.New("log service is nil")
	}
	if s.db == nil {
		return nil, errors.New("db is nil")
	}
	if limit <= 0 {
		return nil, errors.New("limit must be positive")
	}

	query := s.db.WithContext(ctx).Order("datetime desc").Limit(limit)
	if eventID != "" {
		query = query.Where("event_id = ?", eventID)
	}

	var logs []models.Log
	if err := query.Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("get logs: %w", err)
	}

	return logs, nil
}

func (s *LogService) TruncateLogs(ctx context.Context) (int, error) {
	if s == nil {
		return 0, errors.New("log service is nil")
	}
	if s.db == nil {
		return 0, errors.New("db is nil")
	}

	result := s.db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Log{})
	if result.Error != nil {
		return 0, fmt.Errorf("truncate logs: %w", result.Error)
	}

	return int(result.RowsAffected), nil
}
