## Context

The backend's `go.mod` specifies Go 1.24.0. However, the dependency `github.com/gin-contrib/cors@v1.7.7` explicitly requires Go >= 1.25.0. This prevents `go mod download` from succeeding during the Docker build process.

## Goals / Non-Goals

**Goals:**
- Successfully build the backend Docker container without Go version mismatch errors.
- Keep the `go.mod` requirement at Go 1.24.0 to align with the chosen stable alpine builder image.

**Non-Goals:**
- Rewriting the API layer to remove `gin-contrib/cors`.

## Decisions

- **Decision: Downgrade `gin-contrib/cors` (and others if needed)**
  - **Rationale:** A slightly older version of the `cors` middleware (such as `1.7.2`) does not strictly require Go 1.25. Downgrading it resolves the build issue while retaining the functionality on Go 1.24.0.
  - **Alternatives:** Upgrade the entire project to Go 1.25.0. This is rejected because standard `golang:1.25-alpine` images might be too bleeding edge or currently unavailable, whereas `1.24` is stable.

## Risks / Trade-offs

- **Risk: Missing out on the latest bug fixes in downgraded packages.**
  - **Mitigation:** The application can function correctly with an older stable version. We can revisit upgrading when Go 1.25 becomes the universally supported default.
