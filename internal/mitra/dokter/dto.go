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
	VoucherValue     float64 `json:"voucher_value"`
	PlatformFee      float64 `json:"platform_fee"`
	THRBonus         float64 `json:"thr_bonus"`
}

type CompleteOrderTxData struct {
	Note        string
	Attachments []string
}
