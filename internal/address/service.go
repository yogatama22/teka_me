package address

import "teka-api/internal/models"

func SetPrimary(userID uint, addr *models.UserAddress) error {
	if addr.IsPrimary {
		// reset semua primary lain
		if err := ResetPrimary(userID); err != nil {
			return err
		}
	}
	return nil
}
