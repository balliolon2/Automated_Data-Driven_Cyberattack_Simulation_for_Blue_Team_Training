# SOC Trainer REST API Documentation Guide

This document outlines how to access, authenticate with, and utilize the interactive **OpenAPI (Swagger UI)** documentation for the **SOC Trainer** backend platform.

---

## 🌐 Quick Access URLs

When the backend container or local Go server is running, you can access the documentation through the following endpoints:

| Interface | URL | Description |
| :--- | :--- | :--- |
| **Swagger UI** | [`http://localhost:8080/api/docs/index.html`](http://localhost:8080/api/docs/index.html) | Interactive web interface to explore and execute API calls directly. |
| **OpenAPI Spec (JSON)** | [`http://localhost:8080/api/docs/doc.json`](http://localhost:8080/api/docs/doc.json) | Machine-readable OpenAPI specification (useful for client generators or Postman imports). |
| **OpenAPI Spec (YAML)** | `docs/swagger.yaml` | Static repository file version for CI/CD checks and documentation parsers. |

---

## 🔐 Authentication & Interactive Testing Workflow

Protected endpoints (such as `/api/exams/*`, `/api/simulation/*`, and `/api/analytics/*`) require a signed JSON Web Token (JWT) passed via the `Authorization` header.

### Step-by-Step Authorization Flow:

1. **Obtain a JWT Token**:
   - Scroll down to the **Auth** section in Swagger UI.
   - Expand `POST /api/login` (or `POST /api/register` if testing a new user account).
   - Click **Try it out**, fill in the credentials:
     ```json
     {
       "email": "analyst@soc.local",
       "password": "password123"
     }
     ```
   - Click **Execute**.
   - Copy the value of the `token` string from the response payload.

2. **Authorize Swagger UI**:
   - Scroll to the top right of the Swagger UI interface and click the green **Authorize 🔓** button.
   - In the modal dialog under **BearerAuth (apiKey)**, type `Bearer ` followed by your copied token:
     ```text
     Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
     ```
   - Click **Authorize**, then click **Close**.
   - The lock icon will now appear locked 🔒, indicating that all subsequent requests will automatically include the JWT bearer token.

---

## 📚 Endpoint Categories

The API is logically organized into four primary domain tags:

### 1. `Auth`
Handles user account lifecycle and token generation.
- `POST /api/register`: Registers a new learner profile.
- `POST /api/login`: Authenticates an existing user and returns an authentication token.

### 2. `Exams`
Governs the CompTIA Security+ 701 Pre-Test and Post-Test evaluation cycles.
- `GET /api/exams/sample`: Public preview endpoint retrieving a randomized test question.
- `POST /api/exams/pre-test`: Initializes or resumes the 30-question Pre-Test (domain distribution: 4/7/5/8/6).
- `GET /api/exams/session`: Hydrates the learner's active exam session and questions upon reload.
- `POST /api/exams/submit-answer`: Submits single question answers and computes domain scores upon completion.
- `POST /api/exams/post-test`: Initializes the 30-question Post-Test with guaranteed **zero question overlap** from the Pre-Test.

### 3. `Simulation`
Orchestrates the adaptive training loop and OT/IT incident investigations.
- `GET /api/simulation/status`: Checks domain proficiency percentages against the 70.0% threshold.
- `POST /api/simulation/start`: Generates or retrieves an incident scenario targeting the learner's weakest domain.
- `GET /api/simulation/session`: Restores active simulation scenarios and telemetry logs.
- `POST /api/simulation/log-query`: Executes and audits search queries against synthetic SIEM/telemetry logs.
- `POST /api/simulation/submit`: Evaluates incident response triage (TP/FP classification, Key Findings, Response Actions) and updates learner proficiency.
- `GET /api/simulation/result/{sessionId}`: Delivers detailed retrospective grading and log investigation audit trails.

### 4. `Analytics`
Exposes telemetry data and research metrics compliant with the formal evaluation protocol.
- `GET /api/analytics/research-summary`: Aggregates pre/post score deltas, scenario loop counts, and overall proficiency gains.

---

## 🛠️ Regenerating the Documentation

Whenever you modify controller annotations or Data Transfer Objects (`dto/*.go`), update the generated OpenAPI specification using the helper scripts:

### Windows (PowerShell):
```powershell
powershell -ExecutionPolicy Bypass -File scripts/generate_docs.ps1
```

### Linux / macOS / Docker (Bash):
```bash
chmod +x scripts/generate_docs.sh
./scripts/generate_docs.sh
```

### Go Generate:
```bash
go generate ./...
```

---

## ⚙️ Configuration & Security Gating

By default, the Swagger documentation is active across development and test environments.

To disable the Swagger UI in production environments, set the following environment variable:

```bash
ENABLE_SWAGGER=false
```

When set to `false`, requests to `/api/docs/*` will return HTTP 404.
