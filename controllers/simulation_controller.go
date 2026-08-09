package controllers

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

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
		_ = db.AutoMigrate(&models.SimulationSession{})
	}
	return &SimulationController{DB: db}
}

// --- Response types ---

type DomainStatus struct {
	DomainID         string  `json:"domain_id"`
	DomainName       string  `json:"domain_name"`
	ProficiencyScore float64 `json:"proficiency_score"`
	Threshold        float64 `json:"threshold"`
	Passed           bool    `json:"passed"`
}

type SimulationStatusResponse struct {
	NeedsTraining      bool           `json:"needs_training"`
	Domains            []DomainStatus `json:"domains"`
	CompletedScenarios int            `json:"completed_scenarios"`
	MaxScenarios       int            `json:"max_scenarios"`
	AllDomainsPassed   bool           `json:"all_domains_passed"`
}

type ScenarioResponse struct {
	ScenarioID  string `json:"scenario_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DomainID    string `json:"domain_id"`
	Difficulty  int    `json:"difficulty"`
	InitialLogs any    `json:"initial_logs"`
	// PlaybookSteps are sent without is_correct and explanation for the user to answer
	PlaybookSteps any `json:"playbook_steps"`
	SanitizedFindings any `json:"sanitized_findings"`
}

type LogQueryInput struct {
	SessionID string `json:"session_id" binding:"required"`
	Query     string `json:"query" binding:"required"`
}

type SubmitScenarioInput struct {
	SessionID        string   `json:"session_id" binding:"required"`
	IsTruePositive   bool     `json:"is_true_positive"`
	TPFPSelected     bool     `json:"tp_fp_selected" binding:"required"`
	SelectedActions  []string `json:"selected_actions" binding:"required"`
	DiscoveredFindings []string `json:"discovered_findings"`
}

// --- Domain name helper ---

var domainNames = map[string]string{
	"domain1": "General Security Concepts",
	"domain2": "Threats, Vulnerabilities, and Mitigations",
	"domain3": "Security Architecture",
	"domain4": "Security Operations",
	"domain5": "Security Program Management and Oversight",
}

// GET /api/simulation/status
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

// POST /api/simulation/start
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

	// Attempt AI Dynamic Scenario Generation if LLM_API_KEY is configured
	if os.Getenv("LLM_API_KEY") != "" {
		domainName := domainNames[weakestDomain]
		// Use Background context with 120s timeout so connection cancels don't abort generation
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
					break
				}
			}
		}

		if err != nil {
			// No scenarios left for any weak domain — try any remaining scenario
			err = sc.DB.Where("status = 'active' AND scenario_id NOT IN (?)", completedSubquery).
				Order("RANDOM()").
				First(&scenario).Error
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
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"error":          "Failed to load scenario. Please retry.",
				"completed":      false,
				"pool_exhausted": true,
			})
		}
		return
	}

	// Create simulation session
	session := models.SimulationSession{
		UserID:         userID.(string),
		ScenarioID:     scenario.ScenarioID,
		Status:         "in_progress",
		GenerationType: generationType,
		FallbackReason: fallbackReason,
		StartedAt:      time.Now(),
	}

	if err := sc.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create simulation session"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session":  session,
		"scenario": sanitizeScenario(scenario),
	})
}

// GET /api/simulation/session
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

// POST /api/simulation/log-query
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

// POST /api/simulation/submit
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

	// Score the submission
	result := scoreSubmission(scenario, input)

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

	// Update simulation session
	now := time.Now()
	session.Status = "completed"
	session.FinalScore = &result.TotalScore
	session.CompletedAt = &now
	session.SkillGap = result.DomainScores
	session.TotalActions = maxStep + 3

	if err := tx.Save(&session).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
		return
	}

	// Update user_skill_profiles with weighted average
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
		} else {
			// Weighted average update
			profile.ProficiencyScore = WeightOld*profile.ProficiencyScore + WeightNew*newScore
			profile.ProficiencyScore = math.Round(profile.ProficiencyScore*100) / 100
			profile.ScenariosCompleted++
			now := time.Now()
			profile.LastPracticed = &now
			tx.Save(&profile)
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"result": result,
	})
}

// GET /api/simulation/result/:sessionId
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
		"domain_proficiencies": domainProficiencies,
	})
}

// --- Scoring logic ---

type ScenarioResult struct {
	TotalScore      float64                `json:"total_score"`
	TPFPCorrect     bool                   `json:"tp_fp_correct"`
	TPFPPoints      float64                `json:"tp_fp_points"`
	FindingsPoints  float64                `json:"findings_points"`
	ResponsePoints  float64                `json:"response_points"`
	DomainScores    map[string]interface{} `json:"domain_scores"`
	DomainScoreMap  map[string]float64     `json:"-"` // internal use for proficiency update
	FindingsDetail  []FindingResult        `json:"findings_detail"`
	ResponseDetail  []ResponseActionResult `json:"response_detail"`
	TPFPExplanation string                 `json:"tp_fp_explanation"`
}

type FindingResult struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	DomainID    string `json:"domain_id"`
	Found       bool   `json:"found"`
	Points      int    `json:"points"`
	Explanation string `json:"explanation"`
}

type ResponseActionResult struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Phase       string `json:"phase"`
	DomainID    string `json:"domain_id"`
	Selected    bool   `json:"selected"`
	IsCorrect   bool   `json:"is_correct"`
	Points      int    `json:"points"`
	Earned      int    `json:"earned"`
	Explanation string `json:"explanation"`
}

func scoreSubmission(scenario models.Scenario, input SubmitScenarioInput) ScenarioResult {
	result := ScenarioResult{
		DomainScores:   make(map[string]interface{}),
		DomainScoreMap: make(map[string]float64),
	}

	// Track raw points per domain
	domainEarned := make(map[string]float64)
	domainMaxPoints := make(map[string]float64)

	// --- 1. Score TP/FP Decision ---
	result.TPFPCorrect = input.IsTruePositive == scenario.IsTruePositive
	result.TPFPExplanation = scenario.TpFpExplanation
	tpfpMaxPoints := 20.0
	if result.TPFPCorrect {
		result.TPFPPoints = tpfpMaxPoints
		domainEarned[scenario.DomainID] += tpfpMaxPoints
	}
	domainMaxPoints[scenario.DomainID] += tpfpMaxPoints

	// --- 2. Score Key Findings ---
	expectedOutcomes, ok := scenario.ExpectedOutcomes.(map[string]interface{})
	if ok {
		keyFindings, ok := expectedOutcomes["key_findings"].([]interface{})
		if ok {
			discoveredSet := make(map[string]bool)
			for _, fID := range input.DiscoveredFindings {
				discoveredSet[fID] = true
			}

			for _, kf := range keyFindings {
				finding, ok := kf.(map[string]interface{})
				if !ok {
					continue
				}

				fID, _ := finding["id"].(string)
				fDesc, _ := finding["description"].(string)
				fDomain, _ := finding["domain_id"].(string)
				fPoints := toFloat64(finding["points"])
				fExplanation, _ := finding["explanation"].(string)

				found := discoveredSet[fID]

				fr := FindingResult{
					ID:          fID,
					Description: fDesc,
					DomainID:    fDomain,
					Found:       found,
					Points:      int(fPoints),
					Explanation: fExplanation,
				}

				if found {
					result.FindingsPoints += fPoints
					domainEarned[fDomain] += fPoints
				}
				domainMaxPoints[fDomain] += fPoints
				result.FindingsDetail = append(result.FindingsDetail, fr)
			}
		}
	}

	// --- 3. Score Response Actions ---
	playbookSteps, ok := scenario.PlaybookSteps.(map[string]interface{})
	if ok {
		selectedSet := make(map[string]bool)
		for _, aID := range input.SelectedActions {
			selectedSet[aID] = true
		}

		phases := []string{"containment", "eradication", "recovery"}
		for _, phase := range phases {
			actions, ok := playbookSteps[phase].([]interface{})
			if !ok {
				continue
			}

			for _, a := range actions {
				action, ok := a.(map[string]interface{})
				if !ok {
					continue
				}

				aID, _ := action["id"].(string)
				aLabel, _ := action["label"].(string)
				aCorrect, _ := action["is_correct"].(bool)
				aDomain, _ := action["domain_id"].(string)
				aPoints := toFloat64(action["points"])
				aExplanation, _ := action["explanation"].(string)

				selected := selectedSet[aID]

				earned := 0
				if selected {
					earned = int(aPoints) // positive if correct, negative if incorrect
					result.ResponsePoints += aPoints
					domainEarned[aDomain] += aPoints
				}

				// For max points calculation, only count positive points
				if aPoints > 0 {
					domainMaxPoints[aDomain] += aPoints
				}

				result.ResponseDetail = append(result.ResponseDetail, ResponseActionResult{
					ID:          aID,
					Label:       aLabel,
					Phase:       phase,
					DomainID:    aDomain,
					Selected:    selected,
					IsCorrect:   aCorrect,
					Points:      int(aPoints),
					Earned:      earned,
					Explanation: aExplanation,
				})
			}
		}
	}

	// --- Calculate per-domain percentage scores ---
	for domainID, maxPts := range domainMaxPoints {
		if maxPts <= 0 {
			continue
		}
		earned := domainEarned[domainID]
		if earned < 0 {
			earned = 0 // Floor at 0
		}
		pct := math.Round((earned/maxPts)*100*100) / 100
		if pct > 100 {
			pct = 100
		}
		result.DomainScoreMap[domainID] = pct
		result.DomainScores[domainID] = map[string]interface{}{
			"earned":     earned,
			"max_points": maxPts,
			"percentage": pct,
		}
	}

	// --- Calculate total weighted score ---
	totalMax := 0.0
	totalEarned := 0.0
	for _, maxPts := range domainMaxPoints {
		totalMax += maxPts
	}
	for _, earned := range domainEarned {
		totalEarned += earned
	}
	if totalEarned < 0 {
		totalEarned = 0
	}
	if totalMax > 0 {
		result.TotalScore = math.Round((totalEarned/totalMax)*100*100) / 100
	}

	return result
}

// Helper to convert interface{} to float64
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return 0
	}
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
