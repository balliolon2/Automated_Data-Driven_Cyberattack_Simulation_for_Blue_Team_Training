# 04: Production Liveness & Cold-Start Keep-Alive Automation

**What to build:** Implement an automated scheduled liveness worker using GitHub Actions cron (or integrated uptime monitor) that executes every 14 minutes. This worker performs a lightweight HTTP GET request to the deployed Render backend health endpoint, preventing Render's free tier from entering hibernation (which occurs after 15 minutes of inactivity). This guarantees that evaluators, students, and instructors experience instant loading and no 30–50 second cold-start delays when accessing the platform, while remaining completely within the zero-cost tier.

**Blocked by:** 03: Continuous Delivery & Free Cloud Deployment Automation

**Status:** completed

- [x] A scheduled workflow `.github/workflows/keep_alive.yml` is created with cron trigger `*/14 * * * *`.
- [x] Workflow also supports manual trigger via `workflow_dispatch` for on-demand health testing.
- [x] The step executes a lightweight `curl -s -f` request to the backend health endpoint (e.g. `GET /api` or `GET /api/health`).
- [x] Failed requests are logged cleanly without causing persistent alerting noise during scheduled maintenance windows.
- [x] Documentation is added to the repository README or deployment docs guiding how to configure the backend health URL secret.
