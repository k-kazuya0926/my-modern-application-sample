package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

var (
	db           *sql.DB
	databaseHost string
	databaseName string
	databasePort string
)

func init() {
	ctx := context.Background()

	databaseHost = os.Getenv("DATABASE_HOST")
	databaseName = os.Getenv("DATABASE_NAME")
	databasePort = os.Getenv("DATABASE_PORT")

	if databaseHost == "" || databaseName == "" || databasePort == "" {
		panic("DATABASE_HOST, DATABASE_NAME, and DATABASE_PORT environment variables must be set.")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to load AWS config: %v", err))
	}

	// Generate IAM authentication token
	authToken, err := auth.BuildAuthToken(
		ctx, databaseHost+":"+databasePort, cfg.Region, "postgres", cfg.Credentials)
	if err != nil {
		panic(fmt.Sprintf("failed to build auth token: %v", err))
	}

	// Build connection string with IAM auth token
	dsn := fmt.Sprintf("host=%s port=%s user=postgres password=%s dbname=%s sslmode=require",
		databaseHost, databasePort, authToken, databaseName)

	// Connect to database
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		panic(fmt.Sprintf("failed to open database connection: %v", err))
	}

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		panic(fmt.Sprintf("failed to ping database: %v", err))
	}

	// Run migrations
	if err := runMigrations(ctx); err != nil {
		panic(fmt.Sprintf("failed to run migrations: %v", err))
	}
}

func runMigrations(ctx context.Context) error {
	// Create migration source from embedded files
	sourceDriver, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %v", err)
	}

	// Create database driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %v", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %v", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	fmt.Println("Migrations completed successfully")
	return nil
}

func handler(ctx context.Context) (string, error) {
	// Execute SELECT 1; using direct database connection with IAM authentication
	var result int
	err := db.QueryRowContext(ctx, "SELECT 1;").Scan(&result)
	if err != nil {
		return "", fmt.Errorf("failed to execute SELECT 1; with IAM auth: %v", err)
	}

	return fmt.Sprintf("access-rds: Successfully executed SELECT 1; with IAM auth, result: %d", result), nil
}

func main() {
	lambda.Start(handler)
}
