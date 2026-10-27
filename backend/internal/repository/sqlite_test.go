package repository_test

import (
	"database/sql"
	"os"
	"testing"

	"financial-approval-system/internal/models"
	"financial-approval-system/internal/repository"
	_ "modernc.org/sqlite"
)

func TestSQLiteIntegration(t *testing.T) {
	// Create an in-memory SQLite database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer db.Close()

	// Load and run sqlite_init.sql
	migrationData, err := os.ReadFile("../../migrations/sqlite_init.sql")
	if err != nil {
		t.Fatalf("failed to read migration: %v", err)
	}

	_, err = db.Exec(string(migrationData))
	if err != nil {
		t.Fatalf("failed to run migration: %v", err)
	}

	// 1. Test User Repository
	userRepo := repository.NewUserRepository(db)
	user, err := userRepo.GetUserByEmail("employee@company.com")
	if err != nil || user == nil {
		t.Fatalf("failed to fetch user: %v", err)
	}
	roles, err := userRepo.GetUserRoles(user.ID)
	if err != nil || len(roles) == 0 {
		t.Fatalf("failed to fetch roles: %v", err)
	}

	// 2. Test Request Repository
	reqRepo := repository.NewRequestRepository(db)
	requests, err := reqRepo.GetByUserID(user.ID)
	if err != nil {
		t.Fatalf("failed to get requests: %v", err)
	}
	if len(requests) == 0 {
		t.Fatalf("expected seeded requests, got 0")
	}

	// 3. Test Approval Repository
	apprRepo := repository.NewApprovalRepository(db)
	pending, err := apprRepo.GetPendingApprovalsForUser(2, []string{"MANAGER"}) // Bob Manager
	if err != nil {
		t.Fatalf("failed to get pending approvals: %v", err)
	}
	t.Logf("Pending approvals for Manager: %d", len(pending))

	// 4. Test Analytics Repository
	analyticsRepo := repository.NewAnalyticsRepository(db)
	summary, err := analyticsRepo.GetSummary(2, true)
	if err != nil {
		t.Fatalf("failed to get analytics summary: %v", err)
	}
	t.Logf("Total requests in summary: %d, Approved amount: %f", summary.TotalRequests, summary.ApprovedAmount)

	tiers, err := analyticsRepo.GetTierBreakdown(2, true)
	if err != nil {
		t.Fatalf("failed to get tier breakdown: %v", err)
	}
	if len(tiers) != 3 {
		t.Fatalf("expected 3 tiers, got %d", len(tiers))
	}

	// 5. Test Approval Processing (Manager approving Request 1)
	updatedReq, err := apprRepo.ProcessApproval(1, 2, []string{"MANAGER"}, "APPROVED", "Approved for team celebration")
	if err != nil {
		t.Fatalf("failed to process approval: %v", err)
	}
	if updatedReq.Status != models.StatusApproved {
		t.Fatalf("expected status APPROVED for tier 1 under 10k, got %s", updatedReq.Status)
	}
}
