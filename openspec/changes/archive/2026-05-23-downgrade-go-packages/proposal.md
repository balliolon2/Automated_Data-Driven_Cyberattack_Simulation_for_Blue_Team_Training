## Why

When building the backend container, the `github.com/gin-contrib/cors@v1.7.7` package requires Go >= 1.25.0, causing the `go mod download` step to fail. Since we are using Go 1.24.0 for stability and builder image availability, we need to downgrade this package (and potentially others) to a version compatible with Go 1.24.0.

## What Changes

- Downgrade `github.com/gin-contrib/cors` to a version compatible with Go 1.24.0 (e.g., `v1.7.2` or similar).
- Verify if any other direct dependencies (like `github.com/gin-gonic/gin`) enforce Go 1.25 and downgrade them if necessary.

## Capabilities

### New Capabilities
- `package-downgrade`: Compatible dependency specification

### Modified Capabilities
None

## Impact

- `go.mod` (dependency versions)
- `go.sum` (dependency checksums)
