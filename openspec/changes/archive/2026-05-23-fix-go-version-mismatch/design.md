## Context

The system currently fails to build the backend container because the Go version specified in `go.mod` (`go 1.25.6`) requires a higher version of Go than the builder image provided in the `Dockerfile` (`golang:1.22-alpine`). Additionally, the `docker-compose.yml` file contains a `version: '3.8'` attribute, which is obsolete in modern Docker Compose and triggers a warning during deployment.

## Goals / Non-Goals

**Goals:**
- Fix the Go version mismatch to allow `docker-compose up -d` to succeed.
- Eliminate the obsolete version warning from Docker Compose.

**Non-Goals:**
- Upgrading Go packages or addressing unrelated build issues.
- Modifying frontend Docker configuration.

## Decisions

- **Decision: Downgrade `go.mod` to `go 1.24.0` and update `Dockerfile` to `golang:1.24-alpine`**
  - **Rationale:** `1.25.6` is not yet available as a standard Alpine builder image (it might have been a typo or pre-release dependency). We will align both the module and the builder to Go `1.24`, which is the latest stable release. This resolves the mismatch while staying up to date.
  - **Alternatives:** Downgrade `go.mod` to `1.22.x` and leave the `Dockerfile` at `1.22-alpine`. Given modern dependency requirements, `1.24` is a better long-term default if compatible.

- **Decision: Remove `version` attribute from `docker-compose.yml`**
  - **Rationale:** The `version` attribute is deprecated by the Compose specification. Removing it suppresses the warning without affecting functionality.

## Risks / Trade-offs

- **Risk: Go version misalignment with developer local environments**
  - **Mitigation:** Document the Go version change so developers using `go run` locally can ensure their system Go version is at least `1.24.0`.
