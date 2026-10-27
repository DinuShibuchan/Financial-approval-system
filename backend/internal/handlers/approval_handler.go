package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"financial-approval-system/internal/middleware"
	"financial-approval-system/internal/models"
	"financial-approval-system/internal/repository"
)

type ApprovalHandler struct {
	approvalRepo *repository.ApprovalRepository
}

func NewApprovalHandler(approvalRepo *repository.ApprovalRepository) *ApprovalHandler {
	return &ApprovalHandler{approvalRepo: approvalRepo}
}

// ApproveRequest approves the request at its current level
func (h *ApprovalHandler) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r)
	if err != nil || userID == 0 {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	userRoles, err := middleware.GetRolesFromContext(r)
	if err != nil {
		http.Error(w, `{"error": "Forbidden: roles missing"}`, http.StatusForbidden)
		return
	}

	reqID, err := middleware.GetIDFromURLPath(r, "/api/requests")
	if err != nil {
		http.Error(w, `{"error": "Invalid request ID"}`, http.StatusBadRequest)
		return
	}

	var payload models.ApproveRejectPayload
	// Body is optional for approval
	_ = json.NewDecoder(r.Body).Decode(&payload)

	updatedReq, err := h.approvalRepo.ProcessApproval(reqID, userID, userRoles, "APPROVED", payload.Reason)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(updatedReq)
}

// RejectRequest rejects the request with a mandatory reason
func (h *ApprovalHandler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r)
	if err != nil || userID == 0 {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	userRoles, err := middleware.GetRolesFromContext(r)
	if err != nil {
		http.Error(w, `{"error": "Forbidden: roles missing"}`, http.StatusForbidden)
		return
	}

	reqID, err := middleware.GetIDFromURLPath(r, "/api/requests")
	if err != nil {
		http.Error(w, `{"error": "Invalid request ID"}`, http.StatusBadRequest)
		return
	}

	var payload models.ApproveRejectPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(payload.Reason) == "" {
		http.Error(w, `{"error": "Rejection reason is mandatory"}`, http.StatusBadRequest)
		return
	}

	updatedReq, err := h.approvalRepo.ProcessApproval(reqID, userID, userRoles, "REJECTED", payload.Reason)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(updatedReq)
}

// GetApprovalHistory returns audit logs of approvals and rejections for a request
func (h *ApprovalHandler) GetApprovalHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	reqID, err := middleware.GetIDFromURLPath(r, "/api/requests")
	if err != nil {
		http.Error(w, `{"error": "Invalid request ID"}`, http.StatusBadRequest)
		return
	}

	history, err := h.approvalRepo.GetApprovalHistory(reqID)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch approval history"}`, http.StatusInternalServerError)
		return
	}

	if history == nil {
		history = []*models.ApprovalHistory{}
	}

	json.NewEncoder(w).Encode(history)
}

// GetPendingApprovals returns requests pending approval specifically for the current user's roles
func (h *ApprovalHandler) GetPendingApprovals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r)
	if err != nil || userID == 0 {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	userRoles, err := middleware.GetRolesFromContext(r)
	if err != nil {
		http.Error(w, `{"error": "Forbidden: roles missing"}`, http.StatusForbidden)
		return
	}

	requests, err := h.approvalRepo.GetPendingApprovalsForUser(userID, userRoles)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch pending approvals: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	if requests == nil {
		requests = []*models.Request{}
	}

	json.NewEncoder(w).Encode(requests)
}
