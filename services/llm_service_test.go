package services

import (
	"context"
	"encoding/json"
	"fmt"
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

			logs := normalizePhaseStepSlice(scenario.InitialLogs)
			if len(logs) < 28 || len(logs) > 46 {
				t.Errorf("%s: expected 28..46 fallback logs, got %d", domainID, len(logs))
			}
			for _, entry := range logs {
				msg, _ := entry["message"].(string)
				norm := strings.ToLower(msg)
				for _, bStr := range BannedLegacyLogSubstrings {
					if strings.Contains(norm, bStr) {
						t.Errorf("%s: fallback log contains banned legacy boilerplate %q: %q", domainID, bStr, msg)
					}
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

	t.Run("rejects banned legacy log boilerplate substrings", func(t *testing.T) {
		orig := normalizePhaseStepSlice(baseScenario.InitialLogs)
		for _, sample := range []string{
			"SIEM Alert Rule [CORR-9901]: Multi-stage incident correlation triggered on host WS-DEV09 involving IP 198.51.100.42",
			"Routine network telemetry event 21 logged for internal-service-21.corp.local",
			"Kerberos ticket TGT requested for user account user0@corp.local",
		} {
			s := cloneScenarioForTest(baseScenario)
			logs := cloneLogSliceForTest(orig)
			logs[0]["message"] = sample
			s.InitialLogs = logs
			if err := ValidateScenarioCandidate(s); err == nil {
				t.Errorf("expected ValidateScenarioCandidate to reject banned log boilerplate %q", sample)
			}
		}
	})

	t.Run("rejects more than 3 identical normalized log messages", func(t *testing.T) {
		s := cloneScenarioForTest(baseScenario)
		logs := cloneLogSliceForTest(normalizePhaseStepSlice(baseScenario.InitialLogs))
		for _, idx := range []int{0, 1, 2, 3} {
			logs[idx]["message"] = "  Duplicate heartbeat check  "
		}
		logs[1]["message"] = "duplicate heartbeat check"
		s.InitialLogs = logs
		if err := ValidateScenarioCandidate(s); err == nil {
			t.Errorf("expected ValidateScenarioCandidate to reject >3 identical log messages")
		}
	})
}

func TestSynthesizeFullScenario_BackfillsLegacySingleActionsAndFalsePositive(t *testing.T) {
	t.Run("backfills contextual distractors when LLM returns legacy single-string actions", func(t *testing.T) {
		micro := aiMicroScenario{
			Title:               "Suspicious Web Shell Execution",
			TitleTh:             "ตรวจพบการรัน Web Shell ต้องสงสัย",
			Description:         "Web server spawned anomalous shell processes.",
			DescriptionTh:       "เว็บเซิร์ฟเวอร์แตกโพรเซส shell ผิดปกติบน SRV-WEB01",
			IsTruePositive:      true,
			TpFpExplanation:     "True positive web shell activity.",
			TpFpExplanationTh:   "เหตุการณ์นี้เป็น True Positive จากกิจกรรม web shell",
			AttackerIP:          "185.220.101.45",
			TargetHost:          "SRV-WEB01",
			AttackLogMessage:    "w3wp.exe spawned cmd.exe /c whoami",
			KeyFinding:          "Web shell process execution on SRV-WEB01",
			KeyFindingTh:        "พบการรันโพรเซส Web Shell บน SRV-WEB01",
			ContainmentAction:   "Isolate SRV-WEB01 from the production VLAN immediately and block 185.220.101.45 at the perimeter firewall",
			ContainmentActionTh: "แยกโฮสต์ SRV-WEB01 ออกจาก production VLAN ทันทีและบล็อก 185.220.101.45 ที่ perimeter firewall",
			EradicationAction:   "Remove the uploaded web shell script from SRV-WEB01 webroot and patch the vulnerable upload handler",
			EradicationActionTh: "ลบสคริปต์ web shell ออกจาก webroot ของ SRV-WEB01 และแพตช์จุดอัปโหลดไฟล์ที่มีช่องโหว่",
			RecoveryAction:      "Rebuild SRV-WEB01 from a verified gold image and monitor outbound connections for 72 hours",
			RecoveryActionTh:    "สร้าง SRV-WEB01 ใหม่จาก gold image ที่ตรวจสอบแล้วและเฝ้าระวังการเชื่อมต่อขาออก 72 ชั่วโมง",
		}

		s := synthesizeFullScenario(micro, "domain2", "Threats, Vulnerabilities, and Mitigations")
		if err := ValidateScenarioCandidate(s); err != nil {
			t.Fatalf("expected synthesized scenario from bilingual micro struct to pass validation, got: %v", err)
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
			Title:             "Authorized Vulnerability Scanner Traffic",
			TitleTh:           "ทราฟฟิกจากระบบสแกนช่องโหว่ที่ได้รับอนุญาต",
			Description:       "High volume port scan from internal security scanner.",
			DescriptionTh:     "พบการสแกนพอร์ตจำนวนมากจากเครื่องสแกนช่องโหว่ภายในไปยัง SRV-APP02",
			IsTruePositive:    false,
			TpFpExplanation:   "False positive: scheduled internal Qualys scan.",
			TpFpExplanationTh: "เหตุการณ์นี้เป็น False Positive จากการสแกนตามกำหนดการของ Qualys ภายใน",
			AttackerIP:        "10.0.99.10",
			TargetHost:        "SRV-APP02",
			AttackLogMessage:  "Multiple SYN packets from 10.0.99.10 to SRV-APP02",
			KeyFinding:        "Traffic matches scheduled maintenance window CRQ-4012",
			KeyFindingTh:      "ทราฟฟิกตรงกับรอบบำรุงรักษาที่ได้รับอนุญาต CRQ-4012",
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
				t.Errorf("static scenario %q (%s) failed bilingual validation: %v", sc.Title, sc.DomainID, err)
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
		Translations: map[string]interface{}{
			"th": map[string]interface{}{
				"title":             "เหตุการณ์ Web Shell บนระบบเดิม",
				"description":       "ตรวจพบ shell ต้องสงสัยบน SRV-WEB01 เชื่อมต่อจาก 185.220.101.45",
				"tp_fp_explanation": "เป็น True Positive จากการรัน web shell",
				"findings": map[string]interface{}{
					"f1": map[string]interface{}{
						"description": "พบไฟล์ web shell บนโฮสต์",
						"explanation": "ยืนยันจากบันทึกการแตกโพรเซส",
					},
				},
				"playbook_steps": map[string]interface{}{
					"c1": map[string]interface{}{
						"label":       "แยกเครื่อง SRV-WEB01 ออกจากเครือข่ายทันที บล็อก IP ผู้โจมตี 185.220.101.45 และหยุดโพรเซส powershell.exe กับ w3wp.exe",
						"explanation": "หยุดความเสี่ยงด้านการควบคุมขอบเขตทันที",
					},
					"e1": map[string]interface{}{
						"label":       "ลบไฟล์ web shell ออกจาก webroot ติดตั้ง SRV-WEB01 ใหม่ แพตช์ช่องโหว่ และหมุนเวียน credentials ทั้งหมด",
						"explanation": "กำจัดกลไกการฝังตัวของภัยคุกคามอย่างถาวร",
					},
					"r1": map[string]interface{}{
						"label":       "กู้คืน SRV-WEB01 จาก backup ที่สะอาด เปิดการเฝ้าระวังเพิ่มเติมและตรวจสอบความสมบูรณ์ตลอด 72 ชั่วโมง",
						"explanation": "รับรองการคืนระบบงานอย่างปลอดภัย",
					},
				},
			},
		},
		InitialLogs:     make([]map[string]interface{}, 15),
		ExpectedOutcomes: map[string]interface{}{
			"key_findings": []map[string]interface{}{
				{"id": "f1", "description": "Web shell found", "domain_id": "domain2", "points": 10, "evidence_log_indices": []int{0}, "verdict_support": "malicious"},
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
		if strings.EqualFold(strings.TrimSpace(stored[i].Status), "archived") {
			continue
		}
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

func cloneLogSliceForTest(orig []map[string]interface{}) []map[string]interface{} {
	logs := make([]map[string]interface{}, len(orig))
	for i, entry := range orig {
		cp := make(map[string]interface{}, len(entry))
		for k, v := range entry {
			cp[k] = v
		}
		logs[i] = cp
	}
	return logs
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

func TestBuildScenarioDiversityContext_RotationHistoryAndCadence(t *testing.T) {
	t.Run("maps every canonical SY0-701 subtopic to TP and FP archetypes", func(t *testing.T) {
		for _, spec := range OrderedTheorySubtopics {
			tpCtx := BuildScenarioDiversityContext(spec.DomainID, nil)
			if tpCtx.TargetSubTopicID == "" || tpCtx.TargetArchetype == "" || tpCtx.TargetBriefingStyle == "" {
				t.Fatalf("empty diversity context for %s", spec.DomainID)
			}
			tpArch := archetypeForSubtopic(spec.SubTopicID, true)
			fpArch := archetypeForSubtopic(spec.SubTopicID, false)
			if tpArch == "" || fpArch == "" || tpArch == fpArch {
				t.Fatalf("expected distinct non-empty TP/FP archetypes for %s, got %q and %q", spec.SubTopicID, tpArch, fpArch)
			}
		}
	})

	t.Run("rotates across unvisited subtopics, briefing styles, and ~70/30 TP-FP cadence", func(t *testing.T) {
		var history []ScenarioHistoryEntry
		seenSubtopics := make(map[string]bool)
		seenStyles := make(map[string]bool)
		fpCount := 0
		tpCount := 0

		for step := 0; step < 9; step++ {
			ctx := BuildScenarioDiversityContext("domain4", history)
			if ctx.TargetSubTopicID == "" || ctx.TargetSubTopicTitle == "" {
				t.Fatalf("step %d: missing subtopic metadata", step)
			}
			if step < 9 {
				if seenSubtopics[ctx.TargetSubTopicID] {
					t.Fatalf("step %d: repeated subtopic %s before covering all 9 domain4 subtopics", step, ctx.TargetSubTopicID)
				}
				seenSubtopics[ctx.TargetSubTopicID] = true
			}
			seenStyles[ctx.TargetBriefingStyle] = true

			if ctx.TargetIsTruePositive {
				tpCount++
			} else {
				fpCount++
			}

			// Verify recent exclusions match history
			if step > 0 && len(ctx.RecentTitles) == 0 {
				t.Fatalf("step %d: expected RecentTitles populated", step)
			}

			// Simulate sequential history (most recent first)
			entry := ScenarioHistoryEntry{
				Title:          "Incident " + ctx.TargetSubTopicID,
				SubTopicID:     ctx.TargetSubTopicID,
				BriefingStyle:  ctx.TargetBriefingStyle,
				TargetHost:     "HOST-" + ctx.TargetSubTopicID,
				AttackerIP:     "198.51.100." + string(rune('1'+step)),
				IsTruePositive: ctx.TargetIsTruePositive,
			}
			history = append([]ScenarioHistoryEntry{entry}, history...)
		}

		if len(seenSubtopics) != 9 {
			t.Fatalf("expected all 9 domain4 subtopics covered, got %d", len(seenSubtopics))
		}
		if len(seenStyles) != 5 {
			t.Fatalf("expected all 5 briefing styles rotated, got %d (%v)", len(seenStyles), seenStyles)
		}
		if fpCount != 3 || tpCount != 6 {
			t.Fatalf("expected 6 TP and 3 FP (~67/33 cadence) over 9 runs, got TP=%d FP=%d", tpCount, fpCount)
		}
	})

	t.Run("prevents consecutive false positives and populates recent exclusions", func(t *testing.T) {
		history := []ScenarioHistoryEntry{
			{Title: "Recent Alpha", SubTopicID: "1.1", TargetHost: "DC-01", AttackerIP: "203.0.113.10", IsTruePositive: false, BriefingStyle: "shift_handover_brief"},
			{Title: "Recent Beta", SubTopicID: "1.2", TargetHost: "APP-02", AttackerIP: "198.51.100.22", IsTruePositive: true, BriefingStyle: "siem_escalation_ticket"},
		}
		ctx := BuildScenarioDiversityContext("domain1", history)
		if !ctx.TargetIsTruePositive {
			t.Fatalf("expected TargetIsTruePositive=true immediately after a false positive")
		}
		if ctx.TargetSubTopicID != "1.3" {
			t.Fatalf("expected next unvisited subtopic 1.3, got %s", ctx.TargetSubTopicID)
		}
		if ctx.TargetBriefingStyle == "shift_handover_brief" {
			t.Fatalf("expected briefing style to rotate away from most recent style")
		}
		if len(ctx.RecentTitles) != 2 || ctx.RecentTitles[0] != "Recent Alpha" {
			t.Fatalf("unexpected RecentTitles: %v", ctx.RecentTitles)
		}
		if len(ctx.RecentTargetHosts) != 2 || ctx.RecentTargetHosts[0] != "DC-01" {
			t.Fatalf("unexpected RecentTargetHosts: %v", ctx.RecentTargetHosts)
		}
		if len(ctx.RecentAttackerIPs) != 2 || ctx.RecentAttackerIPs[0] != "203.0.113.10" {
			t.Fatalf("unexpected RecentAttackerIPs: %v", ctx.RecentAttackerIPs)
		}
	})

	t.Run("rotates all 5 briefing styles across bounded 10-entry history and ignores SkipCadenceVote", func(t *testing.T) {
		var capped []ScenarioHistoryEntry
		for i := 0; i < 10; i++ {
			capped = append(capped, ScenarioHistoryEntry{
				Title:           fmt.Sprintf("Old %d", i),
				SubTopicID:      "4.1",
				BriefingStyle:   BriefingStyles[i%len(BriefingStyles)],
				SkipCadenceVote: true,
			})
		}

		// Latest resolved votes are 2 consecutive True Positives even though index 0 had SkipCadenceVote=true.
		capped[1].SkipCadenceVote = false
		capped[1].IsTruePositive = true
		capped[2].SkipCadenceVote = false
		capped[2].IsTruePositive = true

		seenCappedStyles := make(map[string]bool)
		lastStyle := capped[0].BriefingStyle
		for step := 0; step < 5; step++ {
			capped[0].BriefingStyle = lastStyle
			ctx := BuildScenarioDiversityContext("domain4", capped)
			seenCappedStyles[ctx.TargetBriefingStyle] = true
			if step == 0 && ctx.TargetIsTruePositive {
				t.Fatalf("expected TargetIsTruePositive=false after two SkipCadenceVote-filtered True Positives")
			}
			lastStyle = ctx.TargetBriefingStyle
		}
		if len(seenCappedStyles) != 5 {
			t.Fatalf("expected all 5 briefing styles rotated even when capped at 10 entries, got %d (%v)", len(seenCappedStyles), seenCappedStyles)
		}
	})
}

func TestSynthesizeFullScenario_NoHardcodedInvestigationCluesFooter(t *testing.T) {
	ai := aiMicroScenario{
		Title:           "API Gateway JWT Signing Key Confusion",
		Description:     "Escalation from API gateway monitoring: forged RS256/HS256 token accepted on api-edge-02 from 198.51.100.77.",
		SubTopicID:      "1.4",
		Archetype:       "jwt_key_confusion",
		BriefingStyle:   "siem_escalation_ticket",
		IsTruePositive:  true,
		TpFpExplanation: "Verified token signature validation flaw exploited from 198.51.100.77.",
		AttackerIP:      "198.51.100.77",
		TargetHost:      "api-edge-02",
		IncidentLogs: []aiIncidentLog{
			{Source: "nginx", Severity: "warning", Host: "api-edge-02", SrcIP: "198.51.100.77", Message: "POST /oauth/token returned 200 with unexpected HS256 alg header"},
			{Source: "edr", Severity: "high", Host: "api-edge-02", SrcIP: "198.51.100.77", Message: "Admin API scope granted without JWKS key match on api-edge-02"},
			{Source: "postgresql", Severity: "high", Host: "db-auth-01", SrcIP: "198.51.100.77", Message: "Bulk read of api_keys table triggered via forged bearer token"},
			{Source: "nids", Severity: "critical", Host: "api-edge-02", SrcIP: "198.51.100.77", Message: "Exfiltration response 2.4MB dispatched to 198.51.100.77"},
		},
	}

	scenario, errs := synthesizeFullScenarioWithErrors(ai, "domain1", "General Security Concepts")
	if len(errs) > 0 {
		t.Fatalf("unexpected synthesis errors: %v", errs)
	}
	if strings.Contains(scenario.Description, "INVESTIGATION CLUES:") {
		t.Fatalf("expected legacy INVESTIGATION CLUES footer to be removed, got %q", scenario.Description)
	}
}

func TestExtractPrimaryScenarioEntities_SkipsNoiseRowZero(t *testing.T) {
	sc := GenerateFallbackScenario("domain2", "Threats, Vulnerabilities, and Mitigations")
	host, ip := ExtractPrimaryScenarioEntities(sc)
	if host != "WS-FINANCE04" {
		t.Fatalf("ExtractPrimaryScenarioEntities host = %q, want WS-FINANCE04", host)
	}
	if ip != "185.220.101.99" {
		t.Fatalf("ExtractPrimaryScenarioEntities ip = %q, want 185.220.101.99", ip)
	}
}

func TestValidateScenarioCandidate_RejectsThaiInLogsAndIncompleteTranslations(t *testing.T) {
	t.Run("rejects Thai characters inside initial_logs", func(t *testing.T) {
		s := GenerateFallbackScenario("domain4", "Security Operations")
		logs := normalizePhaseStepSlice(s.InitialLogs)
		logs[0]["message"] = "แจ้งเตือนความผิดปกติจากโฮสต์ WS-DEV09"
		s.InitialLogs = logs
		if err := ValidateScenarioCandidate(s); err == nil {
			t.Fatalf("expected ValidateScenarioCandidate to reject Thai characters inside initial_logs")
		}
	})

	t.Run("rejects scenario missing translations.th", func(t *testing.T) {
		s := GenerateFallbackScenario("domain1", "General Security Concepts")
		s.Translations = nil
		if err := ValidateScenarioCandidate(s); err == nil {
			t.Fatalf("expected ValidateScenarioCandidate to reject scenario with missing translations.th")
		}
	})

	t.Run("rejects scenario with non-Thai text in translations.th.title", func(t *testing.T) {
		s := GenerateFallbackScenario("domain3", "Security Architecture")
		if transMap, ok := s.Translations.(map[string]interface{}); ok {
			if thMap, ok := transMap["th"].(map[string]interface{}); ok {
				thMap["title"] = "English Title Without Thai Characters"
			}
		}
		if err := ValidateScenarioCandidate(s); err == nil {
			t.Fatalf("expected ValidateScenarioCandidate to reject translations.th.title without Thai characters")
		}
	})

	t.Run("rejects scenario missing a playbook_steps action translation", func(t *testing.T) {
		s := GenerateFallbackScenario("domain2", "Threats, Vulnerabilities, and Mitigations")
		if transMap, ok := s.Translations.(map[string]interface{}); ok {
			if thMap, ok := transMap["th"].(map[string]interface{}); ok {
				if pbMap, ok := thMap["playbook_steps"].(map[string]map[string]string); ok {
					delete(pbMap, "c1")
				} else if pbAny, ok := thMap["playbook_steps"].(map[string]interface{}); ok {
					delete(pbAny, "c1")
				}
			}
		}
		if err := ValidateScenarioCandidate(s); err == nil {
			t.Fatalf("expected ValidateScenarioCandidate to reject missing translations.th.playbook_steps[c1]")
		}
	})
}

func TestGenerateFallbackScenario_BilingualAllDomainsComplete(t *testing.T) {
	domains := []string{"domain1", "domain2", "domain3", "domain4", "domain5"}
	for _, domainID := range domains {
		sc := GenerateFallbackScenario(domainID, domainID)
		transMap, ok := sc.Translations.(map[string]interface{})
		if !ok {
			t.Fatalf("%s: expected non-nil Translations map, got %T", domainID, sc.Translations)
		}
		thRaw, ok := transMap["th"]
		if !ok {
			t.Fatalf("%s: missing translations.th", domainID)
		}
		rawBytes, err := json.Marshal(thRaw)
		if err != nil {
			t.Fatalf("%s: marshal translations.th failed: %v", domainID, err)
		}
		var th struct {
			Title           string `json:"title"`
			Description     string `json:"description"`
			TpFpExplanation string `json:"tp_fp_explanation"`
			PlaybookSteps   map[string]struct {
				Label       string `json:"label"`
				Explanation string `json:"explanation"`
			} `json:"playbook_steps"`
			Findings map[string]struct {
				Description string `json:"description"`
				Explanation string `json:"explanation"`
			} `json:"findings"`
		}
		if err := json.Unmarshal(rawBytes, &th); err != nil {
			t.Fatalf("%s: unmarshal translations.th failed: %v", domainID, err)
		}
		if !ContainsThaiRune(th.Title) || !ContainsThaiRune(th.Description) || !ContainsThaiRune(th.TpFpExplanation) {
			t.Errorf("%s: expected Thai characters in title, description, and tp_fp_explanation", domainID)
		}
		if len(th.PlaybookSteps) < 9 {
			t.Errorf("%s: expected at least 9 translated playbook steps, got %d", domainID, len(th.PlaybookSteps))
		}
		if len(th.Findings) < 1 {
			t.Errorf("%s: expected at least 1 translated finding", domainID)
		}
	}
}

func TestTranslateScenarioToThai_Unit(t *testing.T) {
	base := GenerateFallbackScenario("domain4", "Security Operations")
	expectedTrans := normalizeAnyStringMap(base.Translations)
	expectedTh := normalizeAnyStringMap(expectedTrans["th"])
	legacy := *base
	legacy.Translations = nil

	t.Run("returns error cleanly when LLM_API_KEY is unset", func(t *testing.T) {
		t.Setenv("LLM_API_KEY", "")
		if _, err := TranslateScenarioToThai(context.Background(), &legacy); err == nil || !strings.Contains(err.Error(), "LLM_API_KEY") {
			t.Fatalf("expected LLM_API_KEY configuration error, got %v", err)
		}
	})

	t.Run("builds complete source schema prompt and validates fenced Thai JSON", func(t *testing.T) {
		rawThBytes, err := json.Marshal(expectedTh)
		if err != nil {
			t.Fatalf("failed to marshal expected Thai fixture: %v", err)
		}
		var promptSeen string
		merged, err := translateScenarioToThaiWithCaller(context.Background(), &legacy, func(ctx context.Context, messages []chatMessage, temperature float64) (string, error) {
			if len(messages) != 2 {
				t.Fatalf("expected 2 chat messages, got %d", len(messages))
			}
			promptSeen = messages[1].Content
			return "```json\n" + string(rawThBytes) + "\n```", nil
		})
		if err != nil {
			t.Fatalf("unexpected translation error: %v", err)
		}
		for _, id := range []string{"c1", "c2", "c3", "e1", "e2", "e3", "r1", "r2", "r3", "f1"} {
			if !strings.Contains(promptSeen, `"`+id+`"`) {
				t.Errorf("expected prompt to include ID %q", id)
			}
		}

		candidate := legacy
		candidate.Translations = merged
		if valErr := ValidateScenarioTranslationsTh(&candidate); valErr != nil {
			t.Fatalf("translated scenario failed ValidateScenarioTranslationsTh: %v", valErr)
		}
		if ContainsThaiRune(fmt.Sprintf("%v", candidate.InitialLogs)) {
			t.Fatalf("translation must never inject Thai into InitialLogs")
		}
	})

	t.Run("rejects incomplete or English-only playbook translations", func(t *testing.T) {
		missingStepJSON, err := json.Marshal(map[string]interface{}{
			"title":             expectedTh["title"],
			"description":       expectedTh["description"],
			"tp_fp_explanation": expectedTh["tp_fp_explanation"],
			"playbook_steps":    map[string]interface{}{},
			"findings":          expectedTh["findings"],
		})
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		if _, transErr := translateScenarioToThaiWithCaller(context.Background(), &legacy, func(ctx context.Context, messages []chatMessage, temperature float64) (string, error) {
			return string(missingStepJSON), nil
		}); transErr == nil {
			t.Fatalf("expected error when playbook_steps IDs are missing")
		}

		englishOnlyStepJSON, err := json.Marshal(map[string]interface{}{
			"title":             expectedTh["title"],
			"description":       expectedTh["description"],
			"tp_fp_explanation": expectedTh["tp_fp_explanation"],
			"playbook_steps": map[string]interface{}{
				"c1": map[string]string{
					"label":       "Isolate host WS-DEV09 immediately.",
					"explanation": "อธิบายภาษาไทย",
				},
			},
			"findings": expectedTh["findings"],
		})
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		if _, transErr := translateScenarioToThaiWithCaller(context.Background(), &legacy, func(ctx context.Context, messages []chatMessage, temperature float64) (string, error) {
			return string(englishOnlyStepJSON), nil
		}); transErr == nil || !strings.Contains(transErr.Error(), "Thai text") {
			t.Fatalf("expected Thai rune validation error for English step label, got %v", transErr)
		}
	})
}
