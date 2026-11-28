package services

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

	"cybersim/dto"
)

func assertClose(t *testing.T, label string, got, want, tol float64) {
	t.Helper()
	if math.IsNaN(got) || math.IsInf(got, 0) || math.Abs(got-want) > tol {
		t.Fatalf("%s: got %.6f, want %.6f (tol %.6f)", label, got, want, tol)
	}
}

func TestStudentTCriticalAndPValueReference(t *testing.T) {
	// Known exact t_0.975 critical values (two-tailed alpha=0.05)
	assertClose(t, "t_crit(df=1)", studentTCritical95(1), 12.706205, 1e-4)
	assertClose(t, "t_crit(df=4)", studentTCritical95(4), 2.776445, 1e-4)
	assertClose(t, "t_crit(df=9)", studentTCritical95(9), 2.262157, 1e-4)
	assertClose(t, "t_crit(df=29)", studentTCritical95(29), 2.045230, 1e-4)

	// Two-tailed p-value sanity
	assertClose(t, "t_p(0, df=5)", twoTailedTPValue(0, 5), 1.0, 1e-6)
	assertClose(t, "t_p(2.776445, df=4)", twoTailedTPValue(2.776445, 4), 0.05, 1e-4)
}

func TestComputeDescriptiveStat_SmallNAndFull(t *testing.T) {
	// n = 0
	s0 := ComputeDescriptiveStat("overall", "CSUQ", nil, bandCSUQ)
	if s0.N != 0 || s0.Mean != nil || s0.SD != nil || s0.Status != "insufficient_n" {
		t.Fatalf("expected empty insufficient_n, got %+v", s0)
	}

	// n = 1 -> mean/median/min/max present, SD and 95% CI nil
	s1 := ComputeDescriptiveStat("overall", "CSUQ", []float64{6.25}, bandCSUQ)
	if s1.N != 1 || s1.Mean == nil || *s1.Mean != 6.25 || s1.SD != nil || s1.CILow != nil || s1.Status != "insufficient_n" {
		t.Fatalf("unexpected n=1 descriptive stat: %+v", s1)
	}

	// n = 5 -> [10, 20, 30, 40, 50]: mean=30, SD=15.8114, SE=7.0711, t_crit(4)=2.776445 => CI=[10.3676, 49.6324], Q1=20, Med=30, Q3=40, IQR=20
	s5 := ComputeDescriptiveStat("mental_demand", "TLX", []float64{40, 10, 50, 20, 30}, bandTLX)
	if s5.Status != "ok" || s5.N != 5 {
		t.Fatalf("expected ok status for n=5, got %+v", s5)
	}
	assertClose(t, "mean", *s5.Mean, 30.0, 1e-3)
	assertClose(t, "sd", *s5.SD, 15.81, 1e-2)
	assertClose(t, "ci_low", *s5.CILow, 10.37, 1e-2)
	assertClose(t, "ci_high", *s5.CIHigh, 49.63, 1e-2)
	assertClose(t, "q1", *s5.Q1, 20.0, 1e-4)
	assertClose(t, "median", *s5.Median, 30.0, 1e-4)
	assertClose(t, "q3", *s5.Q3, 40.0, 1e-4)
	assertClose(t, "iqr", *s5.IQR, 20.0, 1e-4)
	if s5.Band != "moderate" {
		t.Fatalf("expected TLX moderate band for mean 30.0, got %q", s5.Band)
	}
}

func TestComputeCronbachAlpha_ReferenceAndGuardrails(t *testing.T) {
	// 4 respondents x 3 items:
	// R1: [6, 6, 7] -> T=19
	// R2: [5, 5, 6] -> T=16
	// R3: [4, 5, 4] -> T=13
	// R4: [7, 7, 7] -> T=21
	// Var(X1) = 1.6667, Var(X2) = 0.9167, Var(X3) = 2.0000 => sum = 4.5833
	// Var(T) = 12.2500 => alpha = (3/2)*(1 - 4.5833/12.2500) = 0.9388
	rows := [][]float64{
		{6, 6, 7},
		{5, 5, 6},
		{4, 5, 4},
		{7, 7, 7},
	}
	res := ComputeCronbachAlpha("csuq_interqual", 3, rows, 6)
	if res.Status != "ok" || res.CompleteCasesN != 4 || res.ExcludedNACases != 2 {
		t.Fatalf("unexpected alpha metadata: %+v", res)
	}
	assertClose(t, "cronbach_alpha", *res.Alpha, 0.9388, 1e-3)
	if res.Band != "excellent" {
		t.Fatalf("expected excellent band, got %q", res.Band)
	}

	// Zero variance across items -> zero_variance guard
	zeroRows := [][]float64{{5, 5, 5}, {5, 5, 5}, {5, 5, 5}}
	zRes := ComputeCronbachAlpha("csuq_interqual", 3, zeroRows, 3)
	if zRes.Alpha != nil || zRes.Status != "zero_variance" {
		t.Fatalf("expected zero_variance guard, got %+v", zRes)
	}

	// n < 2 -> insufficient_n
	oneRow := ComputeCronbachAlpha("csuq_interqual", 3, [][]float64{{5, 6, 7}}, 2)
	if oneRow.Alpha != nil || oneRow.Status != "insufficient_n" {
		t.Fatalf("expected insufficient_n guard, got %+v", oneRow)
	}
}

func TestComputePairedTTestAndWilcoxon_ReferenceAndGuardrails(t *testing.T) {
	// Pre:  [50, 55, 60, 45, 70]
	// Post: [75, 80, 78, 70, 88]
	// Diff: [25, 25, 18, 25, 18] => mean=22.2, sd=3.834058, t=12.9471, df=4, dz=5.7901
	pre := []float64{50, 55, 60, 45, 70}
	post := []float64{75, 80, 78, 70, 88}

	tRes, wRes := ComputePairedInference(pre, post)
	if tRes.Status != "ok" || tRes.DF != 4 {
		t.Fatalf("unexpected paired t-test status: %+v", tRes)
	}
	assertClose(t, "t_mean_diff", *tRes.MeanDiff, 22.2, 1e-3)
	assertClose(t, "t_sd_diff", *tRes.SDDiff, 3.83, 1e-2)
	assertClose(t, "t_stat", *tRes.TStat, 12.9471, 1e-2)
	assertClose(t, "cohens_dz", *tRes.CohensDz, 5.7901, 1e-2)
	if tRes.EffectBand != "large" || *tRes.PValue >= 0.001 {
		t.Fatalf("expected large significant effect, got p=%v band=%s", tRes.PValue, tRes.EffectBand)
	}

	// With ties [25,25,25,18,18] and all positive -> W-=0, W+=15, r_rb=1.0
	if wRes.Status != "ok" || wRes.EffectiveN != 5 || *wRes.WStat != 0 || *wRes.RankBiserial != 1.0 {
		t.Fatalf("unexpected wilcoxon with ties: %+v", wRes)
	}

	// Untied exact n=5 all positive [10, 12, 15, 20, 25]: exact two-tailed p = 2*(1/32) = 0.0625
	_, wExact := ComputePairedInference([]float64{50, 50, 50, 50, 50}, []float64{60, 62, 65, 70, 75})
	if wExact.Method != "exact" || wExact.Status != "ok" {
		t.Fatalf("expected exact wilcoxon method for untied n=5, got %+v", wExact)
	}
	assertClose(t, "wilcoxon_exact_p", *wExact.PValue, 0.0625, 1e-4)

	// Guardrails: n=2 (<3) gates both t-test and wilcoxon; n=4 gates wilcoxon only
	t2, w2 := ComputePairedInference([]float64{50, 60}, []float64{70, 80})
	if t2.Status != "insufficient_n" || w2.Status != "insufficient_n" {
		t.Fatalf("expected n=2 insufficient_n for both tests, got t=%s w=%s", t2.Status, w2.Status)
	}
	t4, w4 := ComputePairedInference([]float64{50, 55, 60, 65}, []float64{70, 72, 80, 85})
	if t4.Status != "ok" || w4.Status != "insufficient_n" {
		t.Fatalf("expected n=4 ok for t-test and insufficient_n for wilcoxon, got t=%s w=%s", t4.Status, w4.Status)
	}

	// Zero-diff count drops below effectiveN 5 -> wilcoxon insufficient_n
	_, wZeroDrop := ComputePairedInference([]float64{50, 50, 50, 50, 50}, []float64{50, 60, 70, 80, 90})
	if wZeroDrop.ZeroDiffCount != 1 || wZeroDrop.EffectiveN != 4 || wZeroDrop.Status != "insufficient_n" {
		t.Fatalf("expected zero-diff drop to gate wilcoxon at effective_n=4, got %+v", wZeroDrop)
	}
}

func TestComputeSpearmanCorrelation_ReferenceAndGuardrails(t *testing.T) {
	// Perfect positive monotonic: [10, 20, 30, 40, 50] vs [1, 2, 3, 4, 5] => rho = 1.0, p = 0.0
	perfect := ComputeSpearmanCorrelation("gain_vs_csuq", []float64{10, 20, 30, 40, 50}, []float64{1, 2, 3, 4, 5})
	if perfect.Status != "ok" || perfect.Band != "strong" {
		t.Fatalf("unexpected perfect spearman: %+v", perfect)
	}
	assertClose(t, "spearman_perfect_rho", *perfect.Rho, 1.0, 1e-6)
	assertClose(t, "spearman_perfect_p", *perfect.PValue, 0.0, 1e-6)

	// Known tied dataset: x=[1,2,2,4,5], y=[10,20,30,20,50]
	// R(x)=[1, 2.5, 2.5, 4, 5], R(y)=[1, 2.5, 4, 2.5, 5] => sumProd=7.25, sumSq=9.5 => rho=7.25/9.5=0.763158
	tied := ComputeSpearmanCorrelation("tied_pair", []float64{1, 2, 2, 4, 5}, []float64{10, 20, 30, 20, 50})
	assertClose(t, "spearman_tied_rho", *tied.Rho, 0.7632, 1e-3)

	// n < 3 -> insufficient_n
	small := ComputeSpearmanCorrelation("small", []float64{1, 2}, []float64{3, 4})
	if small.Status != "insufficient_n" || small.Rho != nil {
		t.Fatalf("expected n=2 insufficient_n, got %+v", small)
	}

	// Zero variance -> zero_variance
	zv := ComputeSpearmanCorrelation("zv", []float64{5, 5, 5, 5}, []float64{1, 2, 3, 4})
	if zv.Status != "zero_variance" || zv.Rho != nil {
		t.Fatalf("expected zero_variance guard, got %+v", zv)
	}
}

func TestBuildPaperStatisticsReport_FullPipelineAndJSONSafety(t *testing.T) {
	records := []AdminLearnerEvaluationRecord{
		makeAdminLearnerRecord("u1-uuid-0001", "u1", 50, 80, 6, 85, "Great"),
		makeAdminLearnerRecord("u2-uuid-0002", "u2", 55, 82, 7, 90, ""),
		makeAdminLearnerRecord("u3-uuid-0003", "u3", 60, 85, 5, 75, "Nice"),
		makeAdminLearnerRecord("u4-uuid-0004", "u4", 48, 76, 6, 80, ""),
		makeAdminLearnerRecord("u5-uuid-0005", "u5", 62, 91, 7, 95, "Clear"),
	}
	cohort := dto.EvaluationCohortFunnelDTO{
		TotalLearners: 10, CompletedPreTestLearners: 8, CompletedPostTestLearners: 6,
		QuestionnaireRespondents: 5, ResponseRatePct: 83.33,
	}

	report := BuildPaperStatisticsReport(records, cohort)
	if !report.IsPilotSample {
		t.Fatalf("expected IsPilotSample=true for n=5 (<30)")
	}
	if len(report.CSUQSubscales) != 4 || len(report.CSUQItems) != 16 || len(report.NASATLX) != 8 {
		t.Fatalf("unexpected descriptor counts: subscales=%d items=%d tlx=%d", len(report.CSUQSubscales), len(report.CSUQItems), len(report.NASATLX))
	}
	if len(report.Reliability) != 5 || len(report.Correlations) != 3 {
		t.Fatalf("unexpected reliability/correlation counts: rel=%d corr=%d", len(report.Reliability), len(report.Correlations))
	}

	// Verify ParticipantPoints chronological order, pseudonymization, nil handling & JSON safety
	if len(report.ParticipantPoints) != 5 {
		t.Fatalf("expected 5 ParticipantPoints, got %d", len(report.ParticipantPoints))
	}

	// Out-of-order DESC timestamps must sort ASC deterministically without PII leakage
	rLate := makeAdminLearnerRecord("bbbb2222-uuid", "late_nick", 60, 88, 7, 90, "")
	rLate.Metric.EvalDate = time.Date(2026, 11, 28, 12, 0, 0, 0, time.UTC)
	rEarly := makeAdminLearnerRecord("aaaa1111-uuid", "early_nick", 45, 75, 6, 80, "")
	rEarly.Metric.EvalDate = time.Date(2026, 11, 28, 9, 0, 0, 0, time.UTC)
	rNilCsuq := makeAdminLearnerRecord("cccc3333-uuid", "nil_csuq_nick", 50, 70, 6, 75, "")
	rNilCsuq.Metric.EvalDate = time.Date(2026, 11, 28, 15, 0, 0, 0, time.UTC)
	csuqCopy := rNilCsuq.Metric.CSUQScores.(dto.CSUQScoresPayloadDTO)
	csuqCopy.Subscales.Overall = nil
	rNilCsuq.Metric.CSUQScores = csuqCopy

	orderedReport := BuildPaperStatisticsReport([]AdminLearnerEvaluationRecord{rNilCsuq, rLate, rEarly}, cohort)
	if len(orderedReport.ParticipantPoints) != 3 {
		t.Fatalf("expected 3 ParticipantPoints, got %d", len(orderedReport.ParticipantPoints))
	}
	if orderedReport.ParticipantPoints[0].ParticipantCode != "P-AAAA1111" ||
		orderedReport.ParticipantPoints[1].ParticipantCode != "P-BBBB2222" ||
		orderedReport.ParticipantPoints[2].ParticipantCode != "P-CCCC3333" {
		t.Fatalf("expected chronological ASC order P-AAAA1111, P-BBBB2222, P-CCCC3333, got %+v", orderedReport.ParticipantPoints)
	}
	if orderedReport.ParticipantPoints[0].CSUQOverall == nil || orderedReport.ParticipantPoints[2].CSUQOverall != nil {
		t.Fatalf("expected nil CSUQOverall preserved safely for null respondent: %+v", orderedReport.ParticipantPoints[2])
	}
	rawOrdered, _ := json.Marshal(orderedReport)
	if strings.Contains(string(rawOrdered), "early_nick") || strings.Contains(string(rawOrdered), "aaaa1111-uuid") {
		t.Fatalf("ParticipantPoints leaked nickname or raw userID: %s", string(rawOrdered))
	}

	// Verify clean JSON serialization (zero NaN / Inf) and non-null empty array even on n=0 empty dataset
	emptyReport := BuildPaperStatisticsReport(nil, dto.EvaluationCohortFunnelDTO{})
	if emptyReport.ParticipantPoints == nil || len(emptyReport.ParticipantPoints) != 0 {
		t.Fatalf("expected initialized non-nil empty ParticipantPoints slice on n=0, got %+v", emptyReport.ParticipantPoints)
	}
	for _, payload := range []any{report, emptyReport} {
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		if strings.Contains(string(raw), "NaN") || strings.Contains(string(raw), "Inf") {
			t.Fatalf("serialized JSON contains NaN/Inf: %s", string(raw))
		}
	}

	// n = 30 clears pilot flag
	cohort30 := make([]AdminLearnerEvaluationRecord, 0, 30)
	for i := 0; i < 30; i++ {
		cohort30 = append(cohort30, makeAdminLearnerRecord("u30-uuid", "user", 50+float64(i%5), 80+float64(i%6), 6+(i%2), 80+(i%4)*5, ""))
	}
	full30 := BuildPaperStatisticsReport(cohort30, dto.EvaluationCohortFunnelDTO{QuestionnaireRespondents: 30})
	if full30.IsPilotSample {
		t.Fatalf("expected IsPilotSample=false when n=30")
	}
}
