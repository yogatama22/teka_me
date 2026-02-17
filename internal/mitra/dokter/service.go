package dokter

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"teka-api/internal/models"
	"teka-api/internal/realtime/firebase"
	"teka-api/internal/realtime/redis"
	"time"

	"gorm.io/gorm"
)

type Service struct {
	Repo *Repository
	Hub  *OrderHub
}

func NewService(r *Repository, hub *OrderHub) *Service {
	return &Service{
		Repo: r,
		Hub:  hub,
	}
}

// REGISTER MITRA
func (s *Service) RegisterMitra(
	userID int,
	req models.CreateMitraRequest,
	files map[string]string,
) error {

	// 1. user_mitra
	if err := s.Repo.CreateUserMitra(userID, req.JobCategoryID, req.JobSubCategoryID); err != nil {
		return err
	}

	// 2 role mitra (role_id = 2, inactive)
	if err := s.Repo.CreateUserRole(userID, 2, false); err != nil {
		return err
	}

	// 3 mitra_detail
	if err := s.Repo.CreateMitraDetail(
		userID,
		req.JobCategoryID,
		req.JobSubCategoryID,
		req.Latitude,
		req.Longitude,
	); err != nil {
		return err
	}

	// 4. dokumen
	for docType, url := range files {
		if err := s.Repo.CreateDocument(userID, docType, url); err != nil {
			return err
		}
	}

	return nil
}

// SERACH
// -------------------------
// SERVICE LAYER: SearchDoctor
// -------------------------
func (s *Service) SearchDoctor(
	ctx context.Context,
	customerID int64,
	req SearchDoctorRequest,
) ([]models.DoctorSearchResult, int64, error) {

	// 1️⃣ Search dokter
	doctors, err := s.Repo.SearchDoctors(
		ctx,
		req.JobCategoryID,
		req.JobSubCategoryID,
		req.Latitude,
		req.Longitude,
	)
	if err != nil {
		return nil, 0, err
	}
	if len(doctors) == 0 {
		return nil, 0, errors.New("no doctor available")
	}

	// 2️⃣ Optional fields
	var voucherID *int64
	if req.VoucherID != 0 {
		v := req.VoucherID
		voucherID = &v
	}

	var subCatID *int64
	if req.JobSubCategoryID != 0 {
		s := req.JobSubCategoryID
		subCatID = &s
	}

	// 3️⃣ Create customer request
	requestID, err := s.Repo.CreateCustomerRequest(
		ctx,
		models.CustomerRequest{
			CustomerID:       customerID,
			JobCategoryID:    req.JobCategoryID,
			JobSubCategoryID: subCatID,
			Keluhan:          req.Keluhan,
			Latitude:         req.Latitude,
			Longitude:        req.Longitude,
			Radius:           req.Radius,
			Price:            req.Price,
			VoucherID:        voucherID,
		},
	)
	if err != nil {
		return nil, 0, err
	}

	// 4️⃣ Create offers
	if err := s.Repo.CreateOffers(ctx, requestID, doctors); err != nil {
		return nil, 0, err
	}

	// 5️⃣ Push notif ke mitra pertama
	firstDoctor := doctors[0]
	tokens, err := s.Repo.GetFCMTokensByUserID(ctx, firstDoctor.MitraID)
	if err != nil {
		log.Println("GetFCMTokens error:", err)
		return doctors, requestID, nil
	}

	if len(tokens) == 0 {
		log.Println("No FCM token for mitra:", firstDoctor.MitraID)
		return doctors, requestID, nil
	}

	// 6️⃣ Goroutine async untuk FCM + logging
	go func(tokens []string, requestID int64, mitraID int64) {
		// timeout supaya gak hang
		fcmCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Payload notification + data
		results := firebase.SendFCMToTokens(
			fcmCtx,
			tokens,
			"Ada Orderan Baru 🚨",
			"Ada customer membutuhkan layanan kamu",
			map[string]string{
				"type":       "NEW_ORDER",
				"request_id": strconv.FormatInt(requestID, 10),
			},
		)

		// Simpan log ke DB per token
		for token, err := range results {
			status := "SENT"
			var errStr *string
			if err != nil {
				status = "FAILED"
				s := err.Error()
				errStr = &s
			}

			if dbErr := s.Repo.LogFCMResult(mitraID, token, status, errStr); dbErr != nil {
				log.Println("❌ Failed to log FCM result:", dbErr)
			}
		}
	}(tokens, requestID, firstDoctor.MitraID)

	return doctors, requestID, nil
}

// -------------------------------
// CANCEL ORDERAN
// -------------------------------
func (s *Service) CancelOrder(ctx context.Context, requestID, customerID int64) error {
	// Ambil request
	req, err := s.Repo.GetCustomerRequest(ctx, requestID)
	if err != nil {
		return err
	}

	// Cek ownership
	if req.CustomerID != customerID {
		return errors.New("not your request")
	}

	// Cek status
	if req.Status == "matched" || req.Status == "completed" {
		return errors.New("cannot cancel this order")
	}

	// Cancel request + offers
	return s.Repo.CancelCustomerRequest(ctx, requestID)
}

// START DETAIL CUSTOMER DI MITRA
func (s *Service) GetCurrentOffer(
	ctx context.Context,
	mitraID int64,
) (*models.CurrentOffer, error) {

	offer, err := s.Repo.GetCurrentOfferForMitra(ctx, mitraID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("no active offer")
		}
		return nil, err
	}

	return offer, nil
}

//START DETAIL CUSTOMER DI MITRA

func (s *Service) GetNextOfferForMitra(ctx context.Context, mitraID int64) (models.MitraOffer, error) {
	return s.Repo.GetNextOfferForMitra(ctx, mitraID)
}

// AcceptOffer service layer
func (s *Service) AcceptOffer(
	ctx context.Context,
	offerID int64,
	mitraID int64,
) error {

	// Ambil offer + lokasi
	o, err := s.Repo.GetOfferForAccept(ctx, offerID, mitraID)
	if err != nil {
		return err
	}

	// Accept offer + cancel lainnya + create service order
	return s.Repo.AcceptOfferAndCreateOrder(ctx, o)
}

// START DETAIL DOKTER DI CUSTOMER
func (s *Service) GetCustomerServiceOrders(
	ctx context.Context,
	customerID int64,
) ([]models.CustomerServiceOrder, error) {

	return s.Repo.GetCustomerServiceOrders(ctx, customerID)
}

// AKTIF ORDER
func (s *Service) GetCurrentOrder(
	ctx context.Context,
	userID int64,
	isMitra bool,
) (*models.ActiveServiceOrder, error) {

	return s.Repo.GetActiveServiceOrder(ctx, userID, isMitra)
}

// Complete service
func (s *Service) CompleteOrder(
	ctx context.Context,
	mitraID int64,
	req *models.CompleteOrderRequest,
) error {

	if req.Subtotal <= 0 {
		return errors.New("invalid subtotal")
	}

	amount := req.Subtotal - req.Discount

	data := CompleteOrderTxData{
		Amount:        amount,
		Subtotal:      req.Subtotal,
		Discount:      req.Discount,
		PlatformFee:   req.PlatformFee,
		MitraIncome:   amount - req.PlatformFee,
		PaymentMethod: req.PaymentMethod,
		Note:          req.Note,
		Attachments:   req.Attachments,
	}

	// 1️⃣ Complete order (DB transaction) — BIARIN
	if err := s.Repo.CompleteOrder(ctx, req.OrderID, mitraID, data); err != nil {
		return err
	}

	// 2️⃣ Update status jadi COMPLETED (4) — CAST DI SINI AJA
	if err := s.UpdateServiceOrderStatus(ctx, int(req.OrderID), 4); err != nil {
		return err
	}

	return nil
}

// CompleteOrderUser (Customer completing the order)
func (s *Service) CompleteOrderUser(
	ctx context.Context,
	customerID int64,
	req *models.CompleteOrderRequest,
) error {

	if req.Subtotal <= 0 {
		return errors.New("invalid subtotal")
	}

	amount := req.Subtotal - req.Discount

	// 1️⃣ Complete order & Potong Saldo (Atomic)
	// - Cek saldo di saldo_role_transactions
	// - Insert mutasi saldo baru
	// - Update status order_transactions jadi Paid
	// - Update status service_orders jadi 6 (Finished)
	if err := s.Repo.DeductCustomerBalance(ctx, customerID, req.OrderID, amount); err != nil {
		return err
	}

	// 2️⃣ Broadcast Update
	s.Hub.Broadcast(int(req.OrderID), map[string]interface{}{
		"event":     "order_status_updated",
		"order_id":  req.OrderID,
		"status_id": 6,
	})

	// 3️⃣ Delete Chat History from Redis
	redis.Rdb.Del(ctx, fmt.Sprintf("order_chat:%d", req.OrderID))

	return nil
}

func (s *Service) GetRatingSummary(ctx context.Context, mitraID int64) (models.RatingSummary, error) {
	return s.Repo.GetMitraRatingSummary(ctx, mitraID)
}

func (s *Service) GetRatingHistory(ctx context.Context, mitraID int64) ([]models.DayHistory, error) {
	return s.Repo.GetMitraRatingHistory(ctx, mitraID)
}

func (s *Service) GetOrderSummary(ctx context.Context, mitraID int64) (models.OrderSummary, error) {
	return s.Repo.GetMitraOrderSummary(ctx, mitraID)
}

func (s *Service) GetOrderHistory(ctx context.Context, mitraID int64) ([]models.OrderDayHistory, error) {
	return s.Repo.GetMitraOrderHistory(ctx, mitraID)
}

// Rate doctor
func (s *Service) RateDoctor(
	ctx context.Context,
	customerID int64,
	req models.RatingRequest,
) error {

	// 1. Ambil mitra dari service_order
	mitraID, err := s.Repo.GetServiceForRating(
		ctx,
		req.ServiceOrderID,
		customerID,
	)
	if err != nil || mitraID == 0 {
		return fmt.Errorf("service not found or not completed")
	}

	// 2. Insert rating
	return s.Repo.InsertRating(
		ctx,
		req.ServiceOrderID,
		mitraID,
		customerID,
		req.Rating,
		req.Review,
	)
}

// service/reject_offer.go

// Ambil voucher aktif customer
func (s *Service) GetActiveVoucher(ctx context.Context, userID int64) ([]models.Voucher, error) {
	return s.Repo.GetActiveVoucher(ctx, userID)
}

// WORKER
func (s *Service) RunOfferTimeoutWorker(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second) // cek tiap 3 detik
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			offers, err := s.Repo.GetExpiredOffers(ctx, 10)
			if err != nil {
				log.Println("timeout worker error:", err)
				continue
			}

			for _, offer := range offers {
				err := s.Repo.TimeoutAndMoveNext(
					ctx,
					offer.RequestID,
					offer.Sequence,
				)
				if err != nil {
					log.Println("process timeout error:", err)
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

var allowedTransitions = map[int16][]int16{
	1: {2, 5},
	2: {3, 5, 6},
	3: {4, 5, 6},
	4: {6},
}

func (s *Service) UpdateServiceOrderStatus(
	ctx context.Context,
	orderID int,
	newStatusID int16,
) error {

	currentStatus, err := s.Repo.GetServiceOrderStatus(ctx, orderID)
	if err != nil {
		return err
	}

	if currentStatus == 4 || currentStatus == 5 {
		return errors.New("order already finalized")
	}

	valid := false
	for _, s := range allowedTransitions[currentStatus] {
		if s == newStatusID {
			valid = true
			break
		}
	}
	if !valid {
		return errors.New("invalid status transition")
	}

	exists, err := s.Repo.ServiceOrderStatusExists(ctx, newStatusID)
	if err != nil || !exists {
		return errors.New("invalid status_id")
	}

	// ✅ UPDATE DB
	if err := s.Repo.UpdateServiceOrderStatus(
		ctx,
		orderID,
		currentStatus,
		newStatusID,
	); err != nil {
		return err
	}

	// 🔥 BROADCAST REALTIME
	s.Hub.Broadcast(orderID, map[string]interface{}{
		"event":     "order_status_updated",
		"order_id":  orderID,
		"status_id": newStatusID,
	})

	// 🛡️ DELETE CHAT HISTORY IF FINALIZED (4: COMPLETED_BY_MITRA, 5: CANCELLED, 6: FINISHED)
	if newStatusID == 5 || newStatusID == 6 {
		redis.Rdb.Del(ctx, fmt.Sprintf("order_chat:%d", orderID))
	}

	return nil
}
