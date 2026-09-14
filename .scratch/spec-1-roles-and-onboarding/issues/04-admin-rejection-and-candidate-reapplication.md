# 04: Admin Rejection with Reason & Candidate Re-application Flow

**What to build:** Enable Platform Admins to reject a candidate's specialist application from the review dashboard by providing a mandatory explanation note (`rejection_reason`). Transition the application status to `rejected`, record reviewer metadata, and dispatch an in-app notification to the candidate with the feedback. Present an "Application Declined" card on the candidate's profile page displaying the administrator's feedback along with an "Edit & Re-apply" action. Clicking this action allows the candidate to revise their bio, update external links, upload new documents, and re-submit, resetting the application status back to `pending` and returning it to the Admin review queue without requiring a new account.

**Blocked by:** 03: Admin Portal for Reviewing, Document Preview & Specialist Approval

**Status:** ready-for-agent

- [ ] Rejecting an application in the admin portal requires entering a non-empty explanation string (`rejection_reason`).
- [ ] Submitting a rejection updates the application status to `'rejected'`, persists the `rejection_reason`, and records `reviewed_by` and `reviewed_at`.
- [ ] Rejection creates an in-app notification to the candidate: *"Your Specialist application was not approved: [reason]. You may revise and re-apply."*
- [ ] The user's account remains active in the `'learner'` role throughout and after rejection.
- [ ] The candidate's profile page renders an "Application Declined" card displaying the admin's exact feedback.
- [ ] The profile page provides an "Edit & Re-apply" button opening the application form pre-filled with previous information.
- [ ] Re-submitting the application updates documents and transitions the status back to `'pending'`.
- [ ] The re-submitted application immediately re-appears in the Admin review queue at `/admin/specialists`.
- [ ] Automated integration tests verify rejection without reason is rejected, valid rejection saves the reason and sends notification, and re-application resets status to pending.
