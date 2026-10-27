package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"financial-approval-system/internal/models"
)

type AnalyticsRepository struct {
	db *sql.DB
}

func NewAnalyticsRepository(db *sql.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// GetSummary calculates high-level KPI metrics
func (r *AnalyticsRepository) GetSummary(userID int, isPrivileged bool) (*models.DashboardSummary, error) {
	summary := &models.DashboardSummary{}

	whereClause := ""
	args := []interface{}{}
	if !isPrivileged {
		whereClause = "WHERE requested_by = $1"
		args = append(args, userID)
	}

	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) AS total_count,
			COUNT(CASE WHEN status = 'PENDING_APPROVAL' THEN 1 END) AS pending_count,
			COUNT(CASE WHEN status = 'APPROVED' THEN 1 END) AS approved_count,
			COUNT(CASE WHEN status = 'REJECTED' THEN 1 END) AS rejected_count,
			COUNT(CASE WHEN status = 'DRAFT' THEN 1 END) AS draft_count,
			COALESCE(SUM(amount), 0) AS total_amount,
			COALESCE(SUM(CASE WHEN status = 'APPROVED' THEN amount ELSE 0 END), 0) AS approved_amount,
			COALESCE(SUM(CASE WHEN status = 'PENDING_APPROVAL' THEN amount ELSE 0 END), 0) AS pending_amount
		FROM requests
		%s
	`, whereClause)

	err := r.db.QueryRow(query, args...).Scan(
		&summary.TotalRequests,
		&summary.PendingRequests,
		&summary.ApprovedRequests,
		&summary.RejectedRequests,
		&summary.DraftRequests,
		&summary.TotalAmount,
		&summary.ApprovedAmount,
		&summary.PendingAmount,
	)
	if err != nil {
		return nil, err
	}

	return summary, nil
}

// GetCategoryBreakdown returns spending metrics by request type
func (r *AnalyticsRepository) GetCategoryBreakdown(userID int, isPrivileged bool) ([]*models.CategoryBreakdown, error) {
	whereClause := ""
	args := []interface{}{}
	if !isPrivileged {
		whereClause = "WHERE requested_by = $1"
		args = append(args, userID)
	}

	query := fmt.Sprintf(`
		SELECT request_type, COUNT(*), COALESCE(SUM(amount), 0)
		FROM requests
		%s
		GROUP BY request_type
	`, whereClause)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.CategoryBreakdown
	var overallTotal float64

	for rows.Next() {
		item := &models.CategoryBreakdown{}
		if err := rows.Scan(&item.Type, &item.Count, &item.TotalAmount); err != nil {
			return nil, err
		}
		overallTotal += item.TotalAmount
		list = append(list, item)
	}

	for _, item := range list {
		if overallTotal > 0 {
			item.Percentage = (item.TotalAmount / overallTotal) * 100.0
		}
	}

	return list, nil
}

// GetTierBreakdown returns amount distribution by approval tiers
func (r *AnalyticsRepository) GetTierBreakdown(userID int, isPrivileged bool) ([]*models.TierBreakdown, error) {
	whereClause := ""
	args := []interface{}{}
	if !isPrivileged {
		whereClause = "WHERE requested_by = $1"
		args = append(args, userID)
	}

	query := fmt.Sprintf(`
		SELECT 
			COUNT(CASE WHEN amount < 10000 THEN 1 END) AS t1_count,
			COALESCE(SUM(CASE WHEN amount < 10000 THEN amount ELSE 0 END), 0) AS t1_amount,
			COUNT(CASE WHEN amount >= 10000 AND amount <= 50000 THEN 1 END) AS t2_count,
			COALESCE(SUM(CASE WHEN amount >= 10000 AND amount <= 50000 THEN amount ELSE 0 END), 0) AS t2_amount,
			COUNT(CASE WHEN amount > 50000 THEN 1 END) AS t3_count,
			COALESCE(SUM(CASE WHEN amount > 50000 THEN amount ELSE 0 END), 0) AS t3_amount
		FROM requests
		%s
	`, whereClause)

	var t1Count, t2Count, t3Count int
	var t1Amount, t2Amount, t3Amount float64

	err := r.db.QueryRow(query, args...).Scan(
		&t1Count, &t1Amount,
		&t2Count, &t2Amount,
		&t3Count, &t3Amount,
	)
	if err != nil {
		return nil, err
	}

	tiers := []*models.TierBreakdown{
		{
			Tier:        1,
			TierName:    "Tier 1",
			RangeLabel:  "< ₹10,000",
			Count:       t1Count,
			TotalAmount: t1Amount,
			Approver:    "Manager",
		},
		{
			Tier:        2,
			TierName:    "Tier 2",
			RangeLabel:  "₹10,000 – ₹50,000",
			Count:       t2Count,
			TotalAmount: t2Amount,
			Approver:    "Manager → Finance",
		},
		{
			Tier:        3,
			TierName:    "Tier 3",
			RangeLabel:  "> ₹50,000",
			Count:       t3Count,
			TotalAmount: t3Amount,
			Approver:    "Manager → Finance → Director",
		},
	}

	return tiers, nil
}

// GetRecentActivity returns the latest requests and approvals
func (r *AnalyticsRepository) GetRecentActivity(limit int) ([]*models.RecentActivityItem, error) {
	if limit <= 0 {
		limit = 8
	}

	query := `
		SELECT r.id, r.request_number, r.request_type, COALESCE(u.name, 'User'),
		       r.amount, r.status, r.updated_at
		FROM requests r
		LEFT JOIN users u ON r.requested_by = u.id
		ORDER BY r.updated_at DESC
		LIMIT $1
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*models.RecentActivityItem
	for rows.Next() {
		item := &models.RecentActivityItem{}
		if err := rows.Scan(
			&item.ID,
			&item.RequestNumber,
			&item.RequestType,
			&item.RequestedBy,
			&item.Amount,
			&item.Status,
			&item.Timestamp,
		); err != nil {
			return nil, err
		}
		activities = append(activities, item)
	}

	return activities, nil
}

// GetFilteredReport retrieves requests matching complex criteria
func (r *AnalyticsRepository) GetFilteredReport(filter models.ReportFilter) ([]*models.Request, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if strings.TrimSpace(filter.Status) != "" && filter.Status != "ALL" {
		conditions = append(conditions, fmt.Sprintf("r.status = $%d", argIndex))
		args = append(args, strings.ToUpper(filter.Status))
		argIndex++
	}

	if strings.TrimSpace(filter.Type) != "" && filter.Type != "ALL" {
		conditions = append(conditions, fmt.Sprintf("r.request_type = $%d", argIndex))
		args = append(args, strings.ToUpper(filter.Type))
		argIndex++
	}

	if filter.MinAmount > 0 {
		conditions = append(conditions, fmt.Sprintf("r.amount >= $%d", argIndex))
		args = append(args, filter.MinAmount)
		argIndex++
	}

	if filter.MaxAmount > 0 {
		conditions = append(conditions, fmt.Sprintf("r.amount <= $%d", argIndex))
		args = append(args, filter.MaxAmount)
		argIndex++
	}

	if strings.TrimSpace(filter.StartDate) != "" {
		conditions = append(conditions, fmt.Sprintf("r.created_at >= $%d", argIndex))
		args = append(args, filter.StartDate)
		argIndex++
	}

	if strings.TrimSpace(filter.EndDate) != "" {
		conditions = append(conditions, fmt.Sprintf("r.created_at <= $%d", argIndex))
		args = append(args, filter.EndDate)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT r.id, r.request_number, r.request_type, r.requested_by, u.name,
		       r.amount, r.description, r.status, r.current_approval_level,
		       r.created_at, r.updated_at
		FROM requests r
		LEFT JOIN users u ON r.requested_by = u.id
		%s
		ORDER BY r.created_at DESC
	`, whereClause)

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
