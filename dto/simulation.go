package dto

import "time"

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
	NeedsTraining          bool              `json:"needs_training" example:"true"`
	Domains                []DomainStatusDTO `json:"domains,omitempty"`
	CompletedScenarios     int               `json:"completed_scenarios" example:"2"`
	MaxScenarios           int               `json:"max_scenarios" example:"10"`
	AllDomainsPassed       bool              `json:"all_domains_passed" example:"false"`
	TheoryCompleted        bool              `json:"theory_completed" example:"true"`
	CompletedTheoryDomains []string          `json:"completed_theory_domains"`
	CompletedSubtopicCount int               `json:"completed_subtopic_count" example:"28"`
	TotalSubtopics         int               `json:"total_subtopics" example:"28"`
	CompletedSubtopics     []string          `json:"completed_subtopics"`
	Error                  string            `json:"error,omitempty" example:"No skill profile found. Please complete the pre-test first."`
}

// ScenarioResponse represents the sanitized scenario presented to the learner
type ScenarioResponse struct {
	ScenarioID  string `json:"scenario_id" example:"sc-chiller-dos-01"`
	Title       string `json:"title" example:"Industrial Chiller DoS Attack"`
	Description string `json:"description" example:"Anomalous telemetry detected in OT cooling network."`
	DomainID    string `json:"domain_id" example:"domain4"`
	Difficulty  int    `json:"difficulty" example:"2"`
	InitialLogs any    `json:"initial_logs"`
	// PlaybookSteps is retained for frontend JSON contract compatibility
	PlaybookSteps any `json:"playbook_steps"`
	// ResponseActions is the canonical domain model representation
	ResponseActions any `json:"response_actions,omitempty"`
	// TotalFindings is the only pre-submission finding disclosure. Evidence
	// indices, finding descriptions, and verdict markers stay server-side.
	TotalFindings int `json:"total_findings" example:"2"`
	Translations  any `json:"translations,omitempty"`
}

// ActiveSimulationResponse contains the active session and corresponding scenario
type ActiveSimulationResponse struct {
	Session            any              `json:"session"`
	Scenario           ScenarioResponse `json:"scenario"`
	Actions            any              `json:"actions,omitempty"`
	SelectedLogIndices []int            `json:"selected_log_indices" example:"0,3"`
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
	SelectedActions    []string `json:"selected_actions" binding:"required" example:"act-contain-01"`
	SelectedLogIndices []int    `json:"selected_log_indices" example:"0,3"`
}

// SelectLogRequest toggles one log in the learner's selected evidence set.
// LogIndex is a pointer so index zero is accepted as a present value.
type SelectLogRequest struct {
	SessionID string `json:"session_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	LogIndex  *int   `json:"log_index" binding:"required" example:"0"`
	Selected  *bool  `json:"selected" binding:"required" example:"true"`
}

// ClearLogSelectionRequest clears every selected log for a simulation session.
type ClearLogSelectionRequest struct {
	SessionID string `json:"session_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// EvidenceSelectionResponse is the server-replayed selected evidence set.
type EvidenceSelectionResponse struct {
	SelectedLogIndices []int `json:"selected_log_indices" example:"0,3"`
	SelectedCount      int   `json:"selected_count" example:"2"`
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
	CalculationVersion string            `json:"calculation_version" example:"proficiency-v1"`
	RubricVersion      string            `json:"rubric_version" example:"rubric-v2-evidence-click"`
	ParticipantID      any               `json:"participant_id"`
	HasPreTest         bool              `json:"has_pre_test" example:"true"`
	PreTestScore       *float64          `json:"pre_test_score" example:"65.0"`
	HasPostTest        bool              `json:"has_post_test" example:"true"`
	PostTestScore      *float64          `json:"post_test_score" example:"85.0"`
	ImprovementScore   *float64          `json:"improvement_score" example:"20.0"`
	AllDomainsPassed       bool              `json:"all_domains_passed" example:"true"`
	EligibleForPost        bool              `json:"eligible_for_post" example:"true"`
	TheoryCompleted        bool                       `json:"theory_completed" example:"true"`
	CompletedTheoryDomains []string                   `json:"completed_theory_domains"`
	TheoryCompletedAt      *time.Time                 `json:"theory_completed_at,omitempty"`
	CompletedScenarios     int                        `json:"completed_scenarios" example:"3"`
	MaxScenarios           int                        `json:"max_scenarios" example:"10"`
	TotalActions           int                        `json:"total_actions" example:"18"`
	TotalElapsedSec        int                        `json:"total_elapsed_sec" example:"720"`
	AverageProficiency     float64                    `json:"average_proficiency" example:"78.5"`
	DomainProfiles         []DomainStatusDTO          `json:"domain_profiles,omitempty"`
	LearnSubtopicTelemetry *LearnSubtopicTelemetryDTO `json:"learn_subtopic_telemetry,omitempty"`
	QuestionnaireSubmitted   bool                       `json:"questionnaire_submitted" example:"true"`
	QuestionnaireSubmittedAt *time.Time                 `json:"questionnaire_submitted_at,omitempty"`
	CSUQOverall              *float64                   `json:"csuq_overall,omitempty" example:"6.33"`
	NASATLXOverall           *float64                   `json:"nasa_tlx_overall,omitempty" example:"34.17"`
}

// LearnDomainBreakdownDTO summarizes Sub-topic completion and sample accuracy for one domain.
type LearnDomainBreakdownDTO struct {
	DomainID           string `json:"domain_id" example:"domain1"`
	CompletedSubtopics int    `json:"completed_subtopics" example:"4"`
	TotalSubtopics     int    `json:"total_subtopics" example:"4"`
	SampleAssigned     int    `json:"sample_assigned" example:"8"`
	SampleAttempted    int    `json:"sample_attempted" example:"8"`
	SampleCorrect      int    `json:"sample_correct" example:"6"`
}

// LearnSubtopicTelemetryDTO captures Learning Theory Sub-topic engagement and formative accuracy metrics.
type LearnSubtopicTelemetryDTO struct {
	CompletedSubtopicsCount  int                       `json:"completed_subtopics_count" example:"28"`
	TotalSubtopics           int                       `json:"total_subtopics" example:"28"`
	CompletedSubtopics       []string                  `json:"completed_subtopics"`
	SampleQuestionsAssigned  int                       `json:"sample_questions_assigned" example:"56"`
	SampleQuestionsAttempted int                       `json:"sample_questions_attempted" example:"56"`
	SampleQuestionsCorrect   int                       `json:"sample_questions_correct" example:"44"`
	SampleAccuracyPercent    float64                   `json:"sample_accuracy_percent" example:"78.57"`
	DomainBreakdown          []LearnDomainBreakdownDTO `json:"domain_breakdown"`
}
