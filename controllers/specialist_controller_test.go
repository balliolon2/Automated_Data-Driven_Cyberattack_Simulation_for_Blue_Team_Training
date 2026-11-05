package controllers

import (
	"reflect"
	"testing"

	"cybersim/models"
	"cybersim/services"
)

// TestSessionSnapshotDrivesSpecialistEvidenceReview verifies that the review
// uses the immutable ScenarioSnapshot rather than the live scenarios row. The
// two scenarios deliberately disagree about which log index proves the finding.
func TestSessionSnapshotDrivesSpecialistEvidenceReview(t *testing.T) {
	snapshotScenario, err := decodeScenarioSnapshot(models.ScenarioSnapshot{
		SnapshotID: "snap-1",
		ScenarioID: "scen-snapshot",
		ScenarioData: map[string]interface{}{
			"scenario_id": "scen-snapshot",
			"title":       "Snapshot title",
			"initial_logs": []interface{}{
				map[string]interface{}{"message": "snapshot index zero"},
				map[string]interface{}{"message": "snapshot evidence"},
			},
			"expected_outcomes": map[string]interface{}{
				"key_findings": []interface{}{
					map[string]interface{}{
						"id":                   "f1",
						"description":          "Snapshot finding",
						"domain_id":            "domain1",
						"points":               10,
						"evidence_log_indices": []interface{}{1},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("decodeScenarioSnapshot() error = %v", err)
	}

	liveScenario := models.Scenario{
		ScenarioID: "scen-snapshot",
		Title:      "Live edited title",
		InitialLogs: []interface{}{
			map[string]interface{}{"message": "live evidence"},
			map[string]interface{}{"message": "live index one"},
		},
		ExpectedOutcomes: map[string]interface{}{
			"key_findings": []interface{}{
				map[string]interface{}{
					"id":                   "f1",
					"description":          "Live finding",
					"domain_id":            "domain1",
					"points":               10,
					"evidence_log_indices": []interface{}{0},
				},
			},
		},
	}

	actions := []models.SessionAction{
		{
			ActionType: services.SelectEvidenceActionType,
			StepOrder:  1,
			Payload:    map[string]interface{}{"operation": "toggle", "log_index": 1, "selected": true},
		},
	}
	session := models.SimulationSession{
		SessionID:     "sess-snapshot",
		ScenarioID:    "scen-snapshot",
		RubricVersion: services.RubricVersion,
	}

	snapshotReview := services.BuildEvidenceReviewForSession(snapshotScenario, session, actions)
	liveReview := services.BuildEvidenceReviewForSession(liveScenario, session, actions)

	if len(snapshotReview.EvidenceFindings) != 1 || !snapshotReview.EvidenceFindings[0].Found {
		t.Fatalf("snapshot review must credit the finding from snapshot evidence, got %+v", snapshotReview.EvidenceFindings)
	}
	if !reflect.DeepEqual(snapshotReview.EvidenceFindings[0].MatchedLogIndices, []int{1}) {
		t.Fatalf("snapshot matched indices = %v, want [1]", snapshotReview.EvidenceFindings[0].MatchedLogIndices)
	}
	if got := snapshotReview.EvidenceFindings[0].EvidenceLogs[0].Message; got != "snapshot evidence" {
		t.Fatalf("snapshot evidence message = %q, want snapshot evidence", got)
	}
	if snapshotReview.NoiseSelectedCount != 0 || snapshotReview.EvidencePenalty != 0 {
		t.Fatalf("snapshot review noise = %d penalty = %f, want 0/0",
			snapshotReview.NoiseSelectedCount, snapshotReview.EvidencePenalty)
	}

	if liveReview.EvidenceFindings[0].Found {
		t.Fatal("live row would credit index 0, so using it would diverge from the scored snapshot")
	}
	if liveReview.NoiseSelectedCount != 1 || liveReview.EvidencePenalty != 2 {
		t.Fatalf("live review noise = %d penalty = %f, want 1/2",
			liveReview.NoiseSelectedCount, liveReview.EvidencePenalty)
	}
}
