package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEmbeddedScenarioSeedIsAvailable(t *testing.T) {
	if len(scenarioSeedJSON) == 0 {
		t.Fatalf("embedded scenarios_seed.json is empty")
	}

	var scenarios []struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal(scenarioSeedJSON, &scenarios); err != nil {
		t.Fatalf("embedded scenarios_seed.json is not valid JSON: %v", err)
	}
	if len(scenarios) == 0 {
		t.Fatalf("embedded scenarios_seed.json contains no scenarios")
	}
	for i, scenario := range scenarios {
		if strings.TrimSpace(scenario.Title) == "" {
			t.Fatalf("embedded scenario %d has no title", i)
		}
	}
}
