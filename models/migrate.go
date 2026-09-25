package models

import (
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
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

	// 4. Ensure community and specialist tables exist (idempotent DDL from init.sql)
	ddl := []string{
		`CREATE TABLE IF NOT EXISTS specialist_applications (
			application_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
			status VARCHAR NOT NULL DEFAULT 'pending',
			bio TEXT NOT NULL,
			resume_path VARCHAR NOT NULL,
			certificate_path VARCHAR,
			linkedin_url VARCHAR,
			portfolio_url VARCHAR,
			rejection_reason TEXT,
			reviewed_by UUID REFERENCES users(user_id),
			reviewed_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT now(),
			updated_at TIMESTAMP DEFAULT now()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_specialist_applications_user ON specialist_applications(user_id)`,
		`CREATE TABLE IF NOT EXISTS analysis_threads (
			thread_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			author_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
			scenario_id UUID REFERENCES scenarios(scenario_id) ON DELETE SET NULL,
			title VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			tags JSONB,
			upvote_count INT DEFAULT 0,
			view_count INT DEFAULT 0,
			is_pinned BOOLEAN DEFAULT false,
			is_locked BOOLEAN DEFAULT false,
			created_at TIMESTAMP DEFAULT now(),
			updated_at TIMESTAMP DEFAULT now()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_analysis_threads_scenario ON analysis_threads(scenario_id)`,
		`CREATE INDEX IF NOT EXISTS idx_analysis_threads_author ON analysis_threads(author_id)`,
		`CREATE TABLE IF NOT EXISTS thread_comments (
			comment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			thread_id UUID REFERENCES analysis_threads(thread_id) ON DELETE CASCADE,
			user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
			parent_comment_id UUID REFERENCES thread_comments(comment_id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			upvote_count INT DEFAULT 0,
			is_pinned BOOLEAN DEFAULT false,
			is_hidden BOOLEAN DEFAULT false,
			created_at TIMESTAMP DEFAULT now(),
			updated_at TIMESTAMP DEFAULT now()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_thread_comments_thread ON thread_comments(thread_id)`,
		`CREATE INDEX IF NOT EXISTS idx_thread_comments_parent ON thread_comments(parent_comment_id)`,
		`CREATE TABLE IF NOT EXISTS thread_upvotes (
			thread_id UUID REFERENCES analysis_threads(thread_id) ON DELETE CASCADE,
			user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
			created_at TIMESTAMP DEFAULT now(),
			PRIMARY KEY (thread_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS comment_upvotes (
			comment_id UUID REFERENCES thread_comments(comment_id) ON DELETE CASCADE,
			user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
			created_at TIMESTAMP DEFAULT now(),
			PRIMARY KEY (comment_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_notifications (
			notification_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
			title VARCHAR(120) NOT NULL,
			message TEXT NOT NULL,
			link_url VARCHAR(255),
			is_read BOOLEAN DEFAULT false,
			created_at TIMESTAMP DEFAULT now()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_user_notifications_user ON user_notifications(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_notifications_unread ON user_notifications(user_id, is_read)`,
	}

	for _, stmt := range ddl {
		if err := db.Exec(stmt).Error; err != nil {
			log.Printf("Migration DDL error: %v", err)
			return err
		}
	}

	// 5. AutoMigrate remaining model configurations
	if err := db.AutoMigrate(
		&SpecialistApplication{},
		&AnalysisThread{},
		&ThreadComment{},
		&ThreadUpvote{},
		&CommentUpvote{},
		&UserNotification{},
	); err != nil {
		log.Printf("Error during AutoMigrate: %v", err)
		return err
	}

	// 6. Provision or synchronize admin user from environment variables (.env)
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	adminNickname := os.Getenv("ADMIN_NICKNAME")
	if adminNickname == "" {
		adminNickname = "admin"
	}

	if adminEmail != "" && adminPassword != "" {
		var existingAdmin User
		result := db.Where("email = ? AND role = ?", adminEmail, "admin").First(&existingAdmin)
		if result.Error != nil {
			// Admin does not exist, provision new admin user
			hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
			if err == nil {
				admin := User{
					Email:        adminEmail,
					Nickname:     adminNickname,
					PasswordHash: string(hash),
					Role:         "admin",
					CurrentTier:  1,
					IsActive:     true,
				}
				if err := db.Create(&admin).Error; err != nil {
					log.Printf("Warning creating admin account: %v", err)
				} else {
					log.Printf("Admin account provisioned from environment: %s", adminEmail)
				}
			} else {
				log.Printf("Error hashing admin password: %v", err)
			}
		} else {
			// Admin account exists: verify or update password if changed in .env
			if err := bcrypt.CompareHashAndPassword([]byte(existingAdmin.PasswordHash), []byte(adminPassword)); err != nil {
				hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
				if err == nil {
					db.Model(&existingAdmin).Update("password_hash", string(hash))
					log.Printf("Admin password synchronized with current .env for: %s", adminEmail)
				}
			}
		}
	} else {
		var adminCount int64
		db.Model(&User{}).Where("role = ?", "admin").Count(&adminCount)
		if adminCount == 0 {
			log.Println("Notice: No admin account found in database. Configure ADMIN_EMAIL and ADMIN_PASSWORD in .env to provision an initial administrator.")
		}
	}

	log.Println("Database schema migration completed successfully.")
	return nil
}
