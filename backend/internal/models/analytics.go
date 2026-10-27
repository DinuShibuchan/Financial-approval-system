package models

import "time"

// DashboardSummary represents the high-level KPI metrics
type DashboardSummary struct {
	TotalRequests    int     `json:"total_requests"`
	PendingRequests  int     `json:"pending_requests"`
	ApprovedRequests int     `json:"approved_requests"`
	RejectedRequests int     `json:"rejected_requests"`
	DraftRequests    int     `json:"draft_requests"`
	TotalAmount      float64 `json:"total_amount"`
	ApprovedAmount   float64 `json:"approved_amount"`
	PendingAmount    float64 `json:"pending_amount"`
}

// CategoryBreakdown represents metrics grouped by request type
type CategoryBreakdown struct {
	Type        string  `json:"type"`
	Count       int     `json:"count"`
	TotalAmount float64 `json:"total_amount"`
	Percentage  float64 `json:"percentage"`
}

// StatusBreakdown represents metrics grouped by status
type StatusBreakdown struct {
	Status      string  `json:"status"`
	Count       int     `json:"count"`
	TotalAmount float64 `json:"total_amount"`
}

// TierBreakdown represents metrics grouped by threshold tier
type TierBreakdown struct {
	Tier        int     `json:"tier"`
	TierName    string  `json:"tier_name"`
	RangeLabel  string  `json:"range_label"`
	Count       int     `json:"count"`
	TotalAmount float64 `json:"total_amount"`
	Approver    string  `json:"approver"`
}

// RecentActivityItem represents an event in the recent activity feed
type RecentActivityItem struct {
	ID            int       `json:"id"`
	RequestNumber string    `json:"request_number"`
	RequestType   string    `json:"request_type"`
	RequestedBy   string    `json:"requested_by"`
	Amount        float64   `json:"amount"`
	Status        string    `json:"status"`
	Timestamp     time.Time `json:"timestamp"`
	Action        string    `json:"action,omitempty"`
}

// ReportFilter represents query parameters for reports
type ReportFilter struct {
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
	Type      string  `json:"type"`
	Status    string  `json:"status"`
	MinAmount float64 `json:"min_amount"`
	MaxAmount float64 `json:"max_amount"`
}
