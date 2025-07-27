package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	var err error
	host := os.Getenv("DATABASE_HOST")
	name := os.Getenv("DATABASE_NAME")
	port := os.Getenv("DATABASE_PORT")
	
	if host == "" || name == "" || port == "" {
		panic("DATABASE_HOST, DATABASE_NAME, and DATABASE_PORT environment variables must be set")
	}

	connStr := fmt.Sprintf("host=%s port=%s dbname=%s sslmode=require", host, port, name)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(fmt.Sprintf("failed to open database connection: %v", err))
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
}

func handler(ctx context.Context) (string, error) {
	var result int
	err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return "", fmt.Errorf("failed to execute SELECT 1: %v", err)
	}

	return fmt.Sprintf("access-vpc: Successfully executed SELECT 1, result: %d", result), nil
}

func main() {
	lambda.Start(handler)
}
