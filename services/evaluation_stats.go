package services

import (
	"math"
	"sort"
	"strings"
	"time"

	"cybersim/dto"
)

const (
	statStatusOK            = "ok"
	statStatusInsufficientN = "insufficient_n"
	statStatusZeroVariance  = "zero_variance"
)

func safePtr4(val float64) *float64 {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return nil
	}
	v := math.Round(val*10000) / 10000
	return &v
}

func safePtr2(val float64) *float64 {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return nil
	}
	v := roundEval2(val)
	return &v
}

func betacf(a, b, x float64) float64 {
	const maxIter = 200
	const eps = 3.0e-12
	const fpMin = 1.0e-30

	qab := a + b
	qap := a + 1.0
	qam := a - 1.0
	c := 1.0
	d := 1.0 - qab*x/qap
	if math.Abs(d) < fpMin {
		d = fpMin
	}
	d = 1.0 / d
	h := d

	for m := 1; m <= maxIter; m++ {
		mf := float64(m)
		m2 := 2.0 * mf
		aa := mf * (b - mf) * x / ((qam + m2) * (a + m2))
		d = 1.0 + aa*d
		if math.Abs(d) < fpMin {
			d = fpMin
		}
		c = 1.0 + aa/c
		if math.Abs(c) < fpMin {
			c = fpMin
		}
		d = 1.0 / d
		h *= d * c

		aa = -(a + mf) * (qab + mf) * x / ((a + m2) * (qap + m2))
		d = 1.0 + aa*d
		if math.Abs(d) < fpMin {
			d = fpMin
		}
		c = 1.0 + aa/c
		if math.Abs(c) < fpMin {
			c = fpMin
		}
		d = 1.0 / d
		del := d * c
		h *= del
		if math.Abs(del-1.0) < eps {
			break
		}
	}
	return h
}

// regularizedIncompleteBeta evaluates I_x(a, b) accurately via Lentz continued fraction.
func regularizedIncompleteBeta(a, b, x float64) float64 {
	if x <= 0.0 {
		return 0.0
	}
	if x >= 1.0 {
		return 1.0
	}
	lna, _ := math.Lgamma(a)
	lnb, _ := math.Lgamma(b)
	lnab, _ := math.Lgamma(a + b)
	bt := math.Exp(lnab - lna - lnb + a*math.Log(x) + b*math.Log(1.0-x))
	if x < (a+1.0)/(a+b+2.0) {
		return bt * betacf(a, b, x) / a
	}
	return 1.0 - bt*betacf(b, a, 1.0-x)/b
}

func twoTailedTPValue(absT float64, df int) float64 {
	if df <= 0 || math.IsNaN(absT) {
		return 1.0
	}
	t := math.Abs(absT)
	if t == 0 {
		return 1.0
	}
	if math.IsInf(t, 1) {
		return 0.0
	}
	dfF := float64(df)
	x := dfF / (dfF + t*t)
	p := regularizedIncompleteBeta(dfF/2.0, 0.5, x)
	if p < 0 {
		return 0
	}
	if p > 1 {
		return 1
	}
	return p
}

func studentTCritical95(df int) float64 {
	if df <= 0 {
		return 0
	}
	low, high := 0.0, 1000.0
	for i := 0; i < 80; i++ {
		mid := 0.5 * (low + high)
		if twoTailedTPValue(mid, df) > 0.05 {
			low = mid
		} else {
			high = mid
		}
		if high-low < 1e-10 {
			break
		}
	}
	return 0.5 * (low + high)
}

// quantileR7 matches R type=7 / NumPy linear percentile interpolation on a sorted slice.
func quantileR7(sorted []float64, p float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return sorted[0]
	}
	h := float64(n-1) * p
	lo := int(math.Floor(h))
	hi := int(math.Ceil(h))
	frac := h - float64(lo)
	return sorted[lo] + frac*(sorted[hi]-sorted[lo])
}

func bandCSUQ(mean float64) string {
	switch {
	case mean >= 6.00:
		return "excellent"
	case mean >= 5.00:
		return "good"
	case mean >= 4.00:
		return "moderate"
	default:
		return "needs_improvement"
	}
}

func bandTLX(mean float64) string {
	switch {
	case mean < 30.0:
		return "low"
	case mean < 50.0:
		return "moderate"
	case mean < 80.0:
		return "high"
	default:
		return "very_high"
	}
}

func bandNone(_ float64) string {
	return "neutral"
}

// ComputeDescriptiveStat computes n, mean, sample SD, Student-t 95% CI, R-7 quartiles, min/max, and band.
func ComputeDescriptiveStat(key, group string, raw []float64, bandFn func(float64) string) dto.DescriptiveStatDTO {
	vals := make([]float64, 0, len(raw))
	for _, v := range raw {
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			vals = append(vals, v)
		}
	}
	n := len(vals)
	if n == 0 {
		return dto.DescriptiveStatDTO{
			Key:    key,
			Group:  group,
			N:      0,
			Band:   "insufficient_n",
			Status: "insufficient_n",
		}
	}

	sorted := append([]float64(nil), vals...)
	sort.Float64s(sorted)

	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	mean := sum / float64(n)
	q1 := quantileR7(sorted, 0.25)
	median := quantileR7(sorted, 0.50)
	q3 := quantileR7(sorted, 0.75)

	stat := dto.DescriptiveStatDTO{
		Key:    key,
		Group:  group,
		N:      n,
		Mean:   safePtr2(mean),
		Median: safePtr2(median),
		Q1:     safePtr2(q1),
		Q3:     safePtr2(q3),
		IQR:    safePtr2(q3 - q1),
		Min:    safePtr2(sorted[0]),
		Max:    safePtr2(sorted[n-1]),
		Band:   bandFn(mean),
		Status: "ok",
	}

	if n < 2 {
		stat.Status = "insufficient_n"
		return stat
	}

	sq := 0.0
	for _, v := range vals {
		d := v - mean
		sq += d * d
	}
	sd := math.Sqrt(sq / float64(n-1))
	se := sd / math.Sqrt(float64(n))
	tCrit := studentTCritical95(n - 1)
	margin := tCrit * se

	stat.SD = safePtr2(sd)
	stat.SE = safePtr2(se)
	stat.CILow = safePtr2(mean - margin)
	stat.CIHigh = safePtr2(mean + margin)
	return stat
}

func sampleVariance(vals []float64) float64 {
	n := len(vals)
	if n < 2 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	mean := sum / float64(n)
	sq := 0.0
	for _, v := range vals {
		d := v - mean
		sq += d * d
	}
	return sq / float64(n-1)
}

func bandCronbachAlpha(alpha float64) string {
	switch {
	case alpha >= 0.90:
		return "excellent"
	case alpha >= 0.80:
		return "good"
	case alpha >= 0.70:
		return "acceptable"
	case alpha >= 0.60:
		return "questionable"
	case alpha >= 0.50:
		return "poor"
	default:
		return "unacceptable"
	}
}

// ComputeCronbachAlpha computes listwise complete-case Cronbach's alpha over k items.
func ComputeCronbachAlpha(key string, itemCount int, completeRows [][]float64, totalRows int) dto.CronbachAlphaDTO {
	completeN := len(completeRows)
	excluded := totalRows - completeN
	if excluded < 0 {
		excluded = 0
	}
	res := dto.CronbachAlphaDTO{
		Key:             key,
		ItemCount:       itemCount,
		CompleteCasesN:  completeN,
		ExcludedNACases: excluded,
	}
	if completeN < 2 || itemCount < 2 {
		res.Band = "insufficient_n"
		res.Status = "insufficient_n"
		return res
	}

	sumItemVars := 0.0
	rowTotals := make([]float64, completeN)
	for j := 0; j < itemCount; j++ {
		col := make([]float64, completeN)
		for i := 0; i < completeN; i++ {
			col[i] = completeRows[i][j]
			rowTotals[i] += completeRows[i][j]
		}
		sumItemVars += sampleVariance(col)
	}

	totalVar := sampleVariance(rowTotals)
	if totalVar <= 1e-12 {
		res.Band = "zero_variance"
		res.Status = "zero_variance"
		return res
	}

	k := float64(itemCount)
	alpha := (k / (k - 1.0)) * (1.0 - sumItemVars/totalVar)
	res.Alpha = safePtr4(alpha)
	res.Band = bandCronbachAlpha(alpha)
	res.Status = "ok"
	return res
}

func bandCohensD(d float64) string {
	a := math.Abs(d)
	switch {
	case a < 0.20:
		return "negligible"
	case a < 0.50:
		return "small"
	case a < 0.80:
		return "medium"
	default:
		return "large"
	}
}

func exactWilcoxonTwoTailedP(wMin int, m int) float64 {
	maxSum := m * (m + 1) / 2
	ways := make([]uint64, maxSum+1)
	ways[0] = 1
	for r := 1; r <= m; r++ {
		for s := maxSum; s >= r; s-- {
			ways[s] += ways[s-r]
		}
	}
	if wMin > maxSum {
		wMin = maxSum
	}
	var tail uint64
	for s := 0; s <= wMin; s++ {
		tail += ways[s]
	}
	totalPerms := math.Ldexp(1.0, m)
	p := 2.0 * float64(tail) / totalPerms
	if p > 1.0 {
		return 1.0
	}
	return p
}

// ComputePairedInference runs both Paired t-test (+ Cohen's d_z) and Wilcoxon signed-rank (+ r_rb).
func ComputePairedInference(pre, post []float64) (dto.PairedTTestDTO, dto.WilcoxonTestDTO) {
	n := len(pre)
	if len(post) < n {
		n = len(post)
	}
	diffs := make([]float64, n)
	for i := 0; i < n; i++ {
		diffs[i] = post[i] - pre[i]
	}

	// 1. Paired t-test
	tRes := dto.PairedTTestDTO{N: n}
	if n < 3 {
		tRes.EffectBand = "insufficient_n"
		tRes.Status = "insufficient_n"
	} else {
		sumD := 0.0
		for _, d := range diffs {
			sumD += d
		}
		meanD := sumD / float64(n)
		sdD := math.Sqrt(sampleVariance(diffs))
		tRes.DF = n - 1
		tRes.MeanDiff = safePtr2(meanD)
		tRes.SDDiff = safePtr2(sdD)
		if sdD <= 1e-12 {
			tRes.EffectBand = "zero_variance"
			tRes.Status = "zero_variance"
		} else {
			se := sdD / math.Sqrt(float64(n))
			tStat := meanD / se
			pVal := twoTailedTPValue(tStat, n-1)
			margin := studentTCritical95(n-1) * se
			dz := meanD / sdD

			tRes.TStat = safePtr4(tStat)
			tRes.PValue = safePtr4(pVal)
			tRes.CILow = safePtr2(meanD - margin)
			tRes.CIHigh = safePtr2(meanD + margin)
			tRes.CohensDz = safePtr4(dz)
			tRes.EffectBand = bandCohensD(dz)
			tRes.Status = "ok"
		}
	}

	// 2. Wilcoxon signed-rank test (excluding zero differences)
	type signedAbs struct {
		abs  float64
		sign int
	}
	nonZero := make([]signedAbs, 0, n)
	zeroCount := 0
	for _, d := range diffs {
		if math.Abs(d) <= 1e-12 {
			zeroCount++
			continue
		}
		sign := 1
		if d < 0 {
			sign = -1
		}
		nonZero = append(nonZero, signedAbs{abs: math.Abs(d), sign: sign})
	}

	m := len(nonZero)
	wRes := dto.WilcoxonTestDTO{
		N:             n,
		EffectiveN:    m,
		ZeroDiffCount: zeroCount,
	}
	if n < 5 || m < 5 {
		wRes.Method = "insufficient_n"
		wRes.Status = "insufficient_n"
		return tRes, wRes
	}

	absVals := make([]float64, m)
	for i, item := range nonZero {
		absVals[i] = item.abs
	}
	ranks, tieTerm, hasTies := averageRanksAndTieSum(absVals)

	wPlus := 0.0
	wMinus := 0.0
	for k := 0; k < m; k++ {
		if nonZero[k].sign > 0 {
			wPlus += ranks[k]
		} else {
			wMinus += ranks[k]
		}
	}

	wStat := math.Min(wPlus, wMinus)
	rankBiserial := (wPlus - wMinus) / (wPlus + wMinus)
	wRes.WPlus = safePtr2(wPlus)
	wRes.WMinus = safePtr2(wMinus)
	wRes.WStat = safePtr2(wStat)
	wRes.RankBiserial = safePtr4(rankBiserial)

	var pVal float64
	if m <= 20 && !hasTies {
		wRes.Method = "exact"
		pVal = exactWilcoxonTwoTailedP(int(math.Round(wStat)), m)
	} else {
		wRes.Method = "normal_approx"
		mf := float64(m)
		mu := mf * (mf + 1.0) / 4.0
		sigmaSq := (mf*(mf+1.0)*(2.0*mf+1.0))/24.0 - tieTerm/48.0
		if sigmaSq <= 1e-12 {
			pVal = 1.0
		} else {
			diff := math.Abs(wStat-mu) - 0.5
			if diff <= 0 {
				pVal = 1.0
			} else {
				z := diff / math.Sqrt(sigmaSq)
				pVal = math.Erfc(z / math.Sqrt2)
			}
		}
	}
	wRes.PValue = safePtr4(pVal)
	wRes.Status = statStatusOK
	return tRes, wRes
}

func averageRanksAndTieSum(vals []float64) ([]float64, float64, bool) {
	n := len(vals)
	type idxVal struct {
		idx int
		val float64
	}
	items := make([]idxVal, n)
	for i, v := range vals {
		items[i] = idxVal{idx: i, val: v}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].val < items[j].val
	})
	ranks := make([]float64, n)
	tieTerm := 0.0
	hasTies := false
	for i := 0; i < n; {
		j := i + 1
		for j < n && math.Abs(items[j].val-items[i].val) <= 1e-12 {
			j++
		}
		count := j - i
		if count > 1 {
			hasTies = true
			cf := float64(count)
			tieTerm += cf*cf*cf - cf
		}
		avg := 0.5 * (float64(i+1) + float64(j))
		for k := i; k < j; k++ {
			ranks[items[k].idx] = avg
		}
		i = j
	}
	return ranks, tieTerm, hasTies
}

func averageRanks(vals []float64) []float64 {
	ranks, _, _ := averageRanksAndTieSum(vals)
	return ranks
}

func bandSpearman(rho float64) string {
	a := math.Abs(rho)
	switch {
	case a < 0.10:
		return "negligible"
	case a < 0.30:
		return "weak"
	case a < 0.50:
		return "moderate"
	default:
		return "strong"
	}
}

// ComputeSpearmanCorrelation computes tie-safe Spearman's rho and two-tailed Student-t p-value.
func ComputeSpearmanCorrelation(key string, x, y []float64) dto.SpearmanCorrelationDTO {
	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	res := dto.SpearmanCorrelationDTO{Key: key, N: n}
	if n < 3 {
		res.Band = "insufficient_n"
		res.Status = "insufficient_n"
		return res
	}

	rx := averageRanks(x[:n])
	ry := averageRanks(y[:n])

	meanX := 0.0
	meanY := 0.0
	for i := 0; i < n; i++ {
		meanX += rx[i]
		meanY += ry[i]
	}
	meanX /= float64(n)
	meanY /= float64(n)

	num := 0.0
	denX := 0.0
	denY := 0.0
	for i := 0; i < n; i++ {
		dx := rx[i] - meanX
		dy := ry[i] - meanY
		num += dx * dy
		denX += dx * dx
		denY += dy * dy
	}

	den := math.Sqrt(denX * denY)
	if den <= 1e-12 {
		res.Band = "zero_variance"
		res.Status = "zero_variance"
		return res
	}

	rho := num / den
	if rho > 1.0 {
		rho = 1.0
	}
	if rho < -1.0 {
		rho = -1.0
	}

	var pVal float64
	if math.Abs(rho) >= 1.0-1e-12 {
		pVal = 0.0
	} else {
		tStat := math.Abs(rho) * math.Sqrt(float64(n-2)/(1.0-rho*rho))
		pVal = twoTailedTPValue(tStat, n-2)
	}

	res.Rho = safePtr4(rho)
	res.PValue = safePtr4(pVal)
	res.Band = bandSpearman(rho)
	res.Status = "ok"
	return res
}

func extractCompleteItemMatrix(summaries []dto.EvaluationMetricSummaryDTO, keys []string) [][]float64 {
	rows := make([][]float64, 0, len(summaries))
	for _, s := range summaries {
		if s.CSUQScores.Items == nil {
			continue
		}
		row := make([]float64, 0, len(keys))
		complete := true
		for _, k := range keys {
			ptr := s.CSUQScores.Items[k]
			if ptr == nil {
				complete = false
				break
			}
			row = append(row, float64(*ptr))
		}
		if complete {
			rows = append(rows, row)
		}
	}
	return rows
}

func csuqItemSubscaleGroup(key string) string {
	switch key {
	case "q1", "q2", "q3", "q4", "q5", "q6":
		return "SYSUSE"
	case "q7", "q8", "q9", "q10", "q11", "q12":
		return "INFOQUAL"
	case "q13", "q14", "q15":
		return "INTERQUAL"
	default:
		return "OVERALL"
	}
}

type pairedFloatSeries struct {
	x []float64
	y []float64
}

func (p *pairedFloatSeries) append(xv, yv float64) {
	p.x = append(p.x, xv)
	p.y = append(p.y, yv)
}

type cohortMetricSeries struct {
	summaries        []dto.EvaluationMetricSummaryDTO
	sysuseVals       []float64
	infoqualVals     []float64
	interqualVals    []float64
	csuqOverallVals  []float64
	itemBuckets      map[string][]float64
	mentalVals       []float64
	physicalVals     []float64
	temporalVals     []float64
	perfRatingVals   []float64
	perfWorkloadVals []float64
	effortVals       []float64
	frustrationVals  []float64
	tlxOverallVals   []float64
	tlxCompleteRows  [][]float64
	preVals          []float64
	postVals         []float64
	impVals          []float64
	diffVals         []float64
	qualCount        int
	gainVsCSUQ       pairedFloatSeries
	gainVsTLX        pairedFloatSeries
	csuqVsTLX        pairedFloatSeries
}

func collectCohortMetricSeries(records []AdminLearnerEvaluationRecord) cohortMetricSeries {
	n := len(records)
	series := cohortMetricSeries{
		summaries:        make([]dto.EvaluationMetricSummaryDTO, 0, n),
		sysuseVals:       make([]float64, 0, n),
		infoqualVals:     make([]float64, 0, n),
		interqualVals:    make([]float64, 0, n),
		csuqOverallVals:  make([]float64, 0, n),
		itemBuckets:      make(map[string][]float64, len(CSUQItemKeys)),
		mentalVals:       make([]float64, 0, n),
		physicalVals:     make([]float64, 0, n),
		temporalVals:     make([]float64, 0, n),
		perfRatingVals:   make([]float64, 0, n),
		perfWorkloadVals: make([]float64, 0, n),
		effortVals:       make([]float64, 0, n),
		frustrationVals:  make([]float64, 0, n),
		tlxOverallVals:   make([]float64, 0, n),
		tlxCompleteRows:  make([][]float64, 0, n),
		preVals:          make([]float64, 0, n),
		postVals:         make([]float64, 0, n),
		impVals:          make([]float64, 0, n),
		diffVals:         make([]float64, 0, n),
	}

	for _, rec := range records {
		s := EvaluationMetricToSummaryDTO(rec.Metric)
		series.summaries = append(series.summaries, s)

		series.preVals = append(series.preVals, s.PreScore)
		series.postVals = append(series.postVals, s.PostScore)
		series.impVals = append(series.impVals, s.ImprovementPct)
		diff := s.PostScore - s.PreScore
		series.diffVals = append(series.diffVals, diff)

		if strings.TrimSpace(s.Notes) != "" ||
			strings.TrimSpace(s.QualitativeFeedback.MostImpressive) != "" ||
			strings.TrimSpace(s.QualitativeFeedback.AreasToImprove) != "" ||
			strings.TrimSpace(s.QualitativeFeedback.AdditionalComments) != "" {
			series.qualCount++
		}

		if s.CSUQScores.Subscales.SysUse != nil {
			series.sysuseVals = append(series.sysuseVals, *s.CSUQScores.Subscales.SysUse)
		}
		if s.CSUQScores.Subscales.InfoQual != nil {
			series.infoqualVals = append(series.infoqualVals, *s.CSUQScores.Subscales.InfoQual)
		}
		if s.CSUQScores.Subscales.InterQual != nil {
			series.interqualVals = append(series.interqualVals, *s.CSUQScores.Subscales.InterQual)
		}
		if s.CSUQScores.Subscales.Overall != nil {
			series.csuqOverallVals = append(series.csuqOverallVals, *s.CSUQScores.Subscales.Overall)
			series.gainVsCSUQ.append(diff, *s.CSUQScores.Subscales.Overall)
		}
		if s.CSUQScores.Items != nil {
			for _, k := range CSUQItemKeys {
				if v := s.CSUQScores.Items[k]; v != nil {
					series.itemBuckets[k] = append(series.itemBuckets[k], float64(*v))
				}
			}
		}

		if s.NASATLXScores.SchemaVersion != "" {
			d := s.NASATLXScores.Dimensions
			series.mentalVals = append(series.mentalVals, float64(d.MentalDemand))
			series.physicalVals = append(series.physicalVals, float64(d.PhysicalDemand))
			series.temporalVals = append(series.temporalVals, float64(d.TemporalDemand))
			series.perfRatingVals = append(series.perfRatingVals, float64(d.PerformanceRating))
			series.perfWorkloadVals = append(series.perfWorkloadVals, float64(d.PerformanceWorkload))
			series.effortVals = append(series.effortVals, float64(d.Effort))
			series.frustrationVals = append(series.frustrationVals, float64(d.Frustration))
			series.tlxOverallVals = append(series.tlxOverallVals, s.NASATLXScores.OverallWorkload)
			series.tlxCompleteRows = append(series.tlxCompleteRows, []float64{
				float64(d.MentalDemand),
				float64(d.PhysicalDemand),
				float64(d.TemporalDemand),
				float64(d.PerformanceWorkload),
				float64(d.Effort),
				float64(d.Frustration),
			})

			series.gainVsTLX.append(diff, s.NASATLXScores.OverallWorkload)
			if s.CSUQScores.Subscales.Overall != nil {
				series.csuqVsTLX.append(*s.CSUQScores.Subscales.Overall, s.NASATLXScores.OverallWorkload)
			}
		}
	}
	return series
}

// BuildPaperStatisticsReport computes the full IEEE/thesis statistical suite from learner evaluations.
func BuildPaperStatisticsReport(
	records []AdminLearnerEvaluationRecord,
	cohort dto.EvaluationCohortFunnelDTO,
) dto.AdminPaperStatisticsResponse {
	series := collectCohortMetricSeries(records)

	csuqSubscales := []dto.DescriptiveStatDTO{
		ComputeDescriptiveStat("sysuse", "CSUQ", series.sysuseVals, bandCSUQ),
		ComputeDescriptiveStat("infoqual", "CSUQ", series.infoqualVals, bandCSUQ),
		ComputeDescriptiveStat("interqual", "CSUQ", series.interqualVals, bandCSUQ),
		ComputeDescriptiveStat("overall", "CSUQ", series.csuqOverallVals, bandCSUQ),
	}

	csuqItems := make([]dto.DescriptiveStatDTO, 0, len(CSUQItemKeys))
	for _, k := range CSUQItemKeys {
		csuqItems = append(csuqItems, ComputeDescriptiveStat(k, csuqItemSubscaleGroup(k), series.itemBuckets[k], bandCSUQ))
	}

	nasaTLX := []dto.DescriptiveStatDTO{
		ComputeDescriptiveStat("mental_demand", "Workload", series.mentalVals, bandTLX),
		ComputeDescriptiveStat("physical_demand", "Workload", series.physicalVals, bandTLX),
		ComputeDescriptiveStat("temporal_demand", "Workload", series.temporalVals, bandTLX),
		ComputeDescriptiveStat("performance_rating", "Self-Rating", series.perfRatingVals, bandNone),
		ComputeDescriptiveStat("performance_workload", "Workload", series.perfWorkloadVals, bandTLX),
		ComputeDescriptiveStat("effort", "Workload", series.effortVals, bandTLX),
		ComputeDescriptiveStat("frustration", "Workload", series.frustrationVals, bandTLX),
		ComputeDescriptiveStat("overall_workload", "Workload", series.tlxOverallVals, bandTLX),
	}

	totalRows := len(series.summaries)
	reliability := []dto.CronbachAlphaDTO{
		ComputeCronbachAlpha("csuq_overall", len(CSUQItemKeys), extractCompleteItemMatrix(series.summaries, CSUQItemKeys), totalRows),
		ComputeCronbachAlpha("csuq_sysuse", len(csuqSysUseKeys), extractCompleteItemMatrix(series.summaries, csuqSysUseKeys), totalRows),
		ComputeCronbachAlpha("csuq_infoqual", len(csuqInfoQualKeys), extractCompleteItemMatrix(series.summaries, csuqInfoQualKeys), totalRows),
		ComputeCronbachAlpha("csuq_interqual", len(csuqInterQualKeys), extractCompleteItemMatrix(series.summaries, csuqInterQualKeys), totalRows),
		ComputeCronbachAlpha("nasa_tlx_overall", 6, series.tlxCompleteRows, totalRows),
	}

	pairedT, wilcoxon := ComputePairedInference(series.preVals, series.postVals)
	learningGain := dto.LearningGainStatisticsDTO{
		Pre:         ComputeDescriptiveStat("pre_score", "Exam", series.preVals, bandNone),
		Post:        ComputeDescriptiveStat("post_score", "Exam", series.postVals, bandNone),
		Difference:  ComputeDescriptiveStat("improvement_pct", "Exam", series.diffVals, bandNone),
		PairedTTest: pairedT,
		Wilcoxon:    wilcoxon,
	}

	correlations := []dto.SpearmanCorrelationDTO{
		ComputeSpearmanCorrelation("learning_gain_vs_csuq_overall", series.gainVsCSUQ.x, series.gainVsCSUQ.y),
		ComputeSpearmanCorrelation("learning_gain_vs_nasa_tlx_overall", series.gainVsTLX.x, series.gainVsTLX.y),
		ComputeSpearmanCorrelation("csuq_overall_vs_nasa_tlx_overall", series.csuqVsTLX.x, series.csuqVsTLX.y),
	}

	return dto.AdminPaperStatisticsResponse{
		GeneratedAt:          time.Now().UTC(),
		ProtocolVersion:      "protocol-v1",
		CSUQSchemaVersion:    CSUQSchemaVersion,
		NASATLXSchemaVersion: NASATLXSchemaVersion,
		IsPilotSample:        totalRows < 30,
		Cohort:               cohort,
		CSUQSubscales:        csuqSubscales,
		CSUQItems:            csuqItems,
		NASATLX:              nasaTLX,
		Reliability:          reliability,
		LearningGain:         learningGain,
		Correlations:         correlations,
	}
}
