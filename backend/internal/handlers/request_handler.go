package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"financial-approval-system/internal/middleware"
	"financial-approval-system/internal/models"
	"financial-approval-system/internal/repository"
)

type RequestHandler struct {
	requestRepo *repository.RequestRepository
	userRepo    *repository.UserRepository
}

func NewRequestHandler(requestRepo *repository.RequestRepository, userRepo *repository.UserRepository) *RequestHandler {
	return &RequestHandler{
		requestRepo: requestRepo,
		userRepo:    userRepo,
	}
}

type CreateRequestInput struct {
	RequestType string  `json:"request_type"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Submit      bool    `json:"submit"` // If true, sets status to PENDING_APPROVAL right away
}

// CreateRequest handles creating a draft or directly submitting a financial request
func (h *RequestHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r)
	if err != nil || userID == 0 {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var input CreateRequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validation
	input.RequestType = strings.ToUpper(strings.TrimSpace(input.RequestType))
	if input.RequestType != models.RequestTypeExpense &&
		input.RequestType != models.RequestTypePurchase &&
		input.RequestType != models.RequestTypeTravel {
		http.Error(w, `{"error": "Invalid request_type. Must be EXPENSE, PURCHASE, or TRAVEL"}`, http.StatusBadRequest)
		return
	}

	if input.Amount <= 0 {
		http.Error(w, `{"error": "Amount must be greater than 0"}`, http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(input.Description) == "" {
		http.Error(w, `{"error": "Description is required"}`, http.StatusBadRequest)
		return
	}

	reqStatus := models.StatusDraft
	if input.Submit {
		reqStatus = models.StatusPendingApproval
	}

	newReq := &models.Request{
		RequestType: input.RequestType,
		RequestedBy: userID,
		Amount:      input.Amount,
		Description: input.Description,
		Status:      reqStatus,
	}

	if err := h.requestRepo.Create(newReq); err != nil {
		http.Error(w, `{"error": "Failed to create request: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newReq)
}

// GetMyRequests returns requests created by the authenticated user
func (h *RequestHandler) GetMyRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r)
	if err != nil || userID == 0 {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	requests, err := h.requestRepo.GetByUserID(userID)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch requests"}`, http.StatusInternalServerError)
		return
	}

	if requests == nil {
		requests = []*models.Request{}
	}

	json.NewEncoder(w).Encode(requests)
}

// GetAllRequests returns all requests with optional status and type filtering
func (h *RequestHandler) GetAllRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	status := strings.TrimSpace(r.URL.Query().Get("status"))
	reqType := strings.TrimSpace(r.URL.Query().Get("type"))

	requests, err := h.requestRepo.GetAll(status, reqType)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch requests"}`, http.StatusInternalServerError)
		return
	}

	if requests == nil {
		requests = []*models.Request{}
	}

	json.NewEncoder(w).Encode(requests)
}

// GetRequestByID returns a single request by ID
func (h *RequestHandler) GetRequestByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	id, err := middleware.GetIDFromURLPath(r, "/api/requests")
	if err != nil {
		http.Error(w, `{"error": "Invalid request ID"}`, http.StatusBadRequest)
		return
	}

	req, err := h.requestRepo.GetByID(id)
	if err != nil {
		http.Error(w, `{"error": "Request not found"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(req)
}

// UpdateRequest allows editing of a DRAFT request
func (h *RequestHandler) UpdateRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r)
	if err != nil || userID == 0 {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id, err := middleware.GetIDFromURLPath(r, "/api/requests")
	if err != nil {
		http.Error(w, `{"error": "Invalid request ID"}`, http.StatusBadRequest)
		return
	}

	var input models.CreateRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	input.RequestType = strings.ToUpper(strings.TrimSpace(input.RequestType))
	if input.Amount <= 0 {
		http.Error(w, `{"error": "Amount must be greater than 0"}`, http.StatusBadRequest)
		return
	}

	updated, err := h.requestRepo.Update(id, userID, input.RequestType, input.Amount, input.Description)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

// SubmitRequest changes a DRAFT request to PENDING_APPROVAL
func (h *RequestHandler) SubmitRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r)
	if err != nil || userID == 0 {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id, err := middleware.GetIDFromURLPath(r, "/api/requests")
	if err != nil {
		http.Error(w, `{"error": "Invalid request ID"}`, http.StatusBadRequest)
		return
	}

	submitted, err := h.requestRepo.Submit(id, userID)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(submitted)
}

// CancelRequest changes a DRAFT or PENDING_APPROVAL request to CANCELLED
func (h *RequestHandler) CancelRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r)
	if err != nil || userID == 0 {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id, err := middleware.GetIDFromURLPath(r, "/api/requests")
	if err != nil {
		http.Error(w, `{"error": "Invalid request ID"}`, http.StatusBadRequest)
		return
	}

	cancelled, err := h.requestRepo.Cancel(id, userID)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(cancelled)
}

// DeleteRequest removes a DRAFT request
func (h *RequestHandler) DeleteRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r)
	if err != nil || userID == 0 {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id, err := middleware.GetIDFromURLPath(r, "/api/requests")
	if err != nil {
		http.Error(w, `{"error": "Invalid request ID"}`, http.StatusBadRequest)
		return
	}

	if err := h.requestRepo.Delete(id, userID); err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Request deleted successfully"}`))
}
