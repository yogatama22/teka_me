package firebase

import "teka-api/pkg/database"

func DeleteFCMToken(token string) error {
	return database.DB.
		Table("user_fcm_tokens").
		Where("fcm_token = ?", token).
		Delete(nil).Error
}
