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
	"sort"
	"strings"
	"sync"
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
	MinScenarios         = 1    // Minimum scenarios before allowing completion
	ScoreWeightTPFP      = 0.20 // 20% of scenario score
	ScoreWeightFindings  = 0.40 // 40% of scenario score
	ScoreWeightResponse  = 0.40 // 40% of scenario score
)

type userLockEntry struct {
	mu   sync.Mutex
	refs int
}

type SimulationController struct {
	DB                 *gorm.DB
	startMu            sync.Mutex
	userStartLocks     map[string]*userLockEntry
	transMu            sync.Mutex
	scenarioTransLocks map[string]*userLockEntry
	upgradeStoryFn     func(context.Context, *models.Scenario) (string, map[string]any, error)
}

func NewSimulationController(db *gorm.DB) *SimulationController {
	if db != nil {
		_ = db.AutoMigrate(&models.SimulationSession{}, &models.ScenarioSnapshot{})
		_ = services.EnsureScenarioPlaybookQuality(db, "scenarios_seed.json")
	}
	return &SimulationController{
		DB:                 db,
		userStartLocks:     make(map[string]*userLockEntry),
		scenarioTransLocks: make(map[string]*userLockEntry),
	}
}

func acquireKeyedLock(mu *sync.Mutex, locksPtr *map[string]*userLockEntry, key string) func() {
	mu.Lock()
	if *locksPtr == nil {
		*locksPtr = make(map[string]*userLockEntry)
	}
	locks := *locksPtr
	entry, ok := locks[key]
	if !ok {
		entry = &userLockEntry{}
		locks[key] = entry
	}
	entry.refs++
	mu.Unlock()

	entry.mu.Lock()
	return func() {
		entry.mu.Unlock()
		mu.Lock()
		entry.refs--
		if entry.refs <= 0 && (*locksPtr)[key] == entry {
			delete(*locksPtr, key)
		}
		mu.Unlock()
	}
}

// lockUserStart serializes StartScenario calls per user so concurrent browser reloads,
// React StrictMode double-mounts, or manual retries cannot launch parallel LLM scenario
// generation runs or create duplicate in_progress sessions for the same learner.
func (sc *SimulationController) lockUserStart(userKey string) func() {
	return acquireKeyedLock(&sc.startMu, &sc.userStartLocks, userKey)
}

// lockScenarioTranslation serializes lazy Thai translation backfills per scenario_id
// so concurrent session requests do not issue duplicate LLM translation calls.
func (sc *SimulationController) lockScenarioTranslation(scenarioID string) func() {
	return acquireKeyedLock(&sc.transMu, &sc.scenarioTransLocks, scenarioID)
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
	return selectedLogIndicesFromDB(sc.DB, sessionID)
}

// selectedLogIndicesTx replays the session's selection inside the caller's
// transaction, so submission scores the exact set that a concurrent toggle
// cannot slip past.
func selectedLogIndicesTx(tx *gorm.DB, sessionID string) ([]int, error) {
	var actions []models.SessionAction
	if err := tx.
		Where("session_id = ? AND action_type = ?", sessionID, services.SelectEvidenceActionType).
		Order("step_order asc").
		Find(&actions).Error; err != nil {
		return nil, err
	}
	return services.ReplaySelectedLogIndices(actions), nil
}

func selectedLogIndicesFromDB(db *gorm.DB, sessionID string) []int {
	var actions []models.SessionAction
	if err := db.
		Where("session_id = ? AND action_type = ?", sessionID, services.SelectEvidenceActionType).
		Order("step_order asc").
		Find(&actions).Error; err != nil {
		log.Printf("Failed to replay evidence selection for session %s: %v", sessionID, err)
		return []int{}
	}
	return services.ReplaySelectedLogIndices(actions)
}

// normalizeSelectionIndices sorts and de-duplicates a selection for comparison.
func normalizeSelectionIndices(indices []int) []int {
	seen := make(map[int]struct{}, len(indices))
	normalized := make([]int, 0, len(indices))
	for _, index := range indices {
		if _, exists := seen[index]; exists {
			continue
		}
		seen[index] = struct{}{}
		normalized = append(normalized, index)
	}
	sort.Ints(normalized)
	return normalized
}

func equalIntSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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
	return sessionScenarioFromDB(sc.DB, session)
}

// sessionScenarioTx is sessionScenario bound to the caller's transaction so
// submission reads the exact snapshot it locked.
func sessionScenarioTx(tx *gorm.DB, session models.SimulationSession) (models.Scenario, bool, error) {
	return sessionScenarioFromDB(tx, session)
}

func sessionScenarioFromDB(db *gorm.DB, session models.SimulationSession) (models.Scenario, bool, error) {
	var snapshot models.ScenarioSnapshot
	err := db.Where("session_id = ?", session.SessionID).First(&snapshot).Error
	if err == nil {
		scenario, decodeErr := decodeScenarioSnapshot(snapshot)
		if decodeErr != nil {
			return scenario, true, decodeErr
		}
		// Pre-storytelling / pre-Spec-11 immutable snapshots may omit story or translations;
		// supplement ONLY narrative fields (description & translations) from live scenario row
		// without altering immutable initial_logs, playbook_steps, findings, or verdict.
		if (!services.HasCompleteStory(&scenario) || !hasScenarioTranslations(scenario.Translations)) && db != nil && session.ScenarioID != "" {
			var live models.Scenario
			if liveErr := db.Select("description, translations").Where("scenario_id = ?", session.ScenarioID).First(&live).Error; liveErr == nil {
				if services.HasCompleteStory(&live) {
					if strings.TrimSpace(live.Description) != "" {
						scenario.Description = live.Description
					}
					scenario.Translations = live.Translations
				} else if !hasScenarioTranslations(scenario.Translations) && hasScenarioTranslations(live.Translations) {
					scenario.Translations = live.Translations
				}
			}
		}
		return scenario, true, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Scenario{}, false, err
	}

	var scenario models.Scenario
	if err := db.Where("scenario_id = ?", session.ScenarioID).First(&scenario).Error; err != nil {
		return scenario, false, err
	}
	return scenario, false, nil
}

func hasScenarioTranslations(raw any) bool {
	m, ok := raw.(map[string]interface{})
	if !ok || len(m) == 0 {
		return false
	}
	th, ok := m["th"].(map[string]interface{})
	return ok && len(th) > 0
}

// decodeScenarioSnapshot decodes the JSONB ScenarioData payload into the
// Scenario shape used by scoring and review. The snapshot, not the live
// scenarios row, defines the evidence index space for a session.
func decodeScenarioSnapshot(snapshot models.ScenarioSnapshot) (models.Scenario, error) {
	var scenario models.Scenario
	encoded, err := json.Marshal(snapshot.ScenarioData)
	if err != nil {
		return scenario, err
	}
	if err := json.Unmarshal(encoded, &scenario); err != nil {
		return scenario, err
	}
	if scenario.ScenarioID == "" {
		scenario.ScenarioID = snapshot.ScenarioID
	}
	return scenario, nil
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

func writeSimError(c *gin.Context, status int, code string, msg string) {
	c.JSON(status, gin.H{
		"error": msg,
		"code":  code,
	})
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
		writeSimError(c, http.StatusUnauthorized, "SIM_UNAUTHORIZED", "Unauthorized")
		return
	}

	// Get all user skill profiles
	var profiles []models.UserSkillProfile
	sc.DB.Where("user_id = ?", userID).Find(&profiles)

	if len(profiles) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"error":          "No skill profile found. Please complete the pre-test first.",
			"code":           "SIM_NO_SKILL_PROFILE",
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

	completedTheoryDomains, completedSubtopics, theoryCompleted, _ := sc.getUserTheoryProgressSummary(userID)

	maxScenarios := services.ResolveMaxScenariosCap()
	needsTraining := !allPassed || int(completedCount) < MinScenarios
	if int(completedCount) >= maxScenarios {
		needsTraining = false
	}

	c.JSON(http.StatusOK, SimulationStatusResponse{
		NeedsTraining:          needsTraining,
		Domains:                domains,
		CompletedScenarios:     int(completedCount),
		MaxScenarios:           maxScenarios,
		AllDomainsPassed:       allPassed,
		TheoryCompleted:        theoryCompleted,
		CompletedTheoryDomains: completedTheoryDomains,
		CompletedSubtopicCount: len(completedSubtopics),
		TotalSubtopics:         len(services.OrderedTheorySubtopics),
		CompletedSubtopics:     completedSubtopics,
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
// @Failure 403 {object} dto.TheoryGateErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /simulation/start [post]
func (sc *SimulationController) StartScenario(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		writeSimError(c, http.StatusUnauthorized, "SIM_UNAUTHORIZED", "Unauthorized")
		return
	}

	unlock := sc.lockUserStart(fmt.Sprint(userID))
	defer unlock()

	// Check for existing in-progress session
	var activeSession models.SimulationSession
	err := sc.DB.Where("user_id = ? AND status = 'in_progress'", userID).
		Order("started_at DESC").
		First(&activeSession).Error
	if err == nil {
		// Return existing active session with scenario data
		scenario, fromSnapshot, scenarioErr := sc.sessionScenario(activeSession)
		if scenarioErr != nil {
			log.Printf("Failed to load scenario for session %s: %v", activeSession.SessionID, scenarioErr)
			writeSimError(c, http.StatusInternalServerError, "SIM_SCENARIO_LOAD_FAILED", "Failed to load scenario")
			return
		}
		sc.prepareSessionScenarioResponse(c, &scenario, fromSnapshot)
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
			"code":           "SIM_NO_SKILL_PROFILE",
			"needs_training": true,
		})
		return
	}

	// Stage 2 Theory Gate: Learners must complete all 5 Domain Essentials before
	// starting new simulation scenarios (unless BYPASS_THEORY_GATE=true or Staff).
	if gateErr := evaluateStage2TheoryGate(os.Getenv("BYPASS_THEORY_GATE"), c.GetString("role"), func() ([]string, []string) {
		completedTheoryDomains, completedSubtopics, _, _ := sc.getUserTheoryProgressSummary(userID)
		return completedSubtopics, completedTheoryDomains
	}); gateErr != nil {
		c.JSON(http.StatusForbidden, gateErr)
		return
	}

	var completedCount int64
	sc.DB.Model(&models.SimulationSession{}).
		Where("user_id = ? AND status = 'completed'", userID).
		Count(&completedCount)

	maxScenarios := services.ResolveMaxScenariosCap()
	// Check max cap
	if int(completedCount) >= maxScenarios {
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

	history := sc.loadScenarioDiversityHistory(userID)
	divCtx := services.BuildScenarioDiversityContext(weakestDomain, history)

	var scenario models.Scenario
	generationType := "static_fallback"
	fallbackReason := ""
	selectionReason := fmt.Sprintf("lowest_proficiency: %s (score: %.1f%%, threshold: %.1f%%)", weakestDomain, weakestScore, ProficiencyThreshold)

	// Attempt AI Dynamic Scenario Generation if LLM_API_KEY is configured
	if os.Getenv("LLM_API_KEY") != "" {
		domainName := domainNames[weakestDomain]
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		aiScenario, llmErr := services.GenerateLLMScenarioWithContext(ctx, weakestDomain, domainName, weakestScore, profiles, divCtx)
		if llmErr == nil && aiScenario != nil {
			if err := sc.DB.Create(aiScenario).Error; err == nil {
				scenario = *aiScenario
				generationType = "ai_generated"
				selectionReason = fmt.Sprintf(
					"lowest_proficiency: %s (score: %.1f%%, threshold: %.1f%%) | subtopic: %s | archetype: %s | style: %s",
					weakestDomain, weakestScore, ProficiencyThreshold,
					divCtx.TargetSubTopicID, divCtx.TargetArchetype, divCtx.TargetBriefingStyle,
				)
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
		if allPassed || int(completedCount) >= maxScenarios {
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
		MaxScenarios:       maxScenarios,
		TimeLimitSeconds:   7200,
		CalculationVersion: services.CalculationVersion,
		RubricVersion:      services.RubricVersion,
		BlueprintVersion:   "blueprint-v1",
		StartedAt:          time.Now(),
	}

	if err := sc.DB.Create(&session).Error; err != nil {
		writeSimError(c, http.StatusInternalServerError, "SIM_SESSION_CREATE_FAILED", "Failed to create simulation session")
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
		writeSimError(c, http.StatusUnauthorized, "SIM_UNAUTHORIZED", "Unauthorized")
		return
	}

	// First, check for an in-progress session
	var activeSession models.SimulationSession
	err := sc.DB.Where("user_id = ? AND status = 'in_progress'", userID).
		Order("started_at DESC").
		First(&activeSession).Error
	if err == nil {
		scenario, fromSnapshot, scenarioErr := sc.sessionScenario(activeSession)
		if scenarioErr != nil {
			log.Printf("Failed to load scenario for session %s: %v", activeSession.SessionID, scenarioErr)
			writeSimError(c, http.StatusInternalServerError, "SIM_SCENARIO_LOAD_FAILED", "Failed to load scenario")
			return
		}
		sc.prepareSessionScenarioResponse(c, &scenario, fromSnapshot)
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

	c.JSON(http.StatusNotFound, gin.H{
		"message": "No active simulation session found",
		"code":    "SIM_NO_ACTIVE_SESSION",
	})
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
		writeSimError(c, http.StatusUnauthorized, "SIM_UNAUTHORIZED", "Unauthorized")
		return
	}

	var input LogQueryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeSimError(c, http.StatusBadRequest, "SIM_BAD_REQUEST", err.Error())
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
		writeSimError(c, http.StatusUnauthorized, "SIM_UNAUTHORIZED", "Unauthorized")
		return
	}

	var input dto.SelectLogRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		writeSimError(c, http.StatusBadRequest, "SIM_BAD_REQUEST", err.Error())
		return
	}
	if input.LogIndex == nil || input.Selected == nil {
		writeSimError(c, http.StatusBadRequest, "SIM_BAD_REQUEST", "log_index and selected are required")
		return
	}

	var session models.SimulationSession
	if err := sc.DB.Where("session_id = ?", input.SessionID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeSimError(c, http.StatusNotFound, "SIM_SESSION_NOT_FOUND", "Simulation session not found")
			return
		}
		writeSimError(c, http.StatusInternalServerError, "SIM_SESSION_LOAD_FAILED", "Failed to load simulation session")
		return
	}
	if session.UserID != userID.(string) {
		writeSimError(c, http.StatusForbidden, "SIM_SESSION_FORBIDDEN", "Simulation session is not owned by caller")
		return
	}
	if session.Status != "in_progress" {
		writeSimError(c, http.StatusForbidden, "SIM_SESSION_NOT_ACTIVE", "Simulation session is not in progress")
		return
	}

	// Validate against the session's immutable snapshot, not the live scenario
	// row, so reseeding or editing a scenario cannot shift evidence indices.
	logs, err := sc.sessionScenarioLogs(session)
	if err != nil {
		switch {
		case errors.Is(err, errSelectionLogsUnavailable):
			writeSimError(c, http.StatusBadRequest, "SIM_LOGS_UNAVAILABLE", errSelectionLogsUnavailable.Error())
		case errors.Is(err, gorm.ErrRecordNotFound):
			writeSimError(c, http.StatusNotFound, "SIM_SCENARIO_NOT_FOUND", "Scenario not found")
		default:
			log.Printf("Failed to load scenario logs for session %s: %v", session.SessionID, err)
			writeSimError(c, http.StatusInternalServerError, "SIM_LOGS_LOAD_FAILED", "Failed to load scenario logs")
		}
		return
	}
	if err := validateSelectionLogIndex(logs, *input.LogIndex); err != nil {
		writeSimError(c, http.StatusBadRequest, "SIM_INVALID_LOG_INDEX", err.Error())
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
		writeSimError(c, http.StatusUnauthorized, "SIM_UNAUTHORIZED", "Unauthorized")
		return
	}

	var input dto.ClearLogSelectionRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		writeSimError(c, http.StatusBadRequest, "SIM_BAD_REQUEST", err.Error())
		return
	}

	var session models.SimulationSession
	if err := sc.DB.Where("session_id = ?", input.SessionID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeSimError(c, http.StatusNotFound, "SIM_SESSION_NOT_FOUND", "Simulation session not found")
			return
		}
		writeSimError(c, http.StatusInternalServerError, "SIM_SESSION_LOAD_FAILED", "Failed to load simulation session")
		return
	}
	if session.UserID != userID.(string) {
		writeSimError(c, http.StatusForbidden, "SIM_SESSION_FORBIDDEN", "Simulation session is not owned by caller")
		return
	}
	if session.Status != "in_progress" {
		writeSimError(c, http.StatusForbidden, "SIM_SESSION_NOT_ACTIVE", "Simulation session is not in progress")
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
		writeSimError(c, http.StatusNotFound, "SIM_SESSION_NOT_FOUND", "Simulation session not found")
	case errors.Is(err, errSelectionSessionForbidden):
		writeSimError(c, http.StatusForbidden, "SIM_SESSION_FORBIDDEN", "Simulation session is not owned by caller")
	case errors.Is(err, errSelectionSessionInactive):
		writeSimError(c, http.StatusForbidden, "SIM_SESSION_NOT_ACTIVE", "Simulation session is not in progress")
	default:
		log.Printf("Failed to record session action: %v", err)
		writeSimError(c, http.StatusInternalServerError, "SIM_ACTION_RECORD_FAILED", "Failed to record session action")
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
// @Success 200 {object} dto.SubmitScenarioResponse{result=models.ScenarioResult}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /simulation/submit [post]
func (sc *SimulationController) SubmitScenario(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		writeSimError(c, http.StatusUnauthorized, "SIM_UNAUTHORIZED", "Unauthorized")
		return
	}

	var input SubmitScenarioInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeSimError(c, http.StatusBadRequest, "SIM_BAD_REQUEST", err.Error())
		return
	}

	tx := sc.DB.Begin()
	if tx.Error != nil {
		writeSimError(c, http.StatusInternalServerError, "SIM_SUBMIT_FAILED", "Failed to start submission")
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

	// Load the scenario from the immutable session snapshot. Legacy sessions
	// created before snapshots existed fall back to the live scenario row.
	scenario, _, scenarioErr := sessionScenarioTx(tx, session)
	if scenarioErr != nil {
		tx.Rollback()
		writeSimError(c, http.StatusInternalServerError, "SIM_SCENARIO_NOT_FOUND", "Scenario not found")
		return
	}

	// Replay the server-recorded evidence actions under the session lock. The
	// client selection is reconciliation-only: a mismatch is logged and the
	// server set wins.
	serverSelection, err := selectedLogIndicesTx(tx, input.SessionID)
	if err != nil {
		tx.Rollback()
		writeSimError(c, http.StatusInternalServerError, "SIM_EVIDENCE_REPLAY_FAILED", "Failed to replay evidence selection")
		return
	}
	clientSelection := []int{}
	selectionMismatch := false
	if input.SelectedLogIndices != nil {
		clientSelection = normalizeSelectionIndices(input.SelectedLogIndices)
		selectionMismatch = !equalIntSlices(serverSelection, clientSelection)
		if selectionMismatch {
			log.Printf(
				"Evidence selection mismatch for session %s: server=%v client=%v; scoring the server set",
				input.SessionID, serverSelection, clientSelection,
			)
		}
	}

	// Score the submission from the replayed server selection.
	result := services.ScoreSubmission(scenario, models.SubmitScenarioInput{
		SessionID:          input.SessionID,
		IsTruePositive:     input.IsTruePositive,
		TPFPSelected:       input.TPFPSelected,
		SelectedActions:    input.SelectedActions,
		SelectedLogIndices: serverSelection,
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

	// Record the scored evidence selection for audit. The legacy findings list
	// is no longer part of the submit contract.
	findingsAction, err := appendSessionActionTx(tx, input.SessionID, sessionActionSpec{
		ActionType: "submit_decision",
		Payload: map[string]interface{}{
			"selected_log_indices":        serverSelection,
			"client_selected_log_indices": clientSelection,
			"selection_mismatch":          selectionMismatch,
			"findings_points":             result.FindingsPoints,
			"evidence_penalty":            result.EvidencePenalty,
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
				writeSimError(c, http.StatusInternalServerError, "SIM_PROFILE_UPDATE_FAILED", "Failed to update skill profile")
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
				writeSimError(c, http.StatusInternalServerError, "SIM_PROFILE_UPDATE_FAILED", "Failed to update skill profile")
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
		writeSimError(c, http.StatusInternalServerError, "SIM_PROFILE_UPDATE_FAILED", "Failed to update skill profile")
		return
	}
	if totalProfilesCount < 5 {
		allDomainsAboveThreshold = false
	}

	// Determine completion reason based on stopping rules
	var totalCompletedCount int64
	if err := tx.Model(&models.SimulationSession{}).Where("user_id = ? AND status = 'completed'", userID).Count(&totalCompletedCount).Error; err != nil {
		tx.Rollback()
		writeSimError(c, http.StatusInternalServerError, "SIM_SESSION_UPDATE_FAILED", "Failed to update session")
		return
	}
	// Add current session (which is about to be saved as completed)
	totalCompletedCount++

	maxScenarios := services.ResolveMaxScenariosCap()
	if allDomainsAboveThreshold {
		session.CompletionReason = "all_domains_passed"
	} else if int(totalCompletedCount) >= maxScenarios {
		session.CompletionReason = "max_scenarios_reached"
	} else if session.ElapsedSeconds >= session.TimeLimitSeconds {
		session.CompletionReason = "time_limit_reached"
	} else {
		session.CompletionReason = "learner_completed"
	}

	if err := tx.Save(&session).Error; err != nil {
		tx.Rollback()
		writeSimError(c, http.StatusInternalServerError, "SIM_SESSION_UPDATE_FAILED", "Failed to update session")
		return
	}

	if err := tx.Commit().Error; err != nil {
		writeSimError(c, http.StatusInternalServerError, "SIM_SUBMIT_FAILED", "Failed to submit scenario")
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
// @Success 200 {object} dto.SimulationResultResponse{result=models.ScenarioResult}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /simulation/result/{sessionId} [get]
func (sc *SimulationController) GetResult(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		writeSimError(c, http.StatusUnauthorized, "SIM_UNAUTHORIZED", "Unauthorized")
		return
	}

	sessionID := c.Param("sessionId")

	var session models.SimulationSession
	if err := sc.DB.Where("session_id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		writeSimError(c, http.StatusNotFound, "SIM_SESSION_NOT_FOUND", "Session not found")
		return
	}

	if session.Status != "completed" {
		writeSimError(c, http.StatusBadRequest, "SIM_SESSION_NOT_COMPLETED", "Session is not completed yet")
		return
	}

	// Prefer the immutable session snapshot so the result page shows the same
	// telemetry and evidence indices the learner investigated.
	scenario, _, scenarioErr := sc.sessionScenario(session)
	if scenarioErr != nil {
		log.Printf("Failed to load scenario for result session %s: %v", sessionID, scenarioErr)
	} else {
		reqCtx := context.Background()
		if c != nil && c.Request != nil {
			reqCtx = c.Request.Context()
		}
		sc.ensureScenarioStoryAndTranslation(reqCtx, &scenario)
	}

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

	// The learner payload exposes only how many Key Findings exist. Evidence
	// indices, descriptions, and verdict markers stay server-side.
	return ScenarioResponse{
		ScenarioID:    s.ScenarioID,
		Title:         s.Title,
		Description:   s.Description,
		DomainID:      s.DomainID,
		Difficulty:    s.Difficulty,
		InitialLogs:   s.InitialLogs,
		PlaybookSteps: sanitizedSteps,
		TotalFindings: countKeyFindings(s.ExpectedOutcomes),
		Translations:  sanitizeScenarioTranslations(s.Translations),
	}
}

// sanitizeScenarioTranslations exposes ONLY non-spoiler localized fields (title,
// description, and playbook_steps[id].label) in pre-submission ScenarioResponse.
// tp_fp_explanation, findings, and playbook explanations MUST NEVER be exposed.
func sanitizeScenarioTranslations(raw any) any {
	transMap, ok := raw.(map[string]interface{})
	if !ok || len(transMap) == 0 {
		return nil
	}

	sanitizedLocales := make(map[string]interface{})
	for locale, rawLocalePayload := range transMap {
		localeMap, ok := rawLocalePayload.(map[string]interface{})
		if !ok || len(localeMap) == 0 {
			continue
		}

		out := make(map[string]interface{})
		if title, ok := localeMap["title"].(string); ok && strings.TrimSpace(title) != "" {
			out["title"] = title
		}
		if desc, ok := localeMap["description"].(string); ok && strings.TrimSpace(desc) != "" {
			out["description"] = desc
		}

		if pbRaw, ok := localeMap["playbook_steps"]; ok && pbRaw != nil {
			sanitizedPB := make(map[string]interface{})
			switch pbTyped := pbRaw.(type) {
			case map[string]interface{}:
				for actionID, rawStep := range pbTyped {
					switch stepTyped := rawStep.(type) {
					case map[string]interface{}:
						if lbl, ok := stepTyped["label"].(string); ok && strings.TrimSpace(lbl) != "" {
							sanitizedPB[actionID] = map[string]interface{}{"label": lbl}
						}
					case map[string]string:
						if lbl := strings.TrimSpace(stepTyped["label"]); lbl != "" {
							sanitizedPB[actionID] = map[string]interface{}{"label": stepTyped["label"]}
						}
					}
				}
			case map[string]map[string]string:
				for actionID, stepTyped := range pbTyped {
					if lbl := strings.TrimSpace(stepTyped["label"]); lbl != "" {
						sanitizedPB[actionID] = map[string]interface{}{"label": stepTyped["label"]}
					}
				}
			}
			if len(sanitizedPB) > 0 {
				out["playbook_steps"] = sanitizedPB
			}
		}

		if len(out) > 0 {
			sanitizedLocales[locale] = out
		}
	}

	if len(sanitizedLocales) == 0 {
		return nil
	}
	return sanitizedLocales
}

// countKeyFindings counts the scenario's authored Key Findings regardless of
// whether the JSONB payload decoded to a generic or typed map slice.
func countKeyFindings(expectedOutcomes any) int {
	switch outcomes := expectedOutcomes.(type) {
	case map[string]interface{}:
		return len(findingsFromOutcomes(outcomes["key_findings"]))
	case []interface{}:
		return len(outcomes)
	default:
		return 0
	}
}

func findingsFromOutcomes(value any) []map[string]interface{} {
	switch findings := value.(type) {
	case []interface{}:
		out := make([]map[string]interface{}, 0, len(findings))
		for _, item := range findings {
			if finding, ok := item.(map[string]interface{}); ok {
				out = append(out, finding)
			}
		}
		return out
	case []map[string]interface{}:
		return findings
	default:
		return nil
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
		writeSimError(c, http.StatusUnauthorized, "SIM_UNAUTHORIZED", "Unauthorized")
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

	domainProfileDTOs := make([]dto.DomainStatusDTO, 0, len(profiles))
	for _, p := range profiles {
		domainProfileDTOs = append(domainProfileDTOs, dto.DomainStatusDTO{
			DomainID:         p.DomainID,
			DomainName:       domainNames[p.DomainID],
			ProficiencyScore: math.Round(p.ProficiencyScore*100) / 100,
			Threshold:        ProficiencyThreshold,
			Passed:           p.ProficiencyScore >= ProficiencyThreshold,
		})
	}

	completedTheoryDomains, _, theoryCompleted, theoryCompletedAt := sc.getUserTheoryProgressSummary(userID)
	learnTelemetry := sc.loadLearnSubtopicTelemetry(userID)

	questionnaireSubmitted := false
	var questionnaireSubmittedAt *time.Time
	var csuqOverall *float64
	var nasaTlxOverall *float64
	if sc.DB != nil {
		var evalMetric models.EvaluationMetric
		if err := sc.DB.Where("user_id = ?", userID).Order("eval_date DESC").First(&evalMetric).Error; err == nil {
			questionnaireSubmitted = true
			submittedAt := evalMetric.EvalDate
			questionnaireSubmittedAt = &submittedAt
			summary := services.EvaluationMetricToSummaryDTO(evalMetric)
			if summary.CSUQScores.Subscales.Overall != nil {
				v := *summary.CSUQScores.Subscales.Overall
				csuqOverall = &v
			}
			if summary.NASATLXScores.SchemaVersion != "" {
				v := summary.NASATLXScores.OverallWorkload
				nasaTlxOverall = &v
			}
		}
	}

	c.JSON(http.StatusOK, dto.ResearchSummaryResponse{
		ProtocolVersion:          "protocol-v1",
		CalculationVersion:       services.CalculationVersion,
		RubricVersion:            services.RubricVersion,
		ParticipantID:            userID,
		HasPreTest:               hasPre,
		PreTestScore:             preExam.Score,
		HasPostTest:              hasPost,
		PostTestScore:            postExam.Score,
		ImprovementScore:         scoreImprovement,
		AllDomainsPassed:         allPassed,
		EligibleForPost:          allPassed && len(completedSessions) >= MinScenarios,
		TheoryCompleted:          theoryCompleted,
		CompletedTheoryDomains:   completedTheoryDomains,
		TheoryCompletedAt:        theoryCompletedAt,
		CompletedScenarios:       len(completedSessions),
		MaxScenarios:             services.ResolveMaxScenariosCap(),
		TotalActions:             totalActions,
		TotalElapsedSec:          totalSimElapsed,
		AverageProficiency:       avgProficiency,
		DomainProfiles:           domainProfileDTOs,
		LearnSubtopicTelemetry:   &learnTelemetry,
		QuestionnaireSubmitted:   questionnaireSubmitted,
		QuestionnaireSubmittedAt: questionnaireSubmittedAt,
		CSUQOverall:              csuqOverall,
		NASATLXOverall:           nasaTlxOverall,
	})
}

func (sc *SimulationController) loadLearnSubtopicTelemetry(userID any) dto.LearnSubtopicTelemetryDTO {
	if sc.DB == nil {
		return BuildLearnSubtopicTelemetry(nil, nil)
	}
	var subProg []models.UserTheorySubtopicProgress
	_ = sc.DB.Where("user_id = ?", userID).Find(&subProg).Error
	var sampleRows []models.UserTheorySampleQuestion
	_ = sc.DB.Where("user_id = ?", userID).Find(&sampleRows).Error
	return BuildLearnSubtopicTelemetry(subProg, sampleRows)
}

// BuildLearnSubtopicTelemetry computes Sub-topic completion and formative sample-question
// engagement/accuracy metrics across the canonical 28 SY0-701 objectives.
func BuildLearnSubtopicTelemetry(subProgress []models.UserTheorySubtopicProgress, sampleRows []models.UserTheorySampleQuestion) dto.LearnSubtopicTelemetryDTO {
	completedSubtopics, doneBySubID := buildOrderedCompletedSubtopics(subProgress)

	subToDomain := make(map[string]string, len(services.OrderedTheorySubtopics))
	for _, spec := range services.OrderedTheorySubtopics {
		subToDomain[spec.SubTopicID] = spec.DomainID
	}

	assignedByDomain := make(map[string]int, len(services.OrderedTheoryDomains))
	attemptedByDomain := make(map[string]int, len(services.OrderedTheoryDomains))
	correctByDomain := make(map[string]int, len(services.OrderedTheoryDomains))
	totalAssigned := len(sampleRows)
	totalAttempted := 0
	totalCorrect := 0

	for _, r := range sampleRows {
		domID := subToDomain[r.SubTopicID]
		if domID != "" {
			assignedByDomain[domID]++
		}
		if isSampleAnswered(r.SubmittedAnswer) {
			totalAttempted++
			if domID != "" {
				attemptedByDomain[domID]++
			}
			if r.IsCorrect != nil && *r.IsCorrect {
				totalCorrect++
				if domID != "" {
					correctByDomain[domID]++
				}
			}
		}
	}

	var accuracyPct float64
	if totalAttempted > 0 {
		accuracyPct = math.Round((float64(totalCorrect)/float64(totalAttempted))*10000) / 100
	}

	breakdown := make([]dto.LearnDomainBreakdownDTO, 0, len(services.OrderedTheoryDomains))
	for _, dom := range services.OrderedTheoryDomains {
		subs := services.SubtopicsForDomain(dom.DomainID)
		doneInDom := 0
		for _, sp := range subs {
			if _, ok := doneBySubID[sp.SubTopicID]; ok {
				doneInDom++
			}
		}
		breakdown = append(breakdown, dto.LearnDomainBreakdownDTO{
			DomainID:           dom.DomainID,
			CompletedSubtopics: doneInDom,
			TotalSubtopics:     len(subs),
			SampleAssigned:     assignedByDomain[dom.DomainID],
			SampleAttempted:    attemptedByDomain[dom.DomainID],
			SampleCorrect:      correctByDomain[dom.DomainID],
		})
	}

	return dto.LearnSubtopicTelemetryDTO{
		CompletedSubtopicsCount:  len(completedSubtopics),
		TotalSubtopics:           len(services.OrderedTheorySubtopics),
		CompletedSubtopics:       completedSubtopics,
		SampleQuestionsAssigned:  totalAssigned,
		SampleQuestionsAttempted: totalAttempted,
		SampleQuestionsCorrect:   totalCorrect,
		SampleAccuracyPercent:    accuracyPct,
		DomainBreakdown:          breakdown,
	}
}

func (sc *SimulationController) getUserTheoryProgressSummary(userID any) ([]string, []string, bool, *time.Time) {
	if sc.DB == nil {
		return []string{}, []string{}, false, nil
	}
	var domProgress []models.UserTheoryProgress
	if err := sc.DB.Where("user_id = ?", userID).Find(&domProgress).Error; err != nil {
		return []string{}, []string{}, false, nil
	}
	var subProgress []models.UserTheorySubtopicProgress
	if err := sc.DB.Where("user_id = ?", userID).Find(&subProgress).Error; err != nil {
		return []string{}, []string{}, false, nil
	}
	completedSubtopics, _ := buildOrderedCompletedSubtopics(subProgress)
	completedDomains, byDom := buildEffectiveCompletedDomains(domProgress)
	theoryCompleted := services.IsTheoryGateSatisfied(completedDomains)
	var completedAt *time.Time
	if theoryCompleted {
		var latest time.Time
		for _, p := range byDom {
			if p.CompletedAt.After(latest) {
				latest = p.CompletedAt
			}
		}
		if !latest.IsZero() {
			completedAt = &latest
		}
	}
	return completedDomains, completedSubtopics, theoryCompleted, completedAt
}

// evaluateStage2TheoryGate returns a structured 403 TheoryGateErrorResponse when
// Stage 2 (5/5 Domain Essentials completed) is not yet satisfied, bypassEnv is not
// "true", and the caller is not Staff (specialist/admin).
func evaluateStage2TheoryGate(bypassEnv, role string, loadProgress func() ([]string, []string)) *dto.TheoryGateErrorResponse {
	if bypassEnv == "true" || role == "specialist" || role == "admin" {
		return nil
	}
	completedSubtopics, completedDomains := loadProgress()
	if services.IsTheoryGateSatisfied(completedDomains) {
		return nil
	}
	doneDomCount := len(completedDomains)
	totalDomains := len(services.OrderedTheoryDomains)
	doneSubCount := len(completedSubtopics)
	totalSubtopics := len(services.OrderedTheorySubtopics)
	return &dto.TheoryGateErrorResponse{
		Error:                  "Please complete all 5 Domain Essentials in the Learning Theory Hub before starting simulation scenarios.",
		Code:                   "SIM_THEORY_INCOMPLETE",
		Params:                 map[string]any{"completed_count": doneDomCount, "total_domains": totalDomains},
		Reason:                 "theory_incomplete",
		CompletedCount:         &doneDomCount,
		TotalDomains:           &totalDomains,
		CompletedDomains:       completedDomains,
		CompletedSubtopicCount: &doneSubCount,
		TotalSubtopics:         &totalSubtopics,
		CompletedSubtopics:     completedSubtopics,
	}
}

func (sc *SimulationController) loadScenarioDiversityHistory(userID any) []services.ScenarioHistoryEntry {
	if sc.DB == nil {
		return nil
	}
	var sessions []models.SimulationSession
	if err := sc.DB.
		Where("user_id = ?", userID).
		Order("started_at DESC").
		Limit(10).
		Find(&sessions).Error; err != nil {
		return nil
	}
	return buildHistoryEntriesFromSessions(sessions, func(sess models.SimulationSession) (models.Scenario, bool) {
		scen, _, err := sc.sessionScenario(sess)
		if err != nil {
			return models.Scenario{}, false
		}
		return scen, true
	})
}

func parseSelectionReasonTag(selectionReason, key string) string {
	prefix := strings.ToLower(strings.TrimSpace(key)) + ":"
	for _, part := range strings.Split(selectionReason, "|") {
		trimmed := strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToLower(trimmed), prefix) {
			return strings.TrimSpace(trimmed[len(prefix):])
		}
	}
	return ""
}

func buildHistoryEntriesFromSessions(sessions []models.SimulationSession, resolveScenario func(models.SimulationSession) (models.Scenario, bool)) []services.ScenarioHistoryEntry {
	out := make([]services.ScenarioHistoryEntry, 0, len(sessions))
	for _, sess := range sessions {
		entry := services.ScenarioHistoryEntry{
			SubTopicID:    parseSelectionReasonTag(sess.SelectionReason, "subtopic"),
			BriefingStyle: parseSelectionReasonTag(sess.SelectionReason, "style"),
		}
		if resolveScenario != nil {
			if scen, ok := resolveScenario(sess); ok {
				host, ip := services.ExtractPrimaryScenarioEntities(&scen)
				entry.Title = scen.Title
				entry.TargetHost = host
				entry.AttackerIP = ip
				entry.IsTruePositive = scen.IsTruePositive
			} else {
				entry.SkipCadenceVote = true
			}
		} else {
			entry.SkipCadenceVote = true
		}
		out = append(out, entry)
	}
	return out
}

func (sc *SimulationController) prepareSessionScenarioResponse(c *gin.Context, scenario *models.Scenario, fromSnapshot bool) {
	if !fromSnapshot {
		sc.ensureScenarioRemediated(scenario)
	}
	reqCtx := context.Background()
	if c != nil && c.Request != nil {
		reqCtx = c.Request.Context()
	}
	sc.ensureScenarioStoryAndTranslation(reqCtx, scenario)
}

// ensureScenarioStoryAndTranslation lazily upgrades any legacy scenario lacking complete EN+TH
// 4-beat storytelling metadata (and any missing TH translations) on read, persisting ONLY
// description and translations without altering initial_logs, playbook_steps, findings, or status.
// Fails open cleanly when offline or unconfigured.
func (sc *SimulationController) ensureScenarioStoryAndTranslation(ctx context.Context, scenario *models.Scenario) {
	if sc == nil || scenario == nil {
		return
	}
	if services.HasCompleteStory(scenario) && services.ValidateScenarioTranslationsTh(scenario) == nil {
		return
	}
	if sc.upgradeStoryFn == nil && strings.TrimSpace(os.Getenv("LLM_API_KEY")) == "" {
		return
	}

	if scenario.ScenarioID != "" {
		unlock := sc.lockScenarioTranslation(scenario.ScenarioID)
		defer unlock()

		if sc.DB != nil {
			var live models.Scenario
			if err := sc.DB.Select("description, translations").Where("scenario_id = ?", scenario.ScenarioID).First(&live).Error; err == nil {
				candidate := *scenario
				if strings.TrimSpace(live.Description) != "" {
					candidate.Description = live.Description
				}
				candidate.Translations = live.Translations
				if services.HasCompleteStory(&candidate) && services.ValidateScenarioTranslationsTh(&candidate) == nil {
					scenario.Description = candidate.Description
					scenario.Translations = candidate.Translations
					return
				}
			}
		}
	}

	if ctx == nil {
		ctx = context.Background()
	}
	upgradeCtx, cancel := context.WithTimeout(ctx, 150*time.Second)
	defer cancel()

	upgradeFn := sc.upgradeStoryFn
	if upgradeFn == nil {
		upgradeFn = services.UpgradeScenarioStoryBilingual
	}

	enDesc, translated, err := upgradeFn(upgradeCtx, scenario)
	if err != nil {
		log.Printf("[Story Backfill] Failed to upgrade story for scenario %s: %v", scenario.ScenarioID, err)
		return
	}

	candidate := *scenario
	candidate.Description = enDesc
	candidate.Translations = translated
	if valErr := services.ValidateScenarioTranslationsTh(&candidate); valErr != nil {
		log.Printf("[Story Backfill] Upgraded payload for scenario %s failed TH validation: %v", scenario.ScenarioID, valErr)
		return
	}
	if !services.HasCompleteStory(&candidate) {
		log.Printf("[Story Backfill] Upgraded payload for scenario %s lacks complete EN/TH story beats", scenario.ScenarioID)
		return
	}

	if sc.DB != nil && scenario.ScenarioID != "" {
		if saveErr := sc.DB.Model(&models.Scenario{}).
			Where("scenario_id = ?", scenario.ScenarioID).
			Updates(map[string]interface{}{
				"description":  enDesc,
				"translations": translated,
			}).Error; saveErr != nil {
			log.Printf("[Story Backfill] Failed to persist story upgrade for scenario %s: %v", scenario.ScenarioID, saveErr)
		}
	}
	scenario.Description = enDesc
	scenario.Translations = translated
}
