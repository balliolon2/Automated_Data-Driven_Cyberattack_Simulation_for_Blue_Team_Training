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

type aiFinding struct {
	Description         string   `json:"description"`
	DomainID            string   `json:"domain_id"`
	Points              float64  `json:"points"`
	EvidenceLogMessages []string `json:"evidence_log_messages"`
	VerdictSupport      string   `json:"verdict_support"` // malicious | benign | supporting
	Explanation         string   `json:"explanation"`
}

type aiIncidentLog struct {
	Source            string `json:"source"`
	Severity          string `json:"severity"` // info | low | warning | medium | high | critical
	Host              string `json:"host"`
	SrcIP             string `json:"src_ip,omitempty"`
	DstIP             string `json:"dst_ip,omitempty"`
	User              string `json:"user,omitempty"`
	Message           string `json:"message"`
	RelativeOffsetSec int    `json:"relative_offset_sec,omitempty"`
}

type aiMicroScenario struct {
	Title             string                   `json:"title"`
	Description       string                   `json:"description"`
	SubTopicID        string                   `json:"sub_topic_id,omitempty"`
	Archetype         string                   `json:"archetype,omitempty"`
	BriefingStyle     string                   `json:"briefing_style,omitempty"`
	IsTruePositive    bool                     `json:"is_true_positive"`
	TpFpExplanation   string                   `json:"tp_fp_explanation"`
	AttackerIP        string                   `json:"attacker_ip"`
	TargetHost        string                   `json:"target_host"`
	PrimaryUser       string                   `json:"primary_user,omitempty"`
	IncidentLogs      []aiIncidentLog          `json:"incident_logs,omitempty"`
	AttackLogMessage  string                   `json:"attack_log_message,omitempty"`
	KeyFinding        string                   `json:"key_finding,omitempty"`
	Findings          []aiFinding              `json:"findings,omitempty"`
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

// ScenarioHistoryEntry summarizes one prior Simulation Session for rotation and history exclusion.
type ScenarioHistoryEntry struct {
	Title           string
	SubTopicID      string
	BriefingStyle   string
	TargetHost      string
	AttackerIP      string
	IsTruePositive  bool
	SkipCadenceVote bool
}

// ScenarioDiversityContext defines the target SY0-701 sub-topic, operational archetype,
// narrative briefing style, target TP/FP verdict, and recent entity exclusions for LLM synthesis.
type ScenarioDiversityContext struct {
	TargetSubTopicID     string
	TargetSubTopicTitle  string
	TargetArchetype      string
	TargetBriefingStyle  string
	TargetIsTruePositive bool
	RecentTitles         []string
	RecentTargetHosts    []string
	RecentAttackerIPs    []string
	RecentSubTopicIDs    []string
}

// BriefingStyles enumerates the 5 rotated SOC narrative opening formats.
var BriefingStyles = []string{
	"siem_escalation_ticket",
	"shift_handover_brief",
	"threat_hunting_lead",
	"helpdesk_user_report",
	"cloud_guardrail_alert",
}

var briefingStyleInstructions = map[string]string{
	"siem_escalation_ticket": "Write 'description' as a concise Tier-1 SIEM escalation ticket (e.g., citing rule trigger context, affected host/account, and initial alert telemetry).",
	"shift_handover_brief":   "Write 'description' as a SOC shift handover note from the outgoing analyst summarizing an unfolding anomaly and what telemetry warrants inspection.",
	"threat_hunting_lead":    "Write 'description' as a proactive threat-hunting lead triggered by anomalous baseline telemetry across endpoints, identity, or network logs.",
	"helpdesk_user_report":   "Write 'description' as an escalated IT helpdesk/user report where a user or asset owner flagged unexpected behavior corroborated by system logs.",
	"cloud_guardrail_alert":  "Write 'description' as an automated CSPM, IAM, or DLP compliance guardrail notification flagging a risky policy or data-flow event.",
}

var subtopicArchetypes = map[string][2]string{
	"1.1": {"Compensating Security Control Bypass & Audit Trail Tampering", "Authorized Red-Team Security Control Validation Exercise"},
	"1.2": {"Zero Trust Policy Violation & Unverified Device Access", "Approved Zero Trust Micro-Segmentation Rollout Verification"},
	"1.3": {"Unauthorized Production Change Bypassing CAB Approval", "Emergency Approved Change Request Verification"},
	"1.4": {"Weak TLS Cipher Downgrade & Compromised PKI Certificate", "Scheduled Internal Root CA Certificate Rotation"},
	"2.1": {"Targeted APT Reconnaissance & Threat-Intel IOC Match", "Authorized Threat-Intel Feed Validation Probe"},
	"2.2": {"Spearphishing Attachment & Living-off-the-Land Execution", "Internal Security Awareness Phishing Simulation Drill"},
	"2.3": {"Exploited API Deserialization Vulnerability & Web Shell", "Authorized Authenticated Vulnerability Scanner Sweep"},
	"2.4": {"Beaconing C2 Implant & Lateral SMB/WMI Movement", "Approved Endpoint Telemetry Stress Test Exercise"},
	"2.5": {"Endpoint Patch Drift & Missing EDR Containment Policy", "Scheduled Maintenance Patch Reboot Cascade"},
	"3.1": {"DMZ / OT Network Segmentation Bypass Pivot", "Documented Inter-Zone Firewall Migration Cutover"},
	"3.2": {"Cloud Security Group / Bastion Exposure & Lateral Pivot", "Approved Cloud Staging Bastion Troubleshooting"},
	"3.3": {"Sensitive PII/Cardholder Exfiltration via Cloud Sync or DLP", "Scheduled Compliance Cold-Storage Archival Transfer"},
	"3.4": {"Backup Shadow Copy Destruction & Ransomware Staging", "Planned Disaster Recovery Backup Restore Verification"},
	"4.1": {"Container Workload Escape & Privileged Pod Execution", "Scheduled Kubernetes Node Draining & Rebalance"},
	"4.2": {"Shadow IT Rogue Asset & Unmanaged Endpoint on LAN", "Approved Hardware Asset Onboarding Batch"},
	"4.3": {"Unpatched Critical CVE Active Exploitation Attempt", "Authenticated Maintenance-Window Vulnerability Scan"},
	"4.4": {"DNS TXT Covert Channel & Anomalous Egress Telemetry", "Authorized Telemetry Collector DNS Health Burst"},
	"4.5": {"Firewall or Proxy Policy Tampering & Logging Evasion", "Approved SIEM Correlation Threshold Tuning Rollout"},
	"4.6": {"MFA Push Fatigue & OAuth/SSO Token Hijacking", "Verified Business Travel VPN & MFA Exception"},
	"4.7": {"Compromised CI/CD Automation Token & SOAR Webhook Abuse", "Scheduled Automated Infrastructure-as-Code Sync"},
	"4.8": {"Active Malware Containment & Scheduled Persistence Removal", "Tabletop Incident Response Playbook Verification"},
	"4.9": {"Memory Process Injection & Forensic Artifact Tampering", "Authorized Digital Forensics Triage Collection"},
	"5.1": {"Privileged Governance Override & Separation-of-Duties Breach", "Scheduled IAM Governance Entitlement Review"},
	"5.2": {"Accepted Risk Threshold Breach on Critical Business Asset", "Documented Risk Exception Monitoring Window"},
	"5.3": {"Third-Party Vendor VPN Credential Abuse & Supply-Chain Pivot", "Approved Vendor Support Window Remote Access"},
	"5.4": {"Cleartext Regulated PII/PCI Transmission Compliance Breach", "Annual External Compliance Audit Evidence Collection"},
	"5.5": {"Unauthorized Privilege Escalation Flagged During Internal Audit", "Pre-Authorized Penetration Test Audit Window"},
	"5.6": {"Social Engineering Helpdesk MFA Reset & Account Takeover", "Verified Employee Helpdesk Onboarding Credential Reset"},
}

func archetypeForSubtopic(subTopicID string, isTruePositive bool) string {
	pair, ok := subtopicArchetypes[subTopicID]
	if !ok {
		if isTruePositive {
			return "Suspicious Multi-Stage Intrusion Activity"
		}
		return "Authorized Operational Change Verification"
	}
	if isTruePositive {
		return pair[0]
	}
	return pair[1]
}

// BuildScenarioDiversityContext computes the next SY0-701 sub-topic, operational archetype,
// briefing style, TP/FP cadence target (~70% TP / ~30% FP), and history exclusions from recent sessions.
func BuildScenarioDiversityContext(weakestDomain string, history []ScenarioHistoryEntry) ScenarioDiversityContext {
	if weakestDomain == "" {
		weakestDomain = "domain4"
	}

	// 1. Controlled ~70% TP / ~30% FP cadence (never > 2 consecutive FPs, inject FP after 2 TPs)
	targetTP := true
	var knownVerdicts []bool
	for _, entry := range history {
		if entry.SkipCadenceVote {
			continue
		}
		knownVerdicts = append(knownVerdicts, entry.IsTruePositive)
		if len(knownVerdicts) == 2 {
			break
		}
	}
	if len(knownVerdicts) > 0 {
		if !knownVerdicts[0] {
			targetTP = true
		} else if len(knownVerdicts) >= 2 && knownVerdicts[0] && knownVerdicts[1] {
			targetTP = false
		}
	}

	// 2. Sub-topic rotation across the domain's SY0-701 objectives (unvisited first, else LRU)
	subs := SubtopicsForDomain(weakestDomain)
	if len(subs) == 0 {
		subs = SubtopicsForDomain("domain4")
	}

	lastSeenIdx := make(map[string]int, len(subs))
	for _, spec := range subs {
		lastSeenIdx[spec.SubTopicID] = -1
	}
	for idx, entry := range history {
		id := strings.TrimSpace(entry.SubTopicID)
		if id == "" {
			continue
		}
		if prev, exists := lastSeenIdx[id]; exists && prev == -1 {
			// History is ordered most-recent-first, so smaller idx = more recent
			lastSeenIdx[id] = idx
		}
	}

	chosen := subs[0]
	foundUnvisited := false
	for _, spec := range subs {
		if lastSeenIdx[spec.SubTopicID] == -1 {
			chosen = spec
			foundUnvisited = true
			break
		}
	}
	if !foundUnvisited {
		oldestIdx := -1
		for _, spec := range subs {
			if lastSeenIdx[spec.SubTopicID] > oldestIdx {
				oldestIdx = lastSeenIdx[spec.SubTopicID]
				chosen = spec
			}
		}
	}

	// 3. Briefing style rotation (cycle sequentially across all 5 styles even when history is bounded)
	style := BriefingStyles[len(history)%len(BriefingStyles)]
	if len(history) > 0 && strings.TrimSpace(history[0].BriefingStyle) != "" {
		prevStyle := strings.TrimSpace(history[0].BriefingStyle)
		for idx, candidate := range BriefingStyles {
			if candidate == prevStyle {
				style = BriefingStyles[(idx+1)%len(BriefingStyles)]
				break
			}
		}
		if style == prevStyle {
			style = BriefingStyles[(len(history)+1)%len(BriefingStyles)]
		}
	}

	// 4. Recent history exclusions (deduplicated, bounded to 10)
	addUnique := func(dst []string, seen map[string]bool, val string) []string {
		trimmed := strings.TrimSpace(val)
		if trimmed == "" {
			return dst
		}
		key := strings.ToLower(trimmed)
		if seen[key] {
			return dst
		}
		seen[key] = true
		return append(dst, trimmed)
	}

	var recentTitles, recentHosts, recentIPs, recentSubTopics []string
	seenTitles := make(map[string]bool)
	seenHosts := make(map[string]bool)
	seenIPs := make(map[string]bool)
	seenSubs := make(map[string]bool)

	for _, entry := range history {
		if len(recentTitles) < 10 {
			recentTitles = addUnique(recentTitles, seenTitles, entry.Title)
		}
		if len(recentHosts) < 10 {
			recentHosts = addUnique(recentHosts, seenHosts, entry.TargetHost)
		}
		if len(recentIPs) < 10 {
			recentIPs = addUnique(recentIPs, seenIPs, entry.AttackerIP)
		}
		if len(recentSubTopics) < 10 {
			recentSubTopics = addUnique(recentSubTopics, seenSubs, entry.SubTopicID)
		}
	}

	return ScenarioDiversityContext{
		TargetSubTopicID:     chosen.SubTopicID,
		TargetSubTopicTitle:  chosen.Title,
		TargetArchetype:      archetypeForSubtopic(chosen.SubTopicID, targetTP),
		TargetBriefingStyle:  style,
		TargetIsTruePositive: targetTP,
		RecentTitles:         recentTitles,
		RecentTargetHosts:    recentHosts,
		RecentAttackerIPs:    recentIPs,
		RecentSubTopicIDs:    recentSubTopics,
	}
}

// GenerateLLMScenario calls an OpenAI-compatible API to dynamically synthesize a SOC scenario based on learner skill gaps.
// ponytail: Standard library net/http only, zero heavy external LLM SDK dependencies.
func GenerateLLMScenario(ctx context.Context, weakestDomain string, domainName string, proficiency float64, profiles []models.UserSkillProfile) (*models.Scenario, error) {
	divCtx := BuildScenarioDiversityContext(weakestDomain, nil)
	return GenerateLLMScenarioWithContext(ctx, weakestDomain, domainName, proficiency, profiles, divCtx)
}

// GenerateLLMScenarioWithContext calls an OpenAI-compatible API using rotated SY0-701 sub-topic,
// archetype, briefing style, TP/FP cadence, and session history exclusions.
func GenerateLLMScenarioWithContext(ctx context.Context, weakestDomain string, domainName string, proficiency float64, profiles []models.UserSkillProfile, divCtx ScenarioDiversityContext) (*models.Scenario, error) {
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

	// Build user profile summary and diversity / exclusion directives
	var profileSummary strings.Builder
	for _, p := range profiles {
		profileSummary.WriteString(fmt.Sprintf("- Domain %s: %.1f%% proficiency\n", p.DomainID, p.ProficiencyScore))
	}

	if divCtx.TargetSubTopicID == "" {
		divCtx = BuildScenarioDiversityContext(weakestDomain, nil)
	}
	verdictDirective := "true (True Positive — authentic security threat requiring containment and eradication)"
	if !divCtx.TargetIsTruePositive {
		verdictDirective = "false (False Positive — benign authorized business/maintenance activity backed by a Benign Context Log)"
	}
	styleGuide := briefingStyleInstructions[divCtx.TargetBriefingStyle]
	if styleGuide == "" {
		styleGuide = briefingStyleInstructions["siem_escalation_ticket"]
	}

	var exclusionLines strings.Builder
	if len(divCtx.RecentTitles) > 0 {
		exclusionLines.WriteString(fmt.Sprintf("- Avoid repeating recent plots/titles: %s\n", strings.Join(divCtx.RecentTitles, "; ")))
	}
	if len(divCtx.RecentTargetHosts) > 0 {
		exclusionLines.WriteString(fmt.Sprintf("- Avoid reusing recent target hosts: %s\n", strings.Join(divCtx.RecentTargetHosts, ", ")))
	}
	if len(divCtx.RecentAttackerIPs) > 0 {
		exclusionLines.WriteString(fmt.Sprintf("- Avoid reusing recent IP addresses: %s\n", strings.Join(divCtx.RecentAttackerIPs, ", ")))
	}
	if exclusionLines.Len() == 0 {
		exclusionLines.WriteString("- None yet; choose fresh realistic enterprise hostnames, usernames, and RFC 5737 / internal IPs.\n")
	}

	systemPrompt := "You are a specialized CompTIA Security+ 701 cybersecurity scenario generator. Output ONLY raw JSON. Every scenario MUST author a realistic 4-to-8 event incident_logs chain matching its specific domain and environment (never generic web/CORR-9901 boilerplate), follow the assigned SY0-701 sub-topic, archetype, briefing style, and target verdict, and every incident response phase (containment_steps, eradication_steps, recovery_steps) MUST contain 3 realistic SOC options (1 correct option and 2 plausible technical distractors) with matching sentence length (15-30 words) and specific entity references."
	userPrompt := fmt.Sprintf(`Generate a unique SOC scenario for target domain %s (%s). Current proficiency: %.1f%%.

Target Scenario Blueprint:
- SY0-701 Sub-Topic: %s (%s)
- Operational Archetype: %s
- Required Verdict (is_true_positive): %s
- Narrative Briefing Style (%s): %s Naturally mention the primary target host and actor IP or username in the prose (do NOT append an "INVESTIGATION CLUES:" bullet block).

DO NOT REPEAT (Recent Learner History Exclusions):
%s
Learner Profile:
%s

Rules for Incident Event Chain (incident_logs):
1. Return 4 to 8 chronological telemetry events in "incident_logs".
2. Each event MUST include: source, severity (info, low, warning, medium, high, or critical), host, src_ip, dst_ip, user, message, and relative_offset_sec (increasing integers such as 0, 45, 110, 185).
3. Choose realistic log sources matching THIS specific scenario's technology stack (e.g., okta, azure_ad, ad_controller, aws_cloudtrail, s3_access, k8s_audit, edr, sysmon, file_integrity, dns_server, vpn_gateway, proxy_gateway, nginx, postgresql, dlp_sensor, email_gateway, change_management, hr_portal, itsm_ticketing, nids, firewall). Do NOT force web_app/HTTP 500 or generic SIEM correlation messages unless the plot is specifically a web application incident.
4. Mix severities realistically across the chain (include supporting info/warning telemetry such as authentication, configuration, DNS, or access events alongside higher-severity alerts).
5. If "is_true_positive" is false, at least one event inside "incident_logs" MUST be an authentic benign context log (e.g., source "change_management", "hr_portal", or "itsm_ticketing" showing an approved change CR-XXXX, HR travel record, scheduled maintenance, or authorized scan), and NEVER emit fake "Malicious attack payload signature matched" alerts.
6. The backend automatically weaves your incident_logs into a variable-length (28-46 row) environment-matched SIEM timeline with decoy warning noise.

Rules for Key Findings (findings):
1. Return 1 to 4 findings.
2. Each finding must contain: description, domain_id (one of domain1..domain5), points (positive integer), evidence_log_messages (one or more exact log message strings copied verbatim from incident_logs[].message), verdict_support (malicious, benign, or supporting), and explanation.
3. Every string in evidence_log_messages MUST match an exact "message" value present in your "incident_logs" array.
4. If "is_true_positive" is true, at least one finding MUST have "verdict_support": "malicious".
5. If "is_true_positive" is false, at least one finding MUST have "verdict_support": "benign" and must point to the benign context log in incident_logs.
6. Prefer the minimum decisive set of evidence logs. Never invent unprinted log messages. Keep the union of all evidence_log_messages <= 50%% of the timeline.

Rules for Playbook Options (containment_steps, eradication_steps, recovery_steps):
1. Provide 3 options per phase: 1 option with "is_correct": true and 2 options with "is_correct": false.
2. Length & Specificity Parity: ALL options (both correct and false distractors) MUST be 15-30 words long, written in realistic SOC operational language, and explicitly reference the scenario's target_host, attacker_ip, affected accounts, or services.
3. Plausible Distractors: Never use absurd or generic distractors. Distractors must represent subtle SOC analytical missteps (e.g., premature action that destroys volatile forensic evidence, blocking an internal gateway instead of attacker_ip, applying a control from the wrong IR phase, or restoring from an unverified snapshot before patching).
4. If "is_true_positive" is false (False Positive scenario), "is_correct": true options must focus on verifying authorized change tickets, tuning SIEM correlation thresholds, and documenting benign activity without disrupting production, while "is_correct": false distractors represent unnecessary host quarantine or account lockout.

Output ONLY valid JSON matching this exact structure:
{
  "title": "Short Scenario Title",
  "description": "Brief 2-3 sentence briefing following the assigned briefing style and naming the primary host/actor naturally",
  "sub_topic_id": "%s",
  "archetype": "%s",
  "briefing_style": "%s",
  "is_true_positive": %t,
  "tp_fp_explanation": "Detailed explanation why this is a True or False Positive incident",
  "attacker_ip": "185.220.101.45",
  "target_host": "SRV-WEB01",
  "primary_user": "svc_deploy",
  "incident_logs": [
    {
      "source": "ad_controller",
      "severity": "info",
      "host": "SRV-WEB01",
      "src_ip": "185.220.101.45",
      "dst_ip": "10.20.4.18",
      "user": "svc_deploy",
      "message": "Initial contextual telemetry event matching this scenario",
      "relative_offset_sec": 0
    },
    {
      "source": "edr",
      "severity": "high",
      "host": "SRV-WEB01",
      "src_ip": "185.220.101.45",
      "dst_ip": "10.20.4.18",
      "user": "svc_deploy",
      "message": "Primary suspicious or decisive execution event on SRV-WEB01",
      "relative_offset_sec": 75
    },
    {
      "source": "firewall",
      "severity": "warning",
      "host": "SRV-WEB01",
      "src_ip": "10.20.4.18",
      "dst_ip": "185.220.101.45",
      "user": "svc_deploy",
      "message": "Follow-on outbound connection or lateral movement indicator",
      "relative_offset_sec": 150
    },
    {
      "source": "file_integrity",
      "severity": "critical",
      "host": "SRV-WEB01",
      "src_ip": "185.220.101.45",
      "dst_ip": "10.20.4.18",
      "user": "svc_deploy",
      "message": "Persistence or impact telemetry event on SRV-WEB01",
      "relative_offset_sec": 240
    }
  ],
  "findings": [
    {
      "description": "Decisive finding supported by the evidence",
      "domain_id": "domain4",
      "points": 10,
      "evidence_log_messages": ["Primary suspicious or decisive execution event on SRV-WEB01"],
      "verdict_support": "malicious",
      "explanation": "Why this evidence proves the finding"
    }
  ],
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
}`, weakestDomain, domainName, proficiency, divCtx.TargetSubTopicID, divCtx.TargetSubTopicTitle, divCtx.TargetArchetype, verdictDirective, divCtx.TargetBriefingStyle, styleGuide, exclusionLines.String(), profileSummary.String(), divCtx.TargetSubTopicID, divCtx.TargetArchetype, divCtx.TargetBriefingStyle, divCtx.TargetIsTruePositive)

	messages := []chatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	rawJSON, err := callChatCompletion(ctx, url, cleanKey, model, messages, 0.3)
	if err != nil {
		return nil, err
	}

	micro, err := parseAIMicroScenario(rawJSON)
	if err != nil {
		return nil, err
	}

	// Synthesize the full SIEM environment, validate the evidence contract once
	// more with the LLM, then fall back to a deterministic single finding.
	scenario, evidenceErrs := synthesizeWithEvidenceRetry(*micro, weakestDomain, domainName, func(validationErrors []string, availableMessages []string) (aiMicroScenario, error) {
		retryMessages := make([]chatMessage, 0, len(messages)+1)
		retryMessages = append(retryMessages, messages...)
		retryMessages = append(retryMessages, chatMessage{
			Role:    "user",
			Content: buildEvidenceRetryPrompt(validationErrors, availableMessages),
		})

		retryRaw, retryErr := callChatCompletion(ctx, url, cleanKey, model, retryMessages, 0.1)
		if retryErr != nil {
			return aiMicroScenario{}, retryErr
		}
		revised, parseErr := parseAIMicroScenario(retryRaw)
		if parseErr != nil {
			return aiMicroScenario{}, parseErr
		}
		return *revised, nil
	})
	if len(evidenceErrs) > 0 {
		log.Printf("[LLM] Evidence contract not satisfied after retry; using deterministic fallback: %v", evidenceErrs)
	}

	// Validate candidate scenario against schema & referential contracts
	if valErr := ValidateScenarioCandidate(scenario); valErr != nil {
		return nil, fmt.Errorf("scenario_validation_failed: %w", valErr)
	}

	return scenario, nil
}

// callChatCompletion performs one OpenAI-compatible chat completion with network
// retries and returns the raw message content.
func callChatCompletion(ctx context.Context, url string, apiKey string, model string, messages []chatMessage, temperature float64) (string, error) {
	reqBody := chatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temperature,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal LLM request: %w", err)
	}

	client := &http.Client{Timeout: 120 * time.Second}

	var resp *http.Response
	maxRetries := 2
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Rebuild request for each attempt (body is consumed after first read)
		retryReq, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBytes))
		if reqErr != nil {
			return "", fmt.Errorf("failed to create HTTP request (attempt %d): %w", attempt, reqErr)
		}
		retryReq.Header.Set("Content-Type", "application/json")
		retryReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

		resp, err = client.Do(retryReq)
		if err == nil {
			break
		}
		log.Printf("[LLM] Attempt %d/%d failed: %v", attempt, maxRetries, err)
		if ctx.Err() != nil {
			break
		}
		if attempt < maxRetries {
			select {
			case <-ctx.Done():
				return "", fmt.Errorf("LLM request failed after %d attempt(s): %w", attempt, ctx.Err())
			case <-time.After(3 * time.Second):
			}
		}
	}
	if err != nil {
		return "", fmt.Errorf("LLM request failed after %d attempts: %w", maxRetries, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("[LLM Error Details] URL: %s, Status: %d, Response: %s", url, resp.StatusCode, string(bodyBytes))
		return "", fmt.Errorf("LLM API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode LLM response: %w", err)
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("LLM returned empty choices")
	}

	return stripJSONFences(strings.TrimSpace(chatResp.Choices[0].Message.Content)), nil
}

func stripJSONFences(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "```json") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
		return strings.TrimSpace(strings.TrimSuffix(trimmed, "```"))
	}
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
		return strings.TrimSpace(strings.TrimSuffix(trimmed, "```"))
	}
	return trimmed
}

func parseAIMicroScenario(rawJSON string) (*aiMicroScenario, error) {
	var micro aiMicroScenario
	if err := json.Unmarshal([]byte(strings.TrimSpace(rawJSON)), &micro); err != nil {
		return nil, fmt.Errorf("invalid_json: failed to parse AI micro scenario JSON: %w", err)
	}
	return &micro, nil
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

	// Validate Logs length and anti-repetition rules
	initialLogs := normalizePhaseStepSlice(s.InitialLogs)
	logCount := len(initialLogs)
	if logCount < 10 {
		return fmt.Errorf("content_validation_failed: insufficient logs count (%d)", logCount)
	}
	if err := validateScenarioAntiRepetition(initialLogs); err != nil {
		return err
	}

	// Validate Expected Outcomes and the evidence contract
	if _, ok := s.ExpectedOutcomes.(map[string]interface{}); !ok {
		return fmt.Errorf("schema_validation_failed: expected_outcomes must be an object")
	}
	if err := validateScenarioEvidenceContract(s); err != nil {
		return err
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

// synthesizeFullScenario constructs the complete SIEM environment and returns
// the deterministic single-finding fallback when the generated findings fail the
// evidence contract.
func synthesizeFullScenario(ai aiMicroScenario, weakestDomain string, domainName string) *models.Scenario {
	scenario, _ := synthesizeFullScenarioWithErrors(ai, weakestDomain, domainName)
	return scenario
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
			Title:           "Unauthorized Security Policy Override Attempt",
			Description:     "Domain controller audit telemetry detected an unauthorized Group Policy modification disabling account lockout enforcement via service account SYSTEM_SVC.",
			SubTopicID:      "1.1",
			Archetype:       "iam_policy_tampering",
			BriefingStyle:   "siem_escalation_ticket",
			IsTruePositive:  true,
			TpFpExplanation: "This is a True Positive because unapproved privilege elevation and Group Policy override calls were executed directly on DC01.corp.internal from 192.168.1.150.",
			AttackerIP:      "192.168.1.150",
			TargetHost:      "DC01.corp.internal",
			PrimaryUser:     "SYSTEM_SVC",
			IncidentLogs: []aiIncidentLog{
				{Source: "ad_controller", Severity: "info", Host: "DC01.corp.internal", SrcIP: "192.168.1.150", User: "SYSTEM_SVC", Message: "Kerberos TGT granted for service account SYSTEM_SVC from workstation 192.168.1.150", RelativeOffsetSec: 0},
				{Source: "windows_event", Severity: "warning", Host: "DC01.corp.internal", SrcIP: "192.168.1.150", User: "SYSTEM_SVC", Message: "Event 5136: Directory service object Default Domain Policy modified outside change window", RelativeOffsetSec: 55},
				{Source: "ad_controller", Severity: "critical", Host: "DC01.corp.internal", SrcIP: "192.168.1.150", User: "SYSTEM_SVC", Message: "CRITICAL AUDIT: Security policy 'DisableAccountLockout' applied by unauthorized user 'SYSTEM_SVC'", RelativeOffsetSec: 120},
				{Source: "edr", Severity: "high", Host: "DC01.corp.internal", SrcIP: "192.168.1.150", User: "SYSTEM_SVC", Message: "Remote gpupdate /force invoked over SMB named pipe from 192.168.1.150 targeting DC01.corp.internal", RelativeOffsetSec: 195},
			},
			Findings: []aiFinding{
				{
					Description: "Unauthorized account lockout policy alteration detected on Primary Domain Controller DC01.corp.internal.",
					DomainID:    "domain1",
					Points:      10,
					EvidenceLogMessages: []string{
						"Event 5136: Directory service object Default Domain Policy modified outside change window",
						"CRITICAL AUDIT: Security policy 'DisableAccountLockout' applied by unauthorized user 'SYSTEM_SVC'",
						"Remote gpupdate /force invoked over SMB named pipe from 192.168.1.150 targeting DC01.corp.internal",
					},
					VerdictSupport: "malicious",
					Explanation:    "Directory service modification Event 5136 and DisableAccountLockout audit event confirm unauthorized policy override from 192.168.1.150.",
				},
			},
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
			Title:           "Suspicious Command Line Execution via Spearphishing Payload",
			Description:     "An endpoint EDR sensor flagged encoded PowerShell execution spawned by Microsoft Word after a finance employee opened an attached invoice document.",
			SubTopicID:      "2.2",
			Archetype:       "spearphish_macro_lolbin",
			BriefingStyle:   "helpdesk_user_report",
			IsTruePositive:  true,
			TpFpExplanation: "This is a True Positive because encoded PowerShell commands were executed by WINWORD.EXE and initiated an external C2 callback to 185.220.101.99.",
			AttackerIP:      "185.220.101.99",
			TargetHost:      "WS-FINANCE04",
			PrimaryUser:     "efoster",
			IncidentLogs: []aiIncidentLog{
				{Source: "email_gateway", Severity: "warning", Host: "MAIL-GW-01", SrcIP: "185.220.101.99", User: "efoster", Message: "Attachment Q3_Wire_Adjustment.docm delivered to efoster@corp.local from unverified sender", RelativeOffsetSec: 0},
				{Source: "edr", Severity: "high", Host: "WS-FINANCE04", User: "efoster", Message: "EDR ALERT: powershell.exe -e aW52b2tlLWV4cHJlc3Npb24... executed by WINWORD.EXE", RelativeOffsetSec: 65},
				{Source: "sysmon", Severity: "high", Host: "WS-FINANCE04", User: "efoster", Message: "Registry Run key HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run\\Updater created by powershell.exe", RelativeOffsetSec: 130},
				{Source: "firewall", Severity: "critical", Host: "WS-FINANCE04", SrcIP: "10.20.4.62", DstIP: "185.220.101.99", User: "efoster", Message: "Outbound HTTPS session established from WS-FINANCE04 to threat-intel flagged host 185.220.101.99:443", RelativeOffsetSec: 210},
			},
			Findings: []aiFinding{
				{
					Description: "Encoded PowerShell payload executed via MS Office child process on WS-FINANCE04.",
					DomainID:    "domain2",
					Points:      10,
					EvidenceLogMessages: []string{
						"EDR ALERT: powershell.exe -e aW52b2tlLWV4cHJlc3Npb24... executed by WINWORD.EXE",
						"Registry Run key HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run\\Updater created by powershell.exe",
						"Outbound HTTPS session established from WS-FINANCE04 to threat-intel flagged host 185.220.101.99:443",
					},
					VerdictSupport: "malicious",
					Explanation:    "WINWORD.EXE spawning encoded PowerShell followed by Run-key persistence and outbound C2 confirms active spearphishing compromise.",
				},
			},
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
			Title:           "Cloud Perimeter Firewall Misconfiguration",
			Description:     "Cloud posture monitoring and VPC flow telemetry detected public SSH ingress exposure followed by external login activity on AWS-PROD-DB01.",
			SubTopicID:      "3.2",
			Archetype:       "cloud_security_group_exposure",
			BriefingStyle:   "cloud_guardrail_alert",
			IsTruePositive:  true,
			TpFpExplanation: "This is a True Positive because SSH port 22 was opened to 0.0.0.0/0 on AWS-PROD-DB01 and immediately accessed from 203.0.113.88.",
			AttackerIP:      "203.0.113.88",
			TargetHost:      "AWS-PROD-DB01",
			PrimaryUser:     "ec2-user",
			IncidentLogs: []aiIncidentLog{
				{Source: "aws_cloudtrail", Severity: "warning", Host: "AWS-PROD-DB01", SrcIP: "10.0.3.8", User: "devops_temp", Message: "AuthorizeSecurityGroupIngress sg-0a91f4 opened TCP port 22 to 0.0.0.0/0 on AWS-PROD-DB01", RelativeOffsetSec: 0},
				{Source: "firewall", Severity: "high", Host: "AWS-PROD-DB01", SrcIP: "203.0.113.88", DstIP: "10.0.3.12", Message: "FIREWALL: Inbound SSH connection accepted from 203.0.113.88:49210 to 10.0.3.12:22", RelativeOffsetSec: 70},
				{Source: "edr", Severity: "high", Host: "AWS-PROD-DB01", SrcIP: "203.0.113.88", User: "ec2-user", Message: "Accepted publickey for ec2-user from 203.0.113.88 port 49210 on AWS-PROD-DB01", RelativeOffsetSec: 145},
				{Source: "edr", Severity: "critical", Host: "AWS-PROD-DB01", User: "ec2-user", Message: "Unauthorized cron persistence /etc/cron.d/db_sync_health written on AWS-PROD-DB01", RelativeOffsetSec: 225},
			},
			Findings: []aiFinding{
				{
					Description: "Management SSH interface exposed publicly due to misconfigured AWS Security Group rule.",
					DomainID:    "domain3",
					Points:      10,
					EvidenceLogMessages: []string{
						"AuthorizeSecurityGroupIngress sg-0a91f4 opened TCP port 22 to 0.0.0.0/0 on AWS-PROD-DB01",
						"FIREWALL: Inbound SSH connection accepted from 203.0.113.88:49210 to 10.0.3.12:22",
						"Unauthorized cron persistence /etc/cron.d/db_sync_health written on AWS-PROD-DB01",
					},
					VerdictSupport: "malicious",
					Explanation:    "CloudTrail shows 0.0.0.0/0 port 22 exposure followed by external SSH ingress from 203.0.113.88 and cron persistence.",
				},
			},
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
			Title:           "Suspicious Outbound C2 Beaconing Activity",
			Description:     "Threat hunting telemetry identified periodic outbound HTTP POST beacons from developer workstation WS-DEV09 to an uncategorized external host.",
			SubTopicID:      "4.4",
			Archetype:       "c2_http_beaconing",
			BriefingStyle:   "threat_hunting_lead",
			IsTruePositive:  true,
			TpFpExplanation: "This is a True Positive because a rogue Windows service on WS-DEV09 is beaconing every 60 seconds to threat-intel C2 IP 198.51.100.42.",
			AttackerIP:      "198.51.100.42",
			TargetHost:      "WS-DEV09",
			PrimaryUser:     "dev_mholt",
			IncidentLogs: []aiIncidentLog{
				{Source: "edr", Severity: "warning", Host: "WS-DEV09", User: "dev_mholt", Message: "Unsigned service binary C:\\ProgramData\\SysCache\\svcupd.exe registered as service 'SysCacheSvc'", RelativeOffsetSec: 0},
				{Source: "proxy_gateway", Severity: "high", Host: "WS-DEV09", SrcIP: "10.20.8.19", DstIP: "198.51.100.42", User: "dev_mholt", Message: "Periodic HTTP POST /api/v1/checkin to 198.51.100.42:8080 every 60s from WS-DEV09", RelativeOffsetSec: 80},
				{Source: "nids", Severity: "critical", Host: "WS-DEV09", SrcIP: "10.20.8.19", DstIP: "198.51.100.42", Message: "Cobalt Strike HTTP beacon jitter signature matched toward external IP 198.51.100.42:8080", RelativeOffsetSec: 170},
				{Source: "edr", Severity: "high", Host: "WS-DEV09", User: "dev_mholt", Message: "Process svcupd.exe accessed LSASS memory space on WS-DEV09 prior to outbound transfer", RelativeOffsetSec: 250},
			},
			Findings: []aiFinding{
				{
					Description: "Periodic outbound network traffic matching Cobalt Strike C2 beacon signature on WS-DEV09.",
					DomainID:    "domain4",
					Points:      10,
					EvidenceLogMessages: []string{
						"Unsigned service binary C:\\ProgramData\\SysCache\\svcupd.exe registered as service 'SysCacheSvc'",
						"Periodic HTTP POST /api/v1/checkin to 198.51.100.42:8080 every 60s from WS-DEV09",
						"Cobalt Strike HTTP beacon jitter signature matched toward external IP 198.51.100.42:8080",
					},
					VerdictSupport: "malicious",
					Explanation:    "Unsigned service persistence combined with 60s beacon intervals and Cobalt Strike NIDS signature confirms C2 compromise.",
				},
			},
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
			Title:           "Compliance Breach: Unencrypted Data Transfer",
			Description:     "DLP inspection flagged sensitive PCI-DSS cardholder records leaving payment server SRV-PAYMENT01 over cleartext HTTP toward partner relay 10.0.2.77.",
			SubTopicID:      "5.4",
			Archetype:       "dlp_cleartext_pci_exposure",
			BriefingStyle:   "shift_handover_brief",
			IsTruePositive:  true,
			TpFpExplanation: "This is a True Positive because cleartext PAN credit card numbers were transmitted over unencrypted HTTP in violation of PCI-DSS encryption requirements.",
			AttackerIP:      "10.0.2.77",
			TargetHost:      "SRV-PAYMENT01",
			PrimaryUser:     "svc_checkout",
			IncidentLogs: []aiIncidentLog{
				{Source: "nginx", Severity: "warning", Host: "SRV-PAYMENT01", SrcIP: "10.0.2.19", DstIP: "10.0.2.77", User: "svc_checkout", Message: "HTTP/1.1 cleartext listener bound on port 8081 of SRV-PAYMENT01 after deployment config drift", RelativeOffsetSec: 0},
				{Source: "dlp_sensor", Severity: "critical", Host: "SRV-PAYMENT01", SrcIP: "10.0.2.19", DstIP: "10.0.2.77", User: "svc_checkout", Message: "DLP ALERT: Cleartext credit card pattern matched in HTTP POST body payload to endpoint 10.0.2.77", RelativeOffsetSec: 90},
				{Source: "proxy_gateway", Severity: "high", Host: "SRV-PAYMENT01", SrcIP: "10.0.2.19", DstIP: "10.0.2.77", User: "svc_checkout", Message: "Unencrypted POST /v1/settlement/batch transferred 184 KB to 10.0.2.77:8081 without TLS", RelativeOffsetSec: 165},
				{Source: "compliance_scanner", Severity: "high", Host: "SRV-PAYMENT01", User: "svc_checkout", Message: "PCI-DSS Req 4.2 policy violation: plaintext PAN records transmitted outside encrypted tunnel", RelativeOffsetSec: 230},
			},
			Findings: []aiFinding{
				{
					Description: "PCI-DSS compliance breach caused by cleartext HTTP transmission of customer payment data.",
					DomainID:    "domain5",
					Points:      10,
					EvidenceLogMessages: []string{
						"DLP ALERT: Cleartext credit card pattern matched in HTTP POST body payload to endpoint 10.0.2.77",
						"Unencrypted POST /v1/settlement/batch transferred 184 KB to 10.0.2.77:8081 without TLS",
						"PCI-DSS Req 4.2 policy violation: plaintext PAN records transmitted outside encrypted tunnel",
					},
					VerdictSupport: "malicious",
					Explanation:    "Cleartext HTTP settlement transfers containing cardholder PAN data violate PCI-DSS cryptographic transmission controls.",
				},
			},
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

// ExtractPrimaryScenarioEntities extracts the primary TargetHost and AttackerIP from
// a Scenario's description, evidence findings, or high-severity incident logs (never noise row 0).
func ExtractPrimaryScenarioEntities(s *models.Scenario) (string, string) {
	if s == nil {
		return "SRV-WEB01", "185.220.101.45"
	}
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

	logs := normalizePhaseStepSlice(s.InitialLogs)
	descLower := strings.ToLower(s.Description)

	// Collect authored evidence indices first so incident entities take priority over noise.
	evidenceSet := make(map[int]bool)
	if outcomes, ok := s.ExpectedOutcomes.(map[string]interface{}); ok {
		for _, f := range mapSlice(outcomes["key_findings"]) {
			indices, _ := strictEvidenceIndices(f["evidence_log_indices"], len(logs))
			for _, idx := range indices {
				evidenceSet[idx] = true
			}
		}
	}

	inspectEntry := func(entry map[string]interface{}) {
		if h, ok := entry["host"].(string); ok && host == "" && strings.TrimSpace(h) != "" {
			host = strings.TrimSpace(h)
		}
		dst, _ := entry["dst_ip"].(string)
		src, _ := entry["src_ip"].(string)
		if ip == "" {
			if strings.TrimSpace(dst) != "" && !strings.HasPrefix(dst, "10.") && dst != "127.0.0.1" {
				ip = strings.TrimSpace(dst)
			} else if strings.TrimSpace(src) != "" && !strings.HasPrefix(src, "10.") && src != "127.0.0.1" {
				ip = strings.TrimSpace(src)
			} else if strings.TrimSpace(dst) != "" && dst != "127.0.0.1" {
				ip = strings.TrimSpace(dst)
			} else if strings.TrimSpace(src) != "" && src != "127.0.0.1" {
				ip = strings.TrimSpace(src)
			}
		}
	}

	// Pass 1: entities explicitly mentioned in the narrative description.
	for _, entry := range logs {
		if h, ok := entry["host"].(string); ok && host == "" && len(strings.TrimSpace(h)) > 2 && strings.Contains(descLower, strings.ToLower(strings.TrimSpace(h))) {
			host = strings.TrimSpace(h)
		}
		for _, field := range []string{"src_ip", "dst_ip"} {
			if candidate, ok := entry[field].(string); ok && ip == "" && strings.TrimSpace(candidate) != "" && strings.Contains(descLower, strings.TrimSpace(candidate)) {
				ip = strings.TrimSpace(candidate)
			}
		}
	}

	// Pass 2: high/critical evidence logs, then any evidence logs.
	for _, wantHigh := range []bool{true, false} {
		for idx, entry := range logs {
			if !evidenceSet[idx] {
				continue
			}
			if wantHigh && scenarioLogSeverityBucket(entry) != "high" {
				continue
			}
			inspectEntry(entry)
			if host != "" && ip != "" {
				return host, ip
			}
		}
	}

	// Pass 3: high/critical incident logs anywhere in timeline.
	for _, entry := range logs {
		if scenarioLogSeverityBucket(entry) != "high" {
			continue
		}
		inspectEntry(entry)
		if host != "" && ip != "" {
			break
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

	host, ip := ExtractPrimaryScenarioEntities(s)
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
		if strings.EqualFold(strings.TrimSpace(sc.Status), "archived") {
			continue
		}
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
