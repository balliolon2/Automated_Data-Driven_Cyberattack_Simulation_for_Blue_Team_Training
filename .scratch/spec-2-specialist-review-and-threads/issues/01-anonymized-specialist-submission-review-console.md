# 01: Anonymized Specialist Submission Review Console

**What to build:** Provide a specialized Review Console at `/specialist/reviews` for users with the `specialist` or `admin` role to browse, filter, and inspect completed simulation submissions across all learners. Submissions can be filtered by scenario, security domain, or score threshold (e.g., struggling sessions scoring < 70%). Strictly protect learner privacy by displaying only user Nicknames while completely concealing email addresses and personal identifiers. Allow specialists to drill down into an individual attempt to examine the learner's TP/FP triage choice against ground truth, discovered versus missed key findings, executed response actions, and log query history.

**Blocked by:** Spec 1 Ticket 03: Admin Portal for Reviewing, Document Preview & Specialist Approval

**Status:** ready-for-agent

- [ ] Review console routes and APIs restrict access strictly to `specialist` and `admin` roles, returning 403 Forbidden to standard learners.
- [ ] The review console displays a paginated list of completed simulation sessions with learner nickname, scenario title, completion timestamp, score, and TP/FP outcome.
- [ ] Learner email addresses, user IDs, and personal identifiable information are strictly excluded from all review console API responses and UI views.
- [ ] Console provides functional filters by Scenario, Security Domain (Domains 1–5), and a score threshold filter (e.g., scores < 70%).
- [ ] Clicking an attempt opens a detailed inspection modal or view showing the learner's TP/FP triage decision compared with the scenario's actual state.
- [ ] The inspection view displays a checklist of Key Findings indicating which items the learner successfully identified and which were missed.
- [ ] The inspection view presents the response actions selected by the learner grouped under Containment, Eradication, and Recovery phases.
- [ ] Automated integration tests verify that learners receive 403 Forbidden on review endpoints, while specialists can query submissions and retrieve anonymized investigation data.
