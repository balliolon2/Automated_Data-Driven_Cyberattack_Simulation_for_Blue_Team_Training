package models

import "time"

// EvaluationMetric persists objective Pre/Post and subjective CSUQ/NASA-TLX
// metrics in the pre-existing evaluation_metrics table (Zero-DDL).
type EvaluationMetric struct {
	EvalID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"eval_id"`
	UserID          string    `gorm:"type:uuid;not null" json:"user_id"`
	User            *User     `gorm:"foreignKey:UserID;references:UserID;constraint:-" json:"user,omitempty"`
	PreScore        float64   `gorm:"not null" json:"pre_score"`
	PostScore       float64   `gorm:"not null" json:"post_score"`
	ImprovementPct  float64   `json:"improvement_pct"`
	CSUQScores      any       `gorm:"column:csuq_scores;type:jsonb;serializer:json" json:"csuq_scores"`
	NASATLXScores   any       `gorm:"column:nasa_tlx_scores;type:jsonb;serializer:json" json:"nasa_tlx_scores"`
	SkillGapSummary any       `gorm:"column:skill_gap_summary;type:jsonb;serializer:json" json:"skill_gap_summary"`
	EvalDate        time.Time `gorm:"default:now()" json:"eval_date"`
	Notes           string    `gorm:"type:text" json:"notes"`
}

// TableName explicitly pins GORM to init.sql's evaluation_metrics table.
func (EvaluationMetric) TableName() string {
	return "evaluation_metrics"
}
