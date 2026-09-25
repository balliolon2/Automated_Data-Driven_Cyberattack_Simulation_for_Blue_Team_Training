# 01: Automated Code Quality & DevSecOps Scanning Gate

**What to build:** Implement the primary continuous integration gate on GitHub Actions that runs on every Pull Request targeting `main`. This gate validates code formatting, static typing, and syntax for both the Go backend and React TypeScript frontend. In addition, it integrates an automated DevSecOps security audit suite including secret scanning with Gitleaks (preventing leaks of API keys, JWT secrets, or DB credentials), dependency vulnerability checking with `govulncheck` for Go, and package vulnerability auditing with `npm audit` for frontend dependencies. Any failure in formatting, typing, or critical vulnerabilities terminates the pipeline and prevents merging.

**Blocked by:** None (can start immediately)

**Status:** completed

- [x] A GitHub Actions workflow job executes on all Pull Requests targeting `main` and on pushes to feature branches.
- [x] Go backend code is validated using `go vet ./...` and `golangci-lint` to catch syntax errors and performance issues.
- [x] Frontend code is validated using `npm run lint` (ESLint) and `npx tsc -b` (TypeScript compilation check).
- [x] Automated secret detection using `gitleaks/gitleaks-action` scans commit history in the PR for sensitive tokens and keys.
- [x] Go packages are scanned for known CVE vulnerabilities using official `govulncheck ./...`.
- [x] Frontend dependencies are scanned using `npm audit --audit-level=high` to ensure no high or critical CVEs exist.
- [x] If any check fails, GitHub Actions reports a failing status check that blocks PR merge.
