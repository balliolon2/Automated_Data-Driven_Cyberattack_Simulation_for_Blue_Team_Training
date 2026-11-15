package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cybersim/dto"

	"github.com/gin-gonic/gin"
)

func TestEvaluateStage2TheoryGate(t *testing.T) {
	allFive := []string{"domain1", "domain2", "domain3", "domain4", "domain5"}

	tests := []struct {
		name               string
		bypassEnv          string
		completedDomains   []string
		wantBlocked        bool
		wantCompletedCount int
		wantTotalDomains   int
	}{
		{
			name:               "blocks learner with 0 completed domains",
			bypassEnv:          "",
			completedDomains:   []string{},
			wantBlocked:        true,
			wantCompletedCount: 0,
			wantTotalDomains:   5,
		},
		{
			name:               "blocks learner with 4 of 5 domains",
			bypassEnv:          "false",
			completedDomains:   []string{"domain1", "domain2", "domain3", "domain4"},
			wantBlocked:        true,
			wantCompletedCount: 4,
			wantTotalDomains:   5,
		},
		{
			name:             "allows learner when all 5 domains are completed",
			bypassEnv:        "",
			completedDomains: allFive,
			wantBlocked:      false,
		},
		{
			name:             "allows bypass when BYPASS_THEORY_GATE is true",
			bypassEnv:        "true",
			completedDomains: []string{},
			wantBlocked:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gateErr := evaluateStage2TheoryGate(tt.bypassEnv, func() []string {
				return tt.completedDomains
			})

			if !tt.wantBlocked {
				if gateErr != nil {
					t.Fatalf("expected gate to allow access, got %+v", gateErr)
				}
				return
			}

			if gateErr == nil {
				t.Fatal("expected gate to block access with TheoryGateErrorResponse, got nil")
			}

			// Round-trip through Gin + JSON to verify exact HTTP 403 wire format.
			gin.SetMode(gin.TestMode)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.JSON(http.StatusForbidden, gateErr)

			if rec.Code != http.StatusForbidden {
				t.Fatalf("expected HTTP 403, got %d", rec.Code)
			}

			var decoded dto.TheoryGateErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &decoded); err != nil {
				t.Fatalf("failed to decode error JSON: %v", err)
			}
			if decoded.Reason != "theory_incomplete" {
				t.Fatalf("expected reason=theory_incomplete, got %q", decoded.Reason)
			}
			if decoded.CompletedCount == nil || *decoded.CompletedCount != tt.wantCompletedCount {
				t.Fatalf("expected completed_count=%d, got %v", tt.wantCompletedCount, decoded.CompletedCount)
			}
			if decoded.TotalDomains == nil || *decoded.TotalDomains != tt.wantTotalDomains {
				t.Fatalf("expected total_domains=%d, got %v", tt.wantTotalDomains, decoded.TotalDomains)
			}
			if len(decoded.CompletedDomains) != tt.wantCompletedCount {
				t.Fatalf("expected %d completed_domains, got %v", tt.wantCompletedCount, decoded.CompletedDomains)
			}
		})
	}
}
