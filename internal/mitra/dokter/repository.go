package dokter

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"teka-api/internal/models"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

// GLOBAL PARAMETER
// GetGlobalParameter fetch parameter by code
func (r *Repository) GetGlobalParameter(ctx context.Context, code string) (string, error) {
	var value string
	query := `SELECT parameter_value FROM global_parameter WHERE parameter_code = $1 AND is_active =  'true' LIMIT 1`
	err := r.DB.Raw(query, code).Scan(&value).Error
	if err != nil {
		return "", fmt.Errorf("failed to get global parameter %s: %w", code, err)
	}
	return value, nil
}

// GetMaxRadius khusus untuk MAX_RADIUS
func (r *Repository) GetMaxRadius(ctx context.Context) (float64, error) {
	val, err := r.GetGlobalParameter(ctx, "MAX_RADIUS")
	if err != nil {
		return 0, err
	}

	radius, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid MAX_RADIUS value: %w", err)
	}
	return radius, nil
}

// START INSERT MITRA
func (r *Repository) CreateUserMitra(userID int, jobCatID int, jobSubCatID *int) error {

	query := `
		INSERT INTO user_mitra
			(user_id, job_category_id, job_sub_category_id)
		VALUES
			($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE
		SET
			job_category_id     = EXCLUDED.job_category_id,
			job_sub_category_id = EXCLUDED.job_sub_category_id;
	`

	return r.DB.Exec(query, userID, jobCatID, jobSubCatID).Error
}

// END INSERT MITRA

// START INSERT MITRA DETAIL
func (r *Repository) CreateMitraDetail(userID int, jobCatID int, jobSubCatID *int, lat, lng float64) error {
	query := `
		INSERT INTO mitra_details
			(user_id, job_category_id, job_sub_category_id, latitude, longitude, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE
		SET
			job_category_id     = EXCLUDED.job_category_id,
			job_sub_category_id = EXCLUDED.job_sub_category_id,
			latitude            = EXCLUDED.latitude,
			longitude           = EXCLUDED.longitude,
			updated_at          = NOW();
	`

	return r.DB.Exec(query, userID, jobCatID, jobSubCatID, lat, lng).Error
}

// END INSERT MITRA DETAIL

// START INSERT USER ROLES
func (r *Repository) CreateUserRole(userID, roleID int, active bool) error {
	query := `
        INSERT INTO user_roles (user_id, role_id, active, created_at, updated_at)
        VALUES (?, ?, ?, NOW(), NOW())
        ON CONFLICT (user_id, role_id) DO NOTHING;
    `
	return r.DB.Exec(query, userID, roleID, active).Error
}

// END INSERT USER ROLES

// START INSERT DOKUMEN
func (r *Repository) CreateDocument(userID int, docType, url string) error {
	query := `
        INSERT INTO mitra_documents (user_id, doc_type, file_url, created_at, updated_at)
        VALUES (?,?,?,NOW(),NOW());
    `
	return r.DB.Exec(query, userID, docType, url).Error
}

// END INSERT DOKUMEN

// START CARI DOKTER DALAM RADIUS
func (r *Repository) SearchDoctors(
	ctx context.Context,
	jobCategoryID int64,
	jobSubCategoryID int64,
	lat, lng float64,
) ([]models.DoctorSearchResult, error) {

	var doctors []models.DoctorSearchResult

	maxRadius, err := r.GetMaxRadius(ctx)
	if err != nil {
		maxRadius = 10
	}

	query := `
	SELECT *
FROM (
	SELECT
		u.id AS mitra_id,
		u.nama AS nama,
		md.latitude,
		md.longitude,
		0 AS avg_rating,
		0 AS total_review,
		(
			6371 * acos(
				cos(radians(?)) *
				cos(radians(md.latitude)) *
				cos(radians(md.longitude) - radians(?)) +
				sin(radians(?)) *
				sin(radians(md.latitude))
			)
		) AS distance_km
	FROM mitra_details md
	JOIN users u ON u.id = md.user_id
	JOIN user_roles ur ON ur.user_id = u.id
	WHERE md.job_category_id = ?
	  AND (? = 0 OR md.job_sub_category_id = ?)
	  AND md.availability_status_id = 2
	  AND ur.role_id = 2
	  AND ur.active = true

	  AND NOT EXISTS (
		  SELECT 1
		  FROM request_mitra_offers rmo
		  WHERE rmo.mitra_id = u.id
		    AND rmo.status_id = 1
	  )

	  AND NOT EXISTS (
		  SELECT 1
		  FROM service_orders so
		  WHERE so.mitra_id = u.id
		    AND so.status_id IN (1,2,3)
	  )
) t
WHERE t.distance_km <= ?
ORDER BY t.distance_km ASC

`

	err = r.DB.WithContext(ctx).
		Raw(
			query,
			lat, lng, lat,
			jobCategoryID,
			jobSubCategoryID,
			jobSubCategoryID,
			maxRadius,
		).
		Scan(&doctors).Error

	if err != nil {
		return nil, err
	}

	return doctors, nil
}

// END CARI DOKTER DALAM RADIUS

// START CUSTOMER REQUEST
func (r *Repository) CreateCustomerRequest(
	ctx context.Context,
	req models.CustomerRequest,
) (int64, error) {

	var id int64

	query := `
		INSERT INTO customer_requests (
			customer_id, job_category_id, job_sub_category_id, 
			keluhan, latitude, longitude, radius, price, 
			voucher_id, voucher_value, platform_fee, thr_bonus,
			status_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 1, NOW(), NOW())
		RETURNING id
	`

	err := r.DB.WithContext(ctx).Raw(
		query,
		req.CustomerID,
		req.JobCategoryID,
		req.JobSubCategoryID,
		req.Keluhan,
		req.Latitude,
		req.Longitude,
		req.Radius,
		req.Price,
		req.VoucherID,
		req.VoucherValue,
		req.PlatformFee,
		req.THRBonus,
	).Scan(&id).Error

	return id, err
}

// Ambil customer request
func (r *Repository) GetCustomerRequest(ctx context.Context, requestID int64) (models.CustomerRequest, error) {
	var req models.CustomerRequest
	err := r.DB.WithContext(ctx).
		Raw(`SELECT * FROM customer_requests WHERE id = ?`, requestID).
		Scan(&req).Error
	return req, err
}

// -------------------------------
// CANCEL ORDERAN
// -------------------------------
func (r *Repository) CancelCustomerRequest(ctx context.Context, requestID int64) error {
	tx := r.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Update customer request → hanya yang belum diterima mitra (status_id != 3 / accepted)
	if err := tx.Exec(`
        UPDATE customer_requests
        SET status_id = ?, updated_at = NOW()
        WHERE id = ? AND status_id NOT IN (3)  -- 3 = accepted/matched
    `, 4, requestID).Error; err != nil { // 4 = cancelled
		tx.Rollback()
		return err
	}

	// 2. Update semua offer yang masih waiting (status_id = 1) menjadi cancelled (status_id = 5)
	if err := tx.Exec(`
        UPDATE request_mitra_offers
        SET status_id = ?, responded_at = NOW()
        WHERE request_id = ? AND status_id = 1
    `, 5, requestID).Error; err != nil { // 5 = cancelled sistem
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// START DETAIL CUSTOMER DI DOKTER
func (r *Repository) GetCurrentOfferForMitra(
	ctx context.Context,
	mitraID int64,
) (*models.CurrentOffer, error) {

	var offer models.CurrentOffer

	err := r.DB.WithContext(ctx).Raw(`
		SELECT
			o.id           AS offer_id,
			o.request_id,
			r.customer_id,
			r.keluhan,
			r.latitude,
			r.longitude,
			r.price,
			r.created_at,
			u.nama         AS customer_name,
			u.phone        AS customer_phone
		FROM request_mitra_offers o
		JOIN customer_requests r ON r.id = o.request_id
		JOIN users u ON u.id = r.customer_id
		WHERE o.mitra_id = ?
		  AND o.status_id = 1
		ORDER BY o.sent_at
		
	`, mitraID).Scan(&offer).Error

	if err != nil {
		return nil, err
	}

	if offer.OfferID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &offer, nil
}

// END DETAIL CUSTOMER DI DOKTER

// START CREATE OFFER
func (r *Repository) CreateOffers(
	ctx context.Context,
	requestID int64,
	doctors []models.DoctorSearchResult,
) error {

	tx := r.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for i, d := range doctors {
		var statusID int16
		var sentAt interface{}

		if i == 0 {
			// offer pertama langsung dikirim
			statusID = 1 // waiting
			sentAt = gorm.Expr("NOW()")
		} else {
			// sisanya ANTRI
			statusID = 6 // pending
			sentAt = nil
		}

		if err := tx.Exec(`
			INSERT INTO request_mitra_offers (
				request_id,
				mitra_id,
				sequence,
				status_id,
				sent_at
			)
			VALUES (?, ?, ?, ?, ?)
		`,
			requestID,
			d.MitraID,
			i+1,
			statusID,
			sentAt,
		).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// END CREATE OFFER

// GET OFFER
func (r *Repository) GetOfferForAccept(
	ctx context.Context,
	offerID int64,
	mitraID int64,
) (*models.OfferAcceptData, error) {

	var o models.OfferAcceptData

	err := r.DB.WithContext(ctx).Raw(`
	SELECT
		o.id AS offer_id,
		o.request_id,
		o.mitra_id,

		r.customer_id,
		r.keluhan,
		r.job_category_id,
		r.job_sub_category_id,

		c.nama  AS customer_name,
		c.phone AS customer_phone,

		m.nama  AS mitra_name,
		m.phone AS mitra_phone,

		r.latitude  AS customer_latitude,
		r.longitude AS customer_longitude,

		md.latitude AS mitra_latitude,
		md.longitude AS mitra_longitude,

		r.price,
		r.voucher_id,
		r.platform_fee,
		r.thr_bonus,
		r.voucher_value
	FROM request_mitra_offers o
	JOIN customer_requests r ON r.id = o.request_id
	JOIN users c ON c.id = r.customer_id
	JOIN users m ON m.id = o.mitra_id
	JOIN mitra_details md ON md.user_id = o.mitra_id
	WHERE o.id = ?
	  AND o.mitra_id = ?
	  AND o.status_id = 1
`, offerID, mitraID).Scan(&o).Error

	if err != nil {
		return nil, err
	}

	if o.OfferID == 0 {
		return nil, errors.New("Maaf, waktu penawaran sudah habis")
	}

	return &o, nil
}

// ACCEPT OFFER
func (r *Repository) AcceptOfferAndCreateOrder(
	ctx context.Context,
	o *models.OfferAcceptData,
) error {

	tx := r.DB.WithContext(ctx).Begin()

	// 1️⃣ Accept offer
	res := tx.Exec(`
		UPDATE request_mitra_offers
		SET status_id = 2,
		    responded_at = NOW()
		WHERE id = ? AND status_id = 1
	`, o.OfferID)
	if res.RowsAffected == 0 {
		tx.Rollback()
		return errors.New("Maaf, waktu penawaran sudah habis")
	}

	// 2️⃣ Cancel other offers
	if err := tx.Exec(`
		UPDATE request_mitra_offers
		SET status_id = 5
		WHERE request_id = ? AND id != ? AND status_id IN (1,6)
	`, o.RequestID, o.OfferID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 3️⃣ Update customer request
	if err := tx.Exec(`
		UPDATE customer_requests
		SET status_id = 2,
		    matched_mitra_id = ?,
		    matched_at = NOW()
		WHERE id = ? AND status_id = 1
	`, o.MitraID, o.RequestID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 4️⃣ INSERT service_order + ambil id & created_at
	var orderID int64
	var createdAt time.Time

	if err := tx.Raw(`
		INSERT INTO service_orders (
			request_id,
			customer_id,
			mitra_id,

			customer_name,
			customer_phone,
			mitra_name,
			mitra_phone,

			job_category_id,
			job_sub_category_id,
			keluhan,

			customer_latitude,
			customer_longitude,
			mitra_latitude,
			mitra_longitude,

			price,
			voucher_id,
			platform_fee,
			thr_bonus,
			voucher_value,

			status_id,
			start_time,
			accepted_at,
			created_at,
			updated_at
		)
		VALUES (
			?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?,
			1,
			NOW(), NOW(), NOW(), NOW()
		)
		RETURNING id, created_at
	`,
		o.RequestID,
		o.CustomerID,
		o.MitraID,

		o.CustomerName,
		o.CustomerPhone,
		o.MitraName,
		o.MitraPhone,

		o.JobCategoryID,
		o.JobSubCategoryID,
		o.Keluhan,

		o.CustomerLatitude,
		o.CustomerLongitude,
		o.MitraLatitude,
		o.MitraLongitude,

		o.Price,
		o.VoucherID,
		o.PlatformFee,
		o.THRBonus,
		o.VoucherValue,
	).Row().Scan(&orderID, &createdAt); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Exec(`
	UPDATE service_orders
SET order_number = format(
	'DOK-%s-%s',
	to_char(created_at, 'YYYYMMDDHH24MISSMS'),
	to_char(nextval('myschema.service_order_seq'), 'FM000000')
)
WHERE id = ?;
`, orderID).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// AKTIF ORDER FOR CUSTOMER
func (r *Repository) GetActiveServiceOrderForCustomer(
	ctx context.Context,
	customerID int64,
) (*models.ActiveServiceOrder, error) {

	var row activeServiceOrderRow

	err := r.DB.WithContext(ctx).Raw(`
	SELECT
		so.id,
		so.start_time,
		so.status_id,
		so.order_number,
		sos.code AS status_name,
		so.price,
		so.keluhan,
		so.customer_id,
		so.customer_name  AS customer_nama,
		so.customer_phone AS customer_phone,
		so.mitra_id,
		so.mitra_name     AS mitra_nama,
		so.mitra_phone    AS mitra_phone,
		so.customer_latitude  AS customer_lat,
		so.customer_longitude AS customer_lng,
		so.mitra_latitude     AS mitra_lat,
		so.mitra_longitude    AS mitra_lng
	FROM service_orders so
	JOIN service_order_statuses sos 
		ON sos.id = so.status_id
	WHERE so.customer_id = ?
	  AND so.status_id NOT IN (4,5,6)
	LIMIT 1
`, customerID).Scan(&row).Error

	if err != nil || row.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return mapActiveServiceOrder(row), nil
}

// AKTIF ORDER FOR MITRA
func (r *Repository) GetActiveServiceOrderForMitra(
	ctx context.Context,
	mitraID int64,
) (*models.ActiveServiceOrder, error) {

	var row activeServiceOrderRow

	err := r.DB.WithContext(ctx).Raw(`
	SELECT
		so.id,
		so.start_time,
		so.status_id,
		so.order_number,
		sos.code AS status_name,
		so.price,
		so.keluhan,
		so.customer_id,
		so.customer_name  AS customer_nama,
		so.customer_phone AS customer_phone,
		so.mitra_id,
		so.mitra_name     AS mitra_nama,
		so.mitra_phone    AS mitra_phone,
		so.customer_latitude  AS customer_lat,
		so.customer_longitude AS customer_lng,
		so.mitra_latitude     AS mitra_lat,
		so.mitra_longitude    AS mitra_lng
	FROM service_orders so
	JOIN service_order_statuses sos 
		ON sos.id = so.status_id
	WHERE so.mitra_id = ?
	  AND so.status_id NOT IN (4,5,6)
	LIMIT 1
`, mitraID).Scan(&row).Error

	if err != nil || row.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return mapActiveServiceOrder(row), nil
}

// COMPLETE ORDER
func (r *Repository) CompleteOrder(
	ctx context.Context,
	orderID int64,
	mitraID int64,
	data CompleteOrderTxData,
) error {

	tx := r.DB.WithContext(ctx).Begin()
	now := time.Now()

	res := tx.Exec(`
		UPDATE service_orders
		SET status_id = 4,          -- COMPLETED
		    end_time = ?,
		    updated_at = ?
		WHERE id = ?
		  AND mitra_id = ?
		  AND status_id = 3         -- ARRIVED
	`, now, now, orderID, mitraID)

	if res.Error != nil {
		tx.Rollback()
		return res.Error
	}

	if res.RowsAffected == 0 {
		tx.Rollback()
		return errors.New("order not valid or already completed")
	}

	// 2️⃣ update service order
	if err := tx.Exec(`
		UPDATE service_orders
		SET status_id = 3,
		    end_time = ?,
		    updated_at = ?
		WHERE id = ?
	`, now, now, orderID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 3️⃣ transaksi + NOTE
	// 3️⃣ insert transaction snapshot dengan note, customer/mitra info, voucher
	if err := tx.Exec(`
	INSERT INTO order_transactions (
		order_id,
		customer_id,
		customer_name,
		mitra_id,
		mitra_name,
		platform_fee,
		mitra_income,
		payment_method,
		voucher_id,
		voucher_value,
		thr_bonus,
		order_number,
		completed_at,
		note,
		status_id,
		paid_at,
		created_at,
		updated_at
	)
	SELECT
		so.id,
		so.customer_id,
		so.customer_name,
		so.mitra_id,
		so.mitra_name,
		so.platform_fee, 
		so.price, -- mitra_income
		'wallet', 
		so.voucher_id, 
		so.voucher_value, 
		so.thr_bonus, 
		so.order_number, 
		NOW(), 
		?, 
		1, 
		?, ?, ?
	FROM service_orders so
	WHERE so.id = ?
`,
		data.Note, // note dokter
		now,
		now,
		now,
		orderID,
	).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 4️⃣ foto bukti
	for _, url := range data.Attachments {
		if err := tx.Exec(`
			INSERT INTO order_attachments (order_id, type, url)
			VALUES (?, 'photo', ?)
		`, orderID, url).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// CompleteOrderCustomer (Complete by Customer)
func (r *Repository) CompleteOrderCustomer(
	ctx context.Context,
	orderID int64,
	customerID int64,
) error {
	// Hanya update status service_order menjadi 6
	// Status sebelumnya harus 4 (Completed by Doctor) atau bisa juga still in progress jika flow mengizinkan
	// Sesuai request: "di user hanya update status menjadi 6 di service order"

	query := `
		UPDATE service_orders
		SET status_id = 6,          -- COMPLETED BY USER / FINISHED
		    updated_at = NOW()
		WHERE id = ?
		  AND customer_id = ?
		  AND status_id IN (2, 3, 4)   -- ACCEPTED, ARRIVED, or COMPLETED_BY_DOCTOR
	`
	res := r.DB.WithContext(ctx).Exec(query, orderID, customerID)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return errors.New("order not found or not authorized")
	}

	return nil
}

func (r *Repository) RejectOffer(
	ctx context.Context,
	offerID int64,
) error {

	tx := r.DB.WithContext(ctx).Begin()

	// 1. reject offer sekarang
	if err := tx.Exec(`
		UPDATE request_mitra_offers
		SET status = 'rejected',
		    status_id = 3,
		    responded_at = now()
		WHERE id = ?
		  AND status = 'waiting'
	`, offerID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2. aktifkan offer berikutnya
	if err := tx.Exec(`
		UPDATE request_mitra_offers
		SET status = 'waiting',
		    status_id = 1,
		    sent_at = now()
		WHERE id = (
			SELECT id FROM request_mitra_offers
			WHERE request_id = (
				SELECT request_id FROM request_mitra_offers WHERE id = ?
			)
			AND status = 'cancelled'
			ORDER BY sequence
			LIMIT 1
		)
	`, offerID).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *Repository) GetExpiredOffers(
	ctx context.Context,
	timeoutSeconds int,
) ([]models.ExpiredOffer, error) {

	var offers []models.ExpiredOffer

	err := r.DB.WithContext(ctx).Raw(`
		SELECT o.id, o.request_id, o.sequence, u.nama as mitra_name
		FROM request_mitra_offers o
		JOIN users u ON u.id = o.mitra_id
		WHERE o.status_id = 1
		  AND o.sent_at IS NOT NULL
		  AND o.sent_at <= clock_timestamp() - (? * INTERVAL '1 second')
	`, timeoutSeconds).Scan(&offers).Error

	return offers, err
}

func (r *Repository) TimeoutAndMoveNext(
	ctx context.Context,
	requestID int64,
	sequence int,
) (int64, string, error) {

	tx := r.DB.WithContext(ctx).Begin()

	// 1. timeout current offer
	if err := tx.Exec(`
		UPDATE request_mitra_offers
		SET status_id = 4,
		    responded_at = NOW()
		WHERE request_id = ?
		  AND sequence = ?
		  AND status_id = 1
	`, requestID, sequence).Error; err != nil {
		tx.Rollback()
		return 0, "", err
	}

	// 2. Cari next mitra_id & name
	var next struct {
		MitraID int64
		Nama    string `gorm:"column:nama"`
	}
	err := tx.Raw(`
		SELECT o.mitra_id, u.nama
		FROM request_mitra_offers o
		JOIN users u ON u.id = o.mitra_id
		WHERE o.request_id = ? AND o.sequence = ? AND o.status_id = 6
		LIMIT 1
	`, requestID, sequence+1).Scan(&next).Error

	if err != nil || next.MitraID == 0 {
		// No more offers in queue
		tx.Commit()
		return 0, "", nil
	}

	// 3. activate next offer (ONLY ONE)
	if err := tx.Exec(`
		UPDATE request_mitra_offers
		SET status_id = 1,
		    sent_at = NOW()
		WHERE request_id = ?
		  AND sequence = ?
		  AND status_id = 6
	`, requestID, sequence+1).Error; err != nil {
		tx.Rollback()
		return 0, "", err
	}

	if err := tx.Commit().Error; err != nil {
		return 0, "", err
	}

	return next.MitraID, next.Nama, nil
}

func (r *Repository) GetServiceForRating(
	ctx context.Context,
	serviceOrderID, customerID int64,
) (int64, error) {

	var mitraID int64

	err := r.DB.WithContext(ctx).Raw(`
		SELECT mitra_id
		FROM service_orders
		WHERE id = ?
		  AND customer_id = ?
		  AND status_id = 6
	`, serviceOrderID, customerID).Scan(&mitraID).Error

	return mitraID, err
}

func (r *Repository) InsertRating(
	ctx context.Context,
	serviceOrderID, mitraID, customerID int64,
	rating int,
	review string,
) error {

	return r.DB.WithContext(ctx).Exec(`
		INSERT INTO mitra_ratings
			(service_order_id, mitra_id, customer_id, rating, review)
		VALUES (?, ?, ?, ?, ?)
	`,
		serviceOrderID,
		mitraID,
		customerID,
		rating,
		review,
	).Error
}

// Ambil offer waiting untuk dokter tertentu (urut sequence)
func (r *Repository) GetNextOfferForMitra(ctx context.Context, mitraID int64) (models.MitraOffer, error) {
	var offer models.MitraOffer
	err := r.DB.WithContext(ctx).Raw(`
		SELECT id, request_id, mitra_id, sequence, status
		FROM request_mitra_offers
		WHERE mitra_id = ? AND status = 'waiting'
		ORDER BY sequence ASC
		LIMIT 1
	`, mitraID).Scan(&offer).Error

	fmt.Printf("Repo GetNextOfferForMitra result: %+v\n", offer)
	return offer, err
}

// close
func (r *Repository) CompleteServiceOrder(
	ctx context.Context,
	serviceOrderID, mitraID int64,
) error {

	res := r.DB.WithContext(ctx).Exec(`
		UPDATE service_orders
		SET
			service_status = 'completed',
			end_time = NOW()
		WHERE
			id = ?
			AND mitra_id = ?
			AND service_status = 'ongoing'
	`, serviceOrderID, mitraID)

	if res.RowsAffected == 0 {
		return errors.New("service order not found or already completed")
	}

	return res.Error
}

func (r *Repository) GetCustomerServiceOrders(
	ctx context.Context,
	customerID int64,
) ([]models.CustomerServiceOrder, error) {

	var orders []models.CustomerServiceOrder

	err := r.DB.WithContext(ctx).Raw(`
		SELECT
			so.id,
			so.service_status AS status,
			so.start_time,
			so.end_time,
			so.price,

			u.id AS mitra_id,
			u.nama AS mitra_name,

			cr.latitude,
			cr.longitude
		FROM service_orders so
		JOIN users u ON u.id = so.mitra_id
		JOIN customer_requests cr ON cr.id = so.request_id
		WHERE so.customer_id = ?
		ORDER BY so.created_at DESC
	`, customerID).Scan(&orders).Error

	return orders, err
}

// Ambil voucher aktif customer
func (r *Repository) GetActiveVoucher(ctx context.Context, userID int64) ([]models.Voucher, error) {
	var vouchers []models.Voucher

	query := `
	SELECT v.id, v.code, v.description
	FROM dev.voucher_user vu
	JOIN dev.vouchers v ON v.id = vu.voucher_id
	WHERE vu.user_id = ?
	  AND vu.used_at IS NULL
	  AND v.is_active = TRUE
	  AND v.start_date <= CURRENT_DATE
	  AND v.end_date >= CURRENT_DATE
	`

	if err := r.DB.WithContext(ctx).Raw(query, userID).Scan(&vouchers).Error; err != nil {
		return nil, err
	}

	return vouchers, nil
}

func (r *Repository) GetServiceOrderStatus(
	ctx context.Context,
	orderID int,
) (int16, error) {

	var statusID int16
	err := r.DB.
		WithContext(ctx).
		Table("service_orders").
		Select("status_id").
		Where("id = ?", orderID).
		Scan(&statusID).Error

	if err != nil {
		return 0, err
	}

	if statusID == 0 {
		return 0, gorm.ErrRecordNotFound
	}

	return statusID, nil
}

func (r *Repository) ServiceOrderStatusExists(
	ctx context.Context,
	statusID int16,
) (bool, error) {

	var exists bool
	err := r.DB.
		WithContext(ctx).
		Raw(`
			SELECT EXISTS (
				SELECT 1 FROM service_order_statuses WHERE id = ?
			)
		`, statusID).
		Scan(&exists).Error

	return exists, err
}

func (r *Repository) UpdateServiceOrderStatus(
	ctx context.Context,
	orderID int,
	fromStatus int16,
	toStatus int16,
) error {

	tx := r.DB.WithContext(ctx).
		Table("service_orders").
		Where("id = ? AND status_id = ?", orderID, fromStatus).
		Updates(map[string]interface{}{
			"status_id":  toStatus,
			"updated_at": time.Now(),
		})

	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		return errors.New("status not updated")
	}

	return nil
}

func (r *Repository) GetFCMTokensByUserID(
	ctx context.Context,
	userID int64,
) ([]string, error) {

	var tokens []string

	err := r.DB.WithContext(ctx).
		Table("user_fcm_tokens").
		Where("user_id = ?", userID).
		Pluck("fcm_token", &tokens).Error

	return tokens, err
}

func (r *Repository) LogFCMResult(userID int64, token string, status string, errStr *string) error {
	logEntry := models.FCMLog{
		UserID: userID,
		Token:  token,
		Status: status,
	}
	if errStr != nil {
		logEntry.Error = *errStr
	}

	return r.DB.Create(&logEntry).Error
}

// DeductCustomerBalance deducts amount from user's balance using saldo_role_transactions
func (r *Repository) DeductCustomerBalance(
	ctx context.Context,
	customerID int64,
	orderID int64,
) error {
	log.Printf("[DeductCustomerBalance] Starting for customerID: %d, orderID: %d", customerID, orderID)

	tx := r.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[DeductCustomerBalance] Recovered from panic: %v", r)
			tx.Rollback()
		}
	}()

	// 1. Get order info and amount
	var orderNo string
	var amount float64

	var err error
	if err = tx.Raw(`
		SELECT order_number, (mitra_income + platform_fee + thr_bonus - voucher_value) 
		FROM myschema.order_transactions 
		WHERE order_id = ?
	`, orderID).Row().Scan(&orderNo, &amount); err != nil {
		log.Printf("[DeductCustomerBalance] Error fetching order info: %v", err)
		tx.Rollback()
		return err
	}
	log.Printf("[DeductCustomerBalance] Order number: %s, Amount: %.2f", orderNo, amount)

	// 2. Get latest balance from saldo_role_transactions (Role ID 1 = Customer)
	var latestSaldo int64

	// Lock row terakhir balance user tersebut untuk menghindari race condition
	err = tx.Raw(`
		SELECT saldo_setelah 
		FROM myschema.saldo_role_transactions 
		WHERE user_id = ? 
		ORDER BY id DESC 
		LIMIT 1
		FOR UPDATE
	`, customerID).Scan(&latestSaldo).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("[DeductCustomerBalance] Error fetching latest balance: %v", err)
		tx.Rollback()
		return err
	}
	log.Printf("[DeductCustomerBalance] Latest saldo: %d", latestSaldo)

	// 2. Check sufficiency
	intAmount := int64(amount)
	if latestSaldo < intAmount {
		log.Printf("[DeductCustomerBalance] Insufficient balance. Have: %d, Need: %d", latestSaldo, intAmount)
		tx.Rollback()
		return fmt.Errorf("saldo tidak mencukupi. Saldo saat ini: %d, Dibutuhkan: %d", latestSaldo, intAmount)
	}

	// 3. Insert mutation ke saldo_role_transactions
	newSaldo := latestSaldo - intAmount

	trx := &models.SaldoTransaction{
		UserID:        uint(customerID),
		ReferenceID:   orderNo,
		ReferenceType: "ORDER",
		MutationType:  "OUT",
		CategoryID:    2, // Order Payment
		Amount:        intAmount,
		SaldoSetelah:  newSaldo,
		Description:   "Pembayaran order Dokter",
		CreatedAt:     time.Now(),
	}

	if err := tx.Create(trx).Error; err != nil {
		tx.Rollback()
		return err
	}
	log.Printf("[DeductCustomerBalance] Mutation inserted. New saldo: %d", newSaldo)

	// 4. Update order_transactions status to PAID (status_id = 2)
	resTrx := tx.Exec(`
		UPDATE myschema.order_transactions 
		SET status_id = 2, paid_at = NOW(), updated_at = NOW() 
		WHERE order_id = ?
	`, orderID)
	if resTrx.Error != nil {
		log.Printf("[DeductCustomerBalance] Error updating order_transactions: %v", resTrx.Error)
		tx.Rollback()
		return resTrx.Error
	}
	log.Printf("[DeductCustomerBalance] order_transactions updated. Rows affected: %d", resTrx.RowsAffected)

	// 5. Update service order status to 6 (FINISHED)
	res := tx.Exec(`
		UPDATE myschema.service_orders
		SET status_id = 6, updated_at = NOW()
		WHERE id = ? AND customer_id = ? AND status_id = 4
	`, orderID, customerID)

	if res.Error != nil {
		log.Printf("[DeductCustomerBalance] Error updating service_orders: %v", res.Error)
		tx.Rollback()
		return res.Error
	}

	log.Printf("[DeductCustomerBalance] service_orders updated. Rows affected: %d", res.RowsAffected)

	if res.RowsAffected == 0 {
		// Log detail tambahan kalau gagal
		var currentStatus int16
		tx.Raw(`SELECT status_id FROM myschema.service_orders WHERE id = ?`, orderID).Scan(&currentStatus)
		log.Printf("[DeductCustomerBalance] Details: orderID=%d, customerID=%d, currentStatusInDB=%d", orderID, customerID, currentStatus)

		tx.Rollback()
		return errors.New("order tidak ditemukan atau status tidak valid untuk diselesaikan")
	}

	// 6. Record Mitra Income (Simplified)
	var trans struct {
		MitraID     int64
		MitraIncome float64
	}
	err = tx.Raw(`
		SELECT mitra_id, mitra_income 
		FROM myschema.order_transactions 
		WHERE order_id = ?
	`, orderID).Scan(&trans).Error
	if err != nil {
		log.Printf("[DeductCustomerBalance] Error fetching order_transaction: %v", err)
		tx.Rollback()
		return err
	}

	// Lock row terakhir balance mitra
	var mitraLatestSaldo int64
	err = tx.Raw(`
		SELECT saldo_setelah 
		FROM myschema.saldo_role_transactions 
		WHERE user_id = ? 
		ORDER BY id DESC 
		LIMIT 1
		FOR UPDATE
	`, trans.MitraID).Scan(&mitraLatestSaldo).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("[DeductCustomerBalance] Error fetching mitra balance: %v", err)
		tx.Rollback()
		return err
	}

	mitraIncomeInt := int64(trans.MitraIncome)
	// Jika mitra_income 0 (misal belum terisi), fallback ke amount (bruto)
	if mitraIncomeInt == 0 {
		mitraIncomeInt = intAmount
	}
	mitraNewSaldo := mitraLatestSaldo + mitraIncomeInt

	if err := tx.Create(&models.SaldoTransaction{
		UserID:        uint(trans.MitraID),
		ReferenceID:   orderNo,
		ReferenceType: "ORDER",
		MutationType:  "IN",
		CategoryID:    3, // Pendapatan
		Amount:        mitraIncomeInt,
		SaldoSetelah:  mitraNewSaldo,
		Description:   "Pendapatan order Dokter",
		CreatedAt:     time.Now(),
	}).Error; err != nil {
		log.Printf("[DeductCustomerBalance] Error inserting mitra income: %v", err)
		tx.Rollback()
		return err
	}
	log.Printf("[DeductCustomerBalance] Mitra income recorded. New saldo: %d", mitraNewSaldo)

	log.Printf("[DeductCustomerBalance] Success. Committing transaction.")
	return tx.Commit().Error
}

// GetLatestBalance returns the latest balance for a user and role
func (r *Repository) GetLatestBalance(ctx context.Context, userID int64) (int64, error) {
	var balance int64
	err := r.DB.WithContext(ctx).Raw(`
		SELECT saldo_setelah 
		FROM myschema.saldo_role_transactions 
		WHERE user_id = ? ORDER BY id DESC LIMIT 1
	`, userID).Scan(&balance).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return balance, nil
}

// RequestWithdrawal processes a withdrawal request for a mitra
func (r *Repository) RequestWithdrawal(
	ctx context.Context,
	mitraID int64,
	amount int64,
	bankName, accountNumber, accountHolder string,
) error {
	tx := r.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Get current balance (Role Mitra = 2)
	var latestSaldo int64
	err := tx.Raw(`
		SELECT saldo_setelah 
		FROM myschema.saldo_role_transactions 
		WHERE user_id = ? ORDER BY id DESC LIMIT 1
		FOR UPDATE
	`, mitraID).Scan(&latestSaldo).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return err
	}

	// 2. Check sufficiency
	if latestSaldo < amount {
		tx.Rollback()
		return fmt.Errorf("saldo tidak mencukupi. Saldo saat ini: %d, Dibutuhkan: %d", latestSaldo, amount)
	}

	// 3. Generate transaction no
	transactionNo := fmt.Sprintf("WD-%d-%d", mitraID, time.Now().UnixNano())

	// 4. Insert mutation
	newSaldo := latestSaldo - amount
	if err := tx.Exec(`
		INSERT INTO myschema.saldo_role_transactions (
			user_id, reference_id, reference_type, mutation_type, category_id, amount, saldo_setelah, description, created_at
		) VALUES (?, ?, 'WITHDRAWAL', 'OUT', 4, ?, ?, ?, NOW())
	`,
		mitraID,
		transactionNo,
		amount,
		newSaldo,
		fmt.Sprintf("Penarikan saldo ke %s (%s) a/n %s", bankName, accountNumber, accountHolder),
	).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *Repository) GetMitraRatingSummary(ctx context.Context, mitraID int64) (models.RatingSummary, error) {
	var summary models.RatingSummary

	err := r.DB.WithContext(ctx).Raw(`
		SELECT 
			COUNT(*) as total_ratings, 
			COALESCE(AVG(rating), 0) as average_rating 
		FROM myschema.mitra_ratings 
		WHERE mitra_id = ?
	`, mitraID).Scan(&summary).Error

	return summary, err
}

func (r *Repository) GetMitraRatingHistory(ctx context.Context, mitraID int64) ([]models.MonthlyRatingHistory, error) {
	var results []models.MitraRating

	// Fetch all ratings for history (removed current month filter)
	err := r.DB.WithContext(ctx).Raw(`
		SELECT 
			mr.id, mr.service_order_id, mr.mitra_id, mr.customer_id, 
			so.customer_name,
			mr.rating, mr.review, mr.created_at 
		FROM myschema.mitra_ratings mr
		JOIN myschema.service_orders so ON so.id = mr.service_order_id
		WHERE mr.mitra_id = ? 
		ORDER BY mr.created_at DESC
	`, mitraID).Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Group by month and year in Go
	type groupKey struct {
		Month string
		Year  string
	}
	historyMap := make(map[groupKey][]models.MitraRating)
	var keys []groupKey

	for _, res := range results {
		// res.CreatedAt format: 2026-02-17T14:22:35.605869Z
		t, err := time.Parse(time.RFC3339, res.CreatedAt)
		if err != nil {
			// Fallback to simple extraction if parsing fails
			if len(res.CreatedAt) >= 7 {
				year := res.CreatedAt[0:4]
				month := res.CreatedAt[5:7]
				key := groupKey{Month: month, Year: year}
				if _, exists := historyMap[key]; !exists {
					keys = append(keys, key)
				}
				historyMap[key] = append(historyMap[key], res)
			}
			continue
		}

		month := t.Format("01")
		year := t.Format("2006")
		key := groupKey{Month: month, Year: year}
		if _, exists := historyMap[key]; !exists {
			keys = append(keys, key)
		}
		historyMap[key] = append(historyMap[key], res)
	}

	var history []models.MonthlyRatingHistory
	for _, key := range keys {
		history = append(history, models.MonthlyRatingHistory{
			Month:   key.Month,
			Year:    key.Year,
			Ratings: historyMap[key],
		})
	}

	return history, nil
}

func (r *Repository) GetMitraOrderSummary(ctx context.Context, mitraID int64) (models.OrderSummary, error) {
	var summary models.OrderSummary
	err := r.DB.WithContext(ctx).Raw(`
		SELECT COUNT(*) as total_orders 
		FROM myschema.service_orders 
		WHERE mitra_id = ? AND status_id = 6
	`, mitraID).Scan(&summary).Error
	return summary, err
}

func (r *Repository) GetMitraOrderHistory(ctx context.Context, mitraID int64) ([]models.MonthlyOrderHistory, error) {
	var rows []activeServiceOrderRow

	err := r.DB.WithContext(ctx).Raw(`
		SELECT
			so.id,
			so.start_time,
			so.status_id,
			so.order_number,
			sos.code AS status_name,
			so.price,
			so.keluhan,
			so.customer_id,
			so.customer_name  AS customer_nama,
			so.customer_phone AS customer_phone,
			so.mitra_id,
			so.mitra_name     AS mitra_nama,
			so.mitra_phone    AS mitra_phone,
			so.customer_latitude  AS customer_lat,
			so.customer_longitude AS customer_lng,
			so.mitra_latitude     AS mitra_lat,
			so.mitra_longitude    AS mitra_lng,
			so.created_at
		FROM service_orders so
		JOIN service_order_statuses sos 
			ON sos.id = so.status_id
		WHERE so.mitra_id = ? 
		  AND so.status_id = 6
		ORDER BY so.created_at DESC
	`, mitraID).Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	// Group by month and year in Go
	type groupKey struct {
		Month string
		Year  string
	}
	historyMap := make(map[groupKey][]models.ActiveServiceOrder)
	var keys []groupKey

	for _, row := range rows {
		month := row.CreatedAt.Format("01")
		year := row.CreatedAt.Format("2006")
		key := groupKey{Month: month, Year: year}
		if _, exists := historyMap[key]; !exists {
			keys = append(keys, key)
		}
		historyMap[key] = append(historyMap[key], *mapActiveServiceOrder(row))
	}

	var history []models.MonthlyOrderHistory
	for _, key := range keys {
		history = append(history, models.MonthlyOrderHistory{
			Month:  key.Month,
			Year:   key.Year,
			Orders: historyMap[key],
		})
	}

	return history, nil
}

func (r *Repository) GetMitraEarningsHistory(ctx context.Context, mitraID int64) (models.EarningMonthlyHistory, error) {
	var rows []models.IncomeDetail

	err := r.DB.WithContext(ctx).Raw(`
		SELECT 
			srt.reference_id as transaction_no,
			srt.amount,
			srt.description,
			stc.code as category_name,
			srt.created_at
		FROM myschema.saldo_role_transactions srt
		JOIN myschema.saldo_transaction_categories stc ON stc.id = srt.category_id
		WHERE srt.user_id = ? 
		  AND srt.category_id IN (3, 4)
		ORDER BY srt.created_at DESC
	`, mitraID).Scan(&rows).Error

	if err != nil {
		return models.EarningMonthlyHistory{}, err
	}

	// Group by month/year in Go
	groupMap := make(map[string]*models.EarningMonthGroup)
	var keys []string
	var totalMonthly int64

	for _, row := range rows {
		month := row.CreatedAt.Format("01")
		year := row.CreatedAt.Format("2006")
		key := year + "-" + month

		if _, exists := groupMap[key]; !exists {
			keys = append(keys, key)
			groupMap[key] = &models.EarningMonthGroup{
				Month: month,
				Year:  year,
			}
		}
		groupMap[key].Items = append(groupMap[key].Items, row)
		groupMap[key].MonthlyTotal += row.Amount
		totalMonthly += row.Amount
	}

	var history []models.EarningMonthGroup
	for _, key := range keys {
		history = append(history, *groupMap[key])
	}

	return models.EarningMonthlyHistory{
		TotalAmount: totalMonthly,
		History:     history,
	}, nil
}

// GetOrderCustomerInfo retrieves customer ID and order number for a specific order
func (r *Repository) GetOrderCustomerInfo(ctx context.Context, orderID int) (customerID int64, orderNumber string, err error) {
	var row struct {
		CustomerID  int64
		OrderNumber string
	}

	err = r.DB.WithContext(ctx).Raw(`
		SELECT customer_id, order_number 
		FROM myschema.service_orders 
		WHERE id = ?
	`, orderID).Scan(&row).Error

	return row.CustomerID, row.OrderNumber, err
}

// UpdateOrderLocation updates the mitra's current location in service_orders
func (r *Repository) UpdateOrderLocation(ctx context.Context, orderID int64, lat, lng float64) error {
	return r.DB.WithContext(ctx).Exec(`
		UPDATE service_orders 
		SET mitra_latitude = ?, mitra_longitude = ?, updated_at = NOW() 
		WHERE id = ?
	`, lat, lng, orderID).Error
}

// GetExpiredOrdersToComplete finds orders that Mitra has COMPLETED (status 4) but Customer hasn't acted upon
func (r *Repository) GetExpiredOrdersToComplete(ctx context.Context, timeoutMinutes float64) ([]struct {
	OrderID    int64
	CustomerID int64
	Amount     float64
}, error) {
	var results []struct {
		OrderID    int64
		CustomerID int64
		Amount     float64
	}

	query := `
		SELECT 
			ot.order_id, 
			ot.customer_id, 
			(ot.mitra_income + ot.platform_fee + ot.thr_bonus - ot.voucher_value) as amount
		FROM order_transactions ot
		JOIN service_orders so ON so.id = ot.order_id
		WHERE so.status_id = 4
		  AND so.end_time + (? * interval '1 minute') < NOW()
		  AND ot.status_id = 1 -- PENDING
	`
	err := r.DB.WithContext(ctx).Raw(query, timeoutMinutes).Scan(&results).Error
	return results, err
}
