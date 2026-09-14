# 01: User Unique Nickname & Role Enum Migration

**What to build:** Allow new users to register with a required unique Nickname alongside their email and password. Migrate the database role enum from `('learner', 'expert', 'admin')` to `('learner', 'specialist', 'admin')` while gracefully backfilling default nicknames derived from email prefixes for all pre-existing user records. Prevent duplicate nicknames during registration with immediate validation errors, and display the authenticated user's nickname across the application's navigation header and user profile views.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] New user registration enforces a required, non-empty `nickname` parameter alongside email and password.
- [ ] Attempting to register with a nickname that already exists returns an explicit validation error without creating a duplicate record.
- [ ] Database `role` type supports `'learner'`, `'specialist'`, and `'admin'` with `'learner'` as the default value.
- [ ] Existing users in the database without a nickname are automatically assigned a default nickname from their email prefix during migration.
- [ ] Authentication responses and JWT token claims include the user's assigned role and nickname.
- [ ] The web interface navigation header prominently displays the user's nickname.
- [ ] The user profile page displays the user's current nickname and allows editing it with uniqueness checks.
- [ ] Automated integration tests verify that duplicate nicknames are rejected and successful registrations issue valid authentication tokens containing role and nickname.
