package services

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"cybersim/dto"
)

const (
	nfrStatusPass             = "pass"
	nfrStatusWarn             = "warn"
	nfrStatusFail             = "fail"
	nfrStatusNotMeasured      = "not_measured"
	nfrStatusInsufficientData = "insufficient_data"
)

var (
	secretBearerJWTRe = regexp.MustCompile(`(?i)bearer\s+eyj[a-z0-9._-]{8,}`)
	secretAPIKeyRe    = regexp.MustCompile(`(?i)\bsk-(?:proj-|ant-|live-|test-)?[a-z0-9_-]{8,}`)
	secretGitHubRe    = regexp.MustCompile(`\b(?:ghp|gho|ghu|ghs)_[A-Za-z0-9_]{12,}`)
)

// NFRRawDatabaseFacts holds objective database audit counts gathered from PostgreSQL.
type NFRRawDatabaseFacts struct {
	DBStatus                        string
	DBLatencyMs                     float64
	NonBcryptUserCount              int
	PrePostOverlapCount             int
	OrphanSimSessionCount           int
	NullScoreCompletedExams         int
	ActiveQuestionsCount            int
	TheorySubtopicsCount            int
	ThaiTheoryCount                 int
	ActiveScenariosCount            int
	TotalSimSessions                int
	SessionsWithSnapshot            int
	SnapshotsCompleteMeta           int
	AIGeneratedSessions             int
	StaticFallbackSessions          int
	SystemFailureSessions           int
	ValidSnapshotCount              int
	TotalSnapshotCount              int
	FallbackReasons                 map[string]int
	CSUQOverallMean                 *float64
	NASATLXOverallMean              *float64
	QuestionnaireCount              int
	ExamTotalCount                  int
	ExamCompletedCount              int
	ExamMedianSeconds               *float64
	SimFinishedCount                int
	SimCompletedCount               int
	SimMedianSeconds                *float64
	ThaiViolationInExams            int
	ThaiViolationInLogs             int
	SecretLeakSnapshots             int
	BannedBoilerplateScenarios      int
	EvidenceRatioViolationScenarios int
}

// NFREvaluateInput bundles telemetry, DB audit facts, probes, and optional benchmark payloads.
type NFREvaluateInput struct {
	RunID           *string
	RunLabel        string
	Environment     string
	Source          string
	Notes           string
	EvaluatedAt     time.Time
	Thresholds      dto.NFRThresholdsDTO
	Runtime         dto.NFRRuntimeSnapshotDTO
	Facts           NFRRawDatabaseFacts
	SecurityProbes  []dto.NFRCheckItemDTO
	Posture         NFRPostureInput
	Scalability     *dto.NFRScalabilityBenchmarkDTO
	Maintainability *dto.NFRMaintainabilityReportDTO
}

// SnapshotContainsSecretPattern checks serialized snapshot data for leaked API keys, bearer tokens, or active secrets.
func SnapshotContainsSecretPattern(snapshotJSON, activeLLMKey, activeJWTSecret string) bool {
	if snapshotJSON == "" {
		return false
	}
	if secretBearerJWTRe.MatchString(snapshotJSON) || secretAPIKeyRe.MatchString(snapshotJSON) || secretGitHubRe.MatchString(snapshotJSON) {
		return true
	}
	for _, secret := range []string{strings.TrimSpace(activeLLMKey), strings.TrimSpace(activeJWTSecret)} {
		if len(secret) >= 12 && strings.Contains(snapshotJSON, secret) {
			return true
		}
	}
	return false
}

// ScenarioTextHasBannedBoilerplate checks both initial_logs and playbook_steps text for prohibited legacy strings.
func ScenarioTextHasBannedBoilerplate(logsRaw []map[string]interface{}, playbookRaw map[string]interface{}) bool {
	if _, found := containsBannedLegacyBoilerplate(logsRaw); found {
		return true
	}
	if playbookRaw != nil {
		for _, phase := range []string{"containment", "eradication", "recovery"} {
			for _, step := range normalizePhaseStepSlice(playbookRaw[phase]) {
				label, _ := step["label"].(string)
				if IsBannedBoilerplateLabel(label) {
					return true
				}
			}
		}
	}
	return false
}

// EvaluateNFR deterministically scores all 10 ISO/IEC 25010 quality characteristics under nfr-rubric-v1.
func EvaluateNFR(in NFREvaluateInput) dto.NFRReportResponse {
	th := in.Thresholds
	if th.RubricVersion == "" {
		th = DefaultNFRThresholds()
	}
	evaluatedAt := in.EvaluatedAt.UTC()
	if evaluatedAt.IsZero() {
		evaluatedAt = time.Now().UTC()
	}
	runLabel := strings.TrimSpace(in.RunLabel)
	if runLabel == "" {
		runLabel = "Live System Evaluation"
	}
	envName := strings.TrimSpace(in.Environment)
	if envName == "" {
		envName = "local"
	}
	source := strings.TrimSpace(in.Source)
	if source == "" {
		source = "live"
	}

	fallbackReasons := in.Facts.FallbackReasons
	if fallbackReasons == nil {
		fallbackReasons = map[string]int{}
	}

	dimensions := []dto.NFRDimensionResultDTO{
		evalPerformance(in.Runtime, th),
		evalAvailability(in.Runtime, in.Facts, th),
		evalScalability(in.Runtime, in.Scalability, th),
		evalReliability(in.Runtime, in.Facts),
		evalSecurity(in.SecurityProbes, in.Posture, in.Facts, in.Maintainability),
		evalDataIntegrity(in.Facts),
		evalReproducibility(in.Facts),
		evalUsability(in.Facts, th),
		evalMaintainability(in.Maintainability, th),
		evalSafety(in.Facts),
	}

	summary := rollupNFRSummary(dimensions)

	return dto.NFRReportResponse{
		RunID:         in.RunID,
		RunLabel:      runLabel,
		Environment:   envName,
		Source:        source,
		RubricVersion: th.RubricVersion,
		EvaluatedAt:   evaluatedAt,
		Notes:         strings.TrimSpace(in.Notes),
		Summary:       summary,
		Runtime:       in.Runtime,
		Dimensions:    dimensions,
		Scalability:   in.Scalability,
		ReliabilityBreakdown: dto.NFRReliabilityBreakdownDTO{
			TotalSessions:       in.Facts.TotalSimSessions,
			AIGeneratedCount:    in.Facts.AIGeneratedSessions,
			StaticFallbackCount: in.Facts.StaticFallbackSessions,
			SystemFailureCount:  in.Facts.SystemFailureSessions,
			ValidSnapshotCount:  in.Facts.ValidSnapshotCount,
			TotalSnapshotCount:  in.Facts.TotalSnapshotCount,
			FallbackReasons:     fallbackReasons,
		},
		MaintainabilityReport: in.Maintainability,
		Thresholds:            th,
	}
}

func rollupNFRSummary(dims []dto.NFRDimensionResultDTO) dto.NFRSummaryDTO {
	s := dto.NFRSummaryDTO{
		TotalDimensions: len(dims),
	}
	for _, d := range dims {
		switch d.Status {
		case nfrStatusPass:
			s.PassedCount++
		case nfrStatusWarn:
			s.WarnCount++
		case nfrStatusFail:
			s.FailCount++
		default:
			s.NotMeasuredCount++
		}
	}
	measured := s.PassedCount + s.WarnCount + s.FailCount
	if measured > 0 {
		s.MeasuredPassRate = RoundTelemetry2(float64(s.PassedCount) / float64(measured) * 100.0)
	}

	switch {
	case s.FailCount > 0:
		s.OverallStatus = "fail"
	case s.NotMeasuredCount > 0:
		s.OverallStatus = "partial_pending"
	case s.WarnCount > 0:
		s.OverallStatus = "pass_with_warnings"
	default:
		s.OverallStatus = "all_passed"
	}
	return s
}

func worstCheckStatus(checks []dto.NFRCheckItemDTO, allowPending bool) string {
	hasWarn := false
	hasPass := false
	hasPending := false
	for _, c := range checks {
		switch c.Status {
		case nfrStatusFail:
			return nfrStatusFail
		case nfrStatusWarn:
			hasWarn = true
		case nfrStatusPass:
			hasPass = true
		case nfrStatusNotMeasured, nfrStatusInsufficientData:
			hasPending = true
		}
	}
	if hasWarn {
		return nfrStatusWarn
	}
	if !allowPending && hasPending && !hasPass {
		return nfrStatusInsufficientData
	}
	if hasPass {
		return nfrStatusPass
	}
	return nfrStatusNotMeasured
}

func fmtPtrMs(p *float64) string {
	if p == nil || math.IsNaN(*p) || math.IsInf(*p, 0) {
		return "N/A"
	}
	return fmt.Sprintf("%.2f", *p)
}

func evalLatencyUpperBound(val *float64, passMax, warnMax float64) string {
	if val == nil {
		return nfrStatusInsufficientData
	}
	if *val <= passMax {
		return nfrStatusPass
	}
	if *val <= warnMax {
		return nfrStatusWarn
	}
	return nfrStatusFail
}

func evalPerformance(rt dto.NFRRuntimeSnapshotDTO, th dto.NFRThresholdsDTO) dto.NFRDimensionResultDTO {
	dbStatus := nfrStatusFail
	if rt.DBStatus == "up" {
		if rt.DBLatencyMs <= th.DBPingTargetMs {
			dbStatus = nfrStatusPass
		} else if rt.DBLatencyMs <= th.DBPingWarnMs {
			dbStatus = nfrStatusWarn
		}
	}

	p95Status := evalLatencyUpperBound(rt.NonLLMLatency.P95, th.NonLLMP95TargetMs, th.NonLLMP95WarnMs)
	p99Status := evalLatencyUpperBound(rt.NonLLMLatency.P99, th.NonLLMP99TargetMs, th.NonLLMP99WarnMs)
	llmP95Status := evalLatencyUpperBound(rt.LLMLatency.P95, th.LLMP95TargetMs, th.LLMP95WarnMs)

	checks := []dto.NFRCheckItemDTO{
		{
			Key:      "non_llm_p95_ms",
			Measured: fmtPtrMs(rt.NonLLMLatency.P95),
			Unit:     "ms",
			Target:   fmt.Sprintf("<= %.0f ms (warn <= %.0f)", th.NonLLMP95TargetMs, th.NonLLMP95WarnMs),
			Status:   p95Status,
		},
		{
			Key:      "non_llm_p99_ms",
			Measured: fmtPtrMs(rt.NonLLMLatency.P99),
			Unit:     "ms",
			Target:   fmt.Sprintf("<= %.0f ms (warn <= %.0f)", th.NonLLMP99TargetMs, th.NonLLMP99WarnMs),
			Status:   p99Status,
		},
		{
			Key:      "db_ping_ms",
			Measured: fmt.Sprintf("%.2f", rt.DBLatencyMs),
			Unit:     "ms",
			Target:   fmt.Sprintf("<= %.0f ms (warn <= %.0f)", th.DBPingTargetMs, th.DBPingWarnMs),
			Status:   dbStatus,
		},
		{
			Key:      "llm_generation_p95_ms",
			Measured: fmtPtrMs(rt.LLMLatency.P95),
			Unit:     "ms",
			Target:   fmt.Sprintf("<= %.0f ms (warn <= %.0f)", th.LLMP95TargetMs, th.LLMP95WarnMs),
			Status:   llmP95Status,
		},
	}

	dimStatus := worstCheckStatus([]dto.NFRCheckItemDTO{checks[0], checks[1], checks[2]}, true)
	if llmP95Status == nfrStatusFail {
		dimStatus = nfrStatusFail
	} else if llmP95Status == nfrStatusWarn && dimStatus == nfrStatusPass {
		dimStatus = nfrStatusWarn
	}
	if dbStatus == nfrStatusPass && rt.NonLLMLatency.Count == 0 && dimStatus == nfrStatusPass {
		dimStatus = nfrStatusInsufficientData
	}

	summary := fmt.Sprintf("DB %.1f ms (n=0 API reqs)", rt.DBLatencyMs)
	if rt.NonLLMLatency.P95 != nil {
		summary = fmt.Sprintf("p95 %.1f ms | p99 %.1f ms | DB %.1f ms", *rt.NonLLMLatency.P95, *rt.NonLLMLatency.P99, rt.DBLatencyMs)
	}

	return dto.NFRDimensionResultDTO{
		ID:                "performance",
		ISOCharacteristic: "Performance Efficiency (Time Behaviour)",
		Status:            dimStatus,
		SummaryMetric:     summary,
		TargetExpression:  fmt.Sprintf("API p95 <= %.0f ms, p99 <= %.0f ms, DB <= %.0f ms", th.NonLLMP95TargetMs, th.NonLLMP99TargetMs, th.DBPingTargetMs),
		EvidenceSource:    "Passive HTTP Ring Buffer + LLM Recorder + Live DB Ping",
		Checks:            checks,
	}
}

func evalAvailability(rt dto.NFRRuntimeSnapshotDTO, facts NFRRawDatabaseFacts, th dto.NFRThresholdsDTO) dto.NFRDimensionResultDTO {
	dbOK := rt.DBStatus == "up" && facts.DBStatus == "up"
	dbCheckStatus := nfrStatusFail
	if dbOK {
		dbCheckStatus = nfrStatusPass
	}

	httpAvailStatus := nfrStatusInsufficientData
	availValStr := "N/A"
	if rt.AvailabilityPct != nil {
		availValStr = fmt.Sprintf("%.2f", *rt.AvailabilityPct)
		switch {
		case *rt.AvailabilityPct >= th.AvailabilityTargetPct:
			httpAvailStatus = nfrStatusPass
		case *rt.AvailabilityPct >= th.AvailabilityWarnPct:
			httpAvailStatus = nfrStatusWarn
		default:
			httpAvailStatus = nfrStatusFail
		}
	}

	checks := []dto.NFRCheckItemDTO{
		{
			Key:      "db_connectivity",
			Measured: rt.DBStatus,
			Unit:     "state",
			Target:   "up",
			Status:   dbCheckStatus,
		},
		{
			Key:      "http_non_5xx_pct",
			Measured: availValStr,
			Unit:     "%",
			Target:   fmt.Sprintf(">= %.1f%% (warn >= %.1f%%)", th.AvailabilityTargetPct, th.AvailabilityWarnPct),
			Status:   httpAvailStatus,
		},
		{
			Key:      "process_uptime_seconds",
			Measured: fmt.Sprintf("%d", rt.UptimeSeconds),
			Unit:     "s",
			Target:   "Continuous runtime",
			Status:   nfrStatusPass,
		},
	}

	dimStatus := worstCheckStatus([]dto.NFRCheckItemDTO{checks[0], checks[1]}, true)
	summary := fmt.Sprintf("DB %s | Uptime %ds", rt.DBStatus, rt.UptimeSeconds)
	if rt.AvailabilityPct != nil {
		summary = fmt.Sprintf("%.2f%% Non-5xx | DB %s", *rt.AvailabilityPct, rt.DBStatus)
	}

	return dto.NFRDimensionResultDTO{
		ID:                "availability",
		ISOCharacteristic: "Reliability (Availability)",
		Status:            dimStatus,
		SummaryMetric:     summary,
		TargetExpression:  fmt.Sprintf("DB up & Non-5xx >= %.1f%%", th.AvailabilityTargetPct),
		EvidenceSource:    "Live PostgreSQL SELECT 1 Ping + HTTP Status Counters",
		Checks:            checks,
	}
}

func evalScalability(rt dto.NFRRuntimeSnapshotDTO, bench *dto.NFRScalabilityBenchmarkDTO, th dto.NFRThresholdsDTO) dto.NFRDimensionResultDTO {
	if bench == nil || len(bench.Levels) == 0 {
		return dto.NFRDimensionResultDTO{
			ID:                "scalability",
			ISOCharacteristic: "Performance Efficiency (Capacity)",
			Status:            nfrStatusNotMeasured,
			SummaryMetric:     fmt.Sprintf("Not Measured (peak observed: %d)", rt.PeakConcurrentRequests),
			TargetExpression:  fmt.Sprintf("Concurrency >= 10, p95 <= %.0f ms, Error <= %.1f%%", th.ScalabilityP95TargetMs, th.ScalabilityErrorTargetPct),
			EvidenceSource:    "Safe In-App Concurrency Probe or CLI Benchmark",
			Checks: []dto.NFRCheckItemDTO{
				{
					Key:      "concurrency_benchmark_runs",
					Measured: "0",
					Unit:     "runs",
					Target:   ">= 1 probe or CLI curve",
					Status:   nfrStatusNotMeasured,
				},
				{
					Key:      "passive_peak_concurrency",
					Measured: fmt.Sprintf("%d", rt.PeakConcurrentRequests),
					Unit:     "req",
					Target:   "Observed in-flight peak",
					Status:   nfrStatusInsufficientData,
				},
			},
		}
	}

	maxConc := 0
	worstErr := 0.0
	peakLevel := bench.Levels[0]
	peakRPS := 0.0
	for _, lv := range bench.Levels {
		if lv.ErrorRatePct > worstErr {
			worstErr = lv.ErrorRatePct
		}
		if lv.ThroughputRPS > peakRPS {
			peakRPS = lv.ThroughputRPS
		}
		if lv.Concurrency >= maxConc {
			maxConc = lv.Concurrency
			peakLevel = lv
		}
	}

	errStatus := nfrStatusFail
	if worstErr <= th.ScalabilityErrorTargetPct {
		errStatus = nfrStatusPass
	} else if worstErr <= th.ScalabilityErrorWarnPct {
		errStatus = nfrStatusWarn
	}

	p95Status := nfrStatusFail
	if peakLevel.P95Ms <= th.ScalabilityP95TargetMs {
		p95Status = nfrStatusPass
	} else if peakLevel.P95Ms <= th.ScalabilityP95WarnMs {
		p95Status = nfrStatusWarn
	}

	depthStatus := nfrStatusPass
	if maxConc < 10 {
		depthStatus = nfrStatusWarn
	}

	checks := []dto.NFRCheckItemDTO{
		{
			Key:      "scalability_max_concurrency",
			Measured: fmt.Sprintf("%d", maxConc),
			Unit:     "workers",
			Target:   ">= 10 workers (light preset = warn)",
			Status:   depthStatus,
		},
		{
			Key:      "scalability_p95_ms",
			Measured: fmt.Sprintf("%.2f", peakLevel.P95Ms),
			Unit:     "ms",
			Target:   fmt.Sprintf("<= %.0f ms (warn <= %.0f)", th.ScalabilityP95TargetMs, th.ScalabilityP95WarnMs),
			Status:   p95Status,
		},
		{
			Key:      "scalability_error_rate_pct",
			Measured: fmt.Sprintf("%.2f", worstErr),
			Unit:     "%",
			Target:   fmt.Sprintf("<= %.1f%% (warn <= %.1f%%)", th.ScalabilityErrorTargetPct, th.ScalabilityErrorWarnPct),
			Status:   errStatus,
		},
		{
			Key:      "scalability_peak_rps",
			Measured: fmt.Sprintf("%.1f", peakRPS),
			Unit:     "req/s",
			Target:   "Sustained read throughput",
			Status:   nfrStatusPass,
		},
	}

	return dto.NFRDimensionResultDTO{
		ID:                "scalability",
		ISOCharacteristic: "Performance Efficiency (Capacity)",
		Status:            worstCheckStatus(checks, true),
		SummaryMetric:     fmt.Sprintf("%d workers | %.1f RPS | p95 %.1f ms", maxConc, peakRPS, peakLevel.P95Ms),
		TargetExpression:  fmt.Sprintf(">= 10 workers, p95 <= %.0f ms, Error <= %.1f%%", th.ScalabilityP95TargetMs, th.ScalabilityErrorTargetPct),
		EvidenceSource:    fmt.Sprintf("Benchmark (%s)", bench.Mode),
		Checks:            checks,
	}
}

func evalReliability(rt dto.NFRRuntimeSnapshotDTO, facts NFRRawDatabaseFacts) dto.NFRDimensionResultDTO {
	errStatus := nfrStatusPass
	errMeasured := "0.00"
	if rt.ErrorRate5xxPct != nil {
		errMeasured = fmt.Sprintf("%.2f", *rt.ErrorRate5xxPct)
		switch {
		case *rt.ErrorRate5xxPct <= 1.0:
			errStatus = nfrStatusPass
		case *rt.ErrorRate5xxPct <= 5.0:
			errStatus = nfrStatusWarn
		default:
			errStatus = nfrStatusFail
		}
	} else {
		errStatus = nfrStatusInsufficientData
	}

	sysFailStatus := nfrStatusInsufficientData
	sysFailMeasured := "N/A"
	if facts.TotalSimSessions > 0 {
		rate := RoundTelemetry2(float64(facts.SystemFailureSessions) / float64(facts.TotalSimSessions) * 100.0)
		sysFailMeasured = fmt.Sprintf("%.2f", rate)
		if facts.SystemFailureSessions == 0 {
			sysFailStatus = nfrStatusPass
		} else {
			sysFailStatus = nfrStatusFail
		}
	}

	snapStatus := nfrStatusInsufficientData
	snapMeasured := "N/A"
	if facts.TotalSnapshotCount > 0 {
		rate := RoundTelemetry2(float64(facts.ValidSnapshotCount) / float64(facts.TotalSnapshotCount) * 100.0)
		snapMeasured = fmt.Sprintf("%.2f", rate)
		switch {
		case rate == 100.0:
			snapStatus = nfrStatusPass
		case rate >= 95.0:
			snapStatus = nfrStatusWarn
		default:
			snapStatus = nfrStatusFail
		}
	}

	checks := []dto.NFRCheckItemDTO{
		{
			Key:      "http_5xx_error_rate_pct",
			Measured: errMeasured,
			Unit:     "%",
			Target:   "<= 1.0% (warn <= 5.0%)",
			Status:   errStatus,
		},
		{
			Key:      "simulation_system_failure_pct",
			Measured: sysFailMeasured,
			Unit:     "%",
			Target:   "0.0% (RQ4 zero system failures)",
			Status:   sysFailStatus,
		},
		{
			Key:      "valid_snapshot_delivery_pct",
			Measured: snapMeasured,
			Unit:     "%",
			Target:   "100.0% (AI + Static Fallback valid)",
			Status:   snapStatus,
		},
	}

	dimStatus := worstCheckStatus(checks, true)
	summary := fmt.Sprintf("5xx %s%% | SysFail %d/%d | Snap %d/%d",
		errMeasured, facts.SystemFailureSessions, facts.TotalSimSessions, facts.ValidSnapshotCount, facts.TotalSnapshotCount)

	return dto.NFRDimensionResultDTO{
		ID:                "reliability",
		ISOCharacteristic: "Reliability (Fault Tolerance & Recoverability)",
		Status:            dimStatus,
		SummaryMetric:     summary,
		TargetExpression:  "HTTP 5xx <= 1.0%, RQ4 system_failure = 0.0%, Valid Snapshots = 100%",
		EvidenceSource:    "HTTP 5xx Ring + simulation_sessions & scenario_snapshots",
		Checks:            checks,
	}
}

func evalSecurity(
	probes []dto.NFRCheckItemDTO,
	posture NFRPostureInput,
	facts NFRRawDatabaseFacts,
	maint *dto.NFRMaintainabilityReportDTO,
) dto.NFRDimensionResultDTO {
	checks := make([]dto.NFRCheckItemDTO, 0, len(probes)+9)
	checks = append(checks, probes...)

	jwtStatus := nfrStatusPass
	jwtMeasured := fmt.Sprintf("set (%d chars)", posture.JWTSecretLength)
	if !posture.JWTSecretSet || posture.JWTSecretIsDefault {
		jwtMeasured = "default/unset"
		if posture.GinMode == "release" {
			jwtStatus = nfrStatusFail
		} else {
			jwtStatus = nfrStatusWarn
		}
	} else if posture.JWTSecretLength < 32 {
		jwtMeasured = fmt.Sprintf("short (%d chars)", posture.JWTSecretLength)
		jwtStatus = nfrStatusWarn
	}

	swaggerStatus := nfrStatusPass
	if posture.SwaggerEnabled {
		swaggerStatus = nfrStatusWarn
	}
	bypassStatus := nfrStatusPass
	if posture.BypassPostTestThreshold {
		bypassStatus = nfrStatusWarn
	}
	corsStatus := nfrStatusPass
	if posture.CORSWildcard {
		corsStatus = nfrStatusFail
	}
	bcryptStatus := nfrStatusPass
	if facts.NonBcryptUserCount > 0 {
		bcryptStatus = nfrStatusFail
	}
	ginStatus := nfrStatusPass
	if posture.GinMode != "release" {
		ginStatus = nfrStatusWarn
	}

	checks = append(checks,
		dto.NFRCheckItemDTO{
			Key:      "posture_jwt_secret",
			Measured: jwtMeasured,
			Unit:     "posture",
			Target:   "non-default & >= 32 chars",
			Status:   jwtStatus,
		},
		dto.NFRCheckItemDTO{
			Key:      "posture_swagger_disabled",
			Measured: fmt.Sprintf("disabled=%t", !posture.SwaggerEnabled),
			Unit:     "posture",
			Target:   "disabled in production",
			Status:   swaggerStatus,
		},
		dto.NFRCheckItemDTO{
			Key:      "posture_bypass_disabled",
			Measured: fmt.Sprintf("disabled=%t", !posture.BypassPostTestThreshold),
			Unit:     "posture",
			Target:   "disabled in production",
			Status:   bypassStatus,
		},
		dto.NFRCheckItemDTO{
			Key:      "posture_cors_strict",
			Measured: fmt.Sprintf("wildcard=%t", posture.CORSWildcard),
			Unit:     "posture",
			Target:   "no * origin wildcard",
			Status:   corsStatus,
		},
		dto.NFRCheckItemDTO{
			Key:      "posture_bcrypt_hashes",
			Measured: fmt.Sprintf("%d non-bcrypt", facts.NonBcryptUserCount),
			Unit:     "users",
			Target:   "0 non-bcrypt password hashes",
			Status:   bcryptStatus,
		},
		dto.NFRCheckItemDTO{
			Key:      "posture_gin_release",
			Measured: posture.GinMode,
			Unit:     "mode",
			Target:   "release",
			Status:   ginStatus,
		},
	)

	addScanCheck := func(key string, val *int) {
		if val == nil {
			checks = append(checks, dto.NFRCheckItemDTO{
				Key:      key,
				Measured: "Not attached",
				Unit:     "findings",
				Target:   "0 High/Critical",
				Status:   nfrStatusNotMeasured,
			})
			return
		}
		st := nfrStatusPass
		if *val > 0 {
			st = nfrStatusFail
		}
		checks = append(checks, dto.NFRCheckItemDTO{
			Key:      key,
			Measured: fmt.Sprintf("%d", *val),
			Unit:     "findings",
			Target:   "0 High/Critical",
			Status:   st,
		})
	}

	var govuln, npmAudit, gitleaks *int
	if maint != nil {
		govuln = maint.GovulncheckHighCritical
		npmAudit = maint.NpmAuditHighCritical
		gitleaks = maint.GitleaksFindings
	}
	addScanCheck("scan_govulncheck_high_critical", govuln)
	addScanCheck("scan_npm_audit_high_critical", npmAudit)
	addScanCheck("scan_gitleaks_findings", gitleaks)

	probePass := 0
	for _, p := range probes {
		if p.Status == nfrStatusPass {
			probePass++
		}
	}

	dimStatus := worstCheckStatus(checks, true)
	return dto.NFRDimensionResultDTO{
		ID:                "security",
		ISOCharacteristic: "Security (Authenticity, Integrity, Confidentiality)",
		Status:            dimStatus,
		SummaryMetric:     fmt.Sprintf("%d/%d Auth/RBAC Probes | Non-bcrypt %d", probePass, len(probes), facts.NonBcryptUserCount),
		TargetExpression:  "100% Auth/RBAC/Input probes, bcrypt hashes, strict CORS/JWT, 0 High CVEs",
		EvidenceSource:    "Read-Only httptest Probes + Zero-Secret Runtime Posture + DB Users Audit",
		Checks:            checks,
	}
}

func evalDataIntegrity(facts NFRRawDatabaseFacts) dto.NFRDimensionResultDTO {
	overlapStatus := nfrStatusPass
	if facts.PrePostOverlapCount > 0 {
		overlapStatus = nfrStatusFail
	}

	orphanCount := facts.OrphanSimSessionCount + facts.NullScoreCompletedExams
	orphanStatus := nfrStatusPass
	if orphanCount > 0 {
		orphanStatus = nfrStatusFail
	}

	seedOK := facts.ActiveQuestionsCount >= 540 &&
		facts.TheorySubtopicsCount >= 28 &&
		facts.ThaiTheoryCount >= 33 &&
		facts.ActiveScenariosCount >= 15
	seedStatus := nfrStatusPass
	if !seedOK {
		seedStatus = nfrStatusWarn
	}

	checks := []dto.NFRCheckItemDTO{
		{
			Key:      "pre_post_question_overlap",
			Measured: fmt.Sprintf("%d", facts.PrePostOverlapCount),
			Unit:     "violations",
			Target:   "0 overlaps per learner",
			Status:   overlapStatus,
		},
		{
			Key:      "orphaned_or_null_records",
			Measured: fmt.Sprintf("%d", orphanCount),
			Unit:     "records",
			Target:   "0 orphaned/null completed records",
			Status:   orphanStatus,
		},
		{
			Key:      "curriculum_seed_completeness",
			Measured: fmt.Sprintf("Q:%d Sub:%d TH:%d Scen:%d", facts.ActiveQuestionsCount, facts.TheorySubtopicsCount, facts.ThaiTheoryCount, facts.ActiveScenariosCount),
			Unit:     "counts",
			Target:   "Q>=540, Sub>=28, TH>=33, Scen>=15",
			Status:   seedStatus,
		},
	}

	return dto.NFRDimensionResultDTO{
		ID:                "data_integrity",
		ISOCharacteristic: "Functional Suitability & Integrity",
		Status:            worstCheckStatus(checks, false),
		SummaryMetric:     fmt.Sprintf("Overlap %d | Orphans %d | Q=%d", facts.PrePostOverlapCount, orphanCount, facts.ActiveQuestionsCount),
		TargetExpression:  "0 Pre/Post overlaps, 0 orphans, 540 Qs / 28 Subtopics / 33 TH / 15+ Scenarios",
		EvidenceSource:    "Live PostgreSQL Relational Integrity & Curriculum Seed Audit",
		Checks:            checks,
	}
}

func evalReproducibility(facts NFRRawDatabaseFacts) dto.NFRDimensionResultDTO {
	if facts.TotalSimSessions == 0 {
		return dto.NFRDimensionResultDTO{
			ID:                "reproducibility",
			ISOCharacteristic: "Maintainability / Reliability (Auditability & Replayability)",
			Status:            nfrStatusInsufficientData,
			SummaryMetric:     "Insufficient Data (0 simulation sessions)",
			TargetExpression:  "100% sessions linked to complete immutable snapshots",
			EvidenceSource:    "PostgreSQL simulation_sessions JOIN scenario_snapshots",
			Checks: []dto.NFRCheckItemDTO{
				{Key: "snapshot_linkage_pct", Measured: "N/A", Unit: "%", Target: "100.0%", Status: nfrStatusInsufficientData},
				{Key: "complete_snapshot_metadata_pct", Measured: "N/A", Unit: "%", Target: "100.0%", Status: nfrStatusInsufficientData},
			},
		}
	}

	linkagePct := RoundTelemetry2(float64(facts.SessionsWithSnapshot) / float64(facts.TotalSimSessions) * 100.0)
	metaPct := 0.0
	if facts.TotalSnapshotCount > 0 {
		metaPct = RoundTelemetry2(float64(facts.SnapshotsCompleteMeta) / float64(facts.TotalSnapshotCount) * 100.0)
	}

	pctStatus := func(v float64) string {
		switch {
		case v == 100.0:
			return nfrStatusPass
		case v >= 95.0:
			return nfrStatusWarn
		default:
			return nfrStatusFail
		}
	}

	checks := []dto.NFRCheckItemDTO{
		{
			Key:      "snapshot_linkage_pct",
			Measured: fmt.Sprintf("%.2f", linkagePct),
			Unit:     "%",
			Target:   "100.0% (warn >= 95.0%)",
			Status:   pctStatus(linkagePct),
		},
		{
			Key:      "complete_snapshot_metadata_pct",
			Measured: fmt.Sprintf("%.2f", metaPct),
			Unit:     "%",
			Target:   "100.0% (source, prompt, blueprint, rubric, validation)",
			Status:   pctStatus(metaPct),
		},
	}

	return dto.NFRDimensionResultDTO{
		ID:                "reproducibility",
		ISOCharacteristic: "Maintainability / Reliability (Auditability & Replayability)",
		Status:            worstCheckStatus(checks, false),
		SummaryMetric:     fmt.Sprintf("Link %.1f%% (%d/%d) | Meta %.1f%%", linkagePct, facts.SessionsWithSnapshot, facts.TotalSimSessions, metaPct),
		TargetExpression:  "100% immutable snapshot coverage & complete version metadata",
		EvidenceSource:    "PostgreSQL simulation_sessions & scenario_snapshots Audit",
		Checks:            checks,
	}
}

func evalUsability(facts NFRRawDatabaseFacts, th dto.NFRThresholdsDTO) dto.NFRDimensionResultDTO {
	checks := make([]dto.NFRCheckItemDTO, 0, 4)

	csuqStatus := nfrStatusInsufficientData
	csuqStr := "N/A"
	if facts.QuestionnaireCount > 0 && facts.CSUQOverallMean != nil {
		csuqStr = fmt.Sprintf("%.2f", *facts.CSUQOverallMean)
		switch {
		case *facts.CSUQOverallMean >= th.CSUQTargetMean:
			csuqStatus = nfrStatusPass
		case *facts.CSUQOverallMean >= th.CSUQWarnMean:
			csuqStatus = nfrStatusWarn
		default:
			csuqStatus = nfrStatusFail
		}
	}

	tlxStatus := nfrStatusInsufficientData
	tlxStr := "N/A"
	if facts.QuestionnaireCount > 0 && facts.NASATLXOverallMean != nil {
		tlxStr = fmt.Sprintf("%.2f", *facts.NASATLXOverallMean)
		switch {
		case *facts.NASATLXOverallMean <= th.TLXTargetMax:
			tlxStatus = nfrStatusPass
		case *facts.NASATLXOverallMean <= th.TLXWarnMax:
			tlxStatus = nfrStatusWarn
		default:
			tlxStatus = nfrStatusFail
		}
	}

	examRateStr := "N/A"
	examCheckStatus := nfrStatusInsufficientData
	if facts.ExamTotalCount > 0 {
		rate := RoundTelemetry2(float64(facts.ExamCompletedCount) / float64(facts.ExamTotalCount) * 100.0)
		examRateStr = fmt.Sprintf("%.1f%% (%ds med)", rate, int(ptrValOrZero(facts.ExamMedianSeconds)))
		examCheckStatus = nfrStatusPass
	}

	simRateStr := "N/A"
	simCheckStatus := nfrStatusInsufficientData
	if facts.SimFinishedCount > 0 {
		rate := RoundTelemetry2(float64(facts.SimCompletedCount) / float64(facts.SimFinishedCount) * 100.0)
		simRateStr = fmt.Sprintf("%.1f%% (%ds med)", rate, int(ptrValOrZero(facts.SimMedianSeconds)))
		simCheckStatus = nfrStatusPass
	}

	checks = append(checks,
		dto.NFRCheckItemDTO{
			Key:      "csuq_overall_mean",
			Measured: csuqStr,
			Unit:     "/ 7.0",
			Target:   fmt.Sprintf(">= %.1f (warn >= %.1f, n=%d)", th.CSUQTargetMean, th.CSUQWarnMean, facts.QuestionnaireCount),
			Status:   csuqStatus,
		},
		dto.NFRCheckItemDTO{
			Key:      "nasa_tlx_overall_mean",
			Measured: tlxStr,
			Unit:     "/ 100",
			Target:   fmt.Sprintf("<= %.0f (warn <= %.0f, n=%d)", th.TLXTargetMax, th.TLXWarnMax, facts.QuestionnaireCount),
			Status:   tlxStatus,
		},
		dto.NFRCheckItemDTO{
			Key:      "exam_task_completion",
			Measured: examRateStr,
			Unit:     "rate/time",
			Target:   "Observed exam completion & median duration",
			Status:   examCheckStatus,
		},
		dto.NFRCheckItemDTO{
			Key:      "simulation_task_completion",
			Measured: simRateStr,
			Unit:     "rate/time",
			Target:   "Observed scenario completion & median duration",
			Status:   simCheckStatus,
		},
	)

	dimStatus := nfrStatusInsufficientData
	summary := "Insufficient Data (0 questionnaires)"
	if facts.QuestionnaireCount > 0 && facts.CSUQOverallMean != nil && facts.NASATLXOverallMean != nil {
		dimStatus = worstCheckStatus([]dto.NFRCheckItemDTO{checks[0], checks[1]}, false)
		summary = fmt.Sprintf("CSUQ %.2f/7 | TLX %.1f/100 (n=%d)", *facts.CSUQOverallMean, *facts.NASATLXOverallMean, facts.QuestionnaireCount)
	}

	return dto.NFRDimensionResultDTO{
		ID:                "usability",
		ISOCharacteristic: "Interaction Capability (Usability — #13 Bridge)",
		Status:            dimStatus,
		SummaryMetric:     summary,
		TargetExpression:  fmt.Sprintf("CSUQ >= %.1f/7.0 & Raw NASA-TLX <= %.0f/100", th.CSUQTargetMean, th.TLXTargetMax),
		EvidenceSource:    "evaluation_metrics (#13 CSUQ/NASA-TLX) + exam_sessions & simulation_sessions",
		Checks:            checks,
	}
}

func ptrValOrZero(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

func evalMaintainability(maint *dto.NFRMaintainabilityReportDTO, th dto.NFRThresholdsDTO) dto.NFRDimensionResultDTO {
	if maint == nil {
		return dto.NFRDimensionResultDTO{
			ID:                "maintainability",
			ISOCharacteristic: "Maintainability (Testability & Modifiability)",
			Status:            nfrStatusNotMeasured,
			SummaryMetric:     "Not Measured (awaiting CI / benchmark JSON)",
			TargetExpression:  fmt.Sprintf("100%% CI gates pass & Go coverage >= %.0f%%", th.CoverageTargetPct),
			EvidenceSource:    "Ingested CLI / CI Test & Coverage Gate Report",
			Checks: []dto.NFRCheckItemDTO{
				{
					Key:      "ci_gate_report",
					Measured: "Not attached",
					Unit:     "report",
					Target:   "Attach via CLI --include-ci or JSON import",
					Status:   nfrStatusNotMeasured,
				},
			},
		}
	}

	checks := make([]dto.NFRCheckItemDTO, 0, 6)
	boolCheck := func(key string, val *bool) {
		if val == nil {
			checks = append(checks, dto.NFRCheckItemDTO{
				Key: key, Measured: "N/A", Unit: "gate", Target: "pass", Status: nfrStatusNotMeasured,
			})
			return
		}
		st := nfrStatusFail
		res := "fail"
		if *val {
			st = nfrStatusPass
			res = "pass"
		}
		checks = append(checks, dto.NFRCheckItemDTO{
			Key: key, Measured: res, Unit: "gate", Target: "pass", Status: st,
		})
	}

	boolCheck("go_unit_tests", maint.GoTestPassed)
	if maint.GoCoveragePct == nil {
		checks = append(checks, dto.NFRCheckItemDTO{
			Key:      "go_test_coverage_pct",
			Measured: "N/A",
			Unit:     "%",
			Target:   fmt.Sprintf(">= %.0f%% (warn >= %.0f%%)", th.CoverageTargetPct, th.CoverageWarnPct),
			Status:   nfrStatusNotMeasured,
		})
	} else {
		covSt := nfrStatusFail
		if *maint.GoCoveragePct >= th.CoverageTargetPct {
			covSt = nfrStatusPass
		} else if *maint.GoCoveragePct >= th.CoverageWarnPct {
			covSt = nfrStatusWarn
		}
		checks = append(checks, dto.NFRCheckItemDTO{
			Key:      "go_test_coverage_pct",
			Measured: fmt.Sprintf("%.1f", *maint.GoCoveragePct),
			Unit:     "%",
			Target:   fmt.Sprintf(">= %.0f%% (warn >= %.0f%%)", th.CoverageTargetPct, th.CoverageWarnPct),
			Status:   covSt,
		})
	}
	boolCheck("frontend_unit_tests", maint.FrontendTestPassed)
	boolCheck("frontend_tsc_build", maint.FrontendTscPassed)
	boolCheck("evidence_authoring_gate", maint.EvidenceGatePassed)
	boolCheck("theory_bilingual_gate", maint.TheoryGatePassed)
	if maint.IntegrationHarnessPassed != nil {
		boolCheck("integration_harness", maint.IntegrationHarnessPassed)
	}

	covSummary := "N/A"
	if maint.GoCoveragePct != nil {
		covSummary = fmt.Sprintf("%.1f%%", *maint.GoCoveragePct)
	}
	passGates := 0
	totalGates := 0
	for _, c := range checks {
		if c.Unit == "gate" && c.Status != nfrStatusNotMeasured {
			totalGates++
			if c.Status == nfrStatusPass {
				passGates++
			}
		}
	}

	return dto.NFRDimensionResultDTO{
		ID:                "maintainability",
		ISOCharacteristic: "Maintainability (Testability & Modifiability)",
		Status:            worstCheckStatus(checks, true),
		SummaryMetric:     fmt.Sprintf("%d/%d CI Gates | Go Cov %s", passGates, totalGates, covSummary),
		TargetExpression:  fmt.Sprintf("100%% automated gates pass & Go coverage >= %.0f%%", th.CoverageTargetPct),
		EvidenceSource:    "CLI / CI Gate Report (go test -cover, npm test, tsc, validators)",
		Checks:            checks,
	}
}

func evalSafety(facts NFRRawDatabaseFacts) dto.NFRDimensionResultDTO {
	thaiTotal := facts.ThaiViolationInExams + facts.ThaiViolationInLogs
	thaiSt := nfrStatusPass
	if thaiTotal > 0 {
		thaiSt = nfrStatusFail
	}
	secretSt := nfrStatusPass
	if facts.SecretLeakSnapshots > 0 {
		secretSt = nfrStatusFail
	}
	boilerSt := nfrStatusPass
	if facts.BannedBoilerplateScenarios > 0 {
		boilerSt = nfrStatusFail
	}
	evSt := nfrStatusPass
	if facts.EvidenceRatioViolationScenarios > 0 {
		evSt = nfrStatusFail
	}

	checks := []dto.NFRCheckItemDTO{
		{
			Key:      "english_only_invariant",
			Measured: fmt.Sprintf("%d", thaiTotal),
			Unit:     "violations",
			Target:   "0 Thai chars in Exams & initial_logs",
			Status:   thaiSt,
		},
		{
			Key:      "zero_secret_leak_snapshots",
			Measured: fmt.Sprintf("%d", facts.SecretLeakSnapshots),
			Unit:     "leaks",
			Target:   "0 secret/key leaks in scenario_snapshots",
			Status:   secretSt,
		},
		{
			Key:      "zero_banned_boilerplate",
			Measured: fmt.Sprintf("%d", facts.BannedBoilerplateScenarios),
			Unit:     "scenarios",
			Target:   "0 banned legacy boilerplate strings",
			Status:   boilerSt,
		},
		{
			Key:      "evidence_ratio_bound",
			Measured: fmt.Sprintf("%d", facts.EvidenceRatioViolationScenarios),
			Unit:     "violations",
			Target:   "0 scenarios exceeding 50% evidence log ratio",
			Status:   evSt,
		},
	}

	totalViolations := thaiTotal + facts.SecretLeakSnapshots + facts.BannedBoilerplateScenarios + facts.EvidenceRatioViolationScenarios
	return dto.NFRDimensionResultDTO{
		ID:                "safety",
		ISOCharacteristic: "Safety / Freedom from Risk (Simulated Training Boundary)",
		Status:            worstCheckStatus(checks, false),
		SummaryMetric:     fmt.Sprintf("%d safety boundary violations", totalViolations),
		TargetExpression:  "0 English-only, secret-leak, boilerplate, or evidence-ratio violations",
		EvidenceSource:    "Live PostgreSQL questions, scenarios & scenario_snapshots Scan",
		Checks:            checks,
	}
}
