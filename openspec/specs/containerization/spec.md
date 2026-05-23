## Purpose
Define the containerization layout and specifications for serving the full-stack cyberattack simulation platform.
## Requirements
### Requirement: Full Stack Containerization
The system SHALL provide a `docker-compose.yml` that defines the database, backend, and frontend services.

#### Scenario: Starting the stack
- **WHEN** a user runs `docker-compose up -d`
- **THEN** all three services start successfully and communicate with each other.

### Requirement: Frontend Web Server
The React application SHALL be served via an Nginx web server in its container, using a valid and error-free configuration.

#### Scenario: Client-side routing
- **WHEN** a user requests a deep link like `/exam` directly
- **THEN** Nginx serves `index.html` allowing React Router to handle the route, preventing 404 Not Found errors.

