# 01: Two-Level Threaded Commenting System & Visual Role Badges

**What to build:** Implement the interactive commenting interface underneath Analysis Threads, allowing learners, specialists, and administrators to engage in focused technical dialogue. Support a strict 2-level comment hierarchy (top-level questions and direct nested replies; any reply targeting a level-2 comment is automatically normalized to attach to the root level-1 parent). Display distinct visual role badges next to author nicknames (`[Specialist]`, `[Admin]`, `[Learner]`) to immediately establish credibility. Support basic markdown formatting in comments and enable authors to edit or delete their own posts.

**Blocked by:** Spec 2 Ticket 03: Discussions Feed & Scenario Result Page Integration

**Status:** ready-for-agent

- [ ] The analysis thread view at `/discussions/:id` renders an interactive comment section with an input area for new top-level comments.
- [ ] Authenticated users (any active role) can submit a top-level comment containing 1 to 2000 characters.
- [ ] Users can click "Reply" beneath any existing comment to submit a nested response.
- [ ] If a user replies to an existing level-2 reply, the backend normalizes `parent_comment_id` to the root level-1 comment, maintaining a clean 2-level discussion layout.
- [ ] Author headers display the author's Nickname and a distinctive Role Badge (`[Specialist]`, `[Admin]`, `[Learner]`).
- [ ] Comment bodies render basic markdown elements including bold, italic, bullet lists, and monospace code blocks.
- [ ] Comment authors can edit or delete their own comments with immediate UI updates.
- [ ] Automated integration tests verify that posting top-level comments and nested replies preserves the 2-level hierarchy and includes author badges.
