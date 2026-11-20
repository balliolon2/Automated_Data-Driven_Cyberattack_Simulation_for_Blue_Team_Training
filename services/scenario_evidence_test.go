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
		Title:               "Runtime Evidence Contract Test",
		TitleTh:             "การทดสอบความสมบูรณ์ของหลักฐานในระบบจำลอง",
		Description:         "An alert fired on the test host.",
		DescriptionTh:       "ตรวจพบการแจ้งเตือนบนโฮสต์ทดสอบ SRV-TEST01",
		IsTruePositive:      isTruePositive,
		TpFpExplanation:     "Deterministic explanation for the test scenario.",
		TpFpExplanationTh:   "คำอธิบายผลเฉลยสำหรับการทดสอบสถานการณ์จำลอง",
		AttackerIP:          "198.51.100.7",
		TargetHost:          "SRV-TEST01",
		AttackLogMessage:    "Suspicious process execution detected on SRV-TEST01",
		KeyFinding:          "Suspicious process execution detected on SRV-TEST01",
		KeyFindingTh:        "ตรวจพบการรันโพรเซสต้องสงสัยบน SRV-TEST01",
		ContainmentAction:   "Isolate SRV-TEST01 from the production network immediately and block 198.51.100.7 at the perimeter firewall",
		ContainmentActionTh: "แยกโฮสต์ SRV-TEST01 ออกจากเครือข่ายระบบงานจริงทันทีและบล็อก 198.51.100.7 ที่ perimeter firewall",
		EradicationAction:   "Remove the malicious process from SRV-TEST01 and rotate all exposed service credentials",
		EradicationActionTh: "ลบโพรเซสที่เป็นอันตรายออกจาก SRV-TEST01 และหมุนเวียน credentials ทั้งหมดที่อาจรั่วไหล",
		RecoveryAction:      "Restore SRV-TEST01 from a verified clean image and monitor outbound traffic for 72 hours",
		RecoveryActionTh:    "กู้คืน SRV-TEST01 จาก clean image ที่ตรวจสอบแล้วและเฝ้าระวังทราฟฟิกขาออก 72 ชั่วโมง",
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

	exactMessage := ai.AttackLogMessage
	expectedIdx := -1
	for i, entry := range logs {
		if msg, _ := entry["message"].(string); msg == exactMessage {
			expectedIdx = i
			break
		}
	}
	if expectedIdx < 0 {
		t.Fatalf("expected attack log message %q to appear in interleaved logs", exactMessage)
	}
	if index, ok := resolveEvidenceMessageIndex(logs, exactMessage); !ok || index != expectedIdx {
		t.Fatalf("exact message resolved to (%d, %v), want (%d, true)", index, ok, expectedIdx)
	}

	loose := "  " + strings.ToUpper(exactMessage) + "  "
	if index, ok := resolveEvidenceMessageIndex(logs, loose); !ok || index != expectedIdx {
		t.Fatalf("trimmed/case-insensitive message resolved to (%d, %v), want (%d, true)", index, ok, expectedIdx)
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
			DescriptionTh:       "เชื่อมโยงพบรายการขออนุมัติเปลี่ยนแปลงระบบ",
			DomainID:            "domain5",
			Points:              10,
			EvidenceLogMessages: []string{benignMessage},
			VerdictSupport:      "benign",
			Explanation:         "The change record proves the activity is authorized.",
			ExplanationTh:       "บันทึก Change Control ยืนยันว่ากิจกรรมนี้ได้รับอนุญาต",
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
	if len(logs) < 28 || len(logs) > 46 {
		t.Fatalf("expected 28..46 interleaved logs, got %d logs", len(logs))
	}
	benignIdx := -1
	for i, entry := range logs {
		if msg, _ := entry["message"].(string); msg == benignMessage {
			benignIdx = i
			if source, _ := entry["source"].(string); source != "change_management" {
				t.Fatalf("benign log source = %q, want change_management", source)
			}
			break
		}
	}
	if benignIdx < 0 {
		t.Fatalf("expected supplied benign context message %q in interleaved logs", benignMessage)
	}

	findings := findingsForScenario(t, scenario)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	indices := evidenceIndicesForFinding(t, findings[0])
	if len(indices) != 1 || indices[0] != benignIdx {
		t.Fatalf("benign evidence = %v, want [%d]", indices, benignIdx)
	}
}

func TestSynthesizeFullScenario_FalsePositiveSynthesizesBenignContextLog(t *testing.T) {
	ai := evidenceTestMicro(false)
	ai.Findings = []aiFinding{
		{
			Description:         "Alert-only benign explanation",
			DescriptionTh:       "คำอธิบายเหตุการณ์ปลอดภัยจาก log เดิม",
			DomainID:            "domain4",
			Points:              10,
			EvidenceLogMessages: []string{ai.AttackLogMessage},
			VerdictSupport:      "benign",
			Explanation:         "The LLM omitted a change record.",
			ExplanationTh:       "โมเดลไม่ได้สร้างบันทึก Change Record มาด้วยระบบจึงสังเคราะห์ให้",
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
		if len(indices) == 0 {
			t.Fatalf("expected non-empty TP fallback evidence")
		}
		logs := normalizePhaseStepSlice(scenario.InitialLogs)
		for _, idx := range indices {
			if idx < 0 || idx >= len(logs) {
				t.Fatalf("fallback index %d out of range for %d logs", idx, len(logs))
			}
			msg, _ := logs[idx]["message"].(string)
			if strings.Contains(strings.ToLower(msg), "corr-9901") {
				t.Fatalf("fallback evidence contains legacy CORR-9901 boilerplate: %q", msg)
			}
		}
		if len(indices) >= 2 && indices[len(indices)-1]-indices[0]+1 == len(indices) {
			t.Fatalf("expected interleaved non-contiguous fallback evidence indices, got contiguous %v", indices)
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
		highIdx := -1
		for idx, entry := range normalizePhaseStepSlice(s.InitialLogs) {
			if scenarioLogSeverityBucket(entry) == "high" {
				highIdx = idx
				break
			}
		}
		if highIdx < 0 {
			t.Fatalf("expected at least one high-severity log in base scenario")
		}
		replaceFindingsForTest(t, s, []map[string]interface{}{
			{
				"id":                   "f1",
				"description":          "alert-only benign finding",
				"domain_id":            "domain4",
				"points":               10,
				"evidence_log_indices": []int{highIdx},
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
		if len(findings[0].EvidenceLogMessages) < 2 || !containsString(findings[0].EvidenceLogMessages, ai.AttackLogMessage) {
			t.Fatalf("evidence messages = %v, want contextual attack chain including %q", findings[0].EvidenceLogMessages, ai.AttackLogMessage)
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

	t.Run("keeps valid multi-event non-seed TP and FP scenarios active", func(t *testing.T) {
		tp := synthesizeFullScenario(evidenceTestMicro(true), "domain4", "Security Operations")
		tp.Title = "Valid Non-Seed TP"
		if shouldArchive, err := shouldArchiveScenarioEvidence(tp, seedTitles); shouldArchive || err != nil {
			t.Fatalf("expected valid TP scenario to stay active, got shouldArchive=%v err=%v", shouldArchive, err)
		}

		fp := synthesizeFullScenario(evidenceTestMicro(false), "domain1", "General Security Concepts")
		fp.Title = "Valid Non-Seed FP"
		if shouldArchive, err := shouldArchiveScenarioEvidence(fp, seedTitles); shouldArchive || err != nil {
			t.Fatalf("expected valid FP scenario to stay active, got shouldArchive=%v err=%v", shouldArchive, err)
		}
	})

	t.Run("archives an AI false positive without benign context", func(t *testing.T) {
		ai := evidenceTestMicro(false)
		scenario := synthesizeFullScenario(ai, "domain4", "Security Operations")
		scenario.Title = "AI FP Without Context"
		alertIdx := -1
		for idx, entry := range normalizePhaseStepSlice(scenario.InitialLogs) {
			if !isBenignContextLog(entry) {
				alertIdx = idx
				break
			}
		}
		if alertIdx < 0 {
			t.Fatalf("expected at least one alert-severity log in scenario")
		}
		replaceFindingsForTest(t, scenario, []map[string]interface{}{
			{
				"id":                   "f1",
				"description":          "alert-only benign finding",
				"domain_id":            "domain4",
				"points":               10,
				"evidence_log_indices": []int{alertIdx},
				"verdict_support":      "benign",
			},
		})
		shouldArchive, err := shouldArchiveScenarioEvidence(scenario, seedTitles)
		if !shouldArchive || err == nil {
			t.Fatalf("expected invalid FP AI scenario to be archived, got shouldArchive=%v err=%v", shouldArchive, err)
		}
	})

	t.Run("archives legacy non-seed TP and FP rows containing CORR-9901 or routine boilerplate", func(t *testing.T) {
		for _, isTP := range []bool{true, false} {
			scenario := synthesizeFullScenario(evidenceTestMicro(isTP), "domain4", "Security Operations")
			scenario.Title = "Legacy CORR-9901 AI Scenario"
			logs := normalizePhaseStepSlice(scenario.InitialLogs)
			logs[5]["message"] = "SIEM Alert Rule [CORR-9901]: Multi-stage incident correlation triggered on host SRV-WEB01 involving IP 198.51.100.7"
			scenario.InitialLogs = logs
			shouldArchive, err := shouldArchiveScenarioEvidence(scenario, seedTitles)
			if !shouldArchive || err == nil {
				t.Fatalf("expected legacy CORR-9901 scenario (isTP=%v) to be archived, got shouldArchive=%v err=%v", isTP, shouldArchive, err)
			}
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

	t.Run("archives non-seed scenarios failing ValidateScenarioCandidate", func(t *testing.T) {
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
		shouldArchive, err := shouldArchiveScenarioEvidence(scenario, seedTitles)
		if !shouldArchive || err == nil {
			t.Fatalf("expected scenario failing ValidateScenarioCandidate to be archived, got shouldArchive=%v err=%v", shouldArchive, err)
		}
	})

	t.Run("archives active non-seed AI scenarios missing valid translations.th", func(t *testing.T) {
		scenario := synthesizeFullScenario(evidenceTestMicro(true), "domain4", "Security Operations")
		scenario.Title = "Legacy Unlocalized AI Scenario"
		scenario.Translations = nil
		shouldArchive, err := shouldArchiveScenarioEvidence(scenario, seedTitles)
		if !shouldArchive || err == nil {
			t.Fatalf("expected unlocalized non-seed AI scenario to be archived, got shouldArchive=%v err=%v", shouldArchive, err)
		}
	})
}

func TestHybridMultiEventInterleavedLogs_DiversityAndDecoys(t *testing.T) {
	ai := aiMicroScenario{
		Title:             "KMS Key Enumeration and S3 Data Staging",
		TitleTh:           "การไล่ตรวจสอบ KMS Key และพักข้อมูลบน S3 เพื่อขโมยออกนอกระบบ",
		Description:       "Unusual AWS KMS key listing followed by S3 object staging by a developer role.",
		DescriptionTh:     "ตรวจพบการเรียกดูรายการ AWS KMS key ผิดปกติตามด้วยการคัดลอกข้อมูลบน S3 โดยสิทธิ์นักพัฒนา",
		IsTruePositive:    true,
		TpFpExplanation:   "Unauthorized role assumption used to enumerate KMS keys and stage S3 archives.",
		TpFpExplanationTh: "มีการสวมสิทธิ์ role โดยไม่ได้รับอนุญาตเพื่อสำรวจ KMS key และพักไฟล์เตรียมขโมย",
		AttackerIP:        "198.51.100.84",
		TargetHost:        "aws-s3-finance",
		PrimaryUser:       "dev_ci_runner",
		IncidentLogs: []aiIncidentLog{
			{Source: "iam_service", Severity: "info", Host: "aws-iam-global", SrcIP: "198.51.100.84", User: "dev_ci_runner", Message: "AssumeRole succeeded for role prod-s3-auditor from external source 198.51.100.84"},
			{Source: "aws_cloudtrail", Severity: "warning", Host: "aws-kms-us-east", SrcIP: "198.51.100.84", User: "dev_ci_runner", Message: "ListAliases executed across 14 customer master keys by dev_ci_runner"},
			{Source: "s3_access", Severity: "high", Host: "aws-s3-finance", SrcIP: "198.51.100.84", User: "dev_ci_runner", Message: "GetObject batch downloaded 42 payroll archives from s3://corp-fin-staging"},
			{Source: "aws_cloudtrail", Severity: "critical", Host: "aws-s3-finance", SrcIP: "198.51.100.84", User: "dev_ci_runner", Message: "PutBucketReplication configured toward external account 998877665544"},
		},
		Findings: []aiFinding{
			{
				Description:         "External IAM role assumption and KMS enumeration",
				DescriptionTh:       "การสวมสิทธิ์ IAM Role จากภายนอกและการสำรวจ KMS Key",
				DomainID:            "domain3",
				Points:              10,
				EvidenceLogMessages: []string{"AssumeRole succeeded for role prod-s3-auditor from external source 198.51.100.84", "ListAliases executed across 14 customer master keys by dev_ci_runner"},
				VerdictSupport:      "malicious",
				Explanation:         "Reconnaissance of KMS keys followed unusual external role assumption.",
				ExplanationTh:       "มีการสำรวจรายการ KMS key ต่อเนื่องจากการ AssumeRole จาก IP ภายนอก",
			},
			{
				Description:         "Unauthorized S3 replication staging",
				DescriptionTh:       "การตั้งค่า S3 Replication โดยไม่ได้รับอนุญาตเพื่อพักข้อมูล",
				DomainID:            "domain3",
				Points:              10,
				EvidenceLogMessages: []string{"GetObject batch downloaded 42 payroll archives from s3://corp-fin-staging", "PutBucketReplication configured toward external account 998877665544"},
				VerdictSupport:      "malicious",
				Explanation:         "Bucket replication to an unrecognized account indicates staging for exfiltration.",
				ExplanationTh:       "การเปิดทำ replication ไปยังบัญชีภายนอกที่ไม่รู้จักบ่งชี้ถึงการเตรียมโจรกรรมข้อมูล",
			},
		},
	}

	scenario, errs := synthesizeFullScenarioWithErrors(ai, "domain3", "Security Architecture")
	if len(errs) > 0 {
		t.Fatalf("expected multi-event scenario to synthesize cleanly, got %v", errs)
	}
	if err := ValidateScenarioCandidate(scenario); err != nil {
		t.Fatalf("expected multi-event scenario to pass ValidateScenarioCandidate, got %v", err)
	}

	logs := normalizePhaseStepSlice(scenario.InitialLogs)
	if len(logs) < 28 || len(logs) > 46 {
		t.Fatalf("expected 28..46 logs, got %d", len(logs))
	}

	// Verify monotonic timestamps, <= 2 message repetitions, zero legacy boilerplate
	evidenceSet := make(map[int]bool)
	for _, f := range findingsForScenario(t, scenario) {
		for _, idx := range evidenceIndicesForFinding(t, f) {
			evidenceSet[idx] = true
		}
	}
	if len(evidenceSet) != 4 {
		t.Fatalf("expected 4 mapped evidence indices, got %d", len(evidenceSet))
	}
	var sortedEv []int
	for idx := range evidenceSet {
		sortedEv = append(sortedEv, idx)
	}
	// Sort to verify interleaving (non-contiguous)
	for i := 0; i < len(sortedEv); i++ {
		for j := i + 1; j < len(sortedEv); j++ {
			if sortedEv[j] < sortedEv[i] {
				sortedEv[i], sortedEv[j] = sortedEv[j], sortedEv[i]
			}
		}
	}
	if sortedEv[len(sortedEv)-1]-sortedEv[0]+1 == len(sortedEv) {
		t.Fatalf("expected interleaved non-contiguous evidence indices, got %v", sortedEv)
	}

	decoyCount := 0
	msgCounts := make(map[string]int)
	var prevTS time.Time
	for i, entry := range logs {
		tsStr, _ := entry["timestamp"].(string)
		ts, parseErr := time.Parse(time.RFC3339, tsStr)
		if parseErr != nil {
			t.Fatalf("log %d invalid RFC3339 timestamp %q: %v", i, tsStr, parseErr)
		}
		if i > 0 && !ts.After(prevTS) {
			t.Fatalf("log %d timestamp %s is not strictly after previous %s", i, tsStr, prevTS.Format(time.RFC3339))
		}
		prevTS = ts

		msg, _ := entry["message"].(string)
		norm := strings.ToLower(strings.TrimSpace(msg))
		msgCounts[norm]++
		if msgCounts[norm] > 2 {
			t.Fatalf("message %q repeated %d times (want <= 2)", msg, msgCounts[norm])
		}
		for _, banned := range BannedLegacyLogSubstrings {
			if strings.Contains(norm, banned) {
				t.Fatalf("found banned legacy boilerplate %q in %q", banned, msg)
			}
		}
		if !evidenceSet[i] && scenarioLogSeverityBucket(entry) == "medium" {
			decoyCount++
		}
	}
	if decoyCount < 2 || decoyCount > 4 {
		t.Fatalf("expected 2..4 non-evidence decoy warning/medium logs, got %d", decoyCount)
	}
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
