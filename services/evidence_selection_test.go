package services

import (
	"reflect"
	"testing"
	"time"

	"cybersim/models"
)

func selectionAction(actionType string, payload any, stepOrder int) models.SessionAction {
	return models.SessionAction{
		SessionID:  "session-1",
		StepOrder:  stepOrder,
		ActionType: actionType,
		Payload:    payload,
		Timestamp:  time.Unix(0, 0),
	}
}

func TestReplaySelectedLogIndices(t *testing.T) {
	tests := []struct {
		name    string
		actions []models.SessionAction
		want    []int
	}{
		{
			name:    "no actions returns empty selection",
			actions: nil,
			want:    []int{},
		},
		{
			name: "ignores action types other than select_evidence",
			actions: []models.SessionAction{
				selectionAction("view_log", map[string]any{"log_index": 3, "selected": true}, 1),
				selectionAction("submit_decision", map[string]any{"discovered_findings": []string{"f1"}}, 2),
			},
			want: []int{},
		},
		{
			name: "toggle selects and deselects in step order",
			actions: []models.SessionAction{
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 5, "selected": true}, 1),
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 2, "selected": true}, 2),
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 5, "selected": false}, 3),
			},
			want: []int{2},
		},
		{
			name: "repeated select of the same index does not duplicate it",
			actions: []models.SessionAction{
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 4, "selected": true}, 1),
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 4, "selected": true}, 2),
			},
			want: []int{4},
		},
		{
			name: "deselecting an unselected index is harmless",
			actions: []models.SessionAction{
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 7, "selected": false}, 1),
			},
			want: []int{},
		},
		{
			name: "clear empties the selection",
			actions: []models.SessionAction{
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 1, "selected": true}, 1),
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 2, "selected": true}, 2),
				selectionAction("select_evidence", map[string]any{"operation": "clear"}, 3),
			},
			want: []int{},
		},
		{
			name: "selection after clear only contains later toggles",
			actions: []models.SessionAction{
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 1, "selected": true}, 1),
				selectionAction("select_evidence", map[string]any{"operation": "clear"}, 2),
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 9, "selected": true}, 3),
			},
			want: []int{9},
		},
		{
			name: "index zero is a valid selection",
			actions: []models.SessionAction{
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 0, "selected": true}, 1),
			},
			want: []int{0},
		},
		{
			name: "JSONB float indices are decoded",
			actions: []models.SessionAction{
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": float64(3), "selected": true}, 1),
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": float64(11), "selected": true}, 2),
			},
			want: []int{3, 11},
		},
		{
			name: "malformed payloads are skipped without failing replay",
			actions: []models.SessionAction{
				selectionAction("select_evidence", nil, 1),
				selectionAction("select_evidence", "not-a-payload", 2),
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 12}, 3),
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 13, "selected": "yes"}, 4),
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "selected": true}, 5),
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 2.5, "selected": true}, 6),
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": "4", "selected": true}, 7),
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": -1, "selected": true}, 8),
				selectionAction("select_evidence", map[string]any{"operation": "unknown", "log_index": 1, "selected": true}, 9),
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 6, "selected": true}, 10),
			},
			want: []int{6},
		},
		{
			name: "actions are replayed by step order regardless of input order",
			actions: []models.SessionAction{
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 5, "selected": false}, 3),
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 5, "selected": true}, 1),
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 2, "selected": true}, 2),
			},
			want: []int{2},
		},
		{
			name: "returns indices sorted ascending",
			actions: []models.SessionAction{
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 8, "selected": true}, 1),
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 2, "selected": true}, 2),
				selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 5, "selected": true}, 3),
			},
			want: []int{2, 5, 8},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaySelectedLogIndices(tt.actions)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ReplaySelectedLogIndices() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReplaySelectedLogIndicesIsIdempotent(t *testing.T) {
	actions := []models.SessionAction{
		selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 1, "selected": true}, 1),
		selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 4, "selected": true}, 2),
		selectionAction("select_evidence", map[string]any{"operation": "clear"}, 3),
		selectionAction("select_evidence", map[string]any{"operation": "toggle", "log_index": 4, "selected": true}, 4),
	}

	first := ReplaySelectedLogIndices(actions)
	second := ReplaySelectedLogIndices(actions)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("replaying the same actions twice changed the result: first=%v second=%v", first, second)
	}
	if want := []int{4}; !reflect.DeepEqual(first, want) {
		t.Fatalf("ReplaySelectedLogIndices() = %v, want %v", first, want)
	}
}
