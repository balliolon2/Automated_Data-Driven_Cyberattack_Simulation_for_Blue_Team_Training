package models

import (
	"log"

	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	// 1. Safely ensure 'specialist' value exists in PostgreSQL enum 'role'
	_ = db.Exec(`ALTER TYPE role ADD VALUE IF NOT EXISTS 'specialist'`).Error

	// 2. Add nickname column if missing
	var columnExists bool
	row := db.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'users' AND column_name = 'nickname'
		)
	`).Row()
	if row != nil {
		_ = row.Scan(&columnExists)
	}

	if !columnExists {
		log.Println("Migrating users table: adding nickname column...")
		if err := db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS nickname VARCHAR`).Error; err != nil {
			log.Printf("Warning adding nickname column: %v", err)
		}
	}

	// 3. Backfill any NULL or empty nicknames using email prefix
	_ = db.Exec(`
		UPDATE users 
		SET nickname = split_part(email, '@', 1) 
		WHERE nickname IS NULL OR nickname = ''
	`).Error

	// Ensure unique index on nickname
	_ = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_nickname ON users(nickname)`).Error

	// 4. AutoMigrate SpecialistApplication table
	if err := db.AutoMigrate(&SpecialistApplication{}); err != nil {
		log.Printf("Error automigrating SpecialistApplication: %v", err)
		return err
	}

	// 5. AutoMigrate AnalysisThread table
	if err := db.AutoMigrate(&AnalysisThread{}); err != nil {
		log.Printf("Error automigrating AnalysisThread: %v", err)
		return err
	}

	log.Println("Database schema migration completed successfully.")
	return nil
}
