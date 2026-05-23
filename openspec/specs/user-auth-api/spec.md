## ADDED Requirements

### Requirement: User Registration API
The system SHALL provide an API endpoint to register a new user.

#### Scenario: Successful registration
- **WHEN** a POST request is made to `/api/register` with valid email and password
- **THEN** the system hashes the password, creates a user record in the database, and returns a 201 status code.

### Requirement: User Login API
The system SHALL provide an API endpoint to authenticate a user.

#### Scenario: Successful login
- **WHEN** a POST request is made to `/api/login` with correct credentials
- **THEN** the system verifies the password and returns a 200 status code with a signed JWT token.

### Requirement: Password Security
The system SHALL NOT store user passwords in plain text.

#### Scenario: Database storage
- **WHEN** a user registers or changes their password
- **THEN** the password is automatically hashed using bcrypt before being persisted to the database.
## MODIFIED Requirements

### Requirement: Database Schema Alignment
The Go backend models SHALL accurately reflect the schema defined in `init.sql` to prevent SQL errors.

#### Scenario: Registering a user
- **WHEN** a user registers
- **THEN** GORM inserts only the columns defined in the `users` table (`user_id`, `email`, `password_hash`, `role`, `current_tier`, `created_at`, `last_login`, `is_active`).

