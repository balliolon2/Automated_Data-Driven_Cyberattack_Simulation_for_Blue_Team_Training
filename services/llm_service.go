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

type aiMicroScenario struct {
	Title             string `json:"title"`
	Description       string `json:"description"`
	IsTruePositive    bool   `json:"is_true_positive"`
	TpFpExplanation   string `json:"tp_fp_explanation"`
	AttackerIP        string `json:"attacker_ip"`
	TargetHost        string `json:"target_host"`
	AttackLogMessage  string `json:"attack_log_message"`
	KeyFinding        string `json:"key_finding"`
	ContainmentAction string `json:"containment_action"`
	EradicationAction string `json:"eradication_action"`
	RecoveryAction    string `json:"recovery_action"`
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

	systemPrompt := "You are a specialized CompTIA Security+ 701 cybersecurity scenario generator. Output ONLY raw JSON."
	userPrompt := fmt.Sprintf(`Generate a unique SOC attack scenario for target domain %s (%s). Current proficiency: %.1f%%.

Learner Profile:
%s

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
  "containment_action": "Correct incident containment action step",
  "eradication_action": "Correct incident eradication action step",
  "recovery_action": "Correct incident recovery action step"
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

// ValidateScenarioCandidate validates scenario structure, bounds, and referential integrity
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

	// Validate Playbook Phases
	playbook, ok := s.PlaybookSteps.(map[string]interface{})
	if !ok {
		return fmt.Errorf("schema_validation_failed: playbook_steps must be an object")
	}
	for _, phase := range []string{"containment", "eradication", "recovery"} {
		if _, hasPhase := playbook[phase]; !hasPhase {
			return fmt.Errorf("schema_validation_failed: missing playbook phase '%s'", phase)
		}
	}

	return nil
}

// synthesizeFullScenario constructs a rich SIEM environment around the AI-generated attack core
func synthesizeFullScenario(ai aiMicroScenario, weakestDomain string, domainName string) *models.Scenario {
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

	// Playbook steps with AI correct actions + distractor options
	containmentSteps := []map[string]interface{}{
		{
			"id":          "c1",
			"label":       ai.ContainmentAction,
			"is_correct":  true,
			"domain_id":   weakestDomain,
			"points":      5,
			"explanation": "Directly mitigates active containment risk.",
		},
		{
			"id":          "c2",
			"label":       "Power off all domain controllers immediately",
			"is_correct":  false,
			"domain_id":   weakestDomain,
			"points":      -3,
			"explanation": "Overly aggressive disruption to business infrastructure.",
		},
	}

	eradicationSteps := []map[string]interface{}{
		{
			"id":          "e1",
			"label":       ai.EradicationAction,
			"is_correct":  true,
			"domain_id":   weakestDomain,
			"points":      5,
			"explanation": "Permanently removes threat persistence mechanisms.",
		},
		{
			"id":          "e2",
			"label":       "Reformat all network switches in the subnet",
			"is_correct":  false,
			"domain_id":   weakestDomain,
			"points":      -3,
			"explanation": "Unnecessary action unrelated to endpoint threat eradication.",
		},
	}

	recoverySteps := []map[string]interface{}{
		{
			"id":          "r1",
			"label":       ai.RecoveryAction,
			"is_correct":  true,
			"domain_id":   weakestDomain,
			"points":      5,
			"explanation": "Ensures safe restoration of business operations.",
		},
		{
			"id":          "r2",
			"label":       "Disable MFA enforcement for affected users",
			"is_correct":  false,
			"domain_id":   weakestDomain,
			"points":      -3,
			"explanation": "Degrades baseline security posture.",
		},
	}

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

// GenerateFallbackScenario constructs a synthesized AI scenario when external LLM API is unavailable and static pool is exhausted.
func GenerateFallbackScenario(weakestDomain string, domainName string) *models.Scenario {
	templates := map[string]aiMicroScenario{
		"domain1": {
			Title:             "Unauthorized Security Policy Override Attempt",
			Description:       "A domain controller logged unexpected security policy modifications bypassing central IAM authentication controls.",
			IsTruePositive:    true,
			TpFpExplanation:   "This is a True Positive because unapproved privilege elevation calls were executed directly on the domain controller.",
			AttackerIP:        "192.168.1.150",
			TargetHost:        "DC01.corp.internal",
			AttackLogMessage:  "CRITICAL AUDIT: Security policy 'DisableAccountLockout' applied by unauthorized user 'SYSTEM_SVC'",
			KeyFinding:        "Unauthorized account lockout policy alteration detected on Primary Domain Controller DC01.",
			ContainmentAction: "Revert domain group policy modifications and revoke SYSTEM_SVC token",
			EradicationAction: "Reset system service account credentials and enforce central Kerberos logging",
			RecoveryAction:    "Perform full Active Directory audit and restore security policy baseline",
		},
		"domain2": {
			Title:             "Suspicious Command Line Execution via Spearphishing Payload",
			Description:       "An endpoint IDS generated a high-severity alert following a user opening an obfuscated attachment containing PowerShell execution scripts.",
			IsTruePositive:    true,
			TpFpExplanation:   "This is a True Positive because encoded PowerShell commands were executed immediately following document launch.",
			AttackerIP:        "185.220.101.99",
			TargetHost:        "WS-FINANCE04",
			AttackLogMessage:  "EDR ALERT: powershell.exe -e aW52b2tlLWV4cHJlc3Npb24... executed by WINWORD.EXE",
			KeyFinding:        "Encoded PowerShell payload executed via MS Office child process on WS-FINANCE04.",
			ContainmentAction: "Isolate WS-FINANCE04 host from network segment",
			EradicationAction: "Terminate powershell process tree and purge cached payload files from AppData",
			RecoveryAction:    "Restore endpoint from backup image and re-enable EDR real-time protection",
		},
		"domain3": {
			Title:             "Cloud Perimeter Firewall Misconfiguration",
			Description:       "Monitoring tools flagged an inbound connection to an internal management interface exposed to the public internet.",
			IsTruePositive:    true,
			TpFpExplanation:   "This is a True Positive because SSH port 22 was open to 0.0.0.0/0 on a production web database host.",
			AttackerIP:        "203.0.113.88",
			TargetHost:        "AWS-PROD-DB01",
			AttackLogMessage:  "FIREWALL: Inbound SSH connection accepted from 203.0.113.88:49210 to 10.0.3.12:22",
			KeyFinding:        "Management SSH interface exposed publicly due to misconfigured AWS Security Group rule.",
			ContainmentAction: "Restrict AWS Security Group ingress rule for port 22 to internal bastion IP",
			EradicationAction: "Audit all active SSH sessions on AWS-PROD-DB01 and revoke compromised keypairs",
			RecoveryAction:    "Enforce IaC security compliance scan before applying cloud network changes",
		},
		"domain4": {
			Title:             "Suspicious Outbound C2 Beaconing Activity",
			Description:       "SIEM correlation engine detected periodic outbound HTTP POST requests matching known Command & Control beacon intervals.",
			IsTruePositive:    true,
			TpFpExplanation:   "This is a True Positive because the destination IP is flagged on global threat intelligence feeds.",
			AttackerIP:        "198.51.100.42",
			TargetHost:        "WS-DEV09",
			AttackLogMessage:  "SIEM CORRELATION: Outbound HTTP POST to 198.51.100.42:8080 repeating every 60s",
			KeyFinding:        "Periodic outbound network traffic matching Cobalt Strike C2 beacon signature.",
			ContainmentAction: "Block C2 destination IP 198.51.100.42 at perimeter firewall and proxy",
			EradicationAction: "Identify and remove malicious persistent service on WS-DEV09",
			RecoveryAction:    "Verify complete network isolation of threat actor C2 channels and restore host",
		},
		"domain5": {
			Title:             "Compliance Breach: Unencrypted Data Transfer",
			Description:       "DLP alert reported sensitive PCI-DSS credit card records transmitted over unencrypted HTTP protocol.",
			IsTruePositive:    true,
			TpFpExplanation:   "This is a True Positive because cleartext PAN credit card numbers were intercepted in network traffic.",
			AttackerIP:        "10.0.2.77",
			TargetHost:        "SRV-PAYMENT01",
			AttackLogMessage:  "DLP ALERT: Cleartext credit card pattern matched in HTTP POST body payload to external endpoint",
			KeyFinding:        "PCI-DSS compliance breach caused by cleartext HTTP transmission of customer payment data.",
			ContainmentAction: "Disable cleartext HTTP endpoint and enforce TLS 1.3 encryption channel",
			EradicationAction: "Purge cleartext transmission logs and re-issue compromised API tokens",
			RecoveryAction:    "Perform full PCI-DSS compliance audit and re-certify secure data transfer pipelines",
		},
	}

	tmpl, exists := templates[weakestDomain]
	if !exists {
		tmpl = templates["domain4"]
	}
	tmpl.Title = fmt.Sprintf("%s (Adaptive AI Scenario)", tmpl.Title)

	return synthesizeFullScenario(tmpl, weakestDomain, domainName)
}
