## Context

The backend container fails to build with the error:
`go: go.mod requires go >= 1.25.0 (running go 1.24.13; GOTOOLCHAIN=local)`
This is because `go.mod` has been updated to require `go 1.25.0`, but the builder image in the `Dockerfile` uses `golang:1.24-alpine`.

## Goals / Non-Goals

**Goals:**
- Update the Docker build pipeline to use Go 1.25.
- Ensure the backend container builds successfully.

**Non-Goals:**
- Upgrading Go packages or addressing unrelated build issues.
- Modifying frontend Docker configuration.

## Decisions

- **Decision: Upgrade builder image in Dockerfile to golang:1.25-alpine**
  - **Rationale:** The host environment is running Go 1.25.6, meaning any local `go` commands will update `go.mod` to require Go 1.25.x. Upgrading the builder stage in the `Dockerfile` to Go 1.25 aligns the container build toolchain with the host development environment.
  - **Alternatives:** We could downgrade the `go.mod` Go directive back to `go 1.24.0`, but since Go 1.25 is installed on the host, any local Go utility invocations will continually rewrite the directive to `go 1.25.0` (or `go 1.25.6`). Thus, upgrading the Docker builder is the most robust solution.

## Risks / Trade-offs

- **Risk: Docker build fails if golang:1.25-alpine is not available in docker registry**
  - **Mitigation:** Go 1.25 is a stable release and the official `golang:1.25-alpine` image is widely available in Docker Hub.
