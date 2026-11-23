package controllers

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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

	t.Run("exposes only total_findings from expected outcomes", func(t *testing.T) {
		sanitized := sanitizeScenario(scenario, "sess-total-findings")
		if sanitized.TotalFindings != 1 {
			t.Fatalf("TotalFindings = %d, want 1", sanitized.TotalFindings)
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
		firstEvidenceIdx := firstEvidenceIndexForTest(t, scenario)
		baselineResult := services.ScoreSubmission(scenario, models.SubmitScenarioInput{
			SessionID:          "sess-score-baseline",
			IsTruePositive:     true,
			TPFPSelected:       true,
			SelectedActions:    []string{"c1", "e1", "r1"},
			SelectedLogIndices: []int{firstEvidenceIdx},
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
				SelectedLogIndices: []int{firstEvidenceIdx},
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

func firstEvidenceIndexForTest(t *testing.T, scenario models.Scenario) int {
	t.Helper()
	outcomes, ok := scenario.ExpectedOutcomes.(map[string]interface{})
	if !ok {
		t.Fatalf("expected_outcomes must be map")
	}
	findings := findingsFromOutcomes(outcomes["key_findings"])
	if len(findings) == 0 {
		t.Fatalf("expected at least 1 key finding")
	}
	switch indices := findings[0]["evidence_log_indices"].(type) {
	case []int:
		if len(indices) > 0 {
			return indices[0]
		}
	case []interface{}:
		if len(indices) > 0 {
			if v, ok := indices[0].(int); ok {
				return v
			}
			if v, ok := indices[0].(float64); ok {
				return int(v)
			}
		}
	}
	t.Fatalf("could not extract first evidence index from %v", findings[0])
	return 0
}

func TestBuildHistoryEntriesFromSessions_ExtractsSubtopicStyleAndEntities(t *testing.T) {
	scen1 := *services.GenerateFallbackScenario("domain2", "Threats, Vulnerabilities, and Mitigations")
	scen2 := *services.GenerateFallbackScenario("domain3", "Security Architecture")
	scen2.IsTruePositive = false

	sessions := []models.SimulationSession{
		{
			SessionID:       "sess-2",
			ScenarioID:      "sc-2",
			SelectionReason: "lowest_proficiency: domain3 (score: 55.0%, threshold: 70.0%) | subtopic: 3.2 | archetype: Approved Cloud Staging Bastion Troubleshooting | style: cloud_guardrail_alert",
		},
		{
			SessionID:       "sess-1",
			ScenarioID:      "sc-1",
			SelectionReason: "lowest_proficiency: domain2 (score: 40.0%, threshold: 70.0%) | subtopic: 2.2 | archetype: Spearphishing Attachment & Living-off-the-Land Execution | style: helpdesk_user_report",
		},
	}

	entries := buildHistoryEntriesFromSessions(sessions, func(sess models.SimulationSession) (models.Scenario, bool) {
		if sess.SessionID == "sess-2" {
			return scen2, true
		}
		return scen1, true
	})

	if len(entries) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(entries))
	}
	if entries[0].SubTopicID != "3.2" || entries[0].BriefingStyle != "cloud_guardrail_alert" || entries[0].IsTruePositive {
		t.Fatalf("unexpected first history entry: %+v", entries[0])
	}
	if entries[1].SubTopicID != "2.2" || entries[1].BriefingStyle != "helpdesk_user_report" || entries[1].TargetHost != "WS-FINANCE04" || entries[1].AttackerIP != "185.220.101.99" {
		t.Fatalf("unexpected second history entry: %+v", entries[1])
	}

	unresolved := buildHistoryEntriesFromSessions(sessions[:1], func(models.SimulationSession) (models.Scenario, bool) {
		return models.Scenario{}, false
	})
	if len(unresolved) != 1 || !unresolved[0].SkipCadenceVote {
		t.Fatalf("expected unresolved session scenario to mark SkipCadenceVote=true, got %+v", unresolved)
	}
}

func TestSimulationController_LockUserStartSerializesSameUser(t *testing.T) {
	sc := &SimulationController{}

	t.Run("serializes concurrent calls for the same user and cleans up lock entries", func(t *testing.T) {
		var inFlight atomic.Int32
		var maxInFlight atomic.Int32
		var wg sync.WaitGroup

		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				unlock := sc.lockUserStart("user-same")
				cur := inFlight.Add(1)
				for {
					prev := maxInFlight.Load()
					if cur <= prev || maxInFlight.CompareAndSwap(prev, cur) {
						break
					}
				}
				time.Sleep(15 * time.Millisecond)
				inFlight.Add(-1)
				unlock()
			}()
		}
		wg.Wait()

		if maxInFlight.Load() != 1 {
			t.Fatalf("expected max concurrent calls for same user to be 1, got %d", maxInFlight.Load())
		}
		sc.startMu.Lock()
		remaining := len(sc.userStartLocks)
		sc.startMu.Unlock()
		if remaining != 0 {
			t.Fatalf("expected userStartLocks map to clean up after unlock, got %d entry(ies)", remaining)
		}
	})

	t.Run("allows distinct users to start scenarios concurrently", func(t *testing.T) {
		enteredBoth := make(chan struct{})
		var arrived atomic.Int32
		var wg sync.WaitGroup

		for _, uid := range []string{"user-a", "user-b"} {
			uid := uid
			wg.Add(1)
			go func() {
				defer wg.Done()
				unlock := sc.lockUserStart(uid)
				defer unlock()
				if arrived.Add(1) == 2 {
					close(enteredBoth)
				}
				select {
				case <-enteredBoth:
				case <-time.After(1 * time.Second):
					t.Errorf("distinct users blocked each other on %s", uid)
				}
			}()
		}
		wg.Wait()
	})
}

func TestSanitizeScenario_SanitizesTranslationsWithoutLeaks(t *testing.T) {
	scenario := models.Scenario{
		ScenarioID:      "scen-bilingual-sanitize-01",
		Title:           "MFA Push Fatigue Attack",
		Description:     "Multiple denied MFA pushes detected for user ajones.",
		DomainID:        "domain1",
		Difficulty:      2,
		IsTruePositive:  true,
		TpFpExplanation: "True Positive MFA fatigue",
		InitialLogs: []map[string]interface{}{
			{"timestamp": "2026-07-13T08:00:00Z", "source": "okta", "severity": "high", "message": "MFA push denied"},
		},
		ExpectedOutcomes: map[string]interface{}{
			"key_findings": []map[string]interface{}{
				{"id": "d1_f1", "description": "Repeated MFA denials", "points": 10},
			},
		},
		PlaybookSteps: map[string]interface{}{
			"containment": []map[string]interface{}{
				{"id": "d1_c1", "label": "Revoke active Okta sessions", "is_correct": true, "explanation": "Stops attacker session"},
				{"id": "d1_c2", "label": "Disable firewall completely", "is_correct": false, "explanation": "Dangerous distractor"},
				{"id": "d1_c3", "label": "Reset monitor brightness", "is_correct": false, "explanation": "Unrelated"},
			},
			"eradication": []map[string]interface{}{
				{"id": "d1_e1", "label": "Enforce FIDO2 keys", "is_correct": true, "explanation": "Phishing-resistant"},
				{"id": "d1_e2", "label": "Delete audit logs", "is_correct": false, "explanation": "Destroys evidence"},
				{"id": "d1_e3", "label": "Wipe smartphone", "is_correct": false, "explanation": "Wrong target"},
			},
			"recovery": []map[string]interface{}{
				{"id": "d1_r1", "label": "Rotate credentials and monitor", "is_correct": true, "explanation": "Safe recovery"},
				{"id": "d1_r2", "label": "Disable MFA for 48 hours", "is_correct": false, "explanation": "Weakens security"},
				{"id": "d1_r3", "label": "Reuse old password hash", "is_correct": false, "explanation": "Allows re-entry"},
			},
		},
		Translations: map[string]interface{}{
			"th": map[string]interface{}{
				"title":             "การโจมตี MFA Push Fatigue",
				"description":       "ตรวจพบการแจ้งเตือน MFA ที่ถูกปฏิเสธซ้ำๆ สำหรับผู้ใช้ ajones",
				"tp_fp_explanation": "คำอธิบายเฉลย True Positive ห้ามหลุดก่อนส่งคำตอบ",
				"findings": map[string]interface{}{
					"d1_f1": map[string]interface{}{
						"description": "ข้อความเฉลยจุดสังเกตที่ห้ามหลุด",
						"explanation": "คำอธิบายจุดสังเกตที่ห้ามหลุด",
					},
				},
				"playbook_steps": map[string]interface{}{
					"d1_c1": map[string]interface{}{"label": "เพิกถอน Okta session ทั้งหมดทันที", "explanation": "เหตุผลเฉลยที่ห้ามหลุด", "is_correct": true},
					"d1_c2": map[string]interface{}{"label": "ปิดไฟร์วอลล์ทั้งหมด", "explanation": "เหตุผลเฉลยที่ห้ามหลุด"},
					"d1_c3": map[string]interface{}{"label": "ปรับความสว่างหน้าจอ", "explanation": "เหตุผลเฉลยที่ห้ามหลุด"},
					"d1_e1": map[string]interface{}{"label": "บังคับใช้ FIDO2 WebAuthn", "explanation": "เหตุผลเฉลยที่ห้ามหลุด"},
					"d1_e2": map[string]interface{}{"label": "ลบ Audit Logs", "explanation": "เหตุผลเฉลยที่ห้ามหลุด"},
					"d1_e3": map[string]interface{}{"label": "ล้างเครื่องโทรศัพท์", "explanation": "เหตุผลเฉลยที่ห้ามหลุด"},
					"d1_r1": map[string]interface{}{"label": "หมุนเวียนรหัสผ่านและเฝ้าระวัง", "explanation": "เหตุผลเฉลยที่ห้ามหลุด"},
					"d1_r2": map[string]interface{}{"label": "ยกเว้น MFA 48 ชั่วโมง", "explanation": "เหตุผลเฉลยที่ห้ามหลุด"},
					"d1_r3": map[string]interface{}{"label": "ใช้รหัสผ่านเดิมซ้ำ", "explanation": "เหตุผลเฉลยที่ห้ามหลุด"},
				},
			},
		},
	}

	sanitized := sanitizeScenario(scenario, "sess-bilingual-01")
	transMap, ok := sanitized.Translations.(map[string]interface{})
	if !ok {
		t.Fatalf("expected sanitized Translations to be present map[string]interface{}, got %T", sanitized.Translations)
	}
	thMap, ok := transMap["th"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected sanitized translations.th map, got %T", transMap["th"])
	}
	if thMap["title"] != "การโจมตี MFA Push Fatigue" {
		t.Errorf("unexpected sanitized th title: %v", thMap["title"])
	}
	if thMap["description"] != "ตรวจพบการแจ้งเตือน MFA ที่ถูกปฏิเสธซ้ำๆ สำหรับผู้ใช้ ajones" {
		t.Errorf("unexpected sanitized th description: %v", thMap["description"])
	}
	if _, leaked := thMap["tp_fp_explanation"]; leaked {
		t.Fatalf("answer leak: sanitized translations.th exposed tp_fp_explanation")
	}
	if _, leaked := thMap["findings"]; leaked {
		t.Fatalf("answer leak: sanitized translations.th exposed findings")
	}
	pbMap, ok := thMap["playbook_steps"].(map[string]interface{})
	if !ok || len(pbMap) != 9 {
		t.Fatalf("expected 9 sanitized th playbook_steps entries, got %T (len=%d)", thMap["playbook_steps"], len(pbMap))
	}
	for id, rawStep := range pbMap {
		stepMap, ok := rawStep.(map[string]interface{})
		if !ok {
			t.Fatalf("step %s is not map: %T", id, rawStep)
		}
		if stepMap["label"] == "" {
			t.Errorf("step %s missing label", id)
		}
		if _, leaked := stepMap["explanation"]; leaked {
			t.Fatalf("answer leak: step %s exposed explanation", id)
		}
		if _, leaked := stepMap["is_correct"]; leaked {
			t.Fatalf("answer leak: step %s exposed is_correct", id)
		}
	}
}

func TestEvaluateStage2TheoryGate_IncludesCanonicalCode(t *testing.T) {
	gateErr := evaluateStage2TheoryGate("", "learner", func() ([]string, []string) {
		return []string{"1.1"}, []string{}
	})
	if gateErr == nil {
		t.Fatal("expected TheoryGateErrorResponse")
	}
	if gateErr.Code != "SIM_THEORY_INCOMPLETE" {
		t.Fatalf("expected Code SIM_THEORY_INCOMPLETE, got %q", gateErr.Code)
	}
	if gateErr.Params["completed_subtopics"] != 1 || gateErr.Params["total_subtopics"] != 28 {
		t.Fatalf("unexpected Params: %+v", gateErr.Params)
	}
}

func TestEnsureScenarioStoryAndTranslation(t *testing.T) {
	translatedFixture := services.GenerateFallbackScenario("domain4", "Security Operations")
	translatedFixture.ScenarioID = "scen-legacy-test-01"

	t.Run("no-op when scenario already has complete EN+TH story beats", func(t *testing.T) {
		alreadyComplete := *translatedFixture
		var calls atomic.Int32
		sc := &SimulationController{
			upgradeStoryFn: func(ctx context.Context, s *models.Scenario) (string, map[string]any, error) {
				calls.Add(1)
				return "", map[string]any{}, nil
			},
		}
		sc.ensureScenarioStoryAndTranslation(context.Background(), &alreadyComplete)
		if calls.Load() != 0 {
			t.Fatalf("expected 0 upgrader calls when story and Thai translations are already complete, got %d", calls.Load())
		}
	})

	t.Run("fails open cleanly when LLM_API_KEY is unset", func(t *testing.T) {
		t.Setenv("LLM_API_KEY", "")
		legacy := *translatedFixture
		legacy.Description = "Old abrupt technical line."
		legacy.Translations = nil
		sc := &SimulationController{}
		sc.ensureScenarioStoryAndTranslation(context.Background(), &legacy)
		if legacy.Translations != nil || legacy.Description != "Old abrupt technical line." {
			t.Fatalf("expected legacy scenario to remain unchanged when LLM_API_KEY is unset")
		}
	})

	t.Run("upgrades legacy scenario to 4-beat EN+TH story and exposes sanitized non-spoiler fields", func(t *testing.T) {
		legacy := *translatedFixture
		legacy.Description = "Threat hunting flagged outbound beacons on WS-DEV09."
		legacy.Translations = nil
		expectedTrans, ok := translatedFixture.Translations.(map[string]any)
		if !ok {
			if typedMap, ok2 := translatedFixture.Translations.(map[string]interface{}); ok2 {
				expectedTrans = typedMap
			} else {
				t.Fatalf("unexpected fixture Translations type %T", translatedFixture.Translations)
			}
		}

		var calls atomic.Int32
		sc := &SimulationController{
			upgradeStoryFn: func(ctx context.Context, s *models.Scenario) (string, map[string]any, error) {
				calls.Add(1)
				return translatedFixture.Description, expectedTrans, nil
			},
		}

		sc.ensureScenarioStoryAndTranslation(context.Background(), &legacy)
		if calls.Load() != 1 {
			t.Fatalf("expected exactly 1 upgrader call, got %d", calls.Load())
		}
		if err := services.ValidateScenarioTranslationsTh(&legacy); err != nil {
			t.Fatalf("expected valid Thai translations after story backfill, got error: %v", err)
		}
		if !services.HasCompleteStory(&legacy) {
			t.Fatalf("expected complete EN+TH story beats after backfill")
		}

		// Calling again immediately on the upgraded scenario must not re-invoke LLM.
		sc.ensureScenarioStoryAndTranslation(context.Background(), &legacy)
		if calls.Load() != 1 {
			t.Fatalf("expected cached scenario not to re-run story upgrade, calls=%d", calls.Load())
		}

		sanitized := sanitizeScenario(legacy, "sess-backfill-01")
		if len(strings.Split(sanitized.Description, "\n\n")) != 3 {
			t.Fatalf("expected 3-paragraph EN description in sanitized response, got %q", sanitized.Description)
		}
		transMap, ok := sanitized.Translations.(map[string]interface{})
		if !ok {
			t.Fatalf("expected sanitized Translations map, got %T", sanitized.Translations)
		}
		thMap, ok := transMap["th"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected sanitized translations.th, got %T", transMap["th"])
		}
		thDesc, _ := thMap["description"].(string)
		if len(strings.Split(thDesc, "\n\n")) != 3 {
			t.Fatalf("expected 3-paragraph TH description in sanitized translations.th, got %q", thDesc)
		}
		if _, leaked := thMap["tp_fp_explanation"]; leaked {
			t.Fatalf("sanitized translations.th leaked tp_fp_explanation")
		}
		if _, leaked := thMap["findings"]; leaked {
			t.Fatalf("sanitized translations.th leaked findings")
		}
	})
}
