# 02: Specialist Application Submission & Secure Document Storage

**What to build:** Allow new registrants and existing learners to submit an application to become a Specialist by uploading their Resume (mandatory) and Certificate (optional) files, alongside a professional cybersecurity bio and external links (LinkedIn, Portfolio, Credly). Persist uploaded documents to a protected local volume with randomized UUID filenames to prevent unauthorized enumeration or path traversal. Persist an application record with a `pending` status, keeping the applicant's account active in the `learner` role while displaying a clear "Pending Specialist Approval" status card on their profile.

**Blocked by:** 01: User Unique Nickname & Role Enum Migration

**Status:** ready-for-agent

- [ ] The registration interface includes an "Apply as Specialist" tab or mode accepting bio, resume file, optional certificate file, and external profile URLs.
- [ ] Authenticated learners can access an "Apply for Specialist" form from their profile page to submit application credentials.
- [ ] File uploads validate MIME types (PDF, PNG, JPG, JPEG) and enforce a maximum file size limit of 10MB per file, returning a descriptive error on failure.
- [ ] Validated documents are stored in a secure local storage directory using randomized UUID filenames, completely inaccessible via direct public URLs.
- [ ] A record is created in the specialist applications table with `status = 'pending'`, associated with the applicant's account ID and submission timestamp.
- [ ] The applicant's role remains `'learner'` while the application is pending, allowing full, uninterrupted access to learner training workflows.
- [ ] The applicant's profile page displays an "Application Under Review" card showing submission date, bio summary, and status.
- [ ] Automated integration tests verify that submitting an application creates a pending record and saves the documents to secure storage.
