package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"financial-approval-system/internal/config"
	"financial-approval-system/internal/routes"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	db := config.ConnectDB(cfg)
	defer db.Close()

	// Run migrations
	runMigrations(db, cfg.DBDriver)

	// Setup routes
	handler := routes.SetupRoutes(db, cfg)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("🚀 Server starting on http://localhost%s", addr)
	log.Printf("📋 API Health Check: http://localhost%s/api/health", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func runMigrations(db *sql.DB, driver string) {
	migrationsDir := findMigrationsDir()
	if migrationsDir == "" {
		log.Println("⚠️  Migrations directory not found, skipping migrations")
		return
	}

	var targetFiles []string
	if driver == "sqlite" {
		targetFiles = []string{"sqlite_init.sql"}
	} else {
		targetFiles = []string{"001_init.sql", "002_seed_users.sql"}
	}

	for _, file := range targetFiles {
		filePath := filepath.Join(migrationsDir, file)
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("⚠️  Could not read migration file %s: %v", file, err)
			continue
		}

		// Execute migration
		_, err = db.Exec(string(data))
		if err != nil {
			// Some errors like already existing tables are acceptable notices
			if strings.Contains(strings.ToLower(err.Error()), "already exists") {
				log.Printf("ℹ️  Migration %s: schema already exists", file)
			} else {
				log.Printf("⚠️  Migration %s notice: %v", file, err)
			}
		} else {
			log.Printf("✅ Applied migration: %s", file)
		}
	}
}

func findMigrationsDir() string {
	candidates := []string{
		"migrations",
		"../migrations",
		"../../migrations",
		"backend/migrations",
	}

	for _, dir := range candidates {
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}
	return ""
}
