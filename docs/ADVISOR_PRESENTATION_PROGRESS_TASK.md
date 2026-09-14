# Advisor Presentation Progress Task

## Purpose

This document records the work required after the latest advisor presentation. It converts the advisor's feedback into clear research, engineering, security, and publication tasks for the project team and AI agents.

**Project:** Automated Data-Driven Cyberattack Simulation for Blue Team Training  
**Current direction:** Register → Evaluate User → Pre-test → Learn through Adaptive Blue Team Scenarios → Post-test  
**Target publication venue under investigation:** [IEEE Transactions on Education](https://ieeexplore.ieee.org/xpl/RecentIssue.jsp?punumber=13)

> **Status:** Advisor feedback captured. The items below are investigation and design tasks, not completed claims.

---

## 1. Publication Venue Investigation

### Advisor feedback

Select two or three relevant papers from the IEEE publication venue and discuss them in the next advisor meeting. The project may be developed into a publishable version.

### Tasks

- [ ] Review recent issues and papers from *IEEE Transactions on Education*.
- [ ] Select **2–3 papers** that are closely related to at least one of the following topics:
  - Computer or cybersecurity education
  - Simulation-based or scenario-based learning
  - Adaptive or personalized learning systems
  - Learning analytics and evidence-based educational evaluation
  - Assessment of practical engineering or cybersecurity skills
- [ ] For each selected paper, record:
  - Research problem
  - Learning intervention or system contribution
  - Participant population and sample size
  - Research design
  - Functional measurements and learning outcomes
  - Statistical analysis
  - Evidence supporting the conclusions
  - Limitations
  - Potential connection to this project
- [ ] Compare the selected papers with this project and identify:
  - The potential research gap
  - The proposed contribution of this system
  - The evidence required to support the contribution
  - The expected publication risks
- [ ] Prepare a short comparison table and discussion points for the next advisor meeting.
- [ ] Verify the current author guidelines, article types, submission requirements, and publication options before treating the venue as a final target.

### Publication-oriented requirement

The publication version must go beyond user satisfaction or self-reported opinions. It should provide transparent evidence from system behavior, assessment results, learning outcomes, and reproducible analysis. The current research direction is a one-group pretest–posttest design, but the final design and its limitations must be discussed with the advisor.

---

## 2. Publication-Version User Journey

The publication version should clearly define and instrument the complete learner journey:

```text
Register
  → Evaluate User
  → Pre-test
  → Build User Skill Profile
  → Adaptive Blue Team Learning Scenarios
  → Evaluate Scenario Performance
  → Check Proficiency Threshold
  → Post-test
  → Compare Pre-test, Learning, and Post-test Results
```

### Tasks

- [ ] Define the purpose and measurable output of each stage.
- [ ] Specify the data collected at each stage.
- [ ] Ensure that every score can be traced back to the learner's original answers, actions, scenario, rubric, and system version.
- [ ] Define the conditions under which a learner is eligible to take the Post-test.
- [ ] Record learners who abandon the process or do not reach the proficiency threshold.
- [ ] Report participant flow, missing data, abandonment, and eligibility rather than analyzing only completed cases.
- [ ] Preserve version metadata for the assessment blueprint, scenario, scoring rubric, prompt, and AI model.

### Minimum evidence expected from the system

- Pre-test total score and domain-level scores
- Simulation performance and domain-level proficiency
- True Positive / False Positive classification accuracy
- Key finding discovery score
- Response action score
- Time to complete each activity
- Number of scenarios completed
- Post-test total score and domain-level scores
- Eligibility, abandonment, and missing-data records

---

## 3. Engineering Measurement: Functional and Non-functional Requirements

The system must be evaluated using numerical and reproducible measurements. Requirements should not be written only as qualitative statements such as “the system works well” or “the system is secure.”

### 3.1 Functional requirements

Functional requirements describe what the system must do. Each requirement should have an observable acceptance metric.

| Area | Example requirement | Numerical measurement | Evidence source |
|---|---|---:|---|
| Registration | A new learner can create an account with valid data | Registration success rate ≥ 99% in test runs | API/integration test |
| Authentication | A valid learner can log in and receive an authenticated session | Successful login rate; invalid-login rejection rate | API test and audit log |
| User evaluation | The system creates an initial learner profile from evaluation results | Profile creation completion rate = 100% for valid inputs | Database and API test |
| Pre-test | The system assigns the required assessment blueprint | 30 questions with the specified domain distribution | Integration test |
| Scenario assignment | The system selects scenarios based on the learner's skill gap | Assignment traceability rate = 100% | Skill profile and scenario metadata |
| Investigation | The learner can inspect logs and identify evidence | Key finding discovery score | Scenario evaluation record |
| Triage | The learner can classify an event as TP or FP | TP/FP accuracy | Scoring record |
| Response | The learner can select response actions | Response action accuracy | Scoring record |
| Post-test | Eligible learners can complete a comparable assessment | Completion rate; blueprint validity | Integration and research dataset |
| Assessment integrity | Post-test questions do not overlap with the learner's Pre-test questions | Question overlap = 0 | Database query and verification test |
| Recovery | The learner can resume an interrupted session | Session recovery success rate | Recovery test |
| AI fallback | The system remains usable when the LLM is unavailable | Fallback activation success rate; failed-session rate | Fault-injection test |

> The threshold values above are initial engineering targets. They must be validated and revised with the advisor before being reported as final project requirements.

### 3.2 Non-functional requirements

Non-functional requirements describe quality attributes and operational constraints.

| Quality attribute | Numerical measurement | Suggested reporting |
|---|---:|---|
| Performance | API response latency, p50/p95/p99 | Report milliseconds under a defined workload |
| Availability | Successful request percentage during a test period | Report uptime and failed-request rate |
| Scalability | Concurrent users supported before an agreed degradation limit | Report users, throughput, and latency |
| Reliability | Error rate and successful recovery rate | Include fault-injection results |
| Security | Authentication rejection rate, authorization-test pass rate, dependency findings | Report test cases and severity of findings |
| Data integrity | Invalid or incomplete research records | Target zero invalid records in controlled tests |
| Reproducibility | Percentage of scenarios and scores reproducible from stored metadata | Report replay success rate |
| Usability | Task completion rate and task completion time | Use as secondary evidence, not the only evidence |
| Maintainability | Test coverage or verified critical-path coverage | Report the defined coverage scope |
| Safety | Unsafe external-target execution attempts blocked | Target 100% blocked in security tests |

### Measurement rules

- [ ] Define the workload, environment, dataset, and test procedure for every metric.
- [ ] Define a target, threshold, or acceptance range before collecting final results.
- [ ] Record raw measurements, not only averages.
- [ ] Report sample size, variance or distribution, and test conditions.
- [ ] Distinguish engineering acceptance thresholds from research hypotheses.
- [ ] Do not claim that a requirement is satisfied without test evidence.

---

## 4. System Differentiation and Comparative Investigation

The project must explain how it is better than, or meaningfully different from, existing systems. “Uses AI” is not sufficient as a contribution claim.

### Investigation questions

- [ ] What existing cybersecurity training platforms, cyber ranges, learning systems, or assessment tools address a similar problem?
- [ ] Which systems support adaptive learning based on learner performance?
- [ ] Which systems evaluate investigation behavior rather than only multiple-choice knowledge?
- [ ] Which systems connect security education content to logs, evidence, TP/FP classification, and response actions?
- [ ] Which systems provide measurable Pre-test → Learning → Post-test evidence?
- [ ] Which systems use an LLM, and how do they control output quality, safety, reproducibility, and fallback behavior?
- [ ] Which systems report domain-level or skill-level learning outcomes?

### Comparison dimensions

| Dimension | Existing system evidence | This project | Verification method |
|---|---|---|---|
| Learning model |  | Adaptive Blue Team decision simulation | Documentation and experiment |
| Assessment type |  | Pre-test and Post-test with domain mapping | Blueprint comparison |
| Practical behavior |  | Logs, key findings, TP/FP, response actions | Scenario records |
| Adaptation |  | Scenario selection from the learner skill profile | Assignment trace |
| Evidence |  | Event-level and outcome-level data | Research dataset |
| AI governance |  | Validation, versioning, retry, and static fallback | Fault-injection and audit tests |
| Safety |  | Simulated or authorized data only | Security test suite |
| Research readiness |  | Reproducible metrics and participant-flow reporting | Analysis protocol |

### Expected contribution statement

The final contribution statement must be evidence-based and narrow. A possible direction for investigation is:

> This project investigates whether an adaptive, decision-based Blue Team simulation that connects Security+ domain weaknesses to log investigation, TP/FP classification, response selection, and measurable pre/post assessment can provide stronger evidence of practical cybersecurity learning than assessment-only training.

This is a research hypothesis or contribution direction, not a conclusion. It must be revised after the comparative investigation and experimental results are available.

---

## 5. System Security Review

Because the system handles learner accounts, assessment results, research data, and AI-generated content, security must be treated as a project requirement and an evaluation topic.

### Security areas to review

- [ ] Authentication: password hashing, login protection, token/session handling, expiration, and revocation strategy.
- [ ] Authorization: learners can access only their own sessions and results; administrative functions are protected.
- [ ] Input validation: request payloads, query parameters, file inputs, and LLM outputs are validated.
- [ ] SQL and API security: use parameterized queries, safe ORM patterns, rate limiting where appropriate, and controlled error messages.
- [ ] Session integrity: prevent cross-user access, answer tampering, replay, duplicate submission, and unauthorized Post-test access.
- [ ] Assessment integrity: preserve raw answers, prevent question leakage where possible, and guarantee zero Pre-test/Post-test overlap for the same learner.
- [ ] LLM security: do not send unnecessary PII; defend against prompt injection through retrieved content; validate generated output against a strict schema.
- [ ] AI safety: use only simulated or authorized telemetry; block external targets and unsafe execution instructions.
- [ ] Fallback safety: static fallback scenarios must be reviewed, deterministic, and subject to the same validation rules.
- [ ] Privacy and research ethics: pseudonymous participant IDs, consent, access control, retention, withdrawal, and separation of authentication data from research data.
- [ ] Auditability: record model, prompt, scenario, rubric, validation result, and important state transitions.
- [ ] Dependency and deployment security: secret management, dependency scanning, secure configuration, and least-privilege database access.

### Security evidence to produce

- [ ] Threat model and trust-boundary diagram
- [ ] Security requirements and abuse cases
- [ ] Authentication and authorization test results
- [ ] Input-validation and API security test results
- [ ] LLM output validation and prompt-injection test results
- [ ] Unsafe-target blocking test results
- [ ] Dependency and secret-scan results
- [ ] Security limitations and residual risks

---

## 6. AI Model Selection: GLM-5.3-Flash

The project currently uses **GLM-5.3-Flash**. The publication version must explain why this model was selected instead of presenting the choice as an unsupported preference.

### Model-selection criteria

- [ ] Task quality: ability to generate complete scenario content that follows the Scenario Contract.
- [ ] Structured output: ability to produce valid JSON or another schema-conformant response.
- [ ] Latency: time required to generate a scenario within the system's response budget.
- [ ] Cost efficiency: cost per generated scenario under the project workload.
- [ ] Reliability: success rate, timeout rate, retry rate, and fallback rate.
- [ ] Safety: ability to follow the simulated-training boundary and avoid unsafe external-target instructions.
- [ ] Context handling: ability to use skill-gap, Security+ domain, MITRE technique, and scenario blueprint information.
- [ ] Reproducibility: availability of model/version metadata and prompt-version recording.
- [ ] Deployment compatibility: compatibility with the project's OpenAI-compatible provider interface and operational environment.

### Required comparison method

- [ ] Define a fixed evaluation set of scenario blueprints.
- [ ] Evaluate GLM-5.3-Flash using the same prompts, constraints, and validation rules for every test case.
- [ ] If feasible, compare it with at least one alternative model or a static baseline.
- [ ] Use expert review or a predefined rubric to score correctness, completeness, educational alignment, safety, and schema validity.
- [ ] Record latency, token usage or cost, retry count, validation failures, and fallback activations.
- [ ] Store the model name, provider, model version, prompt version, blueprint version, and generation result for every scenario.
- [ ] Report the limitations of the comparison, including provider changes, model updates, and sample-size constraints.

### Evidence-based rationale template

> GLM-5.3-Flash was selected because it satisfied the project's defined requirements for structured scenario generation, response latency, operational cost, provider compatibility, and safety validation. The selection will be supported by a controlled evaluation against fixed scenario blueprints and a predefined rubric. The system will not rely on the model as the sole source of truth: generated scenarios must pass schema and safety validation, and a reviewed static fallback must remain available.

This paragraph is a template. Replace it with measured results before using it in a thesis or publication.

---

## 7. Deliverables for the Next Advisor Meeting

- [ ] A shortlist of **2–3 IEEE papers** with a comparison table.
- [ ] A one-page publication-version system flow.
- [ ] A functional and non-functional measurement matrix with proposed thresholds.
- [ ] A comparison matrix against existing systems.
- [ ] A first-pass system security review and threat model scope.
- [ ] A GLM-5.3-Flash model-selection plan and evaluation criteria.
- [ ] A list of unresolved decisions requiring advisor approval:
  - Research design and comparison group
  - Participant sample size
  - Proficiency threshold and maximum scenario count
  - Primary outcome and statistical analysis
  - Security and ethics approval requirements
  - Final publication venue

---

## 8. Completion Criteria

This task is complete for the next advisor meeting when:

1. Two or three relevant IEEE papers have been selected and analyzed using the same comparison template.
2. The complete Register → Evaluate User → Pre-test → Learn → Post-test flow is defined with measurable outputs.
3. Functional and non-functional requirements have numerical acceptance metrics and test procedures.
4. The project has a documented, evidence-based differentiation hypothesis against existing systems.
5. Security requirements, threats, and planned verification activities are documented.
6. The GLM-5.3-Flash selection has a reproducible evaluation plan and does not rely only on subjective preference.
7. Open decisions are clearly marked for discussion with the advisor.

> No research conclusion should be claimed until the planned measurements and comparative evidence have been collected and analyzed.
