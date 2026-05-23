## Why

User registration currently fails with a 500 Internal Server Error (`column "updated_at" of relation "users" does not exist`). This occurs because the `User` struct in the Go backend includes GORM's default fields (`UpdatedAt`, `DeletedAt`), which do not exist in the database schema initialized via `init.sql`. The Go model needs to be aligned with the actual database schema so that GORM can execute insert queries successfully.

## What Changes

- Modify the `User` struct in `models/user.go` to exactly match the columns defined in `init.sql`.
- Remove `UpdatedAt` and `DeletedAt`.
- Add `LastLogin` (`*time.Time`) and `IsActive` (`bool`).

## Capabilities

### Modified Capabilities
- `user-auth-api`: The backend data access layer for the user authentication API will now correctly interface with the PostgreSQL database.

## Impact

- Resolves the 500 Internal Server Error during user registration.
- Allows user authentication flows to work end-to-end.
