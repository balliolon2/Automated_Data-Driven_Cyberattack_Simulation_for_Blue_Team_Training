## Context

The project currently runs via manual commands (`go run main.go`, `npm run dev`) with a local PostgreSQL instance. To ensure the application is easily deployable and can run consistently in any environment, we are migrating to a Dockerized architecture using Docker Compose.

## Goals / Non-Goals

**Goals:**
- Containerize the Go backend.
- Containerize the React frontend using a multi-stage build that serves static files via Nginx.
- Configure `docker-compose.yml` to orchestrate `db`, `backend`, and `frontend` services.
- Ensure the `init.sql` script runs automatically when the database container starts.
- Configure proper networking and environment variables.

**Non-Goals:**
- Setting up a CI/CD pipeline or Kubernetes cluster.
- Configuring HTTPS/SSL (will run on HTTP for now).

## Decisions

- **Database Container:** Use `pgvector/pgvector:pg16` image since we need the `vector` extension.
- **Backend Dockerfile:** Multi-stage build using `golang:1.22-alpine`. Build the binary in the first stage and copy it to a minimal `alpine` image in the second stage.
- **Frontend Dockerfile:** Multi-stage build using `node:20-alpine` to build the Vite app, then copy `dist` to `nginx:alpine`.
- **Nginx Config:** Create `nginx.conf` to redirect all 404s to `index.html` to support React Router's client-side routing.
- **Environment Variables:** Define standard database connection strings in `docker-compose.yml` and pass them to the backend container.
- **Ports:** 
  - `db`: 5432 (mapped to 5432)
  - `backend`: 8080 (mapped to 8080)
  - `frontend` (Nginx): 80 (mapped to 80)

## Risks / Trade-offs

- **[Risk] Development Cycle:** Using Docker for active development can be slower than running locally if hot-reloading isn't configured.
  - *Mitigation:* We will keep `docker-compose.yml` optimized for production-like running. Developers can still run `npm run dev` and `go run main.go` against the `db` container for hot-reloading during active coding.
