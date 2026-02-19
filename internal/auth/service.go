package auth

import (
	"errors"
	"fmt"
	"log"
	"teka-api/internal/models"
	"teka-api/pkg/database"
	"teka-api/pkg/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SendOtpToEmail(email string) error {
	otp := utils.GenerateOTP()
	log.Printf("🔑 Generated OTP for %s: %s", email, otp) // TAMBAHKAN INI

	if err := SaveOTP(email, otp); err != nil {
		log.Printf("❌ Failed save OTP to DB: %v", err) // TAMBAHKAN INI
		return err
	}
	log.Println("✅ OTP saved to DB") // TAMBAHKAN INI

	body := fmt.Sprintf(`
		<h2>Your OTP Code</h2>
		<p style="font-size:20px;"><b>%s</b></p>
		<p>Expires in 5 minutes.</p>
	`, otp)

	log.Printf("📧 Sending email to %s...", email) // TAMBAHKAN INI
	if err := utils.SendEmail(email, "Your OTP Code", body); err != nil {
		log.Printf("❌ Failed send email: %v", err) // TAMBAHKAN INI
		return err
	}
	log.Println("✅ Email sent successfully") // TAMBAHKAN INI

	return nil
}

func VerifyAndRegister(email, otp string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {

		// 1. Verify OTP
		if err := VerifyOtpTx(tx, email, otp); err != nil {
			return errors.New("OTP invalid or expired")
		}

		// 2. Get temp user
		temp, err := GetTempUserTx(tx, email)
		if err != nil {
			return errors.New("no pending registration")
		}

		// 3. Create real user
		user := models.User{
			Nama:      temp.Nama,
			Phone:     temp.Phone,
			Email:     temp.Email,
			Password:  temp.Password,
			CreatedAt: utils.NowJakarta(),
			UpdatedAt: utils.NowJakarta(),
		}

		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		// 4. Assign role (PAKAI GORM → schema aman)
		if err := tx.Create(&models.UserRole{
			UserID: user.ID,
			RoleID: 1, // customer
			Active: true,
		}).Error; err != nil {
			return err
		}

		// 5. Delete temp user
		if err := DeleteTempUserTx(tx, email); err != nil {
			return err
		}

		return nil
	})
}

func Login(email, password string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}

	return &user, nil
}
func TopUpBalance(userID uint, amount int64) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Get current balance
		latestSaldo, err := GetLatestSaldo(userID)
		if err != nil {
			return err
		}

		// 2. Calculate new balance
		newSaldo := latestSaldo + amount

		// 3. Generate transaction no
		trxNo := fmt.Sprintf("TOPUP-%d-%d", time.Now().Unix(), userID)

		// 4. Insert transaction
		trx := &models.SaldoTransaction{
			UserID:        userID,
			TransactionNo: trxNo,
			Category:      "topup",
			Amount:        amount,
			SaldoSetelah:  newSaldo,
			Description:   "Topup saldo customer",
			CreatedAt:     time.Now(),
		}

		return InsertSaldoTransaction(tx, trx)
	})
}

func GetTransactions(userID uint) ([]models.SaldoTransaction, error) {
	return GetTransactionHistory(userID)
}
