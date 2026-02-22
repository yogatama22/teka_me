package dokter

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"teka-api/internal/models"
	"teka-api/pkg/helper"
	"teka-api/pkg/middleware"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/minio/minio-go/v7"
)

type Handler struct {
	Service     Service
	MinioClient *minio.Client
	Hub         *OrderHub
}

func NewHandler(service Service, minioClient *minio.Client, hub *OrderHub) *Handler {
	return &Handler{
		Service:     service,
		MinioClient: minioClient,
		Hub:         hub,
	}
}

// START REGISTER MITRA
func (h *Handler) RegisterMitra(c *fiber.Ctx) error {
	// Ambil userID & nama dari JWT
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	userID := int(userIDVal.(uint))

	userName := "unknown"
	if nameVal := c.Locals("nama"); nameVal != nil {
		userName = nameVal.(string)
	}

	// Parse request body
	req := new(models.CreateMitraRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Folder user di MinIO: user-<id>-<nama>
	userDir := fmt.Sprintf("user-%d-%s", userID, helper.SanitizeFileName(userName))

	// Map untuk menyimpan semua file yang di-upload
	files := make(map[string]string)

	// Daftar field file yang diharapkan
	fileFields := []string{"foto", "foto_ktp", "foto_selfie_ktp", "str", "sip"}

	bucket := os.Getenv("S3_BUCKET")

	for _, field := range fileFields {
		if fileHeader, err := c.FormFile(field); err == nil && fileHeader != nil {
			path, err := helper.UploadFileToMinio(
				h.MinioClient,
				bucket, // ✅ BENAR
				userDir,
				fileHeader,
			)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": fmt.Sprintf("failed to upload %s: %v", field, err),
				})
			}
			// Construct full S3 URL using public endpoint
			publicURL := os.Getenv("S3_PUBLIC_URL")
			if publicURL == "" {
				// Fallback to upload endpoint if public URL not set
				publicURL = fmt.Sprintf("https://%s/%s", os.Getenv("S3_ENDPOINT"), bucket)
			}
			fullURL := fmt.Sprintf("%s/%s", publicURL, path)
			files[field] = fullURL
		}
	}

	// Simpan data ke DB via service
	if err := h.Service.RegisterMitra(userID, *req, files); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Mitra registration submitted"})
}

// START MITRA PROFILE
func (h *Handler) GetMyMitraProfile(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Get My Mitra Profile belum diimplementasi"})
}

// END MITRA PROFILE

// START CUSTOMER SEARCH DOKTER
func (h *Handler) SearchDoctor(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	customerID := int64(userIDVal.(uint))

	var req SearchDoctorRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	doctors, requestID, err := h.Service.SearchDoctor(
		c.Context(),
		customerID,
		req,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"request_id": requestID,
		"doctors":    doctors,
	})
}

// -------------------------------
// CANCEL ORDERAN
// -------------------------------
func (h *Handler) CancelOrder(c *fiber.Ctx) error {
	requestID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request id"})
	}

	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	customerID := int64(userIDVal.(uint))

	if err := h.Service.CancelOrder(c.Context(), int64(requestID), customerID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "order cancelled successfully"})
}

// START DETAIL CUSTOMER DI MITRA
func (h *Handler) GetCurrentOffer(c *fiber.Ctx) error {

	mitraID, err := middleware.UserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	offer, err := h.Service.GetCurrentOffer(
		c.Context(),
		int64(mitraID),
	)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "no active offer",
		})
	}

	return c.JSON(fiber.Map{
		"data": offer,
	})
}

// START DOKTER ACCEPT
func (h *Handler) AcceptOffer(c *fiber.Ctx) error {
	offerID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid offer id"})
	}

	mitraID, err := middleware.UserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	if err := h.Service.AcceptOffer(c.Context(), int64(offerID), int64(mitraID)); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "offer accepted"})
}

// START DETAIL DOKTER DI CUSTOMER SETELAH SERVICE ORDER
func (h *Handler) GetMyServiceOrders(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	customerID := int64(userIDVal.(uint))

	orders, err := h.Service.GetCustomerServiceOrders(c.Context(), customerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if len(orders) == 0 {
		return c.Status(404).JSON(fiber.Map{"message": "no service orders found"})
	}

	return c.JSON(fiber.Map{"orders": orders})
}

// AKTIF ORDER CUSTOMER
func (h *Handler) GetCustomerCurrentOrder(c *fiber.Ctx) error {
	customerID, _ := middleware.UserID(c)

	order, err := h.Service.GetCustomerCurrentOrder(
		c.Context(),
		int64(customerID),
	)

	if err != nil {
		return c.JSON(fiber.Map{
			"message": "Belum ada pesanan aktif",
		})
	}

	return c.JSON(fiber.Map{
		"data":    order,
		"message": "Pesanan aktif ditemukan",
	})
}

// AKTIF ORDER MITRA
func (h *Handler) GetMitraCurrentOrder(c *fiber.Ctx) error {
	mitraID, _ := middleware.UserID(c)

	order, err := h.Service.GetMitraCurrentOrder(
		c.Context(),
		int64(mitraID),
	)

	if err != nil {
		return c.JSON(fiber.Map{
			"message": "Belum ada pesanan aktif",
		})
	}

	return c.JSON(fiber.Map{
		"data":    order,
		"message": "Pesanan aktif ditemukan",
	})
}

// CompleteService dari dokter selesai service
func (h *Handler) CompleteOrder(c *fiber.Ctx) error {
	orderID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid order id"})
	}

	mitraIDVal := c.Locals("user_id")
	if mitraIDVal == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	mitraID := int64(mitraIDVal.(uint))

	var req models.CompleteOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	req.OrderID = int64(orderID)

	// 📂 folder MinIO per order
	userDir := fmt.Sprintf("orders/%d", orderID)

	// ambil multiple photos
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid multipart"})
	}

	var photoPaths []string
	if files := form.File["photos"]; len(files) > 0 {
		for _, file := range files {
			path, err := helper.UploadFileToMinio(
				h.MinioClient,
				"uploads",
				userDir,
				file,
			)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "upload photo failed",
				})
			}
			photoPaths = append(photoPaths, path)
		}
	}

	req.Attachments = photoPaths

	if err := h.Service.CompleteOrder(
		c.Context(),
		mitraID,
		&req,
	); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "order completed successfully",
	})
}

// CompleteOrderUser (Customer completing the order)
func (h *Handler) CompleteOrderUser(c *fiber.Ctx) error {
	orderID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid order id"})
	}

	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	customerID := int64(userIDVal.(uint))

	// Customer completion only triggers status update and balance deduction.
	// No request body is needed.
	if err := h.Service.CompleteOrderUser(
		c.Context(),
		customerID,
		&models.CompleteOrderRequest{OrderID: int64(orderID)},
	); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "order completed successfully",
	})
}

// RateDoctor: customer beri rating
func (h *Handler) RateDoctor(c *fiber.Ctx) error {
	var req models.RatingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	customerIDCtx, err := middleware.UserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	customerID := int64(customerIDCtx)

	if err := h.Service.RateDoctor(c.Context(), customerID, req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Rating submitted successfully",
	})
}

func (c *Handler) GetVoucherHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.ParseInt(r.Header.Get("X-User-ID"), 10, 64) // contoh ambil dari JWT
	vouchers, err := c.Service.GetActiveVoucher(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vouchers)
}

// UpdateServiceOrderStatus: mitra update status order
func (h *Handler) UpdateServiceOrderStatus(c *fiber.Ctx) error {
	orderID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid order id"})
	}

	var body struct {
		StatusID int16 `json:"status_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	// update DB
	if err := h.Service.UpdateServiceOrderStatus(
		c.Context(),
		orderID,
		body.StatusID,
	); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// 🔥🔥🔥 INI YANG SELAMA INI HILANG 🔥🔥🔥
	log.Printf("Broadcasting status update: order=%d status=%d", orderID, body.StatusID)

	h.Hub.Broadcast(orderID, fiber.Map{
		"event":     "status_updated",
		"order_id":  orderID,
		"status_id": body.StatusID,
		"ts":        time.Now(),
	})

	return c.JSON(fiber.Map{
		"message": "status updated",
	})
}

// GetRatingSummary: mitra melihat rata-rata rating
func (h *Handler) GetRatingSummary(c *fiber.Ctx) error {
	mitraIDCtx, err := middleware.UserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	mitraID := int64(mitraIDCtx)

	summary, err := h.Service.GetRatingSummary(c.Context(), mitraID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data": summary,
	})
}

// GetRatingHistory: mitra melihat riwayat rating detail grouped by day
func (h *Handler) GetRatingHistory(c *fiber.Ctx) error {
	mitraIDCtx, err := middleware.UserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	mitraID := int64(mitraIDCtx)

	history, err := h.Service.GetRatingHistory(c.Context(), mitraID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data": history,
	})
}

// GetOrderSummary: mitra melihat total order
func (h *Handler) GetOrderSummary(c *fiber.Ctx) error {
	mitraIDCtx, err := middleware.UserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	mitraID := int64(mitraIDCtx)

	summary, err := h.Service.GetOrderSummary(c.Context(), mitraID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data": summary,
	})
}

// GetOrderHistory: mitra melihat riwayat order per hari bulan berjalan
func (h *Handler) GetOrderHistory(c *fiber.Ctx) error {
	mitraIDCtx, err := middleware.UserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	mitraID := int64(mitraIDCtx)

	history, err := h.Service.GetOrderHistory(c.Context(), mitraID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data": history,
	})
}

// GetMitraBalance: mitra melihat saldo
func (h *Handler) GetMitraBalance(c *fiber.Ctx) error {
	mitraIDCtx, err := middleware.UserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	mitraID := int64(mitraIDCtx)

	balance, err := h.Service.GetMitraBalance(c.Context(), mitraID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"balance": balance,
	})
}

// Withdraw: mitra minta penarikan saldo
func (h *Handler) Withdraw(c *fiber.Ctx) error {
	mitraIDCtx, err := middleware.UserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	mitraID := int64(mitraIDCtx)

	var req models.WithdrawRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	if err := h.Service.Withdraw(c.Context(), mitraID, req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "withdrawal request submitted successfully",
	})
}

// GetMitraEarningsHistory: mitra melihat riwayat penarikan saldo
func (h *Handler) GetMitraEarningsHistory(c *fiber.Ctx) error {
	mitraIDCtx, err := middleware.UserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	mitraID := int64(mitraIDCtx)

	history, err := h.Service.GetMitraEarningsHistory(c.Context(), mitraID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(history)
}
