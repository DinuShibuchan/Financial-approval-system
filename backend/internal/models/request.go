package models

import "time"

// Request represents a financial request
type Request struct {
	ID                   int       `json:"id"`
	RequestNumber        string    `json:"request_number"`
	RequestType          string    `json:"request_type"`
	RequestedBy          int       `json:"requested_by"`
	RequestedByName      string    `json:"requested_by_name,omitempty"`
	Amount               float64   `json:"amount"`
	Description          string    `json:"description"`
	Status               string    `json:"status"`
	CurrentApprovalLevel int       `json:"current_approval_level"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// ApprovalHistory records each approval/rejection action
type ApprovalHistory struct {
	ID            int       `json:"id"`
	RequestID     int       `json:"request_id"`
	ApprovalLevel int       `json:"approval_level"`
	ApprovedBy    int       `json:"approved_by"`
	ApprovedByName string  `json:"approved_by_name,omitempty"`
	Action        string    `json:"action"` // APPROVED or REJECTED
	Reason        string    `json:"reason,omitempty"`
	ActedAt       time.Time `json:"acted_at"`
}

// Request type constants
const (
	RequestTypeExpense  = "EXPENSE"
	RequestTypePurchase = "PURCHASE"
	RequestTypeTravel   = "TRAVEL"
)

// Status constants
const (
	StatusDraft           = "DRAFT"
	StatusPendingApproval = "PENDING_APPROVAL"
	StatusApproved        = "APPROVED"
	StatusRejected        = "REJECTED"
	StatusCancelled       = "CANCELLED"
)

// Approval level constants
const (
	ApprovalLevelManager  = 1
	ApprovalLevelFinance  = 2
	ApprovalLevelDirector = 3
)

// CreateRequestPayload is the payload for creating a new request
type CreateRequestPayload struct {
	RequestType string  `json:"request_type"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

// ApproveRejectPayload is the payload for approving or rejecting a request
type ApproveRejectPayload struct {
	Reason string `json:"reason,omitempty"`
}

// GetRequiredApprovalLevels returns the number of approval levels needed based on amount
func GetRequiredApprovalLevels(amount float64) int {
	if amount < 10000 {
		return 1 // Manager only
	}
	if amount <= 50000 {
		return 2 // Manager + Finance
	}
	return 3 // Manager + Finance + Director
}

// GetApprovalLevelRole returns the role name for a given approval level
func GetApprovalLevelRole(level int) string {
	switch level {
	case ApprovalLevelManager:
		return RoleManager
	case ApprovalLevelFinance:
		return RoleFinance
	case ApprovalLevelDirector:
		return RoleDirector
	default:
		return ""
	}
}
