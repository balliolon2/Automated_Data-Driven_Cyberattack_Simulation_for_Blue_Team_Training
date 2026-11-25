package dto

import "time"

// HealthResponse represents the database-backed service health check response.
type HealthResponse struct {
	Status        string  `json:"status" example:"healthy"`
	Service       string  `json:"service" example:"soc-trainer-api"`
	DBStatus      string  `json:"db_status" example:"up"`
	DBLatencyMs   float64 `json:"db_latency_ms" example:"2.41"`
	UptimeSeconds int64   `json:"uptime_seconds" example:"1842"`
}

// NFRThresholdsDTO documents the active numeric acceptance cutoffs for nfr-rubric-v1.
type NFRThresholdsDTO struct {
	RubricVersion              string  `json:"rubric_version" example:"nfr-rubric-v1"`
	NonLLMP95TargetMs          float64 `json:"non_llm_p95_target_ms" example:"300"`
	NonLLMP95WarnMs            float64 `json:"non_llm_p95_warn_ms" example:"500"`
	NonLLMP99TargetMs          float64 `json:"non_llm_p99_target_ms" example:"800"`
	NonLLMP99WarnMs            float64 `json:"non_llm_p99_warn_ms" example:"1200"`
	LLMP95TargetMs             float64 `json:"llm_p95_target_ms" example:"20000"`
	LLMP95WarnMs               float64 `json:"llm_p95_warn_ms" example:"45000"`
	DBPingTargetMs             float64 `json:"db_ping_target_ms" example:"50"`
	DBPingWarnMs               float64 `json:"db_ping_warn_ms" example:"150"`
	AvailabilityTargetPct      float64 `json:"availability_target_pct" example:"99"`
	AvailabilityWarnPct        float64 `json:"availability_warn_pct" example:"95"`
	ScalabilityP95TargetMs     float64 `json:"scalability_p95_target_ms" example:"500"`
	ScalabilityP95WarnMs       float64 `json:"scalability_p95_warn_ms" example:"1000"`
	ScalabilityErrorTargetPct  float64 `json:"scalability_error_target_pct" example:"1"`
	ScalabilityErrorWarnPct    float64 `json:"scalability_error_warn_pct" example:"5"`
	CoverageTargetPct          float64 `json:"coverage_target_pct" example:"60"`
	CoverageWarnPct            float64 `json:"coverage_warn_pct" example:"40"`
	CSUQTargetMean             float64 `json:"csuq_target_mean" example:"5"`
	CSUQWarnMean               float64 `json:"csuq_warn_mean" example:"4"`
	TLXTargetMax               float64 `json:"tlx_target_max" example:"55"`
	TLXWarnMax                 float64 `json:"tlx_warn_max" example:"70"`
}

// NFRPercentilesDTO captures R-7 latency percentiles in milliseconds (nil when Count == 0).
type NFRPercentilesDTO struct {
	Count int      `json:"count" example:"142"`
	Min   *float64 `json:"min" example:"1.2"`
	Mean  *float64 `json:"mean" example:"18.4"`
	P50   *float64 `json:"p50" example:"14.1"`
	P95   *float64 `json:"p95" example:"62.3"`
	P99   *float64 `json:"p99" example:"118.7"`
	Max   *float64 `json:"max" example:"144.0"`
}

// NFREndpointLatencyDTO summarizes latency and error counts for a normalized route template.
type NFREndpointLatencyDTO struct {
	Route      string   `json:"route" example:"GET /api/simulation/status"`
	Category   string   `json:"category" example:"non_llm_api"`
	Count      int      `json:"count" example:"45"`
	ErrorCount int      `json:"error_count" example:"0"`
	MeanMs     *float64 `json:"mean_ms" example:"16.2"`
	P50Ms      *float64 `json:"p50_ms" example:"12.5"`
	P95Ms      *float64 `json:"p95_ms" example:"48.9"`
	P99Ms      *float64 `json:"p99_ms" example:"75.0"`
	MaxMs      *float64 `json:"max_ms" example:"82.1"`
}

// NFRLLMTelemetryDTO summarizes runtime LLM scenario generation telemetry (never exposes secrets).
type NFRLLMTelemetryDTO struct {
	TotalCalls    int               `json:"total_calls" example:"4"`
	SuccessCount  int               `json:"success_count" example:"3"`
	FallbackCount int               `json:"fallback_count" example:"1"`
	ErrorCount    int               `json:"error_count" example:"0"`
	TotalAttempts int               `json:"total_attempts" example:"5"`
	TotalRetries  int               `json:"total_retries" example:"1"`
	Providers     []string          `json:"providers"`
	Models        []string          `json:"models"`
	OutcomeCounts map[string]int    `json:"outcome_counts"`
	Latency       NFRPercentilesDTO `json:"latency"`
}

// NFRRuntimeSnapshotDTO captures passive HTTP/LLM telemetry and database connectivity.
type NFRRuntimeSnapshotDTO struct {
	ProcessStartedAt       time.Time               `json:"process_started_at"`
	WindowStartedAt        time.Time               `json:"window_started_at"`
	UptimeSeconds          int64                   `json:"uptime_seconds" example:"3600"`
	GinMode                string                  `json:"gin_mode" example:"release"`
	TotalRequests          int                     `json:"total_requests" example:"240"`
	NonLLMRequests         int                     `json:"non_llm_requests" example:"236"`
	LLMRequests            int                     `json:"llm_requests" example:"4"`
	Status2xx              int                     `json:"status_2xx" example:"228"`
	Status3xx              int                     `json:"status_3xx" example:"0"`
	Status4xx              int                     `json:"status_4xx" example:"12"`
	Status5xx              int                     `json:"status_5xx" example:"0"`
	AvailabilityPct        *float64                `json:"availability_pct" example:"100"`
	ErrorRate5xxPct        *float64                `json:"error_rate_5xx_pct" example:"0"`
	InFlightCurrent        int                     `json:"in_flight_current" example:"1"`
	PeakConcurrentRequests int                     `json:"peak_concurrent_requests" example:"12"`
	DBStatus               string                  `json:"db_status" example:"up"`
	DBLatencyMs            float64                 `json:"db_latency_ms" example:"2.35"`
	NonLLMLatency          NFRPercentilesDTO       `json:"non_llm_latency"`
	LLMLatency             NFRPercentilesDTO       `json:"llm_latency"`
	LLMCalls               NFRLLMTelemetryDTO      `json:"llm_calls"`
	Endpoints              []NFREndpointLatencyDTO `json:"endpoints"`
}

// NFRCheckItemDTO represents an individual numeric or boolean verification check inside a dimension.
type NFRCheckItemDTO struct {
	Key      string `json:"key" example:"non_llm_p95_ms"`
	Measured string `json:"measured" example:"48.20"`
	Unit     string `json:"unit" example:"ms"`
	Target   string `json:"target" example:"<= 300 ms"`
	Status   string `json:"status" example:"pass"` // pass | warn | fail | not_measured | insufficient_data
}

// NFRDimensionResultDTO captures one evaluated ISO/IEC 25010 quality characteristic.
type NFRDimensionResultDTO struct {
	ID                string            `json:"id" example:"performance"`
	ISOCharacteristic string            `json:"iso_characteristic" example:"Performance Efficiency (Time Behaviour)"`
	Status            string            `json:"status" example:"pass"` // pass | warn | fail | not_measured | insufficient_data
	SummaryMetric     string            `json:"summary_metric" example:"p95 48.2 ms | DB 2.4 ms"`
	TargetExpression  string            `json:"target_expression" example:"Non-LLM p95 <= 300 ms; DB <= 50 ms"`
	EvidenceSource    string            `json:"evidence_source" example:"In-Memory HTTP Ring Buffer & Live DB Ping"`
	Checks            []NFRCheckItemDTO `json:"checks"`
}

// NFRSummaryDTO rolls up all 10 dimensions transparently without conflating pending items with failures.
type NFRSummaryDTO struct {
	OverallStatus    string  `json:"overall_status" example:"partial_pending"` // all_passed | pass_with_warnings | partial_pending | fail
	TotalDimensions  int     `json:"total_dimensions" example:"10"`
	PassedCount      int     `json:"passed_count" example:"7"`
	WarnCount        int     `json:"warn_count" example:"1"`
	FailCount        int     `json:"fail_count" example:"0"`
	NotMeasuredCount int     `json:"not_measured_count" example:"2"`
	MeasuredPassRate float64 `json:"measured_pass_rate" example:"87.5"`
}

// NFRReliabilityBreakdownDTO tracks RQ4 simulation session delivery and fallback reliability.
type NFRReliabilityBreakdownDTO struct {
	TotalSessions       int            `json:"total_sessions" example:"12"`
	AIGeneratedCount    int            `json:"ai_generated_count" example:"8"`
	StaticFallbackCount int            `json:"static_fallback_count" example:"4"`
	SystemFailureCount  int            `json:"system_failure_count" example:"0"`
	ValidSnapshotCount  int            `json:"valid_snapshot_count" example:"12"`
	TotalSnapshotCount  int            `json:"total_snapshot_count" example:"12"`
	FallbackReasons     map[string]int `json:"fallback_reasons"`
}

// NFRScalabilityLevelDTO captures one concurrency wave result from an in-app probe or external benchmark.
type NFRScalabilityLevelDTO struct {
	Concurrency   int     `json:"concurrency" example:"10"`
	TotalRequests int     `json:"total_requests" example:"100"`
	SuccessCount  int     `json:"success_count" example:"100"`
	ErrorCount    int     `json:"error_count" example:"0"`
	ErrorRatePct  float64 `json:"error_rate_pct" example:"0"`
	ThroughputRPS float64 `json:"throughput_rps" example:"312.5"`
	MeanMs        float64 `json:"mean_ms" example:"28.4"`
	P50Ms         float64 `json:"p50_ms" example:"24.0"`
	P95Ms         float64 `json:"p95_ms" example:"55.2"`
	P99Ms         float64 `json:"p99_ms" example:"78.1"`
	MaxMs         float64 `json:"max_ms" example:"84.0"`
	DurationMs    float64 `json:"duration_ms" example:"320.0"`
}

// NFRScalabilityBenchmarkDTO holds a complete scalability probe or CLI benchmark curve.
type NFRScalabilityBenchmarkDTO struct {
	Mode            string                   `json:"mode" example:"in_app_probe"` // in_app_probe | cli_benchmark
	Preset          string                   `json:"preset,omitempty" example:"standard"`
	ExecutedAt      time.Time                `json:"executed_at"`
	TargetEndpoints []string                 `json:"target_endpoints"`
	Levels          []NFRScalabilityLevelDTO `json:"levels"`
}

// NFRMaintainabilityReportDTO captures automated CI/test/coverage gate outcomes and optional CVE scans.
type NFRMaintainabilityReportDTO struct {
	CollectedAt             time.Time `json:"collected_at"`
	GitCommit               string    `json:"git_commit,omitempty" example:"aef5b35"`
	GoTestPassed            *bool     `json:"go_test_passed,omitempty" example:"true"`
	GoCoveragePct           *float64  `json:"go_coverage_pct,omitempty" example:"68.4"`
	FrontendTestPassed      *bool     `json:"frontend_test_passed,omitempty" example:"true"`
	FrontendTscPassed       *bool     `json:"frontend_tsc_passed,omitempty" example:"true"`
	EvidenceGatePassed      *bool     `json:"evidence_gate_passed,omitempty" example:"true"`
	TheoryGatePassed        *bool     `json:"theory_gate_passed,omitempty" example:"true"`
	IntegrationHarnessPassed *bool    `json:"integration_harness_passed,omitempty" example:"true"`
	GovulncheckHighCritical *int      `json:"govulncheck_high_critical,omitempty" example:"0"`
	NpmAuditHighCritical    *int      `json:"npm_audit_high_critical,omitempty" example:"0"`
	GitleaksFindings        *int      `json:"gitleaks_findings,omitempty" example:"0"`
}

// NFRReportResponse is the canonical evaluation report returned by live, run-detail, and JSON export endpoints.
type NFRReportResponse struct {
	RunID                 *string                      `json:"run_id,omitempty"`
	RunLabel              string                       `json:"run_label" example:"Live System Evaluation"`
	Environment           string                       `json:"environment" example:"local"`
	Source                string                       `json:"source" example:"live"`
	RubricVersion         string                       `json:"rubric_version" example:"nfr-rubric-v1"`
	EvaluatedAt           time.Time                    `json:"evaluated_at"`
	Notes                 string                       `json:"notes,omitempty"`
	Summary               NFRSummaryDTO                `json:"summary"`
	Runtime               NFRRuntimeSnapshotDTO        `json:"runtime"`
	Dimensions            []NFRDimensionResultDTO      `json:"dimensions"`
	Scalability           *NFRScalabilityBenchmarkDTO  `json:"scalability,omitempty"`
	ReliabilityBreakdown  NFRReliabilityBreakdownDTO   `json:"reliability_breakdown"`
	MaintainabilityReport *NFRMaintainabilityReportDTO `json:"maintainability_report,omitempty"`
	Thresholds            NFRThresholdsDTO             `json:"thresholds"`
}

// TelemetryResetResponse confirms an in-memory telemetry window reset without modifying persisted DB runs.
type TelemetryResetResponse struct {
	Reset           bool      `json:"reset" example:"true"`
	WindowStartedAt time.Time `json:"window_started_at"`
	Message         string    `json:"message" example:"In-memory HTTP and LLM telemetry window reset."`
}

// CreateNFRRunRequest captures a live snapshot or ingests CLI/JSON benchmark attachments.
type CreateNFRRunRequest struct {
	RunLabel              string                       `json:"run_label" example:"Local Docker Benchmark"`
	Environment           string                       `json:"environment" example:"local-docker"`
	Source                string                       `json:"source" example:"live_snapshot"` // live_snapshot | in_app_probe | cli_benchmark | imported_json
	Notes                 string                       `json:"notes" example:"Baseline run before tuning"`
	Scalability           *NFRScalabilityBenchmarkDTO  `json:"scalability,omitempty"`
	MaintainabilityReport *NFRMaintainabilityReportDTO `json:"maintainability_report,omitempty"`
}

// NFRRunListItemDTO summarizes one persisted nfr_evaluation_runs record.
type NFRRunListItemDTO struct {
	RunID            string    `json:"run_id"`
	RunLabel         string    `json:"run_label"`
	Environment      string    `json:"environment"`
	Source           string    `json:"source"`
	RubricVersion    string    `json:"rubric_version"`
	OverallStatus    string    `json:"overall_status"`
	PassedCount      int       `json:"passed_count"`
	WarnCount        int       `json:"warn_count"`
	FailCount        int       `json:"fail_count"`
	NotMeasuredCount int       `json:"not_measured_count"`
	MeasuredPassRate float64   `json:"measured_pass_rate"`
	NonLLMP95Ms      *float64  `json:"non_llm_p95_ms,omitempty"`
	AvailabilityPct  *float64  `json:"availability_pct,omitempty"`
	Notes            string    `json:"notes,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// NFRRunListResponse lists saved NFR evaluation snapshots ordered newest first.
type NFRRunListResponse struct {
	Runs []NFRRunListItemDTO `json:"runs"`
}

// NFRRunDetailResponse wraps a single saved NFR evaluation run and its full immutable report.
type NFRRunDetailResponse struct {
	Run    NFRRunListItemDTO `json:"run"`
	Report NFRReportResponse `json:"report"`
}

// RunScalabilityProbeRequest triggers the bounded in-app read-only concurrency probe.
type RunScalabilityProbeRequest struct {
	Preset        string `json:"preset" example:"standard"` // light | standard | burst
	Concurrency   int    `json:"concurrency,omitempty" example:"10"`
	TotalRequests int    `json:"total_requests,omitempty" example:"100"`
	SaveRun       bool   `json:"save_run" example:"false"`
	RunLabel      string `json:"run_label,omitempty" example:"In-App Safe Concurrency Probe"`
	Environment   string `json:"environment,omitempty" example:"local"`
}

// ScalabilityProbeResponse returns the scalability benchmark result and optionally the saved run metadata.
type ScalabilityProbeResponse struct {
	Benchmark NFRScalabilityBenchmarkDTO `json:"benchmark"`
	SavedRun  *NFRRunListItemDTO         `json:"saved_run,omitempty"`
}

// DeleteNFRRunResponse confirms deletion of a persisted NFR run.
type DeleteNFRRunResponse struct {
	Deleted bool   `json:"deleted" example:"true"`
	RunID   string `json:"run_id"`
}
