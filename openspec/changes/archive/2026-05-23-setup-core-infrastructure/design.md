## Context

The project is the "Automated Data-Driven Cyberattack Simulation for Blue Team Training". This change initializes the foundational backend infrastructure. The backend must handle standard application logic as well as AI/RAG capabilities based on the DBML design.

## Goals / Non-Goals

**Goals:**
- Provide a reproducible database setup using Docker Compose.
- Initialize PostgreSQL with the `pgvector` extension.
- Create all core tables and enums based on the provided DBML design.
- Scaffold a Go backend using `gorm` (with `pgvector-go`) for database access.
- Integrate `tmc/langchaingo` for future RAG AI features.

**Non-Goals:**
- Implementing the actual RAG document loading or querying logic in this change.
- Creating the full frontend UI.
- Deploying to a production environment.

## Decisions

- **Database Container:** `pgvector/pgvector:pg16` is chosen because it comes pre-installed with the required `vector` extension, saving the complexity of building a custom Dockerfile for PostgreSQL.
- **ORM:** `gorm` is chosen for its simplicity and the availability of the `github.com/pgvector/pgvector-go` plugin, making it easier for a senior project compared to raw SQL or `sqlc` (which requires more setup).
- **AI Framework:** `tmc/langchaingo` is chosen to keep the entire backend in pure Go rather than splitting into a Go API and Python AI service, reducing architectural complexity.

## Risks / Trade-offs

- **[Risk] Langchaingo maturity:** `tmc/langchaingo` is less mature than Python's LangChain.
  - *Mitigation:* For core RAG (embedding generation and similarity search via pgvector), langchaingo is sufficient. If advanced features are missing later, we can implement custom LLM API calls directly in Go.
- **[Risk] Gorm overhead:** `gorm` can be slow for very large queries.
  - *Mitigation:* For a senior project, development speed and maintainability outweigh the minor performance overhead of GORM.

## Migration Plan

- No migration is needed as this is a greenfield initialization. The database will be created fresh via `init.sql`.
