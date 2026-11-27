package services

import (
	"encoding/json"
	"math"
	"strings"

	"cybersim/dto"
)

// ExamMaxDurationSeconds defines the maximum allowed exam elapsed duration cap (2 hours).
const ExamMaxDurationSeconds = 7200

// ExamReviewData bundles computed review totals, domain breakdown, and ordered Q&A items.
type ExamReviewData struct {
	CorrectCount    int64
	IncorrectCount  int64
	UnansweredCount int64
	DomainScores    []dto.DomainScoreDTO
	Results         []dto.DetailedExamResultDTO
}

// RoundExamScore rounds percentages and deltas to 2 decimal places.
func RoundExamScore(val float64) float64 {
	return math.Round(val*100) / 100
}

// ClampExamDuration bounds elapsed seconds within [0, ExamMaxDurationSeconds] and
// preserves the greater of existing vs. newly reported total time.
func ClampExamDuration(reported, existing int) int {
	if reported < 0 {
		reported = 0
	}
	if reported > ExamMaxDurationSeconds {
		reported = ExamMaxDurationSeconds
	}
	if existing > reported {
		return existing
	}
	return reported
}

// ComputeSessionScore returns a persisted session score if present, otherwise computes
// it from correctCount and totalQuestions.
func ComputeSessionScore(storedScore *float64, correctCount int64, totalQuestions int) float64 {
	if storedScore != nil {
		return RoundExamScore(*storedScore)
	}
	if totalQuestions <= 0 {
		return 0
	}
	return RoundExamScore(float64(correctCount) / float64(totalQuestions) * 100)
}

// EnrichExamResults populates canonical Security+ domain names, SY0-701 sub-topic titles,
// and unanswered indicators onto ordered exam results.
func EnrichExamResults(results []dto.DetailedExamResultDTO) {
	domainTitles := make(map[string]string, len(OrderedTheoryDomains))
	for _, d := range OrderedTheoryDomains {
		domainTitles[d.DomainID] = d.DomainName
	}
	subtopicTitles := make(map[string]string, len(OrderedTheorySubtopics))
	for _, spec := range OrderedTheorySubtopics {
		subtopicTitles[spec.SubTopicID] = spec.Title
	}

	for i := range results {
		if name, ok := domainTitles[results[i].DomainID]; ok {
			results[i].DomainName = name
		}
		if results[i].SubTopicID != nil {
			results[i].SubTopicTitle = subtopicTitles[strings.TrimSpace(*results[i].SubTopicID)]
		}
		results[i].IsUnanswered = strings.TrimSpace(results[i].UserAnswer) == ""
	}
}

// SummarizeExamResults aggregates overall outcomes and canonical domain1..domain5 scores.
func SummarizeExamResults(results []dto.DetailedExamResultDTO, threshold float64) *ExamReviewData {
	EnrichExamResults(results)

	type domainAgg struct {
		total      int64
		correct    int64
		incorrect  int64
		unanswered int64
	}
	byDomain := make(map[string]*domainAgg, len(OrderedTheoryDomains))
	var correctCount, incorrectCount, unansweredCount int64

	for _, r := range results {
		agg, ok := byDomain[r.DomainID]
		if !ok {
			agg = &domainAgg{}
			byDomain[r.DomainID] = agg
		}
		agg.total++
		if r.IsCorrect {
			correctCount++
			agg.correct++
		} else if r.IsUnanswered {
			unansweredCount++
			agg.unanswered++
		} else {
			incorrectCount++
			agg.incorrect++
		}
	}

	domainScores := make([]dto.DomainScoreDTO, 0, len(OrderedTheoryDomains))
	for _, spec := range OrderedTheoryDomains {
		agg, ok := byDomain[spec.DomainID]
		if !ok || agg.total == 0 {
			continue
		}
		pct := RoundExamScore(float64(agg.correct) / float64(agg.total) * 100)
		domainScores = append(domainScores, dto.DomainScoreDTO{
			DomainID:     spec.DomainID,
			DomainName:   spec.DomainName,
			TotalCount:   agg.total,
			CorrectCount: agg.correct,
			Incorrect:    agg.incorrect,
			Unanswered:   agg.unanswered,
			Percentage:   pct,
			Threshold:    threshold,
			Passed:       pct >= threshold,
		})
	}

	return &ExamReviewData{
		CorrectCount:    correctCount,
		IncorrectCount:  incorrectCount,
		UnansweredCount: unansweredCount,
		DomainScores:    domainScores,
		Results:         results,
	}
}

// BuildPreComparison calculates overall and 5-domain Pre-Test to Post-Test learning gains.
func BuildPreComparison(
	preSessionID string,
	preStoredScore *float64,
	preTotal int,
	preReview *ExamReviewData,
	postStoredScore *float64,
	postTotal int,
	postReview *ExamReviewData,
) *dto.ExamPreComparisonDTO {
	if preSessionID == "" || preReview == nil || postReview == nil {
		return nil
	}
	preScore := ComputeSessionScore(preStoredScore, preReview.CorrectCount, preTotal)
	postScore := ComputeSessionScore(postStoredScore, postReview.CorrectCount, postTotal)

	preMap := make(map[string]float64, len(preReview.DomainScores))
	for _, d := range preReview.DomainScores {
		preMap[d.DomainID] = d.Percentage
	}
	postMap := make(map[string]float64, len(postReview.DomainScores))
	for _, d := range postReview.DomainScores {
		postMap[d.DomainID] = d.Percentage
	}

	deltas := make([]dto.ExamDomainDeltaDTO, 0, len(OrderedTheoryDomains))
	for _, spec := range OrderedTheoryDomains {
		prePct, hasPre := preMap[spec.DomainID]
		postPct, hasPost := postMap[spec.DomainID]
		if !hasPre && !hasPost {
			continue
		}
		deltas = append(deltas, dto.ExamDomainDeltaDTO{
			DomainID:       spec.DomainID,
			DomainName:     spec.DomainName,
			PrePercentage:  prePct,
			PostPercentage: postPct,
			Delta:          RoundExamScore(postPct - prePct),
		})
	}

	return &dto.ExamPreComparisonDTO{
		PreSessionID:     preSessionID,
		PreScore:         preScore,
		PostScore:        postScore,
		ScoreDelta:       RoundExamScore(postScore - preScore),
		PreCorrectCount:  preReview.CorrectCount,
		PostCorrectCount: postReview.CorrectCount,
		DomainDeltas:     deltas,
	}
}

// SanitizeQuestionTranslations decodes stored per-locale question translations and
// omits nested explanation/correct_answer fields unless includeExplanation is true.
// Returns untyped nil when absent/empty so json:"translations,omitempty" omits the field.
func SanitizeQuestionTranslations(raw any, includeExplanation bool) any {
	if raw == nil {
		return nil
	}

	var decoded map[string]any
	switch v := raw.(type) {
	case map[string]any:
		decoded = v
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" || trimmed == "{}" || trimmed == "null" {
			return nil
		}
		if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
			return nil
		}
	case []byte:
		trimmed := strings.TrimSpace(string(v))
		if trimmed == "" || trimmed == "{}" || trimmed == "null" {
			return nil
		}
		if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
			return nil
		}
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return nil
		}
		if err := json.Unmarshal(b, &decoded); err != nil {
			return nil
		}
	}

	if len(decoded) == 0 {
		return nil
	}

	sanitized := make(map[string]any)
	for localeKey, localeVal := range decoded {
		locMap, ok := localeVal.(map[string]any)
		if !ok || len(locMap) == 0 {
			continue
		}
		cleanLocale := make(map[string]any)
		if qText, ok := locMap["question_text"].(string); ok && strings.TrimSpace(qText) != "" {
			cleanLocale["question_text"] = qText
		}
		if opts, ok := locMap["options"].(map[string]any); ok && len(opts) > 0 {
			cleanOpts := make(map[string]any, len(opts))
			for k, optVal := range opts {
				cleanOpts[k] = optVal
			}
			cleanLocale["options"] = cleanOpts
		} else if optsStr, ok := locMap["options"].(map[string]string); ok && len(optsStr) > 0 {
			cleanOpts := make(map[string]any, len(optsStr))
			for k, optVal := range optsStr {
				cleanOpts[k] = optVal
			}
			cleanLocale["options"] = cleanOpts
		}
		if includeExplanation {
			if exp, ok := locMap["explanation"].(string); ok && strings.TrimSpace(exp) != "" {
				cleanLocale["explanation"] = exp
			}
		}
		if len(cleanLocale) > 0 {
			sanitized[localeKey] = cleanLocale
		}
	}

	if len(sanitized) == 0 {
		return nil
	}
	return sanitized
}
