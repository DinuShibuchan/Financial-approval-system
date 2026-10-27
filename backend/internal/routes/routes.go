package routes

import (
	"database/sql"
	"net/http"
	"strings"

	"financial-approval-system/internal/config"
	"financial-approval-system/internal/handlers"
	"financial-approval-system/internal/middleware"
	"financial-approval-system/internal/repository"
)

// SetupRoutes configures all API routes
func SetupRoutes(db *sql.DB, cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	requestRepo := repository.NewRequestRepository(db)
	approvalRepo := repository.NewApprovalRepository(db)
	analyticsRepo := repository.NewAnalyticsRepository(db)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, cfg)
	reqHandler := handlers.NewRequestHandler(requestRepo, userRepo)
	approvalHandler := handlers.NewApprovalHandler(approvalRepo)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsRepo)

	// ==========================================
	// Public routes (no auth required)
	// ==========================================
	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)

	// Health check
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok", "message": "Financial Approval System API is running"}`))
	})

	// ==========================================
	// Protected routes (auth required)
	// ==========================================
	authMw := middleware.AuthMiddleware(cfg)

	// Current user profile
	mux.Handle("/api/auth/me", authMw(http.HandlerFunc(authHandler.GetMe)))

	// User management
	mux.Handle("/api/users", authMw(http.HandlerFunc(authHandler.GetUsers)))
	mux.Handle("/api/users/", authMw(http.HandlerFunc(authHandler.AssignRoles)))

	// Roles
	mux.Handle("/api/roles", authMw(http.HandlerFunc(authHandler.GetRoles)))

	// ==========================================
	// Dashboard & Analytics routes
	// ==========================================
	mux.Handle("/api/dashboard/summary", authMw(http.HandlerFunc(analyticsHandler.GetSummary)))
	mux.Handle("/api/dashboard/by-type", authMw(http.HandlerFunc(analyticsHandler.GetCategoryBreakdown)))
	mux.Handle("/api/dashboard/by-amount", authMw(http.HandlerFunc(analyticsHandler.GetTierBreakdown)))
	mux.Handle("/api/dashboard/recent", authMw(http.HandlerFunc(analyticsHandler.GetRecentActivity)))

	// ==========================================
	// Reports routes
	// ==========================================
	mux.Handle("/api/reports", authMw(http.HandlerFunc(analyticsHandler.GetReports)))

	// ==========================================
	// Approvals Queue
	// ==========================================
	mux.Handle("/api/approvals/pending", authMw(http.HandlerFunc(approvalHandler.GetPendingApprovals)))

	// ==========================================
	// Financial Request routes
	// ==========================================
	// Dispatcher for /api/requests
	mux.Handle("/api/requests", authMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			reqHandler.CreateRequest(w, r)
			return
		}
		if r.Method == http.MethodGet {
			reqHandler.GetAllRequests(w, r)
			return
		}
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	})))

	mux.Handle("/api/requests/my", authMw(http.HandlerFunc(reqHandler.GetMyRequests)))

	// Subpath dispatcher for /api/requests/{id} and actions:
	mux.Handle("/api/requests/", authMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > len("/api/requests/") {
			if strings.HasSuffix(path, "/submit") && r.Method == http.MethodPost {
				reqHandler.SubmitRequest(w, r)
				return
			}
			if strings.HasSuffix(path, "/cancel") && r.Method == http.MethodPost {
				reqHandler.CancelRequest(w, r)
				return
			}
			if strings.HasSuffix(path, "/approve") && r.Method == http.MethodPost {
				approvalHandler.ApproveRequest(w, r)
				return
			}
			if strings.HasSuffix(path, "/reject") && r.Method == http.MethodPost {
				approvalHandler.RejectRequest(w, r)
				return
			}
			if strings.HasSuffix(path, "/history") && r.Method == http.MethodGet {
				approvalHandler.GetApprovalHistory(w, r)
				return
			}
			if r.Method == http.MethodGet {
				reqHandler.GetRequestByID(w, r)
				return
			}
			if r.Method == http.MethodPut {
				reqHandler.UpdateRequest(w, r)
				return
			}
			if r.Method == http.MethodDelete {
				reqHandler.DeleteRequest(w, r)
				return
			}
		}
		http.Error(w, `{"error": "Method not allowed or invalid path"}`, http.StatusMethodNotAllowed)
	})))

	// ==========================================
	// Apply global middleware
	// ==========================================
	var handler http.Handler = mux
	handler = middleware.JSONMiddleware(handler)
	handler = middleware.CORSMiddleware(handler)

	return handler
}
