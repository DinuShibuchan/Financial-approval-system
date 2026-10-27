package repository

import (
	"database/sql"
	"fmt"
	"financial-approval-system/internal/models"
)

// UserRepository handles all user-related database operations
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser inserts a new user into the database
func (r *UserRepository) CreateUser(name, email, hashedPassword string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(
		`INSERT INTO users (name, email, password) VALUES ($1, $2, $3) 
		 RETURNING id, name, email, created_at, updated_at`,
		name, email, hashedPassword,
	).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Assign default EMPLOYEE role
	_, err = r.db.Exec(
		`INSERT INTO user_roles (user_id, role_id) 
		 SELECT $1, id FROM roles WHERE role_name = 'EMPLOYEE'`,
		user.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to assign default role: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by their email address
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(
		`SELECT id, name, email, password, created_at, updated_at FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

// GetUserByID retrieves a user by their ID
func (r *UserRepository) GetUserByID(id int) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(
		`SELECT id, name, email, password, created_at, updated_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return user, nil
}

// GetUserRoles retrieves all role names for a given user
func (r *UserRepository) GetUserRoles(userID int) ([]string, error) {
	rows, err := r.db.Query(
		`SELECT r.role_name FROM roles r
		 INNER JOIN user_roles ur ON r.id = ur.role_id
		 WHERE ur.user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// AssignRoles assigns roles to a user (replaces existing roles)
func (r *UserRepository) AssignRoles(userID int, roleNames []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Remove existing roles
	_, err = tx.Exec(`DELETE FROM user_roles WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to remove existing roles: %w", err)
	}

	// Assign new roles
	for _, roleName := range roleNames {
		_, err = tx.Exec(
			`INSERT INTO user_roles (user_id, role_id) 
			 SELECT $1, id FROM roles WHERE role_name = $2`,
			userID, roleName,
		)
		if err != nil {
			return fmt.Errorf("failed to assign role %s: %w", roleName, err)
		}
	}

	return tx.Commit()
}

// GetAllUsers retrieves all users with their roles
func (r *UserRepository) GetAllUsers() ([]models.UserResponse, error) {
	rows, err := r.db.Query(
		`SELECT id, name, email, created_at FROM users ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	defer rows.Close()

	var users []models.UserResponse
	for rows.Next() {
		var user models.UserResponse
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt); err != nil {
			return nil, err
		}
		// Get roles for each user
		roles, err := r.GetUserRoles(user.ID)
		if err != nil {
			return nil, err
		}
		user.Roles = roles
		users = append(users, user)
	}
	return users, nil
}

// GetAllRoles retrieves all available roles
func (r *UserRepository) GetAllRoles() ([]models.Role, error) {
	rows, err := r.db.Query(`SELECT id, role_name FROM roles ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("failed to get all roles: %w", err)
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.RoleName); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}
