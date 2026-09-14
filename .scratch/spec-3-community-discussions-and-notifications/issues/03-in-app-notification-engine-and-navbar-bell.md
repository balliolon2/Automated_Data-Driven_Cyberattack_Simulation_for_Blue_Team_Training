# 03: In-App Notification Engine & Navbar Alert Bell

**What to build:** Build an asynchronous notification engine and Navbar notification UI. Automatically dispatch notifications when a user comments on a specialist's thread, when someone replies to an existing comment, or when an administrator approves or rejects a specialist application. Render a notification bell icon with an unread count badge in the global navigation bar. Clicking the bell displays a dropdown preview of recent notifications with timestamps and direct deep links (`/discussions/:threadId#comment-:commentId`). Provide actions to mark notifications as read individually or all at once.

**Blocked by:** 02: Helpful Upvotes, Comment Pinning & Discussion Moderation

**Status:** ready-for-agent

- [ ] Submitting a top-level comment automatically creates an in-app notification for the thread author (unless commenter is the author).
- [ ] Submitting a nested reply automatically creates an in-app notification for the parent comment author (unless replyer is the parent author).
- [ ] Administrator approval or rejection of a specialist application generates an immediate notification to the candidate's account.
- [ ] The global Navbar component includes a notification bell icon featuring a dynamic badge indicator showing the count of unread notifications.
- [ ] Clicking the notification bell displays a dropdown menu listing recent notifications with timestamps and unread markers.
- [ ] Clicking a discussion notification directs the user to the thread and jumps directly to the target comment anchor.
- [ ] Users can mark a single notification as read or click "Mark all as read" to reset the unread counter.
- [ ] Automated integration tests verify that comment replies trigger notification records and that read-state transitions update via API.
