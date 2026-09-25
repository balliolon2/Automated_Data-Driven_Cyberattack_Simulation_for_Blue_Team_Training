# 03: Continuous Delivery & Free Cloud Deployment Automation

**What to build:** Implement an automated Continuous Delivery (CD) workflow triggered upon every push or merge into the `main` branch after CI passes. The workflow triggers the deployment of the Go backend Docker service to Render Free Web Service via Render's Deploy Hook URL and triggers the deployment of the React production bundle to Cloudflare Pages (with zero hosting cost and free SSL). Learners and evaluators immediately see updated features on the live demo environment without requiring manual intervention from the development team.

**Blocked by:** 02: Ephemeral Database & Exam Harness Verification Pipeline

**Status:** completed

- [x] A dedicated CD workflow `.github/workflows/cd.yml` is created and configured to execute exclusively on `push` to `main`.
- [x] Backend deployment step triggers Render Web Service redeployment using an authenticated HTTP POST to `RENDER_DEPLOY_HOOK_URL` secret.
- [x] Frontend deployment step packages `frontend/dist` and deploys to Cloudflare Pages using Cloudflare GitHub integration or Wrangler GitHub Action with `CLOUDFLARE_API_TOKEN` and `CLOUDFLARE_ACCOUNT_ID`.
- [x] Environment variable `VITE_API_URL` is correctly injected to point the frontend client to the production backend endpoint on Render.
- [x] Workflow provides a concise deployment summary and status check in the GitHub commit / release feed.
