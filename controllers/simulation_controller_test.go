package controllers

import (
	"fmt"
	"reflect"
	"testing"

	"cybersim/models"
	"cybersim/services"
)

func TestSanitizeScenario_DeterministicSessionScopedShuffleAndScoringInvariance(t *testing.T) {
	scenario := *services.GenerateFallbackScenario("domain4", "Security Operations")
	scenario.ScenarioID = "scen-shuffle-test-01"

	t.Run("strips sensitive fields and preserves id and label", func(t *testing.T) {
		sanitized := sanitizeScenario(scenario, "sess-alpha-001")
		stepsMap, ok := sanitized.PlaybookSteps.(map[string]interface{})
		if !ok {
			t.Fatalf("expected sanitized PlaybookSteps to be map[string]interface{}")
		}

		for _, phase := range []string{"containment", "eradication", "recovery"} {
			actions, ok := stepsMap[phase].([]map[string]interface{})
			if !ok {
				t.Fatalf("expected phase %s to be []map[string]interface{}", phase)
			}
			if len(actions) < 3 {
				t.Fatalf("expected at least 3 actions in phase %s, got %d", phase, len(actions))
			}
			for _, act := range actions {
				if _, hasID := act["id"]; !hasID {
					t.Errorf("missing id in sanitized action: %v", act)
				}
				if _, hasLabel := act["label"]; !hasLabel {
					t.Errorf("missing label in sanitized action: %v", act)
				}
				if _, hasCorrect := act["is_correct"]; hasCorrect {
					t.Errorf("leaked is_correct in sanitized action: %v", act)
				}
				if _, hasPoints := act["points"]; hasPoints {
					t.Errorf("leaked points in sanitized action: %v", act)
				}
				if _, hasExplanation := act["explanation"]; hasExplanation {
					t.Errorf("leaked explanation in sanitized action: %v", act)
				}
			}
		}
	})

	t.Run("produces identical option order across repeated calls for the same session_id (refresh stability)", func(t *testing.T) {
		sessionID := "sess-refresh-stability-99"
		first := sanitizeScenario(scenario, sessionID)
		firstDirect := sanitizePlaybookSteps(scenario.PlaybookSteps, sessionID)
		if !reflect.DeepEqual(first.PlaybookSteps, firstDirect) {
			t.Fatalf("expected sanitizeScenario and sanitizePlaybookSteps to produce identical PlaybookSteps")
		}
		for i := 0; i < 10; i++ {
			next := sanitizeScenario(scenario, sessionID)
			if !reflect.DeepEqual(first.PlaybookSteps, next.PlaybookSteps) {
				t.Fatalf("iteration %d: option order changed for same sessionID %s", i, sessionID)
			}
		}
	})

	t.Run("distributes correct option across non-zero indices across different session_ids", func(t *testing.T) {
		nonZeroIndexSeen := map[string]bool{
			"containment": false,
			"eradication": false,
			"recovery":    false,
		}
		correctIDs := map[string]string{
			"containment": "c1",
			"eradication": "e1",
			"recovery":    "r1",
		}

		for i := 0; i < 30; i++ {
			sessID := fmt.Sprintf("sess-distribution-%d", i)
			sanitized := sanitizeScenario(scenario, sessID)
			stepsMap := sanitized.PlaybookSteps.(map[string]interface{})

			for _, phase := range []string{"containment", "eradication", "recovery"} {
				actions := stepsMap[phase].([]map[string]interface{})
				if actions[0]["id"] != correctIDs[phase] {
					nonZeroIndexSeen[phase] = true
				}
			}
		}

		for phase, seen := range nonZeroIndexSeen {
			if !seen {
				t.Errorf("phase %s: correct option remained stuck at index 0 across all 30 sessions", phase)
			}
		}
	})

	t.Run("scoring via ScoreSubmission and Domain Breakdown are 100% invariant to shuffled display order", func(t *testing.T) {
		baselineResult := services.ScoreSubmission(scenario, models.SubmitScenarioInput{
			SessionID:          "sess-score-baseline",
			IsTruePositive:     true,
			TPFPSelected:       true,
			SelectedActions:    []string{"c1", "e1", "r1"},
			DiscoveredFindings: []string{"f1"},
		})

		if baselineResult.TotalScore != 100.0 {
			t.Fatalf("expected baseline TotalScore 100.0, got %f", baselineResult.TotalScore)
		}

		for i := 0; i < 10; i++ {
			sessID := fmt.Sprintf("sess-score-shuffled-%d", i)
			_ = sanitizeScenario(scenario, sessID)

			res := services.ScoreSubmission(scenario, models.SubmitScenarioInput{
				SessionID:          sessID,
				IsTruePositive:     true,
				TPFPSelected:       true,
				SelectedActions:    []string{"r1", "c1", "e1"},
				DiscoveredFindings: []string{"f1"},
			})

			if res.TotalScore != baselineResult.TotalScore || res.ResponsePoints != baselineResult.ResponsePoints {
				t.Errorf("session %s: score changed after shuffling: got TotalScore=%f ResponsePoints=%f", sessID, res.TotalScore, res.ResponsePoints)
			}
			if !reflect.DeepEqual(res.DomainScores, baselineResult.DomainScores) {
				t.Errorf("session %s: DomainScores changed after shuffling: got %v, want %v", sessID, res.DomainScores, baselineResult.DomainScores)
			}
			if !reflect.DeepEqual(res.DomainScoreMap, baselineResult.DomainScoreMap) {
				t.Errorf("session %s: DomainScoreMap changed after shuffling: got %v, want %v", sessID, res.DomainScoreMap, baselineResult.DomainScoreMap)
			}
		}
	})
}
