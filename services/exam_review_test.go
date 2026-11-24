package services

import (
	"testing"

	"cybersim/dto"
)

func TestExamReviewService_SummarizeClampAndPreComparison(t *testing.T) {
	if got := ClampExamDuration(-10, 0); got != 0 {
		t.Fatalf("ClampExamDuration negative = %d, want 0", got)
	}
	if got := ClampExamDuration(99999, 100); got != ExamMaxDurationSeconds {
		t.Fatalf("ClampExamDuration upper = %d, want %d", got, ExamMaxDurationSeconds)
	}
	if got := ClampExamDuration(200, 450); got != 450 {
		t.Fatalf("ClampExamDuration existing = %d, want 450", got)
	}

	sub1 := "1.1"
	sub2 := "2.1"
	results := []dto.DetailedExamResultDTO{
		{QuestionID: "q1", DomainID: "domain2", SubTopicID: &sub2, UserAnswer: "", IsCorrect: false},
		{QuestionID: "q2", DomainID: "domain1", SubTopicID: &sub1, UserAnswer: "A", IsCorrect: true},
		{QuestionID: "q3", DomainID: "domain1", SubTopicID: &sub1, UserAnswer: "B", IsCorrect: false},
	}

	rev := SummarizeExamResults(results, 70.0)
	if rev.CorrectCount != 1 || rev.IncorrectCount != 1 || rev.UnansweredCount != 1 {
		t.Fatalf("unexpected counts: %+v", rev)
	}
	if len(rev.DomainScores) != 2 || rev.DomainScores[0].DomainID != "domain1" || rev.DomainScores[1].DomainID != "domain2" {
		t.Fatalf("domain scores not canonically ordered: %+v", rev.DomainScores)
	}
	if rev.Results[0].IsUnanswered != true || rev.Results[0].SubTopicTitle == "" || rev.Results[1].DomainName == "" {
		t.Fatalf("expected enrichment on results: %+v", rev.Results)
	}

	preScore := 50.0
	postScore := 80.0
	preRev := &ExamReviewData{
		CorrectCount: 15,
		DomainScores: []dto.DomainScoreDTO{{DomainID: "domain1", DomainName: "General Security Concepts", Percentage: 50}},
	}
	postRev := &ExamReviewData{
		CorrectCount: 24,
		DomainScores: []dto.DomainScoreDTO{{DomainID: "domain1", DomainName: "General Security Concepts", Percentage: 100}},
	}
	comp := BuildPreComparison("pre-1", &preScore, 30, preRev, &postScore, 30, postRev)
	if comp == nil || comp.ScoreDelta != 30 || len(comp.DomainDeltas) != 1 || comp.DomainDeltas[0].Delta != 50 {
		t.Fatalf("unexpected BuildPreComparison: %+v", comp)
	}
}
