-- Migration: Create initial tables for Multi-Level Financial Approval System
-- Run this against your PostgreSQL database

-- Enable UUID extension (optional, we use SERIAL for simplicity)
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- 1. Users table
-- =====================================================
CREATE TABLE IF NOT EXISTS users (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    email       VARCHAR(150) UNIQUE NOT NULL,
    password    VARCHAR(255) NOT NULL,
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW()
);

-- =====================================================
-- 2. Roles table
-- =====================================================
CREATE TABLE IF NOT EXISTS roles (
    id          SERIAL PRIMARY KEY,
    role_name   VARCHAR(50) UNIQUE NOT NULL
);

-- Seed default roles
INSERT INTO roles (role_name) VALUES ('EMPLOYEE') ON CONFLICT (role_name) DO NOTHING;
INSERT INTO roles (role_name) VALUES ('MANAGER') ON CONFLICT (role_name) DO NOTHING;
INSERT INTO roles (role_name) VALUES ('FINANCE') ON CONFLICT (role_name) DO NOTHING;
INSERT INTO roles (role_name) VALUES ('DIRECTOR') ON CONFLICT (role_name) DO NOTHING;

-- =====================================================
-- 3. User-Roles junction table (many-to-many)
-- =====================================================
CREATE TABLE IF NOT EXISTS user_roles (
    user_id     INT REFERENCES users(id) ON DELETE CASCADE,
    role_id     INT REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- =====================================================
-- 4. Requests table (financial requests)
-- =====================================================
CREATE TABLE IF NOT EXISTS requests (
    id                      SERIAL PRIMARY KEY,
    request_number          VARCHAR(30) UNIQUE NOT NULL,
    request_type            VARCHAR(20) NOT NULL CHECK (request_type IN ('EXPENSE', 'PURCHASE', 'TRAVEL')),
    requested_by            INT REFERENCES users(id),
    amount                  DECIMAL(12, 2) NOT NULL CHECK (amount > 0),
    description             TEXT,
    status                  VARCHAR(20) DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'PENDING_APPROVAL', 'APPROVED', 'REJECTED', 'CANCELLED')),
    current_approval_level  INT DEFAULT 0,
    created_at              TIMESTAMP DEFAULT NOW(),
    updated_at              TIMESTAMP DEFAULT NOW()
);

-- =====================================================
-- 5. Approval History table
-- =====================================================
CREATE TABLE IF NOT EXISTS approval_history (
    id              SERIAL PRIMARY KEY,
    request_id      INT REFERENCES requests(id) ON DELETE CASCADE,
    approval_level  INT NOT NULL,
    approved_by     INT REFERENCES users(id),
    action          VARCHAR(20) NOT NULL CHECK (action IN ('APPROVED', 'REJECTED')),
    reason          TEXT,
    acted_at        TIMESTAMP DEFAULT NOW()
);

-- =====================================================
-- Indexes for performance
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_requests_status ON requests(status);
CREATE INDEX IF NOT EXISTS idx_requests_requested_by ON requests(requested_by);
CREATE INDEX IF NOT EXISTS idx_requests_request_type ON requests(request_type);
CREATE INDEX IF NOT EXISTS idx_approval_history_request_id ON approval_history(request_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
