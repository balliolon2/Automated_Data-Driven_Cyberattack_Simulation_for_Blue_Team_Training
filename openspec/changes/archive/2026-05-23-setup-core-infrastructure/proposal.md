## Why

We are building the Automated Data-Driven Cyberattack Simulation for Blue Team Training as a senior project. To start, we need to establish the core infrastructure including the database schema with pgvector support for our RAG (Retrieval-Augmented Generation) AI feedback system, and the primary Go backend framework.

## What Changes

- Initialize a `docker-compose.yml` to spin up PostgreSQL with `pgvector` and `pgAdmin`.
- Create the initial database schema (`init.sql`) including core tables, enums, assessments, simulation, and analytics.
- Setup a Go backend skeleton using `gorm` (with `pgvector-go`) to connect to the database.
- Integrate `tmc/langchaingo` in the Go backend for the RAG implementation.

## Capabilities

### New Capabilities
- `database-schema`: Core DB schema setup with pgvector.
- `go-backend-base`: Base setup of the Go backend with DB connection and langchaingo integration.

### Modified Capabilities
- 

## Impact

- Creates the foundational data layer and backend service for all future development.
- Allows team members to run the entire backend + database locally via Docker Compose.
