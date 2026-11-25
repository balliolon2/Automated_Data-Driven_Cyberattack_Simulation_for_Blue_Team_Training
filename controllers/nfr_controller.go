package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"cybersim/dto"
	"cybersim/models"
	"cybersim/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const maxNFRPayloadBytes = 1 << 20 // 1 MB upload/request cap

// NFRStore abstracts PostgreSQL health pings, live NFR audits, and persistent run snapshots.
type NFRStore interface {
	PingDB(ctx context.Context) (float64, error)
	GatherLiveNFRFacts(ctx context.Context) (services.NFRRawDatabaseFacts, error)
	CreateNFRRun(run *models.NFREvaluationRun) error
	ListNFRRuns(limit int) ([]models.NFREvaluationRun, error)
	GetNFRRunByID(runID string) (*models.NFREvaluationRun, error)
	DeleteNFRRun(runID string) (bool, error)
}

type gormNFRStore struct {
	DB *gorm.DB
}

func (s *gormNFRStore) PingDB(ctx context.Context) (float64, error) {
	if s.DB == nil {
		return 0, errors.New("database not initialized")
	}
	sqlDB, err := s.DB.DB()
	if err != nil {
		return 0, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	start := time.Now()
	err = sqlDB.PingContext(pingCtx)
	elapsedMs := float64(time.Since(start).Microseconds()) / 1000.0
	if err == nil && elapsedMs < 0.01 {
		elapsedMs = 0.01
	}
	return elapsedMs, err
}

func (s *gormNFRStore) GatherLiveNFRFacts(ctx context.Context) (services.NFRRawDatabaseFacts, error) {
	facts := services.NFRRawDatabaseFacts{
		DBStatus:        "up",
		FallbackReasons: make(map[string]int),
	}
	latencyMs, pingErr := s.PingDB(ctx)
	facts.DBLatencyMs = latencyMs
	if pingErr != nil {
		facts.DBStatus = "down"
		return facts, pingErr
	}

	var nonBcrypt int64
	if err := s.DB.Raw(`
		SELECT COUNT(*) FROM users
		WHERE password_hash NOT LIKE '$2a$%'
		  AND password_hash NOT LIKE '$2b$%'
		  AND password_hash NOT LIKE '$2y$%'
	`).Scan(&nonBcrypt).Error; err != nil {
		return facts, err
	}
	facts.NonBcryptUserCount = int(nonBcrypt)

	var overlapCount int64
	if err := s.DB.Raw(`
		SELECT COUNT(*) FROM (
			SELECT es_pre.user_id, eq_pre.question_id
			FROM exam_sessions es_pre
			JOIN exam_session_questions eq_pre ON eq_pre.session_id = es_pre.session_id
			JOIN exam_sessions es_post ON es_post.user_id = es_pre.user_id AND es_post.exam_type = 'post'
			JOIN exam_session_questions eq_post ON eq_post.session_id = es_post.session_id AND eq_post.question_id = eq_pre.question_id
			WHERE es_pre.exam_type = 'pre'
			GROUP BY es_pre.user_id, eq_pre.question_id
		) AS q_overlaps
	`).Scan(&overlapCount).Error; err != nil {
		return facts, err
	}
	facts.PrePostOverlapCount = int(overlapCount)

	var orphanSim int64
	if err := s.DB.Raw(`
		SELECT COUNT(*) FROM simulation_sessions ss
		LEFT JOIN users u ON u.user_id = ss.user_id
		LEFT JOIN scenarios sc ON sc.scenario_id = ss.scenario_id
		WHERE u.user_id IS NULL OR sc.scenario_id IS NULL
	`).Scan(&orphanSim).Error; err != nil {
		return facts, err
	}
	facts.OrphanSimSessionCount = int(orphanSim)

	var nullExams int64
	if err := s.DB.Raw(`SELECT COUNT(*) FROM exam_sessions WHERE status = 'completed' AND score IS NULL`).Scan(&nullExams).Error; err != nil {
		return facts, err
	}
	facts.NullScoreCompletedExams = int(nullExams)

	var activeQ, subtopics, thTheory, activeScen int64
	_ = s.DB.Raw(`SELECT COUNT(*) FROM questions WHERE is_active = true`).Scan(&activeQ).Error
	_ = s.DB.Raw(`SELECT COUNT(*) FROM theory_subtopics`).Scan(&subtopics).Error
	_ = s.DB.Raw(`SELECT COUNT(*) FROM theory_translations WHERE locale = 'th'`).Scan(&thTheory).Error
	_ = s.DB.Raw(`SELECT COUNT(*) FROM scenarios WHERE status = 'active'`).Scan(&activeScen).Error
	facts.ActiveQuestionsCount = int(activeQ)
	facts.TheorySubtopicsCount = int(subtopics)
	facts.ThaiTheoryCount = int(thTheory)
	facts.ActiveScenariosCount = int(activeScen)

	var simRow struct {
		TotalSessions   int `gorm:"column:total_sessions"`
		AICount         int `gorm:"column:ai_count"`
		FallbackCount   int `gorm:"column:fallback_count"`
		SysFailCount    int `gorm:"column:sys_fail_count"`
		FinishedCount   int `gorm:"column:finished_count"`
		CompletedCount  int `gorm:"column:completed_count"`
		MedianElapsed   *float64 `gorm:"column:median_elapsed"`
	}
	if err := s.DB.Raw(`
		SELECT
			COUNT(*) AS total_sessions,
			COUNT(*) FILTER (WHERE generation_type = 'ai_generated') AS ai_count,
			COUNT(*) FILTER (WHERE generation_type = 'static_fallback') AS fallback_count,
			COUNT(*) FILTER (WHERE failure_class = 'system_failure') AS sys_fail_count,
			COUNT(*) FILTER (WHERE status IN ('completed', 'abandoned')) AS finished_count,
			COUNT(*) FILTER (WHERE status = 'completed') AS completed_count,
			percentile_cont(0.5) WITHIN GROUP (ORDER BY elapsed_seconds)
				FILTER (WHERE status = 'completed' AND elapsed_seconds IS NOT NULL AND elapsed_seconds > 0) AS median_elapsed
		FROM simulation_sessions
	`).Scan(&simRow).Error; err != nil {
		return facts, err
	}
	facts.TotalSimSessions = simRow.TotalSessions
	facts.AIGeneratedSessions = simRow.AICount
	facts.StaticFallbackSessions = simRow.FallbackCount
	facts.SystemFailureSessions = simRow.SysFailCount
	facts.SimFinishedCount = simRow.FinishedCount
	facts.SimCompletedCount = simRow.CompletedCount
	facts.SimMedianSeconds = simRow.MedianElapsed

	var withSnap int64
	_ = s.DB.Raw(`
		SELECT COUNT(DISTINCT ss.session_id)
		FROM simulation_sessions ss
		JOIN scenario_snapshots sn ON sn.session_id = ss.session_id
	`).Scan(&withSnap).Error
	facts.SessionsWithSnapshot = int(withSnap)

	var snapStats struct {
		TotalSnaps        int `gorm:"column:total_snaps"`
		ValidSnaps        int `gorm:"column:valid_snaps"`
		CompleteMetaSnaps int `gorm:"column:complete_meta_snaps"`
	}
	_ = s.DB.Raw(`
		SELECT
			COUNT(*) AS total_snaps,
			COUNT(*) FILTER (WHERE validation_status = 'valid') AS valid_snaps,
			COUNT(*) FILTER (
				WHERE COALESCE(source_type, '') <> ''
				  AND COALESCE(prompt_version, '') <> ''
				  AND COALESCE(blueprint_version, '') <> ''
				  AND COALESCE(rubric_version, '') <> ''
				  AND COALESCE(validation_status, '') <> ''
			) AS complete_meta_snaps
		FROM scenario_snapshots
	`).Scan(&snapStats).Error
	facts.TotalSnapshotCount = snapStats.TotalSnaps
	facts.ValidSnapshotCount = snapStats.ValidSnaps
	facts.SnapshotsCompleteMeta = snapStats.CompleteMetaSnaps

	var reasonRows []struct {
		Reason string `gorm:"column:reason"`
		Cnt    int    `gorm:"column:cnt"`
	}
	if err := s.DB.Raw(`
		SELECT COALESCE(NULLIF(TRIM(fallback_reason), ''), 'unspecified') AS reason, COUNT(*) AS cnt
		FROM simulation_sessions
		WHERE generation_type = 'static_fallback'
		GROUP BY 1
		ORDER BY cnt DESC
		LIMIT 10
	`).Scan(&reasonRows).Error; err == nil {
		for _, r := range reasonRows {
			facts.FallbackReasons[r.Reason] = r.Cnt
		}
	}

	var examStats struct {
		TotalExams     int      `gorm:"column:total_exams"`
		CompletedExams int      `gorm:"column:completed_exams"`
		MedianSeconds  *float64 `gorm:"column:median_seconds"`
	}
	_ = s.DB.Raw(`
		SELECT
			COUNT(*) AS total_exams,
			COUNT(*) FILTER (WHERE status = 'completed') AS completed_exams,
			percentile_cont(0.5) WITHIN GROUP (ORDER BY time_spent_seconds)
				FILTER (WHERE status = 'completed' AND time_spent_seconds IS NOT NULL AND time_spent_seconds > 0) AS median_seconds
		FROM exam_sessions
	`).Scan(&examStats).Error
	facts.ExamTotalCount = examStats.TotalExams
	facts.ExamCompletedCount = examStats.CompletedExams
	facts.ExamMedianSeconds = examStats.MedianSeconds

	// Usability bridge from Issue #13 evaluation_metrics (reuse EvaluationMetricToSummaryDTO)
	var evalRows []models.EvaluationMetric
	if err := s.DB.Model(&models.EvaluationMetric{}).
		Joins("JOIN users ON users.user_id = evaluation_metrics.user_id").
		Where("users.role = ?", "learner").
		Find(&evalRows).Error; err == nil && len(evalRows) > 0 {
		facts.QuestionnaireCount = len(evalRows)
		csuqSum := 0.0
		csuqCnt := 0
		tlxSum := 0.0
		tlxCnt := 0
		for _, m := range evalRows {
			summary := services.EvaluationMetricToSummaryDTO(m)
			if summary.CSUQScores.Subscales.Overall != nil {
				csuqSum += *summary.CSUQScores.Subscales.Overall
				csuqCnt++
			}
			if summary.NASATLXScores.SchemaVersion != "" {
				tlxSum += summary.NASATLXScores.OverallWorkload
				tlxCnt++
			}
		}
		if csuqCnt > 0 {
			avg := services.RoundTelemetry2(csuqSum / float64(csuqCnt))
			facts.CSUQOverallMean = &avg
		}
		if tlxCnt > 0 {
			avg := services.RoundTelemetry2(tlxSum / float64(tlxCnt))
			facts.NASATLXOverallMean = &avg
		}
	}

	// Safety scans: English-only invariant on questions & initial_logs, secret patterns in snapshots, boilerplate & evidence bounds
	var qTexts []string
	if err := s.DB.Model(&models.Question{}).Where("is_active = true").Pluck("question_text", &qTexts).Error; err == nil {
		for _, txt := range qTexts {
			if services.ContainsThaiRune(txt) {
				facts.ThaiViolationInExams++
			}
		}
	}

	var activeScenarios []models.Scenario
	if err := s.DB.Where("status = 'active'").Find(&activeScenarios).Error; err == nil {
		for _, sc := range activeScenarios {
			logsBytes, _ := json.Marshal(sc.InitialLogs)
			if services.ContainsThaiRune(string(logsBytes)) {
				facts.ThaiViolationInLogs++
			}

			var logsSlice []map[string]interface{}
			_ = json.Unmarshal(logsBytes, &logsSlice)

			pbBytes, _ := json.Marshal(sc.PlaybookSteps)
			var pbMap map[string]interface{}
			_ = json.Unmarshal(pbBytes, &pbMap)

			if services.ScenarioTextHasBannedBoilerplate(logsSlice, pbMap) {
				facts.BannedBoilerplateScenarios++
			}

			if len(logsSlice) > 0 {
				eoBytes, _ := json.Marshal(sc.ExpectedOutcomes)
				var eoMap map[string]interface{}
				_ = json.Unmarshal(eoBytes, &eoMap)
				if rawFindings, ok := eoMap["key_findings"].([]interface{}); ok {
					uniqueEvidence := make(map[int]bool)
					for _, rf := range rawFindings {
						if fMap, ok := rf.(map[string]interface{}); ok {
							if idxs, ok := fMap["evidence_log_indices"].([]interface{}); ok {
								for _, rawIdx := range idxs {
									if num, ok := rawIdx.(float64); ok {
										uniqueEvidence[int(num)] = true
									}
								}
							}
						}
					}
					if len(uniqueEvidence)*2 > len(logsSlice) {
						facts.EvidenceRatioViolationScenarios++
					}
				}
			}
		}
	}

	var snapPayloads []string
	if err := s.DB.Raw(`SELECT scenario_data::text FROM scenario_snapshots ORDER BY rendered_at DESC LIMIT 500`).Scan(&snapPayloads).Error; err == nil {
		llmKey := os.Getenv("LLM_API_KEY")
		jwtSecret := os.Getenv("JWT_SECRET")
		for _, raw := range snapPayloads {
			if services.SnapshotContainsSecretPattern(raw, llmKey, jwtSecret) {
				facts.SecretLeakSnapshots++
			}
		}
	}

	return facts, nil
}

func (s *gormNFRStore) CreateNFRRun(run *models.NFREvaluationRun) error {
	return s.DB.Create(run).Error
}

func (s *gormNFRStore) ListNFRRuns(limit int) ([]models.NFREvaluationRun, error) {
	var runs []models.NFREvaluationRun
	err := s.DB.Order("created_at DESC").Limit(limit).Find(&runs).Error
	return runs, err
}

func (s *gormNFRStore) GetNFRRunByID(runID string) (*models.NFREvaluationRun, error) {
	var run models.NFREvaluationRun
	if err := s.DB.Where("run_id = ?", runID).First(&run).Error; err != nil {
		return nil, err
	}
	return &run, nil
}

func (s *gormNFRStore) DeleteNFRRun(runID string) (bool, error) {
	res := s.DB.Where("run_id = ?", runID).Delete(&models.NFREvaluationRun{})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}


// NFRController serves public DB-backed health checks and Admin-only NFR evaluation endpoints.
type NFRController struct {
	store        NFRStore
	telemetry    *services.NFRTelemetryStore
	router       http.Handler
	probeRunning atomic.Bool
}

// NewNFRController constructs an NFRController wired to GORM and the global telemetry store.
func NewNFRController(db *gorm.DB, router http.Handler) *NFRController {
	return &NFRController{
		store:     &gormNFRStore{DB: db},
		telemetry: services.GlobalNFRTelemetry(),
		router:    router,
	}
}

// NewNFRControllerWithStore constructs an NFRController with injectable dependencies for tests.
func NewNFRControllerWithStore(store NFRStore, telemetry *services.NFRTelemetryStore, router http.Handler) *NFRController {
	if telemetry == nil {
		telemetry = services.GlobalNFRTelemetry()
	}
	return &NFRController{
		store:     store,
		telemetry: telemetry,
		router:    router,
	}
}

// HealthCheck godoc
// @Summary Service and database connectivity health check
// @Description Executes a 2-second PostgreSQL ping and reports database status, latency in milliseconds, and service uptime.
// @Tags System
// @Produce json
// @Success 200 {object} dto.HealthResponse
// @Failure 503 {object} dto.HealthResponse
// @Router /health [get]
func (nc *NFRController) HealthCheck(c *gin.Context) {
	latencyMs, err := nc.store.PingDB(c.Request.Context())
	resp := dto.HealthResponse{
		Status:        "healthy",
		Service:       "soc-trainer-api",
		DBStatus:      "up",
		DBLatencyMs:   services.RoundTelemetry2(latencyMs),
		UptimeSeconds: nc.telemetry.UptimeSeconds(),
	}
	if err != nil {
		resp.Status = "degraded"
		resp.DBStatus = "down"
		c.JSON(http.StatusServiceUnavailable, resp)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (nc *NFRController) buildLiveReport(
	c *gin.Context,
	label, envName, source, notes string,
	overrideScale *dto.NFRScalabilityBenchmarkDTO,
	maint *dto.NFRMaintainabilityReportDTO,
) (dto.NFRReportResponse, error) {
	facts, err := nc.store.GatherLiveNFRFacts(c.Request.Context())
	if err != nil && facts.DBStatus != "down" {
		return dto.NFRReportResponse{}, err
	}

	ginMode := gin.Mode()
	runtimeSnap := nc.telemetry.Snapshot(ginMode, facts.DBStatus, facts.DBLatencyMs)
	secProbes := services.RunInProcessSecurityProbe(nc.router, c.GetHeader("Authorization"), nc.telemetry)
	posture := services.LoadNFRPostureInput(ginMode)
	thresholds := services.LoadNFRThresholdsFromEnv()

	scale := overrideScale
	if scale == nil {
		scale = nc.telemetry.GetScalabilityBenchmark()
	}

	report := services.EvaluateNFR(services.NFREvaluateInput{
		RunLabel:        label,
		Environment:     envName,
		Source:          source,
		Notes:           notes,
		EvaluatedAt:     time.Now().UTC(),
		Thresholds:      thresholds,
		Runtime:         runtimeSnap,
		Facts:           facts,
		SecurityProbes:  secProbes,
		Posture:         posture,
		Scalability:     scale,
		Maintainability: maint,
	})
	return report, nil
}

// GetLiveNFR godoc
// @Summary Get live 10-dimension ISO/IEC 25010 NFR evaluation (Admin only)
// @Description Aggregates passive HTTP/LLM ring-buffer telemetry, live PostgreSQL integrity/reproducibility/safety/usability audits, read-only security probes, and the latest scalability probe result.
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.NFRReportResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/nfr/live [get]
func (nc *NFRController) GetLiveNFR(c *gin.Context) {
	envName := strings.TrimSpace(os.Getenv("APP_ENVIRONMENT"))
	if envName == "" {
		envName = "local"
	}
	report, err := nc.buildLiveReport(c, "Live System Evaluation", envName, "live", "", nil, nil)
	if err != nil {
		respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to evaluate live NFR metrics.")
		return
	}
	c.JSON(http.StatusOK, report)
}

// ResetTelemetryWindow godoc
// @Summary Reset in-memory HTTP and LLM telemetry window (Admin only)
// @Description Clears the passive in-memory HTTP and LLM telemetry ring buffers and resets window_started_at without modifying persisted NFR evaluation runs.
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.TelemetryResetResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /admin/nfr/telemetry/reset [post]
func (nc *NFRController) ResetTelemetryWindow(c *gin.Context) {
	winStart := nc.telemetry.ResetWindow()
	c.JSON(http.StatusOK, dto.TelemetryResetResponse{
		Reset:           true,
		WindowStartedAt: winStart,
		Message:         "In-memory HTTP and LLM telemetry window reset.",
	})
}

func runModelToListItem(r models.NFREvaluationRun) dto.NFRRunListItemDTO {
	return dto.NFRRunListItemDTO{
		RunID:            r.RunID,
		RunLabel:         r.RunLabel,
		Environment:      r.Environment,
		Source:           r.Source,
		RubricVersion:    r.RubricVersion,
		OverallStatus:    r.OverallStatus,
		PassedCount:      r.PassedCount,
		WarnCount:        r.WarnCount,
		FailCount:        r.FailCount,
		NotMeasuredCount: r.NotMeasuredCount,
		MeasuredPassRate: r.MeasuredPassRate,
		NonLLMP95Ms:      r.NonLLMP95Ms,
		AvailabilityPct:  r.AvailabilityPct,
		Notes:            r.Notes,
		CreatedAt:        r.CreatedAt,
	}
}

func decodeRunPayload(raw any) (dto.NFRReportResponse, error) {
	if typed, ok := raw.(dto.NFRReportResponse); ok {
		return typed, nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return dto.NFRReportResponse{}, err
	}
	var out dto.NFRReportResponse
	if err := json.Unmarshal(b, &out); err != nil {
		return dto.NFRReportResponse{}, err
	}
	return out, nil
}

func (nc *NFRController) persistNFRReport(c *gin.Context, report dto.NFRReportResponse, source string) (dto.NFRRunDetailResponse, error) {
	var triggeredBy *string
	if uid, _, ok := extractCallerIdentity(c); ok && examUUIDPattern.MatchString(uid) {
		triggeredBy = &uid
	}

	run := models.NFREvaluationRun{
		TriggeredBy:      triggeredBy,
		RunLabel:         report.RunLabel,
		Environment:      report.Environment,
		Source:           source,
		RubricVersion:    report.RubricVersion,
		OverallStatus:    report.Summary.OverallStatus,
		PassedCount:      report.Summary.PassedCount,
		WarnCount:        report.Summary.WarnCount,
		FailCount:        report.Summary.FailCount,
		NotMeasuredCount: report.Summary.NotMeasuredCount,
		MeasuredPassRate: report.Summary.MeasuredPassRate,
		NonLLMP95Ms:      report.Runtime.NonLLMLatency.P95,
		AvailabilityPct:  report.Runtime.AvailabilityPct,
		ReportPayload:    report,
		Notes:            report.Notes,
		CreatedAt:        time.Now().UTC(),
	}
	if err := nc.store.CreateNFRRun(&run); err != nil {
		return dto.NFRRunDetailResponse{}, err
	}
	runID := run.RunID
	report.RunID = &runID
	return dto.NFRRunDetailResponse{
		Run:    runModelToListItem(run),
		Report: report,
	}, nil
}

// CreateNFRRun godoc
// @Summary Capture a live NFR snapshot or ingest a CLI/JSON benchmark run (Admin only)
// @Description Evaluates current live DB/runtime checks merged with optional scalability and CI maintainability attachments (max 1 MB body) and saves an immutable snapshot in nfr_evaluation_runs.
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateNFRRunRequest true "Run snapshot or benchmark ingestion payload"
// @Success 201 {object} dto.NFRRunDetailResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 413 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/nfr/runs [post]
func (nc *NFRController) CreateNFRRun(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxNFRPayloadBytes)
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			respondAPIError(c, http.StatusRequestEntityTooLarge, "NFR_PAYLOAD_TOO_LARGE", "Benchmark payload exceeds the 1 MB maximum size.")
			return
		}
		respondAPIError(c, http.StatusBadRequest, "NFR_INVALID_BENCHMARK_PAYLOAD", "Failed to read request payload.")
		return
	}

	var req dto.CreateNFRRunRequest
	if len(strings.TrimSpace(string(bodyBytes))) > 0 {
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			respondAPIError(c, http.StatusBadRequest, "NFR_INVALID_BENCHMARK_PAYLOAD", "Invalid NFR run JSON payload.")
			return
		}
	}
	if err := services.ValidateCreateNFRRunRequest(&req); err != nil {
		respondAPIError(c, http.StatusBadRequest, "NFR_INVALID_BENCHMARK_PAYLOAD", err.Error())
		return
	}

	source := strings.TrimSpace(req.Source)
	if source == "" {
		source = "live_snapshot"
	}
	label := strings.TrimSpace(req.RunLabel)
	if label == "" {
		label = "Live NFR Snapshot"
	}
	envName := strings.TrimSpace(req.Environment)
	if envName == "" {
		envName = "local"
	}

	report, err := nc.buildLiveReport(c, label, envName, source, req.Notes, req.Scalability, req.MaintainabilityReport)
	if err != nil {
		respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to build NFR evaluation snapshot.")
		return
	}
	detail, err := nc.persistNFRReport(c, report, source)
	if err != nil {
		respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to save NFR evaluation snapshot.")
		return
	}
	c.JSON(http.StatusCreated, detail)
}

// ListNFRRuns godoc
// @Summary List persisted NFR evaluation runs (Admin only)
// @Description Lists immutable NFR snapshots ordered newest first.
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Maximum runs to return (1-200)" default(50)
// @Success 200 {object} dto.NFRRunListResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/nfr/runs [get]
func (nc *NFRController) ListNFRRuns(c *gin.Context) {
	limit := 50
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > 200 {
		limit = 200
	}

	rows, err := nc.store.ListNFRRuns(limit)
	if err != nil {
		respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to list NFR evaluation runs.")
		return
	}
	out := make([]dto.NFRRunListItemDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, runModelToListItem(r))
	}
	c.JSON(http.StatusOK, dto.NFRRunListResponse{Runs: out})
}

// GetNFRRun godoc
// @Summary Get a saved NFR evaluation run by ID (Admin only)
// @Description Returns the immutable NFR report snapshot stored for a specific run UUID.
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param runId path string true "Run UUID"
// @Success 200 {object} dto.NFRRunDetailResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/nfr/runs/{runId} [get]
func (nc *NFRController) GetNFRRun(c *gin.Context) {
	runID := strings.TrimSpace(c.Param("runId"))
	if !isValidExamUUID(runID) {
		respondAPIError(c, http.StatusNotFound, "NFR_RUN_NOT_FOUND", "NFR evaluation run not found.")
		return
	}
	row, err := nc.store.GetNFRRunByID(runID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondAPIError(c, http.StatusNotFound, "NFR_RUN_NOT_FOUND", "NFR evaluation run not found.")
			return
		}
		respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to retrieve NFR evaluation run.")
		return
	}
	report, err := decodeRunPayload(row.ReportPayload)
	if err != nil {
		respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to decode saved NFR report.")
		return
	}
	idCopy := row.RunID
	report.RunID = &idCopy
	c.JSON(http.StatusOK, dto.NFRRunDetailResponse{
		Run:    runModelToListItem(*row),
		Report: report,
	})
}

// DeleteNFRRun godoc
// @Summary Delete a persisted NFR evaluation run (Admin only)
// @Description Permanently deletes a saved NFR evaluation snapshot by UUID.
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param runId path string true "Run UUID"
// @Success 200 {object} dto.DeleteNFRRunResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/nfr/runs/{runId} [delete]
func (nc *NFRController) DeleteNFRRun(c *gin.Context) {
	runID := strings.TrimSpace(c.Param("runId"))
	if !isValidExamUUID(runID) {
		respondAPIError(c, http.StatusNotFound, "NFR_RUN_NOT_FOUND", "NFR evaluation run not found.")
		return
	}
	deleted, err := nc.store.DeleteNFRRun(runID)
	if err != nil {
		respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to delete NFR evaluation run.")
		return
	}
	if !deleted {
		respondAPIError(c, http.StatusNotFound, "NFR_RUN_NOT_FOUND", "NFR evaluation run not found.")
		return
	}
	c.JSON(http.StatusOK, dto.DeleteNFRRunResponse{Deleted: true, RunID: runID})
}

// RunScalabilityProbe godoc
// @Summary Execute bounded safe in-app concurrency probe (Admin only)
// @Description Runs a mutex-guarded, telemetry-isolated read-only concurrency probe (light 5/50, standard 10/100, burst 20/200) and optionally persists a snapshot.
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.RunScalabilityProbeRequest false "Scalability probe options"
// @Success 200 {object} dto.ScalabilityProbeResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/nfr/probe/scalability [post]
func (nc *NFRController) RunScalabilityProbe(c *gin.Context) {
	var req dto.RunScalabilityProbeRequest
	if c.Request.Body != nil {
		bodyBytes, _ := io.ReadAll(io.LimitReader(c.Request.Body, 16384))
		if len(strings.TrimSpace(string(bodyBytes))) > 0 {
			if err := json.Unmarshal(bodyBytes, &req); err != nil {
				respondAPIError(c, http.StatusBadRequest, "NFR_INVALID_PROBE_PARAMS", "Invalid scalability probe request.")
				return
			}
		}
	}

	if _, _, _, err := services.ResolveScalabilityProbeParams(req.Preset, req.Concurrency, req.TotalRequests); err != nil {
		respondAPIError(c, http.StatusBadRequest, "NFR_INVALID_PROBE_PARAMS", err.Error())
		return
	}

	if !nc.probeRunning.CompareAndSwap(false, true) {
		respondAPIError(c, http.StatusConflict, "NFR_PROBE_ALREADY_RUNNING", "A concurrency probe is already executing. Please wait until it completes.")
		return
	}
	defer nc.probeRunning.Store(false)

	bench, err := services.RunSafeScalabilityProbe(nc.router, c.GetHeader("Authorization"), nc.telemetry, req.Preset, req.Concurrency, req.TotalRequests)
	if err != nil {
		respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to execute scalability probe.")
		return
	}

	resp := dto.ScalabilityProbeResponse{Benchmark: bench}
	if req.SaveRun {
		label := strings.TrimSpace(req.RunLabel)
		if label == "" {
			label = "In-App Scalability Probe (" + bench.Preset + ")"
		}
		envName := strings.TrimSpace(req.Environment)
		if envName == "" {
			envName = "local"
		}
		report, buildErr := nc.buildLiveReport(c, label, envName, "in_app_probe", "", &bench, nil)
		if buildErr != nil {
			respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Probe succeeded but failed to build run report.")
			return
		}
		detail, saveErr := nc.persistNFRReport(c, report, "in_app_probe")
		if saveErr != nil {
			respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Probe succeeded but failed to save snapshot.")
			return
		}
		resp.SavedRun = &detail.Run
	}

	c.JSON(http.StatusOK, resp)
}

// ExportNFRReport godoc
// @Summary Export live or saved NFR evaluation report as UTF-8 BOM CSV or JSON (Admin only)
// @Description Streams RFC 4180 CSV bytes (with UTF-8 BOM and formula-injection neutralization) or canonical JSON for the live evaluation or a specified run_id.
// @Tags Admin
// @Produce text/csv
// @Produce application/json
// @Security BearerAuth
// @Param format query string false "Export format (csv or json)" default(json)
// @Param run_id query string false "Optional saved run UUID; omits for live view"
// @Success 200 {object} dto.NFRReportResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/nfr/export [get]
func (nc *NFRController) ExportNFRReport(c *gin.Context) {
	format := strings.ToLower(strings.TrimSpace(c.DefaultQuery("format", "json")))
	if format != "csv" && format != "json" {
		respondAPIError(c, http.StatusBadRequest, "EVALUATION_INVALID_EXPORT_FORMAT", "Unsupported export format. Expected 'csv' or 'json'.")
		return
	}

	runID := strings.TrimSpace(c.Query("run_id"))
	var report dto.NFRReportResponse
	tag := "live"

	if runID != "" {
		if !isValidExamUUID(runID) {
			respondAPIError(c, http.StatusNotFound, "NFR_RUN_NOT_FOUND", "NFR evaluation run not found.")
			return
		}
		row, err := nc.store.GetNFRRunByID(runID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				respondAPIError(c, http.StatusNotFound, "NFR_RUN_NOT_FOUND", "NFR evaluation run not found.")
				return
			}
			respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to retrieve NFR evaluation run.")
			return
		}
		decoded, err := decodeRunPayload(row.ReportPayload)
		if err != nil {
			respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to decode saved NFR report.")
			return
		}
		idCopy := row.RunID
		decoded.RunID = &idCopy
		report = decoded
		if len(runID) >= 8 {
			tag = runID[:8]
		} else {
			tag = runID
		}
	} else {
		liveRep, err := nc.buildLiveReport(c, "Live System Evaluation", "local", "live", "", nil, nil)
		if err != nil {
			respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to evaluate live NFR metrics.")
			return
		}
		report = liveRep
	}

	dateTag := time.Now().UTC().Format("20060102")
	filename := "nfr_evaluation_" + tag + "_" + dateTag

	if format == "json" {
		c.Header("Content-Disposition", "attachment; filename=\""+filename+".json\"")
		c.JSON(http.StatusOK, report)
		return
	}

	csvBytes, err := services.BuildNFRExportCSV(report)
	if err != nil {
		respondAPIError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to generate NFR CSV export.")
		return
	}
	c.Header("Content-Disposition", "attachment; filename=\""+filename+".csv\"")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", csvBytes)
}
