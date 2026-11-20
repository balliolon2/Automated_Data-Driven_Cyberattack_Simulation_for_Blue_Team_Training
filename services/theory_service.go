package services

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"cybersim/models"

	"gorm.io/gorm"
)

// TheoryDomainSpec defines canonical ordering and metadata for one Security+ domain.
type TheoryDomainSpec struct {
	DomainID   string
	DomainName string
	Title      string
	Filename   string
	SourceURL  string
}

// TheoryModuleContent represents a resolved, validated Theory Module ready for delivery.
type TheoryModuleContent struct {
	DomainID   string
	DomainName string
	Title      string
	Content    string
	SourceURL  string
}

// TheorySubtopicSpec defines canonical ordering and metadata for one SY0-701 sub-topic.
type TheorySubtopicSpec struct {
	SubTopicID string
	DomainID   string
	OrderIndex int
	Title      string
	Filename   string
	SourceURL  string
}

// TheorySubtopicContent represents a resolved SY0-701 sub-topic document ready for delivery.
type TheorySubtopicContent struct {
	SubTopicID string
	DomainID   string
	OrderIndex int
	Title      string
	Content    string
	SourceURL  string
}

// OrderedTheoryDomains defines canonical module order and web references.
var OrderedTheoryDomains = []TheoryDomainSpec{
	{
		DomainID:   "domain1",
		DomainName: "General Security Concepts",
		Title:      "Domain 1: General Security Concepts",
		Filename:   "domain1_security_concepts.md",
		SourceURL:  "https://csrc.nist.gov/pubs/sp/800/53/r5/final",
	},
	{
		DomainID:   "domain2",
		DomainName: "Threats, Vulnerabilities, and Mitigations",
		Title:      "Domain 2: Threats, Vulnerabilities, and Mitigations",
		Filename:   "domain2_threats_vulnerabilities.md",
		SourceURL:  "https://owasp.org/Top10/",
	},
	{
		DomainID:   "domain3",
		DomainName: "Security Architecture",
		Title:      "Domain 3: Security Architecture",
		Filename:   "domain3_security_architecture.md",
		SourceURL:  "https://csrc.nist.gov/pubs/sp/800/207/final",
	},
	{
		DomainID:   "domain4",
		DomainName: "Security Operations",
		Title:      "Domain 4: Security Operations",
		Filename:   "domain4_security_operations.md",
		SourceURL:  "https://csrc.nist.gov/pubs/sp/800/61/r3/final",
	},
	{
		DomainID:   "domain5",
		DomainName: "Security Program Management and Oversight",
		Title:      "Domain 5: Security Program Management and Oversight",
		Filename:   "domain5_security_management.md",
		SourceURL:  "https://www.cmu.edu/iso/governance/procedures/incident-response.html",
	},
}

var (
	embeddedTheoryMu sync.RWMutex
	embeddedTheoryFS fs.FS
)

// RegisterEmbeddedTheoryFS registers the embedded theory/*.md filesystem
// compiled into the main binary so fresh deployments always have a fallback.
func RegisterEmbeddedTheoryFS(fsys fs.FS) {
	embeddedTheoryMu.Lock()
	defer embeddedTheoryMu.Unlock()
	embeddedTheoryFS = fsys
}

func getEmbeddedTheoryFS() fs.FS {
	embeddedTheoryMu.RLock()
	defer embeddedTheoryMu.RUnlock()
	return embeddedTheoryFS
}

var requiredTheorySectionPatterns = []struct {
	label   string
	pattern *regexp.Regexp
}{
	{"Section 1 (Core Security+ Theory)", regexp.MustCompile(`(?m)^##\s+1\.\s+Core Security\+ Theory`)},
	{"Section 2 (SOC Telemetry & MITRE)", regexp.MustCompile(`(?m)^##\s+2\.\s+SOC Telemetry`)},
	{"Section 3 (Alert Triage & Playbook)", regexp.MustCompile(`(?m)^##\s+3\.\s+Alert Triage`)},
	{"Section 4 (Key Takeaways & References)", regexp.MustCompile(`(?m)^##\s+4\.\s+Key Takeaways`)},
}

// ValidateTheoryModuleContent verifies that a Theory Module has non-empty prose
// and all four required curriculum sections plus operational guidance.
func ValidateTheoryModuleContent(domainID, content string) error {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	if strings.TrimSpace(normalized) == "" {
		return fmt.Errorf("%s: theory content is empty", domainID)
	}
	for _, req := range requiredTheorySectionPatterns {
		if !req.pattern.MatchString(normalized) {
			return fmt.Errorf("%s: missing required %s", domainID, req.label)
		}
	}
	lower := strings.ToLower(normalized)
	if !strings.Contains(lower, "kql") {
		return fmt.Errorf("%s: missing KQL investigation patterns", domainID)
	}
	if !strings.Contains(lower, "true positive") || !strings.Contains(lower, "false positive") {
		return fmt.Errorf("%s: missing True Positive / False Positive triage guidance", domainID)
	}
	return nil
}

// IsValidTheoryDomain checks whether domainID is one of the five Security+ domains.
func IsValidTheoryDomain(domainID string) bool {
	for _, spec := range OrderedTheoryDomains {
		if spec.DomainID == domainID {
			return true
		}
	}
	return false
}

// CanAccessTheoryHub enforces Stage 1 of the Theory Gate:
// Specialists and Admins always have access; Learners unlock it once Pre-Test
// creates their skill profile rows.
func CanAccessTheoryHub(role string, skillProfileCount int) bool {
	if role == "specialist" || role == "admin" {
		return true
	}
	return skillProfileCount > 0
}

// DetermineRecommendedFocusDomains marks every domain strictly below threshold
// as recommended focus, or if all pass threshold, marks the single lowest domain.
func DetermineRecommendedFocusDomains(scores map[string]float64, threshold float64) map[string]bool {
	focus := make(map[string]bool, len(OrderedTheoryDomains))
	hasWeak := false
	for _, spec := range OrderedTheoryDomains {
		if score, ok := scores[spec.DomainID]; ok && score < threshold {
			focus[spec.DomainID] = true
			hasWeak = true
		}
	}
	if hasWeak || len(scores) == 0 {
		return focus
	}

	lowestID := ""
	lowestScore := 101.0
	for _, spec := range OrderedTheoryDomains {
		score, ok := scores[spec.DomainID]
		if !ok {
			continue
		}
		if lowestID == "" || score < lowestScore {
			lowestID = spec.DomainID
			lowestScore = score
		}
	}
	if lowestID != "" {
		focus[lowestID] = true
	}
	return focus
}

// IsTheoryGateSatisfied returns true only when all five canonical domains have
// been marked completed.
func IsTheoryGateSatisfied(completedDomains []string) bool {
	seen := make(map[string]bool, len(completedDomains))
	for _, id := range completedDomains {
		if IsValidTheoryDomain(id) {
			seen[id] = true
		}
	}
	return len(seen) == len(OrderedTheoryDomains)
}

// LoadTheoryModules returns all five ordered Theory Modules, preferring seeded
// rows in rag_documents when complete and falling back to the embedded/local
// theory Markdown files.
func LoadTheoryModules(db *gorm.DB) ([]TheoryModuleContent, error) {
	if db != nil {
		var docs []models.RagDocument
		if err := db.Where("domain_id IN ? AND title LIKE 'Domain %'", []string{
			"domain1", "domain2", "domain3", "domain4", "domain5",
		}).Find(&docs).Error; err == nil && len(docs) >= len(OrderedTheoryDomains) {
			byDomain := make(map[string]models.RagDocument, len(docs))
			for _, d := range docs {
				if strings.TrimSpace(d.Content) != "" {
					byDomain[d.DomainID] = d
				}
			}
			if len(byDomain) == len(OrderedTheoryDomains) {
				out := make([]TheoryModuleContent, 0, len(OrderedTheoryDomains))
				for _, spec := range OrderedTheoryDomains {
					doc := byDomain[spec.DomainID]
					sourceURL := doc.SourceURL
					if !strings.HasPrefix(sourceURL, "https://") {
						sourceURL = spec.SourceURL
					}
					out = append(out, TheoryModuleContent{
						DomainID:   spec.DomainID,
						DomainName: spec.DomainName,
						Title:      spec.Title,
						Content:    strings.ReplaceAll(doc.Content, "\r\n", "\n"),
						SourceURL:  sourceURL,
					})
				}
				return out, nil
			}
		}
	}

	return loadTheoryModulesFromFS()
}

func loadTheoryModulesFromFS() ([]TheoryModuleContent, error) {
	emb := getEmbeddedTheoryFS()
	out := make([]TheoryModuleContent, 0, len(OrderedTheoryDomains))

	for _, spec := range OrderedTheoryDomains {
		raw, err := readTheoryMarkdown(emb, spec.Filename)
		if err != nil {
			return nil, fmt.Errorf("failed to load theory module %s: %w", spec.DomainID, err)
		}
		normalized := strings.ReplaceAll(raw, "\r\n", "\n")
		out = append(out, TheoryModuleContent{
			DomainID:   spec.DomainID,
			DomainName: spec.DomainName,
			Title:      spec.Title,
			Content:    normalized,
			SourceURL:  spec.SourceURL,
		})
	}
	return out, nil
}

func readTheoryMarkdown(emb fs.FS, filename string) (string, error) {
	if emb != nil {
		if data, err := fs.ReadFile(emb, filepath.ToSlash(filepath.Join("theory", filename))); err == nil {
			return string(data), nil
		}
	}

	candidates := []string{
		filepath.Join("theory", filename),
		filepath.Join("..", "theory", filename),
	}
	for _, path := range candidates {
		if data, err := os.ReadFile(path); err == nil {
			return string(data), nil
		}
	}
	return "", fmt.Errorf("theory markdown file %s not found in embedded FS or local theory/ directory", filename)
}

// OrderedTheorySubtopics defines the exact 28 CompTIA Security+ (SY0-701) objectives.
var OrderedTheorySubtopics = []TheorySubtopicSpec{
	{SubTopicID: "1.1", DomainID: "domain1", OrderIndex: 1, Title: "Compare and contrast various types of security controls", Filename: "domain1/1_1_security_controls.md", SourceURL: "https://csrc.nist.gov/pubs/sp/800/53/r5/final"},
	{SubTopicID: "1.2", DomainID: "domain1", OrderIndex: 2, Title: "Summarize fundamental security concepts", Filename: "domain1/1_2_fundamental_security_concepts.md", SourceURL: "https://csrc.nist.gov/pubs/sp/800/207/final"},
	{SubTopicID: "1.3", DomainID: "domain1", OrderIndex: 3, Title: "Explain change management best practices", Filename: "domain1/1_3_change_management.md", SourceURL: "https://en.wikipedia.org/wiki/Change_control"},
	{SubTopicID: "1.4", DomainID: "domain1", OrderIndex: 4, Title: "Explain the importance of using appropriate cryptographic solutions", Filename: "domain1/1_4_cryptographic_solutions.md", SourceURL: "https://en.wikipedia.org/wiki/Public_key_infrastructure"},
	{SubTopicID: "2.1", DomainID: "domain2", OrderIndex: 1, Title: "Compare and contrast threat actors and intelligence sources", Filename: "domain2/2_1_threat_actors_and_motivations.md", SourceURL: "https://attack.mitre.org/matrices/enterprise/"},
	{SubTopicID: "2.2", DomainID: "domain2", OrderIndex: 2, Title: "Analyze common threat vectors and attack surfaces", Filename: "domain2/2_2_threat_vectors_and_attack_surfaces.md", SourceURL: "https://www.cisa.gov/topics/cyber-threats-and-responses/phishing"},
	{SubTopicID: "2.3", DomainID: "domain2", OrderIndex: 3, Title: "Explain various types of vulnerabilities", Filename: "domain2/2_3_types_of_vulnerabilities.md", SourceURL: "https://owasp.org/Top10/"},
	{SubTopicID: "2.4", DomainID: "domain2", OrderIndex: 4, Title: "Analyze indicators of malicious activity", Filename: "domain2/2_4_indicators_of_malicious_activity.md", SourceURL: "https://en.wikipedia.org/wiki/Indicator_of_compromise"},
	{SubTopicID: "2.5", DomainID: "domain2", OrderIndex: 5, Title: "Explain mitigation techniques used to secure the enterprise", Filename: "domain2/2_5_mitigation_techniques.md", SourceURL: "https://csrc.nist.gov/pubs/sp/800/40/r4/final"},
	{SubTopicID: "3.1", DomainID: "domain3", OrderIndex: 1, Title: "Compare and contrast security architecture models", Filename: "domain3/3_1_security_architecture_models.md", SourceURL: "https://csrc.nist.gov/pubs/sp/800/207/final"},
	{SubTopicID: "3.2", DomainID: "domain3", OrderIndex: 2, Title: "Apply security principles to secure enterprise infrastructure", Filename: "domain3/3_2_enterprise_infrastructure_security.md", SourceURL: "https://en.wikipedia.org/wiki/Defense_in_depth_%28computing%29"},
	{SubTopicID: "3.3", DomainID: "domain3", OrderIndex: 3, Title: "Compare and contrast concepts and strategies to protect data", Filename: "domain3/3_3_data_protection_strategies.md", SourceURL: "https://en.wikipedia.org/wiki/Data_loss_prevention_software"},
	{SubTopicID: "3.4", DomainID: "domain3", OrderIndex: 4, Title: "Explain the importance of resilience and recovery in security architecture", Filename: "domain3/3_4_resilience_and_recovery.md", SourceURL: "https://www.cisa.gov/stopransomware"},
	{SubTopicID: "4.1", DomainID: "domain4", OrderIndex: 1, Title: "Apply common security techniques to computing resources", Filename: "domain4/4_1_computing_resource_security.md", SourceURL: "https://en.wikipedia.org/wiki/Endpoint_detection_and_response"},
	{SubTopicID: "4.2", DomainID: "domain4", OrderIndex: 2, Title: "Explain security implications of hardware, software, and data asset management", Filename: "domain4/4_2_asset_management.md", SourceURL: "https://csrc.nist.gov/pubs/sp/800/88/r1/final"},
	{SubTopicID: "4.3", DomainID: "domain4", OrderIndex: 3, Title: "Explain various activities associated with vulnerability management", Filename: "domain4/4_3_vulnerability_management.md", SourceURL: "https://en.wikipedia.org/wiki/Vulnerability_management"},
	{SubTopicID: "4.4", DomainID: "domain4", OrderIndex: 4, Title: "Explain security alerting and monitoring concepts and tools", Filename: "domain4/4_4_security_alerting_and_monitoring.md", SourceURL: "https://en.wikipedia.org/wiki/Security_information_and_event_management"},
	{SubTopicID: "4.5", DomainID: "domain4", OrderIndex: 5, Title: "Modify enterprise capabilities to enhance security", Filename: "domain4/4_5_modifying_enterprise_capabilities.md", SourceURL: "https://csrc.nist.gov/pubs/sp/800/53/r5/final"},
	{SubTopicID: "4.6", DomainID: "domain4", OrderIndex: 6, Title: "Implement and maintain identity and access management", Filename: "domain4/4_6_identity_and_access_management.md", SourceURL: "https://csrc.nist.gov/pubs/sp/800/63/3/final"},
	{SubTopicID: "4.7", DomainID: "domain4", OrderIndex: 7, Title: "Explain the importance of automation and orchestration", Filename: "domain4/4_7_automation_and_orchestration.md", SourceURL: "https://en.wikipedia.org/wiki/Security_orchestration"},
	{SubTopicID: "4.8", DomainID: "domain4", OrderIndex: 8, Title: "Explain appropriate incident response activities", Filename: "domain4/4_8_incident_response_activities.md", SourceURL: "https://csrc.nist.gov/pubs/sp/800/61/r3/final"},
	{SubTopicID: "4.9", DomainID: "domain4", OrderIndex: 9, Title: "Use data sources to support an investigation and digital forensics", Filename: "domain4/4_9_data_sources_and_digital_forensics.md", SourceURL: "https://csrc.nist.gov/pubs/sp/800/86/final"},
	{SubTopicID: "5.1", DomainID: "domain5", OrderIndex: 1, Title: "Summarize elements of effective security governance", Filename: "domain5/5_1_security_governance.md", SourceURL: "https://en.wikipedia.org/wiki/Information_security_governance"},
	{SubTopicID: "5.2", DomainID: "domain5", OrderIndex: 2, Title: "Explain elements of the risk management process", Filename: "domain5/5_2_risk_management_process.md", SourceURL: "https://csrc.nist.gov/pubs/sp/800/30/r1/final"},
	{SubTopicID: "5.3", DomainID: "domain5", OrderIndex: 3, Title: "Explain third-party risk assessment and management processes", Filename: "domain5/5_3_third_party_risk_management.md", SourceURL: "https://csrc.nist.gov/pubs/sp/800/161/r1/final"},
	{SubTopicID: "5.4", DomainID: "domain5", OrderIndex: 4, Title: "Summarize elements of effective security compliance", Filename: "domain5/5_4_security_compliance.md", SourceURL: "https://en.wikipedia.org/wiki/Regulatory_compliance"},
	{SubTopicID: "5.5", DomainID: "domain5", OrderIndex: 5, Title: "Explain types and purposes of audits and assessments", Filename: "domain5/5_5_audits_and_assessments.md", SourceURL: "https://csrc.nist.gov/pubs/sp/800/115/final"},
	{SubTopicID: "5.6", DomainID: "domain5", OrderIndex: 6, Title: "Implement security awareness practices", Filename: "domain5/5_6_security_awareness_practices.md", SourceURL: "https://csrc.nist.gov/pubs/sp/800/50/r1/final"},
}

var requiredSubtopicSectionPatterns = []struct {
	label   string
	pattern *regexp.Regexp
}{
	{"Section 1 (Core Security+ Theory)", regexp.MustCompile(`(?m)^##\s+1\.\s+Core Security\+ Theory`)},
	{"Section 2 (SOC & Blue-Team Practice)", regexp.MustCompile(`(?m)^##\s+2\.\s+SOC & Blue-Team`)},
	{"Section 3 (References)", regexp.MustCompile(`(?m)^##\s+3\.\s+References`)},
}

// ValidateTheorySubtopicContent ensures each Sub-topic file contains non-empty
// prose, required sections, and clickable https:// web references.
func ValidateTheorySubtopicContent(subTopicID, content string) error {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	if strings.TrimSpace(normalized) == "" {
		return fmt.Errorf("%s: sub-topic content is empty", subTopicID)
	}
	for _, req := range requiredSubtopicSectionPatterns {
		if !req.pattern.MatchString(normalized) {
			return fmt.Errorf("%s: missing required %s", subTopicID, req.label)
		}
	}
	if !strings.Contains(normalized, "](https://") {
		return fmt.Errorf("%s: missing clickable https:// reference links", subTopicID)
	}
	return nil
}

// GetSubtopicSpec returns canonical metadata for a Sub-topic ID if valid.
func GetSubtopicSpec(subTopicID string) (TheorySubtopicSpec, bool) {
	for _, spec := range OrderedTheorySubtopics {
		if spec.SubTopicID == subTopicID {
			return spec, true
		}
	}
	return TheorySubtopicSpec{}, false
}

// GetSubtopicNeighbors returns previous and next Sub-topic IDs in sequential curriculum order.
func GetSubtopicNeighbors(subTopicID string) (string, string) {
	for i, spec := range OrderedTheorySubtopics {
		if spec.SubTopicID == subTopicID {
			prev := ""
			next := ""
			if i > 0 {
				prev = OrderedTheorySubtopics[i-1].SubTopicID
			}
			if i+1 < len(OrderedTheorySubtopics) {
				next = OrderedTheorySubtopics[i+1].SubTopicID
			}
			return prev, next
		}
	}
	return "", ""
}

// SubtopicsForDomain returns all canonical Sub-topic specs belonging to domainID.
func SubtopicsForDomain(domainID string) []TheorySubtopicSpec {
	var out []TheorySubtopicSpec
	for _, spec := range OrderedTheorySubtopics {
		if spec.DomainID == domainID {
			out = append(out, spec)
		}
	}
	return out
}

// IsTheorySubtopicGateSatisfied returns true only when all 28 canonical
// CompTIA Security+ (SY0-701) Sub-topics have been completed.
func IsTheorySubtopicGateSatisfied(completedSubtopics []string) bool {
	seen := make(map[string]bool, len(completedSubtopics))
	for _, id := range completedSubtopics {
		if _, ok := GetSubtopicSpec(id); ok {
			seen[id] = true
		}
	}
	return len(seen) == len(OrderedTheorySubtopics)
}

// MissingSubtopicsForDomain returns canonical Sub-topic IDs under domainID that
// are not yet present in completedSubtopics, preserving curriculum order.
func MissingSubtopicsForDomain(domainID string, completedSubtopics []string) []string {
	done := make(map[string]bool, len(completedSubtopics))
	for _, id := range completedSubtopics {
		done[id] = true
	}
	var missing []string
	for _, spec := range SubtopicsForDomain(domainID) {
		if !done[spec.SubTopicID] {
			missing = append(missing, spec.SubTopicID)
		}
	}
	return missing
}

// LoadTheorySubtopics returns all 28 ordered Sub-topics, preferring seeded rows
// in theory_subtopics when complete and falling back to embedded/local Markdown files.
func LoadTheorySubtopics(db *gorm.DB) ([]TheorySubtopicContent, error) {
	if db != nil {
		var rows []models.TheorySubtopic
		if err := db.Find(&rows).Error; err == nil && len(rows) >= len(OrderedTheorySubtopics) {
			byID := make(map[string]models.TheorySubtopic, len(rows))
			for _, r := range rows {
				if strings.TrimSpace(r.Content) != "" {
					byID[r.SubTopicID] = r
				}
			}
			if len(byID) == len(OrderedTheorySubtopics) {
				out := make([]TheorySubtopicContent, 0, len(OrderedTheorySubtopics))
				for _, spec := range OrderedTheorySubtopics {
					row := byID[spec.SubTopicID]
					sourceURL := row.SourceURL
					if !strings.HasPrefix(sourceURL, "https://") {
						sourceURL = spec.SourceURL
					}
					out = append(out, TheorySubtopicContent{
						SubTopicID: spec.SubTopicID,
						DomainID:   spec.DomainID,
						OrderIndex: spec.OrderIndex,
						Title:      spec.Title,
						Content:    strings.ReplaceAll(row.Content, "\r\n", "\n"),
						SourceURL:  sourceURL,
					})
				}
				return out, nil
			}
		}
	}

	return loadTheorySubtopicsFromFS()
}

func loadTheorySubtopicsFromFS() ([]TheorySubtopicContent, error) {
	emb := getEmbeddedTheoryFS()
	out := make([]TheorySubtopicContent, 0, len(OrderedTheorySubtopics))

	for _, spec := range OrderedTheorySubtopics {
		raw, err := readTheoryMarkdown(emb, spec.Filename)
		if err != nil {
			return nil, fmt.Errorf("failed to load subtopic %s: %w", spec.SubTopicID, err)
		}
		out = append(out, TheorySubtopicContent{
			SubTopicID: spec.SubTopicID,
			DomainID:   spec.DomainID,
			OrderIndex: spec.OrderIndex,
			Title:      spec.Title,
			Content:    strings.ReplaceAll(raw, "\r\n", "\n"),
			SourceURL:  spec.SourceURL,
		})
	}
	return out, nil
}

// ContainsThaiRune reports whether s contains any Thai Unicode code point (U+0E00..U+0E7F).
func ContainsThaiRune(s string) bool {
	for _, r := range s {
		if unicode.In(r, unicode.Thai) || (r >= 0x0E00 && r <= 0x0E7F) {
			return true
		}
	}
	return false
}

// NormalizeLocale maps supported locale identifiers to canonical "en" or "th" (defaulting to "en").
func NormalizeLocale(raw string) string {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	if trimmed == "th" || strings.HasPrefix(trimmed, "th-") || strings.HasPrefix(trimmed, "th_") {
		return "th"
	}
	return "en"
}

// ParseAcceptLanguage parses an HTTP Accept-Language header and returns the highest-priority
// supported locale ("en" or "th"), returning ok=false when neither supported locale matches.
func ParseAcceptLanguage(header string) (string, bool) {
	if strings.TrimSpace(header) == "" {
		return "", false
	}
	type langPref struct {
		locale string
		q      float64
		idx    int
	}
	var prefs []langPref
	for i, part := range strings.Split(header, ",") {
		seg := strings.TrimSpace(part)
		if seg == "" {
			continue
		}
		pieces := strings.Split(seg, ";")
		tag := strings.ToLower(strings.TrimSpace(pieces[0]))
		var locale string
		switch {
		case tag == "th" || strings.HasPrefix(tag, "th-"):
			locale = "th"
		case tag == "en" || strings.HasPrefix(tag, "en-"):
			locale = "en"
		default:
			continue
		}
		q := 1.0
		for _, param := range pieces[1:] {
			kv := strings.SplitN(strings.TrimSpace(param), "=", 2)
			if len(kv) == 2 && strings.EqualFold(strings.TrimSpace(kv[0]), "q") {
				if parsed, err := strconv.ParseFloat(strings.TrimSpace(kv[1]), 64); err == nil {
					q = parsed
				}
			}
		}
		if q <= 0 {
			continue
		}
		prefs = append(prefs, langPref{locale: locale, q: q, idx: i})
	}
	if len(prefs) == 0 {
		return "", false
	}
	sort.SliceStable(prefs, func(i, j int) bool {
		if prefs[i].q != prefs[j].q {
			return prefs[i].q > prefs[j].q
		}
		return prefs[i].idx < prefs[j].idx
	})
	return prefs[0].locale, true
}

// ExtractMarkdownHeadingTitle extracts the top-level `# Title` from Markdown or returns fallback.
func ExtractMarkdownHeadingTitle(content, fallback string) string {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	for _, line := range strings.Split(strings.TrimLeft(normalized, " \t\r\n"), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			heading := strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
			if heading != "" {
				if len(heading) > 255 {
					return heading[:255]
				}
				return heading
			}
			break
		}
	}
	return fallback
}

var (
	markdownHTTPSLinkPattern = regexp.MustCompile(`\]\((https://[^)\s]+)\)`)
	markdownCodeBlockPattern = regexp.MustCompile("(?s)```.*?```")
)

func extractMarkdownHTTPSLinks(content string) map[string]bool {
	links := make(map[string]bool)
	for _, match := range markdownHTTPSLinkPattern.FindAllStringSubmatch(content, -1) {
		if len(match) >= 2 {
			links[match[1]] = true
		}
	}
	return links
}

func validateThaiTheoryCommon(id, enContent, thContent string, sectionCount int) error {
	normalized := strings.ReplaceAll(thContent, "\r\n", "\n")
	if strings.TrimSpace(normalized) == "" {
		return fmt.Errorf("%s: Thai theory content is empty", id)
	}
	if !ContainsThaiRune(normalized) {
		return fmt.Errorf("%s: Thai theory content must contain Thai script", id)
	}
	for i := 1; i <= sectionCount; i++ {
		secPattern := regexp.MustCompile(fmt.Sprintf(`(?m)^##\s+%d\.\s+`, i))
		if !secPattern.MatchString(normalized) {
			return fmt.Errorf("%s: missing required Thai section %d", id, i)
		}
	}
	for _, block := range markdownCodeBlockPattern.FindAllString(normalized, -1) {
		if ContainsThaiRune(block) {
			return fmt.Errorf("%s: code blocks must remain strictly English (found Thai script)", id)
		}
	}
	enLinks := extractMarkdownHTTPSLinks(enContent)
	thLinks := extractMarkdownHTTPSLinks(normalized)
	for url := range enLinks {
		if !thLinks[url] {
			return fmt.Errorf("%s: missing https:// citation from EN source: %s", id, url)
		}
	}
	for url := range thLinks {
		if !enLinks[url] {
			return fmt.Errorf("%s: unregistered extra https:// citation not in EN source: %s", id, url)
		}
	}
	return nil
}

// ValidateThaiTheoryModuleContent validates a Thai Domain Overview Markdown document against its EN counterpart.
func ValidateThaiTheoryModuleContent(domainID, enContent, thContent string) error {
	if err := validateThaiTheoryCommon(domainID, enContent, thContent, 4); err != nil {
		return err
	}
	lower := strings.ToLower(thContent)
	if !strings.Contains(lower, "kql") {
		return fmt.Errorf("%s: missing KQL investigation patterns", domainID)
	}
	if !strings.Contains(lower, "true positive") || !strings.Contains(lower, "false positive") {
		return fmt.Errorf("%s: missing True Positive / False Positive triage guidance", domainID)
	}
	return nil
}

// ValidateThaiTheorySubtopicContent validates a Thai Sub-topic Markdown document against its EN counterpart.
func ValidateThaiTheorySubtopicContent(subTopicID, enContent, thContent string) error {
	normalized := strings.ReplaceAll(thContent, "\r\n", "\n")
	trimmed := strings.TrimSpace(normalized)
	if !strings.HasPrefix(trimmed, "# "+subTopicID) {
		return fmt.Errorf("%s: expected top heading starting with '# %s'", subTopicID, subTopicID)
	}
	return validateThaiTheoryCommon(subTopicID, enContent, normalized, 3)
}

func readLocalizedTheoryMarkdown(emb fs.FS, locale, filename string) (string, error) {
	relSlash := filepath.ToSlash(filepath.Join("theory", locale, filename))
	if emb != nil {
		if data, err := fs.ReadFile(emb, relSlash); err == nil {
			return string(data), nil
		}
	}
	candidates := []string{
		filepath.Join("theory", locale, filename),
		filepath.Join("..", "theory", locale, filename),
	}
	for _, path := range candidates {
		if data, err := os.ReadFile(path); err == nil {
			return string(data), nil
		}
	}
	return "", fmt.Errorf("localized theory file %s/%s not found", locale, filename)
}

func loadTheoryTranslationsByType(db *gorm.DB, entityType, locale string) map[string]models.TheoryTranslation {
	out := make(map[string]models.TheoryTranslation)
	if db == nil {
		return out
	}
	var rows []models.TheoryTranslation
	if err := db.Where("entity_type = ? AND locale = ?", entityType, locale).Find(&rows).Error; err != nil {
		return out
	}
	for _, r := range rows {
		if strings.TrimSpace(r.Content) != "" {
			out[r.EntityID] = r
		}
	}
	return out
}

// LoadTheoryModulesByLocale returns the 5 Domain Overviews localized to locale ("en" or "th"),
// preferring theory_translations in DB, then embedded/local theory/<locale>/, then canonical EN.
func LoadTheoryModulesByLocale(db *gorm.DB, locale string) ([]TheoryModuleContent, error) {
	modules, err := LoadTheoryModules(db)
	if err != nil {
		return nil, err
	}
	normLocale := NormalizeLocale(locale)
	if normLocale == "en" {
		return modules, nil
	}

	dbTranslations := loadTheoryTranslationsByType(db, "domain_overview", normLocale)
	emb := getEmbeddedTheoryFS()

	for i := range modules {
		id := modules[i].DomainID
		if tr, ok := dbTranslations[id]; ok {
			if strings.TrimSpace(tr.Title) != "" {
				modules[i].Title = tr.Title
			}
			modules[i].Content = strings.ReplaceAll(tr.Content, "\r\n", "\n")
			continue
		}
		var filename string
		for _, spec := range OrderedTheoryDomains {
			if spec.DomainID == id {
				filename = spec.Filename
				break
			}
		}
		if filename != "" {
			if raw, readErr := readLocalizedTheoryMarkdown(emb, normLocale, filename); readErr == nil && strings.TrimSpace(raw) != "" {
				normalized := strings.ReplaceAll(raw, "\r\n", "\n")
				modules[i].Title = ExtractMarkdownHeadingTitle(normalized, modules[i].Title)
				modules[i].Content = normalized
			}
		}
	}
	return modules, nil
}

// LoadTheorySubtopicsByLocale returns all 28 Sub-topics localized to locale ("en" or "th"),
// preferring theory_translations in DB, then embedded/local theory/<locale>/, then canonical EN.
func LoadTheorySubtopicsByLocale(db *gorm.DB, locale string) ([]TheorySubtopicContent, error) {
	subtopics, err := LoadTheorySubtopics(db)
	if err != nil {
		return nil, err
	}
	normLocale := NormalizeLocale(locale)
	if normLocale == "en" {
		return subtopics, nil
	}

	dbTranslations := loadTheoryTranslationsByType(db, "subtopic", normLocale)
	emb := getEmbeddedTheoryFS()

	for i := range subtopics {
		id := subtopics[i].SubTopicID
		if tr, ok := dbTranslations[id]; ok {
			if strings.TrimSpace(tr.Title) != "" {
				subtopics[i].Title = tr.Title
			}
			subtopics[i].Content = strings.ReplaceAll(tr.Content, "\r\n", "\n")
			continue
		}
		if spec, ok := GetSubtopicSpec(id); ok {
			if raw, readErr := readLocalizedTheoryMarkdown(emb, normLocale, spec.Filename); readErr == nil && strings.TrimSpace(raw) != "" {
				normalized := strings.ReplaceAll(raw, "\r\n", "\n")
				subtopics[i].Title = ExtractMarkdownHeadingTitle(normalized, subtopics[i].Title)
				subtopics[i].Content = normalized
			}
		}
	}
	return subtopics, nil
}
