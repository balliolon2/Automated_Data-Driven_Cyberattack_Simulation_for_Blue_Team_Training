package services

import (
	"encoding/json"
	"log"
	"math"
	"sort"

	"cybersim/models"
)

// Configuration constants for scoring
const (
	ProficiencyThreshold = 70.0 // Uniform threshold across all domains
	MaxScenarios         = 10   // Maximum scenarios per training session
	MinScenarios         = 1    // Minimum scenarios before allowing completion
	WeightOld            = 0.7  // 1 - alpha: Weight for existing proficiency score (0.70)
	WeightNew            = 0.3  // alpha: Weight for new scenario score (0.30)
	ScoreWeightTPFP      = 0.20 // 20% of scenario score (20 max points)
	ScoreWeightFindings  = 0.40 // 40% of scenario score
	ScoreWeightResponse  = 0.40 // 40% of scenario score
	CalculationVersion   = "proficiency-v1"
	RubricVersion        = "rubric-v2-evidence-click"

	// EvidencePenaltyPerLog is the number of points deducted for each unique
	// selected log that proves no Key Finding.
	EvidencePenaltyPerLog = 2
	// DefaultFindingPoints is used when a Key Finding omits its points value.
	DefaultFindingPoints = 10
)

// ScoreSubmission evaluates learner inputs against scenario criteria and computes the complete Evaluation Result.
//
// Key Findings are credited from the server-recorded evidence selection
// (input.SelectedLogIndices). A finding earns its full points when the selected
// set intersects its authored evidence_log_indices. Selecting several logs from
// the same finding does not stack, and a log shared by findings credits each.
// Every unique selected log outside every evidence set costs two points; the
// findings component floors at zero.
func ScoreSubmission(scenario models.Scenario, input models.SubmitScenarioInput) models.ScenarioResult {
	selected := normalizeIndices(input.SelectedLogIndices)
	selectedSet := make(map[int]struct{}, len(selected))
	for _, index := range selected {
		selectedSet[index] = struct{}{}
	}

	result := models.ScenarioResult{
		DomainScores:       make(map[string]models.DomainScoreDetail),
		DomainScoreMap:     make(map[string]float64),
		FindingsDetail:     make([]models.FindingResult, 0),
		ResponseDetail:     make([]models.ResponseActionResult, 0),
		SelectedLogIndices: selected,
		NoiseLogIndices:    make([]int, 0),
	}

	// Track raw points per domain
	domainEarned := make(map[string]float64)
	domainMaxPoints := make(map[string]float64)

	// --- 1. Score TP/FP Decision ---
	result.TPFPCorrect = input.IsTruePositive == scenario.IsTruePositive
	result.TPFPExplanation = scenario.TpFpExplanation
	tpfpMaxPoints := 20.0
	if result.TPFPCorrect {
		result.TPFPPoints = tpfpMaxPoints
		domainEarned[scenario.DomainID] += tpfpMaxPoints
	}
	domainMaxPoints[scenario.DomainID] += tpfpMaxPoints

	// --- 2. Score Key Findings from the selected evidence set ---
	evidenceUnion := make(map[int]struct{})
	expectedOutcomes, ok := scenario.ExpectedOutcomes.(map[string]interface{})
	if ok {
		for _, finding := range mapSlice(expectedOutcomes["key_findings"]) {
			fID, _ := finding["id"].(string)
			fDesc, _ := finding["description"].(string)
			fDomain, _ := finding["domain_id"].(string)
			fExplanation, _ := finding["explanation"].(string)
			fPoints := findingPoints(finding)
			evidenceIndices := decodeEvidenceLogIndices(finding["evidence_log_indices"], fID)

			for _, index := range evidenceIndices {
				evidenceUnion[index] = struct{}{}
			}

			matched := make([]int, 0)
			for _, index := range evidenceIndices {
				if _, selectedHere := selectedSet[index]; selectedHere {
					matched = append(matched, index)
				}
			}
			sort.Ints(matched)

			earned := 0
			if len(matched) > 0 {
				earned = fPoints
			}

			if earned > 0 {
				domainEarned[fDomain] += float64(earned)
			}
			domainMaxPoints[fDomain] += float64(fPoints)

			result.FindingsDetail = append(result.FindingsDetail, models.FindingResult{
				ID:                 fID,
				Description:        fDesc,
				DomainID:           fDomain,
				Found:              earned > 0,
				Points:             fPoints,
				Earned:             earned,
				EvidenceLogIndices: evidenceIndices,
				MatchedLogIndices:  matched,
				Explanation:        fExplanation,
			})
			result.FindingsPoints += float64(earned)
		}
	}

	// Noise is any selected index outside the union of every evidence set.
	noise := make([]int, 0)
	for _, index := range selected {
		if _, isEvidence := evidenceUnion[index]; !isEvidence {
			noise = append(noise, index)
		}
	}
	result.NoiseLogIndices = noise
	result.EvidencePenalty = float64(EvidencePenaltyPerLog * len(noise))
	result.FindingsPoints = math.Max(0, result.FindingsPoints-result.EvidencePenalty)

	// The evidence penalty is attributed to the scenario's primary domain.
	domainEarned[scenario.DomainID] -= result.EvidencePenalty

	// --- 3. Score Response Actions ---
	playbookSteps, ok := scenario.PlaybookSteps.(map[string]interface{})
	if ok {
		selectedSet := make(map[string]bool)
		for _, aID := range input.SelectedActions {
			selectedSet[aID] = true
		}

		phases := []string{"containment", "eradication", "recovery"}
		for _, phase := range phases {
			for _, action := range mapSlice(playbookSteps[phase]) {
				aID, _ := action["id"].(string)
				aLabel, _ := action["label"].(string)
				aCorrect, _ := action["is_correct"].(bool)
				aDomain, _ := action["domain_id"].(string)
				aPoints := ToFloat64(action["points"])
				aExplanation, _ := action["explanation"].(string)

				selected := selectedSet[aID]

				earned := 0
				if selected {
					earned = int(aPoints) // positive if correct, negative if incorrect
					result.ResponsePoints += aPoints
					domainEarned[aDomain] += aPoints
				}

				// For max points calculation, only count positive points
				if aPoints > 0 {
					domainMaxPoints[aDomain] += aPoints
				}

				result.ResponseDetail = append(result.ResponseDetail, models.ResponseActionResult{
					ID:          aID,
					Label:       aLabel,
					Phase:       phase,
					DomainID:    aDomain,
					Selected:    selected,
					IsCorrect:   aCorrect,
					Points:      int(aPoints),
					Earned:      earned,
					Explanation: aExplanation,
				})
			}
		}
	}

	// --- Calculate per-domain percentage scores ---
	for domainID, maxPts := range domainMaxPoints {
		if maxPts <= 0 {
			continue
		}
		earned := domainEarned[domainID]
		if earned < 0 {
			earned = 0 // Floor at 0
		}
		pct := math.Round((earned/maxPts)*100*100) / 100
		if pct > 100 {
			pct = 100
		}
		result.DomainScoreMap[domainID] = pct
		result.DomainScores[domainID] = models.DomainScoreDetail{
			Earned:     earned,
			MaxPoints:  maxPts,
			Percentage: pct,
		}
	}

	// --- Calculate total weighted score ---
	// The total is composed from the scored components, not from the diagnostic
	// per-domain earned values: the findings penalty is already floored inside
	// FindingsPoints, and domain earned points floor independently.
	totalMax := 0.0
	for _, maxPts := range domainMaxPoints {
		totalMax += maxPts
	}
	totalEarned := result.TPFPPoints + result.FindingsPoints + result.ResponsePoints
	if totalEarned < 0 {
		totalEarned = 0
	}
	if totalMax > 0 {
		pct := math.Round((totalEarned/totalMax)*100*100) / 100
		if pct < 0 {
			pct = 0
		}
		if pct > 100 {
			pct = 100
		}
		result.TotalScore = pct
	}

	return result
}

// ReconstructFromActions reconstructs the complete ScenarioResult from historical SessionActions.
//
// New sessions replay select_evidence actions. Legacy sessions that only carry
// submit_decision.discovered_findings keep their old score by mapping each found
// finding ID to its first authored evidence index and scoring through the same
// evaluator. Legacy finding IDs never produce a noise penalty.
func ReconstructFromActions(scenario models.Scenario, actions []models.SessionAction) models.ScenarioResult {
	var userChoice bool
	var selectedActions []string
	var discoveredFindings []string
	hasSelectionActions := false

	for _, a := range actions {
		if a.ActionType == SelectEvidenceActionType {
			hasSelectionActions = true
		}

		payload, ok := a.Payload.(map[string]interface{})
		if !ok {
			continue
		}

		switch a.ActionType {
		case "triage_alert":
			if uc, exists := payload["user_choice"].(bool); exists {
				userChoice = uc
			}
		case "respond":
			if saList, exists := payload["selected_actions"].([]interface{}); exists {
				for _, item := range saList {
					if s, ok := item.(string); ok {
						selectedActions = append(selectedActions, s)
					}
				}
			} else if saList, exists := payload["selected_actions"].([]string); exists {
				selectedActions = append(selectedActions, saList...)
			}
		case "submit_decision":
			if dfList, exists := payload["discovered_findings"].([]interface{}); exists {
				for _, item := range dfList {
					if s, ok := item.(string); ok {
						discoveredFindings = append(discoveredFindings, s)
					}
				}
			} else if dfList, exists := payload["discovered_findings"].([]string); exists {
				discoveredFindings = append(discoveredFindings, dfList...)
			}
		}
	}

	selectedLogIndices := ReplaySelectedLogIndices(actions)
	if !hasSelectionActions && len(discoveredFindings) > 0 {
		selectedLogIndices = firstEvidenceIndicesForFindings(scenario, discoveredFindings)
	}

	input := models.SubmitScenarioInput{
		SessionID:          "",
		IsTruePositive:     userChoice,
		TPFPSelected:       true,
		SelectedActions:    selectedActions,
		SelectedLogIndices: selectedLogIndices,
	}

	return ScoreSubmission(scenario, input)
}

// CalculateUpdatedProficiency computes the new weighted average domain proficiency
func CalculateUpdatedProficiency(oldScore float64, newScore float64) float64 {
	updated := WeightOld*oldScore + WeightNew*newScore
	return math.Round(updated*100) / 100
}

// ToFloat64 safely converts an interface{} numeric type to float64
func ToFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case json.Number:
		parsed, err := val.Float64()
		if err == nil {
			return parsed
		}
		return 0
	default:
		return 0
	}
}

// normalizeIndices sorts and de-duplicates an evidence selection.
func normalizeIndices(indices []int) []int {
	normalized := make([]int, 0, len(indices))
	seen := make(map[int]struct{}, len(indices))
	for _, index := range indices {
		if _, exists := seen[index]; exists {
			continue
		}
		seen[index] = struct{}{}
		normalized = append(normalized, index)
	}
	sort.Ints(normalized)
	return normalized
}

// mapSlice normalizes JSONB and in-memory object collections to a map slice.
func mapSlice(value any) []map[string]interface{} {
	switch typed := value.(type) {
	case nil:
		return nil
	case []map[string]interface{}:
		return typed
	case []interface{}:
		out := make([]map[string]interface{}, 0, len(typed))
		for _, item := range typed {
			if entry, ok := item.(map[string]interface{}); ok {
				out = append(out, entry)
			}
		}
		return out
	default:
		return nil
	}
}

// decodeEvidenceLogIndices defensively decodes an authored JSONB evidence array
// into unique, sorted 0-based log indices. It accepts int, float and generic
// JSONB shapes. Fractional, negative, or non-numeric entries are skipped with a
// warning and are never silently truncated.
func decodeEvidenceLogIndices(raw any, findingID string) []int {
	indices := make([]int, 0)

	appendIndex := func(value any) {
		index, ok := evidenceIndexFromAny(value)
		if !ok {
			log.Printf("warning: ignoring invalid evidence_log_indices entry %v for finding %q", value, findingID)
			return
		}
		if index < 0 {
			log.Printf("warning: ignoring negative evidence_log_indices entry %d for finding %q", index, findingID)
			return
		}
		indices = append(indices, index)
	}

	switch values := raw.(type) {
	case nil:
		return []int{}
	case []int:
		for _, value := range values {
			appendIndex(value)
		}
	case []int64:
		for _, value := range values {
			appendIndex(value)
		}
	case []float64:
		for _, value := range values {
			appendIndex(value)
		}
	case []json.Number:
		for _, value := range values {
			appendIndex(value)
		}
	case []interface{}:
		for _, value := range values {
			appendIndex(value)
		}
	default:
		log.Printf("warning: unsupported evidence_log_indices type %T for finding %q", raw, findingID)
		return []int{}
	}

	return normalizeIndices(indices)
}

func evidenceIndexFromAny(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int8:
		return int(typed), true
	case int16:
		return int(typed), true
	case int32:
		return int(typed), true
	case int64:
		return int(typed), true
	case uint:
		return int(typed), true
	case uint8:
		return int(typed), true
	case uint16:
		return int(typed), true
	case uint32:
		return int(typed), true
	case uint64:
		if typed > math.MaxInt64 {
			return 0, false
		}
		return int(typed), true
	case float32:
		return integralIndex(float64(typed))
	case float64:
		return integralIndex(typed)
	case json.Number:
		if parsed, err := typed.Int64(); err == nil {
			return int(parsed), true
		}
		parsed, err := typed.Float64()
		if err != nil {
			return 0, false
		}
		return integralIndex(parsed)
	default:
		return 0, false
	}
}

func integralIndex(value float64) (int, bool) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value != math.Trunc(value) {
		return 0, false
	}
	if value < 0 || value > float64(math.MaxInt32) {
		return 0, false
	}
	return int(value), true
}

// findingPoints returns the configured points for a Key Finding. Missing or
// invalid points fall back to the documented default.
func findingPoints(finding map[string]interface{}) int {
	raw, exists := finding["points"]
	if !exists || raw == nil {
		return DefaultFindingPoints
	}
	points := ToFloat64(raw)
	if points <= 0 || points != math.Trunc(points) {
		log.Printf("warning: invalid points %v for key finding; defaulting to %d", raw, DefaultFindingPoints)
		return DefaultFindingPoints
	}
	return int(points)
}

// firstEvidenceIndicesForFindings maps legacy discovered finding IDs to the
// first authored evidence index of each finding.
func firstEvidenceIndicesForFindings(scenario models.Scenario, findingIDs []string) []int {
	expectedOutcomes, ok := scenario.ExpectedOutcomes.(map[string]interface{})
	if !ok {
		return []int{}
	}
	byID := make(map[string][]int)
	for _, finding := range mapSlice(expectedOutcomes["key_findings"]) {
		fID, _ := finding["id"].(string)
		if _, exists := byID[fID]; exists {
			continue
		}
		byID[fID] = decodeEvidenceLogIndices(finding["evidence_log_indices"], fID)
	}

	indices := make([]int, 0, len(findingIDs))
	for _, findingID := range findingIDs {
		evidence, exists := byID[findingID]
		if !exists || len(evidence) == 0 {
			continue
		}
		indices = append(indices, evidence[0])
	}
	return normalizeIndices(indices)
}
