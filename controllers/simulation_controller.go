package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"cybersim/dto"
	"cybersim/models"
	"cybersim/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Configuration constants
const (
	ProficiencyThreshold = 70.0 // Uniform threshold across all domains
	MaxScenarios         = 10   // Maximum scenarios per training session (configurable cap)
	MinScenarios         = 1    // Minimum scenarios before allowing completion
	WeightOld            = 0.4  // Weight for existing proficiency score
	WeightNew            = 0.6  // Weight for new scenario score
	ScoreWeightTPFP      = 0.20 // 20% of scenario score
	ScoreWeightFindings  = 0.40 // 40% of scenario score
	ScoreWeightResponse  = 0.40 // 40% of scenario score
)

type SimulationController struct {
	DB *gorm.DB
}

func NewSimulationController(db *gorm.DB) *SimulationController {
	if db != nil {
		_ = db.AutoMigrate(&models.SimulationSession{}, &models.ScenarioSnapshot{})
		_ = services.EnsureScenarioPlaybookQuality(db, "scenarios_seed.json")
	}
	return &SimulationController{DB: db}
}

var (
	errSelectionSessionNotFound    = errors.New("simulation session not found")
	errSelectionSessionForbidden   = errors.New("simulation session is not owned by caller")
	errSelectionSessionInactive    = errors.New("simulation session is not in progress")
	errSelectionLogsUnavailable    = errors.New("scenario telemetry logs are unavailable")
	errSelectionLogIndexOutOfRange = errors.New("log_index is out of range")
)

// sessionActionSpec describes one append-only session action.
type sessionActionSpec struct {
	ActionType string
	Payload    any
	IsCorrect  *bool
	Points     int
}

// selectedLogIndices replays the session's select_evidence actions into the
// current evidence set. A read failure degrades to an empty selection so that
// session recovery still works.
func (sc *SimulationController) selectedLogIndices(sessionID string) []int {
	var actions []models.SessionAction
	if err := sc.DB.
		Where("session_id = ? AND action_type = ?", sessionID, services.SelectEvidenceActionType).
		Order("step_order asc").
		Find(&actions).Error; err != nil {
		log.Printf("Failed to replay evidence selection for session %s: %v", sessionID, err)
		return []int{}
	}
	return services.ReplaySelectedLogIndices(actions)
}

// lockSessionRow loads a simulation session under a FOR UPDATE row lock and
// re-checks ownership and in-progress status on the locked row. Every session
// action writer takes this lock before allocating MAX(step_order)+1.
func lockSessionRow(tx *gorm.DB, sessionID, userID string) (models.SimulationSession, error) {
	var session models.SimulationSession
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("session_id = ?", sessionID).
		First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return session, errSelectionSessionNotFound
		}
		return session, err
	}
	if session.UserID != userID {
		return session, errSelectionSessionForbidden
	}
	if session.Status != "in_progress" {
		return session, errSelectionSessionInactive
	}
	return session, nil
}

// nextStepOrder allocates the next step order for a session. Callers must hold
// the session row lock so the MAX+1 read cannot race another append.
func nextStepOrder(tx *gorm.DB, sessionID string) (int, error) {
	var maxStep int
	if err := tx.Model(&models.SessionAction{}).
		Where("session_id = ?", sessionID).
		Select("COALESCE(MAX(step_order), 0)").
		Scan(&maxStep).Error; err != nil {
		return 0, err
	}
	return maxStep + 1, nil
}

// appendSessionActionTx inserts one action using the caller's transaction. The
// caller must already hold the session row lock.
func appendSessionActionTx(tx *gorm.DB, sessionID string, spec sessionActionSpec) (models.SessionAction, error) {
	step, err := nextStepOrder(tx, sessionID)
	if err != nil {
		return models.SessionAction{}, err
	}

	action := models.SessionAction{
		SessionID:  sessionID,
		StepOrder:  step,
		ActionType: spec.ActionType,
		Payload:    spec.Payload,
		IsCorrect:  spec.IsCorrect,
		Points:     spec.Points,
		Timestamp:  time.Now(),
	}
	if err := tx.Create(&action).Error; err != nil {
		return models.SessionAction{}, err
	}
	return action, nil
}

// appendSessionAction appends one session action in its own transaction. The
// simulation session row is locked FOR UPDATE, ownership/status are re-checked,
// step order is allocated as MAX+1, and a unique violation is retried once.
func (sc *SimulationController) appendSessionAction(sessionID, userID string, spec sessionActionSpec) (models.SessionAction, error) {
	const maxAttempts = 2

	for attempt := 0; attempt < maxAttempts; attempt++ {
		var created models.SessionAction
		err := sc.DB.Transaction(func(tx *gorm.DB) error {
			if _, err := lockSessionRow(tx, sessionID, userID); err != nil {
				return err
			}
			action, err := appendSessionActionTx(tx, sessionID, spec)
			if err != nil {
				return err
			}
			created = action
			return nil
		})

		if err == nil {
			return created, nil
		}
		if attempt == 0 && isUniqueViolation(err) {
			continue
		}
		return models.SessionAction{}, err
	}

	return models.SessionAction{}, fmt.Errorf("failed to allocate step order after retry")
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "23505") ||
		strings.Contains(message, "duplicate key") ||
		strings.Contains(message, "unique constraint")
}

// scenarioLogCount reports the number of initial logs and whether the payload
// is an array. JSONB decodes to []interface{}, but tests and cached scenarios
// may hand over other slice shapes.
func scenarioLogCount(logs interface{}) (int, bool) {
	if logs == nil {
		return 0, false
	}
	switch value := logs.(type) {
	case []interface{}:
		return len(value), true
	case []map[string]interface{}:
		return len(value), true
	}

	reflected := reflect.ValueOf(logs)
	if reflected.Kind() != reflect.Slice {
		return 0, false
	}
	return reflected.Len(), true
}

// validateSelectionLogIndex validates an evidence index against the log array
// that defines the session's index space.
func validateSelectionLogIndex(logs any, index int) error {
	if index < 0 {
		return errSelectionLogIndexOutOfRange
	}
	count, ok := scenarioLogCount(logs)
	if !ok {
		return errSelectionLogsUnavailable
	}
	if index >= count {
		return errSelectionLogIndexOutOfRange
	}
	return nil
}

// snapshotInitialLogs extracts the initial telemetry logs captured in a
// ScenarioSnapshot. JSONB decodes to a map; a freshly-built snapshot may carry
// a typed Scenario.
func snapshotInitialLogs(data any) (any, bool) {
	switch value := data.(type) {
	case nil:
		return nil, false
	case map[string]interface{}:
		logs, ok := value["initial_logs"]
		if !ok || logs == nil {
			return nil, false
		}
		return logs, true
	case models.Scenario:
		if value.InitialLogs == nil {
			return nil, false
		}
		return value.InitialLogs, true
	case *models.Scenario:
		if value == nil || value.InitialLogs == nil {
			return nil, false
		}
		return value.InitialLogs, true
	case []byte:
		var decoded struct {
			InitialLogs any `json:"initial_logs"`
		}
		if err := json.Unmarshal(value, &decoded); err != nil || decoded.InitialLogs == nil {
			return nil, false
		}
		return decoded.InitialLogs, true
	case json.RawMessage:
		var decoded struct {
			InitialLogs any `json:"initial_logs"`
		}
		if err := json.Unmarshal(value, &decoded); err != nil || decoded.InitialLogs == nil {
			return nil, false
		}
		return decoded.InitialLogs, true
	default:
		return nil, false
	}
}

// sessionScenario returns the scenario captured in the session's immutable
// snapshot. The bool reports whether the snapshot was used. Legacy sessions
// without a snapshot fall back to the live scenario row so old attempts remain
// recoverable.
func (sc *SimulationController) sessionScenario(session models.SimulationSession) (models.Scenario, bool, error) {
	var snapshot models.ScenarioSnapshot
	err := sc.DB.Where("session_id = ?", session.SessionID).First(&snapshot).Error
	if err == nil {
		var scenario models.Scenario
		encoded, marshalErr := json.Marshal(snapshot.ScenarioData)
		if marshalErr != nil {
			return scenario, true, marshalErr
		}
		if unmarshalErr := json.Unmarshal(encoded, &scenario); unmarshalErr != nil {
			return scenario, true, unmarshalErr
		}
		if scenario.ScenarioID == "" {
			scenario.ScenarioID = snapshot.ScenarioID
		}
		return scenario, true, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Scenario{}, false, err
	}

	var scenario models.Scenario
	if err := sc.DB.Where("scenario_id = ?", session.ScenarioID).First(&scenario).Error; err != nil {
		return scenario, false, err
	}
	return scenario, false, nil
}

// sessionScenarioLogs returns the telemetry log array that defines the
// evidence index space for a session. It prefers the ScenarioSnapshot captured
// at session start so indices cannot drift if the scenario row is edited.
func (sc *SimulationController) sessionScenarioLogs(session models.SimulationSession) (any, error) {
	var snapshot models.ScenarioSnapshot
	err := sc.DB.Where("session_id = ?", session.SessionID).First(&snapshot).Error
	if err == nil {
		logs, ok := snapshotInitialLogs(snapshot.ScenarioData)
		if !ok {
			return nil, errSelectionLogsUnavailable
		}
		return logs, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var scenario models.Scenario
	if err := sc.DB.Where("scenario_id = ?", session.ScenarioID).First(&scenario).Error; err != nil {
		return nil, err
	}
	return scenario.InitialLogs, nil
}

func (sc *SimulationController) ensureScenarioRemediated(scenario *models.Scenario) {
	if scenario == nil {
		return
	}
	if services.UpgradeScenarioPlaybookSteps(scenario) && sc.DB != nil && scenario.ScenarioID != "" {
		_ = sc.DB.Model(&models.Scenario{}).Where("scenario_id = ?", scenario.ScenarioID).Update("playbook_steps", scenario.PlaybookSteps).Error
	}
}

// --- DTO Aliases ---
type DomainStatus = dto.DomainStatusDTO
type SimulationStatusResponse = dto.SimulationStatusResponse
type ScenarioResponse = dto.ScenarioResponse
type LogQueryInput = dto.LogQueryRequest
type SubmitScenarioInput = dto.SubmitScenarioRequest

// --- Domain name helper ---

var domainNames = map[string]string{
	"domain1": "General Security Concepts",
	"domain2": "Threats, Vulnerabilities, and Mitigations",
	"domain3": "Security Architecture",
	"domain4": "Security Operations",
	"domain5": "Security Program Management and Oversight",
}

// GetStatus godoc
// @Summary Get learner simulation status
// @Description Returns per-domain proficiency status, completed scenario count, and training eligibility.
// @Tags Simulation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SimulationStatusResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /simulation/status [get]
func (sc *SimulationController) GetStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get all user skill profiles
	var profiles []models.UserSkillProfile
	sc.DB.Where("user_id = ?", userID).Find(&profiles)

	if len(profiles) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"error":          "No skill profile found. Please complete the pre-test first.",
			"needs_training": true,
		})
		return
	}

	// Build domain status
	domains := make([]DomainStatus, 0, len(profiles))
	allPassed := (len(profiles) >= 5)
	for _, p := range profiles {
		passed := p.ProficiencyScore >= ProficiencyThreshold
		if !passed {
			allPassed = false
		}
		domains = append(domains, DomainStatus{
			DomainID:         p.DomainID,
			DomainName:       domainNames[p.DomainID],
			ProficiencyScore: math.Round(p.ProficiencyScore*100) / 100,
			Threshold:        ProficiencyThreshold,
			Passed:           passed,
		})
	}

	// Count completed scenarios
	var completedCount int64
	sc.DB.Model(&models.SimulationSession{}).
		Where("user_id = ? AND status = 'completed'", userID).
		Count(&completedCount)

	needsTraining := !allPassed || int(completedCount) < MinScenarios
	if int(completedCount) >= MaxScenarios {
		needsTraining = false
	}

	c.JSON(http.StatusOK, SimulationStatusResponse{
		NeedsTraining:      needsTraining,
		Domains:            domains,
		CompletedScenarios: int(completedCount),
		MaxScenarios:       MaxScenarios,
		AllDomainsPassed:   allPassed,
	})
}

// StartScenario godoc
// @Summary Start next adaptive simulation scenario
// @Description Synthesizes or retrieves the next scenario focused on the learner's weakest domain. If an active simulation exists, resumes it.
// @Tags Simulation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.ActiveSimulationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /simulation/start [post]
func (sc *SimulationController) StartScenario(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Check for existing in-progress session
	var activeSession models.SimulationSession
	err := sc.DB.Where("user_id = ? AND status = 'in_progress'", userID).First(&activeSession).Error
	if err == nil {
		// Return existing active session with scenario data
		scenario, fromSnapshot, scenarioErr := sc.sessionScenario(activeSession)
		if scenarioErr != nil {
			log.Printf("Failed to load scenario for session %s: %v", activeSession.SessionID, scenarioErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load scenario"})
			return
		}
		if !fromSnapshot {
			sc.ensureScenarioRemediated(&scenario)
		}
		c.JSON(http.StatusOK, gin.H{
			"session":              activeSession,
			"scenario":             sanitizeScenario(scenario, activeSession.SessionID),
			"selected_log_indices": sc.selectedLogIndices(activeSession.SessionID),
		})
		return
	}

	// Check if training is still needed
	var profiles []models.UserSkillProfile
	sc.DB.Where("user_id = ?", userID).Find(&profiles)

	if len(profiles) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"error":          "No skill profile found. Please complete the pre-test first.",
			"needs_training": true,
		})
		return
	}

	var completedCount int64
	sc.DB.Model(&models.SimulationSession{}).
		Where("user_id = ? AND status = 'completed'", userID).
		Count(&completedCount)

	// Check max cap
	if int(completedCount) >= MaxScenarios {
		c.JSON(http.StatusOK, gin.H{
			"message":     "Maximum scenario limit reached",
			"completed":   true,
			"cap_reached": true,
			"status":      "max_scenarios_reached",
		})
		return
	}

	// Check if all domains passed (and minimum met)
	allPassed := (len(profiles) >= 5)
	for _, p := range profiles {
		if p.ProficiencyScore < ProficiencyThreshold {
			allPassed = false
			break
		}
	}
	if allPassed && int(completedCount) >= MinScenarios {
		c.JSON(http.StatusOK, gin.H{
			"message":   "All domains meet the proficiency threshold",
			"completed": true,
			"status":    "eligible_for_post_test",
		})
		return
	}

	// Find the weakest domain
	weakestDomain := ""
	weakestScore := 101.0
	for _, p := range profiles {
		if p.ProficiencyScore < weakestScore {
			weakestScore = p.ProficiencyScore
			weakestDomain = p.DomainID
		}
	}

	var scenario models.Scenario
	generationType := "static_fallback"
	fallbackReason := ""
	selectionReason := fmt.Sprintf("lowest_proficiency: %s (score: %.1f%%, threshold: %.1f%%)", weakestDomain, weakestScore, ProficiencyThreshold)

	// Attempt AI Dynamic Scenario Generation if LLM_API_KEY is configured
	if os.Getenv("LLM_API_KEY") != "" {
		domainName := domainNames[weakestDomain]
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		aiScenario, llmErr := services.GenerateLLMScenario(ctx, weakestDomain, domainName, weakestScore, profiles)
		if llmErr == nil && aiScenario != nil {
			if err := sc.DB.Create(aiScenario).Error; err == nil {
				scenario = *aiScenario
				generationType = "ai_generated"
				fallbackReason = ""
				log.Printf("[LLM] Successfully generated dynamic AI scenario: %s (ID: %s)", scenario.Title, scenario.ScenarioID)
			} else {
				fallbackReason = fmt.Sprintf("Failed to save AI scenario to database: %v", err)
				log.Printf("[LLM] %s. Falling back to static scenario pool.", fallbackReason)
			}
		} else {
			if llmErr != nil {
				fallbackReason = llmErr.Error()
			} else {
				fallbackReason = "LLM service returned empty scenario"
			}
			log.Printf("[LLM] Dynamic scenario generation fallback: %s", fallbackReason)
		}
	} else {
		fallbackReason = "LLM_API_KEY environment variable is not configured"
	}

	// Fallback to static scenario pool if LLM was skipped or failed
	if scenario.ScenarioID == "" {
		// Get already-completed scenario IDs for this user
		completedSubquery := sc.DB.Table("simulation_sessions").
			Select("scenario_id").
			Where("user_id = ? AND status IN ('completed', 'in_progress')", userID)

		// Try to find a scenario for the weakest domain that hasn't been done
		err = sc.DB.Where("domain_id = ? AND status = 'active' AND scenario_id NOT IN (?)", weakestDomain, completedSubquery).
			Order("RANDOM()").
			First(&scenario).Error

		if err != nil {
			// No scenarios left for weakest domain — try other weak domains
			for _, p := range profiles {
				if p.DomainID == weakestDomain {
					continue
				}
				if p.ProficiencyScore >= ProficiencyThreshold {
					continue
				}
				err = sc.DB.Where("domain_id = ? AND status = 'active' AND scenario_id NOT IN (?)", p.DomainID, completedSubquery).
					Order("RANDOM()").
					First(&scenario).Error
				if err == nil {
					selectionReason = fmt.Sprintf("fallback_weak_domain: %s (score: %.1f%%)", p.DomainID, p.ProficiencyScore)
					break
				}
			}
		}

		if err != nil {
			// No scenarios left for any weak domain — try any remaining scenario
			err = sc.DB.Where("status = 'active' AND scenario_id NOT IN (?)", completedSubquery).
				Order("RANDOM()").
				First(&scenario).Error
			if err == nil {
				selectionReason = "fallback_any_active_scenario"
			}
		}
	}

	// Safety Net: Synthesize dynamic AI scenario if external LLM failed and static pool is exhausted
	if scenario.ScenarioID == "" {
		domainName := domainNames[weakestDomain]
		fallbackScenario := services.GenerateFallbackScenario(weakestDomain, domainName)
		if err := sc.DB.Create(fallbackScenario).Error; err == nil {
			scenario = *fallbackScenario
			generationType = "ai_generated"
			fallbackReason = "Synthesized adaptive AI scenario (LLM API offline / static pool exhausted)"
			selectionReason = fmt.Sprintf("synthesized_fallback: %s", weakestDomain)
			log.Printf("[LLM Synthesis Guarantee] Created dynamic AI scenario: %s (ID: %s)", scenario.Title, scenario.ScenarioID)
			err = nil
		}
	}

	if err != nil && scenario.ScenarioID == "" {
		// Scenario pool exhausted for static scenarios and fallback failed
		if allPassed || int(completedCount) >= MaxScenarios {
			c.JSON(http.StatusOK, gin.H{
				"message":        "All required scenarios completed",
				"completed":      true,
				"pool_exhausted": true,
				"status":         "eligible_for_post_test",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"error":          "Failed to load scenario. Please retry.",
				"completed":      false,
				"pool_exhausted": true,
				"status":         "scenario_unavailable",
				"failure_class":  "system_failure",
			})
		}
		return
	}

	sc.ensureScenarioRemediated(&scenario)

	// Create simulation session with full audit fields
	session := models.SimulationSession{
		UserID:             userID.(string),
		ScenarioID:         scenario.ScenarioID,
		Status:             "in_progress",
		GenerationType:     generationType,
		FallbackReason:     fallbackReason,
		SelectionReason:    selectionReason,
		ThresholdValue:     ProficiencyThreshold,
		MaxScenarios:       MaxScenarios,
		TimeLimitSeconds:   7200,
		CalculationVersion: services.CalculationVersion,
		RubricVersion:      services.RubricVersion,
		BlueprintVersion:   "blueprint-v1",
		StartedAt:          time.Now(),
	}

	if err := sc.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create simulation session"})
		return
	}

	// Create Immutable Scenario Snapshot
	provider := "internal"
	modelName := "static"
	if generationType == "ai_generated" {
		provider = "openai-compatible"
		modelName = os.Getenv("LLM_MODEL")
		if modelName == "" {
			modelName = "gpt-4o-mini"
		}
	}

	snapshot := models.ScenarioSnapshot{
		SessionID:        session.SessionID,
		ScenarioID:       scenario.ScenarioID,
		ScenarioData:     scenario,
		SourceType:       generationType,
		Provider:         provider,
		Model:            modelName,
		PromptVersion:    "prompt-v1",
		BlueprintVersion: "blueprint-v1",
		RubricVersion:    services.RubricVersion,
		ValidationStatus: "valid",
		FallbackReason:   fallbackReason,
		SelectionReason:  selectionReason,
		RenderedAt:       time.Now(),
	}
	if err := sc.DB.Create(&snapshot).Error; err != nil {
		log.Printf("Failed to persist scenario snapshot for session %s: %v", session.SessionID, err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"session":              session,
		"scenario":             sanitizeScenario(scenario, session.SessionID),
		"selected_log_indices": []int{},
	})
}

// GetActiveSession godoc
// @Summary Get active simulation session
// @Description Recovers the in-progress simulation session and scenario data.
// @Tags Simulation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.ActiveSimulationResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /simulation/session [get]
func (sc *SimulationController) GetActiveSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// First, check for an in-progress session
	var activeSession models.SimulationSession
	err := sc.DB.Where("user_id = ? AND status = 'in_progress'", userID).First(&activeSession).Error
	if err == nil {
		scenario, fromSnapshot, scenarioErr := sc.sessionScenario(activeSession)
		if scenarioErr != nil {
			log.Printf("Failed to load scenario for session %s: %v", activeSession.SessionID, scenarioErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load scenario"})
			return
		}
		if !fromSnapshot {
			sc.ensureScenarioRemediated(&scenario)
		}
		c.JSON(http.StatusOK, gin.H{
			"type":                 "in_progress",
			"session":              activeSession,
			"scenario":             sanitizeScenario(scenario, activeSession.SessionID),
			"selected_log_indices": sc.selectedLogIndices(activeSession.SessionID),
		})
		return
	}

	// Check for a recently completed session (not yet acknowledged)
	// A session is "not acknowledged" if it's the most recent completed session
	// and no newer in-progress session exists
	var lastCompleted models.SimulationSession
	err = sc.DB.Where("user_id = ? AND status = 'completed'", userID).
		Order("completed_at DESC").
		First(&lastCompleted).Error
	if err == nil {
		// Return result data for this session
		c.JSON(http.StatusOK, gin.H{
			"type":       "completed",
			"session":    lastCompleted,
			"session_id": lastCompleted.SessionID,
		})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "No active simulation session found"})
}

// LogQuery godoc
// @Summary Execute and record a telemetry log query
// @Description Records a query action executed by the learner against telemetry logs for auditing and investigation scoring.
// @Tags Simulation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.LogQueryRequest true "Log Query Request"
// @Success 200 {object} dto.LogQueryResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /simulation/log-query [post]
func (sc *SimulationController) LogQuery(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input LogQueryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Record the query as a session action. The helper locks the session row,
	// re-checks ownership/status, and allocates the step order atomically.
	action, err := sc.appendSessionAction(input.SessionID, userID.(string), sessionActionSpec{
		ActionType: "view_log",
		Payload: map[string]interface{}{
			"query": input.Query,
		},
	})
	if err != nil {
		writeSessionActionError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recorded":  true,
		"action_id": action.ActionID,
	})
}

// SelectLog godoc
// @Summary Toggle one telemetry log as selected evidence
// @Description Records a select_evidence toggle action for an in-progress simulation session and returns the replayed selection.
// @Tags Simulation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.SelectLogRequest true "Evidence selection toggle"
// @Success 200 {object} dto.EvidenceSelectionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /simulation/select-log [post]
func (sc *SimulationController) SelectLog(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input dto.SelectLogRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.LogIndex == nil || input.Selected == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "log_index and selected are required"})
		return
	}

	var session models.SimulationSession
	if err := sc.DB.Where("session_id = ?", input.SessionID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Simulation session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load simulation session"})
		return
	}
	if session.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Simulation session is not owned by caller"})
		return
	}
	if session.Status != "in_progress" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Simulation session is not in progress"})
		return
	}

	// Validate against the session's immutable snapshot, not the live scenario
	// row, so reseeding or editing a scenario cannot shift evidence indices.
	logs, err := sc.sessionScenarioLogs(session)
	if err != nil {
		switch {
		case errors.Is(err, errSelectionLogsUnavailable):
			c.JSON(http.StatusBadRequest, gin.H{"error": errSelectionLogsUnavailable.Error()})
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Scenario not found"})
		default:
			log.Printf("Failed to load scenario logs for session %s: %v", session.SessionID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load scenario logs"})
		}
		return
	}
	if err := validateSelectionLogIndex(logs, *input.LogIndex); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload := map[string]interface{}{
		"operation": "toggle",
		"log_index": *input.LogIndex,
		"selected":  *input.Selected,
	}
	if _, err := sc.appendSessionAction(session.SessionID, userID.(string), sessionActionSpec{
		ActionType: services.SelectEvidenceActionType,
		Payload:    payload,
	}); err != nil {
		writeSessionActionError(c, err)
		return
	}

	indices := sc.selectedLogIndices(session.SessionID)
	c.JSON(http.StatusOK, dto.EvidenceSelectionResponse{
		SelectedLogIndices: indices,
		SelectedCount:      len(indices),
	})
}

// ClearLogSelection godoc
// @Summary Clear all selected evidence logs
// @Description Records one select_evidence clear action for an in-progress simulation session.
// @Tags Simulation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ClearLogSelectionRequest true "Clear evidence selection"
// @Success 200 {object} dto.EvidenceSelectionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /simulation/select-log/clear [post]
func (sc *SimulationController) ClearLogSelection(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input dto.ClearLogSelectionRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var session models.SimulationSession
	if err := sc.DB.Where("session_id = ?", input.SessionID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Simulation session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load simulation session"})
		return
	}
	if session.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Simulation session is not owned by caller"})
		return
	}
	if session.Status != "in_progress" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Simulation session is not in progress"})
		return
	}

	if _, err := sc.appendSessionAction(session.SessionID, userID.(string), sessionActionSpec{
		ActionType: services.SelectEvidenceActionType,
		Payload:    map[string]interface{}{"operation": "clear"},
	}); err != nil {
		writeSessionActionError(c, err)
		return
	}

	// Replay instead of assuming an empty set so the response stays
	// server-authoritative if a toggle committed immediately after the clear.
	indices := sc.selectedLogIndices(session.SessionID)
	c.JSON(http.StatusOK, dto.EvidenceSelectionResponse{
		SelectedLogIndices: indices,
		SelectedCount:      len(indices),
	})
}

func writeSessionActionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errSelectionSessionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Simulation session not found"})
	case errors.Is(err, errSelectionSessionForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "Simulation session is not owned by caller"})
	case errors.Is(err, errSelectionSessionInactive):
		c.JSON(http.StatusForbidden, gin.H{"error": "Simulation session is not in progress"})
	default:
		log.Printf("Failed to record session action: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record session action"})
	}
}

// SubmitScenario godoc
// @Summary Submit scenario response actions and triage decision
// @Description Evaluates the learner's TP/FP classification, response actions, and discovered key findings, updating proficiency scores via weighted average.
// @Tags Simulation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.SubmitScenarioRequest true "Scenario Submission"
// @Success 200 {object} dto.SubmitScenarioResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /simulation/submit [post]
func (sc *SimulationController) SubmitScenario(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input SubmitScenarioInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := sc.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start submission"})
		return
	}

	// Acquire the session row lock before reading or scoring anything so an
	// evidence toggle cannot commit between scoring and lock acquisition. The
	// lock also re-checks ownership and status on the locked row.
	session, err := lockSessionRow(tx, input.SessionID, userID.(string))
	if err != nil {
		tx.Rollback()
		writeSessionActionError(c, err)
		return
	}

	// Get the scenario
	var scenario models.Scenario
	if err := tx.Where("scenario_id = ?", session.ScenarioID).First(&scenario).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Scenario not found"})
		return
	}

	// Score the submission using the deep evaluator module. The scoring formula
	// is intentionally unchanged; only the transaction/lock ordering is tighter.
	result := services.ScoreSubmission(scenario, models.SubmitScenarioInput{
		SessionID:          input.SessionID,
		IsTruePositive:     input.IsTruePositive,
		TPFPSelected:       input.TPFPSelected,
		SelectedActions:    input.SelectedActions,
		DiscoveredFindings: input.DiscoveredFindings,
		SelectedLogIndices: input.SelectedLogIndices,
	})

	// Record the TP/FP decision action
	if _, err := appendSessionActionTx(tx, input.SessionID, sessionActionSpec{
		ActionType: "triage_alert",
		Payload: map[string]interface{}{
			"user_choice":    input.IsTruePositive,
			"correct_answer": scenario.IsTruePositive,
		},
		IsCorrect: &result.TPFPCorrect,
		Points:    int(result.TPFPPoints),
	}); err != nil {
		tx.Rollback()
		writeSessionActionError(c, err)
		return
	}

	// Record the response action
	if _, err := appendSessionActionTx(tx, input.SessionID, sessionActionSpec{
		ActionType: "respond",
		Payload: map[string]interface{}{
			"selected_actions": input.SelectedActions,
		},
		Points: int(result.ResponsePoints),
	}); err != nil {
		tx.Rollback()
		writeSessionActionError(c, err)
		return
	}

	// Record findings submission
	findingsAction, err := appendSessionActionTx(tx, input.SessionID, sessionActionSpec{
		ActionType: "submit_decision",
		Payload: map[string]interface{}{
			"discovered_findings": input.DiscoveredFindings,
		},
		Points: int(result.FindingsPoints),
	})
	if err != nil {
		tx.Rollback()
		writeSessionActionError(c, err)
		return
	}

	// Update simulation session with complete evaluation result snapshot
	now := time.Now()
	session.Status = "completed"
	session.FailureClass = "learner_progress"
	session.CalculationVersion = services.CalculationVersion
	session.RubricVersion = services.RubricVersion
	session.FinalScore = &result.TotalScore
	session.CompletedAt = &now
	session.SkillGap = result
	session.TotalActions = findingsAction.StepOrder
	session.ElapsedSeconds = int(time.Since(session.StartedAt).Seconds())

	// Update user_skill_profiles with weighted average
	allDomainsAboveThreshold := true
	for domainID, newScore := range result.DomainScoreMap {
		var profile models.UserSkillProfile
		err := tx.Where("user_id = ? AND domain_id = ?", userID, domainID).First(&profile).Error
		if err != nil {
			// Profile doesn't exist for this domain — create it
			profile = models.UserSkillProfile{
				UserID:             userID.(string),
				DomainID:           domainID,
				ProficiencyScore:   newScore,
				ScenariosCompleted: 1,
			}
			now := time.Now()
			profile.LastPracticed = &now
			if err := tx.Create(&profile).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update skill profile"})
				return
			}
			if newScore < ProficiencyThreshold {
				allDomainsAboveThreshold = false
			}
		} else {
			// Weighted average update using Evaluator calculation
			profile.ProficiencyScore = services.CalculateUpdatedProficiency(profile.ProficiencyScore, newScore)
			profile.ScenariosCompleted++
			now := time.Now()
			profile.LastPracticed = &now
			if err := tx.Save(&profile).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update skill profile"})
				return
			}
			if profile.ProficiencyScore < ProficiencyThreshold {
				allDomainsAboveThreshold = false
			}
		}
	}

	// Check total profiles count across all 5 domains
	var totalProfilesCount int64
	if err := tx.Model(&models.UserSkillProfile{}).Where("user_id = ?", userID).Count(&totalProfilesCount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update skill profile"})
		return
	}
	if totalProfilesCount < 5 {
		allDomainsAboveThreshold = false
	}

	// Determine completion reason based on stopping rules
	var totalCompletedCount int64
	if err := tx.Model(&models.SimulationSession{}).Where("user_id = ? AND status = 'completed'", userID).Count(&totalCompletedCount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
		return
	}
	// Add current session (which is about to be saved as completed)
	totalCompletedCount++

	if allDomainsAboveThreshold {
		session.CompletionReason = "all_domains_passed"
	} else if int(totalCompletedCount) >= MaxScenarios {
		session.CompletionReason = "max_scenarios_reached"
	} else if session.ElapsedSeconds >= session.TimeLimitSeconds {
		session.CompletionReason = "time_limit_reached"
	} else {
		session.CompletionReason = "learner_completed"
	}

	if err := tx.Save(&session).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit scenario"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":            result,
		"completion_reason": session.CompletionReason,
		"eligible":          allDomainsAboveThreshold,
	})
}

// GetResult godoc
// @Summary Get detailed retrospective result of a simulation session
// @Description Returns the complete evaluation breakdown, action audit trail, and updated domain proficiencies for a completed simulation session.
// @Tags Simulation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param sessionId path string true "Simulation Session UUID"
// @Success 200 {object} dto.SimulationResultResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /simulation/result/{sessionId} [get]
func (sc *SimulationController) GetResult(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	sessionID := c.Param("sessionId")

	var session models.SimulationSession
	if err := sc.DB.Where("session_id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	if session.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session is not completed yet"})
		return
	}

	// Get the scenario (full details for review)
	var scenario models.Scenario
	sc.DB.Where("scenario_id = ?", session.ScenarioID).First(&scenario)

	// Get all actions for this session
	var actions []models.SessionAction
	sc.DB.Where("session_id = ?", sessionID).Order("step_order asc").Find(&actions)

	// Retrieve or reconstruct evaluation result
	var evalResult models.ScenarioResult
	if session.SkillGap != nil {
		gapBytes, err := json.Marshal(session.SkillGap)
		if err == nil {
			_ = json.Unmarshal(gapBytes, &evalResult)
		}
	}

	// Fallback for legacy sessions if evalResult was not populated
	if evalResult.TotalScore == 0 && len(evalResult.FindingsDetail) == 0 && len(evalResult.ResponseDetail) == 0 {
		evalResult = services.ReconstructFromActions(scenario, actions)
	}

	// Get current proficiency scores
	var profiles []models.UserSkillProfile
	sc.DB.Where("user_id = ?", userID).Find(&profiles)

	domainProficiencies := make([]DomainStatus, 0, len(profiles))
	for _, p := range profiles {
		domainProficiencies = append(domainProficiencies, DomainStatus{
			DomainID:         p.DomainID,
			DomainName:       domainNames[p.DomainID],
			ProficiencyScore: math.Round(p.ProficiencyScore*100) / 100,
			Threshold:        ProficiencyThreshold,
			Passed:           p.ProficiencyScore >= ProficiencyThreshold,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"session":              session,
		"scenario":             scenario,
		"actions":              actions,
		"result":               evalResult,
		"domain_proficiencies": domainProficiencies,
	})
}

// sanitizeScenario removes answer data and deterministically shuffles Response Actions using sessionID
func sanitizeScenario(s models.Scenario, sessionID string) ScenarioResponse {
	seedBase := sessionID
	if seedBase == "" {
		seedBase = s.ScenarioID
	}
	// Strip is_correct and explanation from playbook_steps and deterministically shuffle per phase
	sanitizedSteps := sanitizePlaybookSteps(s.PlaybookSteps, seedBase)

	// Extract only IDs and evidence log indices from expected_outcomes
	var sanitizedFindings []map[string]interface{}
	outcomesMap, ok := s.ExpectedOutcomes.(map[string]interface{})
	if ok {
		keyFindings, ok := outcomesMap["key_findings"].([]interface{})
		if ok {
			for _, kf := range keyFindings {
				finding, ok := kf.(map[string]interface{})
				if !ok {
					continue
				}

				fID, _ := finding["id"].(string)
				indices, _ := finding["evidence_log_indices"].([]interface{})

				sanitizedFindings = append(sanitizedFindings, map[string]interface{}{
					"id":                   fID,
					"evidence_log_indices": indices,
				})
			}
		}
	}

	return ScenarioResponse{
		ScenarioID:        s.ScenarioID,
		Title:             s.Title,
		Description:       s.Description,
		DomainID:          s.DomainID,
		Difficulty:        s.Difficulty,
		InitialLogs:       s.InitialLogs,
		PlaybookSteps:     sanitizedSteps,
		SanitizedFindings: sanitizedFindings,
	}
}

// sanitizePlaybookSteps strips evaluation keys and deterministically shuffles Response Actions per phase using sessionID
func sanitizePlaybookSteps(steps any, sessionID string) any {
	stepsMap, ok := steps.(map[string]interface{})
	if !ok {
		return steps
	}

	sanitized := make(map[string]interface{})
	phases := []string{"containment", "eradication", "recovery"}

	for _, phase := range phases {
		var rawActions []map[string]interface{}

		switch val := stepsMap[phase].(type) {
		case []interface{}:
			for _, item := range val {
				if m, ok := item.(map[string]interface{}); ok {
					rawActions = append(rawActions, m)
				}
			}
		case []map[string]interface{}:
			rawActions = val
		}

		sanitizedActions := make([]map[string]interface{}, 0, len(rawActions))
		for _, action := range rawActions {
			if id, hasID := action["id"]; hasID {
				label, _ := action["label"].(string)
				sanitizedActions = append(sanitizedActions, map[string]interface{}{
					"id":    id,
					"label": label,
				})
			}
		}

		if len(sanitizedActions) > 1 && sessionID != "" {
			h := fnv.New64a()
			_, _ = h.Write([]byte(fmt.Sprintf("%s:%s", sessionID, phase)))
			rng := rand.New(rand.NewSource(int64(h.Sum64())))
			rng.Shuffle(len(sanitizedActions), func(i, j int) {
				sanitizedActions[i], sanitizedActions[j] = sanitizedActions[j], sanitizedActions[i]
			})
		}

		sanitized[phase] = sanitizedActions
	}

	return sanitized
}

// GetResearchSummary godoc
// @Summary Get research evaluation summary
// @Description Returns structured research and experiment metrics across pre-test, scenario loop, and post-test pursuant to the evaluation protocol.
// @Tags Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.ResearchSummaryResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /analytics/research-summary [get]
func (sc *SimulationController) GetResearchSummary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 1. Get Pre-test details
	var preExam models.ExamSession
	hasPre := sc.DB.Where("user_id = ? AND exam_type = 'pre' AND status = 'completed'", userID).
		Order("completed_at DESC").
		First(&preExam).Error == nil

	// 2. Get Post-test details
	var postExam models.ExamSession
	hasPost := sc.DB.Where("user_id = ? AND exam_type = 'post' AND status = 'completed'", userID).
		Order("completed_at DESC").
		First(&postExam).Error == nil

	// 3. Get User Skill Profiles
	var profiles []models.UserSkillProfile
	sc.DB.Where("user_id = ?", userID).Find(&profiles)

	allPassed := len(profiles) >= 5
	avgProficiency := 0.0
	for _, p := range profiles {
		avgProficiency += p.ProficiencyScore
		if p.ProficiencyScore < ProficiencyThreshold {
			allPassed = false
		}
	}
	if len(profiles) > 0 {
		avgProficiency = math.Round((avgProficiency/float64(len(profiles)))*100) / 100
	}

	// 4. Get simulation sessions count & total actions
	var completedSessions []models.SimulationSession
	sc.DB.Where("user_id = ? AND status = 'completed'", userID).Find(&completedSessions)

	totalActions := 0
	totalSimElapsed := 0
	for _, s := range completedSessions {
		totalActions += s.TotalActions
		totalSimElapsed += s.ElapsedSeconds
	}

	var scoreImprovement *float64
	if hasPre && hasPost && preExam.Score != nil && postExam.Score != nil {
		diff := math.Round((*postExam.Score-*preExam.Score)*100) / 100
		scoreImprovement = &diff
	}

	c.JSON(http.StatusOK, gin.H{
		"protocol_version":    "protocol-v1",
		"calculation_version": services.CalculationVersion,
		"rubric_version":      services.RubricVersion,
		"participant_id":      userID,
		"has_pre_test":        hasPre,
		"pre_test_score":      preExam.Score,
		"has_post_test":       hasPost,
		"post_test_score":     postExam.Score,
		"improvement_score":   scoreImprovement,
		"all_domains_passed":  allPassed,
		"eligible_for_post":   allPassed && len(completedSessions) >= MinScenarios,
		"completed_scenarios": len(completedSessions),
		"max_scenarios":       MaxScenarios,
		"total_actions":       totalActions,
		"total_elapsed_sec":   totalSimElapsed,
		"average_proficiency": avgProficiency,
		"domain_profiles":     profiles,
	})
}
