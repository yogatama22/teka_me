package models

import "time"

type Dokter struct {
	ID                int64      `db:"id" json:"id"`
	CategoryID        int64      `db:"job_category_id" json:"category_id"`
	SubCategoryID     *int64     `db:"job_sub_category_id" json:"sub_category_id,omitempty"`
	Name              string     `db:"name" json:"name"`
	Price             float64    `db:"price" json:"price"`
	Unit              string     `db:"unit" json:"unit"`
	LocationLat       float64    `db:"lat" json:"lat"`
	LocationLng       float64    `db:"lng" json:"lng"`
	StartDate         time.Time  `db:"start_date" json:"start_date"`
	EndDate           *time.Time `db:"end_date" json:"end_date,omitempty"`
	IsActive          bool       `db:"is_active" json:"is_active"`
	VoucherID         *int64     `db:"voucher_id,omitempty" json:"voucher_id,omitempty"`
	VoucherCode       *string    `db:"voucher_code,omitempty" json:"voucher_code,omitempty"`
	VoucherType       *string    `db:"discount_type,omitempty" json:"voucher_type,omitempty"`
	VoucherValue      *float64   `db:"discount_value,omitempty" json:"voucher_value,omitempty"`
	PriceAfterVoucher *float64   `json:"price_after_voucher,omitempty"`
}

type CreateMitraRequest struct {
	FullName         string  `form:"full_name"`
	Phone            string  `form:"phone"`
	Email            string  `form:"email"`
	JobCategoryID    int     `form:"job_category_id"`
	JobSubCategoryID *int    `form:"job_sub_category_id"`
	Latitude         float64 `form:"latitude"`
	Longitude        float64 `form:"longitude"`
}

type MitraDocument struct {
	UserID  int
	DocType string
	FileURL string
}

type DoctorSearchResult struct {
	MitraID     int64   `json:"mitra_id"`
	Nama        string  `json:"nama"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	AvgRating   float64 `json:"avg_rating"`
	TotalReview int64   `json:"total_review"`
	DistanceKm  float64 `json:"distance_km"`
	IsBusy      bool    `json:"is_busy"` // tambahan untuk alert
}

type CustomerRequest struct {
	CustomerID       int64
	JobCategoryID    int64
	JobSubCategoryID *int64
	Keluhan          string
	Latitude         float64
	Longitude        float64
	Radius           float64
	Price            float64
	VoucherID        *int64
	VoucherCode      *string
	VoucherValue     *float64
	Status           string
}

type CurrentOffer struct {
	OfferID    int64 `json:"offer_id"`
	RequestID  int64 `json:"request_id"`
	CustomerID int64 `json:"customer_id"`

	Keluhan   string    `json:"keluhan"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`

	CustomerName  string `json:"customer_name"`
	CustomerPhone string `json:"customer_phone"`
}

type MitraOffer struct {
	ID          int64
	RequestID   int64
	MitraID     int64
	Sequence    int
	Status      string
	StatusID    int16
	SentAt      *time.Time
	ExpiredAt   *time.Time
	RespondedAt *time.Time
}

type RatingRequest struct {
	ServiceOrderID int64  `json:"service_order_id"`
	Rating         int    `json:"rating"` // 1–5
	Review         string `json:"review"`
}

// ServiceOrder merepresentasikan service yang dijalankan oleh dokter/mitra
type ServiceOrder struct {
	ID          int64      `json:"id"`
	RequestID   int64      `json:"request_id"`
	MitraID     int64      `json:"mitra_id"`
	CustomerID  int64      `json:"customer_id"`
	Status      string     `json:"status"` // ongoing, completed, cancelled
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	StartedAt   *time.Time `json:"started_at"`   // opsional
	CompletedAt *time.Time `json:"completed_at"` // opsional
}

type OrderSummary struct {
	TotalOrders int64 `json:"total_orders"`
}

type OrderDayHistory struct {
	Date   string               `json:"date"`
	Orders []ActiveServiceOrder `json:"orders"`
}

type EarningMonthGroup struct {
	Month        string         `json:"month"`
	Year         string         `json:"year"`
	MonthlyTotal int64          `json:"monthly_total"`
	Items        []IncomeDetail `json:"items"`
}

type IncomeDetail struct {
	TransactionNo string    `json:"transaction_no"`
	Amount        int64     `json:"amount"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

type EarningMonthlyHistory struct {
	TotalAmount int64               `json:"total_amount"`
	History     []EarningMonthGroup `json:"history"`
}

type CustomerServiceOrder struct {
	ID        int64      `json:"id"`
	Status    string     `json:"status"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	Price     *float64   `json:"price"`

	MitraID   int64  `json:"mitra_id"`
	MitraName string `json:"mitra_name"`

	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type ExpiredOffer struct {
	ID        int64
	RequestID int64
	Sequence  int
}

type OfferAcceptData struct {
	OfferID           int64
	RequestID         int64
	MitraID           int64
	CustomerID        int64
	Keluhan           string
	JobCategoryID     int64
	JobSubCategoryID  *int64
	CustomerName      string
	CustomerPhone     string
	MitraName         string
	MitraPhone        string
	CustomerLatitude  float64
	CustomerLongitude float64
	MitraLatitude     float64
	MitraLongitude    float64
	Price             float64
	VoucherID         *int64
}

type RequestMitraOffer struct {
	ID          int64      `gorm:"column:id;primaryKey"`
	RequestID   int64      `gorm:"column:request_id"`
	MitraID     int64      `gorm:"column:mitra_id"`
	Sequence    int        `gorm:"column:sequence"`
	Status      string     `gorm:"column:status"`
	StatusID    int16      `gorm:"column:status_id"`
	SentAt      *time.Time `gorm:"column:sent_at"`
	ExpiredAt   *time.Time `gorm:"column:expired_at"`
	RespondedAt *time.Time `gorm:"column:responded_at"`
}

func (RequestMitraOffer) TableName() string {
	return "request_mitra_offers"
}

type CancelRequest struct {
	RequestID int64 `json:"request_id"`
}

type ActiveServiceOrder struct {
	ID          int64        `json:"id"`
	StartTime   time.Time    `json:"start_time"`
	StatusID    int16        `json:"status_id"`
	StatusName  string       `json:"code"`
	OrderNumber string       `json:"order_number"`
	Price       float64      `json:"price"` // ✅ NEW
	Keluhan     string       `json:"keluhan"`
	Customer    CustomerInfo `json:"customer"`
	Mitra       MitraInfo    `json:"mitra"`
	CustomerLat float64      `json:"customer_lat"`
	CustomerLng float64      `json:"customer_lng"`
	MitraLat    float64      `json:"mitra_lat"`
	MitraLng    float64      `json:"mitra_lng"`
	CreatedAt   time.Time    `json:"created_at"`
}

type CustomerInfo struct {
	ID    int64  `json:"id"`
	Nama  string `json:"nama"`
	Phone string `json:"phone"`
}

type MitraInfo struct {
	ID    int64  `json:"id"`
	Nama  string `json:"nama"`
	Phone string `json:"phone"`
}

type OrderTransaction struct {
	ID int64 `json:"id"`

	OrderID int64 `json:"order_id"`

	CustomerID   int64  `json:"customer_id"`
	CustomerName string `json:"customer_name"`

	MitraID   int64  `json:"mitra_id"`
	MitraName string `json:"mitra_name"`

	Amount      float64 `json:"amount"`
	Subtotal    float64 `json:"subtotal"`
	Discount    float64 `json:"discount"`
	PlatformFee float64 `json:"platform_fee"`
	MitraIncome float64 `json:"mitra_income"`

	PaymentMethod string     `json:"payment_method"`
	StatusID      int16      `json:"status_id"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`

	VoucherID    *int64   `json:"voucher_id,omitempty"`
	VoucherCode  *string  `json:"voucher_code,omitempty"`
	VoucherValue *float64 `json:"voucher_value,omitempty"`

	OrderCode   string     `json:"order_code"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CompleteOrderRequest struct {
	OrderID       int64    `json:"-"`
	Subtotal      float64  `json:"subtotal"`
	Discount      float64  `json:"discount"`
	PlatformFee   float64  `json:"platform_fee"`
	PaymentMethod string   `json:"payment_method"`
	Note          string   `json:"note"`
	Attachments   []string `json:"attachments"`
}

type ChatMessage struct {
	OrderID    int64     `json:"order_id"`
	SenderID   int64     `json:"sender_id"`
	SenderType string    `json:"sender_type"` // "mitra" or "customer"
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}

type FCMLog struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	Status    string    `json:"status"` // SENT / FAILED
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}
