## 1. Backend Dockerization
- [x] 1.1 Create `Dockerfile` in the root directory for the Go backend (multi-stage build using alpine).

## 2. Frontend Dockerization
- [x] 2.1 Create `nginx.conf` in the `frontend/` directory to configure Nginx for React Router.
- [x] 2.2 Create `Dockerfile` in the `frontend/` directory for building Vite app and serving via Nginx.

## 3. Docker Compose Configuration
- [x] 3.1 Create `docker-compose.yml` in the root directory.
- [x] 3.2 Add `db` service using `pgvector/pgvector:pg16` and mapping `./init.sql:/docker-entrypoint-initdb.d/init.sql`.
- [x] 3.3 Add `backend` service, mapping port `8080` and passing `DATABASE_URL` environment variable.
- [x] 3.4 Add `frontend` service, mapping port `80`.

## 4. Environment Variables Setup
- [x] 4.1 Update Go backend `main.go` to use `os.Getenv("DATABASE_URL")` if available, falling back to localhost for local dev.
- [x] 4.2 Update frontend `axios` base URL configuration if needed to point to `/api` (or configure Nginx to proxy `/api` requests to backend).
