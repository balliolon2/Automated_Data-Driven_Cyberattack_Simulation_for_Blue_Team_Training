## Context

The backend DSN contains `TimeZone=Asia/Bangkok`. The database connection fails in GORM if the timezone database cannot be loaded. In Alpine, this database is provided by the `tzdata` package which is not installed by default in `alpine:latest`.

## Goals / Non-Goals

**Goals:**
- Ensure the backend starts and connects to the database successfully.
- Maintain a stable timezone configuration (Asia/Bangkok).

**Non-Goals:**
- Removing the `TimeZone` setting from the database DSN.
- Upgrading Go packages or other dependencies.

## Decisions

- **Decision: Install the `tzdata` Alpine package in the final image**
  - **Rationale:** This is the standard method for resolving timezone issues in minimal Docker environments. It installs timezone files directly into the OS so that Go's `time` package can resolve the database timezone parameter correctly.
  - **Alternatives:** Import `_ "time/tzdata"` in Go `main.go`. This was rejected because installing it via the system package keeps the Go binary cleaner and is standard practice for containerized deployments.

## Risks / Trade-offs

- **Risk: Slight increase in image size**
  - **Mitigation:** The `tzdata` package size is very small (~3MB) and does not significantly impact deployment times.
