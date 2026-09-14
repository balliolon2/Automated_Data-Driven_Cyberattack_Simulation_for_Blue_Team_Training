# 03: Admin Portal for Reviewing, Document Preview & Specialist Approval

**What to build:** Build an administrative review dashboard at `/admin/specialists` restricted to users with the `admin` role. The portal presents a queue of pending applications displaying candidate Nickname, Email, submission timestamp, bio, and external links. Implement a secure, authenticated document streaming endpoint that strictly verifies the caller is an Admin before serving resume or certificate files. Allow the Admin to click "Approve", which atomically transitions the application status to `approved`, promotes the applicant's account role to `specialist`, records reviewer metadata, and dispatches an in-app notification to the newly approved Specialist.

**Blocked by:** 02: Specialist Application Submission & Secure Document Storage

**Status:** ready-for-agent

- [ ] Administrative review endpoints and views enforce strict authorization checks, returning 403 Forbidden to non-admin callers.
- [ ] An administrative portal at `/admin/specialists` renders a list of pending applications with applicant nickname, email, submission date, bio, and portfolio links.
- [ ] Resumes and certificates are streamed through an authenticated endpoint (`/api/admin/applications/:id/files/:type`) that confirms admin privileges or applicant ownership.
- [ ] Clicking "Approve" in the admin dashboard updates the application record to `status = 'approved'` and sets `reviewed_by` and `reviewed_at`.
- [ ] Approval immediately updates the target user's `role` to `'specialist'` in the database.
- [ ] Approval creates an in-app notification informing the user: *"Your application to become a Specialist has been approved!"*
- [ ] Subsequent logins or token refreshes for the approved applicant carry the updated `'specialist'` role claim.
- [ ] Automated integration tests confirm that non-admins cannot approve applications, while valid admin approvals successfully elevate user roles and trigger notifications.
