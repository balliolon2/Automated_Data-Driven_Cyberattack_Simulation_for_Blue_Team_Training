## Why

The backend container fails to start and keeps restarting because it cannot connect to the database, throwing `failed to connect database: unknown time zone Asia/Bangkok`. The final container stage uses a minimal Alpine image that lacks the timezone database (`tzdata`).

## What Changes

- Add the `tzdata` package to the runner stage of the backend `Dockerfile`.

## Capabilities

### New Capabilities

None

### Modified Capabilities

- `go-backend-base`: Ensures successful database connection with timezone settings on startup.

## Impact

- `Dockerfile` (adds `apk add --no-cache tzdata` in the runner stage).
