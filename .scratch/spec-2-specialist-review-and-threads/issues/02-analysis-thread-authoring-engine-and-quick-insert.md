# 02: Analysis Thread Authoring Engine with Markdown & Scenario Quick-Insert

**What to build:** Implement the authoring pipeline for Specialists and Admins to draft, publish, and maintain authoritative Analysis Threads (กระทู้). Provide a split-pane Markdown editor with live preview, log/KQL query syntax highlighting, and a "Quick-Insert" toolbar that pulls in scenario blueprint details (alert telemetry narrative, key findings checklist, expected response actions). Persist threads in the database with optional scenario association and tags, allowing authors and administrators to edit or delete their posts.

**Blocked by:** 01: Anonymized Specialist Submission Review Console

**Status:** ready-for-agent

- [ ] Creation endpoints and editor routes restrict authoring strictly to users with the `specialist` or `admin` role, rejecting learners with 403 Forbidden.
- [ ] Thread creation interface includes inputs for title, optional scenario association dropdown, domain/topic tags, and a markdown body editor.
- [ ] Editor provides a side-by-side or tabbed live Markdown preview supporting headings, lists, tables, callout blocks, and code formatting.
- [ ] Code snippets within the markdown preview support syntax highlighting for KQL queries, SIEM syslogs, and JSON event structures.
- [ ] A Quick-Insert template toolbar allows the specialist to inject the linked scenario's alert description and key findings table into the draft with one click.
- [ ] Publishing a thread persists the record in the database with author reference, scenario reference, tag array, and creation timestamp.
- [ ] Authors can edit and delete their own threads, while Platform Admins can edit or remove any thread across the platform.
- [ ] Automated integration tests verify that learners cannot post threads, and that specialists can create, retrieve, update, and delete scenario-linked threads.
