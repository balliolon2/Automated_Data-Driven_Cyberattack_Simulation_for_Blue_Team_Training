# ADR-0003: LLM RAG-Driven Dynamic Scenario Generation with Static Pool Fallback

## Status
Accepted

## Context
The platform requires dynamic incident scenarios tailored specifically to individual learners based on their evaluated skill gaps from the CompTIA Security+ pre-test and ongoing simulation performance. The scenario generation must support standard OpenAI SDK compatible endpoints (such as OpenAI, OpenRouter, Groq, DeepSeek, LocalAI, Ollama, or vLLM), while remaining zero-downtime resilient even if LLM keys are unconfigured or external AI services experience latency/failures.

## Decision
1. **OpenAI SDK Compatible Client**:
   Implemented `services/llm_service.go` using Go standard library (`net/http` + `encoding/json`), completely eliminating external LLM dependencies (Ponytail Ultra principles). Configurable via `LLM_API_KEY`, `LLM_BASE_URL` (default: `https://api.openai.com/v1`), and `LLM_MODEL` (default: `gpt-4o-mini`).

2. **RAG Skill Gap Context**:
   When `StartScenario` is called, the backend queries the learner's `user_skill_profiles`, identifies their weakest security domain and proficiency scores, and constructs a RAG prompt instructing the LLM to generate a complete 30-50 telemetry log scenario with key findings, evidence indices, and response playbooks.

3. **Graceful Static Fallback**:
   If `LLM_API_KEY` is not provided or if the LLM request encounters any network/parsing error, the controller automatically falls back to querying the static scenario pool in PostgreSQL (`scenarios` table), ensuring the user never experiences a broken or failed session start.

4. **High-Tech Loading Experience**:
   On the frontend (`SimulationPage.tsx`), a glowing dual-axis cyber radar spinner with cycling threat-synthesis status messages is rendered while the server generates and prepares the virtual environment.

## Consequences
- **Maximum Adaptability**: Scenarios are dynamically tailored to the learner's real-time weak domains.
- **Provider Agnostic**: Works out-of-the-box with any OpenAI-compatible provider.
- **Zero Breakage**: Fallback guarantees 100% operational uptime regardless of LLM API availability.
