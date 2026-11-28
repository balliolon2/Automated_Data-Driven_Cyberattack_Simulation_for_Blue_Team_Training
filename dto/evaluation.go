package dto

import "time"

// NASATLXInputDTO captures the learner's 6 Raw NASA-TLX dimension ratings (0-100, step 5).
// Pointer fields ensure explicit 0 ratings are distinguished from omitted fields.
type NASATLXInputDTO struct {
	MentalDemand      *int `json:"mental_demand" example:"60"`
	PhysicalDemand    *int `json:"physical_demand" example:"10"`
	TemporalDemand    *int `json:"temporal_demand" example:"45"`
	PerformanceRating *int `json:"performance_rating" example:"85"`
	Effort            *int `json:"effort" example:"55"`
	Frustration       *int `json:"frustration" example:"20"`
}

// QualitativeFeedbackDTO holds the three optional qualitative impression prompts.
type QualitativeFeedbackDTO struct {
	MostImpressive     string `json:"most_impressive" example:"The SIEM log investigation felt realistic."`
	AreasToImprove     string `json:"areas_to_improve" example:"Add more cloud IAM scenarios."`
	AdditionalComments string `json:"additional_comments" example:"Bilingual hints helped me understand domain concepts."`
}

// SubmitQuestionnaireRequest is the learner payload sent to POST /api/evaluation/questionnaire.
type SubmitQuestionnaireRequest struct {
	Locale              string                 `json:"locale" example:"en"`
	CSUQ                map[string]*int        `json:"csuq" binding:"required"`
	NASATLX             NASATLXInputDTO        `json:"nasa_tlx"`
	QualitativeFeedback QualitativeFeedbackDTO `json:"qualitative_feedback"`
}

// CSUQSubscalesDTO captures computed CSUQ v3 subscale averages (nil if all items in a subscale were N/A).
type CSUQSubscalesDTO struct {
	SysUse    *float64 `json:"sysuse" example:"6.33"`
	InfoQual  *float64 `json:"infoqual" example:"6.20"`
	InterQual *float64 `json:"interqual" example:"6.33"`
	Overall   *float64 `json:"overall" example:"6.33"`
}

// CSUQValidCountsDTO captures non-N/A item counts per CSUQ subscale.
type CSUQValidCountsDTO struct {
	SysUse    int `json:"sysuse" example:"6"`
	InfoQual  int `json:"infoqual" example:"5"`
	InterQual int `json:"interqual" example:"3"`
	Overall   int `json:"overall" example:"15"`
}

// CSUQScoresPayloadDTO is the canonical JSONB structure persisted in evaluation_metrics.csuq_scores.
type CSUQScoresPayloadDTO struct {
	SchemaVersion string             `json:"schema_version" example:"csuq-v3-asc-7"`
	ScaleMin      int                `json:"scale_min" example:"1"`
	ScaleMax      int                `json:"scale_max" example:"7"`
	Direction     string             `json:"direction" example:"higher_is_better"`
	Locale        string             `json:"locale" example:"en"`
	Items         map[string]*int    `json:"items"`
	Subscales     CSUQSubscalesDTO   `json:"subscales"`
	ValidCounts   CSUQValidCountsDTO `json:"valid_counts"`
}

// NASATLXDimensionsDTO records both intuitive performance_rating and workload-normalized performance_workload.
type NASATLXDimensionsDTO struct {
	MentalDemand        int `json:"mental_demand" example:"60"`
	PhysicalDemand      int `json:"physical_demand" example:"10"`
	TemporalDemand      int `json:"temporal_demand" example:"45"`
	PerformanceRating   int `json:"performance_rating" example:"85"`
	PerformanceWorkload int `json:"performance_workload" example:"15"`
	Effort              int `json:"effort" example:"55"`
	Frustration         int `json:"frustration" example:"20"`
}

// NASATLXScoresPayloadDTO is the canonical JSONB structure persisted in evaluation_metrics.nasa_tlx_scores.
type NASATLXScoresPayloadDTO struct {
	SchemaVersion   string               `json:"schema_version" example:"nasa-tlx-rtlx-v1"`
	Variant         string               `json:"variant" example:"raw_tlx"`
	ScaleMin        int                  `json:"scale_min" example:"0"`
	ScaleMax        int                  `json:"scale_max" example:"100"`
	Step            int                  `json:"step" example:"5"`
	Dimensions      NASATLXDimensionsDTO `json:"dimensions"`
	OverallWorkload float64              `json:"overall_workload" example:"34.17"`
}

// EvaluationSkillGapSnapshotDTO stores the server-snapshotted training context in skill_gap_summary.
type EvaluationSkillGapSnapshotDTO struct {
	SnapshotVersion     string                 `json:"snapshot_version" example:"eval-snapshot-v1"`
	SubmittedLocale     string                 `json:"submitted_locale" example:"en"`
	PreSessionID        *string                `json:"pre_session_id,omitempty"`
	PostSessionID       string                 `json:"post_session_id" example:"987e6543-e21b-12d3-a456-426614174000"`
	CompletedScenarios  int                    `json:"completed_scenarios" example:"4"`
	AverageProficiency  float64                `json:"average_proficiency" example:"78.4"`
	Domains             []DomainStatusDTO      `json:"domains"`
	QualitativeFeedback QualitativeFeedbackDTO `json:"qualitative_feedback"`
}

// EvaluationMetricSummaryDTO is the canonical read-only evaluation summary returned to learners.
type EvaluationMetricSummaryDTO struct {
	EvalID              string                 `json:"eval_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	PreScore            float64                `json:"pre_score" example:"60.0"`
	PostScore           float64                `json:"post_score" example:"83.33"`
	ImprovementPct      float64                `json:"improvement_pct" example:"23.33"`
	SubmittedLocale     string                 `json:"submitted_locale" example:"en"`
	CompletedScenarios  int                    `json:"completed_scenarios" example:"4"`
	AverageProficiency  float64                `json:"average_proficiency" example:"78.4"`
	CSUQScores          CSUQScoresPayloadDTO   `json:"csuq_scores"`
	NASATLXScores       NASATLXScoresPayloadDTO `json:"nasa_tlx_scores"`
	QualitativeFeedback QualitativeFeedbackDTO `json:"qualitative_feedback"`
	Notes               string                 `json:"notes"`
	EvalDate            time.Time              `json:"eval_date"`
}

// QuestionnaireStatusResponse reports caller eligibility, preview status, and any existing submission.
type QuestionnaireStatusResponse struct {
	Role                 string                      `json:"role" example:"learner"`
	IsPreviewOnly        bool                        `json:"is_preview_only" example:"false"`
	HasCompletedPostTest bool                        `json:"has_completed_post_test" example:"true"`
	LatestPostSessionID  *string                     `json:"latest_post_session_id" example:"987e6543-e21b-12d3-a456-426614174000"`
	Eligible             bool                        `json:"eligible" example:"true"`
	IneligibleReason     string                      `json:"ineligible_reason" example:""`
	Submitted            bool                        `json:"submitted" example:"false"`
	Submission           *EvaluationMetricSummaryDTO `json:"submission"`
}

// SubmitQuestionnaireResponse wraps the created evaluation record.
type SubmitQuestionnaireResponse struct {
	Message    string                     `json:"message" example:"Post-test evaluation questionnaire submitted successfully."`
	Submission EvaluationMetricSummaryDTO `json:"submission"`
}

// EvaluationCohortFunnelDTO summarizes learner progression and questionnaire response rate.
type EvaluationCohortFunnelDTO struct {
	TotalLearners               int     `json:"total_learners" example:"30"`
	CompletedPreTestLearners    int     `json:"completed_pre_test_learners" example:"28"`
	CompletedPostTestLearners   int     `json:"completed_post_test_learners" example:"24"`
	QuestionnaireRespondents    int     `json:"questionnaire_respondents" example:"20"`
	ResponseRatePct             float64 `json:"response_rate_pct" example:"83.33"`
}

// AdminMeanSDDTO holds count, arithmetic mean, and sample SD (nil when n < 2).
type AdminMeanSDDTO struct {
	Key  string   `json:"key" example:"overall"`
	N    int      `json:"n" example:"20"`
	Mean *float64 `json:"mean" example:"6.25"`
	SD   *float64 `json:"sd" example:"0.54"`
}

// AdminLearningGainSummaryDTO summarizes Pre/Post/Gain across respondents.
type AdminLearningGainSummaryDTO struct {
	N                   int      `json:"n" example:"20"`
	PreMean             *float64 `json:"pre_mean" example:"61.5"`
	PostMean            *float64 `json:"post_mean" example:"83.2"`
	ImprovementMean     *float64 `json:"improvement_mean" example:"21.7"`
	ImprovementSD       *float64 `json:"improvement_sd" example:"9.4"`
	QualitativeCount    int      `json:"qualitative_count" example:"14"`
}

// AdminEvaluationsSummaryDTO bundles aggregate CSUQ, NASA-TLX, and learning-gain KPIs.
type AdminEvaluationsSummaryDTO struct {
	CSUQSubscales     []AdminMeanSDDTO            `json:"csuq_subscales"`
	NASATLXDimensions []AdminMeanSDDTO            `json:"nasa_tlx_dimensions"`
	LearningGain      AdminLearningGainSummaryDTO `json:"learning_gain"`
}

// AdminEvaluationRowDTO represents a pseudonymized learner response row (never exposes email).
type AdminEvaluationRowDTO struct {
	EvalID              string                  `json:"eval_id"`
	ParticipantCode     string                  `json:"participant_code" example:"P-123E4567"`
	Nickname            string                  `json:"nickname" example:"shadow_analyst"`
	SubmittedLocale     string                  `json:"submitted_locale" example:"en"`
	EvalDate            time.Time               `json:"eval_date"`
	PreScore            float64                 `json:"pre_score"`
	PostScore           float64                 `json:"post_score"`
	ImprovementPct      float64                 `json:"improvement_pct"`
	CompletedScenarios  int                     `json:"completed_scenarios"`
	AverageProficiency  float64                 `json:"average_proficiency"`
	CSUQScores          CSUQScoresPayloadDTO    `json:"csuq_scores"`
	NASATLXScores       NASATLXScoresPayloadDTO `json:"nasa_tlx_scores"`
	QualitativeFeedback QualitativeFeedbackDTO  `json:"qualitative_feedback"`
	Notes               string                  `json:"notes"`
}

// AdminEvaluationsResponse powers the Admin operational Post-Test Evaluations Dashboard.
type AdminEvaluationsResponse struct {
	Cohort    EvaluationCohortFunnelDTO  `json:"cohort"`
	Summary   AdminEvaluationsSummaryDTO `json:"summary"`
	Responses []AdminEvaluationRowDTO    `json:"responses"`
}

// AdminEvaluationsExportJSONResponse structures the downloadable JSON dataset export.
type AdminEvaluationsExportJSONResponse struct {
	ExportedAt time.Time                  `json:"exported_at"`
	Cohort     EvaluationCohortFunnelDTO  `json:"cohort"`
	Summary    AdminEvaluationsSummaryDTO `json:"summary"`
	Responses  []AdminEvaluationRowDTO    `json:"responses"`
}

// DescriptiveStatDTO captures publication-grade descriptive statistics and 95% Student-t CI.
type DescriptiveStatDTO struct {
	Key    string   `json:"key" example:"overall"`
	Group  string   `json:"group,omitempty" example:"SYSUSE"`
	N      int      `json:"n" example:"20"`
	Mean   *float64 `json:"mean" example:"6.12"`
	SD     *float64 `json:"sd" example:"0.54"`
	SE     *float64 `json:"se" example:"0.12"`
	CILow  *float64 `json:"ci_low" example:"5.87"`
	CIHigh *float64 `json:"ci_high" example:"6.37"`
	Median *float64 `json:"median" example:"6.19"`
	Q1     *float64 `json:"q1" example:"5.75"`
	Q3     *float64 `json:"q3" example:"6.50"`
	IQR    *float64 `json:"iqr" example:"0.75"`
	Min    *float64 `json:"min" example:"5.00"`
	Max    *float64 `json:"max" example:"6.94"`
	Band   string   `json:"band" example:"excellent"`
	Status string   `json:"status" example:"ok"`
}

// CronbachAlphaDTO reports internal consistency reliability over listwise complete cases.
type CronbachAlphaDTO struct {
	Key             string   `json:"key" example:"csuq_overall"`
	ItemCount       int      `json:"item_count" example:"16"`
	CompleteCasesN  int      `json:"complete_cases_n" example:"18"`
	ExcludedNACases int      `json:"excluded_na_cases" example:"2"`
	Alpha           *float64 `json:"alpha" example:"0.91"`
	Band            string   `json:"band" example:"excellent"`
	Status          string   `json:"status" example:"ok"`
}

// PairedTTestDTO holds parametric Pre-to-Post paired t-test and Cohen's d_z.
type PairedTTestDTO struct {
	N          int      `json:"n" example:"20"`
	DF         int      `json:"df" example:"19"`
	MeanDiff   *float64 `json:"mean_diff" example:"21.50"`
	SDDiff     *float64 `json:"sd_diff" example:"9.10"`
	CILow      *float64 `json:"ci_low" example:"17.24"`
	CIHigh     *float64 `json:"ci_high" example:"25.76"`
	TStat      *float64 `json:"t_stat" example:"10.57"`
	PValue     *float64 `json:"p_value" example:"0.0001"`
	CohensDz   *float64 `json:"cohens_dz" example:"2.36"`
	EffectBand string   `json:"effect_band" example:"large"`
	Status     string   `json:"status" example:"ok"`
}

// WilcoxonTestDTO holds non-parametric Wilcoxon signed-rank test and rank-biserial r_rb.
type WilcoxonTestDTO struct {
	N             int      `json:"n" example:"20"`
	EffectiveN    int      `json:"effective_n" example:"19"`
	ZeroDiffCount int      `json:"zero_diff_count" example:"1"`
	Method        string   `json:"method" example:"exact"`
	WStat         *float64 `json:"w_stat" example:"0.0"`
	WPlus         *float64 `json:"w_plus" example:"190.0"`
	WMinus        *float64 `json:"w_minus" example:"0.0"`
	PValue        *float64 `json:"p_value" example:"0.0001"`
	RankBiserial  *float64 `json:"rank_biserial" example:"1.0"`
	Status        string   `json:"status" example:"ok"`
}

// LearningGainStatisticsDTO bundles Pre, Post, Difference descriptives and both inferential tests.
type LearningGainStatisticsDTO struct {
	Pre         DescriptiveStatDTO `json:"pre"`
	Post        DescriptiveStatDTO `json:"post"`
	Difference  DescriptiveStatDTO `json:"difference"`
	PairedTTest PairedTTestDTO     `json:"paired_t_test"`
	Wilcoxon    WilcoxonTestDTO    `json:"wilcoxon"`
}

// SpearmanCorrelationDTO reports non-parametric rank correlation between two constructs.
type SpearmanCorrelationDTO struct {
	Key    string   `json:"key" example:"learning_gain_vs_csuq_overall"`
	N      int      `json:"n" example:"20"`
	Rho    *float64 `json:"rho" example:"0.54"`
	PValue *float64 `json:"p_value" example:"0.014"`
	Band   string   `json:"band" example:"strong"`
	Status string   `json:"status" example:"ok"`
}

// ParticipantPlotPointDTO holds a pseudonymized respondent's plotted metrics for statistical figures (zero PII).
type ParticipantPlotPointDTO struct {
	ParticipantCode string   `json:"participant_code" example:"P-A1B2C3D4"`
	PreScore        float64  `json:"pre_score" example:"46.7"`
	PostScore       float64  `json:"post_score" example:"76.7"`
	ImprovementPct  float64  `json:"improvement_pct" example:"30.0"`
	CSUQOverall     *float64 `json:"csuq_overall,omitempty" example:"6.25"`
	NASATLXOverall  *float64 `json:"nasa_tlx_overall,omitempty" example:"35.0"`
}

// AdminPaperStatisticsResponse powers the Admin-only Paper-Ready Statistics workbench.
type AdminPaperStatisticsResponse struct {
	GeneratedAt          time.Time                   `json:"generated_at"`
	ProtocolVersion      string                      `json:"protocol_version" example:"protocol-v1"`
	CSUQSchemaVersion    string                      `json:"csuq_schema_version" example:"csuq-v3-asc-7"`
	NASATLXSchemaVersion string                      `json:"nasa_tlx_schema_version" example:"nasa-tlx-rtlx-v1"`
	IsPilotSample        bool                        `json:"is_pilot_sample" example:"true"`
	Cohort               EvaluationCohortFunnelDTO   `json:"cohort"`
	CSUQSubscales        []DescriptiveStatDTO        `json:"csuq_subscales"`
	CSUQItems            []DescriptiveStatDTO        `json:"csuq_items"`
	NASATLX              []DescriptiveStatDTO        `json:"nasa_tlx"`
	Reliability          []CronbachAlphaDTO          `json:"reliability"`
	LearningGain         LearningGainStatisticsDTO   `json:"learning_gain"`
	Correlations         []SpearmanCorrelationDTO    `json:"correlations"`
	ParticipantPoints    []ParticipantPlotPointDTO   `json:"participant_points"`
}


