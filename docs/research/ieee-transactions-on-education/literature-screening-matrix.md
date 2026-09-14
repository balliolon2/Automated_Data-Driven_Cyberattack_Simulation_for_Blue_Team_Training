# Literature Screening Matrix — IEEE Transactions on Education

**Search Run:** `ieee-toe-2026-09-10-01`
**Date:** 2026-09-10
**Source:** IEEE Xplore (REST search, `punumber=13`)
**Year window:** 2021–2026 (foundational exceptions labeled)
**Status:** Preliminary — awaits user confirmation at Checkpoint

---

## 1. Configuration Snapshot

| Setting | Value |
|---|---|
| Target publication | IEEE Transactions on Education |
| Primary publication types | Research article |
| Year window | 2021–2026 |
| Candidate pool target | 10–15 |
| Query results per query | First 20 |
| Scoring weights | Evidence 30% / Learning Outcome 25% / Practical Cybersecurity 20% / Project Relevance 15% / Adaptivity 10% |
| Scale | 0–5 per criterion (shared rubric) |
| Tie-breakers | Evidence Quality, then Full-text availability |

### Query Groups executed

| Group | Queries | Results (unique in ToE) |
|---|---|---|
| Cybersecurity Education | `"cybersecurity education"`, `"security education"` | 13 |
| SOC / Security Operations Training | `"security operations" training`, `"SOC analyst"`, `SOC "incident triage"` | **0** |
| Incident Response Education | `"incident response" education/training` | **0** |
| Cyber Range / Simulation-Based Learning | `"cyber range"`, `"simulation-based learning" security` | **0** |
| Adaptive / Personalized Cybersecurity Learning | `adaptive learning cybersecurity`, `personalized learning security education` | 1 |
| Learning Analytics / Pretest-Posttest | `"pretest-posttest" learning`, `"pre-test post-test" assessment` | 6 |
| *Expansion (14 queries)* | SOC, blue team, CTF, hands-on, intelligent tutoring, etc. | 18 unique (mostly pre-2021 or adjacent) |

**Full-text availability note:** All six candidates are behind the IEEE paywall (`access=locked`) in this environment. Evidence status is therefore **Abstract Verified** (fetched from IEEE Xplore document metadata endpoints). No paywall bypass was attempted. Research design details cited below come from structured abstracts, which for IEEE ToE include Contribution / Background / Methodology / Findings fields — a strong but still abstract-level evidence base.

---

## 2. Candidate Pool — `Candidate Pool Below Target`

**Candidate pool count: 6** (rule: Metadata Verified+, year window 2021–2026). **Target was 10–15.**

Reason: SOC / security-operations-specific queries returned zero results within IEEE Transactions on Education — the journal's cybersecurity coverage centers on *education methodology* rather than SOC-specific training systems. Expansion queries were run (per workflow §2) and recovered additional records, but most were pre-2021 foundational or adjacent (general engineering education).

**Per workflow §2 Step 4:** the pool is reported as-is; no padding from other sources, no forced Top 2–3 ranking beyond what evidence supports. The three strongest candidates below are recommendable, but the user may prefer to accept fewer than three or approve a second search run.

---

## 3. Ranked Candidates (weighted score /100)

### 1. A Study of Cybersecurity Education Using a Present-Test-Practice-Assess Model — **91.0**

- Workman, Luévanos & Mai, *IEEE Trans. Education*, 65(1), Feb 2022, pp. 40–45. DOI: [10.1109/TE.2021.3086025](https://ieeexplore.ieee.org/document/9452797) · 31 citations (Scopus)
- **Criterion scores:** Evidence 5 · Learning Outcome 5 · Practical Training 5 · Project Relevance 4 · Adaptivity 2
- **Why it ranks first:** This is the closest methodological anchor for our project found in ToE. It *randomly assigned* students to four instruction modes (simulation, live competitive, combined, traditional), used **pretest scores as covariate and post-test as dependent variable**, and measured *applied behavioral performance* — exactly the evidence structure our Pre-test → Simulation → Post-test design needs. Its finding that "simulations improved learning performance over traditional instruction, with the greatest outcomes when simulations were combined with live competitive activities" directly supports the simulation-based intervention at the heart of our system.
- **Limitations (from abstract):** undergraduate CS students; single-institution; "applied exam" is the outcome measure, not longitudinal retention.

### 2. Immersive Virtual Environment for Automotive Security Education — **71.0**

- *IEEE Trans. Education*, 2026, pp. 86–101. DOI: [10.1109/TE.2026.3654077](https://ieeexplore.ieee.org/document/11372737)
- **Criterion scores:** Evidence 4 / LO 4 / PT 4 / PR 3 / Adaptivity 1
- **Relevance:** VR-based hands-on platform for automotive cyber-attack (ranging sensor) education; **user studies showed statistically significant improvements in knowledge, engagement, and self-efficacy**. Useful as recent evidence that immersive scenario platforms produce measurable learning gains, and as a contrast point (immersive environment vs. our decision-based log-investigation approach).

### 3. Blockchain Integration for Practical Cybersecurity Education With Scalable Cyber Infrastructure — **67.0**

- *IEEE Trans. Education*, 2025, pp. 543–552. DOI: [10.1109/TE.2025.3626489](https://ieeexplore.ieee.org/document/11229940)
- **Criterion scores:** Evidence 3 / LO 4 / PT 4 / PR 3 / Adaptivity 2
- **Relevance:** hands-on, evidence-based learning on the FABRIC testbed; evaluation via comprehensive student surveys — which triggers ToE's own caution about self-report data, making it a useful negative example for our stronger behavioral-evidence design.

### 4. Cybersecurity Education in the Age of AI: A Novel Proactive and Collaborative Learning Paradigm — **65.0**

- *IEEE Trans. Education*, 2024, pp. 395–404. DOI: [10.1109/TE.2023.3337337](https://ieeexplore.ieee.org/document/10367784) · 15 citations
- **Relevance:** learning paradigm for AI-cybersecurity skills using immersive scenarios and game-based collaborative learning; reports encouraging engagement results but work is ongoing — evidence quality is moderate (no controlled outcomes evident at abstract level).

### 5. A Systematic Literature Review on Cyber Security Education for Children — **59.0**

- *IEEE Trans. Education*, 2023, pp. 274–286. DOI: [10.1109/TE.2022.3231019](https://ieeexplore.ieee.org/document/10022010) · 51 citations
- **Relevance:** PRISMA-based SLR (412 papers screened, 44 analyzed). Its explicit finding that "existing literature lacks evaluation of comparative effectiveness" of innovative methods is a citable **research-gap statement** supporting our contribution. Domain (children) is far from SOC training.

### 6. Understanding Cybersecurity Education Gaps in Europe — **50.0**

- *IEEE Trans. Education*, 2024, pp. 190–201. DOI: [10.1109/TE.2023.3340868](https://ieeexplore.ieee.org/document/10380620) · 11 citations
- **Relevance:** PESTLE + stakeholder survey analysis of education-system factors; background/motivation material, not a training intervention.

---

## 4. Not Assessable / Insufficient Evidence

- All six candidates: **full text not accessible** (paywall) → research design details beyond the structured abstract (sample sizes, statistical tests, effect sizes) are `Not Assessable` at this evidence level. The 2022 paper (#1) states its design (random assignment, ANCOVA-style covariate use) in the abstract, which is why its Evidence score reaches 5.
- Pre-2021 security-core records (7, e.g. *Cloud-Based Virtual Laboratory* 2014, *NetSecLab* 2011, *Game Theory for Adversarial Thinking* 2017) are retained as `Foundational Exception` background, not counted in the pool.

---

## 5. Research Gap (Working Hypothesis — not a conclusion)

Based on abstract-level evidence across the screened pool (`Supported Interpretation`):

1. IEEE ToE publishes strong work on simulation-based and hands-on cybersecurity education (papers #1–#4), but **no paper in this run addresses SOC-analyst-specific decision training** (log investigation, TP/FP triage, response selection) — SOC/security-operations queries returned zero ToE results.
2. The 2022 closest-neighbor uses simulations but **static, non-adaptive scenarios**; no screened paper generates scenarios adaptively from learner skill gaps.
3. The 2023 SLR explicitly notes the literature lacks comparative-effectiveness evaluation — while our system's Pre/Post behavioral design is built to supply exactly that.
4. **Working hypothesis:** an adaptive, decision-based Blue Team simulation with measurable pre/post evidence is publishable in ToE if we can demonstrate rigorous evidence beyond self-report (per ToE aims & scope).

## 6. Open Questions for User Confirmation

1. **Accept pool of 6?** Or run a second search run (e.g., add non-ToE IEEE venues as auxiliary, or additional query groups) before locking Top 2–3?
2. **Confirm Top 2–3:** recommended #1 (PTPA Model, 2022), #2 (Automotive IVE, 2026), #3 (Blockchain/FABRIC, 2025) — or substitute #4 (AI Paradigm) for #3 given its closer AI angle?
3. **Full-text access:** do you have IEEE access via the university? If yes, the matrix can be upgraded to `Full Text Verified` for the three picks before the advisor meeting.
