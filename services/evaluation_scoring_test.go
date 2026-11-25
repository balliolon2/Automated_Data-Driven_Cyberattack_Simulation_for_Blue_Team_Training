package services

import (
	"errors"
	"strings"
	"testing"

	"cybersim/dto"
)

func intPtr(v int) *int {
	return &v
}

func makeValidCSUQInput() map[string]*int {
	return map[string]*int{
		"q1": intPtr(6), "q2": intPtr(7), "q3": intPtr(6), "q4": intPtr(6), "q5": intPtr(7), "q6": intPtr(6),
		"q7": nil, "q8": intPtr(6), "q9": intPtr(6), "q10": intPtr(6), "q11": intPtr(7), "q12": intPtr(6),
		"q13": intPtr(7), "q14": intPtr(6), "q15": intPtr(6), "q16": intPtr(7),
	}
}

func makeValidTLXInput() dto.NASATLXInputDTO {
	return dto.NASATLXInputDTO{
		MentalDemand:      intPtr(60),
		PhysicalDemand:    intPtr(10),
		TemporalDemand:    intPtr(45),
		PerformanceRating: intPtr(85),
		Effort:            intPtr(55),
		Frustration:       intPtr(20),
	}
}

func TestValidateAndScoreCSUQ_ValidWithNA(t *testing.T) {
	res, err := ValidateAndScoreCSUQ(makeValidCSUQInput(), "th")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.SchemaVersion != "csuq-v3-asc-7" || res.Direction != "higher_is_better" || res.Locale != "th" {
		t.Fatalf("unexpected metadata: %+v", res)
	}
	// sysuse: (6+7+6+6+7+6)/6 = 38/6 = 6.33
	if res.Subscales.SysUse == nil || *res.Subscales.SysUse != 6.33 || res.ValidCounts.SysUse != 6 {
		t.Fatalf("unexpected sysuse: %v (count=%d)", res.Subscales.SysUse, res.ValidCounts.SysUse)
	}
	// infoqual (q7 nil): (6+6+6+7+6)/5 = 31/5 = 6.20
	if res.Subscales.InfoQual == nil || *res.Subscales.InfoQual != 6.20 || res.ValidCounts.InfoQual != 5 {
		t.Fatalf("unexpected infoqual: %v (count=%d)", res.Subscales.InfoQual, res.ValidCounts.InfoQual)
	}
	// interqual: (7+6+6)/3 = 19/3 = 6.33
	if res.Subscales.InterQual == nil || *res.Subscales.InterQual != 6.33 || res.ValidCounts.InterQual != 3 {
		t.Fatalf("unexpected interqual: %v (count=%d)", res.Subscales.InterQual, res.ValidCounts.InterQual)
	}
	// overall (15 items, q7 nil): (38+31+19+7)/15 = 95/15 = 6.33
	if res.Subscales.Overall == nil || *res.Subscales.Overall != 6.33 || res.ValidCounts.Overall != 15 {
		t.Fatalf("unexpected overall: %v (count=%d)", res.Subscales.Overall, res.ValidCounts.Overall)
	}
}

func TestValidateAndScoreCSUQ_Errors(t *testing.T) {
	allNA := make(map[string]*int, 16)
	for _, k := range CSUQItemKeys {
		allNA[k] = nil
	}
	if _, err := ValidateAndScoreCSUQ(allNA, "en"); !errors.Is(err, ErrQuestionnaireAllNA) {
		t.Fatalf("expected ErrQuestionnaireAllNA, got %v", err)
	}

	missingKey := makeValidCSUQInput()
	delete(missingKey, "q16")
	if _, err := ValidateAndScoreCSUQ(missingKey, "en"); !errors.Is(err, ErrQuestionnaireInvalidCSUQ) {
		t.Fatalf("expected ErrQuestionnaireInvalidCSUQ for missing key, got %v", err)
	}

	extraKey := makeValidCSUQInput()
	extraKey["q17"] = intPtr(5)
	if _, err := ValidateAndScoreCSUQ(extraKey, "en"); !errors.Is(err, ErrQuestionnaireInvalidCSUQ) {
		t.Fatalf("expected ErrQuestionnaireInvalidCSUQ for extra key, got %v", err)
	}

	outOfRange := makeValidCSUQInput()
	outOfRange["q1"] = intPtr(8)
	if _, err := ValidateAndScoreCSUQ(outOfRange, "en"); !errors.Is(err, ErrQuestionnaireInvalidCSUQ) {
		t.Fatalf("expected ErrQuestionnaireInvalidCSUQ for score > 7, got %v", err)
	}

	belowRange := makeValidCSUQInput()
	belowRange["q2"] = intPtr(0)
	if _, err := ValidateAndScoreCSUQ(belowRange, "en"); !errors.Is(err, ErrQuestionnaireInvalidCSUQ) {
		t.Fatalf("expected ErrQuestionnaireInvalidCSUQ for score < 1, got %v", err)
	}
}

func TestValidateAndScoreCSUQ_SubscaleAllNAYieldsNilSubscale(t *testing.T) {
	input := makeValidCSUQInput()
	for _, k := range []string{"q7", "q8", "q9", "q10", "q11", "q12"} {
		input[k] = nil
	}
	res, err := ValidateAndScoreCSUQ(input, "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Subscales.InfoQual != nil || res.ValidCounts.InfoQual != 0 {
		t.Fatalf("expected nil infoqual when all items are N/A, got %v", res.Subscales.InfoQual)
	}
	if res.Subscales.Overall == nil || res.ValidCounts.Overall != 10 {
		t.Fatalf("expected overall with 10 valid items, got count=%d", res.ValidCounts.Overall)
	}
}

func TestValidateAndScoreNASATLX_ValidAndZeroAccepted(t *testing.T) {
	res, err := ValidateAndScoreNASATLX(makeValidTLXInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// workloads: 60 + 10 + 45 + (100-85=15) + 55 + 20 = 205 / 6 = 34.17
	if res.Dimensions.PerformanceRating != 85 || res.Dimensions.PerformanceWorkload != 15 {
		t.Fatalf("expected performance inversion 85 -> 15, got rating=%d workload=%d", res.Dimensions.PerformanceRating, res.Dimensions.PerformanceWorkload)
	}
	if res.OverallWorkload != 34.17 {
		t.Fatalf("expected overall workload 34.17, got %v", res.OverallWorkload)
	}

	// Zero values must be accepted as valid integers
	allZero := dto.NASATLXInputDTO{
		MentalDemand: intPtr(0), PhysicalDemand: intPtr(0), TemporalDemand: intPtr(0),
		PerformanceRating: intPtr(100), Effort: intPtr(0), Frustration: intPtr(0),
	}
	zeroRes, err := ValidateAndScoreNASATLX(allZero)
	if err != nil {
		t.Fatalf("unexpected error on 0 ratings: %v", err)
	}
	if zeroRes.OverallWorkload != 0 {
		t.Fatalf("expected 0 overall workload, got %v", zeroRes.OverallWorkload)
	}
}

func TestValidateAndScoreNASATLX_Errors(t *testing.T) {
	missing := makeValidTLXInput()
	missing.PhysicalDemand = nil
	if _, err := ValidateAndScoreNASATLX(missing); !errors.Is(err, ErrQuestionnaireInvalidTLX) {
		t.Fatalf("expected ErrQuestionnaireInvalidTLX for nil field, got %v", err)
	}

	nonStep5 := makeValidTLXInput()
	nonStep5.Effort = intPtr(53)
	if _, err := ValidateAndScoreNASATLX(nonStep5); !errors.Is(err, ErrQuestionnaireInvalidTLX) {
		t.Fatalf("expected ErrQuestionnaireInvalidTLX for non-multiple-of-5, got %v", err)
	}

	tooHigh := makeValidTLXInput()
	tooHigh.Frustration = intPtr(105)
	if _, err := ValidateAndScoreNASATLX(tooHigh); !errors.Is(err, ErrQuestionnaireInvalidTLX) {
		t.Fatalf("expected ErrQuestionnaireInvalidTLX for >100, got %v", err)
	}
}

func TestSanitizeAndFormatQualitativeFeedback(t *testing.T) {
	fb, notes, err := SanitizeQualitativeFeedback(dto.QualitativeFeedbackDTO{
		MostImpressive:     "  Great SIEM log flow  ",
		AreasToImprove:     "",
		AdditionalComments: "   Helpful bilingual tips   ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.MostImpressive != "Great SIEM log flow" || fb.AreasToImprove != "" || fb.AdditionalComments != "Helpful bilingual tips" {
		t.Fatalf("unexpected trimmed feedback: %+v", fb)
	}
	wantNotes := "[Most Impressive]\nGreat SIEM log flow\n\n[Additional Comments]\nHelpful bilingual tips"
	if notes != wantNotes {
		t.Fatalf("expected notes:\n%q\ngot:\n%q", wantNotes, notes)
	}

	_, emptyNotes, err := SanitizeQualitativeFeedback(dto.QualitativeFeedbackDTO{})
	if err != nil || emptyNotes != "" {
		t.Fatalf("expected empty notes without error, got %q err=%v", emptyNotes, err)
	}

	// Over 2000 Unicode runes (Thai characters are 3 bytes each, verify rune counting)
	tooLong := strings.Repeat("ก", MaxQualitativeFeedbackRunes+1)
	if _, _, err := SanitizeQualitativeFeedback(dto.QualitativeFeedbackDTO{MostImpressive: tooLong}); !errors.Is(err, ErrQuestionnaireFeedbackTooLong) {
		t.Fatalf("expected ErrQuestionnaireFeedbackTooLong, got %v", err)
	}

	// Exactly 2000 Unicode runes should succeed
	exactLimit := strings.Repeat("ก", MaxQualitativeFeedbackRunes)
	if _, _, err := SanitizeQualitativeFeedback(dto.QualitativeFeedbackDTO{MostImpressive: exactLimit}); err != nil {
		t.Fatalf("expected exact 2000 runes to succeed, got %v", err)
	}
}
