package dokter

type SearchDoctorRequest struct {
	JobCategoryID    int64   `json:"job_category_id"`
	JobSubCategoryID int64   `json:"job_sub_category_id"`
	Keluhan          string  `json:"keluhan"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	Radius           float64 `json:"radius"`
	Price            float64 `json:"price"`
	VoucherID        int64   `json:"voucher_id"`
}

type CompleteOrderTxData struct {
	Amount        float64
	Subtotal      float64
	Discount      float64
	PlatformFee   float64
	MitraIncome   float64
	PaymentMethod string
	Note          string
	Attachments   []string
}
