package services

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

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

const theorySourceRefURL = "docs/REFERENCES.md#14-learning-theory-curriculum-provenance-r2r5"

var OrderedTheoryDomains = []TheoryDomainSpec{
	{
		DomainID:   "domain1",
		DomainName: "General Security Concepts",
		Title:      "Domain 1: General Security Concepts",
		Filename:   "domain1_security_concepts.md",
		SourceURL:  theorySourceRefURL,
	},
	{
		DomainID:   "domain2",
		DomainName: "Threats, Vulnerabilities, and Mitigations",
		Title:      "Domain 2: Threats, Vulnerabilities, and Mitigations",
		Filename:   "domain2_threats_vulnerabilities.md",
		SourceURL:  theorySourceRefURL,
	},
	{
		DomainID:   "domain3",
		DomainName: "Security Architecture",
		Title:      "Domain 3: Security Architecture",
		Filename:   "domain3_security_architecture.md",
		SourceURL:  theorySourceRefURL,
	},
	{
		DomainID:   "domain4",
		DomainName: "Security Operations",
		Title:      "Domain 4: Security Operations",
		Filename:   "domain4_security_operations.md",
		SourceURL:  theorySourceRefURL,
	},
	{
		DomainID:   "domain5",
		DomainName: "Security Program Management and Oversight",
		Title:      "Domain 5: Security Program Management and Oversight",
		Filename:   "domain5_security_management.md",
		SourceURL:  theorySourceRefURL,
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
					if sourceURL == "" {
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
