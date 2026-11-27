package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cybersim/dto"
	"cybersim/models"
	"cybersim/services"

	"github.com/gin-gonic/gin"
)

func TestEvaluateStage2TheoryGate(t *testing.T) {
	allFive := []string{"domain1", "domain2", "domain3", "domain4", "domain5"}
	fourOfFive := []string{"domain1", "domain2", "domain3", "domain4"}

	tests := []struct {
		name               string
		bypassEnv          string
		role               string
		completedDomains   []string
		completedSubtopics []string
		wantBlocked        bool
		wantCompletedCount int
		wantTotalDomains   int
	}{
		{
			name:               "blocks learner with 0 completed domains",
			bypassEnv:          "",
			role:               "learner",
			completedDomains:   []string{},
			completedSubtopics: []string{},
			wantBlocked:        true,
			wantCompletedCount: 0,
			wantTotalDomains:   5,
		},
		{
			name:               "blocks learner with 4 of 5 completed domains",
			bypassEnv:          "false",
			role:               "learner",
			completedDomains:   fourOfFive,
			completedSubtopics: []string{"1.1", "1.2"},
			wantBlocked:        true,
			wantCompletedCount: 4,
			wantTotalDomains:   5,
		},
		{
			name:               "allows learner when all 5 domain essentials are completed even with 0 subtopics",
			bypassEnv:          "",
			role:               "learner",
			completedDomains:   allFive,
			completedSubtopics: []string{},
			wantBlocked:        false,
		},
		{
			name:               "allows specialist exemption even with 0 completed domains",
			bypassEnv:          "",
			role:               "specialist",
			completedDomains:   []string{},
			completedSubtopics: []string{},
			wantBlocked:        false,
		},
		{
			name:               "allows bypass when BYPASS_THEORY_GATE is true",
			bypassEnv:          "true",
			role:               "learner",
			completedDomains:   []string{},
			completedSubtopics: []string{},
			wantBlocked:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gateErr := evaluateStage2TheoryGate(tt.bypassEnv, tt.role, func() ([]string, []string) {
				return tt.completedSubtopics, tt.completedDomains
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
			if decoded.Reason != "theory_incomplete" || decoded.Code != "SIM_THEORY_INCOMPLETE" {
				t.Fatalf("expected reason=theory_incomplete code=SIM_THEORY_INCOMPLETE, got %+v", decoded)
			}
			if decoded.CompletedCount == nil || *decoded.CompletedCount != tt.wantCompletedCount {
				t.Fatalf("expected completed_count=%d, got %v", tt.wantCompletedCount, decoded.CompletedCount)
			}
			if decoded.TotalDomains == nil || *decoded.TotalDomains != tt.wantTotalDomains {
				t.Fatalf("expected total_domains=%d, got %v", tt.wantTotalDomains, decoded.TotalDomains)
			}
		})
	}
}

func TestBuildLearnSubtopicTelemetry(t *testing.T) {
	now := time.Now()
	ansA := "A"
	ansB := "B"
	corrTrue := true
	corrFalse := false

	subProgress := []models.UserTheorySubtopicProgress{
		{UserID: "u1", SubTopicID: "1.1", CompletedAt: now},
		{UserID: "u1", SubTopicID: "2.1", CompletedAt: now},
	}
	sampleRows := []models.UserTheorySampleQuestion{
		{UserID: "u1", SubTopicID: "1.1", QuestionID: "q1", OrderIndex: 1, SubmittedAnswer: &ansA, IsCorrect: &corrTrue},
		{UserID: "u1", SubTopicID: "1.1", QuestionID: "q2", OrderIndex: 2, SubmittedAnswer: &ansB, IsCorrect: &corrFalse},
		{UserID: "u1", SubTopicID: "2.1", QuestionID: "q3", OrderIndex: 1, SubmittedAnswer: &ansA, IsCorrect: &corrTrue},
		{UserID: "u1", SubTopicID: "2.1", QuestionID: "q4", OrderIndex: 2},
	}

	tel := BuildLearnSubtopicTelemetry(subProgress, sampleRows)
	if tel.CompletedSubtopicsCount != 2 || tel.TotalSubtopics != 28 {
		t.Fatalf("expected 2/28 completed subtopics, got %d/%d", tel.CompletedSubtopicsCount, tel.TotalSubtopics)
	}
	if tel.SampleQuestionsAssigned != 4 || tel.SampleQuestionsAttempted != 3 || tel.SampleQuestionsCorrect != 2 {
		t.Fatalf("unexpected sample metrics: %+v", tel)
	}
	// 2 / 3 = 66.67
	if tel.SampleAccuracyPercent != 66.67 {
		t.Fatalf("expected sample_accuracy_percent=66.67, got %v", tel.SampleAccuracyPercent)
	}
	if len(tel.DomainBreakdown) != 5 || tel.DomainBreakdown[0].DomainID != "domain1" || tel.DomainBreakdown[0].CompletedSubtopics != 1 || tel.DomainBreakdown[0].TotalSubtopics != 4 {
		t.Fatalf("unexpected domain1 breakdown: %+v", tel.DomainBreakdown)
	}
}

func TestBuildEffectiveCompletedDomains_HonorsDirectAndAutoCompletions(t *testing.T) {
	now := time.Now().Add(-24 * time.Hour)
	allRows := make([]models.UserTheoryProgress, 0, len(services.OrderedTheoryDomains))
	for _, dom := range services.OrderedTheoryDomains {
		allRows = append(allRows, models.UserTheoryProgress{
			UserID:      "learner-essentials",
			DomainID:    dom.DomainID,
			CompletedAt: now,
		})
	}

	allFive, byFive := buildEffectiveCompletedDomains(allRows)
	if len(allFive) != len(services.OrderedTheoryDomains) || len(byFive) != len(services.OrderedTheoryDomains) {
		t.Fatalf("expected all %d domains honored directly without subtopics, got %v (map len %d)",
			len(services.OrderedTheoryDomains), allFive, len(byFive))
	}

	partialRows := []models.UserTheoryProgress{
		{UserID: "learner-essentials", DomainID: "domain4", CompletedAt: now},
		{UserID: "learner-essentials", DomainID: "domain1", CompletedAt: now},
	}
	ordered, byPartial := buildEffectiveCompletedDomains(partialRows)
	if len(ordered) != 2 || ordered[0] != "domain1" || ordered[1] != "domain4" {
		t.Fatalf("expected canonical order [domain1 domain4], got %v", ordered)
	}
	if _, ok := byPartial["domain2"]; ok {
		t.Fatalf("uncompleted domain2 should not appear in effective domain map")
	}
}
