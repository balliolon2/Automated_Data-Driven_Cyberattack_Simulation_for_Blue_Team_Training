# Technology Stack

**Project:** Automated Data-Driven Cyberattack Simulation for Continuous Blue Team Workforce Training
**Researched:** 2026-05-22

## Recommended Stack

### Core Framework (Frontend)
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| React + Vite | 18.x/19.x + 6.x | SPA Frontend | Dictated by PROJECT.md constraint. Vite provides lightning-fast HMR and optimized builds. |
| TypeScript | 5.x | Language | End-to-end type safety across the stack, reducing runtime errors. |
| Tailwind CSS | 3.4/4.x | Styling | Dictated by PROJECT.md. Rapid UI prototyping to mimic Splunk/TheHive dashboards. |
| shadcn/ui | Latest | UI Components | Provides unstyled, accessible components that integrate perfectly with Tailwind. |
| TanStack Query | 5.x | Data Fetching | Manages server state, caching, and loading states seamlessly. |

### Core Framework (Backend)
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Node.js | 22.x LTS | Runtime | Dictated by PROJECT.md. Ensures excellent ecosystem support and non-blocking I/O. |
| Hono (or Fastify) | Latest | HTTP Server | Extremely low overhead and high performance compared to Express, crucial for the `<2s` RAG response time target. |
| BullMQ | 5.x | Job Queue | Best-in-class robust Redis-based queue for Node.js. Needed for the `<30s` async scenario generation constraint. |

### Database & ORM
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| PostgreSQL | 16+ | Primary DB | Dictated by PROJECT.md. Relational structure required for the defined DBML schema. |
| pgvector | 0.6+ | Vector Store | Dictated by PROJECT.md. Enables native HNSW indexing for RAG similarity search without requiring a separate vector DB. |
| Drizzle ORM | 0.30+ | ORM | Lightweight, highly performant, type-safe. Allows raw SQL for complex pgvector similarity queries while keeping type safety. |
| Redis | 7.x | Queue/Cache Storage | Necessary backend for BullMQ job queue. |

### AI & RAG
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| OpenAI API | Latest | LLM Provider | Industry standard for text generation. |
| Vercel AI SDK | 3.x | AI Integration | Streamlines prompt management, tool calling, and API communication with less bloat than LangChain. |
| OpenAI Embeddings | text-embedding-3-small | Embedding Model | Fast, cheap, and highly performant for converting Markdown exam assets into vectors. |

### Infrastructure (GCP)
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| GCP Cloud Run | N/A | Serverless Compute | Dictated by PROJECT.md (GCP Autoscaling). Scales instantly to 50 concurrent learners and scales to zero when unused. |
| Cloud SQL | 16 | Managed DB | Fully managed PostgreSQL with native pgvector support. |
| Memorystore | 7 | Managed Redis | Managed Redis required for the BullMQ instance. |
| Cloud Storage | N/A | Asset Storage | For serving the raw `/examination/` MD files if needed outside of DB. |

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| Backend | Node.js (TypeScript) | Go | While Go is highly performant, Node.js allows for a unified TypeScript full-stack ecosystem, sharing types/schemas (like Drizzle schemas) between the React frontend and backend. |
| ORM | Drizzle ORM | Prisma | Prisma's rust-based query engine can add overhead and makes raw pgvector HNSW index queries cumbersome compared to Drizzle's SQL-like builder. |
| AI Framework | Vercel AI SDK | LangChain.js | LangChain introduces excessive abstraction and latency, conflicting with the strict `<2s` response time constraint. |
| State Management | TanStack Query + Zustand | Redux Toolkit | Redux introduces unnecessary boilerplate for an application focused heavily on async server data rather than complex client state. |

## Installation

```bash
# Frontend
npm create vite@latest frontend -- --template react-ts
cd frontend
npm install tailwindcss @tailwindcss/vite
npm install @tanstack/react-query zustand lucide-react

# Backend
npm init -y
npm install hono @hono/node-server
npm install pg drizzle-orm bullmq ioredis ai @ai-sdk/openai
npm install -D typescript tsx drizzle-kit @types/pg
```

## Anti-Patterns & Pitfalls to Avoid

- **Do NOT use Prisma for pgvector:** Prisma struggles with advanced PostgreSQL extensions like HNSW indexes and proximity searches. Use Drizzle ORM or Kysely for fine-grained SQL control.
- **Do NOT use thick AI frameworks (LangChain):** Given the `<2s` response time requirement, you need raw performance and minimal library overhead. Use standard SDKs (like Vercel AI SDK Core) to keep latency low.
- **Do NOT reinvent job queues:** Async scenario generation `<30s` requires a durable queue. Rely on BullMQ/Redis instead of in-memory maps or setTimeout, which will fail during GCP Cloud Run auto-scaling.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Frontend Stack | HIGH | React + Vite + Tailwind is explicitly requested and remains the robust industry standard. |
| Backend Stack | HIGH | Node.js + Hono/Fastify provides the concurrency needed. BullMQ is the de facto standard for Node queues. |
| Database & ORM | HIGH | PostgreSQL + pgvector is requested. Drizzle ORM is widely accepted as the optimal TS ORM for advanced Postgres usage. |
| Cloud/Infra | HIGH | Cloud Run + Cloud SQL directly address the autoscaling (50 concurrent) and pgvector constraints on GCP. |

## Sources
- System constraints from `.planning/PROJECT.md`
- Modern Node.js ecosystem standards (2025/2026)
- Drizzle ORM documentation for pgvector integration
