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
	allTwentyEight := make([]string, 0, len(services.OrderedTheorySubtopics))
	for _, s := range services.OrderedTheorySubtopics {
		allTwentyEight = append(allTwentyEight, s.SubTopicID)
	}
	twentySeven := allTwentyEight[:len(allTwentyEight)-1]

	tests := []struct {
		name                   string
		bypassEnv              string
		role                   string
		completedSubtopics     []string
		wantBlocked            bool
		wantCompletedSubtopics int
		wantTotalSubtopics     int
	}{
		{
			name:                   "blocks learner with 0 completed subtopics",
			bypassEnv:              "",
			role:                   "learner",
			completedSubtopics:     []string{},
			wantBlocked:            true,
			wantCompletedSubtopics: 0,
			wantTotalSubtopics:     28,
		},
		{
			name:                   "blocks learner with 27 of 28 subtopics (legacy 5/5 domains insufficient)",
			bypassEnv:              "false",
			role:                   "learner",
			completedSubtopics:     twentySeven,
			wantBlocked:            true,
			wantCompletedSubtopics: 27,
			wantTotalSubtopics:     28,
		},
		{
			name:               "allows learner when all 28 subtopics are completed",
			bypassEnv:          "",
			role:               "learner",
			completedSubtopics: allTwentyEight,
			wantBlocked:        false,
		},
		{
			name:               "allows specialist exemption even with 0 completed subtopics",
			bypassEnv:          "",
			role:               "specialist",
			completedSubtopics: []string{},
			wantBlocked:        false,
		},
		{
			name:               "allows bypass when BYPASS_THEORY_GATE is true",
			bypassEnv:          "true",
			role:               "learner",
			completedSubtopics: []string{},
			wantBlocked:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gateErr := evaluateStage2TheoryGate(tt.bypassEnv, tt.role, func() ([]string, []string) {
				return tt.completedSubtopics, []string{}
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
			if decoded.CompletedSubtopicCount == nil || *decoded.CompletedSubtopicCount != tt.wantCompletedSubtopics {
				t.Fatalf("expected completed_subtopic_count=%d, got %v", tt.wantCompletedSubtopics, decoded.CompletedSubtopicCount)
			}
			if decoded.TotalSubtopics == nil || *decoded.TotalSubtopics != tt.wantTotalSubtopics {
				t.Fatalf("expected total_subtopics=%d, got %v", tt.wantTotalSubtopics, decoded.TotalSubtopics)
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

func TestBuildEffectiveCompletedDomains_ExcludesLegacyWithoutSubtopics(t *testing.T) {
	now := time.Now().Add(-24 * time.Hour)
	allLegacyRows := make([]models.UserTheoryProgress, 0, len(services.OrderedTheoryDomains))
	for _, dom := range services.OrderedTheoryDomains {
		allLegacyRows = append(allLegacyRows, models.UserTheoryProgress{
			UserID:      "legacy-t4",
			DomainID:    dom.DomainID,
			CompletedAt: now,
		})
	}

	noneDone, byNone := buildEffectiveCompletedDomains(allLegacyRows, []string{})
	if len(noneDone) != 0 || len(byNone) != 0 {
		t.Fatalf("expected 0 effective completed domains with 0 subtopics, got %v (map len %d)", noneDone, len(byNone))
	}

	partialPlusDom2 := []string{"1.1", "1.2", "1.3"}
	for _, s := range services.SubtopicsForDomain("domain2") {
		partialPlusDom2 = append(partialPlusDom2, s.SubTopicID)
	}
	onlyDom2, byDom2 := buildEffectiveCompletedDomains(allLegacyRows, partialPlusDom2)
	if len(onlyDom2) != 1 || onlyDom2[0] != "domain2" {
		t.Fatalf("expected only [domain2] when domain1 is 3/4 and domain2 is 7/7, got %v", onlyDom2)
	}
	if _, ok := byDom2["domain1"]; ok {
		t.Fatalf("incomplete domain1 should not be present in effective domain map")
	}
	if _, ok := byDom2["domain2"]; !ok {
		t.Fatalf("completed domain2 missing from effective domain map")
	}

	allTwentyEight := make([]string, 0, len(services.OrderedTheorySubtopics))
	for _, s := range services.OrderedTheorySubtopics {
		allTwentyEight = append(allTwentyEight, s.SubTopicID)
	}
	allFive, _ := buildEffectiveCompletedDomains(allLegacyRows, allTwentyEight)
	if len(allFive) != len(services.OrderedTheoryDomains) {
		t.Fatalf("expected all %d domains when all 28 subtopics are complete, got %v", len(services.OrderedTheoryDomains), allFive)
	}
}
