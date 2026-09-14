# 03: Discussions Feed & Scenario Result Page Integration

**What to build:** Build the public and learner-facing discussion discovery surfaces. Create a central `/discussions` feed page where learners can browse, search, and filter analysis threads by domain, scenario, or tags, viewing author Nicknames and `[Specialist]` badges. Build the dedicated thread view at `/discussions/:id` rendering the full analysis writeup. Directly integrate with the core learner workflow: on `ScenarioResultPage` (after a learner finishes a simulation attempt), render a "Specialist Insights & Community Discussions" section below their domain breakdown, automatically surfacing any analysis threads linked to that `scenario_id` so learners can immediately read expert breakdowns and compare strategies.

**Blocked by:** 02: Analysis Thread Authoring Engine with Markdown & Scenario Quick-Insert

**Status:** ready-for-agent

- [ ] A central `/discussions` catalog page renders published analysis threads sorted by latest or most viewed.
- [ ] The feed allows filtering by CompTIA Security+ Domain, specific Scenario, or custom topic tags.
- [ ] Thread preview cards display the author's unique Nickname, a distinct `[Specialist]` badge, creation date, and linked scenario title.
- [ ] Navigating to `/discussions/:id` renders the full thread body with formatted Markdown, highlighted code blocks, and author identity.
- [ ] Accessing `/discussions/:id` atomically increments the thread's `view_count` in the database.
- [ ] Upon completing a simulation session on `ScenarioResultPage`, the application queries `GET /api/threads?scenario_id=:id`.
- [ ] If analysis threads exist for that scenario, a dedicated "Specialist Insights & Community Discussions" card renders below the evaluation metrics.
- [ ] Learners can click directly from the result page card to navigate into the analysis thread without leaving the training context.
- [ ] Automated integration tests verify that querying threads by `scenario_id` returns linked records with correct author nickname and specialist role attributes.
