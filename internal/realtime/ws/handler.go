package ws

import (
"context"
"encoding/json"
"fmt"
"log"
"strconv"
"sync"
"teka-api/internal/models"
"teka-api/internal/realtime/redis"
"teka-api/pkg/database"
"time"

"github.com/gofiber/fiber/v2"
"github.com/gofiber/websocket/v2"
)

var (
OrderClients = make(map[string]map[*websocket.Conn]bool)
ChatClients  = make(map[string]map[*websocket.Conn]bool)
clientsMu    sync.RWMutex
)

func WebSocketHandler(c *websocket.Conn) {
orderID := c.Params("orderID")

clientsMu.Lock()
if _, ok := OrderClients[orderID]; !ok {
OrderClients[orderID] = make(map[*websocket.Conn]bool)
}
OrderClients[orderID][c] = true
clientsMu.Unlock()

log.Println("WebSocket connected (Order):", orderID)

defer func() {
clientsMu.Lock()
delete(OrderClients[orderID], c)
clientsMu.Unlock()
c.Close()
log.Println("WebSocket disconnected (Order):", orderID)
}()

for {
if _, _, err := c.ReadMessage(); err != nil {
break
}
}
}

func ChatWebSocketHandler(c *websocket.Conn) {
orderIDStr := c.Params("orderID")
log.Printf("🔌 [STEP 1] Chat WebSocket connection attempt: orderID=%s\n", orderIDStr)

orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
if err != nil {
log.Printf("❌ [STEP 2] Invalid order ID: %v\n", err)
c.WriteJSON(fiber.Map{"error": "invalid order ID"})
c.Close()
return
}
log.Printf("✅ [STEP 2] Order ID parsed: %d\n", orderID)

// Get user info from locals (set by JWT middleware)
userIDVal := c.Locals("user_id")
if userIDVal == nil {
log.Printf("❌ [STEP 3] Unauthorized: user_id not found in locals\n")
c.WriteJSON(fiber.Map{"error": "unauthorized"})
c.Close()
return
}
userID := int64(userIDVal.(uint))
log.Printf("✅ [STEP 3] User authenticated: userID=%d\n", userID)

// 🔍 Deduce sender type and validate membership in the order
log.Printf("🔍 [STEP 4] Querying database for order %d...\n", orderID)
var order models.ServiceOrder
err = database.DB.Select("mitra_id, customer_id").Where("id = ?", orderID).First(&order).Error
if err != nil {
log.Printf("❌ [STEP 4] Order not found: %v\n", err)
c.WriteJSON(fiber.Map{"error": "order not found"})
c.Close()
return
}
log.Printf("✅ [STEP 4] Order found: customerID=%d, mitraID=%d\n", order.CustomerID, order.MitraID)

senderType := ""
if userID == order.MitraID {
senderType = "mitra"
} else if userID == order.CustomerID {
senderType = "customer"
} else {
log.Printf("❌ [STEP 5] User %d not authorized for order %d (customer=%d, mitra=%d)\n",
userID, orderID, order.CustomerID, order.MitraID)
c.WriteJSON(fiber.Map{"error": "not authorized for this order"})
c.Close()
return
}
log.Printf("✅ [STEP 5] User authorized as: %s\n", senderType)

clientsMu.Lock()
if _, ok := ChatClients[orderIDStr]; !ok {
ChatClients[orderIDStr] = make(map[*websocket.Conn]bool)
}
ChatClients[orderIDStr][c] = true
clientCount := len(ChatClients[orderIDStr])
clientsMu.Unlock()

log.Printf("✅ [STEP 6] Chat WebSocket connected: order=%s user=%d type=%s (total clients: %d)\n",
orderIDStr, userID, senderType, clientCount)

// 1️⃣ Load History from Redis (if available)
ctx := context.Background()
cacheKey := fmt.Sprintf("order_chat:%s", orderIDStr)
if redis.Rdb != nil {
log.Printf("📜 [STEP 7] Loading chat history from Redis...\n")
history, err := redis.Rdb.LRange(ctx, cacheKey, 0, -1).Result()
if err == nil {
log.Printf("📜 [STEP 7] Loaded %d messages from history for order %s\n", len(history), orderIDStr)
for _, msgStr := range history {
var msg models.ChatMessage
if err := json.Unmarshal([]byte(msgStr), &msg); err == nil {
if err := c.WriteJSON(msg); err != nil {
log.Printf("⚠️ Error sending history message: %v\n", err)
break
}
}
}
} else {
log.Printf("⚠️ [STEP 7] Error loading chat history: %v\n", err)
}
} else {
log.Printf("⚠️ [STEP 7] Redis not available, skipping history\n")
}

defer func() {
clientsMu.Lock()
delete(ChatClients[orderIDStr], c)
remainingClients := len(ChatClients[orderIDStr])
clientsMu.Unlock()
c.Close()
log.Printf("🔌 Chat WebSocket disconnected: order=%s user=%d (remaining clients: %d)\n",
orderIDStr, userID, remainingClients)
}()

log.Printf("🎧 [STEP 8] Entering message loop for user %d...\n", userID)
for {
var req struct {
Message string `json:"message"`
}
if err := c.ReadJSON(&req); err != nil {
log.Printf("⚠️ Error reading message from user %d: %v\n", userID, err)
break
}

if req.Message == "" {
continue
}

// 2️⃣ Create Chat Message
msg := models.ChatMessage{
OrderID:    orderID,
SenderID:   userID,
SenderType: senderType,
Message:    req.Message,
CreatedAt:  time.Now(),
}

log.Printf("📨 Received message from user=%d (%s) in order=%s: \"%s\"",
userID, senderType, orderIDStr, req.Message)

// 3️⃣ Save to Redis (if available)
if redis.Rdb != nil {
msgJSON, err := json.Marshal(msg)
if err != nil {
log.Printf("Error marshaling message: %v", err)
} else {
if err := redis.Rdb.RPush(ctx, cacheKey, msgJSON).Err(); err != nil {
log.Printf("Error saving message to Redis: %v", err)
}
if err := redis.Rdb.Expire(ctx, cacheKey, 24*time.Hour).Err(); err != nil {
log.Printf("Error setting Redis expiry: %v", err)
}
}
}

// 4️⃣ Broadcast to all in room
BroadcastChat(orderIDStr, msg)
}
}

func BroadcastChat(orderID string, msg models.ChatMessage) {
clientsMu.Lock()
defer clientsMu.Unlock()

if clients, ok := ChatClients[orderID]; ok {
clientCount := len(clients)
log.Printf("📡 Broadcasting message to %d client(s) in order %s", clientCount, orderID)

successCount := 0
failCount := 0

for conn := range clients {
err := conn.WriteJSON(msg)
if err != nil {
log.Printf("❌ Error sending message to client: %v", err)
conn.Close()
delete(clients, conn)
failCount++
} else {
successCount++
}
}

log.Printf("✅ Broadcast complete: %d succeeded, %d failed", successCount, failCount)
} else {
log.Printf("⚠️ No clients found for order %s in ChatClients map", orderID)
}
}

func PublishStatus(orderID, status string) {
log.Printf("PublishStatus called: orderID=%s, status=%s\n", orderID, status)

clientsMu.RLock()
defer clientsMu.RUnlock()

if clients, ok := OrderClients[orderID]; ok {
for conn := range clients {
err := conn.WriteMessage(websocket.TextMessage, []byte(status))
if err != nil {
log.Println("Error sending WS message:", err)
conn.Close()
}
}
}
}
