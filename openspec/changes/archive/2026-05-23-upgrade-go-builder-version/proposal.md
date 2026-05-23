## Why

The backend fails to build via `docker-compose up -d` because the Go version specified in `go.mod` (`go 1.25.0`) is newer than the builder image used in the `Dockerfile` (`golang:1.24-alpine`). Because the host system has Go 1.25.6 installed, running Go commands locally automatically updates `go.mod` to Go 1.25, which continuously breaks the Go 1.24 Docker build.

## What Changes

- Update the `Dockerfile` to use a `golang:1.25-alpine` builder image.

## Capabilities

### New Capabilities

None

### Modified Capabilities

- `docker-build`: Aligns the Go version requirements for container builds.

## Impact

- `Dockerfile`: Updates the builder stage base image.
