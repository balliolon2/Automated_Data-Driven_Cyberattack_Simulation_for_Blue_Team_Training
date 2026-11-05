package dto

import "time"

// SpecialistSubmissionSummary represents an anonymized learner attempt
type SpecialistSubmissionSummary struct {
	SessionID             string  `json:"session_id"`
	LearnerNickname       string  `json:"learner_nickname"`
	LearnerTier           int     `json:"learner_tier"`
	ScenarioID            string  `json:"scenario_id"`
	ScenarioTitle         string  `json:"scenario_title"`
	DomainID              string  `json:"domain_id"`
	DomainName            string  `json:"domain_name"`
	FinalScore            float64 `json:"final_score"`
	TPFPCorrect           bool    `json:"tp_fp_correct"`
	FindingsFoundCount    int     `json:"findings_found_count"`
	TotalFindingsCount    int     `json:"total_findings_count"`
	EvidenceSelectedCount int     `json:"evidence_selected_count"`
	NoiseSelectedCount    int     `json:"noise_selected_count"`
	EvidencePenalty       float64 `json:"evidence_penalty"`
	ActionsCount          int     `json:"actions_count"`
	CompletedAt           string  `json:"completed_at"`
}

// SpecialistSubmissionDetail provides an in-depth breakdown of an attempt
type SpecialistSubmissionDetail struct {
	SessionID               string                     `json:"session_id"`
	LearnerNickname         string                     `json:"learner_nickname"`
	LearnerTier             int                        `json:"learner_tier"`
	ScenarioID              string                     `json:"scenario_id"`
	ScenarioTitle           string                     `json:"scenario_title"`
	ScenarioDescription     string                     `json:"scenario_description"`
	DomainID                string                     `json:"domain_id"`
	IsTruePositive          bool                       `json:"is_true_positive"`
	TpFpExplanation         string                     `json:"tp_fp_explanation"`
	LearnerChoice           *bool                      `json:"learner_choice"`
	TPFPCorrect             bool                       `json:"tp_fp_correct"`
	TPFPPoints              float64                    `json:"tp_fp_points"`
	DiscoveredFindings      []string                   `json:"discovered_findings"`
	ExpectedFindings        any                        `json:"expected_findings"`
	SelectedActions         []string                   `json:"selected_actions"`
	ExpectedActions         any                        `json:"expected_actions"`
	FinalScore              float64                    `json:"final_score"`
	SelectedLogIndices      []int                      `json:"selected_log_indices"`
	EvidenceFindings        []EvidenceFindingDTO       `json:"evidence_findings"`
	NoiseSelections         []EvidenceLogDTO           `json:"noise_selections"`
	GrossFindingsPoints     float64                    `json:"gross_findings_points"`
	FindingsPoints          float64                    `json:"findings_points"`
	EvidencePenalty         float64                    `json:"evidence_penalty"`
	EvidenceReviewAvailable bool                       `json:"evidence_review_available"`
	Timeline                []SubmissionTimelineAction `json:"timeline"`
	CompletedAt             string                     `json:"completed_at"`
}

// EvidenceLogDTO is a Specialist-only telemetry snippet for one evidence row.
type EvidenceLogDTO struct {
	Index     int    `json:"index"`
	Timestamp string `json:"timestamp"`
	Source    string `json:"source"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
}

// EvidenceFindingDTO is the Specialist-only evidence review for one Key Finding.
type EvidenceFindingDTO struct {
	ID                 string           `json:"id"`
	Description        string           `json:"description"`
	DomainID           string           `json:"domain_id"`
	Found              bool             `json:"found"`
	Points             int              `json:"points"`
	Earned             int              `json:"earned"`
	EvidenceLogIndices []int            `json:"evidence_log_indices"`
	MatchedLogIndices  []int            `json:"matched_log_indices"`
	EvidenceLogs       []EvidenceLogDTO `json:"evidence_logs"`
	Explanation        string           `json:"explanation"`
}

// SubmissionTimelineAction represents an action taken during investigation
type SubmissionTimelineAction struct {
	ActionID   int64     `json:"action_id"`
	StepOrder  int       `json:"step_order"`
	ActionType string    `json:"action_type"`
	Payload    any       `json:"payload"`
	IsCorrect  *bool     `json:"is_correct"`
	Points     int       `json:"points"`
	Timestamp  time.Time `json:"timestamp"`
}

// CreateThreadRequest payload for authoring a new analysis thread
type CreateThreadRequest struct {
	Title      string   `json:"title" binding:"required,min=5,max=255"`
	Content    string   `json:"content" binding:"required,min=20"`
	ScenarioID *string  `json:"scenario_id,omitempty"`
	Tags       []string `json:"tags,omitempty"`
}

// UpdateThreadRequest payload for editing an existing thread
type UpdateThreadRequest struct {
	Title   string   `json:"title" binding:"required,min=5,max=255"`
	Content string   `json:"content" binding:"required,min=20"`
	Tags    []string `json:"tags,omitempty"`
}

// ThreadAuthorDTO public metadata of the thread author
type ThreadAuthorDTO struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}

// ThreadScenarioDTO summary of the linked scenario
type ThreadScenarioDTO struct {
	ScenarioID     string `json:"scenario_id"`
	Title          string `json:"title"`
	DomainID       string `json:"domain_id"`
	MitreTechnique string `json:"mitre_technique,omitempty"`
}

// ThreadResponse represents a serialized analysis thread
type ThreadResponse struct {
	ThreadID       string             `json:"thread_id"`
	Author         ThreadAuthorDTO    `json:"author"`
	Scenario       *ThreadScenarioDTO `json:"scenario,omitempty"`
	Title          string             `json:"title"`
	Content        string             `json:"content"`
	Tags           []string           `json:"tags"`
	UpvoteCount    int                `json:"upvote_count"`
	ViewCount      int                `json:"view_count"`
	IsPinned       bool               `json:"is_pinned"`
	IsLocked       bool               `json:"is_locked"`
	UserHasUpvoted bool               `json:"user_has_upvoted"`
	CreatedAt      string             `json:"created_at"`
	UpdatedAt      string             `json:"updated_at"`
}
