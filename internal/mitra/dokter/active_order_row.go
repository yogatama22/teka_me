package dokter

import "time"

type activeServiceOrderRow struct {
	ID          int64
	RequestID   int64
	StartTime   time.Time
	StatusID    int16
	StatusName  string
	OrderNumber string
	Price       float64
	Keluhan     string

	CustomerID    int64
	CustomerNama  string
	CustomerPhone string

	MitraID    int64
	MitraNama  string
	MitraPhone string

	CustomerLat float64
	CustomerLng float64
	MitraLat    float64
	MitraLng    float64
	CreatedAt   time.Time
}
