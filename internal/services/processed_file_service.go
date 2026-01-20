package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"solback/internal/models"

	"gorm.io/gorm"
)

type ProcessedFileService struct {
	db *gorm.DB
}

func NewProcessedFileService(db *gorm.DB) (*ProcessedFileService, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	return &ProcessedFileService{db: db}, nil
}

func (s *ProcessedFileService) IsProcessed(ctx context.Context, filename string) (bool, error) {
	if s == nil {
		return false, errors.New("processed file service is nil")
	}
	if s.db == nil {
		return false, errors.New("db is nil")
	}
	if filename == "" {
		return false, errors.New("filename is empty")
	}

	var count int64
	if err := s.db.WithContext(ctx).Model(&models.ProcessedFile{}).Where("zip_filename = ?", filename).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check processed file: %w", err)
	}

	return count > 0, nil
}

func (s *ProcessedFileService) MarkProcessed(ctx context.Context, filename string) error {
	if s == nil {
		return errors.New("processed file service is nil")
	}
	if s.db == nil {
		return errors.New("db is nil")
	}
	if filename == "" {
		return errors.New("filename is empty")
	}

	entry := models.ProcessedFile{
		ZipFilename: filename,
		ProcessedAt: time.Now().UTC(),
	}

	if err := s.db.WithContext(ctx).Where("zip_filename = ?", filename).FirstOrCreate(&entry).Error; err != nil {
		return fmt.Errorf("mark processed file: %w", err)
	}

	return nil
}
