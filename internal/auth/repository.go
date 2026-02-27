package auth

import (
	"context"
	"errors"
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
	now := time.Now()

	// 1. Hapus token ini dari user lain (mencegah token digunakan oleh user yang berbeda di device yang sama)
	database.DB.Exec("DELETE FROM user_fcm_tokens WHERE fcm_token = ? AND user_id != ?", fcmToken, userID)

	// 2. Bersihkan SEMUA token lama milik user ini (ensuring single device)
	database.DB.Exec("DELETE FROM user_fcm_tokens WHERE user_id = ?", userID)

	// 3. Masukkan record baru
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

// ================= BALANCE & TRANSACTIONS ====================

func GetLatestSaldo(userID uint) (int64, error) {
	var saldo int64
	err := database.DB.
		Table("myschema.saldo_role_transactions").
		Select("saldo_setelah").
		Where("user_id = ?", userID).
		Order("id DESC").
		Limit(1).
		Scan(&saldo).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	return saldo, err
}

func InsertSaldoTransaction(tx *gorm.DB, trx *models.SaldoTransaction) error {
	return tx.Create(trx).Error
}

func GetTransactionHistory(userID uint) ([]models.SaldoTransaction, error) {
	var results []models.SaldoTransaction
	err := database.DB.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&results).Error
	return results, err
}
func GetCustomerEarningsHistoryFromDB(ctx context.Context, userID uint) (models.EarningMonthlyHistory, error) {
	var rows []models.IncomeDetail

	err := database.DB.WithContext(ctx).Raw(`
		SELECT 
			srt.reference_id as transaction_no,
			srt.amount,
			srt.description,
			stc.code as category_name,
			srt.created_at
		FROM myschema.saldo_role_transactions srt
		JOIN myschema.saldo_transaction_categories stc ON stc.id = srt.category_id
		WHERE srt.user_id = ? 
		  AND srt.category_id IN (1, 2)
		ORDER BY srt.created_at DESC
	`, userID).Scan(&rows).Error

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
