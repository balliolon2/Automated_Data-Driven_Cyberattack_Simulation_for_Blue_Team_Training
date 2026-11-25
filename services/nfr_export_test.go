package services

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"cybersim/dto"
)

func TestBuildNFRExportCSV_AndValidateCreateNFRRunRequest(t *testing.T) {
	rep := EvaluateNFR(healthyBaseInput())
	rep.RunLabel = "=MaliciousFormula"
	rep.Notes = "+cmd|'/C calc'!A0"
	rep.Runtime.Endpoints = []dto.NFREndpointLatencyDTO{
		{
			Route:      "GET /api/health",
			Category:   "non_llm_api",
			Count:      10,
			ErrorCount: 0,
			P50Ms:      floatPtrNFR(1.5),
			P95Ms:      floatPtrNFR(3.2),
			P99Ms:      floatPtrNFR(4.1),
		},
	}

	csvBytes, err := BuildNFRExportCSV(rep)
	if err != nil {
		t.Fatalf("BuildNFRExportCSV failed: %v", err)
	}
	if !bytes.HasPrefix(csvBytes, []byte{0xEF, 0xBB, 0xBF}) {
		t.Fatalf("expected UTF-8 BOM prefix")
	}
	csvStr := string(csvBytes)
	if !strings.Contains(csvStr, "'=MaliciousFormula") || !strings.Contains(csvStr, "'+cmd|'/C calc'!A0") {
		t.Fatalf("expected CSV formula neutralization on label and notes")
	}
	if !strings.Contains(csvStr, "scalability_level") || !strings.Contains(csvStr, "runtime_endpoint") {
		t.Fatalf("expected scalability_level and runtime_endpoint sections in CSV")
	}

	if err := ValidateCreateNFRRunRequest(&dto.CreateNFRRunRequest{Source: "invalid_source"}); err == nil {
		t.Fatalf("expected invalid source rejection")
	}
	if err := ValidateCreateNFRRunRequest(&dto.CreateNFRRunRequest{
		Source: "cli_benchmark",
		Scalability: &dto.NFRScalabilityBenchmarkDTO{
			Mode:       "cli_benchmark",
			ExecutedAt: time.Now().UTC(),
			Levels: []dto.NFRScalabilityLevelDTO{
				{Concurrency: 10, TotalRequests: 50, ErrorRatePct: 120},
			},
		},
	}); err == nil {
		t.Fatalf("expected out-of-range error_rate_pct rejection")
	}
	if err := ValidateCreateNFRRunRequest(&dto.CreateNFRRunRequest{
		Source: "imported_json",
		MaintainabilityReport: &dto.NFRMaintainabilityReportDTO{
			GoCoveragePct: floatPtrNFR(65.2),
		},
	}); err != nil {
		t.Fatalf("expected valid benchmark payload, got %v", err)
	}
}
