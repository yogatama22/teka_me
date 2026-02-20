package firebase

import "teka-api/pkg/database"

func DeleteFCMToken(token string) error {
	return database.DB.
		Table("user_fcm_tokens").
		Where("fcm_token = ?", token).
		Delete(nil).Error
}

// GetFCMTokensByUserID retrieves FCM tokens for a specific user
func GetFCMTokensByUserID(userID uint) ([]string, error) {
	var rows []struct {
		FCMToken string `gorm:"column:fcm_token"`
	}

	err := database.DB.
		Table("user_fcm_tokens").
		Select("fcm_token").
		Where("user_id = ?", userID).
		Find(&rows).Error

	if err != nil {
		return nil, err
	}

	tokens := make([]string, 0)
	for _, r := range rows {
		if r.FCMToken != "" {
			tokens = append(tokens, r.FCMToken)
		}
	}
	return tokens, nil
}
