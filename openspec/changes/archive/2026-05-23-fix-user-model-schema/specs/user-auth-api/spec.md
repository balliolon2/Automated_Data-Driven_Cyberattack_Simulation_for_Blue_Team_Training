## MODIFIED Requirements

### Requirement: Database Schema Alignment
The Go backend models SHALL accurately reflect the schema defined in `init.sql` to prevent SQL errors.

#### Scenario: Registering a user
- **WHEN** a user registers
- **THEN** GORM inserts only the columns defined in the `users` table (`user_id`, `email`, `password_hash`, `role`, `current_tier`, `created_at`, `last_login`, `is_active`).
