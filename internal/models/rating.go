package models

type RatingSummary struct {
	TotalRatings  int64   `json:"total_ratings"`
	AverageRating float64 `json:"average_rating"`
}

type DayHistory struct {
	Date    string        `json:"date"`
	Ratings []MitraRating `json:"ratings"`
}

type MonthlyRatingHistory struct {
	Month   string        `json:"month"`
	Year    string        `json:"year"`
	Ratings []MitraRating `json:"ratings"`
}

type MitraRating struct {
	ID             int64  `gorm:"primaryKey" json:"id"`
	ServiceOrderID int64  `json:"service_order_id"`
	MitraID        int64  `json:"mitra_id"`
	CustomerID     int64  `json:"customer_id"`
	CustomerName   string `json:"customer_name"`
	Rating         int    `json:"rating"`
	Review         string `json:"review"`
	CreatedAt      string `json:"created_at"`
}

func (MitraRating) TableName() string {
	return "myschema.mitra_ratings"
}
