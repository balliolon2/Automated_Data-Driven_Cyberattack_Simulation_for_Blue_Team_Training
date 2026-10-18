package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"cybersim/models"
	"gorm.io/gorm"
)

// OpenAI-compatible Chat Completion API request/response structs
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type aiResponseActionOption struct {
	Label       string `json:"label"`
	IsCorrect   bool   `json:"is_correct"`
	Explanation string `json:"explanation"`
}

type aiMicroScenario struct {
	Title             string                   `json:"title"`
	Description       string                   `json:"description"`
	IsTruePositive    bool                     `json:"is_true_positive"`
	TpFpExplanation   string                   `json:"tp_fp_explanation"`
	AttackerIP        string                   `json:"attacker_ip"`
	TargetHost        string                   `json:"target_host"`
	AttackLogMessage  string                   `json:"attack_log_message"`
	KeyFinding        string                   `json:"key_finding"`
	ContainmentAction string                   `json:"containment_action,omitempty"`
	EradicationAction string                   `json:"eradication_action,omitempty"`
	RecoveryAction    string                   `json:"recovery_action,omitempty"`
	ContainmentSteps  []aiResponseActionOption `json:"containment_steps,omitempty"`
	EradicationSteps  []aiResponseActionOption `json:"eradication_steps,omitempty"`
	RecoverySteps     []aiResponseActionOption `json:"recovery_steps,omitempty"`
}

var bannedBoilerplateLabels = map[string]bool{
	"power off all domain controllers immediately": true,
	"reformat all network switches in the subnet":  true,
	"disable mfa enforcement for affected users":   true,
}

// IsBannedBoilerplateLabel checks whether a Response Action label matches any legacy boilerplate distractor.
func IsBannedBoilerplateLabel(label string) bool {
	normalized := strings.ToLower(strings.TrimSpace(label))
	return bannedBoilerplateLabels[normalized]
}

// GenerateLLMScenario calls an OpenAI-compatible API to dynamically synthesize a SOC scenario based on learner skill gaps.
// ponytail: Standard library net/http only, zero heavy external LLM SDK dependencies.
func GenerateLLMScenario(ctx context.Context, weakestDomain string, domainName string, proficiency float64, profiles []models.UserSkillProfile) (*models.Scenario, error) {
	rawApiKey := os.Getenv("LLM_API_KEY")
	if rawApiKey == "" {
		return nil, fmt.Errorf("LLM_API_KEY is not configured")
	}
	cleanKey := strings.TrimSpace(strings.Trim(rawApiKey, "\"'"))

	baseURL := os.Getenv("LLM_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	baseURL = strings.TrimSpace(strings.Trim(baseURL, "\"'"))
	baseURL = strings.TrimSuffix(baseURL, "/")

	var url string
	if strings.HasSuffix(baseURL, "/chat/completions") {
		url = baseURL
	} else {
		url = fmt.Sprintf("%s/chat/completions", baseURL)
	}

	model := os.Getenv("LLM_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}
	model = strings.TrimSpace(strings.Trim(model, "\"'"))

	// Build user profile summary for RAG context
	var profileSummary strings.Builder
	for _, p := range profiles {
		profileSummary.WriteString(fmt.Sprintf("- Domain %s: %.1f%% proficiency\n", p.DomainID, p.ProficiencyScore))
	}

	systemPrompt := "You are a specialized CompTIA Security+ 701 cybersecurity scenario generator. Output ONLY raw JSON. Every incident response phase (containment_steps, eradication_steps, recovery_steps) MUST contain 3 realistic SOC options (1 correct option and 2 plausible technical distractors) with matching sentence length (15-30 words) and specific entity references."
	userPrompt := fmt.Sprintf(`Generate a unique SOC attack scenario for target domain %s (%s). Current proficiency: %.1f%%.

Learner Profile:
%s

Rules for Playbook Options (containment_steps, eradication_steps, recovery_steps):
1. Provide 3 options per phase: 1 option with "is_correct": true and 2 options with "is_correct": false.
2. Length & Specificity Parity: ALL options (both correct and false distractors) MUST be 15-30 words long, written in realistic SOC operational language, and explicitly reference the scenario's target_host, attacker_ip, affected accounts, or services.
3. Plausible Distractors: Never use absurd or generic distractors. Distractors must represent subtle SOC analytical missteps (e.g., premature action that destroys volatile forensic evidence, blocking an internal gateway instead of attacker_ip, applying a control from the wrong IR phase, or restoring from an unverified snapshot before patching).
4. If "is_true_positive" is false (False Positive scenario), "is_correct": true options must focus on verifying authorized change tickets, tuning SIEM correlation thresholds, and documenting benign activity without disrupting production, while "is_correct": false distractors represent unnecessary host quarantine or account lockout.

Output ONLY valid JSON matching this exact structure:
{
  "title": "Short Attack Title",
  "description": "Brief 2-sentence scenario overview for SOC analyst",
  "is_true_positive": true,
  "tp_fp_explanation": "Detailed explanation why this is a True or False Positive incident",
  "attacker_ip": "185.220.101.45",
  "target_host": "SRV-WEB01",
  "attack_log_message": "Descriptive technical SIEM/EDR log message showing the malicious exploit/traffic",
  "key_finding": "Detailed description of the primary security finding discovered in telemetry",
  "containment_steps": [
    {"label": "Detailed correct containment procedure referencing SRV-WEB01 and 185.220.101.45...", "is_correct": true, "explanation": "Why this containment action is effective"},
    {"label": "Plausible but flawed containment procedure referencing SRV-WEB01 and 185.220.101.45...", "is_correct": false, "explanation": "Why this distractor is flawed"},
    {"label": "Second plausible but flawed containment procedure referencing SRV-WEB01 and 185.220.101.45...", "is_correct": false, "explanation": "Why this distractor is flawed"}
  ],
  "eradication_steps": [
    {"label": "Detailed correct eradication procedure referencing SRV-WEB01...", "is_correct": true, "explanation": "Why this eradication action is effective"},
    {"label": "Plausible but flawed eradication procedure referencing SRV-WEB01...", "is_correct": false, "explanation": "Why this distractor is flawed"},
    {"label": "Second plausible but flawed eradication procedure referencing SRV-WEB01...", "is_correct": false, "explanation": "Why this distractor is flawed"}
  ],
  "recovery_steps": [
    {"label": "Detailed correct recovery procedure referencing SRV-WEB01...", "is_correct": true, "explanation": "Why this recovery action is effective"},
    {"label": "Plausible but flawed recovery procedure referencing SRV-WEB01...", "is_correct": false, "explanation": "Why this distractor is flawed"},
    {"label": "Second plausible but flawed recovery procedure referencing SRV-WEB01...", "is_correct": false, "explanation": "Why this distractor is flawed"}
  ]
}`, weakestDomain, domainName, proficiency, profileSummary.String())

	reqBody := chatCompletionRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.3,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal LLM request: %w", err)
	}

	client := &http.Client{Timeout: 120 * time.Second}

	var resp *http.Response
	maxRetries := 2
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Rebuild request for each attempt (body is consumed after first read)
		retryReq, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBytes))
		if reqErr != nil {
			return nil, fmt.Errorf("failed to create HTTP request (attempt %d): %w", attempt, reqErr)
		}
		retryReq.Header.Set("Content-Type", "application/json")
		retryReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cleanKey))

		resp, err = client.Do(retryReq)
		if err == nil {
			break
		}
		log.Printf("[LLM] Attempt %d/%d failed: %v", attempt, maxRetries, err)
		if attempt < maxRetries {
			time.Sleep(3 * time.Second)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("LLM request failed after %d attempts: %w", maxRetries, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("[LLM Error Details] URL: %s, Status: %d, Response: %s", url, resp.StatusCode, string(bodyBytes))
		return nil, fmt.Errorf("LLM API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode LLM response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("LLM returned empty choices")
	}

	rawJSON := strings.TrimSpace(chatResp.Choices[0].Message.Content)

	// Clean codeblock markdown wrappers if model includes them
	if strings.HasPrefix(rawJSON, "```json") {
		rawJSON = strings.TrimPrefix(rawJSON, "```json")
		rawJSON = strings.TrimSuffix(rawJSON, "```")
		rawJSON = strings.TrimSpace(rawJSON)
	} else if strings.HasPrefix(rawJSON, "```") {
		rawJSON = strings.TrimPrefix(rawJSON, "```")
		rawJSON = strings.TrimSuffix(rawJSON, "```")
		rawJSON = strings.TrimSpace(rawJSON)
	}

	var micro aiMicroScenario
	if err := json.Unmarshal([]byte(rawJSON), &micro); err != nil {
		return nil, fmt.Errorf("invalid_json: failed to parse AI micro scenario JSON: %w", err)
	}

	// Synthesize full 45-log SIEM environment and scenario object around the AI-generated incident core
	scenario := synthesizeFullScenario(micro, weakestDomain, domainName)

	// Validate candidate scenario against schema & referential contracts
	if valErr := ValidateScenarioCandidate(scenario); valErr != nil {
		return nil, fmt.Errorf("scenario_validation_failed: %w", valErr)
	}

	return scenario, nil
}

// ValidateScenarioCandidate validates scenario structure, bounds, anti-predictability rules, and referential integrity
func ValidateScenarioCandidate(s *models.Scenario) error {
	if s == nil {
		return fmt.Errorf("scenario is nil")
	}
	if strings.TrimSpace(s.Title) == "" {
		return fmt.Errorf("schema_validation_failed: title is required")
	}
	if s.DomainID == "" {
		return fmt.Errorf("schema_validation_failed: domain_id is required")
	}
	validDomains := map[string]bool{"domain1": true, "domain2": true, "domain3": true, "domain4": true, "domain5": true}
	if !validDomains[s.DomainID] {
		return fmt.Errorf("referential_validation_failed: unknown domain_id '%s'", s.DomainID)
	}

	// Validate Logs length
	logCount := 0
	if logsList, ok := s.InitialLogs.([]map[string]interface{}); ok {
		logCount = len(logsList)
	} else if rawSlice, ok := s.InitialLogs.([]interface{}); ok {
		logCount = len(rawSlice)
	}
	if logCount < 10 {
		return fmt.Errorf("content_validation_failed: insufficient logs count (%d)", logCount)
	}

	// Validate Expected Outcomes
	outcomes, ok := s.ExpectedOutcomes.(map[string]interface{})
	if !ok {
		return fmt.Errorf("schema_validation_failed: expected_outcomes must be an object")
	}
	rawFindings, exists := outcomes["key_findings"]
	if !exists {
		return fmt.Errorf("schema_validation_failed: key_findings missing from expected_outcomes")
	}
	findingsCount := 0
	if fList, ok := rawFindings.([]map[string]interface{}); ok {
		findingsCount = len(fList)
	} else if fSlice, ok := rawFindings.([]interface{}); ok {
		findingsCount = len(fSlice)
	}
	if findingsCount == 0 {
		return fmt.Errorf("content_validation_failed: at least 1 key finding is required")
	}

	// Validate Playbook Phases & Anti-Predictability Contract
	playbook, ok := s.PlaybookSteps.(map[string]interface{})
	if !ok {
		return fmt.Errorf("schema_validation_failed: playbook_steps must be an object")
	}

	seenLabels := make(map[string]bool)
	for _, phase := range []string{"containment", "eradication", "recovery"} {
		rawPhase, hasPhase := playbook[phase]
		if !hasPhase {
			return fmt.Errorf("schema_validation_failed: missing playbook phase '%s'", phase)
		}

		steps := normalizePhaseStepSlice(rawPhase)
		if len(steps) < 3 || len(steps) > 4 {
			return fmt.Errorf("content_validation_failed: phase '%s' has %d options, must have 3 to 4 options", phase, len(steps))
		}

		correctCount := 0
		distractorCount := 0
		for _, step := range steps {
			label, _ := step["label"].(string)
			trimmed := strings.TrimSpace(label)
			if trimmed == "" {
				return fmt.Errorf("content_validation_failed: empty option label in phase '%s'", phase)
			}
			if IsBannedBoilerplateLabel(trimmed) {
				return fmt.Errorf("content_validation_failed: banned boilerplate option label %q in phase '%s'", trimmed, phase)
			}

			normKey := strings.ToLower(trimmed)
			if seenLabels[normKey] {
				return fmt.Errorf("content_validation_failed: duplicate option label %q in phase '%s'", trimmed, phase)
			}
			seenLabels[normKey] = true

			isCorrect, _ := step["is_correct"].(bool)
			if isCorrect {
				correctCount++
			} else {
				distractorCount++
			}
		}

		if correctCount < 1 || correctCount > 2 {
			return fmt.Errorf("content_validation_failed: phase '%s' must have 1 to 2 correct options, got %d", phase, correctCount)
		}
		if distractorCount < 2 {
			return fmt.Errorf("content_validation_failed: phase '%s' must have at least 2 distractor options, got %d", phase, distractorCount)
		}
	}

	return nil
}

func normalizePhaseStepSlice(rawPhase interface{}) []map[string]interface{} {
	switch typed := rawPhase.(type) {
	case []map[string]interface{}:
		return typed
	case []interface{}:
		out := make([]map[string]interface{}, 0, len(typed))
		for _, item := range typed {
			if m, ok := item.(map[string]interface{}); ok {
				out = append(out, m)
			}
		}
		return out
	default:
		return nil
	}
}

// synthesizeFullScenario constructs a rich SIEM environment around the AI-generated attack core
func synthesizeFullScenario(ai aiMicroScenario, weakestDomain string, domainName string) *models.Scenario {
	if weakestDomain == "" {
		weakestDomain = "domain4"
	}
	if ai.Title == "" {
		ai.Title = fmt.Sprintf("AI-Generated Incident (%s)", domainName)
	}
	if ai.AttackerIP == "" {
		ai.AttackerIP = "185.220.101.45"
	}
	if ai.TargetHost == "" {
		ai.TargetHost = "SRV-WEB01"
	}
	if ai.AttackLogMessage == "" {
		ai.AttackLogMessage = fmt.Sprintf("Suspicious outbound network connection detected from %s to %s", ai.TargetHost, ai.AttackerIP)
	}
	if ai.KeyFinding == "" {
		ai.KeyFinding = fmt.Sprintf("Identified malicious activity originating from %s targeting %s", ai.AttackerIP, ai.TargetHost)
	}

	// Ensure description includes explicit Investigation Clues for learner
	if !strings.Contains(ai.Description, "Target Host:") && !strings.Contains(ai.Description, "INVESTIGATION CLUES") {
		ai.Description = fmt.Sprintf("%s\n\nINVESTIGATION CLUES:\n- Target Host: %s\n- Suspected Attacker IP: %s\n- Alert Vector: %s", ai.Description, ai.TargetHost, ai.AttackerIP, ai.Title)
	}

	now := time.Now().UTC()

	// Generate 45 telemetry logs (15 baseline + 5 evidence logs + 25 surrounding logs)
	logs := make([]map[string]interface{}, 0, 45)

	// Baseline noise logs (indices 0..14)
	for i := 0; i < 15; i++ {
		t := now.Add(time.Duration(i*20) * time.Second).Format(time.RFC3339)
		logs = append(logs, map[string]interface{}{
			"timestamp":      t,
			"source":         "ad_controller",
			"severity":       "info",
			"severity_level": 1,
			"message":        fmt.Sprintf("Kerberos ticket TGT requested for user account user%d@corp.local", i),
			"src_ip":         fmt.Sprintf("10.0.1.%d", 10+i),
			"user":           fmt.Sprintf("user%d", i),
			"host":           fmt.Sprintf("WS-WORKSTATION%d", i+1),
		})
	}

	// Attack evidence logs (indices 15..19)
	t15 := now.Add(time.Duration(310) * time.Second).Format(time.RFC3339)
	logs = append(logs, map[string]interface{}{
		"timestamp":      t15,
		"source":         "firewall",
		"severity":       "warning",
		"severity_level": 3,
		"message":        fmt.Sprintf("Perimeter Firewall: Connection request initiated from %s to %s:443 [ACTION: ALLOWED]", ai.TargetHost, ai.AttackerIP),
		"src_ip":         "10.0.1.55",
		"dst_ip":         ai.AttackerIP,
		"user":           "sys_admin",
		"host":           ai.TargetHost,
	})

	t16 := now.Add(time.Duration(320) * time.Second).Format(time.RFC3339)
	logs = append(logs, map[string]interface{}{
		"timestamp":      t16,
		"source":         "edr",
		"severity":       "high",
		"severity_level": 4,
		"message":        ai.AttackLogMessage,
		"src_ip":         "10.0.1.55",
		"dst_ip":         ai.AttackerIP,
		"user":           "sys_admin",
		"host":           ai.TargetHost,
	})

	t17 := now.Add(time.Duration(330) * time.Second).Format(time.RFC3339)
	logs = append(logs, map[string]interface{}{
		"timestamp":      t17,
		"source":         "ids",
		"severity":       "critical",
		"severity_level": 4,
		"message":        fmt.Sprintf("IDS Alert: Malicious attack payload signature matched on incoming traffic from %s targeting host %s", ai.AttackerIP, ai.TargetHost),
		"src_ip":         ai.AttackerIP,
		"dst_ip":         "10.0.1.55",
		"user":           "sys_admin",
		"host":           ai.TargetHost,
	})

	t18 := now.Add(time.Duration(340) * time.Second).Format(time.RFC3339)
	logs = append(logs, map[string]interface{}{
		"timestamp":      t18,
		"source":         "web_app",
		"severity":       "high",
		"severity_level": 4,
		"message":        fmt.Sprintf("HTTP 500 Internal Server Error - Anomaly detected on %s via POST payload from %s", ai.TargetHost, ai.AttackerIP),
		"src_ip":         ai.AttackerIP,
		"dst_ip":         "10.0.1.55",
		"user":           "www-data",
		"host":           ai.TargetHost,
	})

	t19 := now.Add(time.Duration(350) * time.Second).Format(time.RFC3339)
	logs = append(logs, map[string]interface{}{
		"timestamp":      t19,
		"source":         "siem_correlator",
		"severity":       "critical",
		"severity_level": 4,
		"message":        fmt.Sprintf("SIEM Alert Rule [CORR-9901]: Multi-stage incident correlation triggered on host %s involving IP %s", ai.TargetHost, ai.AttackerIP),
		"src_ip":         ai.AttackerIP,
		"dst_ip":         "10.0.1.55",
		"user":           "sys_admin",
		"host":           ai.TargetHost,
	})

	// Trailing noise logs (indices 20..44)
	for i := 20; i < 45; i++ {
		t := now.Add(time.Duration(i*25) * time.Second).Format(time.RFC3339)
		source := "proxy_gateway"
		if i%2 == 0 {
			source = "dns_server"
		} else if i%3 == 0 {
			source = "dhcp_server"
		}
		logs = append(logs, map[string]interface{}{
			"timestamp":      t,
			"source":         source,
			"severity":       "info",
			"severity_level": 1,
			"message":        fmt.Sprintf("Routine network telemetry event %d logged for internal-service-%d.corp.local", i, i),
			"src_ip":         fmt.Sprintf("10.0.2.%d", i),
			"user":           "service_acct",
			"host":           fmt.Sprintf("SRV-NODE%d", i),
		})
	}

	containmentSteps := buildPhaseStepsWithDistractors("containment", "c", weakestDomain, ai.ContainmentAction, ai.ContainmentSteps, ai)
	eradicationSteps := buildPhaseStepsWithDistractors("eradication", "e", weakestDomain, ai.EradicationAction, ai.EradicationSteps, ai)
	recoverySteps := buildPhaseStepsWithDistractors("recovery", "r", weakestDomain, ai.RecoveryAction, ai.RecoverySteps, ai)

	expectedOutcomes := map[string]interface{}{
		"key_findings": []map[string]interface{}{
			{
				"id":                   "f1",
				"description":          ai.KeyFinding,
				"domain_id":            weakestDomain,
				"points":               10,
				"evidence_log_indices": []int{15, 16, 17, 18, 19},
				"explanation":          ai.TpFpExplanation,
			},
		},
	}

	playbookSteps := map[string]interface{}{
		"containment": containmentSteps,
		"eradication": eradicationSteps,
		"recovery":    recoverySteps,
	}

	return &models.Scenario{
		Title:            ai.Title,
		Description:      ai.Description,
		DomainID:         weakestDomain,
		Difficulty:       3,
		IsTruePositive:   ai.IsTruePositive,
		TpFpExplanation:  ai.TpFpExplanation,
		InitialLogs:      logs,
		ExpectedOutcomes: expectedOutcomes,
		PlaybookSteps:    playbookSteps,
		Status:           "active",
	}
}

func buildPhaseStepsWithDistractors(phase string, idPrefix string, domainID string, legacyCorrectAction string, providedSteps []aiResponseActionOption, ai aiMicroScenario) []map[string]interface{} {
	steps := make([]map[string]interface{}, 0, 4)
	seen := make(map[string]bool)
	correctCount := 0
	distractorCount := 0

	addOption := func(label string, isCorrect bool, explanation string) {
		if len(steps) >= 4 {
			return
		}
		if isCorrect && correctCount >= 2 {
			return
		}
		trimmed := strings.TrimSpace(label)
		if trimmed == "" || IsBannedBoilerplateLabel(trimmed) {
			return
		}
		norm := strings.ToLower(trimmed)
		if seen[norm] {
			return
		}
		seen[norm] = true

		pts := -3
		if isCorrect {
			pts = 5
			correctCount++
		} else {
			distractorCount++
		}
		if strings.TrimSpace(explanation) == "" {
			if isCorrect {
				explanation = fmt.Sprintf("Directly addresses %s objectives for %s.", phase, ai.TargetHost)
			} else {
				explanation = fmt.Sprintf("Inappropriate %s action that either destroys forensic evidence or fails to mitigate risk on %s.", phase, ai.TargetHost)
			}
		}

		steps = append(steps, map[string]interface{}{
			"id":          fmt.Sprintf("%s%d", idPrefix, len(steps)+1),
			"label":       trimmed,
			"is_correct":  isCorrect,
			"domain_id":   domainID,
			"points":      pts,
			"explanation": explanation,
		})
	}

	for _, opt := range providedSteps {
		addOption(opt.Label, opt.IsCorrect, opt.Explanation)
	}

	if correctCount == 0 && strings.TrimSpace(legacyCorrectAction) != "" {
		addOption(legacyCorrectAction, true, fmt.Sprintf("Directly executes validated %s procedure on %s.", phase, ai.TargetHost))
	}

	if correctCount == 0 {
		defaultCorrect := defaultContextualCorrectOption(phase, ai)
		addOption(defaultCorrect.Label, true, defaultCorrect.Explanation)
	}

	if distractorCount < 2 || len(steps) < 3 {
		candidates := buildContextualDistractors(phase, domainID, ai)
		for _, cand := range candidates {
			if distractorCount >= 2 && len(steps) >= 3 {
				break
			}
			addOption(cand.Label, false, cand.Explanation)
		}
	}

	return steps
}

func defaultContextualCorrectOption(phase string, ai aiMicroScenario) aiResponseActionOption {
	host := ai.TargetHost
	if host == "" {
		host = "SRV-WEB01"
	}
	ip := ai.AttackerIP
	if ip == "" {
		ip = "185.220.101.45"
	}

	if !ai.IsTruePositive {
		switch phase {
		case "containment":
			return aiResponseActionOption{
				Label:       fmt.Sprintf("Verify authorized change control records for traffic between %s and %s while keeping production services online without isolating the host.", host, ip),
				IsCorrect:   true,
				Explanation: "Confirms benign activity against authorized change records without causing unnecessary production downtime.",
			}
		case "eradication":
			return aiResponseActionOption{
				Label:       fmt.Sprintf("Document the verified benign telemetry pattern from %s on %s and refine SIEM correlation exclusions to reduce false positive alert fatigue.", ip, host),
				IsCorrect:   true,
				Explanation: "Improves detection fidelity without modifying legitimate system binaries or configurations.",
			}
		default:
			return aiResponseActionOption{
				Label:       fmt.Sprintf("Close the incident ticket as a verified False Positive and confirm normal baseline monitoring remains active for %s and source %s.", host, ip),
				IsCorrect:   true,
				Explanation: "Completes the incident lifecycle cleanly while preserving standard monitoring coverage.",
			}
		}
	}

	switch phase {
	case "containment":
		return aiResponseActionOption{
			Label:       fmt.Sprintf("Isolate %s from the production network VLAN immediately and block all inbound and outbound sessions involving %s at the perimeter firewall.", host, ip),
			IsCorrect:   true,
			Explanation: "Immediately halts active command-and-control communication and prevents lateral movement from the compromised host.",
		}
	case "eradication":
		return aiResponseActionOption{
			Label:       fmt.Sprintf("Terminate unauthorized processes on %s, remove persistence artifacts linked to %s, and rotate all credentials exposed on the host.", host, ip),
			IsCorrect:   true,
			Explanation: "Eliminates attacker persistence mechanisms and invalidates potentially harvested credentials.",
		}
	default:
		return aiResponseActionOption{
			Label:       fmt.Sprintf("Restore %s from a verified clean image, validate service integrity, and enforce enhanced SIEM telemetry monitoring for traffic from %s.", host, ip),
			IsCorrect:   true,
			Explanation: "Safely returns the service to production with heightened visibility against re-infection.",
		}
	}
}

func buildContextualDistractors(phase string, domainID string, ai aiMicroScenario) []aiResponseActionOption {
	host := ai.TargetHost
	if host == "" {
		host = "SRV-WEB01"
	}
	ip := ai.AttackerIP
	if ip == "" {
		ip = "185.220.101.45"
	}

	domainFocus := map[string]string{
		"domain1": "identity and access control",
		"domain2": "endpoint exploit mitigation",
		"domain3": "network zone architecture",
		"domain4": "SOC telemetry correlation",
		"domain5": "compliance audit governance",
	}[domainID]
	if domainFocus == "" {
		domainFocus = "security operations"
	}

	if !ai.IsTruePositive {
		switch phase {
		case "containment":
			return []aiResponseActionOption{
				{
					Label:       fmt.Sprintf("Immediately quarantine %s at the EDR layer and blackhole all subnet routing toward %s before reviewing %s change logs.", host, ip, domainFocus),
					IsCorrect:   false,
					Explanation: "Prematurely isolates a production system during a benign event without verifying authorization logs.",
				},
				{
					Label:       fmt.Sprintf("Revoke all active Kerberos tickets and disable service accounts operating on %s in response to the benign traffic from %s.", host, ip),
					IsCorrect:   false,
					Explanation: "Disrupts legitimate business authentication workflows for a verified False Positive alert.",
				},
				{
					Label:       fmt.Sprintf("Configure an emergency perimeter firewall drop rule for the entire internal subnet hosting %s to block connections from %s.", host, ip),
					IsCorrect:   false,
					Explanation: "Causes a broad network outage for benign authorized activity.",
				},
			}
		case "eradication":
			return []aiResponseActionOption{
				{
					Label:       fmt.Sprintf("Wipe the operating system partition on %s and revoke TLS certificates associated with %s despite confirmed benign %s verification.", host, ip, domainFocus),
					IsCorrect:   false,
					Explanation: "Destructive eradication actions are inappropriate when telemetry confirms a False Positive.",
				},
				{
					Label:       fmt.Sprintf("Permanently disable the SIEM correlation rule and EDR telemetry agent on %s so future events from %s are ignored.", host, ip),
					IsCorrect:   false,
					Explanation: "Disabling security sensors creates a critical visibility blind spot rather than properly tuning the rule.",
				},
				{
					Label:       fmt.Sprintf("Force a global password reset across all domain users who accessed %s during the window of activity from %s.", host, ip),
					IsCorrect:   false,
					Explanation: "Unnecessary enterprise-wide disruption for a benign operational event.",
				},
			}
		default:
			return []aiResponseActionOption{
				{
					Label:       fmt.Sprintf("Roll back %s to a week-old backup snapshot and purge all database transactions recorded during the connection from %s.", host, ip),
					IsCorrect:   false,
					Explanation: "Causes severe production data loss when no malicious tampering occurred.",
				},
				{
					Label:       fmt.Sprintf("Migrate %s to an unmonitored DMZ VLAN and bypass %s inspection for traffic originating from %s.", host, domainFocus, ip),
					IsCorrect:   false,
					Explanation: "Weakens network architecture security controls and removes inspection visibility.",
				},
				{
					Label:       fmt.Sprintf("Grant permanent local administrator exceptions on %s for all future connections matching %s without MFA verification.", host, ip),
					IsCorrect:   false,
					Explanation: "Violates least-privilege principles and introduces unnecessary access risk.",
				},
			}
		}
	}

	switch phase {
	case "containment":
		return []aiResponseActionOption{
			{
				Label:       fmt.Sprintf("Perform an immediate hard power-off of %s and flush volatile RAM caches to terminate active sockets connected to %s.", host, ip),
				IsCorrect:   false,
				Explanation: "Destroys volatile memory artifacts and active network state required for forensic root-cause investigation.",
			},
			{
				Label:       fmt.Sprintf("Reroute suspicious traffic from %s targeting %s into the internal corporate backup VLAN while bypassing %s controls.", ip, host, domainFocus),
				IsCorrect:   false,
				Explanation: "Exposes the internal backup network segment to lateral movement instead of containing the threat.",
			},
			{
				Label:       fmt.Sprintf("Block the internal default gateway and core DNS resolver for the entire subnet of %s rather than filtering attacker endpoint %s.", host, ip),
				IsCorrect:   false,
				Explanation: "Causes collateral outage across healthy production hosts without isolating the compromised endpoint.",
			},
		}
	case "eradication":
		return []aiResponseActionOption{
			{
				Label:       fmt.Sprintf("Clear all local security event logs and %s history on %s before restarting the service connected to %s.", domainFocus, host, ip),
				IsCorrect:   false,
				Explanation: "Erases critical audit trails without removing the underlying vulnerability or persistence mechanism.",
			},
			{
				Label:       fmt.Sprintf("Restart the affected application daemon on %s and whitelist %s in the host firewall to suppress recurring IDS alerts.", host, ip),
				IsCorrect:   false,
				Explanation: "Leaves persistent backdoors intact on disk and explicitly permits further attacker access.",
			},
			{
				Label:       fmt.Sprintf("Delete only temporary browser cache files on %s while leaving scheduled tasks and service accounts touched by %s unmodified.", host, ip),
				IsCorrect:   false,
				Explanation: "Fails to remediate privilege escalation and persistence mechanisms established during the intrusion.",
			},
		}
	default:
		return []aiResponseActionOption{
			{
				Label:       fmt.Sprintf("Restore %s directly from the most recent unverified snapshot taken after %s initiated exploitation without applying %s patches.", host, ip, domainFocus),
				IsCorrect:   false,
				Explanation: "Reintroduces the unpatched vulnerability and potentially restores dormant attacker persistence.",
			},
			{
				Label:       fmt.Sprintf("Reconnect %s to the production VLAN immediately and suspend EDR behavioral monitoring for 48 hours to avoid performance overhead after %s.", host, ip),
				IsCorrect:   false,
				Explanation: "Removes endpoint visibility during the critical post-incident observation window.",
			},
			{
				Label:       fmt.Sprintf("Temporarily bypass multi-factor authentication and strict firewall ACLs on %s so users can rapidly resume work following the %s incident.", host, ip),
				IsCorrect:   false,
				Explanation: "Degrades authentication and network security posture precisely when the environment is at elevated risk of re-compromise.",
			},
		}
	}
}

// GenerateFallbackScenario constructs a synthesized AI scenario when external LLM API is unavailable and static pool is exhausted.
func GenerateFallbackScenario(weakestDomain string, domainName string) *models.Scenario {
	templates := map[string]aiMicroScenario{
		"domain1": {
			Title:            "Unauthorized Security Policy Override Attempt",
			Description:      "A domain controller logged unexpected security policy modifications bypassing central IAM authentication controls.",
			IsTruePositive:   true,
			TpFpExplanation:  "This is a True Positive because unapproved privilege elevation calls were executed directly on the domain controller.",
			AttackerIP:       "192.168.1.150",
			TargetHost:       "DC01.corp.internal",
			AttackLogMessage: "CRITICAL AUDIT: Security policy 'DisableAccountLockout' applied by unauthorized user 'SYSTEM_SVC'",
			KeyFinding:       "Unauthorized account lockout policy alteration detected on Primary Domain Controller DC01.",
			ContainmentSteps: []aiResponseActionOption{
				{
					Label:       "Revert the unauthorized Group Policy modification on DC01.corp.internal immediately, revoke active Kerberos tickets for SYSTEM_SVC, and restrict administrative RPC access from 192.168.1.150.",
					IsCorrect:   true,
					Explanation: "Directly halts unauthorized policy enforcement and revokes the compromised service session on DC01.corp.internal.",
				},
				{
					Label:       "Demote DC01.corp.internal from the Active Directory forest immediately and flush all volatile security event logs generated by 192.168.1.150.",
					IsCorrect:   false,
					Explanation: "Demoting a primary domain controller ungracefully causes enterprise outage and destroys volatile audit evidence.",
				},
				{
					Label:       "Migrate SYSTEM_SVC to the Domain Admins group on DC01.corp.internal while routing traffic from 192.168.1.150 through an unmonitored management VLAN.",
					IsCorrect:   false,
					Explanation: "Escalates attacker privileges and removes network telemetry visibility.",
				},
			},
			EradicationSteps: []aiResponseActionOption{
				{
					Label:       "Rotate the SYSTEM_SVC service account credentials, remove unauthorized GPO links on DC01.corp.internal, and audit host 192.168.1.150 for lateral movement tools.",
					IsCorrect:   true,
					Explanation: "Eliminates the compromised credential vector and removes unauthorized policy persistence.",
				},
				{
					Label:       "Restart the Netlogon service on DC01.corp.internal without rotating SYSTEM_SVC credentials or inspecting 192.168.1.150 for credential dumping utilities.",
					IsCorrect:   false,
					Explanation: "Restarting the service leaves the compromised credentials active for immediate re-exploitation.",
				},
				{
					Label:       "Delete the entire SYSVOL replication directory on DC01.corp.internal to clear policy files touched by 192.168.1.150 without backing up baseline GPOs.",
					IsCorrect:   false,
					Explanation: "Deleting SYSVOL breaks domain-wide policy replication across all enterprise endpoints.",
				},
			},
			RecoverySteps: []aiResponseActionOption{
				{
					Label:       "Restore the verified baseline Group Policy configuration on DC01.corp.internal, enforce tiered IAM access controls, and monitor authentication logs from 192.168.1.150.",
					IsCorrect:   true,
					Explanation: "Restores enforced account lockout baselines and ensures continuous monitoring against policy tampering.",
				},
				{
					Label:       "Re-enable the SYSTEM_SVC account on DC01.corp.internal with its previous password hash so legacy scripts on 192.168.1.150 do not fail.",
					IsCorrect:   false,
					Explanation: "Reusing compromised credentials allows the attacker to immediately regain administrative control.",
				},
				{
					Label:       "Disable Kerberos pre-authentication and audit logging on DC01.corp.internal to accelerate domain replication after the 192.168.1.150 incident.",
					IsCorrect:   false,
					Explanation: "Weakens domain authentication security and blinds the SOC to subsequent credential attacks.",
				},
			},
		},
		"domain2": {
			Title:            "Suspicious Command Line Execution via Spearphishing Payload",
			Description:      "An endpoint IDS generated a high-severity alert following a user opening an obfuscated attachment containing PowerShell execution scripts.",
			IsTruePositive:   true,
			TpFpExplanation:  "This is a True Positive because encoded PowerShell commands were executed immediately following document launch.",
			AttackerIP:       "185.220.101.99",
			TargetHost:       "WS-FINANCE04",
			AttackLogMessage: "EDR ALERT: powershell.exe -e aW52b2tlLWV4cHJlc3Npb24... executed by WINWORD.EXE",
			KeyFinding:       "Encoded PowerShell payload executed via MS Office child process on WS-FINANCE04.",
			ContainmentSteps: []aiResponseActionOption{
				{
					Label:       "Isolate workstation WS-FINANCE04 from the corporate network via EDR containment and block outbound C2 connections to 185.220.101.99 at the perimeter firewall.",
					IsCorrect:   true,
					Explanation: "Contains the endpoint while preserving memory state and blocks external command-and-control communication.",
				},
				{
					Label:       "Force an immediate hard power-off of WS-FINANCE04 and clear local firewall connection tables referencing 185.220.101.99 before capturing memory.",
					IsCorrect:   false,
					Explanation: "Destroys volatile RAM artifacts needed to decode the in-memory PowerShell payload.",
				},
				{
					Label:       "Move WS-FINANCE04 into the finance server database VLAN while allowing outbound traffic to 185.220.101.99 to observe attacker behavior.",
					IsCorrect:   false,
					Explanation: "Exposes critical finance servers to lateral movement and data exfiltration.",
				},
			},
			EradicationSteps: []aiResponseActionOption{
				{
					Label:       "Terminate the malicious WINWORD.EXE and powershell.exe process tree on WS-FINANCE04, purge staged payloads in AppData, and remove registry run keys.",
					IsCorrect:   true,
					Explanation: "Completely removes the active malware processes, staged binaries, and user-level persistence.",
				},
				{
					Label:       "Uninstall Microsoft Office on WS-FINANCE04 while leaving the spawned powershell.exe background process connected to 185.220.101.99 running.",
					IsCorrect:   false,
					Explanation: "Uninstalling the parent application does not terminate an already-spawned living-off-the-land process.",
				},
				{
					Label:       "Delete Windows Event Viewer PowerShell operational logs on WS-FINANCE04 to clear alerts triggered by 185.220.101.99 without scanning for persistence.",
					IsCorrect:   false,
					Explanation: "Destroys forensic evidence while leaving the malware active on the endpoint.",
				},
			},
			RecoverySteps: []aiResponseActionOption{
				{
					Label:       "Re-image WS-FINANCE04 from a verified clean baseline, enforce Attack Surface Reduction macro rules, and monitor traffic for callbacks to 185.220.101.99.",
					IsCorrect:   true,
					Explanation: "Guarantees a clean workstation state and hardens Office against future macro-based payloads.",
				},
				{
					Label:       "Restore WS-FINANCE04 from an unverified system restore point created after the 185.220.101.99 callback and disable EDR script-block logging.",
					IsCorrect:   false,
					Explanation: "Restores potentially infected artifacts and disables critical PowerShell visibility.",
				},
				{
					Label:       "Reconnect WS-FINANCE04 to the production network immediately and grant local administrator rights to the user to reinstall tools after the 185.220.101.99 alert.",
					IsCorrect:   false,
					Explanation: "Increases privilege escalation risk on an endpoint that has not been verified clean.",
				},
			},
		},
		"domain3": {
			Title:            "Cloud Perimeter Firewall Misconfiguration",
			Description:      "Monitoring tools flagged an inbound connection to an internal management interface exposed to the public internet.",
			IsTruePositive:   true,
			TpFpExplanation:  "This is a True Positive because SSH port 22 was open to 0.0.0.0/0 on a production web database host.",
			AttackerIP:       "203.0.113.88",
			TargetHost:       "AWS-PROD-DB01",
			AttackLogMessage: "FIREWALL: Inbound SSH connection accepted from 203.0.113.88:49210 to 10.0.3.12:22",
			KeyFinding:       "Management SSH interface exposed publicly due to misconfigured AWS Security Group rule.",
			ContainmentSteps: []aiResponseActionOption{
				{
					Label:       "Restrict the AWS Security Group ingress rule on AWS-PROD-DB01 to allow port 22 only from the internal bastion host and terminate active sessions from 203.0.113.88.",
					IsCorrect:   true,
					Explanation: "Immediately closes public exposure of the management port and severs the unauthorized external SSH session.",
				},
				{
					Label:       "Delete the entire production VPC routing table for AWS-PROD-DB01 while keeping the SSH session from 203.0.113.88 open on the host.",
					IsCorrect:   false,
					Explanation: "Causes a complete production outage without terminating the local attacker process tree.",
				},
				{
					Label:       "Change the SSH daemon port on AWS-PROD-DB01 from 22 to 2222 while leaving 0.0.0.0/0 ingress open to 203.0.113.88.",
					IsCorrect:   false,
					Explanation: "Security through obscurity does not remediate public exposure to the attacker IP.",
				},
			},
			EradicationSteps: []aiResponseActionOption{
				{
					Label:       "Audit authorized_keys and PAM configurations on AWS-PROD-DB01, revoke any SSH keypairs accessed during the 203.0.113.88 session, and remove rogue cron jobs.",
					IsCorrect:   true,
					Explanation: "Eliminates unauthorized SSH keys and scheduled persistence left on the database host.",
				},
				{
					Label:       "Restart the sshd daemon on AWS-PROD-DB01 without inspecting ~/.ssh/authorized_keys or checking for persistence planted by 203.0.113.88.",
					IsCorrect:   false,
					Explanation: "Leaves unauthorized SSH keys and backdoors intact on the host.",
				},
				{
					Label:       "Truncate /var/log/auth.log on AWS-PROD-DB01 to clear failed login warnings from 203.0.113.88 without rotating credentials.",
					IsCorrect:   false,
					Explanation: "Destroys authentication audit logs while leaving compromised credentials valid.",
				},
			},
			RecoverySteps: []aiResponseActionOption{
				{
					Label:       "Enforce automated Infrastructure-as-Code policy checks to block 0.0.0.0/0 SSH rules on AWS-PROD-DB01 and monitor VPC Flow Logs for 203.0.113.88.",
					IsCorrect:   true,
					Explanation: "Prevents configuration drift recurrence through automated guardrails and continuous flow monitoring.",
				},
				{
					Label:       "Re-open port 22 to 0.0.0.0/0 on AWS-PROD-DB01 temporarily so external contractors can test connectivity after the 203.0.113.88 incident.",
					IsCorrect:   false,
					Explanation: "Re-exposes the critical database management interface to the public internet.",
				},
				{
					Label:       "Disable AWS CloudTrail and VPC Flow Logs on AWS-PROD-DB01 to reduce cloud logging costs following the 203.0.113.88 investigation.",
					IsCorrect:   false,
					Explanation: "Eliminates essential cloud audit telemetry needed to detect repeat intrusion attempts.",
				},
			},
		},
		"domain4": {
			Title:            "Suspicious Outbound C2 Beaconing Activity",
			Description:      "SIEM correlation engine detected periodic outbound HTTP POST requests matching known Command & Control beacon intervals.",
			IsTruePositive:   true,
			TpFpExplanation:  "This is a True Positive because the destination IP is flagged on global threat intelligence feeds.",
			AttackerIP:       "198.51.100.42",
			TargetHost:       "WS-DEV09",
			AttackLogMessage: "SIEM CORRELATION: Outbound HTTP POST to 198.51.100.42:8080 repeating every 60s",
			KeyFinding:       "Periodic outbound network traffic matching Cobalt Strike C2 beacon signature.",
			ContainmentSteps: []aiResponseActionOption{
				{
					Label:       "Block outbound traffic to C2 IP 198.51.100.42 at the perimeter firewall and web proxy, and isolate endpoint WS-DEV09 via the EDR console.",
					IsCorrect:   true,
					Explanation: "Severs the command-and-control communication channel and prevents lateral movement from WS-DEV09.",
				},
				{
					Label:       "Unplug the core switch uplink for the entire engineering building housing WS-DEV09 and flush firewall logs referencing 198.51.100.42.",
					IsCorrect:   false,
					Explanation: "Disrupts all engineering operations unnecessarily and destroys network evidence.",
				},
				{
					Label:       "Configure a NAT port-forwarding rule on WS-DEV09 to redirect inbound packets from 198.51.100.42 to the internal domain controller.",
					IsCorrect:   false,
					Explanation: "Directly exposes core domain infrastructure to the external command-and-control server.",
				},
			},
			EradicationSteps: []aiResponseActionOption{
				{
					Label:       "Identify and terminate the Cobalt Strike beacon process on WS-DEV09, delete the malicious Windows service binary, and rotate developer credentials.",
					IsCorrect:   true,
					Explanation: "Removes the active C2 implant, eliminates service persistence, and invalidates compromised credentials.",
				},
				{
					Label:       "Add an exclusion in the EDR agent on WS-DEV09 for the directory containing the 198.51.100.42 beacon executable to stop alert generation.",
					IsCorrect:   false,
					Explanation: "Silences alerts while allowing the C2 implant to continue operating unhindered.",
				},
				{
					Label:       "Clear the local DNS resolver cache on WS-DEV09 without terminating the persistent service beaconing to 198.51.100.42.",
					IsCorrect:   false,
					Explanation: "Flushing DNS cache does not stop an active process communicating directly with an IP address.",
				},
			},
			RecoverySteps: []aiResponseActionOption{
				{
					Label:       "Rebuild WS-DEV09 from a verified gold image, validate that zero outbound beacons to 198.51.100.42 occur, and monitor proxy telemetry for 72 hours.",
					IsCorrect:   true,
					Explanation: "Ensures clean host restoration and verifies complete C2 eradication through heightened monitoring.",
				},
				{
					Label:       "Return WS-DEV09 to the production VLAN without re-imaging and whitelist 198.51.100.42 on the outbound proxy to prevent false alarms.",
					IsCorrect:   false,
					Explanation: "Leaves the host compromised and explicitly authorizes C2 traffic.",
				},
				{
					Label:       "Disable outbound proxy SSL/TLS inspection for WS-DEV09 so developer builds run faster after the 198.51.100.42 incident.",
					IsCorrect:   false,
					Explanation: "Blinds network security monitoring to encrypted C2 channels.",
				},
			},
		},
		"domain5": {
			Title:            "Compliance Breach: Unencrypted Data Transfer",
			Description:      "DLP alert reported sensitive PCI-DSS credit card records transmitted over unencrypted HTTP protocol.",
			IsTruePositive:   true,
			TpFpExplanation:  "This is a True Positive because cleartext PAN credit card numbers were intercepted in network traffic.",
			AttackerIP:       "10.0.2.77",
			TargetHost:       "SRV-PAYMENT01",
			AttackLogMessage: "DLP ALERT: Cleartext credit card pattern matched in HTTP POST body payload to external endpoint",
			KeyFinding:       "PCI-DSS compliance breach caused by cleartext HTTP transmission of customer payment data.",
			ContainmentSteps: []aiResponseActionOption{
				{
					Label:       "Disable the unencrypted HTTP listener on SRV-PAYMENT01 immediately, block unauthorized egress to 10.0.2.77, and enforce TLS 1.3 transport rules.",
					IsCorrect:   true,
					Explanation: "Immediately stops cleartext cardholder data exposure while preserving secure encrypted channels.",
				},
				{
					Label:       "Publish the captured PCAP files from SRV-PAYMENT01 containing cleartext PAN data to a public ticketing bucket for review of 10.0.2.77.",
					IsCorrect:   false,
					Explanation: "Causes a secondary PCI-DSS data breach by exposing sensitive cardholder data in an unencrypted repository.",
				},
				{
					Label:       "Disable the Data Loss Prevention sensor monitoring SRV-PAYMENT01 and 10.0.2.77 so compliance violation alerts stop firing.",
					IsCorrect:   false,
					Explanation: "Suppresses compliance monitoring while allowing cleartext cardholder data leakage to continue.",
				},
			},
			EradicationSteps: []aiResponseActionOption{
				{
					Label:       "Purge cleartext PAN data from unencrypted proxy logs between SRV-PAYMENT01 and 10.0.2.77, rotate payment API keys, and remove legacy HTTP bindings.",
					IsCorrect:   true,
					Explanation: "Sanitizes exposed cardholder records from logs and permanently removes the insecure transport configuration.",
				},
				{
					Label:       "Retain unencrypted HTTP bindings on SRV-PAYMENT01 for backward compatibility with 10.0.2.77 while compressing the payload with gzip.",
					IsCorrect:   false,
					Explanation: "Compression is not encryption and violates PCI-DSS cryptographic transmission requirements.",
				},
				{
					Label:       "Delete all historical compliance audit records on SRV-PAYMENT01 to hide the unencrypted transfer to 10.0.2.77 from external assessors.",
					IsCorrect:   false,
					Explanation: "Violates regulatory audit retention mandates and governance ethics.",
				},
			},
			RecoverySteps: []aiResponseActionOption{
				{
					Label:       "Conduct a formal PCI-DSS compliance validation on SRV-PAYMENT01, verify HSTS and TLS 1.3 enforcement, and document the breach notification workflow for 10.0.2.77.",
					IsCorrect:   true,
					Explanation: "Fulfills regulatory governance obligations and verifies technical cryptographic controls before resuming full operations.",
				},
				{
					Label:       "Resume payment processing on SRV-PAYMENT01 using self-signed SSLv3 certificates with 10.0.2.77 without notifying the compliance officer.",
					IsCorrect:   false,
					Explanation: "SSLv3 is cryptographically broken and non-compliant with PCI-DSS standards.",
				},
				{
					Label:       "Exempt SRV-PAYMENT01 and 10.0.2.77 from quarterly vulnerability scans and DLP inspection to streamline payment throughput.",
					IsCorrect:   false,
					Explanation: "Removes mandatory PCI-DSS oversight and continuous monitoring controls.",
				},
			},
		},
	}

	tmpl, exists := templates[weakestDomain]
	if !exists {
		tmpl = templates["domain4"]
	}
	tmpl.Title = fmt.Sprintf("%s (Adaptive AI Scenario)", tmpl.Title)

	return synthesizeFullScenario(tmpl, weakestDomain, domainName)
}

// extractHostAndIPFromScenario extracts TargetHost and AttackerIP entities from a Scenario's description or logs.
func extractHostAndIPFromScenario(s *models.Scenario) (string, string) {
	host := ""
	ip := ""

	for _, line := range strings.Split(s.Description, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- Target Host:") {
			host = strings.TrimSpace(strings.TrimPrefix(trimmed, "- Target Host:"))
		} else if strings.HasPrefix(trimmed, "- Suspected Attacker IP:") {
			ip = strings.TrimSpace(strings.TrimPrefix(trimmed, "- Suspected Attacker IP:"))
		}
	}

	if host == "" || ip == "" {
		logs := normalizePhaseStepSlice(s.InitialLogs)
		if len(logs) > 15 {
			if h, ok := logs[15]["host"].(string); ok && host == "" && h != "" {
				host = h
			}
			if dst, ok := logs[15]["dst_ip"].(string); ok && ip == "" && dst != "" {
				ip = dst
			} else if src, ok := logs[15]["src_ip"].(string); ok && ip == "" && src != "" {
				ip = src
			}
		}
	}

	if host == "" {
		host = "SRV-WEB01"
	}
	if ip == "" {
		ip = "185.220.101.45"
	}
	return host, ip
}

// UpgradeScenarioPlaybookSteps inspects a Scenario and upgrades its PlaybookSteps in-place if it fails anti-predictability validation.
// Returns true if the Scenario's PlaybookSteps were modified.
func UpgradeScenarioPlaybookSteps(s *models.Scenario) bool {
	if s == nil {
		return false
	}
	if ValidateScenarioCandidate(s) == nil {
		return false
	}

	domainID := s.DomainID
	if domainID == "" {
		domainID = "domain4"
	}

	host, ip := extractHostAndIPFromScenario(s)
	ai := aiMicroScenario{
		Title:           s.Title,
		Description:     s.Description,
		IsTruePositive:  s.IsTruePositive,
		TpFpExplanation: s.TpFpExplanation,
		TargetHost:      host,
		AttackerIP:      ip,
	}

	existingPB, _ := s.PlaybookSteps.(map[string]interface{})
	newPB := make(map[string]interface{}, 3)

	phasePrefixes := map[string]string{
		"containment": "c",
		"eradication": "e",
		"recovery":    "r",
	}

	for _, phase := range []string{"containment", "eradication", "recovery"} {
		var preserved []aiResponseActionOption
		if existingPB != nil {
			for _, rawStep := range normalizePhaseStepSlice(existingPB[phase]) {
				label, _ := rawStep["label"].(string)
				trimmed := strings.TrimSpace(label)
				if trimmed == "" || IsBannedBoilerplateLabel(trimmed) {
					continue
				}
				isCorrect, _ := rawStep["is_correct"].(bool)
				explanation, _ := rawStep["explanation"].(string)
				preserved = append(preserved, aiResponseActionOption{
					Label:       trimmed,
					IsCorrect:   isCorrect,
					Explanation: explanation,
				})
			}
		}

		newPB[phase] = buildPhaseStepsWithDistractors(phase, phasePrefixes[phase], domainID, "", preserved, ai)
	}

	s.PlaybookSteps = newPB
	return true
}

// EnsureScenarioPlaybookQuality scans all stored Scenarios in the database and remediates any row that fails anti-predictability validation.
func EnsureScenarioPlaybookQuality(db *gorm.DB, seedPath string) error {
	if db == nil {
		return nil
	}

	seedByTitle := make(map[string]models.Scenario)
	if seedPath == "" {
		seedPath = "scenarios_seed.json"
	}
	if rawBytes, err := os.ReadFile(seedPath); err == nil {
		var seedScenarios []models.Scenario
		if json.Unmarshal(rawBytes, &seedScenarios) == nil {
			for _, sc := range seedScenarios {
				seedByTitle[strings.ToLower(strings.TrimSpace(sc.Title))] = sc
			}
		}
	}

	var stored []models.Scenario
	if err := db.Find(&stored).Error; err != nil {
		return err
	}

	upgradedCount := 0
	for i := range stored {
		sc := &stored[i]
		if ValidateScenarioCandidate(sc) == nil {
			continue
		}

		// If title matches a static seed scenario, sync its canonical playbook_steps
		if seedMatch, ok := seedByTitle[strings.ToLower(strings.TrimSpace(sc.Title))]; ok {
			sc.PlaybookSteps = seedMatch.PlaybookSteps
		}

		// If still invalid (e.g. legacy AI scenario), upgrade dynamically
		if ValidateScenarioCandidate(sc) != nil {
			UpgradeScenarioPlaybookSteps(sc)
		}

		if err := db.Model(&models.Scenario{}).Where("scenario_id = ?", sc.ScenarioID).Update("playbook_steps", sc.PlaybookSteps).Error; err == nil {
			upgradedCount++
		}
	}

	if upgradedCount > 0 {
		log.Printf("[Scenario Quality Gate] Auto-remediated %d legacy scenario(s) in database.", upgradedCount)
	}
	return nil
}

