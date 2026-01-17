package repo

import (
	"errors"
	"fmt"

	"solback/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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

	if err := db.AutoMigrate(&models.Source{}, &models.Log{}); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	return nil
}
