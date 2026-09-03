package dto

// DomainStatusDTO represents proficiency metrics for a specific domain
type DomainStatusDTO struct {
	DomainID         string  `json:"domain_id" example:"domain1"`
	DomainName       string  `json:"domain_name" example:"General Security Concepts"`
	ProficiencyScore float64 `json:"proficiency_score" example:"75.5"`
	Threshold        float64 `json:"threshold" example:"70.0"`
	Passed           bool    `json:"passed" example:"true"`
}

// SimulationStatusResponse describes the learner's overall scenario loop status
type SimulationStatusResponse struct {
	NeedsTraining      bool              `json:"needs_training" example:"true"`
	Domains            []DomainStatusDTO `json:"domains,omitempty"`
	CompletedScenarios int               `json:"completed_scenarios" example:"2"`
	MaxScenarios       int               `json:"max_scenarios" example:"10"`
	AllDomainsPassed   bool              `json:"all_domains_passed" example:"false"`
	Error              string            `json:"error,omitempty" example:"No skill profile found. Please complete the pre-test first."`
}

// ScenarioResponse represents the sanitized scenario presented to the learner
type ScenarioResponse struct {
	ScenarioID        string `json:"scenario_id" example:"sc-chiller-dos-01"`
	Title             string `json:"title" example:"Industrial Chiller DoS Attack"`
	Description       string `json:"description" example:"Anomalous telemetry detected in OT cooling network."`
	DomainID          string `json:"domain_id" example:"domain4"`
	Difficulty        int    `json:"difficulty" example:"2"`
	InitialLogs       any    `json:"initial_logs"`
	// PlaybookSteps is retained for frontend JSON contract compatibility
	PlaybookSteps     any    `json:"playbook_steps"`
	// ResponseActions is the canonical domain model representation
	ResponseActions   any    `json:"response_actions,omitempty"`
	SanitizedFindings any    `json:"sanitized_findings"`
}

// ActiveSimulationResponse contains the active session and corresponding scenario
type ActiveSimulationResponse struct {
	Session  any              `json:"session"`
	Scenario ScenarioResponse `json:"scenario"`
	Actions  any              `json:"actions,omitempty"`
}

// LogQueryRequest defines the query string executed by the learner against telemetry logs
type LogQueryRequest struct {
	SessionID string `json:"session_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Query     string `json:"query" binding:"required" example:"DeviceName == 'Chiller-01'"`
}

// LogQueryResponse confirms the recorded query action
type LogQueryResponse struct {
	Recorded bool `json:"recorded" example:"true"`
	ActionID uint `json:"action_id" example:"42"`
}

// SubmitScenarioRequest defines the triage decision and actions chosen by the learner
type SubmitScenarioRequest struct {
	SessionID          string   `json:"session_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	IsTruePositive     bool     `json:"is_true_positive" example:"true"`
	TPFPSelected       bool     `json:"tp_fp_selected" binding:"required" example:"true"`
	SelectedActions    []string `json:"selected_actions" binding:"required" example:"[\"act-contain-01\"]"`
	DiscoveredFindings []string `json:"discovered_findings" example:"[\"find-auth-fail\"]"`
}

// SubmitScenarioResponse represents the evaluation summary returned after finishing a scenario
type SubmitScenarioResponse struct {
	Result           any    `json:"result"`
	CompletionReason string `json:"completion_reason" example:"learner_completed"`
	Eligible         bool   `json:"eligible" example:"false"`
}

// SimulationResultResponse contains the detailed retrospective review for a completed scenario
type SimulationResultResponse struct {
	Session             any               `json:"session"`
	Scenario            any               `json:"scenario"`
	Actions             any               `json:"actions"`
	Result              any               `json:"result"`
	DomainProficiencies []DomainStatusDTO `json:"domain_proficiencies"`
}

// ResearchSummaryResponse provides experimental analytics across pre-test, scenarios, and post-test
type ResearchSummaryResponse struct {
	ProtocolVersion    string            `json:"protocol_version" example:"protocol-v1"`
	CalculationVersion string            `json:"calculation_version" example:"v1.0-weighted-0.4-0.6"`
	RubricVersion      string            `json:"rubric_version" example:"v1.0-20-40-40"`
	ParticipantID      any               `json:"participant_id"`
	HasPreTest         bool              `json:"has_pre_test" example:"true"`
	PreTestScore       *float64          `json:"pre_test_score" example:"65.0"`
	HasPostTest        bool              `json:"has_post_test" example:"true"`
	PostTestScore      *float64          `json:"post_test_score" example:"85.0"`
	ImprovementScore   *float64          `json:"improvement_score" example:"20.0"`
	AllDomainsPassed   bool              `json:"all_domains_passed" example:"true"`
	EligibleForPost    bool              `json:"eligible_for_post" example:"true"`
	CompletedScenarios int               `json:"completed_scenarios" example:"3"`
	MaxScenarios       int               `json:"max_scenarios" example:"10"`
	TotalActions       int               `json:"total_actions" example:"18"`
	TotalElapsedSec    int               `json:"total_elapsed_sec" example:"720"`
	AverageProficiency float64           `json:"average_proficiency" example:"78.5"`
	DomainProfiles     []DomainStatusDTO `json:"domain_profiles,omitempty"`
}
