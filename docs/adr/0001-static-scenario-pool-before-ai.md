# Static scenario pool before AI generation

Scenarios are seeded as static JSON data and loaded into the database via a script, rather than generated on-the-fly by an AI model.

The thesis proposes AI-generated scenarios, but during the build phase we need deterministic, reviewable content to validate the scoring engine, session recovery, domain proficiency tracking, and the full pre-test → scenario loop → post-test pipeline. Static data makes bugs in the platform distinguishable from bugs in generated content.

The system is designed so the interface between frontend and backend is identical regardless of scenario source — the backend returns a scenario object and the frontend renders it. When the platform is stable, the static pool can be replaced by an AI generation endpoint behind the same API contract.

## Considered Options

- **AI generation from day one**: Would deliver the thesis vision immediately but couples platform bugs with content-quality bugs, making both harder to diagnose.
- **Hybrid (static + AI fallback)**: Adds complexity without clear benefit during the build phase.

## Consequences

- Scenario content must be manually authored or batch-generated and reviewed before seeding.
- The `scenarios` table schema and API response shape must not assume static-only data (no hardcoded IDs in frontend logic, etc.).
