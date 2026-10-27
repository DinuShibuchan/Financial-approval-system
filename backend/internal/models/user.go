package models

import "time"

// User represents a system user
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Never expose password in JSON
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserResponse is the safe user representation for API responses
type UserResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
}

// Role represents a system role
type Role struct {
	ID       int    `json:"id"`
	RoleName string `json:"role_name"`
}

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	UserID int `json:"user_id"`
	RoleID int `json:"role_id"`
}

// RegisterRequest is the payload for user registration
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest is the payload for user login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse is returned after successful login
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// AssignRoleRequest is the payload for assigning roles to a user
type AssignRoleRequest struct {
	RoleNames []string `json:"role_names"`
}

// Predefined role constants
const (
	RoleEmployee = "EMPLOYEE"
	RoleManager  = "MANAGER"
	RoleFinance  = "FINANCE"
	RoleDirector = "DIRECTOR"
)
