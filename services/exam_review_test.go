package services

import (
	"encoding/json"
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

func TestSanitizeQuestionTranslations(t *testing.T) {
	if got := SanitizeQuestionTranslations(nil, false); got != nil {
		t.Fatalf("expected nil for nil input, got %+v", got)
	}
	if got := SanitizeQuestionTranslations(map[string]any{}, true); got != nil {
		t.Fatalf("expected nil for empty map, got %+v", got)
	}
	marshaledEmpty, err := json.Marshal(struct {
		Translations any `json:"translations,omitempty"`
	}{Translations: SanitizeQuestionTranslations(nil, false)})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if string(marshaledEmpty) != "{}" {
		t.Fatalf("expected untyped nil to be omitted by omitempty, got %s", string(marshaledEmpty))
	}

	raw := map[string]any{
		"th": map[string]any{
			"question_text":  "คำถามทดสอบ",
			"options":        map[string]any{"A": "ก", "B": "ข"},
			"explanation":    "คำตอบที่ถูกต้อง: A. เพราะเหตุผล",
			"correct_answer": "A",
		},
	}

	// Pre-submission: must strip explanation (and any correct_answer), preserving stem + options
	pre, ok := SanitizeQuestionTranslations(raw, false).(map[string]any)
	if !ok || pre == nil {
		t.Fatal("expected sanitized map for active exam")
	}
	thPre, ok := pre["th"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested th map, got %+v", pre)
	}
	if thPre["question_text"] != "คำถามทดสอบ" || thPre["options"] == nil {
		t.Fatalf("expected th question_text and options preserved, got %+v", thPre)
	}
	if _, leaked := thPre["explanation"]; leaked {
		t.Fatalf("pre-submission payload leaked explanation: %+v", thPre)
	}
	if _, leaked := thPre["correct_answer"]; leaked {
		t.Fatalf("pre-submission payload leaked correct_answer: %+v", thPre)
	}

	// Post-submission: preserves explanation alongside stem and options from JSON string too
	rawJSON := `{"th":{"question_text":"คำถามจาก JSON","options":{"A":"ก"},"explanation":"คำตอบที่ถูกต้อง: A. ดีมาก"}}`
	post, ok := SanitizeQuestionTranslations(rawJSON, true).(map[string]any)
	if !ok || post == nil {
		t.Fatal("expected decoded map from JSON string")
	}
	thPost := post["th"].(map[string]any)
	if thPost["explanation"] != "คำตอบที่ถูกต้อง: A. ดีมาก" || thPost["question_text"] != "คำถามจาก JSON" {
		t.Fatalf("expected post-submission th fields preserved, got %+v", thPost)
	}
}
