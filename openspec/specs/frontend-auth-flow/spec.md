## ADDED Requirements

### Requirement: Registration UI
The frontend SHALL provide a user interface for new users to register.

#### Scenario: Registering
- **WHEN** a user fills out the registration form and submits
- **THEN** a request is sent to the backend API and the user is redirected to the login page upon success.

### Requirement: Login and Session Management
The frontend SHALL handle user login and manage the authenticated session.

#### Scenario: Logging in
- **WHEN** a user submits the login form with valid credentials
- **THEN** the received JWT token is stored in `localStorage` and the user is redirected to the Exam or Dashboard page.
