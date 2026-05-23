## 1. Backend Authentication API (Go)

- [x] 1.1 Install Go dependencies (`gin-gonic/gin`, `golang.org/x/crypto/bcrypt`, `golang-jwt/jwt`)
- [x] 1.2 Update `main.go` to setup Gin router and connect to Postgres via GORM
- [x] 1.3 Create `User` model struct mapped to the `users` table
- [x] 1.4 Implement `/api/register` endpoint (parse JSON, hash password, save to DB)
- [x] 1.5 Implement `/api/login` endpoint (parse JSON, find user, compare hash)
- [x] 1.6 Implement JWT token generation on successful login
- [x] 1.7 Add CORS middleware to Gin to allow requests from the React dev server

## 2. Frontend Authentication UI (React)

- [x] 2.1 Install `axios` or configure `fetch` utility for API calls
- [x] 2.2 Create `src/pages/RegisterPage.tsx` with a registration form
- [x] 2.3 Update `App.tsx` to add `/register` route
- [x] 2.4 Update `LoginPage.tsx` form to manage state (email, password) and submit to `/api/login`
- [x] 2.5 Update `RegisterPage.tsx` form to manage state and submit to `/api/register`
- [x] 2.6 Implement `localStorage` logic to save JWT token upon successful login
- [x] 2.7 Update `Layout.tsx` header to show "Logout" if user is logged in
