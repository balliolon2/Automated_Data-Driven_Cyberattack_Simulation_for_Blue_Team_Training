## Context

We are implementing the user authentication system for the CyberSim platform. The system needs to support user registration and login. The backend is written in Go and uses PostgreSQL. The frontend is a React + Vite application. 

## Goals / Non-Goals

**Goals:**
- Setup the Gin web framework on the Go backend.
- Implement `/api/register` and `/api/login` endpoints.
- Securely hash user passwords using `bcrypt`.
- Issue JWT tokens for authenticated sessions.
- Connect the React frontend to the backend API via `fetch` or `axios`.
- Store the JWT in `localStorage` and manage user session state in React.

**Non-Goals:**
- OAuth/Social Login (e.g., Google, GitHub).
- Password reset or email verification flows at this stage.

## Decisions

- **Framework:** Gin (`github.com/gin-gonic/gin`) because of its high performance, middleware support, and ease of use for JSON APIs.
- **Password Hashing:** `golang.org/x/crypto/bcrypt`.
- **Session Management:** JWT (`github.com/golang-jwt/jwt`). The token will contain the user's ID and role, signed with a secret key.
- **CORS:** We will configure Gin to allow cross-origin requests from the React frontend during development.
- **Frontend Storage:** `localStorage` is chosen for simplicity in this project. While HTTP-only cookies are more secure against XSS, `localStorage` is sufficient for this scope and easier to implement with a decoupled Go backend and Vite dev server.

## Risks / Trade-offs

- **[Risk] JWT Secret Exposure:** The JWT signing key must be kept secret.
  - *Mitigation:* Load the JWT secret from environment variables rather than hardcoding it.
- **[Risk] LocalStorage XSS:** Storing tokens in `localStorage` makes them accessible to JavaScript, increasing the risk of XSS attacks.
  - *Mitigation:* The React application must sanitize inputs and avoid rendering raw HTML. Given the scope of this project, `localStorage` is acceptable, but we acknowledge the trade-off.
