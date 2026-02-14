# WebSocket Chat Testing Guide

## Overview
This guide provides instructions for testing the WebSocket chat functionality for Order #1 with both Customer and Mitra connections.

## Fixed Issues in Handler

### Critical Fixes Applied to `internal/realtime/ws/handler.go`:

1. **Line 56-61**: Added error handling for `strconv.ParseInt` - prevents invalid order IDs
2. **Line 109-112**: Added error handling when sending chat history to prevent connection issues
3. **Line 115-116**: Added logging for Redis errors when loading history
4. **Line 149-159**: Added proper error handling for JSON marshaling and Redis operations
5. **Line 166-180**: Fixed `BroadcastChat` to use write lock and properly clean up failed connections

### Improvements:
- Better error logging throughout
- Proper connection cleanup on failures
- No silent failures in Redis operations
- Memory leak prevention in broadcast function

## Test Credentials

### Customer (Ranti Issara)
- **User ID**: 7
- **Token**: `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InJhbnRpaXNzYXJhNzJAZ21haWwuY29tIiwiZXhwIjoxNzcxMTYzNDczLCJpZCI6NywibmFtYSI6IlJhbnRpIElzc2FyYSIsInBob25lIjoiMDgxMzk4ODMyODMxIn0.nGTEDnlz5ztpRRPtnpqe0CmeGtq5vTf-LGpwTlQ-toY`

### Mitra (Andri Prasutio)
- **User ID**: 6
- **Token**: `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImFuZHJpcHJhc3V0aW9AZ21haWwuY29tIiwiZXhwIjoxNzcxMTYzNDMwLCJpZCI6NiwibmFtYSI6IkFuZHJpIFByYXN1dGlvIiwicGhvbmUiOiIwODEzOTg4MzI4MzAifQ.FoYn6r-z0MNBSJX3298mlrjbDAlA_Z9ixYXUZ_-Wocw`

## WebSocket Endpoint
```
ws://localhost:8080/api/realtime/chat/1
```

## Testing Methods

### Method 1: HTML Browser Test (Recommended for Visual Testing)

1. **Start your server**:
   ```bash
   go run main.go
   ```

2. **Open the test page**:
   ```bash
   open test_chat_ws.html
   # or simply double-click the file
   ```

3. **Test steps**:
   - Click "🔌 Connect Customer" to connect as customer
   - Click "🔌 Connect Mitra" to connect as mitra
   - Type messages in either chat box and press Enter or click Send
   - Watch messages appear in both chat windows in real-time
   - Test chat history by disconnecting and reconnecting

### Method 2: Node.js CLI Test (For Automated Testing)

1. **Install dependencies** (if not already installed):
   ```bash
   npm install ws
   ```

2. **Run the test script**:
   ```bash
   node test_chat_ws.js
   ```

3. **Interactive menu options**:
   - `1` - Connect Customer
   - `2` - Connect Mitra
   - `3` - Send message as Customer
   - `4` - Send message as Mitra
   - `5` - Disconnect Customer
   - `6` - Disconnect Mitra
   - `7` - Auto test (automated conversation)
   - `8` - Exit

4. **Recommended test flow**:
   ```
   Select: 7 (Auto test)
   ```
   This will automatically:
   - Connect both clients
   - Send a realistic conversation
   - Demonstrate real-time messaging

### Method 3: Manual WebSocket Testing with wscat

1. **Install wscat**:
   ```bash
   npm install -g wscat
   ```

2. **Connect as Customer**:
   ```bash
   wscat -c "ws://localhost:8080/api/realtime/chat/1" \
     -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InJhbnRpaXNzYXJhNzJAZ21haWwuY29tIiwiZXhwIjoxNzcxMTYzNDczLCJpZCI6NywibmFtYSI6IlJhbnRpIElzc2FyYSIsInBob25lIjoiMDgxMzk4ODMyODMxIn0.nGTEDnlz5ztpRRPtnpqe0CmeGtq5vTf-LGpwTlQ-toY"
   ```

3. **Send a message**:
   ```json
   {"message": "Hello from customer!"}
   ```

4. **In another terminal, connect as Mitra**:
   ```bash
   wscat -c "ws://localhost:8080/api/realtime/chat/1" \
     -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImFuZHJpcHJhc3V0aW9AZ21haWwuY29tIiwiZXhwIjoxNzcxMTYzNDMwLCJpZCI6NiwibmFtYSI6IkFuZHJpIFByYXN1dGlvIiwicGhvbmUiOiIwODEzOTg4MzI4MzAifQ.FoYn6r-z0MNBSJX3298mlrjbDAlA_Z9ixYXUZ_-Wocw"
   ```

## Expected Behavior

### On Connection:
- Server logs: `🔌 Attempting Chat WebSocket connection: orderID=1`
- Server logs: `Chat WebSocket connected: order=1 user=X type=customer/mitra`
- Client receives chat history (if any exists in Redis)

### On Message Send:
- Sender sees their message immediately
- Message is saved to Redis with 24-hour expiry
- All connected clients in the same order room receive the message
- Message format:
  ```json
  {
    "order_id": 1,
    "sender_id": 7,
    "sender_type": "customer",
    "message": "Hello!",
    "created_at": "2026-02-14T14:27:00Z"
  }
  ```

### On Disconnect:
- Server logs: `Chat WebSocket disconnected: 1`
- Client removed from active connections map

## Testing Checklist

- [ ] Customer can connect successfully
- [ ] Mitra can connect successfully
- [ ] Customer can send messages
- [ ] Mitra can send messages
- [ ] Messages appear in real-time for both parties
- [ ] Chat history loads on reconnection
- [ ] Invalid order ID returns error
- [ ] Unauthorized user (wrong token) cannot connect
- [ ] User not part of order cannot connect
- [ ] Connection cleanup works properly
- [ ] Redis saves messages correctly
- [ ] Messages expire after 24 hours

## Troubleshooting

### Connection Refused
- Ensure server is running on port 8080
- Check if Redis is running
- Verify database connection

### Unauthorized Error
- Check if JWT tokens are valid and not expired
- Verify JWT middleware is properly configured
- Ensure user IDs match order participants

### Messages Not Appearing
- Check server logs for errors
- Verify Redis is accessible
- Check if both clients are connected to the same order ID

### Token Expiry
If tokens expire, decode them to check expiry:
```bash
# Customer token expires: 2026-02-16 (timestamp: 1771163473)
# Mitra token expires: 2026-02-16 (timestamp: 1771163430)
```

## Architecture Notes

### WebSocket Flow:
1. Client connects with JWT token in Authorization header
2. Server validates token via JWT middleware
3. Server checks if user is part of the order (customer or mitra)
4. Server adds client to `ChatClients` map
5. Server sends chat history from Redis
6. Client can send/receive messages in real-time
7. All messages broadcast to all clients in the same order room

### Data Storage:
- **Redis Key**: `order_chat:{orderID}`
- **Expiry**: 24 hours
- **Format**: JSON array of ChatMessage objects

### Security:
- JWT authentication required
- User must be either customer or mitra of the order
- Order membership validated against database

## Files Created

1. `test_chat_ws.html` - Beautiful browser-based testing interface
2. `test_chat_ws.js` - Node.js CLI testing script
3. `WEBSOCKET_TEST_README.md` - This documentation

## Next Steps

1. Test with the provided scripts
2. Verify all functionality works as expected
3. Test edge cases (invalid tokens, wrong order IDs, etc.)
4. Monitor server logs for any errors
5. Consider adding unit tests for the handler functions
