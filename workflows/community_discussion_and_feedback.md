# Workflow: Community Discussion & Feedback

## Objective
Provide an interactive, 2-level discussion and mentoring environment within Analysis Threads, enabling Learners to ask clarifying questions about scenario decisions, allowing Specialists to provide guidance, and facilitating platform-wide knowledge sharing with in-app notifications.

---

## 1. Loop Profile

- **Loop Name**: Discussion & Feedback Loop
- **Frequency / Cadence**: Event-driven (user submits comment or reply)
- **Primary Actors**:
  - **Learner**: Asks questions, shares investigation difficulties, receives guidance
  - **Specialist**: Provides expert clarification, pins high-value takeaways
  - **Admin**: Moderates content and ensures platform standards

---

## 2. Trigger

- **Event**:
  - *Trigger A (New Inquiry)*: A Learner submits a top-level comment on an Analysis Thread.
  - *Trigger B (Direct Guidance)*: A Specialist or peer submits a nested reply to an existing comment.
  - *Trigger C (Helpful Vote)*: A user upvotes a thread or comment.

---

## 3. Data Schema & Requirements

### Thread Comments Table (`thread_comments`)
- `comment_id`: `UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `thread_id`: `UUID REFERENCES analysis_threads(thread_id) ON DELETE CASCADE`
- `user_id`: `UUID REFERENCES users(user_id) ON DELETE CASCADE`
- `parent_comment_id`: `UUID REFERENCES thread_comments(comment_id) ON DELETE CASCADE` (NULL for top-level, non-null for level-2 replies)
- `content`: `TEXT NOT NULL`
- `upvote_count`: `INT DEFAULT 0`
- `is_pinned`: `BOOLEAN DEFAULT false`
- `is_hidden`: `BOOLEAN DEFAULT false`
- `created_at`: `TIMESTAMP DEFAULT now()`
- `updated_at`: `TIMESTAMP DEFAULT now()`

### Comment Upvotes Table (`comment_upvotes`)
- `comment_id`: `UUID REFERENCES thread_comments(comment_id) ON DELETE CASCADE`
- `user_id`: `UUID REFERENCES users(user_id) ON DELETE CASCADE`
- `created_at`: `TIMESTAMP DEFAULT now()`
- `PRIMARY KEY (comment_id, user_id)`

### User Notifications Table (`user_notifications`)
- `notification_id`: `UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `user_id`: `UUID REFERENCES users(user_id) ON DELETE CASCADE`
- `title`: `VARCHAR(120) NOT NULL`
- `message`: `TEXT NOT NULL`
- `link_url`: `VARCHAR(255)`
- `is_read`: `BOOLEAN DEFAULT false`
- `created_at`: `TIMESTAMP DEFAULT now()`

---

## 4. Pipeline Execution (Push Right)

1. **Submission & Sanitization**:
   - Limit comment length (1 to 2000 characters).
   - Sanitize plain text / markdown to prevent XSS.
   - Enforce 2-level nesting hierarchy: if `parent_comment_id` points to a comment that itself has a parent, normalize to point to the top-level root comment.
2. **Author Identity Enrichment**:
   - Fetch author's `nickname` and `role`.
   - Render designated role badges:
     - `[Specialist]`: High-contrast indigo badge with verification shield icon.
     - `[Admin]`: Rose/amber badge with terminal shield icon.
     - `[Learner]`: Subtle graphite badge.
3. **Notification Dispatch**:
   - If top-level comment: Notify thread author (`analysis_threads.author_id`) unless commenter is author.
   - If level-2 reply: Notify parent comment author (`parent_comment.user_id`) unless replyer is parent author.
   - Insert notification into `user_notifications`.
   - Real-time / polling badge counter updates Navbar notification bell.

---

## 5. Checkpoint (Moderation & Pinning)

- **Self-Management**:
  - Comment authors can edit or delete their own comments.
- **Specialist Authority (Thread Author)**:
  - Can **[Pin]** up to 3 comments to the top of the discussion list. Pinned comments receive an amber accent and "Pinned by Author" marker.
  - Can **[Hide]** spam or irrelevant comments under their thread.
- **Admin Authority**:
  - Global moderation permissions: edit, delete, pin, or hide any comment across all threads.

---

## 6. Outcomes & Downstream Effects

1. **Continuous Mentorship**:
   - The thread functions as an asynchronous coaching session where recurring mistakes made in simulation sessions are addressed directly by Specialists.
2. **Reputation & Engagement**:
   - Upvotes update `upvote_count` transactionally.
   - High-upvote comments and threads gain priority sorting in the `/discussions` feed.
3. **Loop Closure**:
   - Learners who asked questions receive notifications with direct jump-links (`/discussions/:threadId#comment-:commentId`) to continue the learning conversation.
