---
title: "Spec 1: User Roles, Nickname & Specialist Onboarding/Approval Pipeline"
labels: ["ready-for-agent"]
status: "ready-for-agent"
---

# Spec 1: User Roles, Nickname & Specialist Onboarding/Approval Pipeline

## Problem Statement

The platform currently operates without explicit role-based distinctions beyond a default learner role, preventing the integration of human cyber defense experts (Specialists) and platform administrators (Admins). Users also lack a friendly, non-identifying public display name (Nickname), risking the exposure of private email addresses in collaborative contexts. Additionally, there is no structured mechanism for subject-matter experts to submit proof of their credentials (such as Resumes and Professional Certificates) or for Administrators to vet, verify, and approve candidates prior to granting elevated specialist privileges.

## Solution

1. Extend the platform identity model with a formal three-tier role structure (`learner`, `specialist`, and `admin`) and enforce a mandatory, unique `nickname` across all users, backfilling existing accounts gracefully.
2. Establish a secure onboarding and application pipeline allowing new applicants and existing learners to apply for the Specialist role by providing professional credentials (resume and certificate files, LinkedIn/portfolio links, and a professional bio).
3. Provide an Administrative checkpoint portal where Admins can securely inspect submitted qualifications and either approve the candidate (immediately promoting their role to `specialist`) or reject the application with actionable feedback, allowing candidates to revise and re-apply from their profile.

## User Stories

1. As a new user, I want to register an account with an email, password, and a unique nickname, so that I have a personalized and privacy-safe identity on the platform.
2. As a new user, I want the system to reject my registration if my desired nickname is already taken, so that every user has a distinct display identity.
3. As a prospective specialist, I want to apply for the Specialist role during initial registration by uploading my resume and professional certificate, so that I do not need to go through multiple onboarding steps.
4. As an existing learner, I want to access an "Apply for Specialist" option from my profile page, so that I can upgrade my existing account without losing my training history.
5. As a specialist applicant, I want to provide external professional references (such as LinkedIn and Credly/portfolio links) along with a brief cybersecurity bio, so that administrators have context on my expertise.
6. As a specialist applicant, I want file size and file type restrictions enforced on my uploads, so that I know immediately if my submitted documents are valid (e.g., PDF or images under 10MB).
7. As a specialist applicant awaiting review, I want my account to function as a standard learner, so that I can continue practicing scenarios while my application is evaluated.
8. As a platform administrator, I want to view a centralized list of all pending specialist applications, so that I can review candidate submissions in a timely manner.
9. As a platform administrator, I want to securely inspect and download candidate resume and certificate files through authenticated endpoints, so that sensitive applicant documents are never exposed to the public.
10. As a platform administrator, I want to approve a qualified applicant with a single action, so that their account is promoted to Specialist and they can immediately begin analyzing scenarios.
11. As a platform administrator, I want to reject an unqualified or incomplete application by providing a mandatory explanation note, so that the applicant understands what is missing or required.
12. As a specialist applicant whose application was rejected, I want to see the administrator's feedback on my profile page, so that I know exactly why my application was declined.
13. As a specialist applicant whose application was rejected, I want to update my information, upload new documents, and re-apply directly from my profile, so that I do not need to create a new account.
14. As an approved specialist, I want my subsequent logins and JWT authentication tokens to reflect my `specialist` role, so that client applications can unlock expert-only tools and routes.
15. As an existing user registered before the nickname requirement, I want the system to automatically assign me a default nickname from my email prefix during database migration, so that my account remains functional without manual intervention.

## Implementation Decisions

### Architectural Decisions & State Management
- **Role Migration**: Migrate the PostgreSQL `role` enum type from `('learner', 'expert', 'admin')` to `('learner', 'specialist', 'admin')`.
- **Application State Lifecycle**: Candidate applications move through an explicit state progression:
  ```
  [Initiated] ──> [Pending] ──(Admin Approve)──> [Approved] (User Role -> specialist)
                     │
                     └──(Admin Reject + Reason)──> [Rejected] ──(Candidate Re-apply)──> [Pending]
  ```
- **Document Security**: Uploaded resumes and certificates are persisted to a protected volume path (`/uploads/applications/{application_id}/`) with randomly generated UUID filenames. Files are never served statically or publicly; they are accessible only via a JWT-authenticated route verifying that the caller is an Admin or the application owner.

### Schema Specifications

- **Users Table Modifications**:
  - `nickname`: `VARCHAR UNIQUE NOT NULL`
  - `role`: PostgreSQL enum `role` updated to `('learner', 'specialist', 'admin')` with default `'learner'`

- **Specialist Applications Table**:
  - `application_id`: UUID PK, default `gen_random_uuid()`
  - `user_id`: UUID FK referencing `users(user_id)` with `ON DELETE CASCADE`
  - `status`: VARCHAR(20) NOT NULL default `'pending'` (`pending`, `approved`, `rejected`)
  - `bio`: TEXT NOT NULL
  - `resume_path`: VARCHAR(255) NOT NULL
  - `certificate_path`: VARCHAR(255)
  - `linkedin_url`: VARCHAR(255)
  - `portfolio_url`: VARCHAR(255)
  - `rejection_reason`: TEXT
  - `reviewed_by`: UUID FK referencing `users(user_id)`
  - `reviewed_at`: TIMESTAMP
  - `created_at`: TIMESTAMP default `now()`
  - `updated_at`: TIMESTAMP default `now()`

### API Contracts

- `POST /api/register`
  - Multipart form or JSON payload including `email`, `password`, `nickname`. Supports optional multipart fields for `is_specialist_applicant`, `bio`, `resume`, `certificate`, `linkedin_url`, `portfolio_url`.
  - Response: 201 Created with confirmation message.
- `POST /api/specialist/apply`
  - Authenticated (Bearer JWT). Multipart form with `bio`, `resume` (file), `certificate` (optional file), `linkedin_url`, `portfolio_url`.
  - Response: 201 Created with application summary.
- `GET /api/specialist/application-status`
  - Authenticated (Bearer JWT). Returns current user's latest application status, rejection reason (if any), and submission dates.
- `GET /api/admin/applications`
  - Authenticated (Requires `admin` role). Query parameters for filtering by `status` (`pending`, `approved`, `rejected`).
  - Response: Array of application summaries including candidate nickname, email, submission timestamp, and document links.
- `GET /api/admin/applications/:id/files/:file_type`
  - Authenticated (Requires `admin` role or matching applicant `user_id`). Serves the binary document stream (PDF/image) with proper Content-Type and download headers.
- `POST /api/admin/applications/:id/approve`
  - Authenticated (Requires `admin` role). Updates application status to `approved` and updates applicant's `role` to `specialist`.
- `POST /api/admin/applications/:id/reject`
  - Authenticated (Requires `admin` role). Payload: `{"reason": "string"}` (mandatory). Updates application status to `rejected`.

## Testing Decisions

- **Good Test Criteria**: Tests must verify the system exclusively through external HTTP boundaries (the Highest Seam), validating status codes, response payloads, role elevation in issued JWTs, and file download protections without mocking database operations or coupling to internal function signatures.
- **Tested Modules**: Auth Controller, Specialist Application Controller, Admin Verification Middleware, File Storage Provider, and Database Migration Routines.
- **Prior Art**: Follows the established integration testing pattern in `scripts/test_integration.py`, which performs real HTTP calls against `/api/register` and `/api/login` and asserts token contents.

## Out of Scope

- Automated optical character recognition (OCR) or AI-based certificate parsing.
- Third-party OAuth integration (e.g., verifying LinkedIn or Credly badges via OAuth APIs).
- Interactive video interviewing or scheduling modules.

## Further Notes

- Existing database records must be protected during migration; email prefixes will populate default nicknames when migrating legacy rows.
- Triage label: `ready-for-agent`
