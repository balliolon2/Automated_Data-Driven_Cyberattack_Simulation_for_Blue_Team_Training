package services

import (
	"testing"

	"cybersim/models"
)

func createMockScenario() models.Scenario {
	return models.Scenario{
		ScenarioID:      "scen-test-01",
		Title:           "Test Incident Scenario",
		DomainID:        "domain4",
		IsTruePositive:  true,
		TpFpExplanation: "Malicious C2 beaconing confirmed via threat intel matching destination IP.",
		ExpectedOutcomes: map[string]interface{}{
			"key_findings": []interface{}{
				map[string]interface{}{
					"id":                   "f1",
					"description":          "C2 beaconing detected",
					"domain_id":            "domain4",
					"points":               10.0,
					"evidence_log_indices": []interface{}{15, 16},
					"explanation":          "Direct communication with known bad actor.",
				},
				map[string]interface{}{
					"id":                   "f2",
					"description":          "Privilege escalation attempt",
					"domain_id":            "domain2",
					"points":               10.0,
					"evidence_log_indices": []interface{}{17},
					"explanation":          "Exploitation of local vulnerability.",
				},
			},
		},
		PlaybookSteps: map[string]interface{}{
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
		},
	}
}

func TestScoreSubmission_PerfectScore(t *testing.T) {
	scenario := createMockScenario()
	input := models.SubmitScenarioInput{
		SessionID:          "sess-01",
		IsTruePositive:     true,
		TPFPSelected:       true,
		SelectedActions:    []string{"c1", "e1", "r1"},
		DiscoveredFindings: []string{"f1", "f2"},
	}

	result := ScoreSubmission(scenario, input)

	if !result.TPFPCorrect {
		t.Errorf("expected TPFPCorrect to be true, got false")
	}
	if result.TPFPPoints != 20.0 {
		t.Errorf("expected TPFPPoints 20.0, got %f", result.TPFPPoints)
	}
	if result.FindingsPoints != 20.0 {
		t.Errorf("expected FindingsPoints 20.0, got %f", result.FindingsPoints)
	}
	if result.ResponsePoints != 15.0 {
		t.Errorf("expected ResponsePoints 15.0, got %f", result.ResponsePoints)
	}
	if result.TotalScore != 100.0 {
		t.Errorf("expected TotalScore 100.0, got %f", result.TotalScore)
	}

	// Verify all findings marked found
	for _, f := range result.FindingsDetail {
		if !f.Found {
			t.Errorf("expected finding %s to be found", f.ID)
		}
	}

	// Verify responses
	for _, r := range result.ResponseDetail {
		if r.Selected && !r.IsCorrect {
			t.Errorf("did not expect incorrect response to be selected: %s", r.ID)
		}
	}
}

func TestScoreSubmission_WrongTPFP(t *testing.T) {
	scenario := createMockScenario()
	input := models.SubmitScenarioInput{
		SessionID:          "sess-02",
		IsTruePositive:     false, // Wrong! Scenario is True Positive
		TPFPSelected:       true,
		SelectedActions:    []string{"c1", "e1", "r1"},
		DiscoveredFindings: []string{"f1", "f2"},
	}

	result := ScoreSubmission(scenario, input)

	if result.TPFPCorrect {
		t.Errorf("expected TPFPCorrect to be false, got true")
	}
	if result.TPFPPoints != 0.0 {
		t.Errorf("expected TPFPPoints 0.0, got %f", result.TPFPPoints)
	}
	// Total max points: 20 (TP/FP) + 20 (Findings) + 15 (Responses) = 55
	// Total earned: 0 + 20 + 15 = 35 -> 35 / 55 * 100 = 63.64%
	if result.TotalScore >= 100.0 || result.TotalScore <= 60.0 {
		t.Errorf("expected TotalScore around 63.64, got %f", result.TotalScore)
	}
}

func TestScoreSubmission_DistractorPenalty(t *testing.T) {
	scenario := createMockScenario()
	input := models.SubmitScenarioInput{
		SessionID:          "sess-03",
		IsTruePositive:     true,
		TPFPSelected:       true,
		SelectedActions:    []string{"c1", "c2", "e1", "r1"}, // Selected c2 (distractor with -3 pts)
		DiscoveredFindings: []string{"f1", "f2"},
	}

	result := ScoreSubmission(scenario, input)

	// Response points: 5 (c1) + (-3) (c2) + 5 (e1) + 5 (r1) = 12
	if result.ResponsePoints != 12.0 {
		t.Errorf("expected ResponsePoints 12.0, got %f", result.ResponsePoints)
	}

	// Verify c2 response detail
	var c2Detail *models.ResponseActionResult
	for i := range result.ResponseDetail {
		if result.ResponseDetail[i].ID == "c2" {
			c2Detail = &result.ResponseDetail[i]
			break
		}
	}
	if c2Detail == nil {
		t.Fatalf("c2 response detail not found")
	}
	if !c2Detail.Selected || c2Detail.IsCorrect || c2Detail.Earned != -3 {
		t.Errorf("c2 detail mismatch: Selected=%v, IsCorrect=%v, Earned=%d", c2Detail.Selected, c2Detail.IsCorrect, c2Detail.Earned)
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
		t.Errorf("expected reconstructed TPFPCorrect to be true")
	}
	if reconstructed.TotalScore != 100.0 {
		t.Errorf("expected reconstructed TotalScore 100.0, got %f", reconstructed.TotalScore)
	}
}

func TestCalculateUpdatedProficiency(t *testing.T) {
	// Formula: 0.7 * old + 0.3 * new (alpha = 0.30)
	// Example: old = 50.0, new = 100.0 -> 0.7 * 50 + 0.3 * 100 = 35 + 30 = 65.0
	updated := CalculateUpdatedProficiency(50.0, 100.0)
	if updated != 65.0 {
		t.Errorf("expected updated proficiency 65.0, got %f", updated)
	}

	// Example: old = 80.0, new = 40.0 -> 0.7 * 80 + 0.3 * 40 = 56 + 12 = 68.0
	updated2 := CalculateUpdatedProficiency(80.0, 40.0)
	if updated2 != 68.0 {
		t.Errorf("expected updated proficiency 68.0, got %f", updated2)
	}
}
