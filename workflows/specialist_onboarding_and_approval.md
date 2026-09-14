# Workflow: Specialist Onboarding & Admin Approval

## Objective
Establish an end-to-end verified onboarding pathway for subject-matter experts (Specialists) to join the SOC Trainer platform, submitting verifiable credentials (resume, professional certificates, bio) that are reviewed and approved by Platform Administrators before being granted specialist permissions.

---

## 1. Loop Profile

- **Loop Name**: Specialist Onboarding Loop
- **Frequency / Cadence**: Event-driven (upon registration or learner role-upgrade request)
- **Primary Actors**:
  - **Applicant**: Learner or new registrant applying for Specialist role
  - **Admin**: Reviewer with approval/rejection authority

---

## 2. Trigger

- **Event**: 
  - *Case A (New User)*: User navigates to `/register`, selects the "Apply as Specialist" tab, fills in required registration credentials plus qualification documents, and submits.
  - *Case B (Existing Learner)*: Authenticated Learner navigates to `/profile`, clicks "Apply for Specialist", and submits qualification documents.

---

## 3. Data Schema & Requirements

### User Table Adjustments
- `users.nickname`: `VARCHAR UNIQUE NOT NULL` (backfilled from email prefix for existing users).
- `users.role`: Updated PostgreSQL enum `('learner', 'specialist', 'admin')`.

### Specialist Application Table (`specialist_applications`)
- `application_id`: `UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `user_id`: `UUID REFERENCES users(user_id) ON DELETE CASCADE`
- `status`: `VARCHAR NOT NULL DEFAULT 'pending'` (`pending`, `approved`, `rejected`)
- `bio`: `TEXT NOT NULL`
- `resume_path`: `VARCHAR NOT NULL` (relative secure path)
- `certificate_path`: `VARCHAR` (relative secure path, optional if credential link provided)
- `linkedin_url`: `VARCHAR`
- `portfolio_url`: `VARCHAR`
- `rejection_reason`: `TEXT`
- `reviewed_by`: `UUID REFERENCES users(user_id)`
- `reviewed_at`: `TIMESTAMP`
- `created_at`: `TIMESTAMP DEFAULT now()`
- `updated_at`: `TIMESTAMP DEFAULT now()`

---

## 4. Pipeline Execution (Push Right)

Before requesting Admin intervention, the system performs all automated verifications:

1. **Validation**:
   - Verify `nickname` uniqueness across `users`.
   - Validate file types (allowed: `.pdf`, `.png`, `.jpg`, `.jpeg`).
   - Enforce file size limit (maximum 10MB per file).
2. **Account Provisioning**:
   - If new user, create record in `users` with `role = 'learner'` and default tier.
   - Hash password with bcrypt.
3. **Artifact Persistence**:
   - Store documents in local volume: `/uploads/applications/{application_id}/resume_{uuid}.ext` and `certificate_{uuid}.ext`.
   - Restrict direct web access (no public URL).
4. **State Initialization**:
   - Record created in `specialist_applications` with status `pending`.
   - Notify Admin panel of pending count.

---

## 5. Checkpoint (Admin Review)

- **Actor**: Platform Admin (`role = 'admin'`).
- **Location**: Admin Portal at `/admin/specialists`.
- **The Brief**:
  - **Candidate Header**: Nickname, Email, Submission Date.
  - **Experience Overview**: Bio text, clickable LinkedIn / Portfolio external links.
  - **Document Inspection**: Secure preview/download links powered by `GET /api/admin/applications/:id/files/:type` (JWT protected).
  - **Decision Controls**:
    - **[Approve Button]**: Prompts confirmation modal.
    - **[Reject Button]**: Opens rejection modal with a **required** `rejection_reason` text field.

---

## 6. Outcomes & Downstream Effects

### Path A: Approval
1. Database state updates:
   - `specialist_applications.status = 'approved'`
   - `specialist_applications.reviewed_by = admin_id`
   - `specialist_applications.reviewed_at = NOW()`
   - `users.role = 'specialist'`
2. Notification:
   - In-app notification dispatched to user: *"Your application to become a Specialist has been approved! You now have access to the Review Dashboard and Analysis Thread authoring."*
3. Permissions:
   - Next JWT issued or active session refresh grants `role: "specialist"`.

### Path B: Rejection
1. Database state updates:
   - `specialist_applications.status = 'rejected'`
   - `specialist_applications.rejection_reason = input_reason`
   - `specialist_applications.reviewed_by = admin_id`
   - `specialist_applications.reviewed_at = NOW()`
   - `users.role` remains `'learner'`.
2. Notification & Visibility:
   - In-app notification dispatched: *"Your Specialist application was reviewed and not approved: [rejection_reason]. You may revise and re-apply."*
   - Profile page displays application status card with rejection feedback.
3. Re-application:
   - User can click "Update Application & Re-apply", upload revised documents, and trigger a new review cycle without creating a new user account.
