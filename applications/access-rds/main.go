package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	_ "github.com/lib/pq"
)

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
