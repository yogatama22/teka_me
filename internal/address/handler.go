package address

import (
	"strconv"
	"teka-api/internal/models"

	"github.com/gofiber/fiber/v2"
)

func GetUserAddresses(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	addresses, err := GetAllByUserID(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var resp []models.UserAddressResponse
	for _, addr := range addresses {
		resp = append(resp, models.MapAddressToResponse(addr))
	}

	return c.JSON(resp)
}

func CreateUserAddress(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var body struct {
		TypeID     uint   `json:"type_id"`
		Address    string `json:"address"`
		City       string `json:"city"`
		PostalCode string `json:"postal_code"`
		Phone      string `json:"phone"`
		IsPrimary  bool   `json:"is_primary"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	addr := models.UserAddress{
		UserID:     userID,
		TypeID:     body.TypeID,
		Address:    body.Address,
		City:       body.City,
		PostalCode: body.PostalCode,
		Phone:      body.Phone,
		IsPrimary:  body.IsPrimary,
	}

	if err := SetPrimary(userID, &addr); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if err := Create(&addr); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(models.MapAddressToResponse(addr))
}

func UpdateUserAddress(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	addrID, _ := strconv.Atoi(c.Params("id"))

	var addr models.UserAddress
	var err error
	if addr, err = GetByID(userID, uint(addrID)); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "address not found"})
	}

	var body struct {
		TypeID     uint   `json:"type_id"`
		Address    string `json:"address"`
		City       string `json:"city"`
		PostalCode string `json:"postal_code"`
		Phone      string `json:"phone"`
		IsPrimary  bool   `json:"is_primary"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	addr.TypeID = body.TypeID
	addr.Address = body.Address
	addr.City = body.City
	addr.PostalCode = body.PostalCode
	addr.Phone = body.Phone
	addr.IsPrimary = body.IsPrimary

	if err := SetPrimary(userID, &addr); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if err := Update(&addr); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(models.MapAddressToResponse(addr))
}

func DeleteUserAddress(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	addrID, _ := strconv.Atoi(c.Params("id"))

	if err := Delete(userID, uint(addrID)); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "address deleted"})
}

func CreateAddressType(c *fiber.Ctx) error {
	var body struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	t := models.AddressType{Name: body.Name}

	if err := CreateType(&t); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(t)
}

func GetAddressTypes(c *fiber.Ctx) error {
	types, err := GetAllTypes()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(types)
}

func DeleteAddressType(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	if err := DeleteType(uint(id)); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "address type deleted"})
}
