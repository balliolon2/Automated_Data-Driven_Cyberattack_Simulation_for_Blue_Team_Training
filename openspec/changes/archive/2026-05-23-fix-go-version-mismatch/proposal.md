## Why

The backend fails to build via `docker-compose up -d` due to a Go version mismatch. The `go.mod` file requires `go 1.25.6`, but the `Dockerfile` pulls an older `golang:1.22-alpine` image. Additionally, the `docker-compose.yml` uses an obsolete `version: '3.8'` attribute which throws a warning during deployment.

## What Changes

- Update the `Dockerfile` to use a `golang:1.24-alpine` builder image (or match the Go version).
- Downgrade `go.mod` to `go 1.24` (or match the Dockerfile's version).
- Remove the obsolete `version: '3.8'` attribute from `docker-compose.yml`.

## Capabilities

### New Capabilities
- `docker-build`: Backend container build specification

### Modified Capabilities
None

## Impact

- `Dockerfile` (Go builder image tag)
- `go.mod` (Go version requirement)
- `docker-compose.yml` (removed version attribute)
