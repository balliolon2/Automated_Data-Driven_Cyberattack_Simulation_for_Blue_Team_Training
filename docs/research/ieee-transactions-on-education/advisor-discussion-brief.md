# Advisor Discussion Brief — IEEE Transactions on Education Paper Selection

**Search Run:** `ieee-toe-2026-09-10-01` · **Date:** 2026-09-10
**Status:** Top 2–3 and preliminary research gap **confirmed by user** on 2026-09-10
**Evidence level:** Abstract Verified (full text behind IEEE paywall; no institutional access available at time of run)
**Full data:** [literature-screening-matrix.md](literature-screening-matrix.md) · [records.json](records.json)

---

## 1. Recommended Papers

1. **Workman, M., Luévanos, J. A., & Mai, B.** "A Study of Cybersecurity Education Using a Present-Test-Practice-Assess Model." *IEEE Transactions on Education*, 65(1), Feb 2022, pp. 40–45. DOI: [10.1109/TE.2021.3086025](https://ieeexplore.ieee.org/document/9452797)
2. **"Immersive Virtual Environment for Automotive Security Education."** *IEEE Transactions on Education*, 2026, pp. 86–101. DOI: [10.1109/TE.2026.3654077](https://ieeexplore.ieee.org/document/11372737)
3. **"Blockchain Integration for Practical Cybersecurity Education With Scalable Cyber Infrastructure."** *IEEE Transactions on Education*, 2025, pp. 543–552. DOI: [10.1109/TE.2025.3626489](https://ieeexplore.ieee.org/document/11229940)

## 2. Selection Rationale

Screened from 30 unique IEEE ToE records (6 query groups + 14 expansion queries, 2021–2026 window) and scored on weighted criteria (Evidence Quality 30%, Learning Outcome 25%, Practical Cybersecurity Training 20%, Project Relevance 15%, Adaptivity 10%):

| Paper | Score /100 | Why selected |
|---|---:|---|
| PTPA Model (2022) | 91.0 | Randomized multi-modality study; **pretest as covariate, post-test as DV**; measures *applied behavior* — the methodological blueprint closest to our system |
| Automotive IVE (2026) | 71.0 | Recent (2026); user studies with statistically significant knowledge/engagement/self-efficacy gains; immersive-platform contrast to our web-based approach |
| Blockchain/FABRIC (2025) | 67.0 | Hands-on cyber-infrastructure education with evidence-based learning design; its survey-only evaluation illustrates the self-report limitation ToE warns about |

Not selected: AI Paradigm (65.0, evidence ongoing), Children SLR (59.0, domain mismatch but citable for the gap), PESTLE Europe (50.0, policy analysis).

## 3. Paper-by-Paper Comparison

| Dimension | PTPA (2022) | Automotive IVE (2026) | Blockchain/FABRIC (2025) |
|---|---|---|---|
| Intervention | Simulation + live competitive activities vs. traditional instruction | VR immersive environment (sensor attacks) | Hands-on labs on national cyber infrastructure |
| Research design | Randomized 4-section; pretest covariate / post-test DV | User studies (pre/post, significance reported) | Student surveys (self-report) |
| Outcome type | Applied behavioral performance | Knowledge, engagement, self-efficacy | Perceived practical skills, understanding |
| Adaptivity | None (fixed modalities) | None | None |
| SOC relevance | Closest — organizational cybersecurity behavior | Domain-specific (automotive) | General cybersecurity practice |
| Evidence status | Abstract Verified | Abstract Verified | Abstract Verified |

## 4. Research Gap (confirmed preliminary)

1. **No SOC-analyst decision training in ToE:** all SOC / security-operations query groups returned zero results in the journal; no screened paper trains log investigation, TP/FP triage, or response-action selection.
2. **No adaptive scenario generation:** the closest neighbor (PTPA, 2022) uses fixed simulation modalities; no paper generates scenarios adaptively from individual learner skill gaps.
3. **Comparative-effectiveness deficit:** the 2023 SLR in this journal explicitly finds the literature "lacks evaluation of comparative effectiveness" — the precise weakness our Pre-test → Adaptive Simulation → Post-test behavioral design addresses.

## 5. Proposed Contribution and Evidence Needed

**Working hypothesis (not a conclusion):** an adaptive, decision-based Blue Team simulation — mapping Security+ domain skill gaps to scenario generation, with log investigation, TP/FP classification, and response selection scored against measurable pre/post assessment — is a publishable contribution to IEEE ToE *provided* evidence goes beyond self-report (per the journal's stated aims & scope).

Evidence our system must produce:

- Pre/post total and per-domain scores under a matched blueprint (like PTPA's covariate design)
- Scenario-level behavioral outcomes: TP/FP accuracy, key-finding discovery, response-action accuracy
- Participant flow, attrition, and non-eligible reporting (one-group pretest–posttest, limitations stated)
- AI governance evidence: schema validation, prompt/model versioning, static fallback reliability
- Numerical functional/non-functional requirement metrics (per `docs/ADVISOR_PRESENTATION_PROGRESS_TASK.md`)

## 6. Questions for Advisor Decision

1. Is IEEE Transactions on Education the right venue, given its evidence bar (beyond self-report) and that our SOC-specific angle fills its coverage gap?
2. Research design: is one-group pretest–posttest acceptable as a pilot, or does the advisor expect a comparison group?
3. Target participant count for a pilot study credible at ToE?
4. Is institutional IEEE access available to strengthen evidence to full-text level before submission?
5. Should the adaptive-scenario LLM evaluation (GLM-5.3-Flash selection rationale) be part of the paper's methodology section?

---

### Caveats

- All three papers: evidence status `Abstract Verified` — research design details (sample sizes, statistical tests, effect sizes) are `Not Assessable` until full text is obtained. Upgrade path is planned.
- Candidate pool was below target (6 of 10–15): SOC-specific queries return zero in this journal. Accepted by user; a second search run can widen to other IEEE venues as auxiliary references if the advisor suggests.
