package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"solback/internal/models"

	"gorm.io/gorm"
)

var ErrInvalidPeriod = errors.New("invalid period")

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

func (s *DataService) StoreAuctionResults(ctx context.Context, results AuctionResults, eventID *string) (int, error) {
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
			SourceFile:                  results.SourceFile,
			Participants:                results.Participants,
			Year:                        year,
			Month:                       month,
			Region:                      row.Region,
			Technology:                  row.Technology,
			TotalVolumeAuctioned:        row.TotalVolumeAuctioned,
			TotalVolumeSold:             row.TotalVolumeSold,
			WeightedAvgPriceEurPerMwh:   row.WeightedAvgPriceEurPerMwh,
			MyTotalVolume:               row.MyTotalVolume,
			MyWeightedAvgPriceEurPerMwh: row.MyWeightedAvgPriceEurPerMwh,
			NumberOfWinners:             row.NumberOfWinners,
		})
	}

	if err := s.db.WithContext(ctx).Create(&records).Error; err != nil {
		failMsg := fmt.Sprintf("store data rows=%d source_file=%s: %v", len(records), results.SourceFile, err)
		_ = s.logService.CreateLog(ctx, eventID, LogActionDataStore, LogOutcomeFail, &failMsg)
		return 0, fmt.Errorf("store auction results: %w", err)
	}

	successMsg := fmt.Sprintf("stored rows=%d source_file=%s", len(records), results.SourceFile)
	_ = s.logService.CreateLog(ctx, eventID, LogActionDataStore, LogOutcomeSuccess, &successMsg)

	return len(records), nil
}

func (s *DataService) GetData(ctx context.Context, period string, technology string) ([]models.AuctionResult, error) {
	if s == nil {
		return nil, errors.New("data service is nil")
	}
	if s.db == nil {
		return nil, errors.New("db is nil")
	}

	query := s.db.WithContext(ctx).Model(&models.AuctionResult{})

	period = strings.TrimSpace(period)
	if period != "" {
		startYear, endYear, err := parsePeriod(period)
		if err != nil {
			return nil, err
		}
		query = query.Where("year >= ? AND year <= ?", startYear, endYear)
	}

	technology = strings.TrimSpace(technology)
	if technology != "" {
		query = query.Where("lower(technology) = lower(?)", technology)
	}

	var results []models.AuctionResult
	if err := query.Order("year, month, region, technology").Find(&results).Error; err != nil {
		return nil, fmt.Errorf("get data: %w", err)
	}

	return results, nil
}

func (s *DataService) DeleteData(ctx context.Context) (int, error) {
	if s == nil {
		return 0, errors.New("data service is nil")
	}
	if s.db == nil {
		return 0, errors.New("db is nil")
	}
	if s.logService == nil {
		return 0, errors.New("log service is nil")
	}

	result := s.db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.AuctionResult{})
	if result.Error != nil {
		failMsg := fmt.Sprintf("delete data: %v", result.Error)
		_ = s.logService.CreateLog(ctx, nil, LogActionDataStore, LogOutcomeFail, &failMsg)
		return 0, fmt.Errorf("delete data: %w", result.Error)
	}

	count := int(result.RowsAffected)
	successMsg := fmt.Sprintf("deleted rows=%d", count)
	_ = s.logService.CreateLog(ctx, nil, LogActionDataStore, LogOutcomeSuccess, &successMsg)

	return count, nil
}

func parsePeriod(period string) (int, int, error) {
	parts := strings.Split(period, "-")
	if len(parts) != 2 {
		return 0, 0, ErrInvalidPeriod
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, ErrInvalidPeriod
	}
	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, ErrInvalidPeriod
	}
	if start > end {
		return 0, 0, ErrInvalidPeriod
	}

	return start, end, nil
}
