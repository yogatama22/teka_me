package address

import (
	"teka-api/internal/models"
	"teka-api/pkg/database"
)

func GetAllByUserID(userID uint) ([]models.UserAddress, error) {
	var addresses []models.UserAddress
	err := database.DB.Preload("Type").Where("user_id = ?", userID).Find(&addresses).Error
	return addresses, err
}

func GetByID(userID, id uint) (models.UserAddress, error) {
	var addr models.UserAddress
	err := database.DB.Preload("Type").Where("user_id = ? AND id = ?", userID, id).First(&addr).Error
	return addr, err
}

func Create(addr *models.UserAddress) error {
	return database.DB.Create(addr).Error
}

func Update(addr *models.UserAddress) error {
	return database.DB.Save(addr).Error
}

func Delete(userID, id uint) error {
	return database.DB.Where("user_id = ? AND id = ?", userID, id).Delete(&models.UserAddress{}).Error
}

func ResetPrimary(userID uint) error {
	return database.DB.Model(&models.UserAddress{}).Where("user_id = ?", userID).Update("is_primary", false).Error
}

func CreateType(t *models.AddressType) error {
	return database.DB.Create(t).Error
}

func GetAllTypes() ([]models.AddressType, error) {
	var types []models.AddressType
	err := database.DB.Find(&types).Error
	return types, err
}

func DeleteType(id uint) error {
	return database.DB.Delete(&models.AddressType{}, id).Error
}
