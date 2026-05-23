## Why

To simplify local development and prepare for production deployment, the CyberSim platform needs a containerized environment. A `docker-compose` setup allows developers to spin up the entire stack—PostgreSQL database, Go backend, React frontend, and an Nginx web server—with a single command (`docker-compose up`). This ensures consistency across different development machines and eliminates the need to manually install dependencies or start services individually.

## What Changes

- Create a `Dockerfile` for the Go backend (multi-stage build).
- Create a `Dockerfile` for the React frontend (multi-stage build with Nginx).
- Create a `docker-compose.yml` file to orchestrate the following services:
  - `db`: PostgreSQL with pgvector and pgcrypto, using `init.sql`.
  - `backend`: The Go API server.
  - `frontend`: The React application served by Nginx.
- Configure networking and environment variables so the services can communicate with each other correctly.
- Add an Nginx configuration file for the frontend to handle React Router's client-side routing.

## Capabilities

### New Capabilities
- `containerization`: The ability to run the full application stack in isolated Docker containers via Docker Compose.

### Modified Capabilities
- 

## Impact

- Development setup becomes trivial: just run `docker-compose up -d`.
- Avoids "it works on my machine" issues by standardizing the OS and runtime environments.
- The frontend will be served through Nginx, simulating a production-like architecture.
