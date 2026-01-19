package services

import (
	"context"
	"errors"
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

	results, err := service.GetData(context.Background(), "2024-2025", "Solar")
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

	if _, err := service.GetData(context.Background(), "2024", ""); !errors.Is(err, ErrInvalidPeriod) {
		t.Fatalf("expected ErrInvalidPeriod, got %v", err)
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
