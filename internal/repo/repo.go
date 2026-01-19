package repo

import (
	"errors"
	"fmt"

	"solback/internal/config"
	"solback/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const defaultSourceConfigPath = "config.json"

func Connect(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("dsn is empty")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	return db, nil
}

func Migrate(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}

	if err := db.AutoMigrate(&models.Source{}, &models.Log{}, &models.AuctionResult{}); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	if err := ensureDefaultSource(db); err != nil {
		return fmt.Errorf("ensure default source: %w", err)
	}

	return nil
}

func ensureDefaultSource(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}

	var count int64
	if err := db.Model(&models.Source{}).Count(&count).Error; err != nil {
		return fmt.Errorf("count sources: %w", err)
	}
	if count > 0 {
		return nil
	}

	cfg, err := config.LoadSourceConfig(defaultSourceConfigPath)
	if err != nil {
		return err
	}

	comment := cfg.Source.Comment
	source := models.Source{
		URL:     cfg.Source.URL,
		Comment: &comment,
	}
	if err := db.Create(&source).Error; err != nil {
		return fmt.Errorf("create default source: %w", err)
	}

	return nil
}
