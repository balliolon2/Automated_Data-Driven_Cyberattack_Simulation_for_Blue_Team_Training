package services

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"cybersim/dto"
	"cybersim/models"
)

func makeAdminLearnerRecord(userID, nickname string, pre, post float64, q1, perfRating int, impressive string) AdminLearnerEvaluationRecord {
	csuqInput := makeValidCSUQInput()
	csuqInput["q1"] = intPtr(q1)
	csuqPayload, _ := ValidateAndScoreCSUQ(csuqInput, "th")

	tlxInput := makeValidTLXInput()
	tlxInput.PerformanceRating = intPtr(perfRating)
	tlxPayload, _ := ValidateAndScoreNASATLX(tlxInput)

	fb, notes, _ := SanitizeQualitativeFeedback(dto.QualitativeFeedbackDTO{
		MostImpressive: impressive,
	})

	snap := dto.EvaluationSkillGapSnapshotDTO{
		SnapshotVersion:     EvalSnapshotVersion,
		SubmittedLocale:     "th",
		PostSessionID:       "post-sess-1",
		CompletedScenarios:  4,
		AverageProficiency:  78.5,
		QualitativeFeedback: fb,
	}

	return AdminLearnerEvaluationRecord{
		UserID:   userID,
		Nickname: nickname,
		Metric: models.EvaluationMetric{
			EvalID:          "eval-" + userID,
			UserID:          userID,
			PreScore:        pre,
			PostScore:       post,
			ImprovementPct:  roundEval2(post - pre),
			CSUQScores:      csuqPayload,
			NASATLXScores:   tlxPayload,
			SkillGapSummary: snap,
			EvalDate:        time.Date(2026, 11, 25, 10, 0, 0, 0, time.UTC),
			Notes:           notes,
		},
	}
}

func TestBuildAdminEvaluationsResponse_SummaryAndPseudonymization(t *testing.T) {
	records := []AdminLearnerEvaluationRecord{
		makeAdminLearnerRecord("abcdef12-3456-7890-abcd-ef1234567890", "alice_soc", 55.0, 80.0, 6, 80, "=HYPERLINK(\"http://evil\")"),
		makeAdminLearnerRecord("12345678-90ab-cdef-1234-567890abcdef", "bob_blue", 65.0, 90.0, 7, 90, ""),
	}
	cohort := dto.EvaluationCohortFunnelDTO{
		TotalLearners:             10,
		CompletedPreTestLearners:  8,
		CompletedPostTestLearners: 4,
		QuestionnaireRespondents:  2,
		ResponseRatePct:           50.0,
	}

	resp := BuildAdminEvaluationsResponse(records, cohort)
	if len(resp.Responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(resp.Responses))
	}
	if resp.Responses[0].ParticipantCode != "P-ABCDEF12" || resp.Responses[0].Nickname != "alice_soc" {
		t.Fatalf("unexpected pseudonymized participant row: %+v", resp.Responses[0])
	}
	if resp.Summary.LearningGain.N != 2 || resp.Summary.LearningGain.QualitativeCount != 1 {
		t.Fatalf("unexpected learning gain summary: %+v", resp.Summary.LearningGain)
	}
	if resp.Summary.LearningGain.ImprovementMean == nil || *resp.Summary.LearningGain.ImprovementMean != 25.0 {
		t.Fatalf("expected improvement mean 25.0, got %v", resp.Summary.LearningGain.ImprovementMean)
	}
	if len(resp.Summary.CSUQSubscales) != 4 || len(resp.Summary.NASATLXDimensions) != 7 {
		t.Fatalf("unexpected subscale/dimension summary lengths: csuq=%d tlx=%d", len(resp.Summary.CSUQSubscales), len(resp.Summary.NASATLXDimensions))
	}
}

func TestNeutralizeCSVFormulaAndBuildCSV(t *testing.T) {
	cases := map[string]string{
		"=cmd|' /C calc'!A0": "'=cmd|' /C calc'!A0",
		"+SUM(A1:A2)":        "'+SUM(A1:A2)",
		"-1+1":               "'-1+1",
		"@SUM(1,1)":          "'@SUM(1,1)",
		"\tcmd":              "'\tcmd",
		"\rcmd":              "'\rcmd",
		"Normal Thai ภาษาไทย": "Normal Thai ภาษาไทย",
		"":                   "",
	}
	for in, want := range cases {
		if got := NeutralizeCSVFormula(in); got != want {
			t.Fatalf("NeutralizeCSVFormula(%q) = %q, want %q", in, got, want)
		}
	}

	records := []AdminLearnerEvaluationRecord{
		makeAdminLearnerRecord("abcdef12-3456-7890-abcd-ef1234567890", "+evil_nick", 55.0, 80.0, 6, 80, "=1+1"),
	}
	resp := BuildAdminEvaluationsResponse(records, dto.EvaluationCohortFunnelDTO{})
	csvBytes, err := BuildEvaluationsCSV(resp.Responses)
	if err != nil {
		t.Fatalf("BuildEvaluationsCSV failed: %v", err)
	}
	if !bytes.HasPrefix(csvBytes, []byte{0xEF, 0xBB, 0xBF}) {
		t.Fatalf("expected UTF-8 BOM prefix")
	}

	reader := csv.NewReader(bytes.NewReader(csvBytes[3:]))
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("CSV parse error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected header + 1 data row, got %d", len(rows))
	}
	header := rows[0]
	row := rows[1]
	if header[0] != "participant_code" || header[len(header)-1] != "notes" {
		t.Fatalf("unexpected CSV header boundaries: first=%s last=%s", header[0], header[len(header)-1])
	}
	if row[1] != "'+evil_nick" {
		t.Fatalf("expected formula-neutralized nickname, got %q", row[1])
	}
	if !strings.HasPrefix(row[len(row)-4], "'=1+1") {
		t.Fatalf("expected formula-neutralized most_impressive, got %q", row[len(row)-4])
	}
	// q7 was nil (N/A) in makeValidCSUQInput -> empty cell
	if row[13] != "" {
		t.Fatalf("expected empty cell for N/A q7, got %q", row[13])
	}
}
