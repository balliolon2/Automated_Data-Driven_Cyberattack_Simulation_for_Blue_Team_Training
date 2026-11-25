package services

import (
	"encoding/json"
	"errors"
	"math"
	"strings"
	"unicode/utf8"

	"cybersim/dto"
	"cybersim/models"
)

const (
	CSUQSchemaVersion         = "csuq-v3-asc-7"
	NASATLXSchemaVersion      = "nasa-tlx-rtlx-v1"
	EvalSnapshotVersion       = "eval-snapshot-v1"
	MaxQualitativeFeedbackRunes = 2000
)

var (
	ErrQuestionnaireInvalidCSUQ     = errors.New("invalid CSUQ payload")
	ErrQuestionnaireAllNA           = errors.New("all CSUQ items cannot be N/A")
	ErrQuestionnaireInvalidTLX      = errors.New("invalid NASA-TLX payload")
	ErrQuestionnaireFeedbackTooLong = errors.New("qualitative feedback exceeds 2000 characters")
)

// CSUQItemKeys defines the exact 16 items of CSUQ Version 3 in canonical order.
var CSUQItemKeys = []string{
	"q1", "q2", "q3", "q4", "q5", "q6",
	"q7", "q8", "q9", "q10", "q11", "q12",
	"q13", "q14", "q15", "q16",
}

var (
	csuqSysUseKeys    = []string{"q1", "q2", "q3", "q4", "q5", "q6"}
	csuqInfoQualKeys  = []string{"q7", "q8", "q9", "q10", "q11", "q12"}
	csuqInterQualKeys = []string{"q13", "q14", "q15"}
)

// NormalizeQuestionnaireLocale normalizes requested survey UI locale to "en" or "th".
func NormalizeQuestionnaireLocale(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "th") {
		return "th"
	}
	return "en"
}

func roundEval2(val float64) float64 {
	return math.Round(val*100) / 100
}

func computeCSUQSubset(items map[string]*int, keys []string) (*float64, int) {
	sum := 0
	count := 0
	for _, k := range keys {
		if v := items[k]; v != nil {
			sum += *v
			count++
		}
	}
	if count == 0 {
		return nil, 0
	}
	avg := roundEval2(float64(sum) / float64(count))
	return &avg, count
}

// ValidateAndScoreCSUQ validates the 16 CSUQ v3 responses (1..7 or nil for N/A),
// rejects all-N/A or malformed keys/ranges, and computes subscale averages.
func ValidateAndScoreCSUQ(raw map[string]*int, locale string) (dto.CSUQScoresPayloadDTO, error) {
	if len(raw) != len(CSUQItemKeys) {
		return dto.CSUQScoresPayloadDTO{}, ErrQuestionnaireInvalidCSUQ
	}

	normalizedItems := make(map[string]*int, len(CSUQItemKeys))
	for _, key := range CSUQItemKeys {
		val, exists := raw[key]
		if !exists {
			return dto.CSUQScoresPayloadDTO{}, ErrQuestionnaireInvalidCSUQ
		}
		if val == nil {
			normalizedItems[key] = nil
			continue
		}
		if *val < 1 || *val > 7 {
			return dto.CSUQScoresPayloadDTO{}, ErrQuestionnaireInvalidCSUQ
		}
		copied := *val
		normalizedItems[key] = &copied
	}

	overallMean, overallCount := computeCSUQSubset(normalizedItems, CSUQItemKeys)
	if overallCount == 0 {
		return dto.CSUQScoresPayloadDTO{}, ErrQuestionnaireAllNA
	}

	sysUseMean, sysUseCount := computeCSUQSubset(normalizedItems, csuqSysUseKeys)
	infoQualMean, infoQualCount := computeCSUQSubset(normalizedItems, csuqInfoQualKeys)
	interQualMean, interQualCount := computeCSUQSubset(normalizedItems, csuqInterQualKeys)

	return dto.CSUQScoresPayloadDTO{
		SchemaVersion: CSUQSchemaVersion,
		ScaleMin:      1,
		ScaleMax:      7,
		Direction:     "higher_is_better",
		Locale:        NormalizeQuestionnaireLocale(locale),
		Items:         normalizedItems,
		Subscales: dto.CSUQSubscalesDTO{
			SysUse:    sysUseMean,
			InfoQual:  infoQualMean,
			InterQual: interQualMean,
			Overall:   overallMean,
		},
		ValidCounts: dto.CSUQValidCountsDTO{
			SysUse:    sysUseCount,
			InfoQual:  infoQualCount,
			InterQual: interQualCount,
			Overall:   overallCount,
		},
	}, nil
}

func validateTLXRating(ptr *int) (int, bool) {
	if ptr == nil {
		return 0, false
	}
	v := *ptr
	if v < 0 || v > 100 || v%5 != 0 {
		return 0, false
	}
	return v, true
}

// ValidateAndScoreNASATLX validates all 6 Raw NASA-TLX ratings (0..100, step 5),
// transforms intuitive Performance rating (0=Failure..100=Perfect) to workload (100-rating),
// and computes unweighted overall_workload.
func ValidateAndScoreNASATLX(input dto.NASATLXInputDTO) (dto.NASATLXScoresPayloadDTO, error) {
	mental, ok1 := validateTLXRating(input.MentalDemand)
	physical, ok2 := validateTLXRating(input.PhysicalDemand)
	temporal, ok3 := validateTLXRating(input.TemporalDemand)
	perfRating, ok4 := validateTLXRating(input.PerformanceRating)
	effort, ok5 := validateTLXRating(input.Effort)
	frustration, ok6 := validateTLXRating(input.Frustration)
	if !(ok1 && ok2 && ok3 && ok4 && ok5 && ok6) {
		return dto.NASATLXScoresPayloadDTO{}, ErrQuestionnaireInvalidTLX
	}

	perfWorkload := 100 - perfRating
	overall := roundEval2(float64(mental+physical+temporal+perfWorkload+effort+frustration) / 6.0)

	return dto.NASATLXScoresPayloadDTO{
		SchemaVersion: NASATLXSchemaVersion,
		Variant:       "raw_tlx",
		ScaleMin:      0,
		ScaleMax:      100,
		Step:          5,
		Dimensions: dto.NASATLXDimensionsDTO{
			MentalDemand:        mental,
			PhysicalDemand:      physical,
			TemporalDemand:      temporal,
			PerformanceRating:   perfRating,
			PerformanceWorkload: perfWorkload,
			Effort:              effort,
			Frustration:         frustration,
		},
		OverallWorkload: overall,
	}, nil
}

// SanitizeQualitativeFeedback trims each optional text field, enforces the 2,000-rune cap,
// and formats non-empty sections into readable notes prose.
func SanitizeQualitativeFeedback(input dto.QualitativeFeedbackDTO) (dto.QualitativeFeedbackDTO, string, error) {
	cleaned := dto.QualitativeFeedbackDTO{
		MostImpressive:     strings.TrimSpace(input.MostImpressive),
		AreasToImprove:     strings.TrimSpace(input.AreasToImprove),
		AdditionalComments: strings.TrimSpace(input.AdditionalComments),
	}

	for _, text := range []string{cleaned.MostImpressive, cleaned.AreasToImprove, cleaned.AdditionalComments} {
		if utf8.RuneCountInString(text) > MaxQualitativeFeedbackRunes {
			return dto.QualitativeFeedbackDTO{}, "", ErrQuestionnaireFeedbackTooLong
		}
	}

	sections := make([]string, 0, 3)
	if cleaned.MostImpressive != "" {
		sections = append(sections, "[Most Impressive]\n"+cleaned.MostImpressive)
	}
	if cleaned.AreasToImprove != "" {
		sections = append(sections, "[Areas to Improve]\n"+cleaned.AreasToImprove)
	}
	if cleaned.AdditionalComments != "" {
		sections = append(sections, "[Additional Comments]\n"+cleaned.AdditionalComments)
	}
	return cleaned, strings.Join(sections, "\n\n"), nil
}

// DecodeJSONBInto decodes a GORM JSONB any-typed field cleanly into a typed struct.
func DecodeJSONBInto(raw any, dest any) error {
	switch v := raw.(type) {
	case nil:
		return nil
	case []byte:
		if len(v) == 0 || string(v) == "null" {
			return nil
		}
		return json.Unmarshal(v, dest)
	case json.RawMessage:
		if len(v) == 0 || string(v) == "null" {
			return nil
		}
		return json.Unmarshal(v, dest)
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" || trimmed == "null" {
			return nil
		}
		return json.Unmarshal([]byte(trimmed), dest)
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return err
		}
		if string(bytes) == "null" {
			return nil
		}
		return json.Unmarshal(bytes, dest)
	}
}

// EvaluationMetricToSummaryDTO converts a persisted EvaluationMetric row into the canonical API summary DTO.
func EvaluationMetricToSummaryDTO(m models.EvaluationMetric) dto.EvaluationMetricSummaryDTO {
	var csuq dto.CSUQScoresPayloadDTO
	_ = DecodeJSONBInto(m.CSUQScores, &csuq)

	var tlx dto.NASATLXScoresPayloadDTO
	_ = DecodeJSONBInto(m.NASATLXScores, &tlx)

	var snap dto.EvaluationSkillGapSnapshotDTO
	_ = DecodeJSONBInto(m.SkillGapSummary, &snap)

	locale := snap.SubmittedLocale
	if locale == "" {
		locale = csuq.Locale
	}
	locale = NormalizeQuestionnaireLocale(locale)

	return dto.EvaluationMetricSummaryDTO{
		EvalID:              m.EvalID,
		PreScore:            roundEval2(m.PreScore),
		PostScore:           roundEval2(m.PostScore),
		ImprovementPct:      roundEval2(m.ImprovementPct),
		SubmittedLocale:     locale,
		CompletedScenarios:  snap.CompletedScenarios,
		AverageProficiency:  roundEval2(snap.AverageProficiency),
		CSUQScores:          csuq,
		NASATLXScores:       tlx,
		QualitativeFeedback: snap.QualitativeFeedback,
		Notes:               m.Notes,
		EvalDate:            m.EvalDate,
	}
}
