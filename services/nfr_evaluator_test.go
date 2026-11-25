package services

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"cybersim/dto"
)

func boolPtrNFR(v bool) *bool       { return &v }
func floatPtrNFR(v float64) *float64 { return &v }
func intPtrNFR(v int) *int          { return &v }

func healthyBaseInput() NFREvaluateInput {
	avail := 100.0
	err5xx := 0.0
	p50 := 18.0
	p95 := 45.0
	p99 := 90.0
	mean := 22.0
	minV := 3.0
	maxV := 120.0
	csuq := 6.10
	tlx := 32.0
	examMed := 900.0
	simMed := 420.0

	return NFREvaluateInput{
		RunLabel:    "Healthy Verification",
		Environment: "production",
		Source:      "live",
		EvaluatedAt: time.Now().UTC(),
		Thresholds:  DefaultNFRThresholds(),
		Runtime: dto.NFRRuntimeSnapshotDTO{
			UptimeSeconds:          1200,
			GinMode:                "release",
			TotalRequests:          100,
			NonLLMRequests:         98,
			LLMRequests:            2,
			Status2xx:              98,
			Status4xx:              2,
			Status5xx:              0,
			AvailabilityPct:        &avail,
			ErrorRate5xxPct:        &err5xx,
			PeakConcurrentRequests: 8,
			DBStatus:               "up",
			DBLatencyMs:            4.2,
			NonLLMLatency: dto.NFRPercentilesDTO{
				Count: 98, Min: &minV, Mean: &mean, P50: &p50, P95: &p95, P99: &p99, Max: &maxV,
			},
			LLMLatency: dto.NFRPercentilesDTO{Count: 0},
		},
		Facts: NFRRawDatabaseFacts{
			DBStatus:               "up",
			DBLatencyMs:            4.2,
			ActiveQuestionsCount:   540,
			TheorySubtopicsCount:   28,
			ThaiTheoryCount:        33,
			ActiveScenariosCount:   15,
			TotalSimSessions:       10,
			SessionsWithSnapshot:   10,
			SnapshotsCompleteMeta:  10,
			AIGeneratedSessions:    7,
			StaticFallbackSessions: 3,
			SystemFailureSessions:  0,
			ValidSnapshotCount:     10,
			TotalSnapshotCount:     10,
			FallbackReasons:        map[string]int{"LLM_API_KEY not configured": 3},
			CSUQOverallMean:        &csuq,
			NASATLXOverallMean:     &tlx,
			QuestionnaireCount:     5,
			ExamTotalCount:         20,
			ExamCompletedCount:     19,
			ExamMedianSeconds:      &examMed,
			SimFinishedCount:       10,
			SimCompletedCount:      10,
			SimMedianSeconds:       &simMed,
		},
		SecurityProbes: []dto.NFRCheckItemDTO{
			{Key: "unauth_admin_nfr_blocked", Measured: "401", Unit: "status", Target: "401", Status: "pass"},
			{Key: "forged_jwt_rejected", Measured: "401", Unit: "status", Target: "401", Status: "pass"},
			{Key: "learner_rbac_denied", Measured: "403", Unit: "status", Target: "403", Status: "pass"},
			{Key: "specialist_rbac_denied", Measured: "403", Unit: "status", Target: "403", Status: "pass"},
			{Key: "unauth_exam_blocked", Measured: "401", Unit: "status", Target: "401", Status: "pass"},
			{Key: "sqli_query_param_safe", Measured: "200", Unit: "status", Target: "< 500", Status: "pass"},
		},
		Posture: NFRPostureInput{
			JWTSecretSet:            true,
			JWTSecretIsDefault:      false,
			JWTSecretLength:         48,
			SwaggerEnabled:          false,
			BypassPostTestThreshold: false,
			CORSWildcard:            false,
			GinMode:                 "release",
		},
		Scalability: &dto.NFRScalabilityBenchmarkDTO{
			Mode:       "cli_benchmark",
			ExecutedAt: time.Now().UTC(),
			Levels: []dto.NFRScalabilityLevelDTO{
				{Concurrency: 10, TotalRequests: 100, SuccessCount: 100, ErrorRatePct: 0, ThroughputRPS: 320, P50Ms: 20, P95Ms: 48, P99Ms: 70},
				{Concurrency: 25, TotalRequests: 100, SuccessCount: 100, ErrorRatePct: 0, ThroughputRPS: 410, P50Ms: 32, P95Ms: 88, P99Ms: 120},
			},
		},
		Maintainability: &dto.NFRMaintainabilityReportDTO{
			CollectedAt:              time.Now().UTC(),
			GoTestPassed:             boolPtrNFR(true),
			GoCoveragePct:            floatPtrNFR(66.5),
			FrontendTestPassed:       boolPtrNFR(true),
			FrontendTscPassed:        boolPtrNFR(true),
			EvidenceGatePassed:       boolPtrNFR(true),
			TheoryGatePassed:         boolPtrNFR(true),
			IntegrationHarnessPassed: boolPtrNFR(true),
			GovulncheckHighCritical:  intPtrNFR(0),
			NpmAuditHighCritical:     intPtrNFR(0),
			GitleaksFindings:         intPtrNFR(0),
		},
	}
}

func findDim(t *testing.T, rep dto.NFRReportResponse, id string) dto.NFRDimensionResultDTO {
	t.Helper()
	for _, d := range rep.Dimensions {
		if d.ID == id {
			return d
		}
	}
	t.Fatalf("dimension %q not found", id)
	return dto.NFRDimensionResultDTO{}
}

func TestEvaluateNFR_AllTenPassAndRollupPrecedence(t *testing.T) {
	in := healthyBaseInput()
	rep := EvaluateNFR(in)

	if len(rep.Dimensions) != 10 {
		t.Fatalf("expected 10 ISO/IEC 25010 dimensions, got %d", len(rep.Dimensions))
	}
	if rep.Summary.OverallStatus != "all_passed" {
		t.Fatalf("expected all_passed, got %s (pass=%d warn=%d fail=%d pending=%d)",
			rep.Summary.OverallStatus, rep.Summary.PassedCount, rep.Summary.WarnCount, rep.Summary.FailCount, rep.Summary.NotMeasuredCount)
	}
	if rep.Summary.MeasuredPassRate != 100.0 {
		t.Fatalf("expected 100.0 measured pass rate, got %v", rep.Summary.MeasuredPassRate)
	}

	rawJSON, err := json.Marshal(rep)
	if err != nil {
		t.Fatalf("json marshal failed: %v", err)
	}
	if strings.Contains(string(rawJSON), "NaN") || strings.Contains(string(rawJSON), "Infinity") {
		t.Fatalf("JSON contains NaN or Infinity")
	}

	// Partial pending when scalability & maintainability are nil and usability n == 0
	partialIn := healthyBaseInput()
	partialIn.Scalability = nil
	partialIn.Maintainability = nil
	partialIn.Facts.QuestionnaireCount = 0
	partialIn.Facts.CSUQOverallMean = nil
	partialIn.Facts.NASATLXOverallMean = nil
	partialRep := EvaluateNFR(partialIn)
	if partialRep.Summary.OverallStatus != "partial_pending" {
		t.Fatalf("expected partial_pending, got %s", partialRep.Summary.OverallStatus)
	}
	if partialRep.Summary.NotMeasuredCount != 3 || partialRep.Summary.FailCount != 0 {
		t.Fatalf("expected 3 pending / 0 fail, got pending=%d fail=%d", partialRep.Summary.NotMeasuredCount, partialRep.Summary.FailCount)
	}
	if findDim(t, partialRep, "scalability").Status != "not_measured" {
		t.Fatalf("expected scalability=not_measured")
	}
	if findDim(t, partialRep, "usability").Status != "insufficient_data" {
		t.Fatalf("expected usability=insufficient_data")
	}
	// Pass rate must be computed strictly over measured (7/7 = 100%)
	if partialRep.Summary.MeasuredPassRate != 100.0 {
		t.Fatalf("expected 100.0%% over 7 measured dimensions, got %v", partialRep.Summary.MeasuredPassRate)
	}

	// Warn overrides partial_pending ONLY? Wait: Q13=A precedence: fail > partial_pending > pass_with_warnings > all_passed
	warnIn := healthyBaseInput()
	warnIn.Posture.SwaggerEnabled = true // causes security warn
	warnRep := EvaluateNFR(warnIn)
	if warnRep.Summary.OverallStatus != "pass_with_warnings" {
		t.Fatalf("expected pass_with_warnings, got %s", warnRep.Summary.OverallStatus)
	}
}

func TestEvaluateNFR_ThresholdsAndSafetyViolations(t *testing.T) {
	// 1. Pre/Post overlap triggers Data Integrity FAIL and overall FAIL
	failIn := healthyBaseInput()
	failIn.Facts.PrePostOverlapCount = 1
	failRep := EvaluateNFR(failIn)
	if findDim(t, failRep, "data_integrity").Status != "fail" || failRep.Summary.OverallStatus != "fail" {
		t.Fatalf("expected data_integrity and overall fail on question overlap")
	}

	// 2. Safety violations fail Safety dimension
	safetyIn := healthyBaseInput()
	safetyIn.Facts.ThaiViolationInLogs = 2
	safetyIn.Facts.SecretLeakSnapshots = 1
	safetyRep := EvaluateNFR(safetyIn)
	if findDim(t, safetyRep, "safety").Status != "fail" {
		t.Fatalf("expected safety fail on Thai logs & secret leaks")
	}

	// 3. Security CVE count > 0 fails Security even if probes pass
	cveIn := healthyBaseInput()
	cveIn.Maintainability.GovulncheckHighCritical = intPtrNFR(1)
	cveRep := EvaluateNFR(cveIn)
	if findDim(t, cveRep, "security").Status != "fail" {
		t.Fatalf("expected security fail when govulncheck_high_critical > 0")
	}

	// 4. Light preset concurrency (<10 workers) gives Scalability WARN when latency/errors pass
	lightIn := healthyBaseInput()
	lightIn.Scalability = &dto.NFRScalabilityBenchmarkDTO{
		Mode:   "in_app_probe",
		Preset: "light",
		Levels: []dto.NFRScalabilityLevelDTO{
			{Concurrency: 5, TotalRequests: 50, SuccessCount: 50, ErrorRatePct: 0, ThroughputRPS: 190, P95Ms: 42},
		},
	}
	lightRep := EvaluateNFR(lightIn)
	if findDim(t, lightRep, "scalability").Status != "warn" {
		t.Fatalf("expected scalability warn for <10 concurrency light preset, got %s", findDim(t, lightRep, "scalability").Status)
	}

	// 5. Threshold env override works
	t.Setenv("NFR_P95_TARGET_MS", "40")
	overridden := LoadNFRThresholdsFromEnv()
	if overridden.NonLLMP95TargetMs != 40 {
		t.Fatalf("expected env override NFR_P95_TARGET_MS=40, got %v", overridden.NonLLMP95TargetMs)
	}
}

func TestNFRSafetyScanners_ZeroSecretAndThaiAndBoilerplate(t *testing.T) {
	if !SnapshotContainsSecretPattern(`{"debug":"Bearer eyJhbGciOi..."}`, "", "") {
		t.Fatalf("expected Bearer pattern detection")
	}
	if !SnapshotContainsSecretPattern(`{"key":"sk-proj-1234567890"}`, "", "") {
		t.Fatalf("expected sk- API key pattern detection")
	}
	if !SnapshotContainsSecretPattern(`{"leak":"my-super-secret-value-xyz"}`, "", "my-super-secret-value-xyz") {
		t.Fatalf("expected active JWT secret detection")
	}
	if SnapshotContainsSecretPattern(`{"host":"SRV-WEB01","message":"Accepted publickey"}`, "", "real-secret-123456") {
		t.Fatalf("clean snapshot should not trigger secret warning")
	}
}
