package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"financial-approval-system/internal/models"
)

type ApprovalRepository struct {
	db *sql.DB
}

func NewApprovalRepository(db *sql.DB) *ApprovalRepository {
	return &ApprovalRepository{db: db}
}

// GetApprovalHistory retrieves the full audit log of actions taken on a request
func (r *ApprovalRepository) GetApprovalHistory(requestID int) ([]*models.ApprovalHistory, error) {
	query := `
		SELECT h.id, h.request_id, h.approval_level, h.approved_by, u.name,
		       h.action, COALESCE(h.reason, ''), h.acted_at
		FROM approval_history h
		LEFT JOIN users u ON h.approved_by = u.id
		WHERE h.request_id = $1
		ORDER BY h.acted_at ASC, h.id ASC
	`
	rows, err := r.db.Query(query, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*models.ApprovalHistory
	for rows.Next() {
		entry := &models.ApprovalHistory{}
		var approverName sql.NullString
		if err := rows.Scan(
			&entry.ID,
			&entry.RequestID,
			&entry.ApprovalLevel,
			&entry.ApprovedBy,
			&approverName,
			&entry.Action,
			&entry.Reason,
			&entry.ActedAt,
		); err != nil {
			return nil, err
		}
		if approverName.Valid {
			entry.ApprovedByName = approverName.String
		}
		history = append(history, entry)
	}
	return history, nil
}

// GetPendingApprovalsForUser returns all requests waiting for approval by the caller based on their roles
func (r *ApprovalRepository) GetPendingApprovalsForUser(userID int, userRoles []string) ([]*models.Request, error) {
	// Determine which approval levels this user can act on
	var allowedLevels []int
	for _, role := range userRoles {
		switch strings.ToUpper(role) {
		case models.RoleManager:
			allowedLevels = append(allowedLevels, models.ApprovalLevelManager) // Level 1
		case models.RoleFinance:
			allowedLevels = append(allowedLevels, models.ApprovalLevelFinance) // Level 2
		case models.RoleDirector:
			allowedLevels = append(allowedLevels, models.ApprovalLevelDirector) // Level 3
		}
	}

	if len(allowedLevels) == 0 {
		return []*models.Request{}, nil
	}

	// Build placeholders for allowedLevels IN ($1, $2, ...)
	placeholders := make([]string, len(allowedLevels))
	args := []interface{}{userID} // $1 is userID to enforce requested_by != userID
	for i, lvl := range allowedLevels {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, lvl)
	}

	query := fmt.Sprintf(`
		SELECT r.id, r.request_number, r.request_type, r.requested_by, u.name,
		       r.amount, r.description, r.status, r.current_approval_level,
		       r.created_at, r.updated_at
		FROM requests r
		LEFT JOIN users u ON r.requested_by = u.id
		WHERE r.status = 'PENDING_APPROVAL'
		  AND r.requested_by != $1
		  AND r.current_approval_level IN (%s)
		ORDER BY r.created_at ASC
	`, strings.Join(placeholders, ", "))

	rows, err := r.db.Query(query, args...)
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

// ProcessApproval handles approving or rejecting a request with full business rule enforcement
func (r *ApprovalRepository) ProcessApproval(requestID, userID int, userRoles []string, action, reason string) (*models.Request, error) {
	action = strings.ToUpper(strings.TrimSpace(action))
	if action != "APPROVED" && action != "REJECTED" {
		return nil, fmt.Errorf("invalid action. Must be APPROVED or REJECTED")
	}

	// Rule 4: Rejection requires a reason
	if action == "REJECTED" && strings.TrimSpace(reason) == "" {
		return nil, fmt.Errorf("rejection requires a valid reason")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Fetch the request
	queryLock := `
		SELECT id, request_number, request_type, requested_by, amount, description, status, current_approval_level, created_at, updated_at
		FROM requests
		WHERE id = $1
	`
	req := &models.Request{}
	err = tx.QueryRow(queryLock, requestID).Scan(
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
			return nil, fmt.Errorf("request not found")
		}
		return nil, err
	}

	// 2. Validate request status
	if req.Status != models.StatusPendingApproval {
		return nil, fmt.Errorf("request is not in PENDING_APPROVAL status (current status: %s)", req.Status)
	}

	// 3. Rule 5: User cannot approve their own request
	if req.RequestedBy == userID {
		return nil, fmt.Errorf("you cannot approve or reject your own financial request")
	}

	// 4. Validate current approval level role requirement
	requiredRole := models.GetApprovalLevelRole(req.CurrentApprovalLevel)
	if requiredRole == "" {
		return nil, fmt.Errorf("invalid current approval level: %d", req.CurrentApprovalLevel)
	}

	hasRole := false
	for _, role := range userRoles {
		if strings.EqualFold(role, requiredRole) {
			hasRole = true
			break
		}
	}
	if !hasRole {
		return nil, fmt.Errorf("insufficient permissions: this request requires approval by role %s at level %d", requiredRole, req.CurrentApprovalLevel)
	}

	// 5. Rule 6: Same approval level cannot approve same request twice
	var levelAlreadyApproved bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM approval_history
			WHERE request_id = $1 AND approval_level = $2 AND action = 'APPROVED'
		)
	`, req.ID, req.CurrentApprovalLevel).Scan(&levelAlreadyApproved)
	if err != nil {
		return nil, err
	}
	if levelAlreadyApproved {
		return nil, fmt.Errorf("approval level %d has already been completed for this request", req.CurrentApprovalLevel)
	}

	currentLevel := req.CurrentApprovalLevel

	// 6. Execute approval or rejection logic
	if action == "REJECTED" {
		req.Status = models.StatusRejected

		updateQuery := `
			UPDATE requests
			SET status = $1, updated_at = CURRENT_TIMESTAMP
			WHERE id = $2
			RETURNING updated_at
		`
		if err := tx.QueryRow(updateQuery, req.Status, req.ID).Scan(&req.UpdatedAt); err != nil {
			return nil, err
		}
	} else {
		// Action is APPROVED: Check if this was the final level needed
		maxRequiredLevels := models.GetRequiredApprovalLevels(req.Amount)
		if currentLevel >= maxRequiredLevels {
			// Fully approved!
			req.Status = models.StatusApproved
		} else {
			// Advance to next level
			req.CurrentApprovalLevel = currentLevel + 1
		}

		updateQuery := `
			UPDATE requests
			SET status = $1, current_approval_level = $2, updated_at = CURRENT_TIMESTAMP
			WHERE id = $3
			RETURNING updated_at
		`
		if err := tx.QueryRow(updateQuery, req.Status, req.CurrentApprovalLevel, req.ID).Scan(&req.UpdatedAt); err != nil {
			return nil, err
		}
	}

	// 7. Insert audit history entry
	insertHistoryQuery := `
		INSERT INTO approval_history (request_id, approval_level, approved_by, action, reason, acted_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
	`
	if _, err := tx.Exec(insertHistoryQuery, req.ID, currentLevel, userID, action, reason); err != nil {
		return nil, fmt.Errorf("failed to record approval history: %w", err)
	}

	// 8. Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return req, nil
}
