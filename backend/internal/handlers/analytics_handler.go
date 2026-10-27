package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"financial-approval-system/internal/middleware"
	"financial-approval-system/internal/models"
	"financial-approval-system/internal/repository"
)

type AnalyticsHandler struct {
	analyticsRepo *repository.AnalyticsRepository
}

func NewAnalyticsHandler(analyticsRepo *repository.AnalyticsRepository) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsRepo: analyticsRepo}
}

func isUserPrivileged(roles []string) bool {
	for _, r := range roles {
		rUpper := strings.ToUpper(r)
		if rUpper == models.RoleManager || rUpper == models.RoleFinance || rUpper == models.RoleDirector {
			return true
		}
	}
	return false
}

// GetSummary returns high-level dashboard KPIs
func (h *AnalyticsHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r)
	if err != nil || userID == 0 {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	roles, _ := middleware.GetRolesFromContext(r)
	isPrivileged := isUserPrivileged(roles)

	summary, err := h.analyticsRepo.GetSummary(userID, isPrivileged)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch summary: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(summary)
}

// GetCategoryBreakdown returns spending grouped by request type
func (h *AnalyticsHandler) GetCategoryBreakdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r)
	if err != nil || userID == 0 {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	roles, _ := middleware.GetRolesFromContext(r)
	isPrivileged := isUserPrivileged(roles)

	breakdown, err := h.analyticsRepo.GetCategoryBreakdown(userID, isPrivileged)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch category breakdown"}`, http.StatusInternalServerError)
		return
	}

	if breakdown == nil {
		breakdown = []*models.CategoryBreakdown{}
	}

	json.NewEncoder(w).Encode(breakdown)
}

// GetTierBreakdown returns requests grouped by threshold tiers
func (h *AnalyticsHandler) GetTierBreakdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r)
	if err != nil || userID == 0 {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	roles, _ := middleware.GetRolesFromContext(r)
	isPrivileged := isUserPrivileged(roles)

	tiers, err := h.analyticsRepo.GetTierBreakdown(userID, isPrivileged)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch tier breakdown"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(tiers)
}

// GetRecentActivity returns the latest activity feed
func (h *AnalyticsHandler) GetRecentActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	limit := 8
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	activities, err := h.analyticsRepo.GetRecentActivity(limit)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch recent activity"}`, http.StatusInternalServerError)
		return
	}

	if activities == nil {
		activities = []*models.RecentActivityItem{}
	}

	json.NewEncoder(w).Encode(activities)
}

// GetReports returns filtered requests for comprehensive reporting
func (h *AnalyticsHandler) GetReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	filter := models.ReportFilter{
		StartDate: r.URL.Query().Get("start_date"),
		EndDate:   r.URL.Query().Get("end_date"),
		Type:      r.URL.Query().Get("type"),
		Status:    r.URL.Query().Get("status"),
	}

	if minStr := r.URL.Query().Get("min_amount"); minStr != "" {
		filter.MinAmount, _ = strconv.ParseFloat(minStr, 64)
	}
	if maxStr := r.URL.Query().Get("max_amount"); maxStr != "" {
		filter.MaxAmount, _ = strconv.ParseFloat(maxStr, 64)
	}

	requests, err := h.analyticsRepo.GetFilteredReport(filter)
	if err != nil {
		http.Error(w, `{"error": "Failed to generate report: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	if requests == nil {
		requests = []*models.Request{}
	}

	json.NewEncoder(w).Encode(requests)
}
