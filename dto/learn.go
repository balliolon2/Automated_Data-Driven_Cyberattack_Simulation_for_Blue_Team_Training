package dto

import "time"

// TheoryDomainDTO represents one CompTIA Security+ domain Theory Module along
// with the learner's proficiency context and completion state.
type TheoryDomainDTO struct {
	DomainID         string     `json:"domain_id" example:"domain1"`
	DomainName       string     `json:"domain_name" example:"General Security Concepts"`
	Title            string     `json:"title" example:"Domain 1: General Security Concepts"`
	Content          string     `json:"content"`
	SourceURL        string     `json:"source_url" example:"docs/REFERENCES.md#14-learning-theory-curriculum-provenance-r2r5"`
	ProficiencyScore float64    `json:"proficiency_score" example:"71.4"`
	Threshold        float64    `json:"threshold" example:"70.0"`
	PassedThreshold  bool       `json:"passed_threshold" example:"true"`
	RecommendedFocus bool       `json:"recommended_focus" example:"false"`
	IsCompleted      bool       `json:"is_completed" example:"true"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}

// LearnDomainsResponse returns all five ordered Theory Modules and completion summary.
type LearnDomainsResponse struct {
	Unlocked         bool              `json:"unlocked" example:"true"`
	TheoryCompleted  bool              `json:"theory_completed" example:"false"`
	CompletedCount   int               `json:"completed_count" example:"2"`
	TotalDomains     int               `json:"total_domains" example:"5"`
	CompletedDomains []string          `json:"completed_domains" example:"domain1,domain2"`
	Domains          []TheoryDomainDTO `json:"domains"`
}

// CompleteTheoryDomainResponse confirms idempotent completion of a domain Theory Module.
type CompleteTheoryDomainResponse struct {
	DomainID         string   `json:"domain_id" example:"domain1"`
	IsCompleted      bool     `json:"is_completed" example:"true"`
	CompletedDomains []string `json:"completed_domains" example:"domain1"`
	CompletedCount   int      `json:"completed_count" example:"1"`
	TotalDomains     int      `json:"total_domains" example:"5"`
	TheoryCompleted  bool     `json:"theory_completed" example:"false"`
}

// TheoryGateErrorResponse describes a blocked Theory Gate response.
type TheoryGateErrorResponse struct {
	Error            string   `json:"error" example:"Please complete the Pre-Test first to unlock Learning Theory."`
	Reason           string   `json:"reason" example:"pre_test_required"`
	Unlocked         *bool    `json:"unlocked,omitempty" example:"false"`
	CompletedCount   *int     `json:"completed_count,omitempty" example:"2"`
	TotalDomains     *int     `json:"total_domains,omitempty" example:"5"`
	CompletedDomains []string `json:"completed_domains,omitempty"`
}
