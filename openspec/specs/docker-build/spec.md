## Purpose
Define the container build process for the backend of the cyberattack simulation platform.
## Requirements
### Requirement: Backend Container Build
The system MUST build the backend container successfully using Go version 1.25.x in `go.mod` and the `Dockerfile` builder image.

#### Scenario: Building the backend container
- **WHEN** `docker-compose up -d` is executed
- **THEN** the backend container builds successfully without Go version mismatch errors.
- **AND** no obsolete version warnings are thrown by Docker Compose.

