package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math"
	"strings"
	"time"

	"cybersim/dto"
	"cybersim/models"
)

// AdminLearnerEvaluationRecord pairs a learner's EvaluationMetric row with pseudonymized identity fields.
type AdminLearnerEvaluationRecord struct {
	UserID   string
	Nickname string
	Metric   models.EvaluationMetric
}

// PseudonymizedParticipantCode generates a deterministic privacy-preserving code (P-XXXXXXXX).
func PseudonymizedParticipantCode(userID string) string {
	clean := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(userID), "-", ""))
	if len(clean) >= 8 {
		return "P-" + clean[:8]
	}
	if clean == "" {
		return "P-UNKNOWN"
	}
	return "P-" + clean
}

func computeMeanAndSampleSD(values []float64) (*float64, *float64) {
	n := len(values)
	if n == 0 {
		return nil, nil
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(n)
	meanRounded := roundEval2(mean)
	if n < 2 {
		return &meanRounded, nil
	}
	sq := 0.0
	for _, v := range values {
		diff := v - mean
		sq += diff * diff
	}
	sd := roundEval2(math.Sqrt(sq / float64(n-1)))
	return &meanRounded, &sd
}

// BuildAdminEvaluationsResponse aggregates KPI summaries and pseudonymized learner rows.
func BuildAdminEvaluationsResponse(
	records []AdminLearnerEvaluationRecord,
	cohort dto.EvaluationCohortFunnelDTO,
) dto.AdminEvaluationsResponse {
	series := collectCohortMetricSeries(records)
	responses := make([]dto.AdminEvaluationRowDTO, 0, len(records))
	for i, rec := range records {
		summary := series.summaries[i]
		responses = append(responses, dto.AdminEvaluationRowDTO{
			EvalID:              summary.EvalID,
			ParticipantCode:     PseudonymizedParticipantCode(rec.UserID),
			Nickname:            rec.Nickname,
			SubmittedLocale:     summary.SubmittedLocale,
			EvalDate:            summary.EvalDate,
			PreScore:            summary.PreScore,
			PostScore:           summary.PostScore,
			ImprovementPct:      summary.ImprovementPct,
			CompletedScenarios:  summary.CompletedScenarios,
			AverageProficiency:  summary.AverageProficiency,
			CSUQScores:          summary.CSUQScores,
			NASATLXScores:       summary.NASATLXScores,
			QualitativeFeedback: summary.QualitativeFeedback,
			Notes:               summary.Notes,
		})
	}

	makeStat := func(key string, vals []float64) dto.AdminMeanSDDTO {
		mean, sd := computeMeanAndSampleSD(vals)
		return dto.AdminMeanSDDTO{Key: key, N: len(vals), Mean: mean, SD: sd}
	}

	preMean, _ := computeMeanAndSampleSD(series.preVals)
	postMean, _ := computeMeanAndSampleSD(series.postVals)
	impMean, impSD := computeMeanAndSampleSD(series.impVals)

	return dto.AdminEvaluationsResponse{
		Cohort: cohort,
		Summary: dto.AdminEvaluationsSummaryDTO{
			CSUQSubscales: []dto.AdminMeanSDDTO{
				makeStat("sysuse", series.sysuseVals),
				makeStat("infoqual", series.infoqualVals),
				makeStat("interqual", series.interqualVals),
				makeStat("overall", series.csuqOverallVals),
			},
			NASATLXDimensions: []dto.AdminMeanSDDTO{
				makeStat("mental_demand", series.mentalVals),
				makeStat("physical_demand", series.physicalVals),
				makeStat("temporal_demand", series.temporalVals),
				makeStat("performance_workload", series.perfWorkloadVals),
				makeStat("effort", series.effortVals),
				makeStat("frustration", series.frustrationVals),
				makeStat("overall_workload", series.tlxOverallVals),
			},
			LearningGain: dto.AdminLearningGainSummaryDTO{
				N:                len(records),
				PreMean:          preMean,
				PostMean:         postMean,
				ImprovementMean:  impMean,
				ImprovementSD:    impSD,
				QualitativeCount: series.qualCount,
			},
		},
		Responses: responses,
	}
}

// NeutralizeCSVFormula prevents spreadsheet formula injection on free-text cells.
func NeutralizeCSVFormula(val string) string {
	if val == "" {
		return ""
	}
	switch val[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + val
	default:
		return val
	}
}

func formatOptionalFloat2(ptr *float64) string {
	if ptr == nil {
		return ""
	}
	return fmt.Sprintf("%.2f", *ptr)
}

func formatOptionalIntPtr(ptr *int) string {
	if ptr == nil {
		return ""
	}
	return fmt.Sprintf("%d", *ptr)
}

// BuildEvaluationsCSV builds an RFC 4180 CSV with UTF-8 BOM and neutralized free-text fields.
func BuildEvaluationsCSV(rows []dto.AdminEvaluationRowDTO) ([]byte, error) {
	var buf bytes.Buffer
	// UTF-8 BOM so Microsoft Excel detects UTF-8 (Thai characters) cleanly.
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(&buf)
	headers := []string{
		"participant_code", "nickname", "eval_date", "locale",
		"pre_score", "post_score", "improvement_pct",
		"csuq_q1", "csuq_q2", "csuq_q3", "csuq_q4", "csuq_q5", "csuq_q6",
		"csuq_q7", "csuq_q8", "csuq_q9", "csuq_q10", "csuq_q11", "csuq_q12",
		"csuq_q13", "csuq_q14", "csuq_q15", "csuq_q16",
		"csuq_sysuse", "csuq_infoqual", "csuq_interqual", "csuq_overall",
		"tlx_mental_demand", "tlx_physical_demand", "tlx_temporal_demand",
		"tlx_performance_rating", "tlx_performance_workload", "tlx_effort",
		"tlx_frustration", "tlx_overall_workload",
		"feedback_most_impressive", "feedback_areas_to_improve", "feedback_additional_comments", "notes",
	}
	if err := w.Write(headers); err != nil {
		return nil, err
	}

	for _, r := range rows {
		rec := make([]string, 0, len(headers))
		rec = append(rec,
			r.ParticipantCode,
			NeutralizeCSVFormula(r.Nickname),
			r.EvalDate.UTC().Format(time.RFC3339),
			r.SubmittedLocale,
			fmt.Sprintf("%.2f", r.PreScore),
			fmt.Sprintf("%.2f", r.PostScore),
			fmt.Sprintf("%.2f", r.ImprovementPct),
		)
		for _, k := range CSUQItemKeys {
			var ptr *int
			if r.CSUQScores.Items != nil {
				ptr = r.CSUQScores.Items[k]
			}
			rec = append(rec, formatOptionalIntPtr(ptr))
		}
		dims := r.NASATLXScores.Dimensions
		rec = append(rec,
			formatOptionalFloat2(r.CSUQScores.Subscales.SysUse),
			formatOptionalFloat2(r.CSUQScores.Subscales.InfoQual),
			formatOptionalFloat2(r.CSUQScores.Subscales.InterQual),
			formatOptionalFloat2(r.CSUQScores.Subscales.Overall),
			fmt.Sprintf("%d", dims.MentalDemand),
			fmt.Sprintf("%d", dims.PhysicalDemand),
			fmt.Sprintf("%d", dims.TemporalDemand),
			fmt.Sprintf("%d", dims.PerformanceRating),
			fmt.Sprintf("%d", dims.PerformanceWorkload),
			fmt.Sprintf("%d", dims.Effort),
			fmt.Sprintf("%d", dims.Frustration),
			fmt.Sprintf("%.2f", r.NASATLXScores.OverallWorkload),
			NeutralizeCSVFormula(r.QualitativeFeedback.MostImpressive),
			NeutralizeCSVFormula(r.QualitativeFeedback.AreasToImprove),
			NeutralizeCSVFormula(r.QualitativeFeedback.AdditionalComments),
			NeutralizeCSVFormula(r.Notes),
		)
		if err := w.Write(rec); err != nil {
			return nil, err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
