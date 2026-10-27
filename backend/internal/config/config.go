package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

// Config holds all configuration for the application
type Config struct {
	DBDriver   string // "sqlite" or "postgres"
	DBPath     string // path for sqlite db file (e.g., "./financial_approval.db")
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	ServerPort string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	driver := getEnv("DB_DRIVER", "sqlite")
	return &Config{
		DBDriver:   driver,
		DBPath:     getEnv("DB_PATH", "./financial_approval.db"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "financial_approval"),
		JWTSecret:  getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		ServerPort: getEnv("SERVER_PORT", "8085"),
	}
}

// ConnectDB establishes a connection to either SQLite or PostgreSQL based on configuration
func ConnectDB(cfg *Config) *sql.DB {
	var db *sql.DB
	var err error

	if cfg.DBDriver == "sqlite" {
		log.Printf("📂 Connecting to SQLite database at: %s", cfg.DBPath)
		db, err = sql.Open("sqlite", cfg.DBPath+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
		if err != nil {
			log.Fatalf("Failed to open SQLite database: %v", err)
		}
	} else {
		connStr := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
		)
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Fatalf("Failed to open PostgreSQL connection: %v", err)
		}
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Printf("✅ [%s] Database connected successfully", cfg.DBDriver)
	return db
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
