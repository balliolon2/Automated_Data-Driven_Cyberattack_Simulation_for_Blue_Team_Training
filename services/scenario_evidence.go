package services

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"cybersim/models"
	"gorm.io/gorm"
)

const (
	// MaxAIFindings is the largest number of Key Findings a generated Scenario may carry.
	MaxAIFindings = 4
	// DefaultAIFindingPoints is used when a generated finding omits or misstates its points.
	DefaultAIFindingPoints = 10
)

var validVerdictSupport = map[string]bool{
	"malicious":  true,
	"benign":     true,
	"supporting": true,
}

var validScenarioDomainIDs = map[string]bool{
	"domain1": true,
	"domain2": true,
	"domain3": true,
	"domain4": true,
	"domain5": true,
}

// mappedAIFinding is an AI finding after its evidence messages have been resolved
// to 0-based log indices.
type mappedAIFinding struct {
	finding aiFinding
	indices []int
}

// buildScenarioLogs constructs the fixed 45-log SIEM environment around the
// AI-generated incident core. Index 16 is always the AI attack log message.
func buildScenarioLogs(ai aiMicroScenario, now time.Time) []map[string]interface{} {
	logs := make([]map[string]interface{}, 0, 46)

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

	return logs
}

// normalizeAIFindings returns the explicit findings, or synthesizes one finding
// from the legacy key_finding / attack_log_message fields.
func normalizeAIFindings(ai aiMicroScenario) []aiFinding {
	if len(ai.Findings) > 0 {
		out := make([]aiFinding, len(ai.Findings))
		copy(out, ai.Findings)
		return out
	}

	description := strings.TrimSpace(ai.KeyFinding)
	if description == "" {
		if ai.IsTruePositive {
			description = fmt.Sprintf("Identified malicious activity originating from %s targeting %s", ai.AttackerIP, ai.TargetHost)
		} else {
			description = fmt.Sprintf("Verified the activity involving %s and %s is benign and authorized", ai.TargetHost, ai.AttackerIP)
		}
	}

	support := "benign"
	if ai.IsTruePositive {
		support = "malicious"
	}

	messages := make([]string, 0, 1)
	if ai.IsTruePositive {
		// Preserve the legacy AI evidence contract: the whole attack-log block
		// (indices 15..19) is the decisive evidence for the single legacy finding.
		messages = append(messages, legacyAttackEvidenceMessages(ai)...)
	} else if strings.TrimSpace(ai.KeyFinding) != "" {
		messages = append(messages, ai.KeyFinding)
	} else if strings.TrimSpace(ai.AttackLogMessage) != "" {
		messages = append(messages, ai.AttackLogMessage)
	}

	return []aiFinding{
		{
			Description:         description,
			DomainID:            "",
			Points:              DefaultAIFindingPoints,
			EvidenceLogMessages: messages,
			VerdictSupport:      support,
			Explanation:         ai.TpFpExplanation,
		},
	}
}

// resolveEvidenceMessageIndex maps an exact log message to its index, then falls
// back to a trimmed, case-insensitive comparison. Duplicate log messages resolve
// to the first occurrence.
func resolveEvidenceMessageIndex(logs []map[string]interface{}, message string) (int, bool) {
	exact := make(map[string]int, len(logs))
	folded := make(map[string]int, len(logs))
	for i, entry := range logs {
		value, _ := entry["message"].(string)
		if _, exists := exact[value]; !exists {
			exact[value] = i
		}
		key := strings.ToLower(strings.TrimSpace(value))
		if key == "" {
			continue
		}
		if _, exists := folded[key]; !exists {
			folded[key] = i
		}
	}

	if index, ok := exact[message]; ok {
		return index, true
	}
	key := strings.ToLower(strings.TrimSpace(message))
	if key == "" {
		return 0, false
	}
	index, ok := folded[key]
	return index, ok
}

// mapAIFindingEvidence resolves a finding's exact evidence messages to log
// indices. Missing, duplicate, and out-of-range mappings are rejected.
func mapAIFindingEvidence(logs []map[string]interface{}, finding aiFinding) ([]int, []string) {
	label := strings.TrimSpace(finding.Description)
	if label == "" {
		label = "unnamed finding"
	}

	if len(finding.EvidenceLogMessages) == 0 {
		return nil, []string{fmt.Sprintf("finding %q has no evidence_log_messages", label)}
	}

	indices := make([]int, 0, len(finding.EvidenceLogMessages))
	errs := make([]string, 0)
	seen := make(map[int]bool, len(finding.EvidenceLogMessages))

	for _, message := range finding.EvidenceLogMessages {
		trimmed := strings.TrimSpace(message)
		if trimmed == "" {
			errs = append(errs, fmt.Sprintf("finding %q has an empty evidence message", label))
			continue
		}

		index, ok := resolveEvidenceMessageIndex(logs, message)
		if !ok {
			errs = append(errs, fmt.Sprintf("finding %q has an unmappable evidence message %q", label, message))
			continue
		}
		if index < 0 || index >= len(logs) {
			errs = append(errs, fmt.Sprintf("finding %q evidence index %d is out of range", label, index))
			continue
		}
		if seen[index] {
			errs = append(errs, fmt.Sprintf("finding %q has duplicate evidence index %d", label, index))
			continue
		}
		seen[index] = true
		indices = append(indices, index)
	}

	return indices, errs
}

// validateMappedFindings enforces the finding count, evidence union, verdict
// support, and False Positive benign-context rules.
func validateMappedFindings(mapped []mappedAIFinding, logs []map[string]interface{}, isTruePositive bool) []string {
	errs := make([]string, 0)

	if len(mapped) == 0 {
		return append(errs, "at least 1 key finding is required")
	}
	if len(mapped) > MaxAIFindings {
		return append(errs, fmt.Sprintf("generated scenario has %d findings; maximum is %d", len(mapped), MaxAIFindings))
	}

	logCount := len(logs)
	union := make(map[int]struct{}, logCount)
	hasMalicious := false
	hasBenign := false
	hasBenignContext := false

	for i, item := range mapped {
		label := strings.TrimSpace(item.finding.Description)
		if label == "" {
			label = fmt.Sprintf("finding %d", i+1)
		}

		if len(item.indices) == 0 {
			errs = append(errs, fmt.Sprintf("finding %q has no evidence", label))
			continue
		}

		seen := make(map[int]bool, len(item.indices))
		for _, index := range item.indices {
			if index < 0 || index >= logCount {
				errs = append(errs, fmt.Sprintf("finding %q evidence index %d is out of range", label, index))
				continue
			}
			if seen[index] {
				errs = append(errs, fmt.Sprintf("finding %q has duplicate evidence index %d", label, index))
				continue
			}
			seen[index] = true
			union[index] = struct{}{}
		}

		support := strings.ToLower(strings.TrimSpace(item.finding.VerdictSupport))
		if support == "" {
			errs = append(errs, fmt.Sprintf("finding %q is missing verdict_support", label))
			continue
		}
		if !validVerdictSupport[support] {
			errs = append(errs, fmt.Sprintf("finding %q has invalid verdict_support %q", label, item.finding.VerdictSupport))
			continue
		}

		switch support {
		case "malicious":
			hasMalicious = true
		case "benign":
			hasBenign = true
			for _, index := range item.indices {
				if index >= 0 && index < logCount && isBenignContextLog(logs[index]) {
					hasBenignContext = true
					break
				}
			}
		}
	}

	cap := logCount / 2
	if len(union) > cap {
		errs = append(errs, fmt.Sprintf("evidence union %d/%d exceeds the 50%% cap (%d logs)", len(union), logCount, cap))
	}

	if isTruePositive && !hasMalicious {
		errs = append(errs, "True Positive requires at least one malicious finding")
	}
	if !isTruePositive {
		if !hasBenign {
			errs = append(errs, "False Positive requires at least one benign finding")
		}
		if !hasBenignContext {
			errs = append(errs, "False Positive requires a benign context log")
		}
	}

	return errs
}

// synthesizedBenignContextMessage builds a deterministic approved-change log
// message for a False Positive Scenario whose LLM response omitted one.
func synthesizedBenignContextMessage(ai aiMicroScenario) string {
	host := strings.TrimSpace(ai.TargetHost)
	if host == "" {
		host = "the affected host"
	}
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(strings.Join([]string{ai.Title, ai.TargetHost, ai.AttackerIP}, "|")))
	changeNumber := 1000 + int(hasher.Sum32()%9000)
	return fmt.Sprintf("CHANGE_MANAGEMENT: Approved change CR-%d authorizes the observed activity on %s during this window.", changeNumber, host)
}

// ensureBenignContextLog guarantees that a False Positive Scenario's log set
// contains a strict benign context log. If the LLM supplied a benign message
// that is absent from the logs, that exact message is appended. Otherwise a
// deterministic approved-change log is synthesized. The context message is also
// attached to an existing benign finding so it becomes evidence.
func ensureBenignContextLog(ai *aiMicroScenario, logs []map[string]interface{}, now time.Time) []map[string]interface{} {
	if ai.IsTruePositive {
		return logs
	}
	for _, entry := range logs {
		if isStrictBenignContextLog(entry) {
			return logs
		}
	}

	benignFindingIndex := -1
	benignMessage := ""
	for i := range ai.Findings {
		if strings.ToLower(strings.TrimSpace(ai.Findings[i].VerdictSupport)) != "benign" {
			continue
		}
		if benignFindingIndex < 0 {
			benignFindingIndex = i
		}
		for _, message := range ai.Findings[i].EvidenceLogMessages {
			trimmed := strings.TrimSpace(message)
			if trimmed == "" {
				continue
			}
			if _, ok := resolveEvidenceMessageIndex(logs, trimmed); !ok {
				benignMessage = trimmed
				break
			}
		}
		if benignMessage != "" {
			break
		}
	}

	source := "change_management"
	if benignMessage == "" {
		benignMessage = synthesizedBenignContextMessage(*ai)
	} else if looksLikeHRContext(benignMessage) {
		source = "hr_system"
	}

	host := strings.TrimSpace(ai.TargetHost)
	if host == "" {
		host = "SRV-WEB01"
	}
	logs = append(logs, map[string]interface{}{
		"timestamp":      now.Add(time.Duration(360) * time.Second).Format(time.RFC3339),
		"source":         source,
		"severity":       "info",
		"severity_level": 1,
		"message":        benignMessage,
		"src_ip":         "10.0.0.5",
		"user":           "change_control",
		"host":           host,
	})

	if benignFindingIndex >= 0 {
		ai.Findings[benignFindingIndex].EvidenceLogMessages = appendUniqueMessage(
			ai.Findings[benignFindingIndex].EvidenceLogMessages,
			benignMessage,
		)
	}

	return logs
}

func legacyAttackEvidenceMessages(ai aiMicroScenario) []string {
	host := ai.TargetHost
	if host == "" {
		host = "SRV-WEB01"
	}
	ip := ai.AttackerIP
	if ip == "" {
		ip = "185.220.101.45"
	}
	messages := make([]string, 0, 5)
	messages = append(messages,
		fmt.Sprintf("Perimeter Firewall: Connection request initiated from %s to %s:443 [ACTION: ALLOWED]", host, ip),
		ai.AttackLogMessage,
		fmt.Sprintf("IDS Alert: Malicious attack payload signature matched on incoming traffic from %s targeting host %s", ip, host),
		fmt.Sprintf("HTTP 500 Internal Server Error - Anomaly detected on %s via POST payload from %s", host, ip),
		fmt.Sprintf("SIEM Alert Rule [CORR-9901]: Multi-stage incident correlation triggered on host %s involving IP %s", host, ip),
	)
	return messages
}

func appendUniqueMessage(messages []string, message string) []string {
	for _, existing := range messages {
		if existing == message {
			return messages
		}
	}
	return append(messages, message)
}

func looksLikeHRContext(message string) bool {
	normalized := strings.ToLower(message)
	for _, marker := range []string{"travel", "out of office", "human resources", "hr approval", "employee"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

// isStrictBenignContextLog identifies explicit benign context sources such as
// change management, HR travel, or scheduled maintenance.
func isStrictBenignContextLog(entry map[string]interface{}) bool {
	source, _ := entry["source"].(string)
	message, _ := entry["message"].(string)
	normalizedSource := strings.ToLower(strings.TrimSpace(source))
	normalizedMessage := strings.ToLower(strings.TrimSpace(message))

	for _, marker := range []string{"change", "hr", "maintenance", "itsm", "ticket", "service_desk"} {
		if strings.Contains(normalizedSource, marker) {
			return true
		}
	}
	for _, marker := range []string{
		"approved change",
		"change request",
		"change ticket",
		"cr-",
		"crq-",
		"travel",
		"out of office",
		"scheduled maintenance",
		"maintenance window",
		"authorized scan",
		"authorized vulnerability",
		"vulnerability scanning",
		"authorized activity",
		"approved maintenance",
	} {
		if strings.Contains(normalizedMessage, marker) {
			return true
		}
	}
	return false
}

// isBenignContextLog accepts strict context logs and non-alert informational
// evidence. Alert-only high/critical evidence never qualifies.
func isBenignContextLog(entry map[string]interface{}) bool {
	if isStrictBenignContextLog(entry) {
		return true
	}
	switch scenarioLogSeverityBucket(entry) {
	case "low":
		return true
	default:
		return false
	}
}

func scenarioLogSeverityBucket(entry map[string]interface{}) string {
	severity, _ := entry["severity"].(string)
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "high", "critical":
		return "high"
	case "info", "low":
		return "low"
	case "medium", "warning":
		return "medium"
	}
	if level, ok := evidenceIndexFromAny(entry["severity_level"]); ok {
		if level >= 4 {
			return "high"
		}
		if level <= 2 {
			return "low"
		}
		return "medium"
	}
	return ""
}

// fallbackAIFinding produces the single deterministic finding used when the
// generated findings cannot satisfy the evidence contract.
func fallbackAIFinding(ai aiMicroScenario, logs []map[string]interface{}, weakestDomain string) mappedAIFinding {
	description := strings.TrimSpace(ai.KeyFinding)
	domain := weakestDomain
	points := DefaultAIFindingPoints
	explanation := ai.TpFpExplanation

	if len(ai.Findings) > 0 {
		first := ai.Findings[0]
		if trimmed := strings.TrimSpace(first.Description); trimmed != "" {
			description = trimmed
		}
		if validScenarioDomainIDs[first.DomainID] {
			domain = first.DomainID
		}
		if normalized := normalizeAIFindingPoints(first.Points); normalized > 0 {
			points = normalized
		}
		if trimmed := strings.TrimSpace(first.Explanation); trimmed != "" {
			explanation = trimmed
		}
	}
	if description == "" {
		if ai.IsTruePositive {
			description = fmt.Sprintf("Identified malicious activity originating from %s targeting %s", ai.AttackerIP, ai.TargetHost)
		} else {
			description = fmt.Sprintf("Verified the activity involving %s and %s is benign and authorized", ai.TargetHost, ai.AttackerIP)
		}
	}

	support := "malicious"
	var indices []int
	if ai.IsTruePositive {
		indices = fallbackAttackEvidenceIndices(logs)
	} else {
		support = "benign"
		indices = benignContextEvidenceIndices(logs)
	}

	return mappedAIFinding{
		finding: aiFinding{
			Description:         description,
			DomainID:            domain,
			Points:              float64(points),
			EvidenceLogMessages: messagesAtIndices(logs, indices),
			VerdictSupport:      support,
			Explanation:         explanation,
		},
		indices: indices,
	}
}

func fallbackAttackEvidenceIndices(logs []map[string]interface{}) []int {
	if len(logs) >= 20 {
		return []int{15, 16, 17, 18, 19}
	}

	high := make([]int, 0, len(logs))
	for i, entry := range logs {
		if scenarioLogSeverityBucket(entry) == "high" {
			high = append(high, i)
		}
	}
	if len(high) > 0 {
		return high
	}

	start := len(logs) - 5
	if start < 0 {
		start = 0
	}
	indices := make([]int, 0, len(logs)-start)
	for i := start; i < len(logs); i++ {
		indices = append(indices, i)
	}
	return indices
}

func benignContextEvidenceIndices(logs []map[string]interface{}) []int {
	strict := make([]int, 0)
	lenient := make([]int, 0)
	for i, entry := range logs {
		if isStrictBenignContextLog(entry) {
			strict = append(strict, i)
		}
		if isBenignContextLog(entry) {
			lenient = append(lenient, i)
		}
	}
	if len(strict) > 0 {
		return strict
	}
	return lenient
}

func messagesAtIndices(logs []map[string]interface{}, indices []int) []string {
	messages := make([]string, 0, len(indices))
	for _, index := range indices {
		if index < 0 || index >= len(logs) {
			continue
		}
		message, _ := logs[index]["message"].(string)
		if message != "" {
			messages = append(messages, message)
		}
	}
	return messages
}

func normalizeAIFindingPoints(points float64) int {
	if math.IsNaN(points) || math.IsInf(points, 0) || points <= 0 || points != math.Trunc(points) || points > math.MaxInt32 {
		return DefaultAIFindingPoints
	}
	return int(points)
}

func buildExpectedOutcomes(mapped []mappedAIFinding, ai aiMicroScenario, weakestDomain string) map[string]interface{} {
	findings := make([]map[string]interface{}, 0, len(mapped))
	for i, item := range mapped {
		description := strings.TrimSpace(item.finding.Description)
		if description == "" {
			description = fmt.Sprintf("Key finding %d", i+1)
		}

		domain := item.finding.DomainID
		if !validScenarioDomainIDs[domain] {
			domain = weakestDomain
		}

		support := strings.ToLower(strings.TrimSpace(item.finding.VerdictSupport))
		if !validVerdictSupport[support] {
			support = "supporting"
		}

		explanation := strings.TrimSpace(item.finding.Explanation)
		if explanation == "" {
			explanation = ai.TpFpExplanation
		}

		findings = append(findings, map[string]interface{}{
			"id":                   fmt.Sprintf("f%d", i+1),
			"description":          description,
			"domain_id":            domain,
			"points":               normalizeAIFindingPoints(item.finding.Points),
			"evidence_log_indices": item.indices,
			"verdict_support":      support,
			"explanation":          explanation,
		})
	}

	return map[string]interface{}{"key_findings": findings}
}

// synthesizeFullScenarioWithErrors synthesizes the complete Scenario and returns
// the evidence-contract errors from the generated findings. The returned
// Scenario is always valid: if the generated findings fail validation, a single
// deterministic finding is used instead.
func synthesizeFullScenarioWithErrors(ai aiMicroScenario, weakestDomain string, domainName string) (*models.Scenario, []string) {
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

	if !strings.Contains(ai.Description, "Target Host:") && !strings.Contains(ai.Description, "INVESTIGATION CLUES") {
		ai.Description = fmt.Sprintf("%s\n\nINVESTIGATION CLUES:\n- Target Host: %s\n- Suspected Attacker IP: %s\n- Alert Vector: %s", ai.Description, ai.TargetHost, ai.AttackerIP, ai.Title)
	}

	ai.Findings = normalizeAIFindings(ai)
	now := time.Now().UTC()
	logs := buildScenarioLogs(ai, now)
	logs = ensureBenignContextLog(&ai, logs, now)

	mapped := make([]mappedAIFinding, 0, len(ai.Findings))
	errs := make([]string, 0)
	for _, finding := range ai.Findings {
		indices, mappingErrs := mapAIFindingEvidence(logs, finding)
		if len(mappingErrs) > 0 {
			errs = append(errs, mappingErrs...)
		}
		mapped = append(mapped, mappedAIFinding{finding: finding, indices: indices})
	}

	errs = append(errs, validateMappedFindings(mapped, logs, ai.IsTruePositive)...)
	if len(errs) > 0 {
		mapped = []mappedAIFinding{fallbackAIFinding(ai, logs, weakestDomain)}
	}

	containmentSteps := buildPhaseStepsWithDistractors("containment", "c", weakestDomain, ai.ContainmentAction, ai.ContainmentSteps, ai)
	eradicationSteps := buildPhaseStepsWithDistractors("eradication", "e", weakestDomain, ai.EradicationAction, ai.EradicationSteps, ai)
	recoverySteps := buildPhaseStepsWithDistractors("recovery", "r", weakestDomain, ai.RecoveryAction, ai.RecoverySteps, ai)

	return &models.Scenario{
		Title:            ai.Title,
		Description:      ai.Description,
		DomainID:         weakestDomain,
		Difficulty:       3,
		IsTruePositive:   ai.IsTruePositive,
		TpFpExplanation:  ai.TpFpExplanation,
		InitialLogs:      logs,
		ExpectedOutcomes: buildExpectedOutcomes(mapped, ai, weakestDomain),
		PlaybookSteps: map[string]interface{}{
			"containment": containmentSteps,
			"eradication": eradicationSteps,
			"recovery":    recoverySteps,
		},
		Status: "active",
	}, errs
}

// synthesizeWithEvidenceRetry runs one evidence-validation retry and then falls
// back to the deterministic single-finding Scenario. The retry callback is
// expected to request a corrected micro scenario from the LLM.
type aiScenarioRetryFunc func(validationErrors []string, availableMessages []string) (aiMicroScenario, error)

func synthesizeWithEvidenceRetry(initial aiMicroScenario, weakestDomain string, domainName string, retry aiScenarioRetryFunc) (*models.Scenario, []string) {
	scenario, validationErrors := synthesizeFullScenarioWithErrors(initial, weakestDomain, domainName)
	if len(validationErrors) == 0 || retry == nil {
		return scenario, validationErrors
	}

	availableMessages := scenarioLogMessages(scenario)
	revised, err := retry(validationErrors, availableMessages)
	if err != nil {
		log.Printf("[LLM] Evidence retry failed: %v", err)
		return scenario, validationErrors
	}

	retriedScenario, retriedErrors := synthesizeFullScenarioWithErrors(revised, weakestDomain, domainName)
	if retriedScenario == nil {
		return scenario, validationErrors
	}
	return retriedScenario, retriedErrors
}

func scenarioLogMessages(scenario *models.Scenario) []string {
	if scenario == nil {
		return nil
	}
	logs := normalizePhaseStepSlice(scenario.InitialLogs)
	messages := make([]string, 0, len(logs))
	for _, entry := range logs {
		message, _ := entry["message"].(string)
		if message != "" {
			messages = append(messages, message)
		}
	}
	return messages
}

func buildEvidenceRetryPrompt(validationErrors []string, availableMessages []string) string {
	var builder strings.Builder
	builder.WriteString("Your previous scenario JSON did not satisfy the Key Finding evidence contract. Validation errors:\n")
	for _, validationError := range validationErrors {
		builder.WriteString("- ")
		builder.WriteString(validationError)
		builder.WriteString("\n")
	}
	builder.WriteString("\nReturn the complete scenario JSON again. Every findings[].evidence_log_messages entry MUST be copied exactly from these generated log messages:\n")
	for _, message := range availableMessages {
		builder.WriteString("- ")
		builder.WriteString(message)
		builder.WriteString("\n")
	}
	builder.WriteString("\nDo not invent messages. Return 1-4 findings. A True Positive needs at least one malicious finding. A False Positive needs at least one benign finding pointing to a benign context log (approved change, HR travel, scheduled maintenance, authorized scan). Keep the evidence union at or below 50% of the logs.")
	return builder.String()
}

// strictEvidenceIndices decodes an authored evidence array and rejects empty,
// non-integer, negative, out-of-range, and duplicate indices.
func strictEvidenceIndices(raw any, logCount int) ([]int, []string) {
	if raw == nil {
		return nil, []string{"evidence_log_indices is missing"}
	}

	indices := make([]int, 0)
	errs := make([]string, 0)
	seen := make(map[int]bool)

	appendValue := func(value any) {
		index, ok := evidenceIndexFromAny(value)
		if !ok {
			errs = append(errs, fmt.Sprintf("evidence_log_indices contains an invalid index %v", value))
			return
		}
		if index < 0 {
			errs = append(errs, fmt.Sprintf("evidence_log_indices contains an invalid index %d", index))
			return
		}
		if index >= logCount {
			errs = append(errs, fmt.Sprintf("evidence_log_indices value %d is out of range for %d logs", index, logCount))
			return
		}
		if seen[index] {
			errs = append(errs, fmt.Sprintf("evidence_log_indices contains duplicate index %d", index))
			return
		}
		seen[index] = true
		indices = append(indices, index)
	}

	switch values := raw.(type) {
	case []int:
		for _, value := range values {
			appendValue(value)
		}
	case []int64:
		for _, value := range values {
			appendValue(value)
		}
	case []float64:
		for _, value := range values {
			appendValue(value)
		}
	case []json.Number:
		for _, value := range values {
			appendValue(value)
		}
	case []interface{}:
		for _, value := range values {
			appendValue(value)
		}
	default:
		return nil, []string{fmt.Sprintf("evidence_log_indices has unsupported type %T", raw)}
	}

	if len(indices) == 0 && len(errs) == 0 {
		errs = append(errs, "evidence_log_indices is empty")
	}
	return indices, errs
}

// validateScenarioEvidenceContract enforces the runtime and static evidence
// contract over a Scenario's expected_outcomes.
func validateScenarioEvidenceContract(s *models.Scenario) error {
	return validateScenarioEvidenceContractWithLegacySupport(s, false)
}

// validateScenarioEvidenceContractForRemediation evaluates pre-T7 rows that omit
// verdict_support. Missing support is inferred from the Scenario's TP/FP flag so
// a legacy True Positive with attack-log evidence remains valid while a legacy
// False Positive still has to prove a benign context log.
func validateScenarioEvidenceContractForRemediation(s *models.Scenario) error {
	return validateScenarioEvidenceContractWithLegacySupport(s, true)
}

func validateScenarioEvidenceContractWithLegacySupport(s *models.Scenario, inferMissingSupport bool) error {
	if s == nil {
		return fmt.Errorf("scenario is nil")
	}

	logs := normalizePhaseStepSlice(s.InitialLogs)
	logCount := len(logs)
	if logCount == 0 {
		return fmt.Errorf("content_validation_failed: initial_logs are unavailable")
	}

	outcomes, ok := s.ExpectedOutcomes.(map[string]interface{})
	if !ok {
		return fmt.Errorf("schema_validation_failed: expected_outcomes must be an object")
	}
	rawFindings, exists := outcomes["key_findings"]
	if !exists {
		return fmt.Errorf("schema_validation_failed: key_findings missing from expected_outcomes")
	}

	findings := mapSlice(rawFindings)
	if len(findings) == 0 {
		return fmt.Errorf("content_validation_failed: at least 1 key finding is required")
	}
	if len(findings) > MaxAIFindings {
		return fmt.Errorf("content_validation_failed: scenario has %d key findings; maximum is %d", len(findings), MaxAIFindings)
	}

	union := make(map[int]struct{}, logCount)
	hasMalicious := false
	hasBenign := false
	hasBenignContext := false

	for i, finding := range findings {
		findingID, _ := finding["id"].(string)
		if strings.TrimSpace(findingID) == "" {
			findingID = fmt.Sprintf("f%d", i+1)
		}

		indices, indexErrs := strictEvidenceIndices(finding["evidence_log_indices"], logCount)
		if len(indexErrs) > 0 {
			return fmt.Errorf("content_validation_failed: finding %q: %s", findingID, strings.Join(indexErrs, "; "))
		}
		if len(indices) == 0 {
			return fmt.Errorf("content_validation_failed: finding %q has no evidence", findingID)
		}

		supportRaw, _ := finding["verdict_support"].(string)
		support := strings.ToLower(strings.TrimSpace(supportRaw))
		if support == "" && inferMissingSupport {
			if s.IsTruePositive {
				support = "malicious"
			} else {
				support = "benign"
			}
		}
		if support == "" {
			return fmt.Errorf("content_validation_failed: finding %q is missing verdict_support", findingID)
		}
		if !validVerdictSupport[support] {
			return fmt.Errorf("content_validation_failed: finding %q has invalid verdict_support %q", findingID, supportRaw)
		}

		for _, index := range indices {
			union[index] = struct{}{}
		}
		switch support {
		case "malicious":
			hasMalicious = true
		case "benign":
			hasBenign = true
			for _, index := range indices {
				if isBenignContextLog(logs[index]) {
					hasBenignContext = true
					break
				}
			}
		}
	}

	cap := logCount / 2
	if len(union) > cap {
		return fmt.Errorf("content_validation_failed: evidence union %d/%d exceeds the 50%% cap (%d logs)", len(union), logCount, cap)
	}
	if s.IsTruePositive && !hasMalicious {
		return fmt.Errorf("content_validation_failed: True Positive requires at least one malicious finding")
	}
	if !s.IsTruePositive {
		if !hasBenign {
			return fmt.Errorf("content_validation_failed: False Positive requires at least one benign finding")
		}
		if !hasBenignContext {
			return fmt.Errorf("content_validation_failed: False Positive requires a benign context log")
		}
	}

	return nil
}

// EnsureScenarioEvidenceQuality loads the Scenario seed from a path and runs the
// startup remediation. The path-based wrapper is kept for tests and tooling; the
// server uses the embedded seed through EnsureScenarioEvidenceQualityFromSeed so
// the seed is available in the runtime image.
func EnsureScenarioEvidenceQuality(db *gorm.DB, seedPath string) error {
	if seedPath == "" {
		seedPath = "scenarios_seed.json"
	}
	rawBytes, err := os.ReadFile(seedPath)
	if err != nil {
		return fmt.Errorf("read scenario seed %q: %w", seedPath, err)
	}
	return EnsureScenarioEvidenceQualityFromSeed(db, rawBytes)
}

// EnsureScenarioEvidenceQualityFromSeed archives legacy False Positive Scenarios
// that do not satisfy the evidence contract. Seed-authored static Scenarios
// (matched by title to the seed), already archived rows, and True Positive rows
// are never candidates. The remediation only updates Scenario status; it never
// modifies Simulation Sessions.
func EnsureScenarioEvidenceQualityFromSeed(db *gorm.DB, seedJSON []byte) error {
	seedTitles, err := loadSeedScenarioTitlesFromJSON(seedJSON)
	if err != nil {
		return err
	}
	if db == nil {
		return nil
	}

	var stored []models.Scenario
	if err := db.Find(&stored).Error; err != nil {
		return err
	}

	archivedCount := 0
	for i := range stored {
		scenario := &stored[i]
		shouldArchive, validationErr := shouldArchiveScenarioEvidence(scenario, seedTitles)
		if !shouldArchive {
			continue
		}
		if updateErr := db.Model(&models.Scenario{}).
			Where("scenario_id = ? AND status <> 'archived'", scenario.ScenarioID).
			Update("status", "archived").Error; updateErr != nil {
			log.Printf("[Scenario Evidence Quality Gate] Failed to archive scenario %q (%s): %v", scenario.Title, scenario.ScenarioID, updateErr)
		} else {
			archivedCount++
			log.Printf("[Scenario Evidence Quality Gate] Archived invalid legacy False Positive scenario %q (%s): %v", scenario.Title, scenario.ScenarioID, validationErr)
		}
	}

	if archivedCount > 0 {
		log.Printf("[Scenario Evidence Quality Gate] Archived %d invalid legacy False Positive scenario(s).", archivedCount)
	}
	return nil
}

// shouldArchiveScenarioEvidence is the pure remediation policy used at startup.
// Static authored Scenarios (matched to the seed by title), already archived
// rows, and True Positive rows are never candidates. Only legacy False Positive
// rows that remain invalid after inferring missing verdict support are archived.
// The caller never touches Simulation Sessions.
func shouldArchiveScenarioEvidence(scenario *models.Scenario, seedTitles map[string]struct{}) (bool, error) {
	if scenario == nil {
		return false, nil
	}
	if strings.EqualFold(strings.TrimSpace(scenario.Status), "archived") {
		return false, nil
	}
	if _, isStatic := seedTitles[normalizeScenarioTitle(scenario.Title)]; isStatic {
		return false, nil
	}

	// Evaluate every candidate with legacy support inference so pre-T7 True
	// Positive rows are validated as malicious rather than rejected for a
	// missing verdict_support field.
	validationErr := validateScenarioEvidenceContractForRemediation(scenario)
	if validationErr == nil {
		return false, nil
	}
	if scenario.IsTruePositive {
		// Only invalid False Positive rows are archived. True Positive rows are
		// left active even when they fail the evidence contract.
		return false, nil
	}
	return true, validationErr
}

func loadSeedScenarioTitles(seedPath string) (map[string]struct{}, error) {
	rawBytes, err := os.ReadFile(seedPath)
	if err != nil {
		return nil, err
	}
	return loadSeedScenarioTitlesFromJSON(rawBytes)
}

func loadSeedScenarioTitlesFromJSON(seedJSON []byte) (map[string]struct{}, error) {
	if len(seedJSON) == 0 {
		return nil, fmt.Errorf("scenario seed is empty")
	}

	var seedScenarios []models.Scenario
	if err := json.Unmarshal(seedJSON, &seedScenarios); err != nil {
		return nil, fmt.Errorf("scenario seed is not valid JSON: %w", err)
	}

	titles := make(map[string]struct{}, len(seedScenarios))
	for _, scenario := range seedScenarios {
		title := normalizeScenarioTitle(scenario.Title)
		if title == "" {
			continue
		}
		titles[title] = struct{}{}
	}
	if len(titles) == 0 {
		return nil, fmt.Errorf("scenario seed contains no scenario titles")
	}
	return titles, nil
}

func normalizeScenarioTitle(title string) string {
	return strings.ToLower(strings.TrimSpace(title))
}
