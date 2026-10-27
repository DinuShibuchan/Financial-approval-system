-- SQLite Initial Schema and Seed Data for Multi-Level Financial Approval System

-- Enable Foreign Keys in SQLite
PRAGMA foreign_keys = ON;

-- 1. Users table
CREATE TABLE IF NOT EXISTS users (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL,
    email       TEXT UNIQUE NOT NULL,
    password    TEXT NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Roles table
CREATE TABLE IF NOT EXISTS roles (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    role_name   TEXT UNIQUE NOT NULL
);

-- Seed default roles
INSERT OR IGNORE INTO roles (id, role_name) VALUES 
(1, 'EMPLOYEE'),
(2, 'MANAGER'),
(3, 'FINANCE'),
(4, 'DIRECTOR');

-- 3. User-Roles junction table
CREATE TABLE IF NOT EXISTS user_roles (
    user_id     INTEGER NOT NULL,
    role_id     INTEGER NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

-- 4. Requests table
CREATE TABLE IF NOT EXISTS requests (
    id                      INTEGER PRIMARY KEY AUTOINCREMENT,
    request_number          TEXT UNIQUE NOT NULL,
    request_type            TEXT NOT NULL CHECK (request_type IN ('EXPENSE', 'PURCHASE', 'TRAVEL')),
    requested_by            INTEGER REFERENCES users(id),
    amount                  REAL NOT NULL CHECK (amount > 0),
    description             TEXT,
    status                  TEXT DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'PENDING_APPROVAL', 'APPROVED', 'REJECTED', 'CANCELLED')),
    current_approval_level  INTEGER DEFAULT 0,
    created_at              TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 5. Approval History table
CREATE TABLE IF NOT EXISTS approval_history (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    request_id      INTEGER REFERENCES requests(id) ON DELETE CASCADE,
    approval_level  INTEGER NOT NULL,
    approved_by     INTEGER REFERENCES users(id),
    action          TEXT NOT NULL CHECK (action IN ('APPROVED', 'REJECTED')),
    reason          TEXT,
    acted_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_requests_status ON requests(status);
CREATE INDEX IF NOT EXISTS idx_requests_requested_by ON requests(requested_by);
CREATE INDEX IF NOT EXISTS idx_requests_request_type ON requests(request_type);
CREATE INDEX IF NOT EXISTS idx_approval_history_request_id ON approval_history(request_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);

-- Seed Default Users (Password is 'password123' for all)
INSERT OR IGNORE INTO users (id, name, email, password) VALUES
(1, 'Alice Employee', 'employee@company.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi'),
(2, 'Bob Manager', 'manager@company.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi'),
(3, 'Carol Finance', 'finance@company.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi'),
(4, 'David Director', 'director@company.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi');

-- Assign User Roles
-- Alice: EMPLOYEE
INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES (1, 1);
-- Bob: EMPLOYEE (1), MANAGER (2)
INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES (2, 1), (2, 2);
-- Carol: EMPLOYEE (1), FINANCE (3)
INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES (3, 1), (3, 3);
-- David: EMPLOYEE (1), DIRECTOR (4)
INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES (4, 1), (4, 4);

-- Seed Sample Financial Requests across all 3 tiers
-- Request 1: Tier 1 (Manager only)
INSERT OR IGNORE INTO requests (id, request_number, request_type, requested_by, amount, description, status, current_approval_level, created_at, updated_at)
VALUES (1, 'REQ-20261028-0001', 'EXPENSE', 1, 4500.00, 'Team lunch celebration following quarterly milestone completion', 'PENDING_APPROVAL', 1, datetime('now', '-2 days'), datetime('now'));

-- Request 2: Tier 2 (Manager -> Finance)
INSERT OR IGNORE INTO requests (id, request_number, request_type, requested_by, amount, description, status, current_approval_level, created_at, updated_at)
VALUES (2, 'REQ-20261028-0002', 'PURCHASE', 1, 28500.00, 'Dell UltraSharp 27-inch 4K ergonomic monitors for frontend development team', 'PENDING_APPROVAL', 2, datetime('now', '-3 days'), datetime('now'));

-- Request 3: Tier 3 (Manager -> Finance -> Director)
INSERT OR IGNORE INTO requests (id, request_number, request_type, requested_by, amount, description, status, current_approval_level, created_at, updated_at)
VALUES (3, 'REQ-20261028-0003', 'TRAVEL', 2, 85000.00, 'International Tech Conference attendance, flights, and accommodation in San Francisco', 'PENDING_APPROVAL', 3, datetime('now', '-4 days'), datetime('now'));

-- Request 4: Approved Request
INSERT OR IGNORE INTO requests (id, request_number, request_type, requested_by, amount, description, status, current_approval_level, created_at, updated_at)
VALUES (4, 'REQ-20261028-0004', 'PURCHASE', 1, 8000.00, 'Annual JetBrains All Products development license renewal', 'APPROVED', 1, datetime('now', '-7 days'), datetime('now', '-5 days'));

-- Request 5: Rejected Request
INSERT OR IGNORE INTO requests (id, request_number, request_type, requested_by, amount, description, status, current_approval_level, created_at, updated_at)
VALUES (5, 'REQ-20261028-0005', 'EXPENSE', 1, 15000.00, 'Executive luxury airport limousine transport', 'REJECTED', 1, datetime('now', '-10 days'), datetime('now', '-9 days'));

-- Seed Approval History Audit Trail
INSERT OR IGNORE INTO approval_history (id, request_id, approval_level, approved_by, action, reason, acted_at) VALUES
(1, 2, 1, 2, 'APPROVED', 'Verified necessity for development productivity', datetime('now', '-2 days')),
(2, 3, 1, 2, 'APPROVED', 'Approved conference trip for leadership representation', datetime('now', '-3 days')),
(3, 3, 2, 3, 'APPROVED', 'Budget allocated under Q4 Corporate Training & Travel', datetime('now', '-1 day')),
(4, 4, 1, 2, 'APPROVED', 'Approved essential IDE tooling', datetime('now', '-5 days')),
(5, 5, 1, 2, 'REJECTED', 'Exceeds standard corporate transport policy limits. Please utilize pre-approved taxi voucher.', datetime('now', '-9 days'));
