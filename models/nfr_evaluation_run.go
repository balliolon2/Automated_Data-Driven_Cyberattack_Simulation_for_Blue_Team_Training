package models

import "time"

// NFREvaluationRun persists an immutable system-level NFR evaluation snapshot in PostgreSQL.
type NFREvaluationRun struct {
	RunID            string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"run_id"`
	TriggeredBy      *string   `gorm:"type:uuid" json:"triggered_by,omitempty"`
	TriggeredByUser  *User     `gorm:"foreignKey:TriggeredBy;references:UserID;constraint:-" json:"triggered_by_user,omitempty"`
	RunLabel         string    `gorm:"type:varchar(120);not null;default:'Live NFR Snapshot'" json:"run_label"`
	Environment      string    `gorm:"type:varchar(64);not null;default:'local'" json:"environment"`
	Source           string    `gorm:"type:varchar(32);not null" json:"source"`
	RubricVersion    string    `gorm:"type:varchar(32);not null;default:'nfr-rubric-v1'" json:"rubric_version"`
	OverallStatus    string    `gorm:"type:varchar(32);not null" json:"overall_status"`
	PassedCount      int       `gorm:"not null;default:0" json:"passed_count"`
	WarnCount        int       `gorm:"not null;default:0" json:"warn_count"`
	FailCount        int       `gorm:"not null;default:0" json:"fail_count"`
	NotMeasuredCount int       `gorm:"not null;default:0" json:"not_measured_count"`
	MeasuredPassRate float64   `gorm:"not null;default:0" json:"measured_pass_rate"`
	NonLLMP95Ms      *float64  `gorm:"column:non_llm_p95_ms" json:"non_llm_p95_ms,omitempty"`
	AvailabilityPct  *float64  `gorm:"column:availability_pct" json:"availability_pct,omitempty"`
	ReportPayload    any       `gorm:"column:report_payload;type:jsonb;serializer:json;not null" json:"report_payload"`
	Notes            string    `gorm:"type:text" json:"notes"`
	CreatedAt        time.Time `gorm:"default:now()" json:"created_at"`
}

func (NFREvaluationRun) TableName() string {
	return "nfr_evaluation_runs"
}
