package services

import (
	"reflect"
	"testing"
	"time"

	"cybersim/models"
)

func reviewLog(message, source, severity string) map[string]interface{} {
	return map[string]interface{}{
		"timestamp": "2026-07-13T08:00:00Z",
		"source":    source,
		"severity":  severity,
		"message":   message,
	}
}

func reviewScenario(findings ...map[string]interface{}) models.Scenario {
	scenario := mockScenarioWithFindings(findings...)
	scenario.InitialLogs = []interface{}{
		reviewLog("benign login", "ad_controller", "info"),
		reviewLog("mfa push", "okta", "warning"),
		reviewLog("token reuse", "okta", "high"),
		reviewLog("dacl change", "ad_controller", "high"),
		reviewLog("outbound allow", "firewall", "info"),
		reviewLog("hr travel", "hr_portal", "info"),
	}
	return scenario
}

func TestBuildEvidenceReview_MatchedMissedSharedNoiseAndPenalty(t *testing.T) {
	scenario := reviewScenario(
		keyFinding("f1", "Push fatigue", "domain1", 10, []interface{}{1, 2}),
		keyFinding("f2", "Token reuse", "domain1", 10, []interface{}{2, 3}),
		keyFinding("f3", "Travel context", "domain1", 10, []interface{}{5}),
	)

	review := BuildEvidenceReview(scenario, []int{2, 5, 0})

	if len(review.EvidenceFindings) != 3 {
		t.Fatalf("EvidenceFindings = %d, want 3", len(review.EvidenceFindings))
	}

	f1 := review.EvidenceFindings[0]
	if !f1.Found || f1.Earned != 10 {
		t.Fatalf("f1 found=%v earned=%d, want found with 10", f1.Found, f1.Earned)
	}
	if !reflect.DeepEqual(f1.MatchedLogIndices, []int{2}) {
		t.Fatalf("f1 matched = %v, want [2]", f1.MatchedLogIndices)
	}
	if !reflect.DeepEqual(f1.EvidenceLogIndices, []int{1, 2}) {
		t.Fatalf("f1 evidence = %v, want [1 2]", f1.EvidenceLogIndices)
	}

	f2 := review.EvidenceFindings[1]
	if !f2.Found || f2.Earned != 10 {
		t.Fatalf("shared log must credit f2, got found=%v earned=%d", f2.Found, f2.Earned)
	}
	if !reflect.DeepEqual(f2.MatchedLogIndices, []int{2}) {
		t.Fatalf("f2 matched = %v, want [2]", f2.MatchedLogIndices)
	}

	f3 := review.EvidenceFindings[2]
	if !f3.Found || f3.Earned != 10 {
		t.Fatalf("f3 found=%v earned=%d, want found with 10", f3.Found, f3.Earned)
	}

	// Index 0 is outside every evidence set, so it is noise. Index 2 is shared
	// and must never be counted as noise.
	if !reflect.DeepEqual(review.SelectedLogIndices, []int{0, 2, 5}) {
		t.Fatalf("SelectedLogIndices = %v, want [0 2 5]", review.SelectedLogIndices)
	}
	if review.NoiseSelectedCount != 1 {
		t.Fatalf("NoiseSelectedCount = %d, want 1", review.NoiseSelectedCount)
	}
	if got := noiseIndices(review); !reflect.DeepEqual(got, []int{0}) {
		t.Fatalf("noise indices = %v, want [0]", got)
	}
	if review.EvidenceSelectedCount != 3 {
		t.Fatalf("EvidenceSelectedCount = %d, want 3", review.EvidenceSelectedCount)
	}
	if review.FindingsFoundCount != 3 || review.TotalFindingsCount != 3 {
		t.Fatalf("findings = %d/%d, want 3/3", review.FindingsFoundCount, review.TotalFindingsCount)
	}
	if review.GrossFindingsPoints != 30 {
		t.Fatalf("GrossFindingsPoints = %f, want 30", review.GrossFindingsPoints)
	}
	if review.EvidencePenalty != 2 {
		t.Fatalf("EvidencePenalty = %f, want 2", review.EvidencePenalty)
	}
	if review.FindingsPoints != 28 {
		t.Fatalf("FindingsPoints = %f, want 28", review.FindingsPoints)
	}
}

func TestBuildEvidenceReview_MissedFindingsExposeAcceptableEvidence(t *testing.T) {
	scenario := reviewScenario(
		keyFinding("f1", "Push fatigue", "domain1", 10, []interface{}{1, 2}),
		keyFinding("f2", "Travel context", "domain1", 10, []interface{}{5}),
	)

	review := BuildEvidenceReview(scenario, []int{4})

	if review.FindingsFoundCount != 0 {
		t.Fatalf("FindingsFoundCount = %d, want 0", review.FindingsFoundCount)
	}
	if review.GrossFindingsPoints != 0 || review.FindingsPoints != 0 {
		t.Fatalf("findings points gross=%f net=%f, want 0", review.GrossFindingsPoints, review.FindingsPoints)
	}
	if review.EvidencePenalty != 2 {
		t.Fatalf("EvidencePenalty = %f, want 2", review.EvidencePenalty)
	}
	if review.NoiseSelectedCount != 1 {
		t.Fatalf("NoiseSelectedCount = %d, want 1", review.NoiseSelectedCount)
	}

	for _, finding := range review.EvidenceFindings {
		if finding.Found || finding.Earned != 0 {
			t.Fatalf("finding %s should be missed with 0 earned, got found=%v earned=%d", finding.ID, finding.Found, finding.Earned)
		}
		if len(finding.EvidenceLogs) == 0 {
			t.Fatalf("missed finding %s must expose acceptable evidence snippets", finding.ID)
		}
	}
	if len(review.NoiseSelections) != 1 || review.NoiseSelections[0].Index != 4 {
		t.Fatalf("noise selections = %+v, want index 4", review.NoiseSelections)
	}
	if review.NoiseSelections[0].Message != "outbound allow" {
		t.Fatalf("noise snippet message = %q, want outbound allow", review.NoiseSelections[0].Message)
	}
}

func TestBuildEvidenceReview_EmptySelectionAndDuplicateIndices(t *testing.T) {
	scenario := reviewScenario(
		keyFinding("f1", "Push fatigue", "domain1", 10, []interface{}{1}),
	)

	empty := BuildEvidenceReview(scenario, nil)
	if empty.EvidenceSelectedCount != 0 || empty.NoiseSelectedCount != 0 || empty.EvidencePenalty != 0 {
		t.Fatalf("empty selection counts = selected %d noise %d penalty %f, want 0/0/0",
			empty.EvidenceSelectedCount, empty.NoiseSelectedCount, empty.EvidencePenalty)
	}
	if empty.FindingsFoundCount != 0 || empty.FindingsPoints != 0 {
		t.Fatalf("empty selection findings = %d/%f, want 0/0", empty.FindingsFoundCount, empty.FindingsPoints)
	}

	duplicated := BuildEvidenceReview(scenario, []int{1, 1, 1})
	if duplicated.EvidenceSelectedCount != 1 || duplicated.FindingsFoundCount != 1 {
		t.Fatalf("duplicate indices must be de-duplicated, got selected=%d found=%d",
			duplicated.EvidenceSelectedCount, duplicated.FindingsFoundCount)
	}
	if duplicated.EvidencePenalty != 0 {
		t.Fatalf("duplicate evidence must not create a penalty, got %f", duplicated.EvidencePenalty)
	}
}

func TestBuildEvidenceReview_FloorAtZero(t *testing.T) {
	scenario := reviewScenario(
		keyFinding("f1", "Push fatigue", "domain1", 5, []interface{}{1}),
	)

	review := BuildEvidenceReview(scenario, []int{0, 2, 3, 4, 5})

	if review.GrossFindingsPoints != 0 {
		t.Fatalf("GrossFindingsPoints = %f, want 0", review.GrossFindingsPoints)
	}
	if review.EvidencePenalty != 10 {
		t.Fatalf("EvidencePenalty = %f, want 10", review.EvidencePenalty)
	}
	if review.FindingsPoints != 0 {
		t.Fatalf("FindingsPoints = %f, want floor 0", review.FindingsPoints)
	}
}

func TestBuildEvidenceReview_SnippetFieldsAndOutOfRange(t *testing.T) {
	scenario := reviewScenario(
		keyFinding("f1", "Push fatigue", "domain1", 10, []interface{}{1}),
	)

	review := BuildEvidenceReview(scenario, []int{1, 99})

	if len(review.EvidenceFindings[0].EvidenceLogs) != 1 {
		t.Fatalf("acceptable evidence snippets = %d, want 1", len(review.EvidenceFindings[0].EvidenceLogs))
	}
	snippet := review.EvidenceFindings[0].EvidenceLogs[0]
	if snippet.Index != 1 || snippet.Timestamp != "2026-07-13T08:00:00Z" || snippet.Source != "okta" ||
		snippet.Severity != "warning" || snippet.Message != "mfa push" {
		t.Fatalf("unexpected evidence snippet: %+v", snippet)
	}

	if review.NoiseSelectedCount != 1 {
		t.Fatalf("NoiseSelectedCount = %d, want 1", review.NoiseSelectedCount)
	}
	if len(review.NoiseSelections) != 1 || review.NoiseSelections[0].Index != 99 {
		t.Fatalf("out-of-range noise selection = %+v, want index 99 with empty snippet", review.NoiseSelections)
	}
	if review.NoiseSelections[0].Message != "" {
		t.Fatalf("out-of-range snippet message = %q, want empty", review.NoiseSelections[0].Message)
	}
}

func TestBuildEvidenceReviewFromActions_LegacyFallbackIsNotRescored(t *testing.T) {
	scenario := reviewScenario(
		keyFinding("f1", "Push fatigue", "domain1", 10, []interface{}{1}),
		keyFinding("f2", "Travel context", "domain1", 10, []interface{}{5}),
	)
	actions := []models.SessionAction{
		{ActionType: "triage_alert", StepOrder: 1, Payload: map[string]interface{}{"user_choice": true}, Timestamp: time.Unix(0, 0)},
		{ActionType: "submit_decision", StepOrder: 2, Payload: map[string]interface{}{
			"discovered_findings": []interface{}{"f1"},
		}, Timestamp: time.Unix(0, 0)},
	}

	review := BuildEvidenceReviewFromActions(scenario, actions)

	if !review.Legacy {
		t.Fatal("expected legacy review when no select_evidence actions exist")
	}
	if !reflect.DeepEqual(review.LegacyDiscoveredFindings, []string{"f1"}) {
		t.Fatalf("legacy findings = %v, want [f1]", review.LegacyDiscoveredFindings)
	}
	if review.FindingsFoundCount != 1 || review.TotalFindingsCount != 2 {
		t.Fatalf("legacy findings = %d/%d, want 1/2", review.FindingsFoundCount, review.TotalFindingsCount)
	}
	if review.EvidenceSelectedCount != 0 || review.NoiseSelectedCount != 0 || review.EvidencePenalty != 0 {
		t.Fatalf("legacy sessions must not synthesize evidence counters, got selected=%d noise=%d penalty=%f",
			review.EvidenceSelectedCount, review.NoiseSelectedCount, review.EvidencePenalty)
	}
	if review.FindingsPoints != 0 || review.GrossFindingsPoints != 0 {
		t.Fatalf("legacy sessions must not be re-scored, got gross=%f net=%f", review.GrossFindingsPoints, review.FindingsPoints)
	}
	if len(review.EvidenceFindings) != 0 || len(review.NoiseSelections) != 0 || len(review.SelectedLogIndices) != 0 {
		t.Fatalf("legacy review must not expose evidence findings, got %+v", review)
	}
}

func TestBuildEvidenceReviewFromActions_LegacyListOutcomesStillCounted(t *testing.T) {
	scenario := models.Scenario{
		ScenarioID: "scen-legacy-list",
		ExpectedOutcomes: []interface{}{
			map[string]interface{}{"id": "f1"},
			map[string]interface{}{"id": "f2"},
		},
	}
	actions := []models.SessionAction{
		{ActionType: "submit_decision", StepOrder: 1, Payload: map[string]interface{}{
			"discovered_findings": []interface{}{"f1"},
		}},
	}

	review := BuildEvidenceReviewFromActions(scenario, actions)

	if !review.Legacy {
		t.Fatal("expected legacy review")
	}
	if review.FindingsFoundCount != 1 || review.TotalFindingsCount != 2 {
		t.Fatalf("legacy findings = %d/%d, want 1/2", review.FindingsFoundCount, review.TotalFindingsCount)
	}
	if review.FindingsPoints != 0 {
		t.Fatalf("legacy review must not re-score, got %f", review.FindingsPoints)
	}
}

func TestBuildEvidenceReviewFromActions_UsesServerSelection(t *testing.T) {
	scenario := reviewScenario(
		keyFinding("f1", "Push fatigue", "domain1", 10, []interface{}{1, 2}),
	)
	actions := []models.SessionAction{
		selectionAction(SelectEvidenceActionType, map[string]any{"operation": "toggle", "log_index": 2, "selected": true}, 1),
		selectionAction(SelectEvidenceActionType, map[string]any{"operation": "toggle", "log_index": 0, "selected": true}, 2),
		selectionAction(SelectEvidenceActionType, map[string]any{"operation": "clear"}, 3),
		selectionAction(SelectEvidenceActionType, map[string]any{"operation": "toggle", "log_index": 1, "selected": true}, 4),
		selectionAction("submit_decision", map[string]any{"discovered_findings": []interface{}{"f1"}}, 5),
	}

	review := BuildEvidenceReviewFromActions(scenario, actions)

	if review.Legacy {
		t.Fatal("selection actions must take precedence over the legacy payload")
	}
	if !reflect.DeepEqual(review.SelectedLogIndices, []int{1}) {
		t.Fatalf("SelectedLogIndices = %v, want [1]", review.SelectedLogIndices)
	}
	if review.FindingsFoundCount != 1 || review.NoiseSelectedCount != 0 || review.EvidencePenalty != 0 {
		t.Fatalf("review = found %d noise %d penalty %f, want 1/0/0",
			review.FindingsFoundCount, review.NoiseSelectedCount, review.EvidencePenalty)
	}
	if review.FindingsPoints != 10 {
		t.Fatalf("FindingsPoints = %f, want 10", review.FindingsPoints)
	}
}

func TestBuildEvidenceReviewFromActions_ZeroSelectionNewRubric(t *testing.T) {
	scenario := reviewScenario(
		keyFinding("f1", "Push fatigue", "domain1", 10, []interface{}{1}),
		keyFinding("f2", "Travel context", "domain1", 10, []interface{}{5}),
	)
	actions := []models.SessionAction{
		{ActionType: "submit_decision", StepOrder: 1, Payload: map[string]interface{}{
			"selected_log_indices": []interface{}{},
		}},
	}

	review := BuildEvidenceReviewFromActions(scenario, actions)

	if review.Legacy {
		t.Fatal("a v2 session that submitted an empty selection is not legacy")
	}
	if !reflect.DeepEqual(review.SelectedLogIndices, []int{}) {
		t.Fatalf("SelectedLogIndices = %v, want []", review.SelectedLogIndices)
	}
	if review.EvidenceSelectedCount != 0 || review.NoiseSelectedCount != 0 || review.EvidencePenalty != 0 {
		t.Fatalf("zero-selection counts = selected %d noise %d penalty %f, want 0/0/0",
			review.EvidenceSelectedCount, review.NoiseSelectedCount, review.EvidencePenalty)
	}
	if review.FindingsFoundCount != 0 || review.TotalFindingsCount != 2 {
		t.Fatalf("findings = %d/%d, want 0/2", review.FindingsFoundCount, review.TotalFindingsCount)
	}
	if review.FindingsPoints != 0 || review.GrossFindingsPoints != 0 {
		t.Fatalf("findings points gross=%f net=%f, want 0/0", review.GrossFindingsPoints, review.FindingsPoints)
	}
	if len(review.EvidenceFindings) != 2 {
		t.Fatalf("EvidenceFindings = %d, want the two missed findings", len(review.EvidenceFindings))
	}
}

func TestBuildEvidenceReviewForSession_UsesRubricVersionForZeroSelection(t *testing.T) {
	scenario := reviewScenario(
		keyFinding("f1", "Push fatigue", "domain1", 10, []interface{}{1}),
	)
	session := models.SimulationSession{RubricVersion: RubricVersion}
	actions := []models.SessionAction{
		{ActionType: "submit_decision", StepOrder: 1, Payload: map[string]interface{}{}},
	}

	review := BuildEvidenceReviewForSession(scenario, session, actions)

	if review.Legacy {
		t.Fatal("a session recorded with the v2 rubric must not be treated as legacy")
	}
	if review.EvidenceSelectedCount != 0 || review.NoiseSelectedCount != 0 || review.EvidencePenalty != 0 {
		t.Fatalf("zero-selection counts = selected %d noise %d penalty %f, want 0/0/0",
			review.EvidenceSelectedCount, review.NoiseSelectedCount, review.EvidencePenalty)
	}
	if review.FindingsFoundCount != 0 || review.TotalFindingsCount != 1 {
		t.Fatalf("findings = %d/%d, want 0/1", review.FindingsFoundCount, review.TotalFindingsCount)
	}
}

func TestEvidenceReviewFromResult_UsesRecordedCounters(t *testing.T) {
	result := models.ScenarioResult{
		SelectedLogIndices:  []int{1, 4},
		NoiseLogIndices:     []int{4},
		GrossFindingsPoints: 10,
		FindingsPoints:      10,
		EvidencePenalty:     2,
		FindingsDetail: []models.FindingResult{
			{ID: "f1", Found: true, Points: 10, Earned: 10, EvidenceLogIndices: []int{1}, MatchedLogIndices: []int{1}},
			{ID: "f2", Found: false, Points: 10, Earned: 0, EvidenceLogIndices: []int{5}},
		},
	}

	review := EvidenceReviewFromResult(result)

	if review.FindingsFoundCount != 1 || review.TotalFindingsCount != 2 {
		t.Fatalf("findings = %d/%d, want 1/2", review.FindingsFoundCount, review.TotalFindingsCount)
	}
	if review.EvidenceSelectedCount != 2 || review.NoiseSelectedCount != 1 {
		t.Fatalf("selection counts = %d/%d, want 2/1", review.EvidenceSelectedCount, review.NoiseSelectedCount)
	}
	if review.GrossFindingsPoints != 10 || review.FindingsPoints != 10 || review.EvidencePenalty != 2 {
		t.Fatalf("points = gross %f net %f penalty %f, want 10/10/2",
			review.GrossFindingsPoints, review.FindingsPoints, review.EvidencePenalty)
	}
}

func noiseIndices(review EvidenceReview) []int {
	indices := make([]int, 0, len(review.NoiseSelections))
	for _, selection := range review.NoiseSelections {
		indices = append(indices, selection.Index)
	}
	return indices
}
