package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cybersim/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestGenerateFallbackScenario_AllDomainsPassValidationAndNoBoilerplate(t *testing.T) {
	domains := map[string]string{
		"domain1": "General Security Concepts",
		"domain2": "Threats, Vulnerabilities, and Mitigations",
		"domain3": "Security Architecture",
		"domain4": "Security Operations",
		"domain5": "Security Program Management and Oversight",
	}

	banned := []string{
		"Power off all domain controllers immediately",
		"Reformat all network switches in the subnet",
		"Disable MFA enforcement for affected users",
	}

	for domainID, domainName := range domains {
		t.Run(domainID, func(t *testing.T) {
			scenario := GenerateFallbackScenario(domainID, domainName)
			if err := ValidateScenarioCandidate(scenario); err != nil {
				t.Fatalf("expected fallback scenario for %s to pass validation, got error: %v", domainID, err)
			}

			playbook, ok := scenario.PlaybookSteps.(map[string]interface{})
			if !ok {
				t.Fatalf("expected playbook_steps to be map[string]interface{}")
			}

			for _, phase := range []string{"containment", "eradication", "recovery"} {
				steps := extractPhaseStepsForTest(t, playbook, phase)
				if len(steps) < 3 || len(steps) > 4 {
					t.Errorf("%s/%s: expected 3-4 options, got %d", domainID, phase, len(steps))
				}

				correctCount := 0
				distractorCount := 0
				for _, step := range steps {
					label, _ := step["label"].(string)
					isCorrect, _ := step["is_correct"].(bool)
					if isCorrect {
						correctCount++
					} else {
						distractorCount++
					}

					for _, b := range banned {
						if strings.EqualFold(strings.TrimSpace(label), strings.TrimSpace(b)) {
							t.Errorf("%s/%s: found banned boilerplate label %q", domainID, phase, label)
						}
					}

					wordCount := len(strings.Fields(label))
					if wordCount < 8 {
						t.Errorf("%s/%s: option label too short (%d words): %q", domainID, phase, wordCount, label)
					}
				}

				if correctCount < 1 {
					t.Errorf("%s/%s: expected at least 1 correct option, got %d", domainID, phase, correctCount)
				}
				if distractorCount < 2 {
					t.Errorf("%s/%s: expected at least 2 distractor options, got %d", domainID, phase, distractorCount)
				}
			}
		})
	}
}

func TestValidateScenarioCandidate_AntiPredictabilityRules(t *testing.T) {
	baseScenario := GenerateFallbackScenario("domain4", "Security Operations")

	t.Run("rejects phase with fewer than 3 options", func(t *testing.T) {
		s := cloneScenarioForTest(baseScenario)
		pb := s.PlaybookSteps.(map[string]interface{})
		pb["containment"] = []map[string]interface{}{
			{"id": "c1", "label": "Isolate host WS-DEV09 from the corporate network segment immediately and block outbound traffic", "is_correct": true, "domain_id": "domain4", "points": 5, "explanation": "Valid"},
			{"id": "c2", "label": "Restart the local DNS resolver service on WS-DEV09 while keeping outbound C2 sockets open", "is_correct": false, "domain_id": "domain4", "points": -3, "explanation": "Invalid"},
		}
		err := ValidateScenarioCandidate(s)
		if err == nil {
			t.Errorf("expected error for phase with only 2 options, got nil")
		}
	})

	t.Run("rejects phase with fewer than 2 distractors", func(t *testing.T) {
		s := cloneScenarioForTest(baseScenario)
		pb := s.PlaybookSteps.(map[string]interface{})
		pb["eradication"] = []map[string]interface{}{
			{"id": "e1", "label": "Remove the malicious scheduled task on WS-DEV09 and purge the C2 beacon binary from disk", "is_correct": true, "domain_id": "domain4", "points": 5, "explanation": "Valid"},
			{"id": "e2", "label": "Revoke compromised service credentials on WS-DEV09 and rotate local administrator password hashes", "is_correct": true, "domain_id": "domain4", "points": 5, "explanation": "Valid"},
			{"id": "e3", "label": "Clear system event logs on WS-DEV09 without removing the persistent background service", "is_correct": false, "domain_id": "domain4", "points": -3, "explanation": "Invalid"},
		}
		err := ValidateScenarioCandidate(s)
		if err == nil {
			t.Errorf("expected error for phase with only 1 distractor, got nil")
		}
	})

	t.Run("rejects banned boilerplate distractor label", func(t *testing.T) {
		s := cloneScenarioForTest(baseScenario)
		pb := s.PlaybookSteps.(map[string]interface{})
		pb["containment"] = []map[string]interface{}{
			{"id": "c1", "label": "Block C2 destination IP 198.51.100.42 at perimeter firewall and isolate WS-DEV09", "is_correct": true, "domain_id": "domain4", "points": 5, "explanation": "Valid"},
			{"id": "c2", "label": "Power off all domain controllers immediately", "is_correct": false, "domain_id": "domain4", "points": -3, "explanation": "Boilerplate"},
			{"id": "c3", "label": "Allow outbound traffic to 198.51.100.42 on WS-DEV09 while disabling firewall logging", "is_correct": false, "domain_id": "domain4", "points": -3, "explanation": "Invalid"},
		}
		err := ValidateScenarioCandidate(s)
		if err == nil {
			t.Errorf("expected error for banned boilerplate label, got nil")
		}
	})

	t.Run("rejects duplicate labels across options", func(t *testing.T) {
		s := cloneScenarioForTest(baseScenario)
		pb := s.PlaybookSteps.(map[string]interface{})
		dupLabel := "Isolate WS-DEV09 from the internal VLAN and block 198.51.100.42 at the perimeter firewall"
		pb["recovery"] = []map[string]interface{}{
			{"id": "r1", "label": dupLabel, "is_correct": true, "domain_id": "domain4", "points": 5, "explanation": "Valid"},
			{"id": "r2", "label": dupLabel, "is_correct": false, "domain_id": "domain4", "points": -3, "explanation": "Duplicate"},
			{"id": "r3", "label": "Restore WS-DEV09 from an unverified snapshot taken after the initial C2 beacon alert", "is_correct": false, "domain_id": "domain4", "points": -3, "explanation": "Invalid"},
		}
		err := ValidateScenarioCandidate(s)
		if err == nil {
			t.Errorf("expected error for duplicate option label, got nil")
		}
	})

	t.Run("rejects phase with more than 4 options", func(t *testing.T) {
		s := cloneScenarioForTest(baseScenario)
		pb := s.PlaybookSteps.(map[string]interface{})
		pb["containment"] = []map[string]interface{}{
			{"id": "c1", "label": "Block outbound C2 traffic to 198.51.100.42 at the perimeter firewall and isolate WS-DEV09", "is_correct": true, "domain_id": "domain4", "points": 5, "explanation": "Valid"},
			{"id": "c2", "label": "Unplug the core switch uplink for the engineering building housing WS-DEV09 immediately", "is_correct": false, "domain_id": "domain4", "points": -3, "explanation": "Invalid"},
			{"id": "c3", "label": "Redirect inbound packets from 198.51.100.42 on WS-DEV09 to the internal domain controller", "is_correct": false, "domain_id": "domain4", "points": -3, "explanation": "Invalid"},
			{"id": "c4", "label": "Flush volatile RAM caches on WS-DEV09 before capturing memory artifacts for 198.51.100.42", "is_correct": false, "domain_id": "domain4", "points": -3, "explanation": "Invalid"},
			{"id": "c5", "label": "Disable the internal DNS resolver for the entire subnet hosting WS-DEV09 and 198.51.100.42", "is_correct": false, "domain_id": "domain4", "points": -3, "explanation": "Invalid"},
		}
		err := ValidateScenarioCandidate(s)
		if err == nil {
			t.Errorf("expected error for phase with 5 options, got nil")
		}
	})
}

func TestSynthesizeFullScenario_BackfillsLegacySingleActionsAndFalsePositive(t *testing.T) {
	t.Run("backfills contextual distractors when LLM returns legacy single-string actions", func(t *testing.T) {
		micro := aiMicroScenario{
			Title:             "Suspicious Web Shell Execution",
			Description:       "Web server spawned anomalous shell processes.",
			IsTruePositive:    true,
			TpFpExplanation:   "True positive web shell activity.",
			AttackerIP:        "185.220.101.45",
			TargetHost:        "SRV-WEB01",
			AttackLogMessage:  "w3wp.exe spawned cmd.exe /c whoami",
			KeyFinding:        "Web shell process execution on SRV-WEB01",
			ContainmentAction: "Isolate SRV-WEB01 from the production VLAN immediately and block 185.220.101.45 at the perimeter firewall",
			EradicationAction: "Remove the uploaded web shell script from SRV-WEB01 webroot and patch the vulnerable upload handler",
			RecoveryAction:    "Rebuild SRV-WEB01 from a verified gold image and monitor outbound connections for 72 hours",
		}

		s := synthesizeFullScenario(micro, "domain2", "Threats, Vulnerabilities, and Mitigations")
		if err := ValidateScenarioCandidate(s); err != nil {
			t.Fatalf("expected synthesized scenario from legacy micro struct to pass validation, got: %v", err)
		}

		pb := s.PlaybookSteps.(map[string]interface{})
		for _, phase := range []string{"containment", "eradication", "recovery"} {
			steps := extractPhaseStepsForTest(t, pb, phase)
			for _, step := range steps {
				isCorrect, _ := step["is_correct"].(bool)
				label, _ := step["label"].(string)
				if !isCorrect {
					if !strings.Contains(label, "SRV-WEB01") && !strings.Contains(label, "185.220.101.45") {
						t.Errorf("expected contextual distractor in %s to reference SRV-WEB01 or 185.220.101.45, got: %q", phase, label)
					}
				}
			}
		}
	})

	t.Run("synthesizes appropriate non-destructive actions and disruptive distractors for False Positive scenarios", func(t *testing.T) {
		micro := aiMicroScenario{
			Title:            "Authorized Vulnerability Scanner Traffic",
			Description:      "High volume port scan from internal security scanner.",
			IsTruePositive:   false,
			TpFpExplanation:  "False positive: scheduled internal Qualys scan.",
			AttackerIP:       "10.0.99.10",
			TargetHost:       "SRV-APP02",
			AttackLogMessage: "Multiple SYN packets from 10.0.99.10 to SRV-APP02",
			KeyFinding:       "Traffic matches scheduled maintenance window CRQ-4012",
		}

		s := synthesizeFullScenario(micro, "domain4", "Security Operations")
		if err := ValidateScenarioCandidate(s); err != nil {
			t.Fatalf("expected synthesized FP scenario to pass validation, got: %v", err)
		}
	})
}

func TestStaticScenarioSeedCatalog_AllScenariosPassValidation(t *testing.T) {
	seedPath := filepath.Join("..", "scenarios_seed.json")
	rawBytes, err := os.ReadFile(seedPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", seedPath, err)
	}

	var scenarios []models.Scenario
	if err := json.Unmarshal(rawBytes, &scenarios); err != nil {
		t.Fatalf("failed to unmarshal scenarios_seed.json: %v", err)
	}

	if len(scenarios) < 15 {
		t.Fatalf("expected at least 15 static scenarios in scenarios_seed.json, got %d", len(scenarios))
	}

	for i := range scenarios {
		sc := &scenarios[i]
		t.Run(sc.Title, func(t *testing.T) {
			if err := ValidateScenarioCandidate(sc); err != nil {
				t.Errorf("static scenario %q (%s) failed anti-predictability validation: %v", sc.Title, sc.DomainID, err)
			}
		})
	}
}

func TestUpgradeScenarioPlaybookSteps_RemediatesLegacyScenarios(t *testing.T) {
	legacyScenario := &models.Scenario{
		ScenarioID:      "legacy-ai-01",
		Title:           "Legacy Web Shell Incident",
		Description:     "Suspicious shell on SRV-WEB01.\n\nINVESTIGATION CLUES:\n- Target Host: SRV-WEB01\n- Suspected Attacker IP: 185.220.101.45\n- Alert Vector: Web Shell",
		DomainID:        "domain2",
		IsTruePositive:  true,
		TpFpExplanation: "True positive web shell execution.",
		InitialLogs:     make([]map[string]interface{}, 15),
		ExpectedOutcomes: map[string]interface{}{
			"key_findings": []map[string]interface{}{
				{"id": "f1", "description": "Web shell found", "domain_id": "domain2", "points": 10},
			},
		},
		PlaybookSteps: map[string]interface{}{
			"containment": []interface{}{
				map[string]interface{}{
					"id":          "c1",
					"label":       "Isolate SRV-WEB01 from the network immediately, block the attacker IP 185.220.101.45 and the C2 destination at the perimeter firewall, and kill the malicious powershell.exe and w3wp.exe child processes.",
					"is_correct":  true,
					"domain_id":   "domain2",
					"points":      5,
					"explanation": "Directly mitigates active containment risk.",
				},
				map[string]interface{}{
					"id":          "c2",
					"label":       "Power off all domain controllers immediately",
					"is_correct":  false,
					"domain_id":   "domain2",
					"points":      -3,
					"explanation": "Overly aggressive disruption to business infrastructure.",
				},
			},
			"eradication": []interface{}{
				map[string]interface{}{
					"id":          "e1",
					"label":       "Remove the web shell files from the web root directory, rebuild or clean-reimage SRV-WEB01 to eliminate any persistence mechanisms, patch the exploited web application vulnerability, and rotate all credentials that were accessible from the compromised host.",
					"is_correct":  true,
					"domain_id":   "domain2",
					"points":      5,
					"explanation": "Permanently removes threat persistence mechanisms.",
				},
				map[string]interface{}{
					"id":          "e2",
					"label":       "Reformat all network switches in the subnet",
					"is_correct":  false,
					"domain_id":   "domain2",
					"points":      -3,
					"explanation": "Unnecessary action unrelated to endpoint threat eradication.",
				},
			},
			"recovery": []interface{}{
				map[string]interface{}{
					"id":          "r1",
					"label":       "Restore SRV-WEB01 from a verified clean backup or rebuilt image, return it to production behind enhanced monitoring rules for web shell indicators and anomalous PowerShell activity, and validate normal application functionality and log integrity over the following 72 hours.",
					"is_correct":  true,
					"domain_id":   "domain2",
					"points":      5,
					"explanation": "Ensures safe restoration of business operations.",
				},
				map[string]interface{}{
					"id":          "r2",
					"label":       "Disable MFA enforcement for affected users",
					"is_correct":  false,
					"domain_id":   "domain2",
					"points":      -3,
					"explanation": "Degrades baseline security posture.",
				},
			},
		},
	}

	if err := ValidateScenarioCandidate(legacyScenario); err == nil {
		t.Fatalf("expected legacy scenario to fail validation before upgrade")
	}

	upgraded := UpgradeScenarioPlaybookSteps(legacyScenario)
	if !upgraded {
		t.Fatalf("expected UpgradeScenarioPlaybookSteps to return true for legacy scenario")
	}

	if err := ValidateScenarioCandidate(legacyScenario); err != nil {
		t.Fatalf("expected upgraded legacy scenario to pass validation, got: %v", err)
	}
}

func TestEnsureScenarioPlaybookQuality_DatabaseIntegration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("skipping DB remediation test (DATABASE_URL not set)")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("skipping DB remediation test (database not reachable): %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil || sqlDB.Ping() != nil {
		t.Skipf("skipping DB remediation test (database ping failed)")
	}

	seedPath := filepath.Join("..", "scenarios_seed.json")
	if err := EnsureScenarioPlaybookQuality(db, seedPath); err != nil {
		t.Fatalf("EnsureScenarioPlaybookQuality failed: %v", err)
	}

	var stored []models.Scenario
	if err := db.Find(&stored).Error; err != nil {
		t.Fatalf("failed to query scenarios: %v", err)
	}

	for i := range stored {
		if err := ValidateScenarioCandidate(&stored[i]); err != nil {
			t.Errorf("stored scenario %q (%s) failed validation after EnsureScenarioPlaybookQuality: %v", stored[i].Title, stored[i].ScenarioID, err)
		}
	}
}

func extractPhaseStepsForTest(t *testing.T, playbook map[string]interface{}, phase string) []map[string]interface{} {
	t.Helper()
	raw, exists := playbook[phase]
	if !exists {
		t.Fatalf("missing phase %s", phase)
	}
	steps := normalizePhaseStepSlice(raw)
	if steps == nil {
		t.Fatalf("unexpected type for phase %s: %T", phase, raw)
	}
	return steps
}

func cloneScenarioForTest(s *models.Scenario) *models.Scenario {
	cp := *s
	oldPB := s.PlaybookSteps.(map[string]interface{})
	newPB := make(map[string]interface{}, len(oldPB))
	for k, v := range oldPB {
		newPB[k] = v
	}
	cp.PlaybookSteps = newPB
	return &cp
}
