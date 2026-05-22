package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	_ "github.com/pgvector/pgvector-go"
	_ "github.com/tmc/langchaingo/llms/openai" // Verify compilation
)

func main() {
	dsn := "host=localhost user=admin password=password dbname=cyber_sim port=5432 sslmode=disable TimeZone=Asia/Bangkok"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	fmt.Println("Successfully connected to the database!")
	
	// Ensure DB connection is valid
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	fmt.Println("Database connection verified via Ping!")
}
