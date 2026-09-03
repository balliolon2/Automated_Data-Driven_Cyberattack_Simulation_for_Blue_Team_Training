package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"cybersim/dto"
	"cybersim/models"
	"cybersim/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Configuration constants
const (
	ProficiencyThreshold = 70.0  // Uniform threshold across all domains
	MaxScenarios         = 10    // Maximum scenarios per training session (configurable cap)
	MinScenarios         = 1     // Minimum scenarios before allowing completion
	WeightOld            = 0.4   // Weight for existing proficiency score
	WeightNew            = 0.6   // Weight for new scenario score
	ScoreWeightTPFP      = 0.20  // 20% of scenario score
	ScoreWeightFindings  = 0.40  // 40% of scenario score
	ScoreWeightResponse  = 0.40  // 40% of scenario score
)

type SimulationController struct {
	DB *gorm.DB
}

func NewSimulationController(db *gorm.DB) *SimulationController {
	if db != nil {
		_ = db.AutoMigrate(&models.SimulationSession{}, &models.ScenarioSnapshot{})
	}
	return &SimulationController{DB: db}
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
		var scenario models.Scenario
		sc.DB.Where("scenario_id = ?", activeSession.ScenarioID).First(&scenario)
		c.JSON(http.StatusOK, gin.H{
			"session":  activeSession,
			"scenario": sanitizeScenario(scenario),
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
	_ = sc.DB.Create(&snapshot)

	c.JSON(http.StatusCreated, gin.H{
		"session":  session,
		"scenario": sanitizeScenario(scenario),
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
		var scenario models.Scenario
		sc.DB.Where("scenario_id = ?", activeSession.ScenarioID).First(&scenario)
		c.JSON(http.StatusOK, gin.H{
			"type":     "in_progress",
			"session":  activeSession,
			"scenario": sanitizeScenario(scenario),
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

	// Verify session ownership
	var session models.SimulationSession
	if err := sc.DB.Where("session_id = ? AND user_id = ? AND status = 'in_progress'", input.SessionID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid or inactive session"})
		return
	}

	// Get the next step order
	var maxStep int
	sc.DB.Model(&models.SessionAction{}).
		Where("session_id = ?", input.SessionID).
		Select("COALESCE(MAX(step_order), 0)").
		Scan(&maxStep)

	// Record the query as a session action
	action := models.SessionAction{
		SessionID:  input.SessionID,
		StepOrder:  maxStep + 1,
		ActionType: "view_log",
		Payload: map[string]interface{}{
			"query": input.Query,
		},
		Timestamp: time.Now(),
	}

	if err := sc.DB.Create(&action).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log query"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recorded": true,
		"action_id": action.ActionID,
	})
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

	// Verify session ownership and status
	var session models.SimulationSession
	if err := sc.DB.Where("session_id = ? AND user_id = ? AND status = 'in_progress'", input.SessionID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid or inactive session"})
		return
	}

	// Get the scenario
	var scenario models.Scenario
	if err := sc.DB.Where("scenario_id = ?", session.ScenarioID).First(&scenario).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Scenario not found"})
		return
	}

	// Score the submission using the deep evaluator module
	result := services.ScoreSubmission(scenario, models.SubmitScenarioInput{
		SessionID:          input.SessionID,
		IsTruePositive:     input.IsTruePositive,
		TPFPSelected:       input.TPFPSelected,
		SelectedActions:    input.SelectedActions,
		DiscoveredFindings: input.DiscoveredFindings,
	})

	tx := sc.DB.Begin()

	// Record the TP/FP decision action
	var maxStep int
	tx.Model(&models.SessionAction{}).
		Where("session_id = ?", input.SessionID).
		Select("COALESCE(MAX(step_order), 0)").
		Scan(&maxStep)

	tpfpAction := models.SessionAction{
		SessionID:  input.SessionID,
		StepOrder:  maxStep + 1,
		ActionType: "triage_alert",
		Payload: map[string]interface{}{
			"user_choice":    input.IsTruePositive,
			"correct_answer": scenario.IsTruePositive,
		},
		IsCorrect: &result.TPFPCorrect,
		Points:    int(result.TPFPPoints),
		Timestamp: time.Now(),
	}
	tx.Create(&tpfpAction)

	// Record the response action
	responseAction := models.SessionAction{
		SessionID:  input.SessionID,
		StepOrder:  maxStep + 2,
		ActionType: "respond",
		Payload: map[string]interface{}{
			"selected_actions": input.SelectedActions,
		},
		Points:    int(result.ResponsePoints),
		Timestamp: time.Now(),
	}
	tx.Create(&responseAction)

	// Record findings submission
	findingsAction := models.SessionAction{
		SessionID:  input.SessionID,
		StepOrder:  maxStep + 3,
		ActionType: "submit_decision",
		Payload: map[string]interface{}{
			"discovered_findings": input.DiscoveredFindings,
		},
		Points:    int(result.FindingsPoints),
		Timestamp: time.Now(),
	}
	tx.Create(&findingsAction)

	// Update simulation session with complete evaluation result snapshot
	now := time.Now()
	session.Status = "completed"
	session.FailureClass = "learner_progress"
	session.CalculationVersion = services.CalculationVersion
	session.RubricVersion = services.RubricVersion
	session.FinalScore = &result.TotalScore
	session.CompletedAt = &now
	session.SkillGap = result
	session.TotalActions = maxStep + 3
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
			tx.Create(&profile)
			if newScore < ProficiencyThreshold {
				allDomainsAboveThreshold = false
			}
		} else {
			// Weighted average update using Evaluator calculation
			profile.ProficiencyScore = services.CalculateUpdatedProficiency(profile.ProficiencyScore, newScore)
			profile.ScenariosCompleted++
			now := time.Now()
			profile.LastPracticed = &now
			tx.Save(&profile)
			if profile.ProficiencyScore < ProficiencyThreshold {
				allDomainsAboveThreshold = false
			}
		}
	}

	// Check total profiles count across all 5 domains
	var totalProfilesCount int64
	tx.Model(&models.UserSkillProfile{}).Where("user_id = ?", userID).Count(&totalProfilesCount)
	if totalProfilesCount < 5 {
		allDomainsAboveThreshold = false
	}

	// Determine completion reason based on stopping rules
	var totalCompletedCount int64
	tx.Model(&models.SimulationSession{}).Where("user_id = ? AND status = 'completed'", userID).Count(&totalCompletedCount)
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

	tx.Commit()

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

// sanitizeScenario removes answer data before sending to the frontend
func sanitizeScenario(s models.Scenario) ScenarioResponse {
	// Strip is_correct and explanation from playbook_steps
	sanitizedSteps := sanitizePlaybookSteps(s.PlaybookSteps)

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

// sanitizePlaybookSteps removes is_correct, explanation, and points from playbook step options
func sanitizePlaybookSteps(steps any) any {
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
		"rubric_version":       services.RubricVersion,
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
