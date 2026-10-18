# 02: Static Scenario Catalog Upgrade & Legacy Database Auto-Remediation

**What to build:** Upgrade all 15 pre-seeded static Scenarios across CompTIA Security+ Domains 1–5 (both True Positive and False Positive Scenarios) so every phase (`containment`, `eradication`, `recovery`) contains 3 to 4 contextually balanced Response Actions (1–2 correct actions and at least 2 plausible distractors with matching length and technical specificity), and implement an automatic database remediation mechanism that detects and upgrades any existing Scenarios in PostgreSQL that still contain legacy 2-option structures or banned boilerplate distractors.

**Blocked by:** 01: Contextual Scenario Playbook Synthesis & Anti-Predictability Validation Gate (#2)

**Status:** completed

- [x] All 15 static Scenarios in the scenario generator script and pre-seeded JSON catalog contain 3 to 4 Response Actions in every phase (`containment`, `eradication`, `recovery`), with at least 1 correct option (`+5` points) and at least 2 plausible, scenario-specific distractors (`-2` to `-3` points) matching the length and operational tone of the correct options.
- [x] Every static Scenario in the pre-seeded catalog passes the Scenario anti-predictability validation gate with zero banned boilerplate strings.
- [x] An idempotent database remediation routine runs on backend startup (and guards Scenario loading during Simulation Session creation/recovery) to inspect existing `scenarios` rows and automatically upgrade any row failing the validation gate by syncing from the updated static catalog or applying the contextual distractor upgrader.
- [x] Automated tests verify that every Scenario in the static seed catalog passes validation and that legacy 2-option boilerplate Scenarios are upgraded in-place into valid 3+ option Scenarios.
