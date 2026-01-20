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
var ErrInvalidGroupPeriod = errors.New("invalid group period")
var ErrInvalidMonthRange = errors.New("invalid month range")
var ErrInvalidSort = errors.New("invalid sort")
var ErrInvalidLimit = errors.New("invalid limit")

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

func (s *DataService) GetData(ctx context.Context, period string, technology string, groupPeriod string, sumTech bool, from string, to string, techIn string, sort string, limit string) ([]models.AuctionResult, error) {
	if s == nil {
		return nil, errors.New("data service is nil")
	}
	if s.db == nil {
		return nil, errors.New("db is nil")
	}

	groupPeriod, err := normalizeGroupPeriod(groupPeriod)
	if err != nil {
		return nil, err
	}
	if groupPeriod == "" && sumTech {
		groupPeriod = "month"
	}

	limitValue, err := parseLimit(limit)
	if err != nil {
		return nil, err
	}

	fromYear, fromMonth, hasFrom, toYear, toMonth, hasTo, err := parseMonthRange(from, to)
	if err != nil {
		return nil, err
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

	techList := parseTechIn(techIn)
	if len(techList) > 0 {
		query = query.Where("lower(technology) IN ?", techList)
	}

	if hasFrom {
		query = query.Where("(year > ?) OR (year = ? AND month >= ?)", fromYear, fromYear, fromMonth)
	}
	if hasTo {
		query = query.Where("(year < ?) OR (year = ? AND month <= ?)", toYear, toYear, toMonth)
	}

	sortParts, err := parseSortParts(sort)
	if err != nil {
		return nil, err
	}

	if groupPeriod != "" {
		selectFields := []string{"year"}
		groupFields := []string{"year"}
		orderFields := []string{"year"}
		if groupPeriod == "month" {
			selectFields = append(selectFields, "month")
			groupFields = append(groupFields, "month")
			orderFields = append(orderFields, "month")
		}
		if !sumTech {
			selectFields = append(selectFields, "technology")
			groupFields = append(groupFields, "technology")
			orderFields = append(orderFields, "technology")
		}

		selectFields = append(
			selectFields,
			"SUM(total_volume_auctioned) AS total_volume_auctioned",
			"SUM(total_volume_sold) AS total_volume_sold",
			"AVG(weighted_avg_price_eur_per_mwh) AS weighted_avg_price_eur_per_mwh",
		)
		query = query.Select(strings.Join(selectFields, ", ")).Group(strings.Join(groupFields, ", "))
		if len(sortParts) > 0 {
			allowed := map[string]bool{"year": true}
			if groupPeriod == "month" {
				allowed["month"] = true
			}
			if !sumTech {
				allowed["technology"] = true
			}
			orderClause, err := buildOrderClause(sortParts, allowed)
			if err != nil {
				return nil, err
			}
			query = query.Order(orderClause)
		} else {
			query = query.Order(strings.Join(orderFields, ", "))
		}
	} else {
		if len(sortParts) > 0 {
			allowed := map[string]bool{"year": true, "month": true, "technology": true}
			orderClause, err := buildOrderClause(sortParts, allowed)
			if err != nil {
				return nil, err
			}
			query = query.Order(orderClause)
		} else {
			query = query.Order("year, month, region, technology")
		}
	}

	if limitValue > 0 {
		query = query.Limit(limitValue)
	}

	var results []models.AuctionResult
	if err := query.Find(&results).Error; err != nil {
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

func normalizeGroupPeriod(groupPeriod string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(groupPeriod))
	if normalized == "" {
		return "", nil
	}
	if normalized != "year" && normalized != "month" {
		return "", ErrInvalidGroupPeriod
	}
	return normalized, nil
}

type sortPart struct {
	Field     string
	Direction string
}

func parseMonthRange(from string, to string) (int, int, bool, int, int, bool, error) {
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	hasFrom := from != ""
	hasTo := to != ""
	if !hasFrom && !hasTo {
		return 0, 0, false, 0, 0, false, nil
	}

	var fromYear, fromMonth int
	var toYear, toMonth int
	var err error
	if hasFrom {
		fromYear, fromMonth, err = parseYearMonth(from)
		if err != nil {
			return 0, 0, false, 0, 0, false, err
		}
	}
	if hasTo {
		toYear, toMonth, err = parseYearMonth(to)
		if err != nil {
			return 0, 0, false, 0, 0, false, err
		}
	}

	if hasFrom && hasTo {
		if fromYear > toYear || (fromYear == toYear && fromMonth > toMonth) {
			return 0, 0, false, 0, 0, false, ErrInvalidMonthRange
		}
	}

	return fromYear, fromMonth, hasFrom, toYear, toMonth, hasTo, nil
}

func parseYearMonth(value string) (int, int, error) {
	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) != 2 {
		return 0, 0, ErrInvalidMonthRange
	}

	year, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || year <= 0 {
		return 0, 0, ErrInvalidMonthRange
	}
	month, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || month < 1 || month > 12 {
		return 0, 0, ErrInvalidMonthRange
	}

	return year, month, nil
}

func parseTechIn(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	techList := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.ToLower(strings.TrimSpace(part))
		if trimmed != "" {
			techList = append(techList, trimmed)
		}
	}

	return techList
}

func parseSortParts(value string) ([]sortPart, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	entries := strings.Split(trimmed, ",")
	parts := make([]sortPart, 0, len(entries))
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			return nil, ErrInvalidSort
		}
		field := entry
		direction := "asc"
		if idx := strings.LastIndex(entry, "_"); idx != -1 {
			field = entry[:idx]
			direction = entry[idx+1:]
		}
		field = strings.ToLower(strings.TrimSpace(field))
		direction = strings.ToLower(strings.TrimSpace(direction))
		if field == "" || direction == "" {
			return nil, ErrInvalidSort
		}
		if field != "year" && field != "month" && field != "technology" {
			return nil, ErrInvalidSort
		}
		if direction != "asc" && direction != "desc" {
			return nil, ErrInvalidSort
		}
		parts = append(parts, sortPart{Field: field, Direction: direction})
	}

	return parts, nil
}

func buildOrderClause(parts []sortPart, allowed map[string]bool) (string, error) {
	if len(parts) == 0 {
		return "", nil
	}

	clauses := make([]string, 0, len(parts))
	for _, part := range parts {
		if !allowed[part.Field] {
			return "", ErrInvalidSort
		}
		clauses = append(clauses, fmt.Sprintf("%s %s", part.Field, strings.ToUpper(part.Direction)))
	}

	return strings.Join(clauses, ", "), nil
}

func parseLimit(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}

	limit, err := strconv.Atoi(value)
	if err != nil || limit <= 0 {
		return 0, ErrInvalidLimit
	}

	return limit, nil
}
