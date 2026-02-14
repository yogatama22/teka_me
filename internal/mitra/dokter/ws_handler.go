package dokter

import (
	"strconv"

	"github.com/gofiber/websocket/v2"
)

func (h *Handler) OrderStatusWS(c *websocket.Conn) {
	orderIDStr := c.Params("order_id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		_ = c.Close()
		return
	}

	h.Hub.Join(orderID, c)
	defer h.Hub.Leave(orderID, c)

	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}
}
