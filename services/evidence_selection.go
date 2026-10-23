package services

import (
	"encoding/json"
	"math"
	"sort"

	"cybersim/models"
)

// SelectEvidenceActionType is the session action type that records evidence
// selection toggles and clears for a simulation session.
const SelectEvidenceActionType = "select_evidence"

// ReplaySelectedLogIndices derives the final selected evidence log index set by
// replaying select_evidence actions in step order. Malformed payloads are
// skipped so a single bad record cannot invalidate the whole selection.
//
// The result is sorted ascending and is always non-nil, which keeps the JSON
// contract as an empty array rather than null for a fresh session.
func ReplaySelectedLogIndices(actions []models.SessionAction) []int {
	selected := make(map[int]struct{})

	ordered := make([]models.SessionAction, len(actions))
	copy(ordered, actions)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].StepOrder < ordered[j].StepOrder
	})

	for _, action := range ordered {
		if action.ActionType != SelectEvidenceActionType {
			continue
		}

		payload, ok := selectionPayloadMap(action.Payload)
		if !ok {
			continue
		}

		operation, ok := payload["operation"].(string)
		if !ok {
			continue
		}

		switch operation {
		case "clear":
			selected = make(map[int]struct{})
		case "toggle":
			index, ok := selectionLogIndex(payload["log_index"])
			if !ok || index < 0 {
				continue
			}
			value, ok := payload["selected"].(bool)
			if !ok {
				continue
			}
			if value {
				selected[index] = struct{}{}
			} else {
				delete(selected, index)
			}
		}
	}

	indices := make([]int, 0, len(selected))
	for index := range selected {
		indices = append(indices, index)
	}
	sort.Ints(indices)
	return indices
}

func selectionPayloadMap(payload any) (map[string]interface{}, bool) {
	switch value := payload.(type) {
	case map[string]interface{}:
		if value == nil {
			return nil, false
		}
		return value, true
	case []byte:
		var decoded map[string]interface{}
		if err := json.Unmarshal(value, &decoded); err != nil || decoded == nil {
			return nil, false
		}
		return decoded, true
	case json.RawMessage:
		var decoded map[string]interface{}
		if err := json.Unmarshal(value, &decoded); err != nil || decoded == nil {
			return nil, false
		}
		return decoded, true
	case string:
		var decoded map[string]interface{}
		if err := json.Unmarshal([]byte(value), &decoded); err != nil || decoded == nil {
			return nil, false
		}
		return decoded, true
	default:
		return nil, false
	}
}

func selectionLogIndex(value any) (int, bool) {
	switch index := value.(type) {
	case int:
		return index, true
	case int8:
		return int(index), true
	case int16:
		return int(index), true
	case int32:
		return int(index), true
	case int64:
		return int(index), true
	case float32:
		return integralFloatToInt(float64(index))
	case float64:
		return integralFloatToInt(index)
	case json.Number:
		parsed, err := index.Int64()
		if err != nil {
			return 0, false
		}
		return int(parsed), true
	default:
		return 0, false
	}
}

func integralFloatToInt(value float64) (int, bool) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value != math.Trunc(value) {
		return 0, false
	}
	if value < -9.007199254740992e15 || value > 9.007199254740992e15 {
		return 0, false
	}
	return int(value), true
}
