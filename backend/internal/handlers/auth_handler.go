package handlers

import (
	"encoding/json"
	"financial-approval-system/internal/config"
	"financial-approval-system/internal/middleware"
	"financial-approval-system/internal/models"
	"financial-approval-system/internal/repository"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userRepo *repository.UserRepository
	cfg      *config.Config
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(userRepo *repository.UserRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, cfg: cfg}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Name == "" || req.Email == "" || req.Password == "" {
		http.Error(w, `{"error": "Name, email, and password are required"}`, http.StatusBadRequest)
		return
	}

	if len(req.Password) < 6 {
		http.Error(w, `{"error": "Password must be at least 6 characters"}`, http.StatusBadRequest)
		return
	}

	// Check if user already exists
	existingUser, err := h.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		http.Error(w, `{"error": "Email already registered"}`, http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error": "Failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	// Create user
	user, err := h.userRepo.CreateUser(req.Name, req.Email, string(hashedPassword))
	if err != nil {
		http.Error(w, `{"error": "Failed to create user"}`, http.StatusInternalServerError)
		return
	}

	// Get user roles
	roles, _ := h.userRepo.GetUserRoles(user.ID)

	// Generate JWT token
	token, err := middleware.GenerateToken(h.cfg, user.ID, user.Email, roles)
	if err != nil {
		http.Error(w, `{"error": "Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	response := models.LoginResponse{
		Token: token,
		User: models.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Roles:     roles,
			CreatedAt: user.CreatedAt,
		},
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Login handles user authentication
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error": "Email and password are required"}`, http.StatusBadRequest)
		return
	}

	// Find user
	user, err := h.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, `{"error": "Invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, `{"error": "Invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	// Get user roles
	roles, _ := h.userRepo.GetUserRoles(user.ID)

	// Generate JWT token
	token, err := middleware.GenerateToken(h.cfg, user.ID, user.Email, roles)
	if err != nil {
		http.Error(w, `{"error": "Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	response := models.LoginResponse{
		Token: token,
		User: models.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Roles:     roles,
			CreatedAt: user.CreatedAt,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// GetUsers handles listing all users (admin/manager only)
func (h *AuthHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	users, err := h.userRepo.GetAllUsers()
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch users"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}

// AssignRoles handles assigning roles to a user
func (h *AuthHandler) AssignRoles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from URL: /api/users/{id}/roles
	path := strings.TrimPrefix(r.URL.Path, "/api/users/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "roles" {
		http.Error(w, `{"error": "Invalid URL"}`, http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, `{"error": "Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	var req models.AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if len(req.RoleNames) == 0 {
		http.Error(w, `{"error": "At least one role is required"}`, http.StatusBadRequest)
		return
	}

	// Validate role names
	validRoles := map[string]bool{
		models.RoleEmployee: true,
		models.RoleManager:  true,
		models.RoleFinance:  true,
		models.RoleDirector: true,
	}
	for _, role := range req.RoleNames {
		if !validRoles[role] {
			http.Error(w, `{"error": "Invalid role: `+role+`"}`, http.StatusBadRequest)
			return
		}
	}

	// Check if target user exists
	user, err := h.userRepo.GetUserByID(userID)
	if err != nil || user == nil {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}

	// Assign roles
	if err := h.userRepo.AssignRoles(userID, req.RoleNames); err != nil {
		http.Error(w, `{"error": "Failed to assign roles"}`, http.StatusInternalServerError)
		return
	}

	// Return updated user
	roles, _ := h.userRepo.GetUserRoles(userID)
	response := models.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Roles:     roles,
		CreatedAt: user.CreatedAt,
	}

	json.NewEncoder(w).Encode(response)
}

// GetRoles handles listing all available roles
func (h *AuthHandler) GetRoles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	roles, err := h.userRepo.GetAllRoles()
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch roles"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(roles)
}

// GetMe returns the current authenticated user's profile
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	user, err := h.userRepo.GetUserByID(userID)
	if err != nil || user == nil {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}

	roles, _ := h.userRepo.GetUserRoles(user.ID)
	response := models.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Roles:     roles,
		CreatedAt: user.CreatedAt,
	}

	json.NewEncoder(w).Encode(response)
}
