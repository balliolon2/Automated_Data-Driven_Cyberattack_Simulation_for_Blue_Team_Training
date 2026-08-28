package services

import (
	"math"

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
	RubricVersion        = "rubric-v1"
)

// ScoreSubmission evaluates learner inputs against scenario criteria and computes the complete Evaluation Result
func ScoreSubmission(scenario models.Scenario, input models.SubmitScenarioInput) models.ScenarioResult {
	result := models.ScenarioResult{
		DomainScores:   make(map[string]models.DomainScoreDetail),
		DomainScoreMap: make(map[string]float64),
		FindingsDetail: make([]models.FindingResult, 0),
		ResponseDetail: make([]models.ResponseActionResult, 0),
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

	// --- 2. Score Key Findings ---
	expectedOutcomes, ok := scenario.ExpectedOutcomes.(map[string]interface{})
	if ok {
		keyFindings, ok := expectedOutcomes["key_findings"].([]interface{})
		if ok {
			discoveredSet := make(map[string]bool)
			for _, fID := range input.DiscoveredFindings {
				discoveredSet[fID] = true
			}

			for _, kf := range keyFindings {
				finding, ok := kf.(map[string]interface{})
				if !ok {
					continue
				}

				fID, _ := finding["id"].(string)
				fDesc, _ := finding["description"].(string)
				fDomain, _ := finding["domain_id"].(string)
				fPoints := ToFloat64(finding["points"])
				fExplanation, _ := finding["explanation"].(string)

				found := discoveredSet[fID]

				fr := models.FindingResult{
					ID:          fID,
					Description: fDesc,
					DomainID:    fDomain,
					Found:       found,
					Points:      int(fPoints),
					Explanation: fExplanation,
				}

				if found {
					result.FindingsPoints += fPoints
					domainEarned[fDomain] += fPoints
				}
				domainMaxPoints[fDomain] += fPoints
				result.FindingsDetail = append(result.FindingsDetail, fr)
			}
		}
	}

	// --- 3. Score Response Actions ---
	playbookSteps, ok := scenario.PlaybookSteps.(map[string]interface{})
	if ok {
		selectedSet := make(map[string]bool)
		for _, aID := range input.SelectedActions {
			selectedSet[aID] = true
		}

		phases := []string{"containment", "eradication", "recovery"}
		for _, phase := range phases {
			actions, ok := playbookSteps[phase].([]interface{})
			if !ok {
				continue
			}

			for _, a := range actions {
				action, ok := a.(map[string]interface{})
				if !ok {
					continue
				}

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
	totalMax := 0.0
	totalEarned := 0.0
	for _, maxPts := range domainMaxPoints {
		totalMax += maxPts
	}
	for _, earned := range domainEarned {
		totalEarned += earned
	}
	if totalEarned < 0 {
		totalEarned = 0
	}
	if totalMax > 0 {
		result.TotalScore = math.Round((totalEarned/totalMax)*100*100) / 100
	}

	return result
}

// ReconstructFromActions reconstructs the complete ScenarioResult from historical SessionActions (for backward compatibility)
func ReconstructFromActions(scenario models.Scenario, actions []models.SessionAction) models.ScenarioResult {
	var userChoice bool
	var selectedActions []string
	var discoveredFindings []string

	for _, a := range actions {
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
			}
		case "submit_decision":
			if dfList, exists := payload["discovered_findings"].([]interface{}); exists {
				for _, item := range dfList {
					if s, ok := item.(string); ok {
						discoveredFindings = append(discoveredFindings, s)
					}
				}
			}
		}
	}

	input := models.SubmitScenarioInput{
		SessionID:          "",
		IsTruePositive:     userChoice,
		TPFPSelected:       true,
		SelectedActions:    selectedActions,
		DiscoveredFindings: discoveredFindings,
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
	case int:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return 0
	}
}
