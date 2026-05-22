## 1. Database Setup

- [x] 1.1 Create `docker-compose.yml` with `pgvector/pgvector:pg16` and `pgAdmin` services
- [x] 1.2 Write `init.sql` with enums and `vector` / `pgcrypto` extensions
- [x] 1.3 Add core tables (`users`, `security_domains`, `mitre_techniques`) to `init.sql`
- [x] 1.4 Add assessment tables (`questions`, `test_attempts`) to `init.sql`
- [x] 1.5 Add simulation tables (`scenarios`, `simulation_sessions`, `session_actions`) to `init.sql`
- [x] 1.6 Add AI & Feedback tables (`rag_documents`, `feedback_logs`) to `init.sql`
- [x] 1.7 Add analytics tables (`user_skill_profiles`, `evaluation_metrics`, `help_click_analytics`) to `init.sql`
- [x] 1.8 Verify the database starts and tables are created correctly using `docker-compose up -d`

## 2. Go Backend Setup

- [x] 2.1 Initialize Go module (`go mod init`)
- [x] 2.2 Install `gorm` and PostgreSQL driver
- [x] 2.3 Install `github.com/pgvector/pgvector-go`
- [x] 2.4 Setup basic `main.go` to connect to the database via GORM
- [x] 2.5 Install `github.com/tmc/langchaingo` and verify compilation
