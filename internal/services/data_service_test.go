package services

import (
	"context"
	"errors"
	"math"
	"testing"

	"solback/internal/models"

	"gorm.io/gorm"
)

func createAuctionResultsTable(t *testing.T, db *gorm.DB) {
	t.Helper()

	query := `CREATE TABLE auction_results (
		id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
		source_file TEXT NOT NULL,
		participants INTEGER NOT NULL,
		year INTEGER NOT NULL,
		month INTEGER NOT NULL,
		region TEXT NOT NULL,
		technology TEXT NOT NULL,
		total_volume_auctioned REAL NOT NULL,
		total_volume_sold REAL NOT NULL,
		weighted_avg_price_eur_per_mwh REAL NOT NULL,
		my_total_volume REAL,
		my_weighted_avg_price_eur_per_mwh REAL,
		number_of_winners INTEGER NOT NULL
	);`
	if err := db.Exec(query).Error; err != nil {
		t.Fatalf("create auction_results table: %v", err)
	}
}

func TestDataServiceStoreAuctionResults(t *testing.T) {
	db := openTestDB(t)
	createAuctionResultsTable(t, db)

	logWriter := &stubLogWriter{}
	service, err := NewDataService(db, logWriter)
	if err != nil {
		t.Fatalf("NewDataService: %v", err)
	}

	results := AuctionResults{
		SourceFile:   "file.xlsx",
		Participants: 34,
		Rows: []AuctionRow{
			{
				Year:                      2025,
				Month:                     8,
				Region:                    "Region1",
				Technology:                "Tech1",
				TotalVolumeAuctioned:      1.2,
				TotalVolumeSold:           1.1,
				WeightedAvgPriceEurPerMwh: 0.3,
				NumberOfWinners:           1,
			},
		},
	}

	count, err := service.StoreAuctionResults(context.Background(), results, nil)
	if err != nil {
		t.Fatalf("StoreAuctionResults: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}

	var stored []models.AuctionResult
	if err := db.Find(&stored).Error; err != nil {
		t.Fatalf("select auction results: %v", err)
	}
	if len(stored) != 1 {
		t.Fatalf("stored rows = %d, want 1", len(stored))
	}
	if stored[0].SourceFile != "file.xlsx" {
		t.Fatalf("source_file = %q, want %q", stored[0].SourceFile, "file.xlsx")
	}
	if len(logWriter.entries) == 0 {
		t.Fatalf("expected log entries")
	}
}

func TestDataServiceGetDataFilters(t *testing.T) {
	db := openTestDB(t)
	createAuctionResultsTable(t, db)

	rows := []models.AuctionResult{
		{
			ID:                        "row-1",
			SourceFile:                "file.xlsx",
			Participants:              10,
			Year:                      2024,
			Month:                     6,
			Region:                    "Region1",
			Technology:                "Solar",
			TotalVolumeAuctioned:      1,
			TotalVolumeSold:           1,
			WeightedAvgPriceEurPerMwh: 0.3,
			NumberOfWinners:           1,
		},
		{
			ID:                        "row-2",
			SourceFile:                "file.xlsx",
			Participants:              10,
			Year:                      2026,
			Month:                     1,
			Region:                    "Region2",
			Technology:                "Wind",
			TotalVolumeAuctioned:      2,
			TotalVolumeSold:           2,
			WeightedAvgPriceEurPerMwh: 0.4,
			NumberOfWinners:           1,
		},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("insert rows: %v", err)
	}

	service, err := NewDataService(db, &stubLogWriter{})
	if err != nil {
		t.Fatalf("NewDataService: %v", err)
	}

	results, err := service.GetData(context.Background(), "2024-2025", "Solar", "", false, "", "", "", "", "")
	if err != nil {
		t.Fatalf("GetData: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("results = %d, want 1", len(results))
	}
	if results[0].Technology != "Solar" {
		t.Fatalf("technology = %q, want %q", results[0].Technology, "Solar")
	}
}

func TestDataServiceGetDataInvalidPeriod(t *testing.T) {
	db := openTestDB(t)
	createAuctionResultsTable(t, db)

	service, err := NewDataService(db, &stubLogWriter{})
	if err != nil {
		t.Fatalf("NewDataService: %v", err)
	}

	if _, err := service.GetData(context.Background(), "2024", "", "", false, "", "", "", "", ""); !errors.Is(err, ErrInvalidPeriod) {
		t.Fatalf("expected ErrInvalidPeriod, got %v", err)
	}
}

func TestDataServiceGetDataInvalidGroupPeriod(t *testing.T) {
	db := openTestDB(t)
	createAuctionResultsTable(t, db)

	service, err := NewDataService(db, &stubLogWriter{})
	if err != nil {
		t.Fatalf("NewDataService: %v", err)
	}

	if _, err := service.GetData(context.Background(), "", "", "quarter", false, "", "", "", "", ""); !errors.Is(err, ErrInvalidGroupPeriod) {
		t.Fatalf("expected ErrInvalidGroupPeriod, got %v", err)
	}
}

func TestDataServiceGetDataInvalidMonthRange(t *testing.T) {
	db := openTestDB(t)
	createAuctionResultsTable(t, db)

	service, err := NewDataService(db, &stubLogWriter{})
	if err != nil {
		t.Fatalf("NewDataService: %v", err)
	}

	if _, err := service.GetData(context.Background(), "", "", "", false, "2024-13", "", "", "", ""); !errors.Is(err, ErrInvalidMonthRange) {
		t.Fatalf("expected ErrInvalidMonthRange, got %v", err)
	}
	if _, err := service.GetData(context.Background(), "", "", "", false, "2025-02", "2025-01", "", "", ""); !errors.Is(err, ErrInvalidMonthRange) {
		t.Fatalf("expected ErrInvalidMonthRange, got %v", err)
	}
}

func TestDataServiceGetDataMonthRangeFilter(t *testing.T) {
	db := openTestDB(t)
	createAuctionResultsTable(t, db)

	rows := []models.AuctionResult{
		{
			ID:                        "row-1",
			SourceFile:                "file.xlsx",
			Participants:              10,
			Year:                      2024,
			Month:                     1,
			Region:                    "Region1",
			Technology:                "Solar",
			TotalVolumeAuctioned:      1,
			TotalVolumeSold:           1,
			WeightedAvgPriceEurPerMwh: 0.3,
			NumberOfWinners:           1,
		},
		{
			ID:                        "row-2",
			SourceFile:                "file.xlsx",
			Participants:              10,
			Year:                      2024,
			Month:                     6,
			Region:                    "Region2",
			Technology:                "Solar",
			TotalVolumeAuctioned:      2,
			TotalVolumeSold:           2,
			WeightedAvgPriceEurPerMwh: 0.4,
			NumberOfWinners:           1,
		},
		{
			ID:                        "row-3",
			SourceFile:                "file.xlsx",
			Participants:              10,
			Year:                      2025,
			Month:                     1,
			Region:                    "Region3",
			Technology:                "Solar",
			TotalVolumeAuctioned:      3,
			TotalVolumeSold:           3,
			WeightedAvgPriceEurPerMwh: 0.5,
			NumberOfWinners:           1,
		},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("insert rows: %v", err)
	}

	service, err := NewDataService(db, &stubLogWriter{})
	if err != nil {
		t.Fatalf("NewDataService: %v", err)
	}

	results, err := service.GetData(context.Background(), "", "", "", false, "2024-05", "2024-12", "", "", "")
	if err != nil {
		t.Fatalf("GetData: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("results = %d, want 1", len(results))
	}
	if results[0].Year != 2024 || results[0].Month != 6 {
		t.Fatalf("year/month = %d/%d, want 2024/6", results[0].Year, results[0].Month)
	}
}

func TestDataServiceGetDataTechInFilter(t *testing.T) {
	db := openTestDB(t)
	createAuctionResultsTable(t, db)

	rows := []models.AuctionResult{
		{
			ID:                        "row-1",
			SourceFile:                "file.xlsx",
			Participants:              10,
			Year:                      2024,
			Month:                     1,
			Region:                    "Region1",
			Technology:                "Solar",
			TotalVolumeAuctioned:      1,
			TotalVolumeSold:           1,
			WeightedAvgPriceEurPerMwh: 0.3,
			NumberOfWinners:           1,
		},
		{
			ID:                        "row-2",
			SourceFile:                "file.xlsx",
			Participants:              10,
			Year:                      2024,
			Month:                     1,
			Region:                    "Region2",
			Technology:                "Hydro",
			TotalVolumeAuctioned:      2,
			TotalVolumeSold:           2,
			WeightedAvgPriceEurPerMwh: 0.4,
			NumberOfWinners:           1,
		},
		{
			ID:                        "row-3",
			SourceFile:                "file.xlsx",
			Participants:              10,
			Year:                      2024,
			Month:                     1,
			Region:                    "Region3",
			Technology:                "Wind",
			TotalVolumeAuctioned:      3,
			TotalVolumeSold:           3,
			WeightedAvgPriceEurPerMwh: 0.5,
			NumberOfWinners:           1,
		},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("insert rows: %v", err)
	}

	service, err := NewDataService(db, &stubLogWriter{})
	if err != nil {
		t.Fatalf("NewDataService: %v", err)
	}

	results, err := service.GetData(context.Background(), "", "", "", false, "", "", "Solar, HYDRO", "", "")
	if err != nil {
		t.Fatalf("GetData: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}
}

func TestDataServiceGetDataSortAndLimit(t *testing.T) {
	db := openTestDB(t)
	createAuctionResultsTable(t, db)

	rows := []models.AuctionResult{
		{
			ID:                        "row-1",
			SourceFile:                "file.xlsx",
			Participants:              10,
			Year:                      2024,
			Month:                     12,
			Region:                    "Region1",
			Technology:                "Solar",
			TotalVolumeAuctioned:      1,
			TotalVolumeSold:           1,
			WeightedAvgPriceEurPerMwh: 0.3,
			NumberOfWinners:           1,
		},
		{
			ID:                        "row-2",
			SourceFile:                "file.xlsx",
			Participants:              10,
			Year:                      2025,
			Month:                     1,
			Region:                    "Region2",
			Technology:                "Solar",
			TotalVolumeAuctioned:      2,
			TotalVolumeSold:           2,
			WeightedAvgPriceEurPerMwh: 0.4,
			NumberOfWinners:           1,
		},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("insert rows: %v", err)
	}

	service, err := NewDataService(db, &stubLogWriter{})
	if err != nil {
		t.Fatalf("NewDataService: %v", err)
	}

	results, err := service.GetData(context.Background(), "", "", "", false, "", "", "", "year_desc,month_desc", "1")
	if err != nil {
		t.Fatalf("GetData: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("results = %d, want 1", len(results))
	}
	if results[0].Year != 2025 || results[0].Month != 1 {
		t.Fatalf("year/month = %d/%d, want 2025/1", results[0].Year, results[0].Month)
	}
}

func TestDataServiceGetDataInvalidLimit(t *testing.T) {
	db := openTestDB(t)
	createAuctionResultsTable(t, db)

	service, err := NewDataService(db, &stubLogWriter{})
	if err != nil {
		t.Fatalf("NewDataService: %v", err)
	}

	if _, err := service.GetData(context.Background(), "", "", "", false, "", "", "", "", "0"); !errors.Is(err, ErrInvalidLimit) {
		t.Fatalf("expected ErrInvalidLimit, got %v", err)
	}
	if _, err := service.GetData(context.Background(), "", "", "", false, "", "", "", "", "nope"); !errors.Is(err, ErrInvalidLimit) {
		t.Fatalf("expected ErrInvalidLimit, got %v", err)
	}
}

func TestDataServiceGetDataInvalidSort(t *testing.T) {
	db := openTestDB(t)
	createAuctionResultsTable(t, db)

	service, err := NewDataService(db, &stubLogWriter{})
	if err != nil {
		t.Fatalf("NewDataService: %v", err)
	}

	if _, err := service.GetData(context.Background(), "", "", "", false, "", "", "", "region_desc", ""); !errors.Is(err, ErrInvalidSort) {
		t.Fatalf("expected ErrInvalidSort, got %v", err)
	}
}

func TestDataServiceGetDataGroupedByYear(t *testing.T) {
	db := openTestDB(t)
	createAuctionResultsTable(t, db)

	rows := []models.AuctionResult{
		{
			ID:                        "row-1",
			SourceFile:                "file.xlsx",
			Participants:              10,
			Year:                      2024,
			Month:                     1,
			Region:                    "North",
			Technology:                "Solar",
			TotalVolumeAuctioned:      1,
			TotalVolumeSold:           1,
			WeightedAvgPriceEurPerMwh: 0.5,
			NumberOfWinners:           1,
		},
		{
			ID:                        "row-2",
			SourceFile:                "file.xlsx",
			Participants:              10,
			Year:                      2024,
			Month:                     2,
			Region:                    "South",
			Technology:                "Solar",
			TotalVolumeAuctioned:      2,
			TotalVolumeSold:           2,
			WeightedAvgPriceEurPerMwh: 1.5,
			NumberOfWinners:           1,
		},
		{
			ID:                        "row-3",
			SourceFile:                "file.xlsx",
			Participants:              10,
			Year:                      2024,
			Month:                     2,
			Region:                    "East",
			Technology:                "Wind",
			TotalVolumeAuctioned:      3,
			TotalVolumeSold:           3,
			WeightedAvgPriceEurPerMwh: 2.5,
			NumberOfWinners:           1,
		},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("insert rows: %v", err)
	}

	service, err := NewDataService(db, &stubLogWriter{})
	if err != nil {
		t.Fatalf("NewDataService: %v", err)
	}

	results, err := service.GetData(context.Background(), "", "", "year", false, "", "", "", "", "")
	if err != nil {
		t.Fatalf("GetData: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}

	var solar models.AuctionResult
	var wind models.AuctionResult
	for _, row := range results {
		if row.Region != "" {
			t.Fatalf("region = %q, want empty", row.Region)
		}
		switch row.Technology {
		case "Solar":
			solar = row
		case "Wind":
			wind = row
		}
	}

	if solar.TotalVolumeAuctioned != 3 || solar.TotalVolumeSold != 3 {
		t.Fatalf("solar totals = %v/%v, want 3/3", solar.TotalVolumeAuctioned, solar.TotalVolumeSold)
	}
	if math.Abs(solar.WeightedAvgPriceEurPerMwh-1.0) > 1e-9 {
		t.Fatalf("solar avg price = %v, want 1.0", solar.WeightedAvgPriceEurPerMwh)
	}
	if wind.TotalVolumeAuctioned != 3 || wind.TotalVolumeSold != 3 {
		t.Fatalf("wind totals = %v/%v, want 3/3", wind.TotalVolumeAuctioned, wind.TotalVolumeSold)
	}
	if math.Abs(wind.WeightedAvgPriceEurPerMwh-2.5) > 1e-9 {
		t.Fatalf("wind avg price = %v, want 2.5", wind.WeightedAvgPriceEurPerMwh)
	}
}

func TestDataServiceGetDataGroupedByMonthSumTech(t *testing.T) {
	db := openTestDB(t)
	createAuctionResultsTable(t, db)

	rows := []models.AuctionResult{
		{
			ID:                        "row-1",
			SourceFile:                "file.xlsx",
			Participants:              10,
			Year:                      2024,
			Month:                     1,
			Region:                    "South",
			Technology:                "Solar",
			TotalVolumeAuctioned:      1,
			TotalVolumeSold:           1,
			WeightedAvgPriceEurPerMwh: 0.5,
			NumberOfWinners:           1,
		},
		{
			ID:                        "row-2",
			SourceFile:                "file.xlsx",
			Participants:              10,
			Year:                      2024,
			Month:                     1,
			Region:                    "West",
			Technology:                "Wind",
			TotalVolumeAuctioned:      3,
			TotalVolumeSold:           2,
			WeightedAvgPriceEurPerMwh: 1.5,
			NumberOfWinners:           1,
		},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("insert rows: %v", err)
	}

	service, err := NewDataService(db, &stubLogWriter{})
	if err != nil {
		t.Fatalf("NewDataService: %v", err)
	}

	results, err := service.GetData(context.Background(), "", "", "month", true, "", "", "", "", "")
	if err != nil {
		t.Fatalf("GetData: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("results = %d, want 1", len(results))
	}
	row := results[0]
	if row.Region != "" {
		t.Fatalf("region = %q, want empty", row.Region)
	}
	if row.TotalVolumeAuctioned != 4 || row.TotalVolumeSold != 3 {
		t.Fatalf("totals = %v/%v, want 4/3", row.TotalVolumeAuctioned, row.TotalVolumeSold)
	}
	if math.Abs(row.WeightedAvgPriceEurPerMwh-1.0) > 1e-9 {
		t.Fatalf("avg price = %v, want 1.0", row.WeightedAvgPriceEurPerMwh)
	}
	if row.Technology != "" {
		t.Fatalf("technology = %q, want empty", row.Technology)
	}
}

func TestDataServiceGetDataSumTechDefaultsToMonth(t *testing.T) {
	db := openTestDB(t)
	createAuctionResultsTable(t, db)

	rows := []models.AuctionResult{
		{
			ID:                        "row-1",
			SourceFile:                "file.xlsx",
			Participants:              10,
			Year:                      2025,
			Month:                     3,
			Region:                    "North",
			Technology:                "Solar",
			TotalVolumeAuctioned:      2,
			TotalVolumeSold:           2,
			WeightedAvgPriceEurPerMwh: 0.8,
			NumberOfWinners:           1,
		},
		{
			ID:                        "row-2",
			SourceFile:                "file.xlsx",
			Participants:              10,
			Year:                      2025,
			Month:                     3,
			Region:                    "South",
			Technology:                "Wind",
			TotalVolumeAuctioned:      4,
			TotalVolumeSold:           3,
			WeightedAvgPriceEurPerMwh: 1.2,
			NumberOfWinners:           1,
		},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("insert rows: %v", err)
	}

	service, err := NewDataService(db, &stubLogWriter{})
	if err != nil {
		t.Fatalf("NewDataService: %v", err)
	}

	results, err := service.GetData(context.Background(), "", "", "", true, "", "", "", "", "")
	if err != nil {
		t.Fatalf("GetData: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("results = %d, want 1", len(results))
	}
	row := results[0]
	if row.Year != 2025 || row.Month != 3 {
		t.Fatalf("year/month = %d/%d, want 2025/3", row.Year, row.Month)
	}
	if row.Region != "" {
		t.Fatalf("region = %q, want empty", row.Region)
	}
	if row.Technology != "" {
		t.Fatalf("technology = %q, want empty", row.Technology)
	}
	if row.TotalVolumeAuctioned != 6 || row.TotalVolumeSold != 5 {
		t.Fatalf("totals = %v/%v, want 6/5", row.TotalVolumeAuctioned, row.TotalVolumeSold)
	}
	if math.Abs(row.WeightedAvgPriceEurPerMwh-1.0) > 1e-9 {
		t.Fatalf("avg price = %v, want 1.0", row.WeightedAvgPriceEurPerMwh)
	}
}

func TestDataServiceDeleteData(t *testing.T) {
	db := openTestDB(t)
	createAuctionResultsTable(t, db)

	row := models.AuctionResult{
		ID:                        "row-1",
		SourceFile:                "file.xlsx",
		Participants:              10,
		Year:                      2024,
		Month:                     6,
		Region:                    "Region1",
		Technology:                "Solar",
		TotalVolumeAuctioned:      1,
		TotalVolumeSold:           1,
		WeightedAvgPriceEurPerMwh: 0.3,
		NumberOfWinners:           1,
	}
	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("insert row: %v", err)
	}

	logWriter := &stubLogWriter{}
	service, err := NewDataService(db, logWriter)
	if err != nil {
		t.Fatalf("NewDataService: %v", err)
	}

	count, err := service.DeleteData(context.Background())
	if err != nil {
		t.Fatalf("DeleteData: %v", err)
	}
	if count != 1 {
		t.Fatalf("deleted = %d, want 1", count)
	}

	var remaining []models.AuctionResult
	if err := db.Find(&remaining).Error; err != nil {
		t.Fatalf("select remaining: %v", err)
	}
	if len(remaining) != 0 {
		t.Fatalf("remaining rows = %d, want 0", len(remaining))
	}
	if len(logWriter.entries) == 0 {
		t.Fatalf("expected log entries")
	}
}
