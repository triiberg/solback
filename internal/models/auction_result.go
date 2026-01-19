package models

type AuctionResult struct {
	ID                          string   `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	SourceFile                  string   `gorm:"type:text;not null" json:"source_file"`
	Participants                int      `gorm:"type:int;not null" json:"participants"`
	Year                        int      `gorm:"type:int;not null" json:"year"`
	Month                       int      `gorm:"type:int;not null" json:"month"`
	Region                      string   `gorm:"type:text;not null" json:"region"`
	Technology                  string   `gorm:"type:text;not null" json:"technology"`
	TotalVolumeAuctioned        float64  `gorm:"type:double precision;not null" json:"total_volume_auctioned"`
	TotalVolumeSold             float64  `gorm:"type:double precision;not null" json:"total_volume_sold"`
	WeightedAvgPriceEurPerMwh   float64  `gorm:"type:double precision;not null" json:"weighted_avg_price_eur_per_mwh"`
	MyTotalVolume               *float64 `gorm:"type:double precision" json:"my_total_volume"`
	MyWeightedAvgPriceEurPerMwh *float64 `gorm:"type:double precision" json:"my_weighted_avg_price_eur_per_mwh"`
	NumberOfWinners             int      `gorm:"type:int;not null" json:"number_of_winners"`
}
