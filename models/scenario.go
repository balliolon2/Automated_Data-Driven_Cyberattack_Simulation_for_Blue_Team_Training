package models

import (
	"time"
)

type Scenario struct {
	ScenarioID       string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"scenario_id"`
	Title            string    `gorm:"type:varchar;not null" json:"title"`
	Description      string    `gorm:"type:text" json:"description"`
	MitreTechnique   *string   `gorm:"type:varchar" json:"mitre_technique"`
	DomainID         string    `gorm:"type:varchar" json:"domain_id"`
	Difficulty       int       `gorm:"default:3" json:"difficulty"`
	InitialLogs      any       `gorm:"type:jsonb;serializer:json;not null" json:"initial_logs"`
	PlaybookSteps    any       `gorm:"type:jsonb;serializer:json" json:"playbook_steps"`
	ExpectedOutcomes any       `gorm:"type:jsonb;serializer:json" json:"expected_outcomes"`
	IsTruePositive   bool      `gorm:"not null;default:true" json:"is_true_positive"`
	TpFpExplanation  string    `gorm:"type:text" json:"tp_fp_explanation"`
	Status           string    `gorm:"type:scenario_status;default:'active'" json:"status"`
	CreatedAt        time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at"`
}

type SimulationSession struct {
	SessionID          string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"session_id"`
	UserID             string     `gorm:"type:uuid;not null" json:"user_id"`
	ScenarioID         string     `gorm:"type:uuid;not null" json:"scenario_id"`
	Status             string     `gorm:"type:session_status;default:'in_progress'" json:"status"`
	CompletionReason   string     `gorm:"type:varchar" json:"completion_reason"`
	FailureClass       string     `gorm:"type:varchar" json:"failure_class"`
	GenerationType     string     `gorm:"type:varchar;default:'static_fallback'" json:"generation_type"`
	FallbackReason     string     `gorm:"type:text" json:"fallback_reason"`
	SelectionReason    string     `gorm:"type:text" json:"selection_reason"`
	FinalScore         *float64   `json:"final_score"`
	SkillGap           any        `gorm:"type:jsonb;serializer:json" json:"skill_gap"`
	ThresholdValue     float64    `gorm:"default:70.0" json:"threshold_value"`
	MaxScenarios       int        `gorm:"default:10" json:"max_scenarios"`
	TimeLimitSeconds   int        `gorm:"default:7200" json:"time_limit_seconds"`
	ElapsedSeconds     int        `gorm:"default:0" json:"elapsed_seconds"`
	CalculationVersion string     `gorm:"type:varchar;default:'proficiency-v1'" json:"calculation_version"`
	RubricVersion      string     `gorm:"type:varchar;default:'rubric-v1'" json:"rubric_version"`
	BlueprintVersion   string     `gorm:"type:varchar;default:'blueprint-v1'" json:"blueprint_version"`
	StartedAt          time.Time  `gorm:"default:now()" json:"started_at"`
	CompletedAt        *time.Time `json:"completed_at"`
	TotalActions       int        `gorm:"default:0" json:"total_actions"`
}

type ScenarioSnapshot struct {
	SnapshotID       string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"snapshot_id"`
	SessionID        string    `gorm:"type:uuid;not null" json:"session_id"`
	ScenarioID       string    `gorm:"type:uuid;not null" json:"scenario_id"`
	ScenarioData     any       `gorm:"type:jsonb;serializer:json;not null" json:"scenario_data"`
	SourceType       string    `gorm:"type:varchar;default:'static_fallback'" json:"source_type"`
	Provider         string    `gorm:"type:varchar" json:"provider"`
	Model            string    `gorm:"type:varchar" json:"model"`
	PromptVersion    string    `gorm:"type:varchar;default:'prompt-v1'" json:"prompt_version"`
	BlueprintVersion string    `gorm:"type:varchar;default:'blueprint-v1'" json:"blueprint_version"`
	RubricVersion    string    `gorm:"type:varchar;default:'rubric-v1'" json:"rubric_version"`
	ValidationStatus string    `gorm:"type:varchar;default:'valid'" json:"validation_status"`
	FallbackReason   string    `gorm:"type:text" json:"fallback_reason"`
	SelectionReason  string    `gorm:"type:text" json:"selection_reason"`
	RenderedAt       time.Time `gorm:"default:now()" json:"rendered_at"`
}

type SessionAction struct {
	ActionID   int64     `gorm:"primaryKey;autoIncrement" json:"action_id"`
	SessionID  string    `gorm:"type:uuid;not null" json:"session_id"`
	StepOrder  int       `gorm:"not null" json:"step_order"`
	ActionType string    `gorm:"type:action_type;not null" json:"action_type"`
	Payload    any       `gorm:"type:jsonb;serializer:json" json:"payload"`
	IsCorrect  *bool     `json:"is_correct"`
	Points     int       `gorm:"default:0" json:"points"`
	Timestamp  time.Time `gorm:"default:now()" json:"timestamp"`
}

type UserSkillProfile struct {
	ProfileID          string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"profile_id"`
	UserID             string     `gorm:"type:uuid;not null" json:"user_id"`
	DomainID           string     `gorm:"type:varchar;not null" json:"domain_id"`
	ProficiencyScore   float64    `gorm:"default:0" json:"proficiency_score"`
	LastPracticed      *time.Time `json:"last_practiced"`
	ScenariosCompleted int        `gorm:"default:0" json:"scenarios_completed"`
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

type DomainScoreDetail struct {
	Earned     float64 `json:"earned"`
	MaxPoints  float64 `json:"max_points"`
	Percentage float64 `json:"percentage"`
}

type ScenarioResult struct {
	TotalScore      float64                      `json:"total_score"`
	TPFPCorrect     bool                         `json:"tp_fp_correct"`
	TPFPPoints      float64                      `json:"tp_fp_points"`
	FindingsPoints  float64                      `json:"findings_points"`
	ResponsePoints  float64                      `json:"response_points"`
	DomainScores    map[string]DomainScoreDetail `json:"domain_scores"`
	DomainScoreMap  map[string]float64           `json:"-"` // internal use for proficiency updates
	FindingsDetail  []FindingResult              `json:"findings_detail"`
	ResponseDetail  []ResponseActionResult       `json:"response_detail"`
	TPFPExplanation string                       `json:"tp_fp_explanation"`
}

type SubmitScenarioInput struct {
	SessionID          string   `json:"session_id" binding:"required"`
	IsTruePositive     bool     `json:"is_true_positive"`
	TPFPSelected       bool     `json:"tp_fp_selected" binding:"required"`
	SelectedActions    []string `json:"selected_actions" binding:"required"`
	DiscoveredFindings []string `json:"discovered_findings"`
}
