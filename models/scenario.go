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
	SessionID   string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"session_id"`
	UserID      string     `gorm:"type:uuid;not null" json:"user_id"`
	ScenarioID  string     `gorm:"type:uuid;not null" json:"scenario_id"`
	Status      string     `gorm:"type:session_status;default:'in_progress'" json:"status"`
	FinalScore  *float64   `json:"final_score"`
	SkillGap    any        `gorm:"type:jsonb;serializer:json" json:"skill_gap"`
	StartedAt   time.Time  `gorm:"default:now()" json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	TotalActions int       `gorm:"default:0" json:"total_actions"`
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
