package dto

import "time"

// TheorySubtopicSummaryDTO provides lightweight tree navigation metadata for one Sub-topic.
type TheorySubtopicSummaryDTO struct {
	SubTopicID          string     `json:"sub_topic_id" example:"1.1"`
	DomainID            string     `json:"domain_id" example:"domain1"`
	OrderIndex          int        `json:"order_index" example:"1"`
	Title               string     `json:"title" example:"Compare and contrast various types of security controls"`
	IsCompleted         bool       `json:"is_completed" example:"true"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	SampleAttemptedCount int       `json:"sample_attempted_count" example:"2"`
	SampleTotalCount     int       `json:"sample_total_count" example:"2"`
}

// TheoryDomainDTO represents one CompTIA Security+ domain Theory Module along
// with the learner's proficiency context and completion state.
type TheoryDomainDTO struct {
	DomainID                   string                     `json:"domain_id" example:"domain1"`
	DomainName                 string                     `json:"domain_name" example:"General Security Concepts"`
	Title                      string                     `json:"title" example:"Domain 1: General Security Concepts"`
	Content                    string                     `json:"content"`
	SourceURL                  string                     `json:"source_url" example:"https://csrc.nist.gov/pubs/sp/800/53/r5/final"`
	ProficiencyScore           float64                    `json:"proficiency_score" example:"71.4"`
	Threshold                  float64                    `json:"threshold" example:"70.0"`
	PassedThreshold            bool                       `json:"passed_threshold" example:"true"`
	RecommendedFocus           bool                       `json:"recommended_focus" example:"false"`
	IsCompleted                bool                       `json:"is_completed" example:"true"`
	CompletedAt                *time.Time                 `json:"completed_at,omitempty"`
	CompletedSubtopicsInDomain int                        `json:"completed_subtopics_in_domain" example:"4"`
	TotalSubtopicsInDomain     int                        `json:"total_subtopics_in_domain" example:"4"`
	Subtopics                  []TheorySubtopicSummaryDTO `json:"subtopics"`
}

// LearnDomainsResponse returns all five ordered Theory Modules and completion summary.
type LearnDomainsResponse struct {
	Unlocked               bool              `json:"unlocked" example:"true"`
	TheoryCompleted        bool              `json:"theory_completed" example:"false"`
	CompletedCount         int               `json:"completed_count" example:"2"`
	TotalDomains           int               `json:"total_domains" example:"5"`
	CompletedSubtopicCount int               `json:"completed_subtopic_count" example:"7"`
	TotalSubtopics         int               `json:"total_subtopics" example:"28"`
	CompletedDomains       []string          `json:"completed_domains" example:"domain1,domain2"`
	CompletedSubtopics     []string          `json:"completed_subtopics" example:"1.1,1.2"`
	Domains                []TheoryDomainDTO `json:"domains"`
}

// CompleteTheoryDomainResponse confirms idempotent completion of a domain Theory Module.
type CompleteTheoryDomainResponse struct {
	DomainID         string   `json:"domain_id" example:"domain1"`
	IsCompleted      bool     `json:"is_completed" example:"true"`
	CompletedDomains []string `json:"completed_domains" example:"domain1"`
	CompletedCount   int      `json:"completed_count" example:"1"`
	TotalDomains     int      `json:"total_domains" example:"5"`
	TheoryCompleted  bool     `json:"theory_completed" example:"false"`
	PreviewMode      bool     `json:"preview_mode" example:"false"`
}

// TheoryGateErrorResponse describes a blocked Theory Gate response.
type TheoryGateErrorResponse struct {
	Error                  string         `json:"error" example:"Please complete the Pre-Test first to unlock Learning Theory."`
	Code                   string         `json:"code,omitempty" example:"LEARN_PRETEST_REQUIRED"`
	Params                 map[string]any `json:"params,omitempty"`
	Reason                 string         `json:"reason" example:"pre_test_required"`
	Unlocked               *bool          `json:"unlocked,omitempty" example:"false"`
	CompletedCount         *int           `json:"completed_count,omitempty" example:"2"`
	TotalDomains           *int           `json:"total_domains,omitempty" example:"5"`
	CompletedDomains       []string       `json:"completed_domains,omitempty"`
	CompletedSubtopicCount *int           `json:"completed_subtopic_count,omitempty" example:"12"`
	TotalSubtopics         *int           `json:"total_subtopics,omitempty" example:"28"`
	CompletedSubtopics     []string       `json:"completed_subtopics,omitempty"`
	MissingSubtopics       []string       `json:"missing_subtopics,omitempty"`
}

// CompleteSubtopicResponse confirms attempt-gated Sub-topic completion and reports domain auto-completion.
type CompleteSubtopicResponse struct {
	SubTopicID                 string   `json:"sub_topic_id" example:"1.4"`
	DomainID                   string   `json:"domain_id" example:"domain1"`
	IsCompleted                bool     `json:"is_completed" example:"true"`
	DomainAutoCompleted        bool     `json:"domain_auto_completed" example:"true"`
	CompletedSubtopicsInDomain int      `json:"completed_subtopics_in_domain" example:"4"`
	TotalSubtopicsInDomain     int      `json:"total_subtopics_in_domain" example:"4"`
	CompletedSubtopicCount     int      `json:"completed_subtopic_count" example:"4"`
	TotalSubtopics             int      `json:"total_subtopics" example:"28"`
	CompletedSubtopics         []string `json:"completed_subtopics"`
	CompletedDomains           []string `json:"completed_domains"`
	CompletedCount             int      `json:"completed_count" example:"1"`
	TotalDomains               int      `json:"total_domains" example:"5"`
	TheoryCompleted            bool     `json:"theory_completed" example:"false"`
	PreviewMode                bool     `json:"preview_mode" example:"false"`
}

// LearnSubtopicSampleQuestionDTO represents one locked formative sample question for a Sub-topic.
// CorrectAnswer and Explanation remain omitted until SubmittedAnswer is non-null.
type LearnSubtopicSampleQuestionDTO struct {
	QuestionID      string     `json:"question_id"`
	OrderIndex      int        `json:"order_index" example:"1"`
	QuestionText    string     `json:"question_text"`
	Options         any        `json:"options"`
	Translations    any        `json:"translations,omitempty"`
	IsFallback      bool       `json:"is_fallback" example:"false"`
	SubmittedAnswer *string    `json:"submitted_answer,omitempty"`
	IsCorrect       *bool      `json:"is_correct,omitempty"`
	AnsweredAt      *time.Time `json:"answered_at,omitempty"`
	CorrectAnswer   *string    `json:"correct_answer,omitempty"`
	Explanation     *string    `json:"explanation,omitempty"`
}

// LearnSubtopicDetailResponse delivers one SY0-701 Sub-topic lesson and its 2 formative sample questions.
type LearnSubtopicDetailResponse struct {
	SubTopicID         string                           `json:"sub_topic_id" example:"1.1"`
	DomainID           string                           `json:"domain_id" example:"domain1"`
	OrderIndex         int                              `json:"order_index" example:"1"`
	Title              string                           `json:"title"`
	Content            string                           `json:"content"`
	SourceURL          string                           `json:"source_url"`
	IsCompleted        bool                             `json:"is_completed"`
	CompletedAt        *time.Time                       `json:"completed_at,omitempty"`
	PreviewMode        bool                             `json:"preview_mode"`
	PreviousSubTopicID string                           `json:"previous_sub_topic_id,omitempty"`
	NextSubTopicID     string                           `json:"next_sub_topic_id,omitempty"`
	AttemptedCount     int                              `json:"attempted_count"`
	RequiredCount      int                              `json:"required_count"`
	ReadyToComplete    bool                             `json:"ready_to_complete"`
	SampleQuestions    []LearnSubtopicSampleQuestionDTO `json:"sample_questions"`
}

// AttemptSampleQuestionRequest validates a Learner's answer choice for a Sub-topic sample question.
type AttemptSampleQuestionRequest struct {
	Answer string `json:"answer" binding:"required" example:"A"`
}

// AttemptSampleQuestionResponse returns immediate graded feedback with first-attempt locking.
type AttemptSampleQuestionResponse struct {
	SubTopicID       string `json:"sub_topic_id" example:"1.2"`
	QuestionID       string `json:"question_id"`
	SubmittedAnswer  string `json:"submitted_answer" example:"B"`
	IsCorrect        bool   `json:"is_correct" example:"true"`
	CorrectAnswer    string `json:"correct_answer" example:"B"`
	Explanation      string `json:"explanation"`
	Translations     any    `json:"translations,omitempty"`
	AttemptedCount   int    `json:"attempted_count" example:"2"`
	RequiredCount    int    `json:"required_count" example:"2"`
	ReadyToComplete  bool   `json:"ready_to_complete" example:"true"`
	AlreadySubmitted bool   `json:"already_submitted" example:"false"`
	PreviewMode      bool   `json:"preview_mode" example:"false"`
}
