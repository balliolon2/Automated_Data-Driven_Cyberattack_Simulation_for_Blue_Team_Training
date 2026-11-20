package services

import (
	"strings"
	"testing"
)

func TestLoadTheoryModules_AllFiveDomainsPassStructure(t *testing.T) {
	modules, err := LoadTheoryModules(nil)
	if err != nil {
		t.Fatalf("LoadTheoryModules(nil) failed: %v", err)
	}
	if len(modules) != 5 {
		t.Fatalf("expected 5 theory modules, got %d", len(modules))
	}

	expectedIDs := []string{"domain1", "domain2", "domain3", "domain4", "domain5"}
	for i, mod := range modules {
		if mod.DomainID != expectedIDs[i] {
			t.Errorf("module %d: got domain_id %q, want %q", i, mod.DomainID, expectedIDs[i])
		}
		if mod.Title == "" || mod.DomainName == "" {
			t.Errorf("module %s has empty title or domain name", mod.DomainID)
		}
		if !strings.HasPrefix(mod.SourceURL, "https://") {
			t.Errorf("module %s expected clickable https SourceURL, got %q", mod.DomainID, mod.SourceURL)
		}
		if !strings.Contains(mod.Content, "](https://") || !strings.Contains(mod.Content, "https://en.wikipedia.org/") {
			t.Errorf("module %s expected clickable Wikipedia/web reference links inside content", mod.DomainID)
		}
		if err := ValidateTheoryModuleContent(mod.DomainID, mod.Content); err != nil {
			t.Errorf("module %s failed structural validation: %v", mod.DomainID, err)
		}
	}
}

func TestValidateTheoryModuleContent(t *testing.T) {
	validDoc := `# Domain 1: Example
## 1. Core Security+ Theory & Exam Concepts
Concepts here.
## 2. SOC Telemetry, Log Indicators & MITRE ATT&CK Mapping
KQL patterns here.
## 3. Alert Triage (TP vs. FP) & 3-Phase Playbook Response Guide
True Positive and False Positive triage.
## 4. Key Takeaways & References
Summary here.`

	if err := ValidateTheoryModuleContent("domain1", validDoc); err != nil {
		t.Fatalf("expected valid document to pass, got %v", err)
	}

	cases := []struct {
		name    string
		content string
	}{
		{name: "empty content", content: "   "},
		{name: "missing section 2", content: "## 1. Core Security+ Theory\n## 3. Alert Triage\nTrue Positive False Positive\n## 4. Key Takeaways"},
		{name: "missing TP/FP guidance", content: "## 1. Core Security+ Theory\n## 2. SOC Telemetry\nKQL\n## 3. Alert Triage\n## 4. Key Takeaways"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateTheoryModuleContent("domain1", tc.content); err == nil {
				t.Errorf("expected error for %s, got nil", tc.name)
			}
		})
	}
}

func TestCanAccessTheoryHub(t *testing.T) {
	cases := []struct {
		name              string
		role              string
		skillProfileCount int
		want              bool
	}{
		{name: "learner before pre-test is locked", role: "learner", skillProfileCount: 0, want: false},
		{name: "learner after pre-test is unlocked", role: "learner", skillProfileCount: 5, want: true},
		{name: "specialist without pre-test is unlocked", role: "specialist", skillProfileCount: 0, want: true},
		{name: "admin without pre-test is unlocked", role: "admin", skillProfileCount: 0, want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CanAccessTheoryHub(tc.role, tc.skillProfileCount)
			if got != tc.want {
				t.Errorf("CanAccessTheoryHub(%q, %d) = %v, want %v", tc.role, tc.skillProfileCount, got, tc.want)
			}
		})
	}
}

func TestDetermineRecommendedFocusDomains(t *testing.T) {
	t.Run("flags all domains strictly below threshold", func(t *testing.T) {
		scores := map[string]float64{
			"domain1": 75.0,
			"domain2": 55.0,
			"domain3": 80.0,
			"domain4": 69.9,
			"domain5": 70.0,
		}
		focus := DetermineRecommendedFocusDomains(scores, 70.0)
		if focus["domain1"] || focus["domain3"] || focus["domain5"] {
			t.Errorf("domains >= 70 should not be flagged when weak domains exist: %+v", focus)
		}
		if !focus["domain2"] || !focus["domain4"] {
			t.Errorf("expected domain2 and domain4 to be flagged, got %+v", focus)
		}
	})

	t.Run("flags single lowest domain when all pass threshold", func(t *testing.T) {
		scores := map[string]float64{
			"domain1": 71.4,
			"domain2": 90.9,
			"domain3": 94.6,
			"domain4": 70.4,
			"domain5": 75.0,
		}
		focus := DetermineRecommendedFocusDomains(scores, 70.0)
		for _, id := range []string{"domain1", "domain2", "domain3", "domain5"} {
			if focus[id] {
				t.Errorf("did not expect %s to be flagged, got %+v", id, focus)
			}
		}
		if !focus["domain4"] {
			t.Errorf("expected lowest domain4 (70.4) to be flagged, got %+v", focus)
		}
	})
}

func TestIsTheoryGateSatisfied(t *testing.T) {
	if IsTheoryGateSatisfied([]string{"domain1", "domain2", "domain3", "domain4"}) {
		t.Errorf("expected gate to be unsatisfied with 4/5 domains")
	}
	if !IsTheoryGateSatisfied([]string{"domain5", "domain1", "domain3", "domain2", "domain4"}) {
		t.Errorf("expected gate to be satisfied when all 5 domains are present")
	}
	if IsTheoryGateSatisfied([]string{"domain1", "domain1", "domain2", "domain3", "domain4"}) {
		t.Errorf("duplicate domain IDs must not satisfy the 5-domain gate")
	}
}

func TestLoadTheorySubtopics_AllTwentyEightPassStructure(t *testing.T) {
	subtopics, err := LoadTheorySubtopics(nil)
	if err != nil {
		t.Fatalf("LoadTheorySubtopics(nil) failed: %v", err)
	}
	expectedIDs := []string{
		"1.1", "1.2", "1.3", "1.4",
		"2.1", "2.2", "2.3", "2.4", "2.5",
		"3.1", "3.2", "3.3", "3.4",
		"4.1", "4.2", "4.3", "4.4", "4.5", "4.6", "4.7", "4.8", "4.9",
		"5.1", "5.2", "5.3", "5.4", "5.5", "5.6",
	}
	if len(subtopics) != len(expectedIDs) {
		t.Fatalf("expected %d subtopics, got %d", len(expectedIDs), len(subtopics))
	}
	for i, sub := range subtopics {
		if sub.SubTopicID != expectedIDs[i] {
			t.Errorf("index %d: got sub_topic_id %q, want %q", i, sub.SubTopicID, expectedIDs[i])
		}
		if sub.Title == "" || sub.DomainID == "" {
			t.Errorf("subtopic %s has empty title or domain_id", sub.SubTopicID)
		}
		if !strings.HasPrefix(sub.SourceURL, "https://") {
			t.Errorf("subtopic %s expected clickable https SourceURL, got %q", sub.SubTopicID, sub.SourceURL)
		}
		if !strings.Contains(sub.Content, "](https://") {
			t.Errorf("subtopic %s expected clickable https links inside content", sub.SubTopicID)
		}
		if err := ValidateTheorySubtopicContent(sub.SubTopicID, sub.Content); err != nil {
			t.Errorf("subtopic %s failed validation: %v", sub.SubTopicID, err)
		}
	}

	prev, next := GetSubtopicNeighbors("1.4")
	if prev != "1.3" || next != "2.1" {
		t.Errorf("neighbors of 1.4 = (%q, %q), want (1.3, 2.1)", prev, next)
	}
	prevFirst, nextFirst := GetSubtopicNeighbors("1.1")
	if prevFirst != "" || nextFirst != "1.2" {
		t.Errorf("neighbors of 1.1 = (%q, %q), want (\"\", 1.2)", prevFirst, nextFirst)
	}
	prevLast, nextLast := GetSubtopicNeighbors("5.6")
	if prevLast != "5.5" || nextLast != "" {
		t.Errorf("neighbors of 5.6 = (%q, %q), want (5.5, \"\")", prevLast, nextLast)
	}
}

func containsThaiScript(s string) bool {
	for _, r := range s {
		if r >= 0x0E00 && r <= 0x0E7F {
			return true
		}
	}
	return false
}

func TestTheoryLocalization_AllThirtyThreeDocsAndGracefulFallback(t *testing.T) {
	modules, err := LoadTheoryModulesByLocale(nil, "th")
	if err != nil {
		t.Fatalf("LoadTheoryModulesByLocale(nil, th) failed: %v", err)
	}
	if len(modules) != 5 {
		t.Fatalf("expected 5 modules, got %d", len(modules))
	}
	enModules, _ := LoadTheoryModules(nil)
	for i := range modules {
		if !containsThaiScript(modules[i].Title) || !containsThaiScript(modules[i].Content) {
			t.Errorf("%s module should be translated in Thai, got title=%q", modules[i].DomainID, modules[i].Title)
		}
		if err := ValidateThaiTheoryModuleContent(modules[i].DomainID, enModules[i].Content, modules[i].Content); err != nil {
			t.Errorf("%s Thai validation failed: %v", modules[i].DomainID, err)
		}
	}

	subtopics, err := LoadTheorySubtopicsByLocale(nil, "th")
	if err != nil {
		t.Fatalf("LoadTheorySubtopicsByLocale(nil, th) failed: %v", err)
	}
	if len(subtopics) != 28 {
		t.Fatalf("expected 28 subtopics, got %d", len(subtopics))
	}
	enSubtopics, _ := LoadTheorySubtopics(nil)
	for i := range subtopics {
		if !containsThaiScript(subtopics[i].Title) || !containsThaiScript(subtopics[i].Content) {
			t.Errorf("subtopic %s should be translated in Thai, got title=%q", subtopics[i].SubTopicID, subtopics[i].Title)
		}
		if err := ValidateThaiTheorySubtopicContent(subtopics[i].SubTopicID, enSubtopics[i].Content, subtopics[i].Content); err != nil {
			t.Errorf("subtopic %s failed validation: %v", subtopics[i].SubTopicID, err)
		}
	}
	if err := ValidateThaiTheorySubtopicContent("1.1", enSubtopics[0].Content, enSubtopics[0].Content); err == nil {
		t.Errorf("expected error when passing English-only stub to ValidateThaiTheorySubtopicContent")
	}

	// Graceful fallback to English when locale has no translations
	fallbackModules, err := LoadTheoryModulesByLocale(nil, "fr")
	if err != nil || len(fallbackModules) != 5 || fallbackModules[0].Title != enModules[0].Title {
		t.Errorf("expected unsupported locale to gracefully fall back to English, err=%v", err)
	}

	// Accept-Language parsing
	if got, ok := ParseAcceptLanguage("th-TH,th;q=0.9,en;q=0.8"); !ok || got != "th" {
		t.Errorf("ParseAcceptLanguage(th-TH,...) = (%q, %v), want (th, true)", got, ok)
	}
	if got, ok := ParseAcceptLanguage("en-US,th;q=0.5"); !ok || got != "en" {
		t.Errorf("ParseAcceptLanguage(en-US,th) = (%q, %v), want (en, true)", got, ok)
	}
	if _, ok := ParseAcceptLanguage("fr-FR"); ok {
		t.Errorf("expected unsupported Accept-Language fr-FR to return ok=false")
	}
}
