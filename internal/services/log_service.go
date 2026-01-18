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

func (s *LogService) CreateLog(ctx context.Context, action string, outcome string, message *string) error {
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

func (s *LogService) GetLogs(ctx context.Context, limit int) ([]models.Log, error) {
	if s == nil {
		return nil, errors.New("log service is nil")
	}
	if s.db == nil {
		return nil, errors.New("db is nil")
	}
	if limit <= 0 {
		return nil, errors.New("limit must be positive")
	}

	var logs []models.Log
	if err := s.db.WithContext(ctx).
		Order("datetime desc").
		Limit(limit).
		Find(&logs).
		Error; err != nil {
		return nil, fmt.Errorf("get logs: %w", err)
	}

	return logs, nil
}
