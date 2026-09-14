---
title: "Spec 3: Community Discussions (2-Level Comments), Upvotes & In-App Notifications"
labels: ["ready-for-agent"]
status: "ready-for-agent"
---

# Spec 3: Community Discussions (2-Level Comments), Upvotes & In-App Notifications

## Problem Statement

After an Analysis Thread is published by a Specialist, there is no interactive mechanism for learners to ask follow-up questions, seek clarification on difficult investigation techniques, or discuss alternate remediation strategies. The platform lacks a bidirectional feedback loop: specialists cannot determine whether their case breakdowns resolved student confusion, learners cannot highlight exceptionally helpful explanations, and users receive no alerting when someone responds to their inquiry or when an administrator acts upon their role application.

## Solution

1. Introduce an interactive, 2-level threaded commenting system beneath Analysis Threads (top-level questions and direct nested replies), enabling focused technical dialogue between Learners, Specialists, and Admins.
2. Establish visual role badges (`[Specialist]`, `[Admin]`, `[Learner]`) alongside unique user nicknames to reinforce credibility, coupled with a transactional 1-vote-per-user Helpful/Upvote system for threads and comments.
3. Provide moderation and curation capabilities, including comment pinning by thread authors and admins to highlight essential conclusions, self-editing/deletion for authors, and content moderation tools for admins.
4. Implement an in-app notification system with a Navbar alert indicator that informs users of comment replies, thread mentions, and Specialist application status decisions.

## User Stories

1. As a learner reading an analysis thread, I want to post a top-level comment with my question, so that the author or other analysts can help me understand a specific triage concept.
2. As a specialist or learner, I want to reply directly to another user's comment, so that our conversation is organized as a distinct 2-level sub-thread.
3. As a user, I want any reply attempted on a level-2 comment to automatically attach to the existing thread branch, so that the discussion never exceeds 2 levels of indentation.
4. As a reader, I want to see distinct visual badges next to author nicknames (`[Specialist]`, `[Admin]`, `[Learner]`), so that I can immediately identify expert insights and official announcements.
5. As a user, I want to toggle a "Helpful" upvote on any analysis thread or comment, so that high-quality technical advice is recognized by the community.
6. As a user, I want the system to ensure I can only cast one upvote per thread or comment, so that ratings cannot be artificially inflated.
7. As a specialist authoring an analysis thread, I want to pin up to 3 high-value comments to the top of the discussion, so that important takeaways or community contributions are immediately visible.
8. As a comment author, I want to edit or delete my own comments, so that I can refine my wording or remove obsolete questions.
9. As a specialist thread author, I want to hide or remove spam or inappropriate comments on my own thread, so that technical quality is maintained.
10. As a platform administrator, I want to edit, hide, or delete any comment or thread across the platform, so that community standards are strictly enforced.
11. As a thread author, I want to receive an in-app notification when someone comments on my thread, so that I can provide prompt follow-up answers.
12. As a commenter, I want to receive an in-app notification when someone replies to my comment, so that I can continue the learning dialogue.
13. As a specialist applicant, I want to receive an in-app notification when an administrator approves or rejects my application, so that I am instantly informed of my new role.
14. As a user, I want to see an unread notification counter badge on the Navbar bell icon, so that I am aware of new activity without leaving my current page.
15. As a user, I want clicking a notification to navigate directly to the relevant thread and jump down to the specific comment anchor, so that I do not have to search for the reply.
16. As a user, I want to mark notifications as read individually or all at once, so that I can manage my alert inbox.

## Implementation Decisions

### Interaction Hierarchy & Thread Flattening
- Discussion comments follow a strict 2-level hierarchy:
  - **Level 1 (Root Comment)**: `parent_comment_id IS NULL`. Represents a primary inquiry or standalone observation.
  - **Level 2 (Nested Reply)**: `parent_comment_id` points to a Level 1 comment.
  - If a user submits a reply to an existing Level 2 comment, the backend normalizes the `parent_comment_id` to point to the Level 1 root ancestor, preventing UI nesting degradation.

### Schema Specifications

- **Thread Comments Table (`thread_comments`)**:
  - `comment_id`: UUID PK, default `gen_random_uuid()`
  - `thread_id`: UUID FK referencing `analysis_threads(thread_id)` with `ON DELETE CASCADE`
  - `user_id`: UUID FK referencing `users(user_id)` with `ON DELETE CASCADE`
  - `parent_comment_id`: UUID FK referencing `thread_comments(comment_id)` with `ON DELETE CASCADE` (nullable)
  - `content`: TEXT NOT NULL (1–2000 characters)
  - `upvote_count`: INT DEFAULT 0
  - `is_pinned`: BOOLEAN DEFAULT false
  - `is_hidden`: BOOLEAN DEFAULT false
  - `created_at`: TIMESTAMP default `now()`
  - `updated_at`: TIMESTAMP default `now()`

- **Thread Upvotes Table (`thread_upvotes`)**:
  - `thread_id`: UUID FK referencing `analysis_threads(thread_id)` with `ON DELETE CASCADE`
  - `user_id`: UUID FK referencing `users(user_id)` with `ON DELETE CASCADE`
  - `created_at`: TIMESTAMP default `now()`
  - `PRIMARY KEY (thread_id, user_id)`

- **Comment Upvotes Table (`comment_upvotes`)**:
  - `comment_id`: UUID FK referencing `thread_comments(comment_id)` with `ON DELETE CASCADE`
  - `user_id`: UUID FK referencing `users(user_id)` with `ON DELETE CASCADE`
  - `created_at`: TIMESTAMP default `now()`
  - `PRIMARY KEY (comment_id, user_id)`

- **User Notifications Table (`user_notifications`)**:
  - `notification_id`: UUID PK, default `gen_random_uuid()`
  - `user_id`: UUID FK referencing `users(user_id)` with `ON DELETE CASCADE`
  - `title`: VARCHAR(120) NOT NULL
  - `message`: TEXT NOT NULL
  - `link_url`: VARCHAR(255)
  - `is_read`: BOOLEAN DEFAULT false
  - `created_at`: TIMESTAMP default `now()`

### API Contracts

- `POST /api/threads/:id/comments`
  - Authenticated (Any active user).
  - Payload: `{"content": "string", "parent_comment_id": "uuid" (optional)}`.
  - Behavior: Creates comment, creates notification for parent author or thread author, returns created comment.
- `GET /api/threads/:id/comments`
  - Authenticated (Any user).
  - Response: Structured list of Level 1 comments, each containing an array of its Level 2 replies, author metadata (`nickname`, `role`), `is_pinned`, and `has_upvoted` boolean for caller.
- `PUT /api/comments/:id`
  - Authenticated (Author or `admin`).
  - Payload: `{"content": "string"}`.
  - Response: 200 OK with updated comment.
- `DELETE /api/comments/:id`
  - Authenticated (Author, Thread Author Specialist, or `admin`).
  - Response: 200 OK with confirmation.
- `POST /api/comments/:id/pin`
  - Authenticated (Thread Author Specialist or `admin`).
  - Response: 200 OK with toggled `is_pinned` status.
- `POST /api/threads/:id/upvote`
  - Authenticated (Any active user).
  - Response: 200 OK with updated thread `upvote_count` and current `has_upvoted` boolean.
- `POST /api/comments/:id/upvote`
  - Authenticated (Any active user).
  - Response: 200 OK with updated comment `upvote_count` and current `has_upvoted` boolean.
- `GET /api/notifications`
  - Authenticated (Any user).
  - Query parameters: `page`, `page_size`, `unread_only` (boolean).
  - Response: Paginated notifications array and `unread_count` integer.
- `PUT /api/notifications/:id/read`
  - Authenticated (Owner of notification).
  - Response: 200 OK.
- `PUT /api/notifications/read-all`
  - Authenticated (Any user).
  - Response: 200 OK marking all unread notifications for caller as read.

## Testing Decisions

- **Good Test Criteria**: Tests must verify behavior through external HTTP boundaries (Highest Seam), ensuring correct parent-child relationship construction in response JSON, transactional upvote counts without duplicates, notification generation upon comment creation, and authorization constraints on deletion and pinning.
- **Tested Modules**: Comment Controller, Upvote Service, Notification Controller, Moderation Service.
- **Prior Art**: Follows HTTP session verification tests in `scripts/test_integration.py`.

## Out of Scope

- Persistent WebSocket / real-time streaming connections (standard HTTP polling/fetching upon page focus is sufficient).
- File/image uploads directly within comment bodies (plain markdown and code formatting only).
- Email newsletter or external SMS delivery of notifications.

## Further Notes

- In-app notification bell polls `GET /api/notifications?unread_only=true` or re-fetches on route navigation.
- Triage label: `ready-for-agent`
