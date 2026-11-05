package services

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"cybersim/models"
)

// EvidenceLogSnippet is the Specialist-facing view of one telemetry row. The
// index is the row's 0-based position in the Scenario's original log array.
type EvidenceLogSnippet struct {
	Index     int
	Timestamp string
	Source    string
	Severity  string
	Message   string
}

// EvidenceFindingReview describes whether the learner's selection proved a Key
// Finding and which authored evidence rows were acceptable or actually matched.
// Earned is the gross per-finding credit, before the evidence penalty.
type EvidenceFindingReview struct {
	ID                 string
	Description        string
	DomainID           string
	Found              bool
	Points             int
	Earned             int
	EvidenceLogIndices []int
	MatchedLogIndices  []int
	EvidenceLogs       []EvidenceLogSnippet
	Explanation        string
}

// EvidenceReview is the pure result of reviewing a selection against a
// Scenario's authored Key Findings and evidence sets.
//
// Legacy is true only for sessions that predate the evidence-click rubric. A
// new-rubric session that submitted an empty selection still produces a
// non-legacy review with zero counts and zero penalty. Legacy sessions expose
// the submit_decision.discovered_findings payload as-is and are never
// re-scored.
type EvidenceReview struct {
	Legacy                   bool
	LegacyDiscoveredFindings []string
	SelectedLogIndices       []int
	EvidenceFindings         []EvidenceFindingReview
	NoiseSelections          []EvidenceLogSnippet
	FindingsFoundCount       int
	TotalFindingsCount       int
	EvidenceSelectedCount    int
	NoiseSelectedCount       int
	GrossFindingsPoints      float64
	FindingsPoints           float64
	EvidencePenalty          float64
}

// BuildEvidenceReview computes the per-finding evidence trail for a selected
// evidence set. A finding is credited in full when the selection intersects its
// authored evidence indices; shared indices credit every finding that lists
// them and are never noise. Every selected index outside every evidence set is
// noise and costs EvidencePenaltyPerLog points.
func BuildEvidenceReview(scenario models.Scenario, selectedLogIndices []int) EvidenceReview {
	selected := normalizeIndices(selectedLogIndices)
	selectedSet := make(map[int]struct{}, len(selected))
	for _, index := range selected {
		selectedSet[index] = struct{}{}
	}

	findings, totalFindings := expectedReviewFindings(scenario)
	logs := mapSlice(scenario.InitialLogs)

	review := EvidenceReview{
		SelectedLogIndices:    selected,
		EvidenceFindings:      make([]EvidenceFindingReview, 0, len(findings)),
		NoiseSelections:       make([]EvidenceLogSnippet, 0),
		TotalFindingsCount:    totalFindings,
		EvidenceSelectedCount: len(selected),
	}

	evidenceUnion := make(map[int]struct{})

	for _, finding := range findings {
		findingID, _ := finding["id"].(string)
		description, _ := finding["description"].(string)
		domainID, _ := finding["domain_id"].(string)
		explanation, _ := finding["explanation"].(string)
		points := findingPoints(finding)
		evidenceIndices := decodeEvidenceLogIndices(finding["evidence_log_indices"], findingID)

		evidenceLogs := make([]EvidenceLogSnippet, 0, len(evidenceIndices))
		matched := make([]int, 0, len(evidenceIndices))
		for _, index := range evidenceIndices {
			evidenceUnion[index] = struct{}{}
			evidenceLogs = append(evidenceLogs, buildEvidenceLogSnippet(logs, index))
			if _, selectedHere := selectedSet[index]; selectedHere {
				matched = append(matched, index)
			}
		}
		sort.Ints(matched)

		found := len(matched) > 0
		earned := 0
		if found {
			earned = points
			review.FindingsFoundCount++
		}
		review.GrossFindingsPoints += float64(earned)

		review.EvidenceFindings = append(review.EvidenceFindings, EvidenceFindingReview{
			ID:                 findingID,
			Description:        description,
			DomainID:           domainID,
			Found:              found,
			Points:             points,
			Earned:             earned,
			EvidenceLogIndices: evidenceIndices,
			MatchedLogIndices:  matched,
			EvidenceLogs:       evidenceLogs,
			Explanation:        explanation,
		})
	}

	for _, index := range selected {
		if _, isEvidence := evidenceUnion[index]; isEvidence {
			continue
		}
		review.NoiseSelections = append(review.NoiseSelections, buildEvidenceLogSnippet(logs, index))
	}
	review.NoiseSelectedCount = len(review.NoiseSelections)
	review.EvidencePenalty = float64(EvidencePenaltyPerLog * review.NoiseSelectedCount)
	review.FindingsPoints = math.Max(0, review.GrossFindingsPoints-review.EvidencePenalty)

	return review
}

// BuildEvidenceReviewFromActions replays a session's select_evidence actions.
// A new-rubric session is detected from the replay actions alone: either a
// select_evidence action exists, or the submit_decision payload carries the
// selected_log_indices audit field (which is present even when the learner
// selected nothing). Legacy sessions fall back to the
// submit_decision.discovered_findings payload and are never re-scored.
func BuildEvidenceReviewFromActions(scenario models.Scenario, actions []models.SessionAction) EvidenceReview {
	return buildEvidenceReviewFromActions(scenario, "", actions)
}

// BuildEvidenceReviewForSession is BuildEvidenceReviewFromActions with the
// session's recorded RubricVersion as a second new-rubric signal. The submitted
// action payload remains the fallback for sessions persisted before the rubric
// version was written.
func BuildEvidenceReviewForSession(scenario models.Scenario, session models.SimulationSession, actions []models.SessionAction) EvidenceReview {
	return buildEvidenceReviewFromActions(scenario, session.RubricVersion, actions)
}

func buildEvidenceReviewFromActions(scenario models.Scenario, rubricVersion string, actions []models.SessionAction) EvidenceReview {
	if isEvidenceRubricSession(rubricVersion, actions) {
		return BuildEvidenceReview(scenario, ReplaySelectedLogIndices(actions))
	}

	discovered := discoveredFindingsFromActions(actions)
	_, totalFindings := expectedReviewFindings(scenario)

	return EvidenceReview{
		Legacy:                   true,
		LegacyDiscoveredFindings: discovered,
		FindingsFoundCount:       len(discovered),
		TotalFindingsCount:       totalFindings,
	}
}

// isEvidenceRubricSession reports whether the attempt was scored against the
// evidence-click rubric. A v2 session that submitted an empty selection has no
// select_evidence actions, so its submit_decision audit payload is the
// discriminator. An explicit RubricVersion match also covers replayed sessions
// whose action rows were pruned.
func isEvidenceRubricSession(rubricVersion string, actions []models.SessionAction) bool {
	if rubricVersion == RubricVersion {
		return true
	}

	for _, action := range actions {
		if action.ActionType == SelectEvidenceActionType {
			return true
		}
		if action.ActionType != "submit_decision" {
			continue
		}
		payload, ok := selectionPayloadMap(action.Payload)
		if !ok {
			continue
		}
		if _, recorded := payload["selected_log_indices"]; recorded {
			return true
		}
	}

	return false
}

// ScenarioResultFromRecorded decodes the JSONB SimulationSession.SkillGap
// payload into the ScenarioResult captured at submission time. It returns false
// when no meaningful recorded result exists, so callers can fall back.
func ScenarioResultFromRecorded(value any) (models.ScenarioResult, bool) {
	var result models.ScenarioResult

	switch recorded := value.(type) {
	case nil:
		return result, false
	case models.ScenarioResult:
		result = recorded
	case *models.ScenarioResult:
		if recorded == nil {
			return result, false
		}
		result = *recorded
	case []byte:
		if err := json.Unmarshal(recorded, &result); err != nil {
			return result, false
		}
	case json.RawMessage:
		if err := json.Unmarshal(recorded, &result); err != nil {
			return result, false
		}
	default:
		encoded, err := json.Marshal(recorded)
		if err != nil {
			return result, false
		}
		if err := json.Unmarshal(encoded, &result); err != nil {
			return result, false
		}
	}

	if result.TotalScore == 0 &&
		result.FindingsPoints == 0 &&
		result.EvidencePenalty == 0 &&
		len(result.FindingsDetail) == 0 &&
		len(result.ResponseDetail) == 0 &&
		len(result.SelectedLogIndices) == 0 &&
		len(result.NoiseLogIndices) == 0 {
		return result, false
	}

	return result, true
}

// EvidenceReviewFromResult exposes the recorded ScenarioResult as read-only
// review counters. It never re-scores an attempt.
func EvidenceReviewFromResult(result models.ScenarioResult) EvidenceReview {
	selected := normalizeIndices(result.SelectedLogIndices)
	noise := normalizeIndices(result.NoiseLogIndices)

	findings := make([]EvidenceFindingReview, 0, len(result.FindingsDetail))
	foundCount := 0
	grossPoints := 0.0
	for _, detail := range result.FindingsDetail {
		if detail.Found {
			foundCount++
		}
		grossPoints += float64(detail.Earned)
		findings = append(findings, EvidenceFindingReview{
			ID:                 detail.ID,
			Description:        detail.Description,
			DomainID:           detail.DomainID,
			Found:              detail.Found,
			Points:             detail.Points,
			Earned:             detail.Earned,
			EvidenceLogIndices: normalizeIndices(detail.EvidenceLogIndices),
			MatchedLogIndices:  normalizeIndices(detail.MatchedLogIndices),
			Explanation:        detail.Explanation,
		})
	}

	return EvidenceReview{
		SelectedLogIndices:    selected,
		EvidenceFindings:      findings,
		FindingsFoundCount:    foundCount,
		TotalFindingsCount:    len(result.FindingsDetail),
		EvidenceSelectedCount: len(selected),
		NoiseSelectedCount:    len(noise),
		GrossFindingsPoints:   grossPoints,
		FindingsPoints:        result.FindingsPoints,
		EvidencePenalty:       result.EvidencePenalty,
	}
}

// expectedReviewFindings normalizes the Scenario's expected outcomes. The
// current contract stores {"key_findings": [...]}; a legacy list payload is
// still counted for display without inventing evidence mappings.
func expectedReviewFindings(scenario models.Scenario) ([]map[string]interface{}, int) {
	switch outcomes := scenario.ExpectedOutcomes.(type) {
	case map[string]interface{}:
		findings := mapSlice(outcomes["key_findings"])
		return findings, len(findings)
	case []map[string]interface{}:
		return nil, len(outcomes)
	case []interface{}:
		return nil, len(outcomes)
	default:
		return nil, 0
	}
}

func discoveredFindingsFromActions(actions []models.SessionAction) []string {
	discovered := make([]string, 0)
	for _, action := range actions {
		if action.ActionType != "submit_decision" {
			continue
		}
		payload, ok := selectionPayloadMap(action.Payload)
		if !ok {
			continue
		}
		switch values := payload["discovered_findings"].(type) {
		case []interface{}:
			for _, value := range values {
				if id, ok := value.(string); ok {
					discovered = append(discovered, id)
				}
			}
		case []string:
			discovered = append(discovered, values...)
		}
	}
	return discovered
}

func buildEvidenceLogSnippet(logs []map[string]interface{}, index int) EvidenceLogSnippet {
	snippet := EvidenceLogSnippet{Index: index}
	if index < 0 || index >= len(logs) {
		return snippet
	}
	entry := logs[index]
	snippet.Timestamp = evidenceLogField(entry, "timestamp")
	snippet.Source = evidenceLogField(entry, "source")
	snippet.Severity = evidenceLogField(entry, "severity")
	snippet.Message = evidenceLogField(entry, "message")
	return snippet
}

func evidenceLogField(log map[string]interface{}, key string) string {
	value, ok := log[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case time.Time:
		return typed.Format(time.RFC3339)
	default:
		return fmt.Sprint(typed)
	}
}
