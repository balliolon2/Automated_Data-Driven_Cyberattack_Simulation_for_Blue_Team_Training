# Domain Notes & Terminology (NOTES.md)

## Core Roles & Terminology

- **Learner (คนเรียนรู้)**: The primary training user undergoing pre-test, scenario simulation loops, and post-test.
- **Specialist (ผู้เชี่ยวชาญ)**: An approved domain expert who reviews learner scenario answers, provides deep technical analysis, and publishes analysis threads (กระทู้).
- **Admin**: Platform administrator who verifies and approves Specialist candidate applications (evaluating resumes, certificates, and credentials) and oversees platform operations.
- **Nickname (ชื่อเล่น)**: Public display name for users across discussions, leaderboards, and reviews, preserving privacy while allowing friendly identification.
- **Specialist Application (การสมัครเป็น Specialist)**: Onboarding process requiring extra qualification artifacts (resume file/URL, professional certificates, security experience summary).
- **Analysis Thread (กระทู้การวิเคราะห์)**: A post authored by a Specialist breaking down scenario solutions, common learner mistakes, triage techniques, and key finding analysis.
- **Discussion / Comment (คอมเมนต์ / ตอบกระทู้)**: Collaborative replies and questions on an analysis thread between Learners, Specialists, and Admins.

## Settled Decisions (Round 1)

1. **Role Architecture**:
   - Roles: `learner`, `specialist`, `admin` (migrating `expert` to `specialist`).
   - Registration flow: Candidates register as `learner` by default. If applying as Specialist, an application record is created in `specialist_applications` with status `pending`.
   - Admin approves the application to promote the user role to `specialist`.
   - Existing Learners can also submit an application to upgrade to `specialist`.

2. **Credentials & Artifacts**:
   - Hybrid mode: Supports uploading resume and certificate files (PDF/Image) stored on server + optional URL links (LinkedIn, portfolio, Credly).

3. **Learner Answers Visibility**:
   - Specialists have a review dashboard to inspect completed simulation sessions filtered by Scenario, Domain, or Score.
   - Learner identity is displayed via **Nickname** to protect email privacy.
   - Specialist views: TP/FP triage decision, discovered key findings, response actions executed, log query attempts, and final score.

4. **Analysis Threads (กระทู้) & Discussion**:
   - Authoring: Restricted to `specialist` and `admin` roles.
   - Scope: Can be linked directly to a `scenario_id` or written as general blue-team analysis.
   - Interaction: 2-level comment hierarchy (top-level comment + direct reply). All roles (`learner`, `specialist`, `admin`) can participate in discussions.

5. **Nickname**:
   - Mandatory for all users at registration.
   - Must be unique across the platform.
   - Serves as the public display name in community discussions and learner review records.

## Settled Decisions (Round 2)

6. **Admin Checkpoint & Rejection Handling**:
   - Rejection requires an Admin to enter an explanation note (`rejection_reason`).
   - The applicant sees the `rejected` status with the reason on their Profile.
   - The applicant can re-apply by updating their profile details and re-uploading documents without creating a new account.

7. **Artifact Storage & File Security**:
   - Resumes and certificates stored under protected storage `/uploads/applications/{application_id}/`.
   - Accessible only to Admins and the application owner via JWT-authenticated endpoint `GET /api/admin/applications/:id/files/:file_type`.
   - Prevent public static directory exposure and enforce UUID naming.

8. **Linking Scenario Results to Analysis Threads**:
   - `ScenarioResultPage` embeds a "Specialist Insights & Discussions" section below evaluation metrics.
   - Directly links to threads tagged with that `scenario_id` so learners can immediately review expert analysis and common mistakes.

9. **Role Badges & Thread Pinning**:
   - Visual badges next to Nickname: `[Specialist]`, `[Admin]`, `[Learner]`.
   - Thread authors (`specialist`) and `admin` can pin key takeaway comments to the top of the thread.

10. **In-App Notification Loop**:
    - Navbar includes an in-app notification indicator (bell) showing:
      - Specialist application approval / rejection status updates.
      - Replies to user comments in analysis threads.

## Settled Decisions (Round 3)

11. **Rich Content & Scenario Log Embedding**:
    - Markdown editor with syntax highlighting for KQL queries, syslog/JSON dumps.
    - Quick-embed widget allowing Specialists to inject scenario summaries (alert details, key findings checklist, expected actions) into their thread.

12. **Moderation & Permissions**:
    - Authors can edit and delete their own threads and comments.
    - Thread authors (Specialists) can hide or delete inappropriate comments on their threads.
    - Admins have master moderation privileges (edit, delete, pin, lock) across all community content.

13. **Existing Account Migration**:
    - Migration script backfills `nickname` using email prefix (with random suffix if duplicate).
    - Users can update their nickname anytime in their Profile.

14. **Helpful / Upvoting System**:
    - 1-upvote per user on threads and comments to highlight quality analyses.

## Settled Decisions (CI/CD & DevSecOps Pipeline)

15. **Zero-Cost Hosting Architecture**:
    - Frontend SPA: Cloudflare Pages (Free tier, fast edge CDN, unlimited bandwidth).
    - Backend Go API: Render.com (Free Web Service tier, Docker runtime).
    - Database: Neon Serverless PostgreSQL (Free tier with pgvector extension support).

16. **Pipeline Stages & DevSecOps Gates**:
    - **Trigger Gate**: PRs to `main` trigger CI only (Linting, DevSecOps scanning, Test Harness, Build). Direct push/merge to `main` triggers CD.
    - **DevSecOps Scans**: Gitleaks (secret sniffing) + `govulncheck` (Go CVEs) + `npm audit` (Frontend dependencies).
    - **Integration Verification**: GitHub Actions spins up `pgvector/pgvector:pg16` service container, initializes `init.sql`, runs `load_exams_to_db.py`, launches Go backend, and runs `test_integration.py` to assert exam 30-question distribution & zero-overlap integrity.

17. **Continuous Delivery & Cold-Start Optimization**:
    - **Backend Deploy**: Triggered via Render Deploy Hook Webhook upon green merge to `main`.
    - **Frontend Deploy**: Automated deploy to Cloudflare Pages.
    - **Keep-Alive Worker**: Scheduled cron every 14 minutes pings the health endpoint to prevent Render Free Tier container hibernation.


