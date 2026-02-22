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
		return nil, 0, errors.New("tidak ada dokter yang tersedia")
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
			VoucherValue:     &req.VoucherValue,
			PlatformFee:      req.PlatformFee,
			THRBonus:         req.THRBonus,
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
			return nil, errors.New("Maaf, waktu penawaran sudah habis")
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

// GET CUSTOMER CURRENT ORDER
func (s *Service) GetCustomerCurrentOrder(
	ctx context.Context,
	customerID int64,
) (*models.ActiveServiceOrder, error) {

	return s.Repo.GetActiveServiceOrderForCustomer(ctx, customerID)
}

// GET MITRA CURRENT ORDER
func (s *Service) GetMitraCurrentOrder(
	ctx context.Context,
	mitraID int64,
) (*models.ActiveServiceOrder, error) {

	return s.Repo.GetActiveServiceOrderForMitra(ctx, mitraID)
}

// Complete service
func (s *Service) CompleteOrder(
	ctx context.Context,
	mitraID int64,
	req *models.CompleteOrderRequest,
) error {
	data := CompleteOrderTxData{
		Note:        req.Note,
		Attachments: req.Attachments,
	}

	// 1️⃣ Complete order (DB transaction)
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

	// 1️⃣ Complete order & Potong Saldo (Atomic)
	// - Cek saldo di saldo_role_transactions
	// - Insert mutasi saldo baru
	// - Update status order_transactions jadi Paid
	// - Update status service_orders jadi 6 (Finished)
	// - Nilai amount diambil otomatis dari service_orders (price + platform_fee + thr_bonus - voucher_value)
	if err := s.Repo.DeductCustomerBalance(ctx, customerID, req.OrderID); err != nil {
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

func (s *Service) GetMitraBalance(ctx context.Context, mitraID int64) (int64, error) {
	return s.Repo.GetLatestBalance(ctx, mitraID)
}

func (s *Service) Withdraw(ctx context.Context, mitraID int64, req models.WithdrawRequest) error {
	if req.Amount <= 0 {
		return errors.New("jumlah penarikan harus lebih dari 0")
	}

	return s.Repo.RequestWithdrawal(
		ctx,
		mitraID,
		req.Amount,
		req.BankName,
		req.AccountNumber,
		req.AccountHolder,
	)
}

func (s *Service) GetRatingSummary(ctx context.Context, mitraID int64) (models.RatingSummary, error) {
	return s.Repo.GetMitraRatingSummary(ctx, mitraID)
}

func (s *Service) GetRatingHistory(ctx context.Context, mitraID int64) ([]models.MonthlyRatingHistory, error) {
	return s.Repo.GetMitraRatingHistory(ctx, mitraID)
}

func (s *Service) GetOrderSummary(ctx context.Context, mitraID int64) (models.OrderSummary, error) {
	return s.Repo.GetMitraOrderSummary(ctx, mitraID)
}

func (s *Service) GetOrderHistory(ctx context.Context, mitraID int64) ([]models.MonthlyOrderHistory, error) {
	return s.Repo.GetMitraOrderHistory(ctx, mitraID)
}

func (s *Service) GetMitraEarningsHistory(ctx context.Context, mitraID int64) (models.EarningMonthlyHistory, error) {
	return s.Repo.GetMitraEarningsHistory(ctx, mitraID)
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
	ticker := time.NewTicker(5 * time.Second) // cek tiap 5 detik
	defer ticker.Stop()

	log.Println("👷 Offer Timeout Worker started")

	for {
		select {
		case <-ticker.C:
			// 1. Ambil timeout dari global parameter (default 60)
			timeoutSeconds := 60
			val, err := s.Repo.GetGlobalParameter(ctx, "OFFER_TIMEOUT_SECONDS")
			if err == nil && val != "" {
				if t, err := strconv.Atoi(val); err == nil {
					timeoutSeconds = t
				}
			}

			// 2. Ambil penawaran yang sudah expired
			offers, err := s.Repo.GetExpiredOffers(ctx, timeoutSeconds)
			if err != nil {
				log.Println("❌ timeout worker error:", err)
				continue
			}

			if len(offers) > 0 {
				log.Printf("🔍 Worker found %d expired offers", len(offers))
			}

			for _, offer := range offers {
				log.Printf("⏰ Offer %d expired for request %d (Dokter: %s), moving to next sequence...",
					offer.ID, offer.RequestID, offer.MitraName)

				nextMitraID, nextMitraName, err := s.Repo.TimeoutAndMoveNext(
					ctx,
					offer.RequestID,
					offer.Sequence,
				)
				if err != nil {
					log.Println("❌ process timeout error:", err)
					continue
				}

				if nextMitraID != 0 {
					log.Printf("🚀 Distributing request %d to Dokter selanjutnya: %s (%d)",
						offer.RequestID, nextMitraName, nextMitraID)
					// Push notif ke mitra selanjutnya
					tokens, err := s.Repo.GetFCMTokensByUserID(ctx, nextMitraID)
					if err != nil {
						log.Println("❌ FCM Error (worker):", err)
						continue
					}

					if len(tokens) > 0 {
						go func(tks []string, mID int64, rID int64) {
							fcmCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
							defer cancel()

							results := firebase.SendFCMToTokens(
								fcmCtx,
								tks,
								"Ada Orderan Baru 🚨",
								"Ada customer membutuhkan layanan kamu",
								map[string]string{
									"type":       "NEW_ORDER",
									"request_id": strconv.FormatInt(rID, 10),
								},
							)

							for token, err := range results {
								status := "SENT"
								var errStr *string
								if err != nil {
									status = "FAILED"
									s := err.Error()
									errStr = &s
								}
								_ = s.Repo.LogFCMResult(mID, token, status, errStr)
							}
						}(tokens, nextMitraID, offer.RequestID)
					}
				} else {
					log.Printf("🏁 Request %d reached end of queue or no more offers.", offer.RequestID)
				}
			}

		case <-ctx.Done():
			log.Println("👷 Offer Timeout Worker stopped")
			return
		}
	}
}

// allowedTransitions: status order yang boleh diubah
var allowedTransitions = map[int16][]int16{
	1: {2, 5},
	2: {3, 5, 6},
	3: {4, 5, 6},
	4: {6},
}

// UpdateServiceOrderStatus: mitra update status order
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

	// 🔔 NOTIFIKASI FCM
	if newStatusID == 2 || newStatusID == 3 {
		go func(oID int, sID int16) {
			fcmCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			customerID, orderNo, err := s.Repo.GetOrderCustomerInfo(fcmCtx, oID)
			if err != nil {
				log.Printf("FCM: failed to get customer info for order %d: %v", oID, err)
				return
			}

			tokens, err := firebase.GetFCMTokensByUserID(uint(customerID))
			if err != nil || len(tokens) == 0 {
				log.Printf("FCM: no tokens found for customer %d", customerID)
				return
			}

			title := "Dokter OTW 🏎️"
			body := fmt.Sprintf("Dokter sedang menuju lokasi untuk order %s", orderNo)
			if sID == 3 {
				title = "Dokter Sampai 🏁"
				body = fmt.Sprintf("Dokter sudah sampai di lokasi untuk order %s", orderNo)
			}

			results := firebase.SendFCMToTokens(fcmCtx, tokens, title, body, map[string]string{
				"type":     "ORDER_STATUS_UPDATE",
				"order_id": strconv.Itoa(oID),
				"status":   strconv.Itoa(int(sID)),
			})

			for token, err := range results {
				status := "SENT"
				var errStr *string
				if err != nil {
					status = "FAILED"
					s := err.Error()
					errStr = &s
				}
				_ = s.Repo.LogFCMResult(customerID, token, status, errStr)
			}
		}(orderID, newStatusID)
	}

	return nil
}

// RunAutoOrderCompletionWorker periodically completes orders not confirmed by customers
func (s *Service) RunAutoOrderCompletionWorker(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute) // cek tiap 5 menit
	defer ticker.Stop()

	log.Println("👷 Auto Order Completion Worker started")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 1. Ambil timeout dari global parameter (default 60 menit)
			timeoutMinutes := 60.0
			val, err := s.Repo.GetGlobalParameter(ctx, "ORDER_AUTO_COMPLETE_DURATION")
			if err == nil && val != "" {
				// Coba parse sebagai float dulu (anggap menit)
				if t, err := strconv.ParseFloat(val, 64); err == nil {
					timeoutMinutes = t
				} else {
					// Jika gagal, coba parse pakai format Go duration (1h, 30m, dll)
					if dur, err := time.ParseDuration(val); err == nil {
						timeoutMinutes = dur.Minutes()
					}
				}
			}

			// 2. Ambil order yang sudah waktunya auto-complete
			orders, err := s.Repo.GetExpiredOrdersToComplete(ctx, timeoutMinutes)
			if err != nil {
				log.Println("❌ auto-complete worker error:", err)
				continue
			}

			if len(orders) > 0 {
				log.Printf("🔍 Auto-complete worker found %d orders to process", len(orders))
			}

			for _, order := range orders {
				log.Printf("🤖 Auto-completing order %d for customer %d...", order.OrderID, order.CustomerID)

				// Re-use existing completion logic: Deduct balance & Update status to 6
				err := s.Repo.DeductCustomerBalance(ctx, order.CustomerID, order.OrderID)
				if err != nil {
					log.Printf("❌ Failed to auto-complete order %d: %v", order.OrderID, err)
					continue
				}

				log.Printf("✅ Order %d auto-completed successfully", order.OrderID)
			}
		}
	}
}
