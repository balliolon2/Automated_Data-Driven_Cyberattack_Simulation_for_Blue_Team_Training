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
