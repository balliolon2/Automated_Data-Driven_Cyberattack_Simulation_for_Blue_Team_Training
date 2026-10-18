# 01: Contextual Scenario Playbook Synthesis & Anti-Predictability Validation Gate

**What to build:** Upgrade the AI Scenario synthesizer, contextual fallback generator, and Scenario candidate validation gate so that every generated Scenario (`ai_generated` and synthesized fallback across all five CompTIA Security+ domains) produces 3 to 4 balanced, context-specific Response Actions in each incident response phase (Containment, Eradication, and Recovery). Each phase must include 1–2 correct Response Actions (`+5` points) and at least 2 plausible, scenario-specific distractor Response Actions (`-2` to `-3` points) that match the sentence length (12–35 words), operational tone, and entity specificity (`target_host`, `attacker_ip`, affected accounts, or services) of the correct options, supporting both True Positive and False Positive Scenarios.

**Blocked by:** None (can start immediately)

**Status:** completed

- [x] The LLM micro-scenario contract and prompt support structured Response Action arrays per phase (`containment`, `eradication`, `recovery`) with `label`, `is_correct`, and `explanation`, while maintaining backward compatibility when single-action fields are returned by backfilling via a contextual distractor generator.
- [x] All five domain fallback templates (`domain1` through `domain5`) and the contextual distractor generator produce 3 to 4 Response Actions per phase (1–2 correct, at least 2 plausible distractors) that interpolate scenario entities (`target_host`, `attacker_ip`, affected service/account) at comparable length and technical specificity.
- [x] False Positive Scenarios (`is_true_positive: false`) synthesize non-destructive verification and tuning actions as correct Response Actions and unnecessary disruptive actions as distractors.
- [x] The Scenario validation gate enforces `>= 3` Response Actions per phase, `>= 1` correct option per phase, `>= 2` distractor options per phase, case-insensitive label uniqueness across the Scenario, and zero occurrences of the banned legacy boilerplate strings (`"Power off all domain controllers immediately"`, `"Reformat all network switches in the subnet"`, `"Disable MFA enforcement for affected users"`).
- [x] Unit tests verify that all five domain fallback Scenarios pass the validation gate and that invalid Scenario candidates (insufficient options, missing distractors, duplicate labels, or banned boilerplate labels) are rejected.
