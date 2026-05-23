## MODIFIED Requirements

### Requirement: Frontend Web Server
The React application SHALL be served via an Nginx web server in its container, using a valid and error-free configuration.

#### Scenario: Client-side routing
- **WHEN** a user requests a deep link like `/exam` directly
- **THEN** Nginx serves `index.html` allowing React Router to handle the route, preventing 404 Not Found errors.
