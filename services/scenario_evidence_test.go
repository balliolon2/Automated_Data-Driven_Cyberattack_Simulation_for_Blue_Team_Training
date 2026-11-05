package services

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"cybersim/models"
)

func evidenceTestMicro(isTruePositive bool) aiMicroScenario {
	return aiMicroScenario{
		Title:             "Runtime Evidence Contract Test",
		Description:       "An alert fired on the test host.",
		IsTruePositive:    isTruePositive,
		TpFpExplanation:   "Deterministic explanation for the test scenario.",
		AttackerIP:        "198.51.100.7",
		TargetHost:        "SRV-TEST01",
		AttackLogMessage:  "Suspicious process execution detected on SRV-TEST01",
		KeyFinding:        "Suspicious process execution detected on SRV-TEST01",
		ContainmentAction: "Isolate SRV-TEST01 from the production network immediately and block 198.51.100.7 at the perimeter firewall",
		EradicationAction: "Remove the malicious process from SRV-TEST01 and rotate all exposed service credentials",
		RecoveryAction:    "Restore SRV-TEST01 from a verified clean image and monitor outbound traffic for 72 hours",
	}
}

func findingsForScenario(t *testing.T, s *models.Scenario) []map[string]interface{} {
	t.Helper()
	outcomes, ok := s.ExpectedOutcomes.(map[string]interface{})
	if !ok {
		t.Fatalf("expected_outcomes is not a map: %T", s.ExpectedOutcomes)
	}
	return mapSlice(outcomes["key_findings"])
}

func evidenceIndicesForFinding(t *testing.T, finding map[string]interface{}) []int {
	t.Helper()
	indices, errs := strictEvidenceIndices(finding["evidence_log_indices"], 1000)
	if len(errs) > 0 {
		t.Fatalf("unexpected evidence validation errors: %v", errs)
	}
	return indices
}

func deepCloneScenarioForTest(t *testing.T, s *models.Scenario) *models.Scenario {
	t.Helper()
	raw, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal scenario: %v", err)
	}
	var clone models.Scenario
	if err := json.Unmarshal(raw, &clone); err != nil {
		t.Fatalf("failed to unmarshal scenario: %v", err)
	}
	return &clone
}

func replaceFindingsForTest(t *testing.T, s *models.Scenario, findings []map[string]interface{}) {
	t.Helper()
	outcomes, ok := s.ExpectedOutcomes.(map[string]interface{})
	if !ok {
		t.Fatalf("expected_outcomes is not a map: %T", s.ExpectedOutcomes)
	}
	outcomes["key_findings"] = findings
}

func TestEvidenceMessageMapping_ExactThenTrimmedCaseInsensitive(t *testing.T) {
	ai := evidenceTestMicro(true)
	logs := buildScenarioLogs(ai, time.Unix(1700000000, 0).UTC())

	exactMessage, _ := logs[16]["message"].(string)
	if index, ok := resolveEvidenceMessageIndex(logs, exactMessage); !ok || index != 16 {
		t.Fatalf("exact message resolved to (%d, %v), want (16, true)", index, ok)
	}

	loose := "  " + strings.ToUpper(exactMessage) + "  "
	if index, ok := resolveEvidenceMessageIndex(logs, loose); !ok || index != 16 {
		t.Fatalf("trimmed/case-insensitive message resolved to (%d, %v), want (16, true)", index, ok)
	}

	if _, ok := resolveEvidenceMessageIndex(logs, "this message was never generated"); ok {
		t.Fatalf("expected unknown message to be unmappable")
	}
}

func TestMapAIFindingEvidence_RejectsDuplicateAndMissingMessages(t *testing.T) {
	ai := evidenceTestMicro(true)
	logs := buildScenarioLogs(ai, time.Unix(1700000000, 0).UTC())

	t.Run("duplicate message is rejected", func(t *testing.T) {
		finding := aiFinding{
			Description:         "duplicate evidence",
			EvidenceLogMessages: []string{ai.AttackLogMessage, ai.AttackLogMessage},
			VerdictSupport:      "malicious",
		}
		_, errs := mapAIFindingEvidence(logs, finding)
		if !containsErrorFragment(errs, "duplicate") {
			t.Fatalf("expected duplicate evidence error, got %v", errs)
		}
	})

	t.Run("missing message is rejected", func(t *testing.T) {
		finding := aiFinding{
			Description:         "missing evidence",
			EvidenceLogMessages: []string{"a message that does not exist in the log set"},
			VerdictSupport:      "malicious",
		}
		_, errs := mapAIFindingEvidence(logs, finding)
		if !containsErrorFragment(errs, "unmappable") {
			t.Fatalf("expected unmappable evidence error, got %v", errs)
		}
	})

	t.Run("empty evidence is rejected", func(t *testing.T) {
		finding := aiFinding{
			Description:         "empty evidence",
			EvidenceLogMessages: nil,
			VerdictSupport:      "malicious",
		}
		_, errs := mapAIFindingEvidence(logs, finding)
		if !containsErrorFragment(errs, "no evidence") {
			t.Fatalf("expected empty evidence error, got %v", errs)
		}
	})
}

func TestStrictEvidenceIndices_RejectsOutOfRangeAndNonInteger(t *testing.T) {
	if _, errs := strictEvidenceIndices([]interface{}{float64(99)}, 10); !containsErrorFragment(errs, "out of range") {
		t.Fatalf("expected out-of-range error, got %v", errs)
	}
	if _, errs := strictEvidenceIndices([]interface{}{1.5}, 10); !containsErrorFragment(errs, "invalid") {
		t.Fatalf("expected invalid index error, got %v", errs)
	}
	if _, errs := strictEvidenceIndices([]interface{}{-1}, 10); !containsErrorFragment(errs, "invalid") {
		t.Fatalf("expected negative index error, got %v", errs)
	}
}

func TestSynthesizeFullScenario_FalsePositiveAppendsSuppliedBenignMessage(t *testing.T) {
	ai := evidenceTestMicro(false)
	benignMessage := "Approved change CR-4242 authorizes maintenance on SRV-TEST01"
	ai.Findings = []aiFinding{
		{
			Description:         "Correlated approved change",
			DomainID:            "domain5",
			Points:              10,
			EvidenceLogMessages: []string{benignMessage},
			VerdictSupport:      "benign",
			Explanation:         "The change record proves the activity is authorized.",
		},
	}

	scenario, errs := synthesizeFullScenarioWithErrors(ai, "domain4", "Security Operations")
	if len(errs) > 0 {
		t.Fatalf("expected FP scenario to synthesize without errors, got %v", errs)
	}
	if err := ValidateScenarioCandidate(scenario); err != nil {
		t.Fatalf("expected synthesized FP scenario to validate, got %v", err)
	}

	logs := normalizePhaseStepSlice(scenario.InitialLogs)
	if len(logs) <= 45 {
		t.Fatalf("expected the benign context log to be appended, got %d logs", len(logs))
	}
	appended := logs[len(logs)-1]
	if message, _ := appended["message"].(string); message != benignMessage {
		t.Fatalf("appended message = %q, want %q", message, benignMessage)
	}
	if source, _ := appended["source"].(string); source != "change_management" {
		t.Fatalf("appended source = %q, want change_management", source)
	}

	findings := findingsForScenario(t, scenario)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	indices := evidenceIndicesForFinding(t, findings[0])
	if len(indices) != 1 || indices[0] != len(logs)-1 {
		t.Fatalf("benign evidence = %v, want [%d]", indices, len(logs)-1)
	}
}

func TestSynthesizeFullScenario_FalsePositiveSynthesizesBenignContextLog(t *testing.T) {
	ai := evidenceTestMicro(false)
	ai.Findings = []aiFinding{
		{
			Description:         "Alert-only benign explanation",
			DomainID:            "domain4",
			Points:              10,
			EvidenceLogMessages: []string{ai.AttackLogMessage},
			VerdictSupport:      "benign",
			Explanation:         "The LLM omitted a change record.",
		},
	}

	scenario, errs := synthesizeFullScenarioWithErrors(ai, "domain4", "Security Operations")
	if len(errs) > 0 {
		t.Fatalf("expected FP scenario to synthesize without errors, got %v", errs)
	}
	if err := ValidateScenarioCandidate(scenario); err != nil {
		t.Fatalf("expected synthesized FP scenario to validate, got %v", err)
	}

	logs := normalizePhaseStepSlice(scenario.InitialLogs)
	var contextIndex = -1
	for i, log := range logs {
		message, _ := log["message"].(string)
		if strings.HasPrefix(message, "CHANGE_MANAGEMENT: Approved change CR-") {
			contextIndex = i
			if !strings.Contains(message, "SRV-TEST01") {
				t.Fatalf("synthesized context message does not reference target host: %q", message)
			}
			break
		}
	}
	if contextIndex < 0 {
		t.Fatalf("expected a synthesized approved-change log")
	}

	findings := findingsForScenario(t, scenario)
	indices := evidenceIndicesForFinding(t, findings[0])
	if !containsIndex(indices, contextIndex) {
		t.Fatalf("benign evidence %v does not include synthesized context index %d", indices, contextIndex)
	}
}

func TestSynthesizeWithEvidenceRetry_UsesSingleRetryAndFallsBack(t *testing.T) {
	t.Run("retry succeeds with corrected evidence", func(t *testing.T) {
		initial := evidenceTestMicro(true)
		initial.Findings = []aiFinding{
			{
				Description:         "Unmappable finding",
				DomainID:            "domain4",
				Points:              10,
				EvidenceLogMessages: []string{"this log message does not exist"},
				VerdictSupport:      "malicious",
			},
		}

		retryCalls := 0
		retry := func(validationErrors []string, availableMessages []string) (aiMicroScenario, error) {
			retryCalls++
			if len(validationErrors) == 0 {
				t.Fatalf("expected validation errors passed to retry")
			}
			if !containsString(availableMessages, initial.AttackLogMessage) {
				t.Fatalf("expected available messages to include the generated attack message")
			}
			revised := evidenceTestMicro(true)
			revised.Findings = []aiFinding{
				{
					Description:         "Corrected finding",
					DomainID:            "domain4",
					Points:              10,
					EvidenceLogMessages: []string{revised.AttackLogMessage},
					VerdictSupport:      "malicious",
				},
			}
			return revised, nil
		}

		scenario, errs := synthesizeWithEvidenceRetry(initial, "domain4", "Security Operations", retry)
		if retryCalls != 1 {
			t.Fatalf("retry calls = %d, want 1", retryCalls)
		}
		if len(errs) > 0 {
			t.Fatalf("expected retried scenario to be valid, got %v", errs)
		}
		findings := findingsForScenario(t, scenario)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if description, _ := findings[0]["description"].(string); description != "Corrected finding" {
			t.Fatalf("finding description = %q, want corrected retry finding", description)
		}
	})

	t.Run("retry still invalid falls back to one valid finding", func(t *testing.T) {
		initial := evidenceTestMicro(true)
		initial.Findings = []aiFinding{
			{
				Description:         "Unmappable finding",
				DomainID:            "domain4",
				Points:              10,
				EvidenceLogMessages: []string{"this log message does not exist"},
				VerdictSupport:      "malicious",
			},
		}

		retryCalls := 0
		retry := func(validationErrors []string, availableMessages []string) (aiMicroScenario, error) {
			retryCalls++
			return initial, nil
		}

		scenario, errs := synthesizeWithEvidenceRetry(initial, "domain4", "Security Operations", retry)
		if retryCalls != 1 {
			t.Fatalf("retry calls = %d, want 1", retryCalls)
		}
		if len(errs) == 0 {
			t.Fatalf("expected validation errors after failed retry")
		}
		if err := ValidateScenarioCandidate(scenario); err != nil {
			t.Fatalf("expected fallback scenario to validate, got %v", err)
		}
		findings := findingsForScenario(t, scenario)
		if len(findings) != 1 {
			t.Fatalf("expected one fallback finding, got %d", len(findings))
		}
		indices := evidenceIndicesForFinding(t, findings[0])
		if len(indices) != 5 || indices[0] != 15 || indices[4] != 19 {
			t.Fatalf("TP fallback evidence = %v, want attack block 15..19", indices)
		}
	})

	t.Run("fallback for false positive uses benign context log", func(t *testing.T) {
		initial := evidenceTestMicro(false)
		initial.Findings = []aiFinding{
			{
				Description:         "Unmappable supporting finding",
				DomainID:            "domain4",
				Points:              10,
				EvidenceLogMessages: []string{"this log message does not exist"},
				VerdictSupport:      "supporting",
			},
		}

		scenario, errs := synthesizeWithEvidenceRetry(initial, "domain4", "Security Operations", nil)
		if len(errs) == 0 {
			t.Fatalf("expected validation errors for unmappable FP finding")
		}
		if err := ValidateScenarioCandidate(scenario); err != nil {
			t.Fatalf("expected FP fallback scenario to validate, got %v", err)
		}
		findings := findingsForScenario(t, scenario)
		indices := evidenceIndicesForFinding(t, findings[0])
		if len(indices) == 0 {
			t.Fatalf("expected FP fallback evidence")
		}
		for _, index := range indices {
			logs := normalizePhaseStepSlice(scenario.InitialLogs)
			if index < 0 || index >= len(logs) {
				t.Fatalf("fallback evidence index %d out of range", index)
			}
			if !isBenignContextLog(logs[index]) {
				t.Fatalf("FP fallback index %d does not identify a benign context log", index)
			}
		}
	})
}

func TestValidateScenarioCandidate_RejectsInvalidEvidenceContracts(t *testing.T) {
	base := synthesizeFullScenario(evidenceTestMicro(true), "domain4", "Security Operations")

	t.Run("rejects more than four findings", func(t *testing.T) {
		s := deepCloneScenarioForTest(t, base)
		findings := make([]map[string]interface{}, 0, 5)
		for i := 0; i < 5; i++ {
			findings = append(findings, map[string]interface{}{
				"id":                   "f" + string(rune('1'+i)),
				"description":          "finding",
				"domain_id":            "domain4",
				"points":               10,
				"evidence_log_indices": []int{i},
				"verdict_support":      "malicious",
			})
		}
		replaceFindingsForTest(t, s, findings)
		if err := ValidateScenarioCandidate(s); err == nil {
			t.Fatalf("expected >4 findings to be rejected")
		}
	})

	t.Run("rejects empty evidence", func(t *testing.T) {
		s := deepCloneScenarioForTest(t, base)
		findings := findingsForScenario(t, s)
		findings[0]["evidence_log_indices"] = []int{}
		replaceFindingsForTest(t, s, findings)
		if err := ValidateScenarioCandidate(s); err == nil {
			t.Fatalf("expected empty evidence to be rejected")
		}
	})

	t.Run("rejects union above fifty percent", func(t *testing.T) {
		s := deepCloneScenarioForTest(t, base)
		findings := make([]map[string]interface{}, 0, 4)
		for i := 0; i < 4; i++ {
			indices := make([]int, 0, 6)
			for j := 0; j < 6; j++ {
				indices = append(indices, i*6+j)
			}
			findings = append(findings, map[string]interface{}{
				"id":                   "f" + string(rune('1'+i)),
				"description":          "finding",
				"domain_id":            "domain4",
				"points":               10,
				"evidence_log_indices": indices,
				"verdict_support":      "malicious",
			})
		}
		replaceFindingsForTest(t, s, findings)
		err := ValidateScenarioCandidate(s)
		if err == nil || !strings.Contains(err.Error(), "union") {
			t.Fatalf("expected union cap error, got %v", err)
		}
	})

	t.Run("rejects missing verdict support", func(t *testing.T) {
		s := deepCloneScenarioForTest(t, base)
		findings := findingsForScenario(t, s)
		delete(findings[0], "verdict_support")
		replaceFindingsForTest(t, s, findings)
		if err := ValidateScenarioCandidate(s); err == nil {
			t.Fatalf("expected missing verdict support to be rejected")
		}
	})

	t.Run("rejects true positive without malicious support", func(t *testing.T) {
		s := deepCloneScenarioForTest(t, base)
		findings := findingsForScenario(t, s)
		findings[0]["verdict_support"] = "supporting"
		replaceFindingsForTest(t, s, findings)
		if err := ValidateScenarioCandidate(s); err == nil {
			t.Fatalf("expected TP without malicious support to be rejected")
		}
	})

	t.Run("rejects false positive without benign support", func(t *testing.T) {
		s := deepCloneScenarioForTest(t, base)
		s.IsTruePositive = false
		findings := findingsForScenario(t, s)
		findings[0]["verdict_support"] = "malicious"
		replaceFindingsForTest(t, s, findings)
		if err := ValidateScenarioCandidate(s); err == nil {
			t.Fatalf("expected FP without benign support to be rejected")
		}
	})

	t.Run("rejects false positive without benign context log", func(t *testing.T) {
		s := deepCloneScenarioForTest(t, base)
		s.IsTruePositive = false
		replaceFindingsForTest(t, s, []map[string]interface{}{
			{
				"id":                   "f1",
				"description":          "alert-only benign finding",
				"domain_id":            "domain4",
				"points":               10,
				"evidence_log_indices": []int{15, 16, 17, 18, 19},
				"verdict_support":      "benign",
			},
		})
		err := ValidateScenarioCandidate(s)
		if err == nil || !strings.Contains(err.Error(), "benign context") {
			t.Fatalf("expected benign context error, got %v", err)
		}
	})
}

func TestNormalizeAIFindings_LegacyKeyFindingCompatibility(t *testing.T) {
	t.Run("legacy true positive synthesizes malicious finding from attack log", func(t *testing.T) {
		ai := evidenceTestMicro(true)
		findings := normalizeAIFindings(ai)
		if len(findings) != 1 {
			t.Fatalf("expected one synthesized finding, got %d", len(findings))
		}
		if findings[0].VerdictSupport != "malicious" {
			t.Fatalf("verdict support = %q, want malicious", findings[0].VerdictSupport)
		}
		if len(findings[0].EvidenceLogMessages) != 5 || !containsString(findings[0].EvidenceLogMessages, ai.AttackLogMessage) {
			t.Fatalf("evidence messages = %v, want the legacy attack-log block including %q", findings[0].EvidenceLogMessages, ai.AttackLogMessage)
		}

		scenario, errs := synthesizeFullScenarioWithErrors(ai, "domain4", "Security Operations")
		if len(errs) > 0 {
			t.Fatalf("expected legacy TP to synthesize without errors, got %v", errs)
		}
		if err := ValidateScenarioCandidate(scenario); err != nil {
			t.Fatalf("expected legacy TP scenario to validate, got %v", err)
		}
	})

	t.Run("legacy false positive synthesizes benign finding from key finding", func(t *testing.T) {
		ai := evidenceTestMicro(false)
		ai.KeyFinding = "Approved change CR-123 covers planned maintenance on the test host"
		ai.Findings = nil

		scenario, errs := synthesizeFullScenarioWithErrors(ai, "domain4", "Security Operations")
		if len(errs) > 0 {
			t.Fatalf("expected legacy FP to synthesize without errors, got %v", errs)
		}
		if err := ValidateScenarioCandidate(scenario); err != nil {
			t.Fatalf("expected legacy FP scenario to validate, got %v", err)
		}

		logs := normalizePhaseStepSlice(scenario.InitialLogs)
		var found bool
		for _, log := range logs {
			message, _ := log["message"].(string)
			if message == ai.KeyFinding {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected legacy key finding message to be appended as benign context")
		}
	})
}

// legacyPreT7ExpectedOutcomes mirrors the pre-T7 AI row shape: one finding with
// the attack-log evidence block and no verdict_support field.
const legacyPreT7ExpectedOutcomes = `{
  "key_findings": [
    {
      "id": "f1",
      "description": "Identified malicious activity from the attack-log block",
      "domain_id": "domain4",
      "points": 10,
      "evidence_log_indices": [15, 16, 17, 18, 19],
      "explanation": "The attack-log block proves the activity is malicious."
    }
  ]
}`

func legacyPreT7Scenario(t *testing.T, isTruePositive bool) *models.Scenario {
	t.Helper()
	ai := evidenceTestMicro(isTruePositive)
	scenario := &models.Scenario{
		ScenarioID:     "legacy-pre-t7",
		Title:          "Legacy Pre-T7 AI Scenario",
		Status:         "active",
		IsTruePositive: isTruePositive,
		InitialLogs:    buildScenarioLogs(ai, time.Unix(1700000000, 0).UTC()),
	}
	var outcomes map[string]interface{}
	if err := json.Unmarshal([]byte(legacyPreT7ExpectedOutcomes), &outcomes); err != nil {
		t.Fatalf("failed to unmarshal legacy expected_outcomes: %v", err)
	}
	scenario.ExpectedOutcomes = outcomes
	return scenario
}

func TestShouldArchiveScenarioEvidence(t *testing.T) {
	seedTitles := map[string]struct{}{"static scenario": {}}

	t.Run("never archives a static authored scenario", func(t *testing.T) {
		static := &models.Scenario{
			Title:            "Static Scenario",
			Status:           "active",
			IsTruePositive:   true,
			InitialLogs:      make([]interface{}, 5),
			ExpectedOutcomes: map[string]interface{}{},
		}
		shouldArchive, _ := shouldArchiveScenarioEvidence(static, seedTitles)
		if shouldArchive {
			t.Fatalf("static authored scenario must never be archived")
		}
	})

	t.Run("archives an AI false positive without benign context", func(t *testing.T) {
		ai := evidenceTestMicro(false)
		scenario := synthesizeFullScenario(ai, "domain4", "Security Operations")
		scenario.Title = "AI FP Without Context"
		replaceFindingsForTest(t, scenario, []map[string]interface{}{
			{
				"id":                   "f1",
				"description":          "alert-only benign finding",
				"domain_id":            "domain4",
				"points":               10,
				"evidence_log_indices": []int{15, 16, 17, 18, 19},
				"verdict_support":      "benign",
			},
		})
		shouldArchive, err := shouldArchiveScenarioEvidence(scenario, seedTitles)
		if !shouldArchive || err == nil {
			t.Fatalf("expected invalid FP AI scenario to be archived, got shouldArchive=%v err=%v", shouldArchive, err)
		}
	})

	t.Run("keeps a pre-T7 AI true positive with attack-log evidence and no verdict support", func(t *testing.T) {
		scenario := legacyPreT7Scenario(t, true)
		shouldArchive, err := shouldArchiveScenarioEvidence(scenario, seedTitles)
		if shouldArchive || err != nil {
			t.Fatalf("expected legacy AI TP with attack-log evidence to remain active, got shouldArchive=%v err=%v", shouldArchive, err)
		}
	})

	t.Run("remediation validation infers malicious support for a legacy true positive", func(t *testing.T) {
		scenario := legacyPreT7Scenario(t, true)
		if err := validateScenarioEvidenceContract(scenario); err == nil {
			t.Fatalf("strict validation should reject the missing verdict_support field")
		}
		if err := validateScenarioEvidenceContractForRemediation(scenario); err != nil {
			t.Fatalf("remediation validation should infer malicious support, got %v", err)
		}
	})

	t.Run("archives a pre-T7 AI false positive with attack-log evidence and no verdict support", func(t *testing.T) {
		scenario := legacyPreT7Scenario(t, false)
		shouldArchive, err := shouldArchiveScenarioEvidence(scenario, seedTitles)
		if !shouldArchive || err == nil {
			t.Fatalf("expected legacy AI FP without benign context to be archived, got shouldArchive=%v err=%v", shouldArchive, err)
		}
	})

	t.Run("keeps a non-seed static false positive that only lacks verdict support", func(t *testing.T) {
		scenario := synthesizeFullScenario(evidenceTestMicro(false), "domain4", "Security Operations")
		scenario.Title = "Non-Seed Static FP"
		logs := normalizePhaseStepSlice(scenario.InitialLogs)
		contextIndex := len(logs) - 1
		replaceFindingsForTest(t, scenario, []map[string]interface{}{
			{
				"id":                   "f1",
				"description":          "approved change context",
				"domain_id":            "domain5",
				"points":               10,
				"evidence_log_indices": []int{contextIndex},
			},
		})
		shouldArchive, err := shouldArchiveScenarioEvidence(scenario, seedTitles)
		if shouldArchive || err != nil {
			t.Fatalf("expected non-seed static FP with benign context to remain active, got shouldArchive=%v err=%v", shouldArchive, err)
		}
	})

	t.Run("never archives a true positive row", func(t *testing.T) {
		ai := evidenceTestMicro(true)
		scenario := synthesizeFullScenario(ai, "domain4", "Security Operations")
		scenario.Title = "Invalid AI TP"
		replaceFindingsForTest(t, scenario, []map[string]interface{}{
			{
				"id":              "f1",
				"description":     "no evidence",
				"domain_id":       "domain4",
				"points":          10,
				"verdict_support": "malicious",
			},
		})
		shouldArchive, _ := shouldArchiveScenarioEvidence(scenario, seedTitles)
		if shouldArchive {
			t.Fatalf("expected only False Positive rows to be archived")
		}
	})
}

func TestEnsureScenarioEvidenceQuality_RequiresSeed(t *testing.T) {
	t.Run("path wrapper returns an error for a missing seed", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "does-not-exist.json")
		if err := EnsureScenarioEvidenceQuality(nil, missing); err == nil {
			t.Fatalf("expected missing seed path to return an error")
		}
	})

	t.Run("empty seed returns an error", func(t *testing.T) {
		if err := EnsureScenarioEvidenceQualityFromSeed(nil, nil); err == nil {
			t.Fatalf("expected empty seed to return an error")
		}
	})

	t.Run("invalid JSON seed returns an error", func(t *testing.T) {
		if err := EnsureScenarioEvidenceQualityFromSeed(nil, []byte("not json")); err == nil {
			t.Fatalf("expected invalid seed JSON to return an error")
		}
	})

	t.Run("seed without titles returns an error", func(t *testing.T) {
		if err := EnsureScenarioEvidenceQualityFromSeed(nil, []byte(`[{"description":"no title"}]`)); err == nil {
			t.Fatalf("expected seed without titles to return an error")
		}
	})

	t.Run("valid seed is accepted", func(t *testing.T) {
		if err := EnsureScenarioEvidenceQualityFromSeed(nil, []byte(`[{"title":"Static Scenario"}]`)); err != nil {
			t.Fatalf("expected valid seed to be accepted, got %v", err)
		}
	})
}

func TestSynthesizedBenignContextMessage_IsDeterministic(t *testing.T) {
	ai := evidenceTestMicro(false)
	first := synthesizedBenignContextMessage(ai)
	second := synthesizedBenignContextMessage(ai)
	if first != second {
		t.Fatalf("expected deterministic message, got %q and %q", first, second)
	}
	pattern := regexp.MustCompile(`^CHANGE_MANAGEMENT: Approved change CR-\d+ authorizes the observed activity on SRV-TEST01 during this window\.$`)
	if !pattern.MatchString(first) {
		t.Fatalf("unexpected synthesized message: %q", first)
	}
}

func containsErrorFragment(errs []string, fragment string) bool {
	for _, err := range errs {
		if strings.Contains(strings.ToLower(err), strings.ToLower(fragment)) {
			return true
		}
	}
	return false
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsIndex(indices []int, want int) bool {
	for _, index := range indices {
		if index == want {
			return true
		}
	}
	return false
}
