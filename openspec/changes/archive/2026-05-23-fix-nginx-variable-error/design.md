## Context

The Nginx configuration file in the frontend application contains a typo in the `X-Forwarded-For` header setup, using `$proxy_addrs` instead of `$proxy_add_x_forwarded_for`. This causes Nginx to crash immediately upon container startup.

## Goals / Non-Goals

**Goals:**
- Correct the Nginx configuration to allow the container to start.
- Ensure reverse proxying to the backend API functions as expected.

**Non-Goals:**
- Adding new Nginx routes or modifications unrelated to the startup crash.

## Decisions

- **Decision: Replace `$proxy_addrs` with `$proxy_add_x_forwarded_for` in `nginx.conf`**
  - **Rationale:** `$proxy_add_x_forwarded_for` is the standard Nginx variable for client IP forwarding. Modifying this typo solves the startup crash without changing routing behavior.
  - **Alternatives:** Remove the `X-Forwarded-For` header line completely. This is rejected because forwarding the client IP is important for security and analytics logging in the backend.

## Risks / Trade-offs

None
