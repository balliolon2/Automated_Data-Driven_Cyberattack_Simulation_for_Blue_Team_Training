package services

import (
	"reflect"
	"testing"

	"cybersim/models"
)

func keyFinding(id, description, domainID string, points int, evidence any) map[string]interface{} {
	return map[string]interface{}{
		"id":                   id,
		"description":          description,
		"domain_id":            domainID,
		"points":               points,
		"evidence_log_indices": evidence,
		"explanation":          description + " explanation.",
	}
}

func mockScenarioWithFindings(findings ...map[string]interface{}) models.Scenario {
	keyFindings := make([]interface{}, 0, len(findings))
	for _, finding := range findings {
		keyFindings = append(keyFindings, finding)
	}
	return models.Scenario{
		ScenarioID:      "scen-test-01",
		Title:           "Test Incident Scenario",
		DomainID:        "domain4",
		IsTruePositive:  true,
		TpFpExplanation: "Malicious C2 beaconing confirmed via threat intel matching destination IP.",
		ExpectedOutcomes: map[string]interface{}{
			"key_findings": keyFindings,
		},
	}
}

func createMockScenario() models.Scenario {
	scenario := mockScenarioWithFindings(
		keyFinding("f1", "C2 beaconing detected", "domain4", 10, []interface{}{15, 16}),
		keyFinding("f2", "Privilege escalation attempt", "domain2", 10, []interface{}{17}),
	)
	scenario.PlaybookSteps = map[string]interface{}{
		"containment": []interface{}{
			map[string]interface{}{
				"id":          "c1",
				"label":       "Isolate infected host from subnet",
				"is_correct":  true,
				"domain_id":   "domain4",
				"points":      5.0,
				"explanation": "Stops lateral movement immediately.",
			},
			map[string]interface{}{
				"id":          "c2",
				"label":       "Reboot core router",
				"is_correct":  false,
				"domain_id":   "domain3",
				"points":      -3.0,
				"explanation": "Causes broad outage without isolating host.",
			},
		},
		"eradication": []interface{}{
			map[string]interface{}{
				"id":          "e1",
				"label":       "Terminate malicious PowerShell process",
				"is_correct":  true,
				"domain_id":   "domain2",
				"points":      5.0,
				"explanation": "Removes active malware process.",
			},
		},
		"recovery": []interface{}{
			map[string]interface{}{
				"id":          "r1",
				"label":       "Restore host from verified clean backup",
				"is_correct":  true,
				"domain_id":   "domain4",
				"points":      5.0,
				"explanation": "Restores operations securely.",
			},
		},
	}
	return scenario
}

func TestScoreSubmission_PerfectScore(t *testing.T) {
	scenario := createMockScenario()
	input := models.SubmitScenarioInput{
		SessionID:          "sess-01",
		IsTruePositive:     true,
		TPFPSelected:       true,
		SelectedActions:    []string{"c1", "e1", "r1"},
		SelectedLogIndices: []int{15, 17},
	}

	result := ScoreSubmission(scenario, input)

	if !result.TPFPCorrect || result.TPFPPoints != 20.0 {
		t.Fatalf("expected correct TP/FP with 20 points, got correct=%v points=%f", result.TPFPCorrect, result.TPFPPoints)
	}
	if result.FindingsPoints != 20.0 {
		t.Fatalf("expected FindingsPoints 20.0, got %f", result.FindingsPoints)
	}
	if result.ResponsePoints != 15.0 {
		t.Fatalf("expected ResponsePoints 15.0, got %f", result.ResponsePoints)
	}
	if result.EvidencePenalty != 0 || len(result.NoiseLogIndices) != 0 {
		t.Fatalf("expected no evidence penalty, got penalty=%f noise=%v", result.EvidencePenalty, result.NoiseLogIndices)
	}
	if !reflect.DeepEqual(result.SelectedLogIndices, []int{15, 17}) {
		t.Fatalf("SelectedLogIndices = %v, want [15 17]", result.SelectedLogIndices)
	}
	if result.TotalScore != 100.0 {
		t.Fatalf("expected TotalScore 100.0, got %f", result.TotalScore)
	}

	for _, finding := range result.FindingsDetail {
		if !finding.Found || finding.Earned != finding.Points {
			t.Fatalf("expected finding %s found with full credit, got found=%v earned=%d points=%d", finding.ID, finding.Found, finding.Earned, finding.Points)
		}
	}
}

func TestScoreSubmission_WrongTPFP(t *testing.T) {
	scenario := createMockScenario()
	result := ScoreSubmission(scenario, models.SubmitScenarioInput{
		SessionID:          "sess-02",
		IsTruePositive:     false,
		TPFPSelected:       true,
		SelectedActions:    []string{"c1", "e1", "r1"},
		SelectedLogIndices: []int{15, 17},
	})

	if result.TPFPCorrect || result.TPFPPoints != 0.0 {
		t.Fatalf("expected incorrect TP/FP with 0 points, got correct=%v points=%f", result.TPFPCorrect, result.TPFPPoints)
	}
	// Total max: 20 (TP/FP) + 20 (Findings) + 15 (Responses) = 55.
	// Total earned: 0 + 20 + 15 = 35 -> 63.64%.
	if result.TotalScore != 63.64 {
		t.Fatalf("expected TotalScore 63.64, got %f", result.TotalScore)
	}
}

func TestScoreSubmission_DistractorPenalty(t *testing.T) {
	scenario := createMockScenario()
	result := ScoreSubmission(scenario, models.SubmitScenarioInput{
		SessionID:          "sess-03",
		IsTruePositive:     true,
		TPFPSelected:       true,
		SelectedActions:    []string{"c1", "c2", "e1", "r1"},
		SelectedLogIndices: []int{15, 17},
	})

	if result.ResponsePoints != 12.0 {
		t.Fatalf("expected ResponsePoints 12.0, got %f", result.ResponsePoints)
	}

	var c2Detail *models.ResponseActionResult
	for i := range result.ResponseDetail {
		if result.ResponseDetail[i].ID == "c2" {
			c2Detail = &result.ResponseDetail[i]
			break
		}
	}
	if c2Detail == nil {
		t.Fatal("c2 response detail not found")
	}
	if !c2Detail.Selected || c2Detail.IsCorrect || c2Detail.Earned != -3 {
		t.Fatalf("c2 detail mismatch: Selected=%v, IsCorrect=%v, Earned=%d", c2Detail.Selected, c2Detail.IsCorrect, c2Detail.Earned)
	}
}

func TestScoreSubmission_EvidenceRules(t *testing.T) {
	scenario := mockScenarioWithFindings(
		keyFinding("f1", "C2 beaconing detected", "domain4", 10, []interface{}{15, 16}),
		keyFinding("f2", "Privilege escalation attempt", "domain2", 10, []interface{}{17}),
	)

	tests := []struct {
		name           string
		selected       []int
		wantFindings   float64
		wantPenalty    float64
		wantNoise      []int
		wantSelected   []int
		wantTotalScore float64
		wantMatchedF1  []int
		wantMatchedF2  []int
		wantFoundF1    bool
		wantFoundF2    bool
	}{
		{
			name:           "one evidence log credits the full finding",
			selected:       []int{15},
			wantFindings:   10,
			wantPenalty:    0,
			wantNoise:      []int{},
			wantSelected:   []int{15},
			wantTotalScore: 75,
			wantMatchedF1:  []int{15},
			wantMatchedF2:  []int{},
			wantFoundF1:    true,
			wantFoundF2:    false,
		},
		{
			name:           "no evidence credits zero",
			selected:       nil,
			wantFindings:   0,
			wantPenalty:    0,
			wantNoise:      []int{},
			wantSelected:   []int{},
			wantTotalScore: 50,
			wantMatchedF1:  []int{},
			wantMatchedF2:  []int{},
			wantFoundF1:    false,
			wantFoundF2:    false,
		},
		{
			name:           "duplicate evidence from the same finding does not stack",
			selected:       []int{15, 16},
			wantFindings:   10,
			wantPenalty:    0,
			wantNoise:      []int{},
			wantSelected:   []int{15, 16},
			wantTotalScore: 75,
			wantMatchedF1:  []int{15, 16},
			wantMatchedF2:  []int{},
			wantFoundF1:    true,
			wantFoundF2:    false,
		},
		{
			name:           "one noise log subtracts two points",
			selected:       []int{15, 99},
			wantFindings:   8,
			wantPenalty:    2,
			wantNoise:      []int{99},
			wantSelected:   []int{15, 99},
			wantTotalScore: 70,
			wantMatchedF1:  []int{15},
			wantMatchedF2:  []int{},
			wantFoundF1:    true,
			wantFoundF2:    false,
		},
		{
			name:           "noise floor prevents negative findings points",
			selected:       []int{99, 100, 101},
			wantFindings:   0,
			wantPenalty:    6,
			wantNoise:      []int{99, 100, 101},
			wantSelected:   []int{99, 100, 101},
			wantTotalScore: 50,
			wantMatchedF1:  []int{},
			wantMatchedF2:  []int{},
			wantFoundF1:    false,
			wantFoundF2:    false,
		},
		{
			name:           "penalty can reduce but not flatten positive credit",
			selected:       []int{15, 99, 100, 101},
			wantFindings:   4,
			wantPenalty:    6,
			wantNoise:      []int{99, 100, 101},
			wantSelected:   []int{15, 99, 100, 101},
			wantTotalScore: 60,
			wantMatchedF1:  []int{15},
			wantMatchedF2:  []int{},
			wantFoundF1:    true,
			wantFoundF2:    false,
		},
		{
			name:           "duplicate selected indices are counted once",
			selected:       []int{15, 15, 99, 99},
			wantFindings:   8,
			wantPenalty:    2,
			wantNoise:      []int{99},
			wantSelected:   []int{15, 99},
			wantTotalScore: 70,
			wantMatchedF1:  []int{15},
			wantMatchedF2:  []int{},
			wantFoundF1:    true,
			wantFoundF2:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScoreSubmission(scenario, models.SubmitScenarioInput{
				SessionID:          "sess-evidence",
				IsTruePositive:     true,
				TPFPSelected:       true,
				SelectedLogIndices: tt.selected,
			})

			if result.FindingsPoints != tt.wantFindings {
				t.Errorf("FindingsPoints = %f, want %f", result.FindingsPoints, tt.wantFindings)
			}
			if result.EvidencePenalty != tt.wantPenalty {
				t.Errorf("EvidencePenalty = %f, want %f", result.EvidencePenalty, tt.wantPenalty)
			}
			if !reflect.DeepEqual(result.NoiseLogIndices, tt.wantNoise) {
				t.Errorf("NoiseLogIndices = %v, want %v", result.NoiseLogIndices, tt.wantNoise)
			}
			if !reflect.DeepEqual(result.SelectedLogIndices, tt.wantSelected) {
				t.Errorf("SelectedLogIndices = %v, want %v", result.SelectedLogIndices, tt.wantSelected)
			}
			if result.TotalScore != tt.wantTotalScore {
				t.Errorf("TotalScore = %f, want %f", result.TotalScore, tt.wantTotalScore)
			}

			f1 := result.FindingsDetail[0]
			f2 := result.FindingsDetail[1]
			if !reflect.DeepEqual(f1.MatchedLogIndices, tt.wantMatchedF1) {
				t.Errorf("f1 MatchedLogIndices = %v, want %v", f1.MatchedLogIndices, tt.wantMatchedF1)
			}
			if !reflect.DeepEqual(f2.MatchedLogIndices, tt.wantMatchedF2) {
				t.Errorf("f2 MatchedLogIndices = %v, want %v", f2.MatchedLogIndices, tt.wantMatchedF2)
			}
			if f1.Found != tt.wantFoundF1 || f2.Found != tt.wantFoundF2 {
				t.Errorf("found flags = (%v, %v), want (%v, %v)", f1.Found, f2.Found, tt.wantFoundF1, tt.wantFoundF2)
			}
			if f1.Found && f1.Earned != f1.Points {
				t.Errorf("found finding earned %d, want full %d", f1.Earned, f1.Points)
			}
			if !f1.Found && f1.Earned != 0 {
				t.Errorf("missed finding earned %d, want 0", f1.Earned)
			}
		})
	}
}

func TestScoreSubmission_SharedEvidenceCreditsBothFindings(t *testing.T) {
	scenario := mockScenarioWithFindings(
		keyFinding("f1", "Shared indicator one", "domain4", 10, []interface{}{15, 16}),
		keyFinding("f2", "Shared indicator two", "domain2", 10, []interface{}{16, 17}),
	)

	result := ScoreSubmission(scenario, models.SubmitScenarioInput{
		SessionID:          "sess-shared",
		IsTruePositive:     true,
		TPFPSelected:       true,
		SelectedLogIndices: []int{16},
	})

	if result.FindingsPoints != 20 {
		t.Fatalf("FindingsPoints = %f, want 20", result.FindingsPoints)
	}
	if result.EvidencePenalty != 0 || len(result.NoiseLogIndices) != 0 {
		t.Fatalf("shared evidence must not be noise, got penalty=%f noise=%v", result.EvidencePenalty, result.NoiseLogIndices)
	}
	for _, finding := range result.FindingsDetail {
		if !finding.Found || !reflect.DeepEqual(finding.MatchedLogIndices, []int{16}) {
			t.Fatalf("finding %s = found %v matched %v, want found with [16]", finding.ID, finding.Found, finding.MatchedLogIndices)
		}
	}
}

func TestScoreSubmission_DomainAttribution(t *testing.T) {
	scenario := mockScenarioWithFindings(
		keyFinding("f1", "Domain four finding", "domain4", 10, []interface{}{15}),
		keyFinding("f2", "Domain two finding", "domain2", 10, []interface{}{17}),
	)

	result := ScoreSubmission(scenario, models.SubmitScenarioInput{
		SessionID:          "sess-domain",
		IsTruePositive:     true,
		TPFPSelected:       true,
		SelectedLogIndices: []int{15, 17, 99},
	})

	if result.FindingsPoints != 18 {
		t.Fatalf("FindingsPoints = %f, want 18", result.FindingsPoints)
	}
	if result.EvidencePenalty != 2 {
		t.Fatalf("EvidencePenalty = %f, want 2", result.EvidencePenalty)
	}

	domain4 := result.DomainScores["domain4"]
	if domain4.Earned != 28 || domain4.MaxPoints != 30 {
		t.Fatalf("domain4 = %+v, want earned 28 / max 30 (TP/FP 20 + finding 10 - penalty 2)", domain4)
	}
	if domain4.Percentage != 93.33 {
		t.Fatalf("domain4 percentage = %f, want 93.33", domain4.Percentage)
	}
	domain2 := result.DomainScores["domain2"]
	if domain2.Earned != 10 || domain2.MaxPoints != 10 || domain2.Percentage != 100 {
		t.Fatalf("domain2 = %+v, want earned 10 / max 10 / 100%%", domain2)
	}
	// Total earned 38 of 40: TP/FP 20 + findings 18.
	if result.TotalScore != 95 {
		t.Fatalf("TotalScore = %f, want 95", result.TotalScore)
	}
}

func TestScoreSubmission_TotalScoreClampsToRange(t *testing.T) {
	scenario := mockScenarioWithFindings(
		keyFinding("f1", "Finding", "domain4", 10, []interface{}{15}),
	)
	scenario.PlaybookSteps = map[string]interface{}{
		"containment": []interface{}{
			map[string]interface{}{
				"id":          "bad",
				"label":       "Destructive action",
				"is_correct":  false,
				"domain_id":   "domain4",
				"points":      -100.0,
				"explanation": "Very harmful.",
			},
		},
	}

	result := ScoreSubmission(scenario, models.SubmitScenarioInput{
		SessionID:          "sess-clamp",
		IsTruePositive:     false,
		TPFPSelected:       true,
		SelectedActions:    []string{"bad"},
		SelectedLogIndices: []int{99},
	})

	if result.TotalScore != 0 {
		t.Fatalf("TotalScore = %f, want clamped to 0", result.TotalScore)
	}
	if result.DomainScores["domain4"].Earned != 0 {
		t.Fatalf("domain4 earned = %f, want floored at 0", result.DomainScores["domain4"].Earned)
	}
}

func TestScoreSubmission_DecodesEvidenceDefensively(t *testing.T) {
	tests := []struct {
		name         string
		evidence     any
		selected     []int
		wantEvidence []int
		wantFound    bool
		wantMatched  []int
	}{
		{
			name:         "accepts []int",
			evidence:     []int{15, 16},
			selected:     []int{15},
			wantEvidence: []int{15, 16},
			wantFound:    true,
			wantMatched:  []int{15},
		},
		{
			name:         "accepts []float64 from JSONB",
			evidence:     []float64{15, 16},
			selected:     []int{16},
			wantEvidence: []int{15, 16},
			wantFound:    true,
			wantMatched:  []int{16},
		},
		{
			name:         "accepts []interface with integral floats",
			evidence:     []interface{}{float64(15), float64(16)},
			selected:     []int{15},
			wantEvidence: []int{15, 16},
			wantFound:    true,
			wantMatched:  []int{15},
		},
		{
			name:         "skips fractional and non-numeric entries without truncating",
			evidence:     []interface{}{15.5, "16", float64(17), 18, -2},
			selected:     []int{17},
			wantEvidence: []int{17, 18},
			wantFound:    true,
			wantMatched:  []int{17},
		},
		{
			name:         "sorts and de-duplicates authored evidence",
			evidence:     []interface{}{18, 15, 18, float64(15)},
			selected:     []int{18},
			wantEvidence: []int{15, 18},
			wantFound:    true,
			wantMatched:  []int{18},
		},
		{
			name:         "unsupported evidence type credits nothing",
			evidence:     "15,16",
			selected:     []int{15},
			wantEvidence: []int{},
			wantFound:    false,
			wantMatched:  []int{},
		},
		{
			name:         "nil evidence credits nothing",
			evidence:     nil,
			selected:     []int{15},
			wantEvidence: []int{},
			wantFound:    false,
			wantMatched:  []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenario := mockScenarioWithFindings(
				keyFinding("f1", "Decoded finding", "domain4", 10, tt.evidence),
			)
			result := ScoreSubmission(scenario, models.SubmitScenarioInput{
				SessionID:          "sess-decode",
				IsTruePositive:     true,
				TPFPSelected:       true,
				SelectedLogIndices: tt.selected,
			})

			finding := result.FindingsDetail[0]
			if !reflect.DeepEqual(finding.EvidenceLogIndices, tt.wantEvidence) {
				t.Errorf("EvidenceLogIndices = %v, want %v", finding.EvidenceLogIndices, tt.wantEvidence)
			}
			if !reflect.DeepEqual(finding.MatchedLogIndices, tt.wantMatched) {
				t.Errorf("MatchedLogIndices = %v, want %v", finding.MatchedLogIndices, tt.wantMatched)
			}
			if finding.Found != tt.wantFound || finding.Earned > 0 != tt.wantFound {
				t.Errorf("found = %v earned = %d, want found=%v", finding.Found, finding.Earned, tt.wantFound)
			}
		})
	}
}

func TestScoreSubmission_DefaultsMissingPoints(t *testing.T) {
	finding := keyFinding("f1", "No points finding", "domain4", 0, []interface{}{15})
	delete(finding, "points")
	scenario := mockScenarioWithFindings(finding)

	result := ScoreSubmission(scenario, models.SubmitScenarioInput{
		SessionID:          "sess-default-points",
		IsTruePositive:     true,
		TPFPSelected:       true,
		SelectedLogIndices: []int{15},
	})

	if result.FindingsDetail[0].Points != DefaultFindingPoints {
		t.Fatalf("Points = %d, want default %d", result.FindingsDetail[0].Points, DefaultFindingPoints)
	}
	if result.FindingsPoints != DefaultFindingPoints {
		t.Fatalf("FindingsPoints = %f, want %d", result.FindingsPoints, DefaultFindingPoints)
	}
}

func TestReconstructFromActions(t *testing.T) {
	scenario := createMockScenario()
	actions := []models.SessionAction{
		{
			ActionType: "triage_alert",
			Payload: map[string]interface{}{
				"user_choice": true,
			},
		},
		{
			ActionType: "submit_decision",
			Payload: map[string]interface{}{
				"discovered_findings": []interface{}{"f1", "f2"},
			},
		},
		{
			ActionType: "respond",
			Payload: map[string]interface{}{
				"selected_actions": []interface{}{"c1", "e1", "r1"},
			},
		},
	}

	reconstructed := ReconstructFromActions(scenario, actions)

	if !reconstructed.TPFPCorrect {
		t.Fatal("expected reconstructed TPFPCorrect to be true")
	}
	if !reflect.DeepEqual(reconstructed.SelectedLogIndices, []int{15, 17}) {
		t.Fatalf("legacy discovered findings should map to first evidence indices [15 17], got %v", reconstructed.SelectedLogIndices)
	}
	if reconstructed.EvidencePenalty != 0 || len(reconstructed.NoiseLogIndices) != 0 {
		t.Fatalf("legacy finding IDs must not produce noise, got penalty=%f noise=%v", reconstructed.EvidencePenalty, reconstructed.NoiseLogIndices)
	}
	if reconstructed.TotalScore != 100.0 {
		t.Fatalf("expected reconstructed TotalScore 100.0, got %f", reconstructed.TotalScore)
	}
}

func TestReconstructFromActions_PrefersSelectionActions(t *testing.T) {
	scenario := createMockScenario()
	actions := []models.SessionAction{
		{
			ActionType: SelectEvidenceActionType,
			Payload:    map[string]interface{}{"operation": "toggle", "log_index": 15, "selected": true},
		},
		{
			ActionType: "triage_alert",
			Payload:    map[string]interface{}{"user_choice": true},
		},
		{
			ActionType: "submit_decision",
			Payload:    map[string]interface{}{"discovered_findings": []interface{}{"f1", "f2"}},
		},
	}

	result := ReconstructFromActions(scenario, actions)

	if !reflect.DeepEqual(result.SelectedLogIndices, []int{15}) {
		t.Fatalf("select_evidence actions should be authoritative, got %v", result.SelectedLogIndices)
	}
	if result.FindingsPoints != 10 {
		t.Fatalf("FindingsPoints = %f, want 10 (only f1 selected)", result.FindingsPoints)
	}
}

func TestReconstructFromActions_LegacyFindingWithoutEvidence(t *testing.T) {
	scenario := mockScenarioWithFindings(
		keyFinding("f1", "No authored evidence", "domain4", 10, []interface{}{}),
	)
	actions := []models.SessionAction{
		{
			ActionType: "triage_alert",
			Payload:    map[string]interface{}{"user_choice": true},
		},
		{
			ActionType: "submit_decision",
			Payload:    map[string]interface{}{"discovered_findings": []interface{}{"f1"}},
		},
	}

	result := ReconstructFromActions(scenario, actions)
	if len(result.SelectedLogIndices) != 0 || result.FindingsPoints != 0 {
		t.Fatalf("finding without evidence must not be invented, got selected=%v findings=%f", result.SelectedLogIndices, result.FindingsPoints)
	}
}

func TestRubricVersionIsEvidenceClick(t *testing.T) {
	if RubricVersion != "rubric-v2-evidence-click" {
		t.Fatalf("RubricVersion = %q, want rubric-v2-evidence-click", RubricVersion)
	}
}

func TestCalculateUpdatedProficiency(t *testing.T) {
	// Formula: 0.7 * old + 0.3 * new (alpha = 0.30)
	updated := CalculateUpdatedProficiency(50.0, 100.0)
	if updated != 65.0 {
		t.Fatalf("expected updated proficiency 65.0, got %f", updated)
	}

	updated2 := CalculateUpdatedProficiency(80.0, 40.0)
	if updated2 != 68.0 {
		t.Fatalf("expected updated proficiency 68.0, got %f", updated2)
	}
}
