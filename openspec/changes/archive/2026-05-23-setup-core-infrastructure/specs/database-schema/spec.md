## ADDED Requirements

### Requirement: Dockerized pgvector database
The system SHALL provide a PostgreSQL database with the pgvector extension enabled via Docker Compose.

#### Scenario: Database startup
- **WHEN** running `docker-compose up`
- **THEN** the database container starts and is accessible on port 5432

### Requirement: Core schema initialization
The system SHALL initialize all core tables, enums, and indexes automatically upon database creation.

#### Scenario: First run schema creation
- **WHEN** the database container is created for the first time
- **THEN** `init.sql` is executed, creating `users`, `scenarios`, `rag_documents`, and all other defined tables
