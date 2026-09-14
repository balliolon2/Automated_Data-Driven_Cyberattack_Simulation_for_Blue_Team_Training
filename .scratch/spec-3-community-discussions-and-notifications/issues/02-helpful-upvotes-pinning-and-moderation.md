# 02: Helpful Upvotes, Comment Pinning & Discussion Moderation

**What to build:** Provide community engagement and content moderation capabilities for analysis threads and comments. Implement a transactional 1-vote-per-user "Helpful / Upvote" system allowing readers to upvote or un-upvote posts and comments. Allow thread authors (Specialists) and Platform Admins to Pin up to 3 high-value takeaway comments to the top of the discussion list, styling them with a distinct highlight and "Pinned by Author" badge. Grant Specialist thread authors the ability to hide spam comments on their threads, and grant Platform Admins global moderation authority to edit, hide, or delete any content across the application.

**Blocked by:** 01: Two-Level Threaded Commenting System & Visual Role Badges

**Status:** ready-for-agent

- [ ] Threads and comments render a "Helpful" upvote toggle displaying the current score and whether the active user has upvoted.
- [ ] Upvotes are enforced via composite primary key constraints in the database, preventing duplicate votes by the same user.
- [ ] Specialist thread authors and Platform Admins can toggle a "Pin" action on comments (capped at a maximum of 3 pinned comments per thread).
- [ ] Pinned comments are anchored at the very top of the comments section with an amber accent and "Pinned by Author" marker.
- [ ] Specialist thread authors can hide or delete inappropriate/spam comments posted under their own threads.
- [ ] Platform Admins have universal privileges to edit, hide, or permanently delete any thread or comment across the system.
- [ ] Automated integration tests verify that upvote toggling increments and decrements correctly without double-voting, and that non-authors/non-admins cannot pin comments.
