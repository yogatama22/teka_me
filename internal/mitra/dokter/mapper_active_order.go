package dokter

import "teka-api/internal/models"

func mapActiveServiceOrder(row activeServiceOrderRow) *models.ActiveServiceOrder {
	return &models.ActiveServiceOrder{
		ID:          row.ID,
		RequestID:   row.RequestID,
		StartTime:   row.StartTime,
		StatusID:    row.StatusID,
		StatusName:  row.StatusName,
		OrderNumber: row.OrderNumber,
		Price:       row.Price,
		Keluhan:     row.Keluhan, // 🔥 INI YANG HILANG

		Customer: models.CustomerInfo{
			ID:    row.CustomerID,
			Nama:  row.CustomerNama,
			Phone: row.CustomerPhone,
		},
		Mitra: models.MitraInfo{
			ID:    row.MitraID,
			Nama:  row.MitraNama,
			Phone: row.MitraPhone,
		},

		CustomerLat: row.CustomerLat,
		CustomerLng: row.CustomerLng,
		MitraLat:    row.MitraLat,
		MitraLng:    row.MitraLng,
		CreatedAt:   row.CreatedAt,
	}
}
