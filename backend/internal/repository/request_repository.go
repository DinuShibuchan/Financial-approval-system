package repository

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"financial-approval-system/internal/models"
)

type RequestRepository struct {
	db *sql.DB
}

func NewRequestRepository(db *sql.DB) *RequestRepository {
	return &RequestRepository{db: db}
}

// GenerateRequestNumber produces a human-friendly unique request code like REQ-20261028-1234
func (r *RequestRepository) GenerateRequestNumber() (string, error) {
	for i := 0; i < 5; i++ {
		dateStr := time.Now().Format("20060102")
		num := rand.Intn(9000) + 1000
		candidate := fmt.Sprintf("REQ-%s-%04d", dateStr, num)

		var exists bool
		err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM requests WHERE request_number = $1)", candidate).Scan(&exists)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	// Fallback to timestamp nanoseconds
	return fmt.Sprintf("REQ-%d", time.Now().UnixNano()/1e6), nil
}

// Create inserts a new request into the database
func (r *RequestRepository) Create(req *models.Request) error {
	reqNum, err := r.GenerateRequestNumber()
	if err != nil {
		return fmt.Errorf("failed to generate request number: %w", err)
	}
	req.RequestNumber = reqNum

	if req.Status == "" {
		req.Status = models.StatusDraft
	}
	if req.Status == models.StatusPendingApproval {
		req.CurrentApprovalLevel = 1
	} else {
		req.CurrentApprovalLevel = 0
	}

	query := `
		INSERT INTO requests (request_number, request_type, requested_by, amount, description, status, current_approval_level, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(
		query,
		req.RequestNumber,
		req.RequestType,
		req.RequestedBy,
		req.Amount,
		req.Description,
		req.Status,
		req.CurrentApprovalLevel,
	).Scan(&req.ID, &req.CreatedAt, &req.UpdatedAt)
}

// GetByID returns a request along with creator name
func (r *RequestRepository) GetByID(id int) (*models.Request, error) {
	query := `
		SELECT r.id, r.request_number, r.request_type, r.requested_by, u.name,
		       r.amount, r.description, r.status, r.current_approval_level,
		       r.created_at, r.updated_at
		FROM requests r
		LEFT JOIN users u ON r.requested_by = u.id
		WHERE r.id = $1
	`
	req := &models.Request{}
	var requestedByName sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&req.ID,
		&req.RequestNumber,
		&req.RequestType,
		&req.RequestedBy,
		&requestedByName,
		&req.Amount,
		&req.Description,
		&req.Status,
		&req.CurrentApprovalLevel,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if requestedByName.Valid {
		req.RequestedByName = requestedByName.String
	}
	return req, nil
}

// GetByUserID returns all requests initiated by a specific user
func (r *RequestRepository) GetByUserID(userID int) ([]*models.Request, error) {
	query := `
		SELECT r.id, r.request_number, r.request_type, r.requested_by, u.name,
		       r.amount, r.description, r.status, r.current_approval_level,
		       r.created_at, r.updated_at
		FROM requests r
		LEFT JOIN users u ON r.requested_by = u.id
		WHERE r.requested_by = $1
		ORDER BY r.created_at DESC
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*models.Request
	for rows.Next() {
		req := &models.Request{}
		var requestedByName sql.NullString
		if err := rows.Scan(
			&req.ID,
			&req.RequestNumber,
			&req.RequestType,
			&req.RequestedBy,
			&requestedByName,
			&req.Amount,
			&req.Description,
			&req.Status,
			&req.CurrentApprovalLevel,
			&req.CreatedAt,
			&req.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if requestedByName.Valid {
			req.RequestedByName = requestedByName.String
		}
		requests = append(requests, req)
	}
	return requests, nil
}

// GetAll returns requests with optional filtering by status and type
func (r *RequestRepository) GetAll(status, reqType string) ([]*models.Request, error) {
	query := `
		SELECT r.id, r.request_number, r.request_type, r.requested_by, u.name,
		       r.amount, r.description, r.status, r.current_approval_level,
		       r.created_at, r.updated_at
		FROM requests r
		LEFT JOIN users u ON r.requested_by = u.id
		WHERE ($1 = '' OR r.status = $1)
		  AND ($2 = '' OR r.request_type = $2)
		ORDER BY r.created_at DESC
	`
	rows, err := r.db.Query(query, status, reqType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*models.Request
	for rows.Next() {
		req := &models.Request{}
		var requestedByName sql.NullString
		if err := rows.Scan(
			&req.ID,
			&req.RequestNumber,
			&req.RequestType,
			&req.RequestedBy,
			&requestedByName,
			&req.Amount,
			&req.Description,
			&req.Status,
			&req.CurrentApprovalLevel,
			&req.CreatedAt,
			&req.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if requestedByName.Valid {
			req.RequestedByName = requestedByName.String
		}
		requests = append(requests, req)
	}
	return requests, nil
}

// Update modifies a request (only valid if current status is DRAFT)
func (r *RequestRepository) Update(id, userID int, reqType string, amount float64, description string) (*models.Request, error) {
	query := `
		UPDATE requests
		SET request_type = $1, amount = $2, description = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4 AND requested_by = $5 AND status = 'DRAFT'
		RETURNING id, request_number, request_type, requested_by, amount, description, status, current_approval_level, created_at, updated_at
	`
	req := &models.Request{}
	err := r.db.QueryRow(query, reqType, amount, description, id, userID).Scan(
		&req.ID,
		&req.RequestNumber,
		&req.RequestType,
		&req.RequestedBy,
		&req.Amount,
		&req.Description,
		&req.Status,
		&req.CurrentApprovalLevel,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("request not found or cannot be edited (only DRAFT requests by owner can be edited)")
		}
		return nil, err
	}
	return req, nil
}

// Submit changes status from DRAFT to PENDING_APPROVAL and sets current_approval_level = 1
func (r *RequestRepository) Submit(id, userID int) (*models.Request, error) {
	query := `
		UPDATE requests
		SET status = 'PENDING_APPROVAL', current_approval_level = 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND requested_by = $2 AND status = 'DRAFT'
		RETURNING id, request_number, request_type, requested_by, amount, description, status, current_approval_level, created_at, updated_at
	`
	req := &models.Request{}
	err := r.db.QueryRow(query, id, userID).Scan(
		&req.ID,
		&req.RequestNumber,
		&req.RequestType,
		&req.RequestedBy,
		&req.Amount,
		&req.Description,
		&req.Status,
		&req.CurrentApprovalLevel,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("request not found or cannot be submitted (only DRAFT requests can be submitted)")
		}
		return nil, err
	}
	return req, nil
}

// Cancel changes status to CANCELLED (allowed for owner if DRAFT or PENDING_APPROVAL)
func (r *RequestRepository) Cancel(id, userID int) (*models.Request, error) {
	query := `
		UPDATE requests
		SET status = 'CANCELLED', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND requested_by = $2 AND status IN ('DRAFT', 'PENDING_APPROVAL')
		RETURNING id, request_number, request_type, requested_by, amount, description, status, current_approval_level, created_at, updated_at
	`
	req := &models.Request{}
	err := r.db.QueryRow(query, id, userID).Scan(
		&req.ID,
		&req.RequestNumber,
		&req.RequestType,
		&req.RequestedBy,
		&req.Amount,
		&req.Description,
		&req.Status,
		&req.CurrentApprovalLevel,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("request not found or cannot be cancelled")
		}
		return nil, err
	}
	return req, nil
}

// Delete permanently removes a draft request
func (r *RequestRepository) Delete(id, userID int) error {
	res, err := r.db.Exec("DELETE FROM requests WHERE id = $1 AND requested_by = $2 AND status = 'DRAFT'", id, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("request not found or cannot be deleted (only DRAFT requests can be deleted)")
	}
	return nil
}
