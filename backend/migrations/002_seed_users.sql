-- =====================================================
-- Seed Data for Financial Approval System
-- Default password for all seed accounts: password123
-- Bcrypt hash ($2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi)
-- =====================================================

-- 1. Insert Test Users
INSERT INTO users (id, name, email, password, created_at, updated_at) VALUES
(1, 'Alice Employee', 'employee@company.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', NOW(), NOW()),
(2, 'Bob Manager', 'manager@company.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', NOW(), NOW()),
(3, 'Carol Finance', 'finance@company.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', NOW(), NOW()),
(4, 'David Director', 'director@company.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', NOW(), NOW())
ON CONFLICT (email) DO NOTHING;

-- Reset sequence if needed
SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));

-- 2. Assign User Roles
-- Alice: EMPLOYEE
INSERT INTO user_roles (user_id, role_id) 
SELECT 1, id FROM roles WHERE role_name = 'EMPLOYEE'
ON CONFLICT DO NOTHING;

-- Bob: EMPLOYEE, MANAGER
INSERT INTO user_roles (user_id, role_id) 
SELECT 2, id FROM roles WHERE role_name IN ('EMPLOYEE', 'MANAGER')
ON CONFLICT DO NOTHING;

-- Carol: EMPLOYEE, FINANCE
INSERT INTO user_roles (user_id, role_id) 
SELECT 3, id FROM roles WHERE role_name IN ('EMPLOYEE', 'FINANCE')
ON CONFLICT DO NOTHING;

-- David: EMPLOYEE, DIRECTOR
INSERT INTO user_roles (user_id, role_id) 
SELECT 4, id FROM roles WHERE role_name IN ('EMPLOYEE', 'DIRECTOR')
ON CONFLICT DO NOTHING;

-- 3. Seed Sample Financial Requests across all 3 tiers
-- Sample 1: Low-value expense (< 10,000) -> Tier 1 (Manager only)
INSERT INTO requests (id, request_number, request_type, requested_by, amount, description, status, current_approval_level, created_at, updated_at)
VALUES
(1, 'REQ-20261028-0001', 'EXPENSE', 1, 4500.00, 'Team lunch celebration following quarterly milestone completion', 'PENDING_APPROVAL', 1, NOW() - INTERVAL '2 days', NOW())
ON CONFLICT (request_number) DO NOTHING;

-- Sample 2: Mid-value purchase (10,000 - 50,000) -> Tier 2 (Manager -> Finance)
INSERT INTO requests (id, request_number, request_type, requested_by, amount, description, status, current_approval_level, created_at, updated_at)
VALUES
(2, 'REQ-20261028-0002', 'PURCHASE', 1, 28500.00, 'Dell UltraSharp 27-inch 4K ergonomic monitors for frontend development team', 'PENDING_APPROVAL', 2, NOW() - INTERVAL '3 days', NOW())
ON CONFLICT (request_number) DO NOTHING;

-- Sample 3: High-value capital outlay (> 50,000) -> Tier 3 (Manager -> Finance -> Director)
INSERT INTO requests (id, request_number, request_type, requested_by, amount, description, status, current_approval_level, created_at, updated_at)
VALUES
(3, 'REQ-20261028-0003', 'TRAVEL', 2, 85000.00, 'International Tech Conference attendance, flights, and accommodation in San Francisco', 'PENDING_APPROVAL', 3, NOW() - INTERVAL '4 days', NOW())
ON CONFLICT (request_number) DO NOTHING;

-- Sample 4: Fully Approved Request
INSERT INTO requests (id, request_number, request_type, requested_by, amount, description, status, current_approval_level, created_at, updated_at)
VALUES
(4, 'REQ-20261028-0004', 'PURCHASE', 1, 8000.00, 'Annual JetBrains All Products development license renewal', 'APPROVED', 1, NOW() - INTERVAL '7 days', NOW() - INTERVAL '5 days')
ON CONFLICT (request_number) DO NOTHING;

-- Sample 5: Rejected Request
INSERT INTO requests (id, request_number, request_type, requested_by, amount, description, status, current_approval_level, created_at, updated_at)
VALUES
(5, 'REQ-20261028-0005', 'EXPENSE', 1, 15000.00, 'Executive luxury airport limousine transport', 'REJECTED', 1, NOW() - INTERVAL '10 days', NOW() - INTERVAL '9 days')
ON CONFLICT (request_number) DO NOTHING;

SELECT setval('requests_id_seq', (SELECT MAX(id) FROM requests));

-- 4. Seed Approval History Audit Trail
INSERT INTO approval_history (request_id, approval_level, approved_by, action, reason, acted_at) VALUES
-- For Request 2 (Level 1 Manager approved, now waiting Level 2 Finance)
(2, 1, 2, 'APPROVED', 'Verified necessity for development productivity', NOW() - INTERVAL '2 days'),
-- For Request 3 (Level 1 Manager and Level 2 Finance approved, now waiting Level 3 Director)
(3, 1, 2, 'APPROVED', 'Approved conference trip for leadership representation', NOW() - INTERVAL '3 days'),
(3, 2, 3, 'APPROVED', 'Budget allocated under Q4 Corporate Training & Travel', NOW() - INTERVAL '1 day'),
-- For Request 4 (Approved by Manager)
(4, 1, 2, 'APPROVED', 'Approved essential IDE tooling', NOW() - INTERVAL '5 days'),
-- For Request 5 (Rejected by Manager with mandatory reason)
(5, 1, 2, 'REJECTED', 'Exceeds standard corporate transport policy limits. Please utilize pre-approved taxi voucher.', NOW() - INTERVAL '9 days')
ON CONFLICT DO NOTHING;
