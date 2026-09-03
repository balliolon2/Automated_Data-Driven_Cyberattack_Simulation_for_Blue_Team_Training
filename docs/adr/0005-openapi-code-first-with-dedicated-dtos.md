# OpenAPI Code-First Documentation with Dedicated DTOs

The Go REST API uses code-first OpenAPI documentation generated via `swaggo/swag` and served through `gin-swagger`, with a dedicated `dto/` package isolating the API contract from database models.

Annotating controllers and isolating request/response DTOs ensures that the documentation is a reliable single source of truth without leaking database internals or sensitive credentials. It also aligns API payload terminology with the domain model in `CONTEXT.md` (e.g. `ResponseAction` instead of `PlaybookStep`). Swagger UI is exposed at `/api/docs/index.html` with bearer authentication enabled, gated by an environment flag for production environments.

## Considered Options

- **Spec-First (OpenAPI YAML)**: Handcrafted YAML specs drift from actual Gin handler implementations unless heavy code-generation pipelines are enforced.
- **Annotating GORM Models directly**: Couples database schema directly to external API responses, exposing database internals and complicating field-level documentation.

## Consequences

- All new endpoints must define request/response structs in `dto/` and include Swag annotations on controller handlers.
- The Swagger documentation must be regenerated via `go generate` or `scripts/generate_docs.ps1` whenever DTOs or endpoints change.
