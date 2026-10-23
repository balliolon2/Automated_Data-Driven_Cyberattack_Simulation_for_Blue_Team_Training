package controllers

import (
	"errors"
	"reflect"
	"testing"

	"cybersim/models"
)

// appendSessionAction/lockSessionRow are DB-backed: they need PostgreSQL row
// locks and the (session_id, step_order) unique constraint. This repository has
// no SQLite driver, so those paths are covered by
// scripts/test_simulation_integration.py instead. The pure pieces that decide
// whether a log index is valid and which log array a session uses are covered
// here.

func TestValidateSelectionLogIndex(t *testing.T) {
	logs := []interface{}{
		map[string]interface{}{"message": "first"},
		map[string]interface{}{"message": "second"},
	}
	typedLogs := []map[string]interface{}{
		{"message": "first"},
		{"message": "second"},
	}

	tests := []struct {
		name    string
		logs    any
		index   int
		wantErr error
	}{
		{name: "index zero is valid", logs: logs, index: 0},
		{name: "last index is valid", logs: logs, index: 1},
		{name: "typed map slice is valid", logs: typedLogs, index: 1},
		{name: "negative index is rejected", logs: logs, index: -1, wantErr: errSelectionLogIndexOutOfRange},
		{name: "index equal to length is rejected", logs: logs, index: 2, wantErr: errSelectionLogIndexOutOfRange},
		{name: "nil logs are unavailable", logs: nil, index: 0, wantErr: errSelectionLogsUnavailable},
		{name: "non-slice logs are unavailable", logs: "not-a-log-array", index: 0, wantErr: errSelectionLogsUnavailable},
		{name: "empty logs reject index zero", logs: []interface{}{}, index: 0, wantErr: errSelectionLogIndexOutOfRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSelectionLogIndex(tt.logs, tt.index)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("validateSelectionLogIndex() = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("validateSelectionLogIndex() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestSnapshotInitialLogs(t *testing.T) {
	logs := []interface{}{
		map[string]interface{}{"message": "first"},
		map[string]interface{}{"message": "second"},
	}

	t.Run("decodes JSONB map snapshot", func(t *testing.T) {
		got, ok := snapshotInitialLogs(map[string]interface{}{
			"scenario_id":  "scen-1",
			"initial_logs": logs,
		})
		if !ok {
			t.Fatal("expected snapshot logs to be found")
		}
		if !reflect.DeepEqual(got, logs) {
			t.Fatalf("snapshotInitialLogs() = %v, want %v", got, logs)
		}
	})

	t.Run("accepts typed scenario snapshot", func(t *testing.T) {
		got, ok := snapshotInitialLogs(models.Scenario{InitialLogs: logs})
		if !ok {
			t.Fatal("expected typed snapshot logs to be found")
		}
		if !reflect.DeepEqual(got, logs) {
			t.Fatalf("snapshotInitialLogs() = %v, want %v", got, logs)
		}
	})

	t.Run("accepts pointer scenario snapshot", func(t *testing.T) {
		got, ok := snapshotInitialLogs(&models.Scenario{InitialLogs: logs})
		if !ok {
			t.Fatal("expected pointer snapshot logs to be found")
		}
		if !reflect.DeepEqual(got, logs) {
			t.Fatalf("snapshotInitialLogs() = %v, want %v", got, logs)
		}
	})

	t.Run("decodes raw JSON snapshot", func(t *testing.T) {
		raw := []byte(`{"initial_logs":[{"message":"first"},{"message":"second"}]}`)
		got, ok := snapshotInitialLogs(raw)
		if !ok {
			t.Fatal("expected raw JSON snapshot logs to be found")
		}
		count, isArray := scenarioLogCount(got)
		if !isArray || count != 2 {
			t.Fatalf("decoded raw JSON logs = %d (array=%v), want 2", count, isArray)
		}
	})

	t.Run("rejects missing and nil logs", func(t *testing.T) {
		cases := []any{
			nil,
			map[string]interface{}{"scenario_id": "scen-1"},
			map[string]interface{}{"initial_logs": nil},
			models.Scenario{},
			&models.Scenario{},
			[]byte(`{"scenario_id":"scen-1"}`),
			"not-a-snapshot",
		}
		for _, data := range cases {
			if got, ok := snapshotInitialLogs(data); ok {
				t.Fatalf("snapshotInitialLogs(%v) = %v, true; want false", data, got)
			}
		}
	})
}

func TestIsUniqueViolation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "unrelated error", err: errors.New("connection reset"), want: false},
		{name: "sqlstate", err: errors.New("ERROR: duplicate key value violates unique constraint \"session_actions_session_id_step_order_key\" (SQLSTATE 23505)"), want: true},
		{name: "duplicate key", err: errors.New("duplicate key value violates unique constraint"), want: true},
		{name: "unique constraint", err: errors.New("unique constraint failed"), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUniqueViolation(tt.err); got != tt.want {
				t.Fatalf("isUniqueViolation() = %v, want %v", got, tt.want)
			}
		})
	}
}
