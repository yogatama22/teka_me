package auth

import (
	"teka-api/internal/models"
	"teka-api/pkg/database"
	"time"

	"gorm.io/gorm"
)

// ================= OTP ====================

func SaveOTP(email, code string) error {
	// Hapus OTP sebelumnya
	database.DB.Where("email = ?", email).Delete(&models.OTP{})

	otp := models.OTP{
		Email:     email,
		Otp:       code,
		ExpiredAt: time.Now().Add(5 * time.Minute),
	}

	return database.DB.Create(&otp).Error
}

func FindOTPTx(tx *gorm.DB, email, otp string) (*models.OTP, error) {
	var data models.OTP
	err := tx.
		Where("email = ? AND otp = ? AND expired_at >= NOW()", email, otp).
		First(&data).Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func VerifyOtpTx(tx *gorm.DB, email, otp string) error {
	data, err := FindOTPTx(tx, email, otp)
	if err != nil {
		return err
	}
	return tx.Delete(&data).Error
}

// ================= TEMP USER ====================

func SaveTempUser(u *models.TempUser) error {
	// bersihkan temp user dengan email yg sama
	database.DB.Where("email = ?", u.Email).Delete(&models.TempUser{})
	return database.DB.Create(u).Error
}

func GetTempUserTx(tx *gorm.DB, email string) (*models.TempUser, error) {
	var u models.TempUser
	err := tx.Where("email = ?", email).First(&u).Error
	return &u, err
}

func DeleteTempUserTx(tx *gorm.DB, email string) error {
	return tx.Where("email = ?", email).Delete(&models.TempUser{}).Error
}

// ================= VALIDASI DATA EXIST ====================

func EmailExists(email string) bool {
	var u models.User
	return database.DB.Where("email = ?", email).First(&u).Error == nil
}

func PhoneExists(phone string) bool {
	var u models.User
	return database.DB.Where("phone = ?", phone).First(&u).Error == nil
}

func SaveFCMToken(userID uint, fcmToken, deviceName, deviceID string) error {
	// Cek apakah user sudah punya token
	var existing struct {
		ID uint
	}
	err := database.DB.Table("user_fcm_tokens").
		Where("user_id = ?", userID).
		First(&existing).Error

	now := time.Now()

	if err == nil {
		// update token & device_name & device_id
		return database.DB.Table("user_fcm_tokens").
			Where("id = ?", existing.ID).
			Updates(map[string]interface{}{
				"fcm_token":   fcmToken,
				"device_name": deviceName,
				"device_id":   deviceID,
				"updated_at":  now,
			}).Error
	}

	// belum ada record, buat baru
	return database.DB.Table("user_fcm_tokens").
		Create(map[string]interface{}{
			"user_id":     userID,
			"fcm_token":   fcmToken,
			"device_name": deviceName,
			"device_id":   deviceID,
			"created_at":  now,
			"updated_at":  now,
		}).Error
}
