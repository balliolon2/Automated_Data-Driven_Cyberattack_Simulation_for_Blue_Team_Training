# SOC Trainer

A web-based platform that trains SOC Tier 1 analysts through adaptive pre-test → scenario simulation → post-test workflows, measuring skill improvement across CompTIA Security+ domains.

## Language

**Scenario**:
A self-contained simulated security incident that a learner investigates and responds to. Contains log data, key findings, and response actions.
_Avoid_: Exercise, challenge, lab

**Scenario Loop**:
The repeating cycle of scenario assignment → investigation → scoring → next scenario, driven by the learner's weakest domain until all domains reach the proficiency threshold.
_Avoid_: Training cycle, simulation round

**Proficiency Score**:
A per-domain percentage (0–100) stored in `user_skill_profiles`, representing a learner's current competence. Updated via weighted average after each scenario.
_Avoid_: Skill score, competency rating

**Proficiency Threshold**:
The minimum proficiency score a learner must reach in every domain before the scenario loop ends. A single value applied uniformly across all five domains.
_Avoid_: Pass mark, cutoff

**Key Finding**:
A specific piece of evidence within a scenario's logs that a learner is expected to discover through querying. Each key finding is tagged with a domain and carries points.
_Avoid_: Clue, indicator, evidence item

**TP/FP Decision**:
The learner's triage classification of whether the scenario's alert represents a True Positive (real threat) or False Positive (benign). Scored against the scenario's `is_true_positive` field.
_Avoid_: Alert classification, triage choice

**Response Action**:
A checkbox option within the Response panel, grouped under Containment, Eradication, or Recovery phases. Each action is marked correct or incorrect and tagged with a domain.
_Avoid_: Playbook step, remediation option

**Simulation Session**:
A single attempt at a single scenario by a learner. Maps 1:1 with a row in `simulation_sessions`.
_Avoid_: Simulation run, training session

**Domain Breakdown**:
The per-domain score analysis shown after a pre-test or scenario completion, displaying each domain's percentage and whether it meets the proficiency threshold.
_Avoid_: Score summary, domain analysis

**Scenario Generation Source**:
The origin classification of a scenario synthesized for a simulation session (`ai_generated` via real-time LLM synthesis or `static_fallback` from pre-seeded pool).
_Avoid_: Scenario type, origin mode

**Fallback Reason**:
The diagnostic explanation stored when an LLM API call fails or is unconfigured, detailing why a static scenario was selected instead.
_Avoid_: Error log, fallback status

**Evaluation Result**:
The canonical structured performance summary generated upon simulation session submission. Contains the overall score, TP/FP outcome, key findings found, response actions executed, and updated per-domain proficiency metrics.
_Avoid_: Score sheet, grading output

