package services

type AuctionPayload struct {
	SourceFile   string
	Participants int
	Headers      []string
	Rows         [][]string
}

type AuctionResults struct {
	SourceFile   string       `json:"source_file"`
	Participants int          `json:"participants"`
	Rows         []AuctionRow `json:"rows"`
}

type AuctionRow struct {
	Year                        float64  `json:"year"`
	Month                       float64  `json:"month"`
	Region                      string   `json:"region"`
	Technology                  string   `json:"technology"`
	TotalVolumeAuctioned        float64  `json:"total_volume_auctioned"`
	TotalVolumeSold             float64  `json:"total_volume_sold"`
	WeightedAvgPriceEurPerMwh   float64  `json:"weighted_avg_price_eur_per_mwh"`
	MyTotalVolume               *float64 `json:"my_total_volume"`
	MyWeightedAvgPriceEurPerMwh *float64 `json:"my_weighted_avg_price_eur_per_mwh"`
	NumberOfWinners             int      `json:"number_of_winners"`
}

type ZipResult struct {
	URL        string
	StatusCode int
	Bytes      []byte
}
