-- Seed sample categories
INSERT INTO categories (id, name, description)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'Electronics', 'Gadgets and devices'),
    ('00000000-0000-0000-0000-000000000002', 'Books', 'Fiction and non-fiction'),
    ('00000000-0000-0000-0000-000000000003', 'Home', 'Home and kitchen items')
ON CONFLICT DO NOTHING;

-- Seed sample products
INSERT INTO products (id, name, description, price, stock, created_at, updated_at)
VALUES
    ('10000000-0000-0000-0000-000000000001', 'Smartphone', 'Latest model smartphone', 69900, 50, NOW(), NOW()),
    ('10000000-0000-0000-0000-000000000002', 'Cooking Pot', 'Non-stick cooking pot', 4500, 120, NOW(), NOW()),
    ('10000000-0000-0000-0000-000000000003', 'Novel', 'Bestselling novel', 1999, 200, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Map products to categories
INSERT INTO product_category (product_id, category_id)
VALUES
    ('10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001'),
    ('10000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000003'),
    ('10000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000002')
ON CONFLICT DO NOTHING;

-- Seed sample users (password: password123)
INSERT INTO users (id, email, full_name, password_hash, role, status, is_verified, created_at, updated_at)
VALUES
    ('20000000-0000-0000-0000-000000000001', 'user1@example.com', 'User One', '$2a$10$UW9DVBNAS8DTIShJMFb1rO9bB46fQoQcq8cFPIwQkjnMQiONw0m7a', 'user', 'active', true, NOW(), NOW()),
    ('20000000-0000-0000-0000-000000000002', 'user2@example.com', 'User Two', '$2a$10$UW9DVBNAS8DTIShJMFb1rO9bB46fQoQcq8cFPIwQkjnMQiONw0m7a', 'client', 'active', true, NOW(), NOW())
ON CONFLICT DO NOTHING;
