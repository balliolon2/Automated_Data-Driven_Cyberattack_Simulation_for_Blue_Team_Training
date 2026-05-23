# Automated Data-Driven Cyberattack Simulation for Blue Team Training

This is a Senior Project for Computer Engineering, Kasetsart University.

This project implements an interactive platform designed to train and assess Blue Teams (defense) through automated, data-driven cyberattack simulations. It features a robust assessment pipeline containing user authentication, randomized Pre/Post tests across core security domains, session state recovery, and analytics dashboards.

---

## 🚀 Core Features

- **Pre-Test & Post-Test Assessment**: Automated evaluation of learner skills using 100-question randomized exams. Questions are dynamically drawn from 5 distinct CompTIA Security+ domains:
  - **Domain 1**: Security Concepts (12 questions)
  - **Domain 2**: Threats, Vulnerabilities, and Mitigations (22 questions)
  - **Domain 3**: Security Architecture (18 questions)
  - **Domain 4**: Security Operations (28 questions)
  - **Domain 5**: Security Management (20 questions)
- **Zero Question Overlap**: The randomization engine ensures that questions presented in the user's Pre-Test are strictly excluded from their Post-Test.
- **Session Recovery**: State-persisted exam sessions ensure that user progress is saved. If a network disconnection or page refresh occurs, the active session is recovered automatically.
- **Data Ingestion Pipeline**: Deterministic markdown-to-JSON parser that parses CompTIA exam questions and automatically seeds security domains and questions into the database.
- **Analytics & Skill Gaps**: Interactive evaluation dashboards utilizing radar charts to display proficiency across the five security domains.

---

## 🛠️ Technology Stack

- **Backend**: Go (version 1.25), Gin Web Framework, GORM (Object-Relational Mapping)
- **Frontend**: React, Vite, TypeScript, Tailwind CSS, Recharts
- **Database**: PostgreSQL 16 with `pgvector` & `pgcrypto` extensions
- **Infrastucture**: Docker & Docker Compose

---

## 📂 Project Structure

```text
├── controllers/          # Go backend controllers (Auth, Exam controllers)
├── models/               # GORM database schemas & struct definitions
├── middlewares/          # JWT authentication middleware
├── frontend/             # React (Vite) frontend application
│   ├── src/
│   │   ├── pages/        # Dashboard, Exam, Evaluation, and Landing pages
│   │   └── components/   # Shared UI components
│   └── Dockerfile        # Frontend Docker configuration
├── examination/          # CompTIA Security+ markdown source exam files
├── scripts/              # Python automation & testing scripts
│   ├── parse_exams.py    # Parses markdown exams to JSON
│   ├── load_exams_to_db.py # Seeds security domains and inserts questions
│   └── test_integration.py # End-to-end integration test suite
├── init.sql              # Database initialization schema (tables, types, indexes)
├── Dockerfile            # Backend Docker configuration
└── docker-compose.yml    # Multi-container service configuration
```

---

## ⚙️ Setup & Running

### Prerequisites
- Docker & Docker Compose
- Python 3.x (with `psycopg2` and `requests` libraries)

### 1. Build and Run the Services
Spin up the PostgreSQL database, Go backend, and React frontend containers:
```bash
docker-compose up -d --build
```

### 2. Ingest Exam Questions
Since the pre-parsed `parsed_questions.json` is already committed in the repository, you can directly seed the security domains and load the questions into the database:
```bash
# 1. Seed security domains and load the questions into Postgres
python scripts/load_exams_to_db.py
```

*(Optional)* If you ever modify the markdown files in `examination/` and need to regenerate the JSON before loading:
```bash
# 2. (Optional) Re-parse raw markdown files in /examination to parsed_questions.json
python scripts/parse_exams.py
```

### 3. Run Integration Tests
Verify the entire API pipeline (registration, login, pre-test start, session recovery, submission, and post-test overlap protection):
```bash
python scripts/test_integration.py
```

---

## 📊 Verification Output example
When running `test_integration.py`, the output should confirm a successful end-to-end test run:
```text
Registering user: test_4037@example.com
Registration successful
Logging in...
Login successful, token retrieved
Starting Pre-Test...
Pre-Test started. Session ID: 5b2ec05a-02ce-4479-ac12-a5a859747062. Questions count: 100
Domain distribution: {'domain3': 18, 'domain1': 12, 'domain4': 28, 'domain5': 20, 'domain2': 22}
Testing Session Recovery...
Session recovery successful. Recovered questions: 100
Submitting answers...
Pre-test completed! Final Score: 0
Starting Post-Test...
Post-Test started. Session ID: a25cf54a-c2f5-40d8-9412-7160f6735177. Questions count: 100
Overlap count between Pre-Test and Post-Test: 0
Verification SUCCESS: Zero question overlap between Pre-Test and Post-Test!
```