# Architecture Research

**Domain:** Cyberattack Simulation & Training Platform
**Researched:** 2026-05-22
**Confidence:** HIGH

## Standard Architecture

### System Overview

```text
┌─────────────────────────────────────────────────────────────┐
│                     Presentation Layer                      │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐ ┌───────────────┐ ┌───────────────────┐   │
│  │ Learner UI   │ │ Admin/ExpertUI│ │ Virtual Workspace │   │
│  │ (Dashboard)  │ │ (Approval)    │ │ (SIEM/Case Mock)  │   │
│  └──────┬───────┘ └───────┬───────┘ └─────────┬─────────┘   │
│         │                 │                   │             │
├─────────┴─────────────────┴───────────────────┴─────────────┤
│                    API & Orchestration Layer                 │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐ ┌─────────────────┐ ┌────────────────┐ │
│  │ Session Manager │ │ Scoring Engine  │ │ Scenario API   │ │
│  └────────┬────────┘ └────────┬────────┘ └───────┬────────┘ │
├───────────┴───────────────────┴──────────────────┴──────────┤
│                  AI & Background Workers                    │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐ ┌─────────────────┐                    │
│  │ Scenario Builder│ │ RAG & Feedback  │                    │
│  │ (Async Queue)   │ │ (Vector Search) │                    │
│  └────────┬────────┘ └────────┬────────┘                    │
├───────────┴───────────────────┴─────────────────────────────┤
│                         Data Layer                          │
│  ┌───────────────┐ ┌───────────────┐                        │
│  │ Relational DB │ │ Vector Store  │                        │
│  │ (PostgreSQL)  │ │ (pgvector)    │                        │
│  └───────────────┘ └───────────────┘                        │
└─────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| **Learner UI** | Pre-tests, unlocking scenarios, tracking progress, skill gap visualizations. | React/Vite + TailwindCSS |
| **Virtual Workspace** | Mimics SIEM (Splunk) and Ticketing (TheHive) for log analysis without real traffic. | React/Vite + mock data services |
| **Admin/Expert UI** | Approving AI-generated content before activation, reviewing audit logs. | React/Vite |
| **Session Manager** | Maintains state of active simulation, handles user actions and log exposure. | Node.js/Go API |
| **Scoring Engine** | Evaluates `session_actions` for correctness, assigns points, updates `user_skill_profiles`. | Node.js/Go API |
| **Scenario API** | Fetches active scenarios, interfaces with builder for new content. | Node.js/Go API |
| **Scenario Builder** | Orchestrates LLM to generate scenarios, handles async job queue (<30s). | Background Worker (BullMQ/Go routines) |
| **RAG & Feedback** | Embeds queries, searches `/examination` materials, generates cited responses (<2s). | Python/Node + LLM Provider |
| **PostgreSQL DB** | Stores users, roles, test attempts, domains, session logs, and auth hashes. | PostgreSQL |
| **Vector Store** | Stores embeddings of domain theory and mitre techniques for fast retrieval. | PostgreSQL + pgvector (HNSW) |

## Recommended Project Structure

```text
src/
├── frontend/                 # React/Vite application
│   ├── features/             # Feature-based component grouping
│   │   ├── dashboard/        # Skill gaps and pre-tests
│   │   ├── siem-mock/        # Splunk-like log querying interface
│   │   ├── case-mock/        # TheHive-like ticketing interface
│   │   └── rag-chat/         # AI feedback assistant
│   └── shared/               # Shared UI components, hooks, auth
├── backend/                  # Node.js/Go API server
│   ├── api/                  # REST/GraphQL controllers & routes
│   ├── core/                 # Business logic (scoring, simulation state)
│   └── services/             # Integrations (DB, Job Queue)
├── ai-workers/               # Async AI/ML operations
│   ├── rag/                  # Embedding generation and retrieval
│   └── generation/           # Scenario synthesis and parsing
└── database/                 # Schema and migrations
    ├── schema/               # DBML and SQL definitions
    └── seeds/                # Initial CompTIA/MITRE data load
```

### Structure Rationale

- **`frontend/features/`:** Groups UI components by bounded context (e.g., SIEM mock is distinct from the dashboard). Keeps complex mock interfaces isolated.
- **`ai-workers/`:** Separating AI generation from the main API prevents long-running scenario generation (<30s) from blocking standard API requests.
- **`database/`:** Centralizes schema (users, domains, actions) ensuring both backend and AI workers share a unified data contract.

## Architectural Patterns

### Pattern 1: State Machine for Simulation Session

**What:** The `Session Manager` models a scenario as a Finite State Machine (FSM). User actions trigger state transitions which unlock new log entries.
**When to use:** When users must investigate step-by-step (e.g., finding IP A allows querying IP A's traffic).
**Trade-offs:** High precision for scoring but requires strict state definitions during scenario generation.

### Pattern 2: Retrieval-Augmented Generation (RAG) with Citation Enforcement

**What:** The RAG pipeline forces the LLM to return specific source citations (`domain_id` or `mitre_ref`) alongside generated feedback.
**When to use:** To meet the constraint that feedback must be theory-backed and traceable.
**Trade-offs:** Slightly higher latency and token usage, but ensures educational accuracy and prevents hallucinations.

### Pattern 3: Async Content Generation Pipeline

**What:** Scenario generation requests are queued. An AI worker processes the queue, generates the scenario, and flags it `scenario_status=pending_approval`.
**When to use:** To meet the <30s generation requirement and the "expert approval" constraint.
**Trade-offs:** Adds complexity (queues, websockets/polling for status updates) but protects system performance and content quality.

## Data Flow

### Request Flow: Scenario Execution & Scoring

```text
[User Action in Workspace]
    ↓
[Frontend API Client] → [Session API] → [Scoring Engine]
    ↓                        ↓                 ↓
[Updated Workspace] ← [State Check] ← [Log session_action (correctness)]
    ↓                                          ↓
[New Simulated Logs] ← [DB: Retrieve allowed logs for state]
```

### Request Flow: RAG Feedback

```text
[User Queries RAG Bot]
    ↓
[Frontend] → [API Server] → [AI Worker: RAG]
    ↓            ↓                 ↓
[Response] ← [Feedback] ← [pgvector: HNSW Search (exam MDs)]
    ↓                              ↓
[Source Citations] ← [LLM: Synthesize Context + Query]
```

## Recommended Build Order

To manage dependencies effectively, the system should be built in the following order:

1. **Foundation (Data & Auth):** Define PostgreSQL schema (DBML), setup pgvector, build user auth (encryption at rest), and load baseline `/examination` materials.
2. **AI & RAG Core:** Build the embedding pipeline and the RAG retrieval function. This ensures the <2s response metric can be tested early.
3. **Scenario Engine (Async):** Develop the async queue for scenario generation and the Admin UI for expert approval.
4. **Simulation Workspace:** Build the React SIEM/TheHive UI mocks and the Backend Session Manager to handle state and log exposure.
5. **Evaluation & Dashboards:** Implement the Scoring Engine to flag `session_actions`, calculate points, and render the post-session `skill_gap` visualizations.

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 50 concurrent | Single API instance, managed PostgreSQL, single AI worker queue. Sufficient for target. |
| 500+ concurrent | Autoscaling API nodes via GCP Cloud Run, multiple job queue workers for generation. |
| 5000+ concurrent | Read replicas for PostgreSQL, caching layer (Redis) for frequent RAG context retrieval. |

### Scaling Priorities

1. **First bottleneck:** LLM API rate limits / generation time. **Fix:** Implement robust queuing and consider batching generation requests or caching common scenario parameters.
2. **Second bottleneck:** Vector search latency as `rag_documents` grows. **Fix:** Ensure pgvector HNSW indexes are properly tuned and cached.

## Anti-Patterns

### Anti-Pattern 1: Live Traffic Injection
**What people do:** Attempting to pipe real network traffic or live PCAP replays into the training platform to increase "realism".
**Why it's wrong:** Violates project constraints (simulation only), introduces massive security risks, and makes deterministic scoring nearly impossible.
**Do this instead:** Use generated static log structures (JSON/CSV) mapped to scenario states.

### Anti-Pattern 2: Synchronous Scenario Generation
**What people do:** Waiting for the LLM to generate the scenario within the HTTP request lifecycle.
**Why it's wrong:** Will lead to HTTP timeouts and locked UI threads if the generation takes close to the 30s limit.
**Do this instead:** Use an async queue pattern and notify the user when the scenario is ready (or pending expert approval).

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| **LLM Provider** | API Client (e.g., OpenAI/Anthropic) | Must enforce JSON schema for citations and scenario structure. |
| **GCP Cloud** | Managed services (Cloud Run, Cloud SQL) | Handles auto-scaling to meet the 50 concurrent user requirement. |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| API ↔ AI Worker | Async Queue (e.g., Redis/BullMQ) | Decouples heavy AI generation from fast user requests. |
| RAG ↔ Database | Direct SQL via pgvector | Requires optimized HNSW indexing for <2s response constraint. |

## Sources

- System constraints and requirements from `.planning/PROJECT.md`
- Standard architecture patterns for Cyber Ranges / BAS (Breach and Attack Simulation) tools
- RAG architecture best practices (Vector DB + LLM synthesis)

---
*Architecture research for: Cyberattack Simulation & Training Platform*
*Researched: 2026-05-22*