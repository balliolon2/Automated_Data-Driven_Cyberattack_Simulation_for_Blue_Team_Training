## Why

The CyberSim platform requires a secure authentication system to identify users, track their progress, and protect access to the exam and scenario simulations. Currently, the database schema has a `users` table, and the frontend has dummy UI pages for Login. We need to implement the backend API and connect the frontend to make the authentication flow functional.

## What Changes

- Implement a Go backend using the Gin web framework to provide RESTful APIs.
- Create `/api/register` and `/api/login` endpoints.
- Use `bcrypt` for secure password hashing before storing them in PostgreSQL via GORM.
- Implement JWT (JSON Web Token) generation for session management.
- Update the React frontend to have a separate `/register` route and page.
- Connect the frontend Login and Register forms to the Go backend APIs.
- Store the JWT token securely in the browser's `localStorage` and handle protected routes.

## Capabilities

### New Capabilities
- `user-auth-api`: The Go backend API endpoints for handling user registration, authentication, and token generation.
- `frontend-auth-flow`: The React components and state management for logging in, registering, and managing the JWT session.

### Modified Capabilities
- 

## Impact

- Allows users to create accounts and log in securely.
- Sets up the Gin framework on the Go backend, which will serve as the foundation for all future API endpoints (e.g., fetching questions, submitting answers).
- Transitions the frontend from static UI pages to a dynamic application with state.
