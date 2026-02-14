-- Check existing orders
SELECT id, customer_id, mitra_id, status_id, created_at 
FROM service_orders 
ORDER BY id DESC 
LIMIT 10;

-- Check if customer and mitra exist
SELECT id, nama, email, phone FROM users WHERE id IN (6, 7);

-- If Order 1 doesn't exist, create it
-- Make sure to adjust this based on your actual table structure
INSERT INTO service_orders (
    id,
    request_id,
    customer_id,
    mitra_id,
    status_id,
    created_at,
    updated_at
) VALUES (
    1,
    1,  -- You may need to create a customer_request first
    7,  -- Customer: Ranti Issara
    6,  -- Mitra: Andri Prasutio
    2,  -- Assuming 2 = ongoing/active status
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE SET
    customer_id = 7,
    mitra_id = 6,
    updated_at = NOW();

-- Verify the order was created
SELECT id, request_id, customer_id, mitra_id, status_id 
FROM service_orders 
WHERE id = 1;
