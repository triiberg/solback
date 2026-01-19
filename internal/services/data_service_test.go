package services

import (
	"context"
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
		weighted_avg_price_eur_per_mwh REAL NOT NULL
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
			},
		},
	}

	count, err := service.StoreAuctionResults(context.Background(), results)
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
