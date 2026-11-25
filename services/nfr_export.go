package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"cybersim/dto"
)

// ValidateCreateNFRRunRequest checks source enum and benchmark bounds on snapshot/import payloads.
func ValidateCreateNFRRunRequest(req *dto.CreateNFRRunRequest) error {
	if req == nil {
		return fmt.Errorf("request body is required")
	}
	src := strings.TrimSpace(req.Source)
	switch src {
	case "", "live_snapshot", "in_app_probe", "cli_benchmark", "imported_json":
	default:
		return fmt.Errorf("unsupported run source %q", src)
	}
	if len([]rune(req.RunLabel)) > 120 {
		return fmt.Errorf("run_label must be <= 120 characters")
	}
	if len([]rune(req.Environment)) > 64 {
		return fmt.Errorf("environment must be <= 64 characters")
	}
	if req.Scalability != nil {
		if len(req.Scalability.Levels) == 0 {
			return fmt.Errorf("scalability.levels must contain at least one level when provided")
		}
		for i, lv := range req.Scalability.Levels {
			if lv.Concurrency <= 0 || lv.Concurrency > 10000 {
				return fmt.Errorf("scalability.levels[%d].concurrency out of range", i)
			}
			if lv.TotalRequests <= 0 || lv.SuccessCount < 0 || lv.ErrorCount < 0 {
				return fmt.Errorf("scalability.levels[%d] counts invalid", i)
			}
			if lv.ErrorRatePct < 0 || lv.ErrorRatePct > 100 {
				return fmt.Errorf("scalability.levels[%d].error_rate_pct must be between 0 and 100", i)
			}
			if lv.ThroughputRPS < 0 || lv.MeanMs < 0 || lv.P50Ms < 0 || lv.P95Ms < 0 || lv.P99Ms < 0 || lv.MaxMs < 0 {
				return fmt.Errorf("scalability.levels[%d] latency/throughput cannot be negative", i)
			}
		}
	}
	if m := req.MaintainabilityReport; m != nil {
		if m.GoCoveragePct != nil && (*m.GoCoveragePct < 0 || *m.GoCoveragePct > 100) {
			return fmt.Errorf("maintainability_report.go_coverage_pct must be between 0 and 100")
		}
		for name, cnt := range map[string]*int{
			"govulncheck_high_critical": m.GovulncheckHighCritical,
			"npm_audit_high_critical":   m.NpmAuditHighCritical,
			"gitleaks_findings":         m.GitleaksFindings,
		} {
			if cnt != nil && *cnt < 0 {
				return fmt.Errorf("maintainability_report.%s cannot be negative", name)
			}
		}
	}
	return nil
}

// BuildNFRExportCSV generates a UTF-8 BOM RFC 4180 CSV dataset with formula-injection neutralization.
func BuildNFRExportCSV(report dto.NFRReportResponse) ([]byte, error) {
	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(&buf)
	headers := []string{
		"section",
		"run_id",
		"run_label",
		"environment",
		"source",
		"rubric_version",
		"evaluated_at",
		"dimension_id",
		"iso_characteristic",
		"check_key",
		"measured",
		"unit",
		"target",
		"status",
		"summary_metric",
		"evidence_source",
	}
	if err := w.Write(headers); err != nil {
		return nil, err
	}

	runIDStr := "live"
	if report.RunID != nil && *report.RunID != "" {
		runIDStr = *report.RunID
	}
	evalTime := report.EvaluatedAt.UTC().Format(time.RFC3339)
	safeLabel := NeutralizeCSVFormula(report.RunLabel)
	safeEnv := NeutralizeCSVFormula(report.Environment)

	// 1. Summary row
	if err := w.Write([]string{
		"summary",
		runIDStr,
		safeLabel,
		safeEnv,
		report.Source,
		report.RubricVersion,
		evalTime,
		"overall",
		"ISO/IEC 25010 Rollup",
		"measured_pass_rate",
		fmt.Sprintf("%.2f", report.Summary.MeasuredPassRate),
		"%",
		fmt.Sprintf("%d pass / %d warn / %d fail / %d pending", report.Summary.PassedCount, report.Summary.WarnCount, report.Summary.FailCount, report.Summary.NotMeasuredCount),
		report.Summary.OverallStatus,
		NeutralizeCSVFormula(report.Notes),
		"nfr_evaluator",
	}); err != nil {
		return nil, err
	}

	// 2. Dimension checks
	for _, dim := range report.Dimensions {
		if len(dim.Checks) == 0 {
			if err := w.Write([]string{
				"dimension", runIDStr, safeLabel, safeEnv, report.Source, report.RubricVersion, evalTime,
				dim.ID, NeutralizeCSVFormula(dim.ISOCharacteristic), "summary",
				NeutralizeCSVFormula(dim.SummaryMetric), "", NeutralizeCSVFormula(dim.TargetExpression),
				dim.Status, NeutralizeCSVFormula(dim.SummaryMetric), NeutralizeCSVFormula(dim.EvidenceSource),
			}); err != nil {
				return nil, err
			}
			continue
		}
		for _, chk := range dim.Checks {
			if err := w.Write([]string{
				"dimension_check",
				runIDStr,
				safeLabel,
				safeEnv,
				report.Source,
				report.RubricVersion,
				evalTime,
				dim.ID,
				NeutralizeCSVFormula(dim.ISOCharacteristic),
				chk.Key,
				NeutralizeCSVFormula(chk.Measured),
				chk.Unit,
				NeutralizeCSVFormula(chk.Target),
				chk.Status,
				NeutralizeCSVFormula(dim.SummaryMetric),
				NeutralizeCSVFormula(dim.EvidenceSource),
			}); err != nil {
				return nil, err
			}
		}
	}

	// 3. Scalability curve levels (when present)
	if report.Scalability != nil {
		for _, lv := range report.Scalability.Levels {
			if err := w.Write([]string{
				"scalability_level",
				runIDStr,
				safeLabel,
				safeEnv,
				report.Scalability.Mode,
				report.RubricVersion,
				evalTime,
				"scalability",
				"Performance Efficiency (Capacity)",
				fmt.Sprintf("concurrency_%d", lv.Concurrency),
				fmt.Sprintf("rps=%.2f;p50=%.2f;p95=%.2f;p99=%.2f;err=%.2f%%", lv.ThroughputRPS, lv.P50Ms, lv.P95Ms, lv.P99Ms, lv.ErrorRatePct),
				"rps/ms/%",
				fmt.Sprintf("%d requests", lv.TotalRequests),
				"measured",
				fmt.Sprintf("%d workers", lv.Concurrency),
				"scalability_benchmark",
			}); err != nil {
				return nil, err
			}
		}
	}

	// 4. Normalized runtime endpoints
	for _, ep := range report.Runtime.Endpoints {
		if err := w.Write([]string{
			"runtime_endpoint",
			runIDStr,
			safeLabel,
			safeEnv,
			report.Source,
			report.RubricVersion,
			evalTime,
			"performance",
			ep.Category,
			NeutralizeCSVFormula(ep.Route),
			fmt.Sprintf("n=%d;p50=%s;p95=%s;p99=%s;err=%d", ep.Count, fmtPtrMs(ep.P50Ms), fmtPtrMs(ep.P95Ms), fmtPtrMs(ep.P99Ms), ep.ErrorCount),
			"ms",
			"normalized_route",
			"measured",
			ep.Category,
			"http_ring_buffer",
		}); err != nil {
			return nil, err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
