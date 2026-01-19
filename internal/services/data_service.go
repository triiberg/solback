package services

import (
	"context"
	"errors"
	"fmt"
	"math"

	"solback/internal/models"

	"gorm.io/gorm"
)

type DataService struct {
	db         *gorm.DB
	logService LogWriter
}

func NewDataService(db *gorm.DB, logService LogWriter) (*DataService, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}
	if logService == nil {
		return nil, errors.New("log service is nil")
	}

	return &DataService{
		db:         db,
		logService: logService,
	}, nil
}

func (s *DataService) StoreAuctionResults(ctx context.Context, results AuctionResults) (int, error) {
	if s == nil {
		return 0, errors.New("data service is nil")
	}
	if s.db == nil {
		return 0, errors.New("db is nil")
	}
	if s.logService == nil {
		return 0, errors.New("log service is nil")
	}
	if results.SourceFile == "" {
		return 0, errors.New("source file is empty")
	}
	if results.Participants <= 0 {
		return 0, errors.New("participants must be positive")
	}
	if len(results.Rows) == 0 {
		return 0, errors.New("rows are empty")
	}

	records := make([]models.AuctionResult, 0, len(results.Rows))
	for _, row := range results.Rows {
		if math.Trunc(row.Year) != row.Year {
			return 0, fmt.Errorf("year is not an integer: %v", row.Year)
		}
		if math.Trunc(row.Month) != row.Month {
			return 0, fmt.Errorf("month is not an integer: %v", row.Month)
		}

		year := int(row.Year)
		month := int(row.Month)

		records = append(records, models.AuctionResult{
			SourceFile:                results.SourceFile,
			Participants:              results.Participants,
			Year:                      year,
			Month:                     month,
			Region:                    row.Region,
			Technology:                row.Technology,
			TotalVolumeAuctioned:      row.TotalVolumeAuctioned,
			TotalVolumeSold:           row.TotalVolumeSold,
			WeightedAvgPriceEurPerMwh: row.WeightedAvgPriceEurPerMwh,
		})
	}

	if err := s.db.WithContext(ctx).Create(&records).Error; err != nil {
		failMsg := fmt.Sprintf("store data rows=%d source_file=%s: %v", len(records), results.SourceFile, err)
		_ = s.logService.CreateLog(ctx, LogActionDataStore, LogOutcomeFail, &failMsg)
		return 0, fmt.Errorf("store auction results: %w", err)
	}

	successMsg := fmt.Sprintf("stored rows=%d source_file=%s", len(records), results.SourceFile)
	_ = s.logService.CreateLog(ctx, LogActionDataStore, LogOutcomeSuccess, &successMsg)

	return len(records), nil
}
