package auth

import (
	"log"
	"teka-api/internal/models"
	"teka-api/pkg/database"
	"teka-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// =========REGISTER======== \\
func RegisterController(c *fiber.Ctx) error {
	var body struct {
		Nama       string `json:"nama"`
		Phone      string `json:"phone"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		ConfirmPwd string `json:"confirm_password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	if body.Password != body.ConfirmPwd {
		return c.Status(400).JSON(fiber.Map{"error": "password not match"})
	}

	// Cek email & phone
	if EmailExists(body.Email) {
		return c.Status(400).JSON(fiber.Map{"error": "email already registered"})
	}
	if PhoneExists(body.Phone) {
		return c.Status(400).JSON(fiber.Map{"error": "phone already registered"})
	}

	// Simpan temp_user
	hashed, _ := utils.HashPassword(body.Password)
	temp := models.TempUser{
		Nama:      body.Nama,
		Phone:     body.Phone,
		Email:     body.Email,
		Password:  hashed,
		CreatedAt: utils.NowJakarta(),
	}
	if err := SaveTempUser(&temp); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed save temp user"})
	}

	// Kirim OTP
	if err := SendOtpToEmail(body.Email); err != nil {
		log.Printf("❌ Failed send OTP to %s: %v", body.Email, err) // TAMBAHKAN INI
		return c.Status(500).JSON(fiber.Map{
			"error":  "failed send otp",
			"detail": err.Error(), // TAMBAHKAN INI untuk development
		})
	}

	return c.JSON(fiber.Map{"message": "OTP sent"})
}

// =========VERIFY======== \\
func VerifyOtpController(c *fiber.Ctx) error {
	var body struct {
		Email string `json:"email"`
		Otp   string `json:"otp"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	if err := VerifyAndRegister(body.Email, body.Otp); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Registration complete",
	})
}

// =========RESEND OTP======== \\
func ResendOtpController(c *fiber.Ctx) error {
	var body struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	// Pastikan email sedang dalam proses registrasi
	var temp models.TempUser
	if err := database.DB.Where("email = ?", body.Email).First(&temp).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "no pending registration for this email"})
	}

	// Hapus OTP lama
	database.DB.Where("email = ?", body.Email).Delete(&models.OTP{})

	// Kirim OTP baru
	if err := SendOtpToEmail(body.Email); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "new OTP sent"})
}

// =========LOGIN======== \\
func LoginController(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	user, err := Login(body.Email, body.Password)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	token, err := utils.GenerateToken(user.ID, user.Nama, user.Email, user.Phone)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed generate token"})
	}

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"token":   token,
	})
}

// =========CEK PIN======== \\
func CheckPin(c *fiber.Ctx) error {
	uid := c.Locals("user_id")
	if uid == nil {
		return c.Status(401).JSON(fiber.Map{"error": "user_id claim missing"})
	}

	userID := uid.(uint)

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	if user.Pin == "" {
		return c.JSON(fiber.Map{"pin": false})
	}

	return c.JSON(fiber.Map{"pin": true})
}

// =========CREATE PIN======== \\
func CreatePin(c *fiber.Ctx) error {
	userIDValue := c.Locals("user_id")
	userID, ok := userIDValue.(uint)
	if !ok {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user_id in token"})
	}

	var body struct {
		Pin string `json:"pin"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	if len(body.Pin) != 6 {
		return c.Status(400).JSON(fiber.Map{"error": "PIN must be 6 digits"})
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	if user.Pin != "" {
		return c.Status(400).JSON(fiber.Map{"error": "PIN already set"})
	}

	hashedPin, _ := utils.HashPassword(body.Pin)
	user.Pin = hashedPin
	database.DB.Save(&user)

	return c.JSON(fiber.Map{"message": "PIN created"})
}

// =========UPDATE PIN======== \\
func UpdatePin(c *fiber.Ctx) error {
	userIDValue := c.Locals("user_id")
	userID, ok := userIDValue.(uint)
	if !ok {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user_id in token"})
	}

	var body struct {
		NewPin string `json:"new_pin"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	// Validasi PIN baru 6 digit
	if len(body.NewPin) != 6 {
		return c.Status(400).JSON(fiber.Map{"error": "New PIN must be 6 digits"})
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	// Hash PIN baru
	hashedNewPin, err := utils.HashPassword(body.NewPin)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to hash new PIN"})
	}

	user.Pin = hashedNewPin
	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update PIN"})
	}

	return c.JSON(fiber.Map{
		"message": "PIN updated",
		// "pin_updated": true,
	})
}

// ====== VERIFY PIN ===== \\
func VerifyPin(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var body struct {
		Pin string `json:"pin"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	if user.Pin == "" {
		return c.Status(400).JSON(fiber.Map{"error": "PIN not set"})
	}

	valid := utils.CheckPassword(user.Pin, body.Pin) // <-- hash vs plain
	return c.JSON(fiber.Map{"valid": valid})
}

// ====== GET PROFILE ===== \\
func GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	userResp, err := GetUserWithActiveRole(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	// ambil saldo sesuai role aktif
	saldo, err := GetSaldoByRole(userID, userResp.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed get saldo"})
	}

	// update saldo di response
	userResp.Saldo = saldo

	return c.JSON(userResp)
}

// ====== SWITCH AKUN ROLES ===== \\
func SwitchRoleController(c *fiber.Ctx) error {
	userId := c.Locals("user_id").(uint)

	var body struct {
		RoleID uint `json:"role_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to start transaction"})
	}

	// pastikan user punya role tsb
	var existing models.UserRole
	if err := tx.
		Where("user_id = ? AND role_id = ?", userId, body.RoleID).
		First(&existing).Error; err != nil {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{"error": "user does not have this role"})
	}

	// NONAKTIFKAN SEMUA ROLE (FORCE false)
	if err := tx.Model(&models.UserRole{}).
		Where("user_id = ?", userId).
		UpdateColumn("active", false).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "failed to deactivate roles"})
	}

	// AKTIFKAN ROLE PILIHAN
	result := tx.Model(&models.UserRole{}).
		Where("user_id = ? AND role_id = ?", userId, body.RoleID).
		UpdateColumn("active", true)

	if result.Error != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "failed to activate role"})
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{"error": "role not updated"})
	}

	// commit
	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to commit transaction"})
	}

	var role models.Role
	_ = database.DB.First(&role, body.RoleID)

	return c.JSON(fiber.Map{
		"message": "role switched successfully",
		"active_role": fiber.Map{
			"role_id":   body.RoleID,
			"role_name": role.Name,
		},
	})
}

// ====== GET ACTIVE ROLE ===== \\
func GetActiveRole(c *fiber.Ctx) error {
	userId := c.Locals("user_id").(uint)

	// struct untuk hasil query
	type ActiveRole struct {
		RoleID   uint   `json:"role_id"`
		RoleName string `json:"role_name"`
	}

	var ar ActiveRole

	err := database.DB.
		Table("user_roles AS ur").
		Select("ur.role_id, r.name AS role_name").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Where("ur.user_id = ? AND ur.active = true", userId).
		First(&ar).Error

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "no active role"})
	}

	return c.JSON(fiber.Map{"active_role": ar})
}

type FCMRequest struct {
	FCMToken   string `json:"fcm_token"`
	DeviceName string `json:"device_name"`
	DeviceID   string `json:"device_id"`
}

func RegisterFCM(c *fiber.Ctx) error {
	// Ambil user_id dari context
	userIDVal := c.Locals("user_id")
	userID, ok := userIDVal.(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	// Bind body
	var body FCMRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	if body.FCMToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "fcm_token required"})
	}

	// Simpan atau update FCM token
	if err := SaveFCMToken(userID, body.FCMToken, body.DeviceName, body.DeviceID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed save fcm token"})
	}

	return c.JSON(fiber.Map{"message": "FCM token registered"})
}

// ========= TOP UP ========= \\
func TopUpHandler(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	// In a real app, we'd get the role from the context or body
	// For now, let's assume it's for the active role
	userResp, err := GetUserWithActiveRole(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	var body struct {
		Amount int64 `json:"amount"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	if body.Amount <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "amount must be greater than 0"})
	}

	if err := TopUpBalance(userID, userResp.RoleID, body.Amount); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to top up", "detail": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Top up successful"})
}

// ========= TRANSACTION HISTORY ========= \\
func TransactionHistoryHandler(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	userResp, err := GetUserWithActiveRole(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	history, err := GetTransactions(userID, userResp.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to get transaction history"})
	}

	return c.JSON(fiber.Map{
		"history": history,
	})
}
