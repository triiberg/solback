package services

import (
	"context"
	"errors"
	"fmt"

	"solback/internal/models"

	"gorm.io/gorm"
)

type SourceService struct {
	db *gorm.DB
}

func NewSourceService(db *gorm.DB) (*SourceService, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	return &SourceService{db: db}, nil
}

func (s *SourceService) GetSources(ctx context.Context) ([]models.Source, error) {
	if s == nil {
		return nil, errors.New("source service is nil")
	}
	if s.db == nil {
		return nil, errors.New("db is nil")
	}

	var sources []models.Source
	if err := s.db.WithContext(ctx).Find(&sources).Error; err != nil {
		return nil, fmt.Errorf("get sources: %w", err)
	}

	return sources, nil
}
