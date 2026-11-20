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
	Label         string `json:"label"`
	LabelTh       string `json:"label_th,omitempty"`
	IsCorrect     bool   `json:"is_correct"`
	Explanation   string `json:"explanation"`
	ExplanationTh string `json:"explanation_th,omitempty"`
}

type aiFinding struct {
	Description         string   `json:"description"`
	DescriptionTh       string   `json:"description_th,omitempty"`
	DomainID            string   `json:"domain_id"`
	Points              float64  `json:"points"`
	EvidenceLogMessages []string `json:"evidence_log_messages"`
	VerdictSupport      string   `json:"verdict_support"` // malicious | benign | supporting
	Explanation         string   `json:"explanation"`
	ExplanationTh       string   `json:"explanation_th,omitempty"`
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
	Title               string                   `json:"title"`
	TitleTh             string                   `json:"title_th,omitempty"`
	Description         string                   `json:"description"`
	DescriptionTh       string                   `json:"description_th,omitempty"`
	SubTopicID          string                   `json:"sub_topic_id,omitempty"`
	Archetype           string                   `json:"archetype,omitempty"`
	BriefingStyle       string                   `json:"briefing_style,omitempty"`
	IsTruePositive      bool                     `json:"is_true_positive"`
	TpFpExplanation     string                   `json:"tp_fp_explanation"`
	TpFpExplanationTh   string                   `json:"tp_fp_explanation_th,omitempty"`
	AttackerIP          string                   `json:"attacker_ip"`
	TargetHost          string                   `json:"target_host"`
	PrimaryUser         string                   `json:"primary_user,omitempty"`
	IncidentLogs        []aiIncidentLog          `json:"incident_logs,omitempty"`
	AttackLogMessage    string                   `json:"attack_log_message,omitempty"`
	KeyFinding          string                   `json:"key_finding,omitempty"`
	KeyFindingTh        string                   `json:"key_finding_th,omitempty"`
	Findings            []aiFinding              `json:"findings,omitempty"`
	ContainmentAction   string                   `json:"containment_action,omitempty"`
	ContainmentActionTh string                   `json:"containment_action_th,omitempty"`
	EradicationAction   string                   `json:"eradication_action,omitempty"`
	EradicationActionTh string                   `json:"eradication_action_th,omitempty"`
	RecoveryAction      string                   `json:"recovery_action,omitempty"`
	RecoveryActionTh    string                   `json:"recovery_action_th,omitempty"`
	ContainmentSteps    []aiResponseActionOption `json:"containment_steps,omitempty"`
	EradicationSteps    []aiResponseActionOption `json:"eradication_steps,omitempty"`
	RecoverySteps       []aiResponseActionOption `json:"recovery_steps,omitempty"`
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

	systemPrompt := "You are a specialized CompTIA Security+ 701 bilingual (English + Thai) cybersecurity scenario generator. Output ONLY raw JSON. Every scenario MUST author a realistic 4-to-8 event English-only incident_logs chain matching its specific domain and environment (never generic web/CORR-9901 boilerplate, zero Thai characters in incident_logs), follow the assigned SY0-701 sub-topic, archetype, briefing style, and target verdict, provide natural Thai translations (*_th fields) preserving technical security terms, hosts, IPs, usernames, log source names, and MITRE IDs verbatim in English, and every incident response phase (containment_steps, eradication_steps, recovery_steps) MUST contain 3 realistic SOC options (1 correct option and 2 plausible technical distractors) with matching sentence length (15-30 words) and specific entity references."
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

Rules for Bilingual Thai Fields (*_th):
1. Provide natural Thai text for title_th, description_th, tp_fp_explanation_th, every findings[].description_th & explanation_th, and every phase step's label_th & explanation_th.
2. Follow Hybrid Bilingual Terminology: use natural Thai prose while keeping hostnames, IPs, usernames, log source names, file paths, command-line strings, and standard security terms/MITRE IDs in English.
3. incident_logs and all IOCs (host, src_ip, dst_ip, user, source, message) MUST remain 100%% English with zero Thai characters.
4. Match the assigned briefing_style tone in both description and description_th.

Output ONLY valid JSON matching this exact structure:
{
  "title": "Short Scenario Title",
  "title_th": "ชื่อสถานการณ์จำลองภาษาไทย (คงชื่อเทคนิค/ระบบเป็นภาษาอังกฤษ)",
  "description": "Brief 2-3 sentence briefing following the assigned briefing style and naming the primary host/actor naturally",
  "description_th": "คำอธิบายสถานการณ์ 2-3 ประโยคเป็นภาษาไทยโดยยังคงชื่อโฮสต์และ IP เป็นภาษาอังกฤษ",
  "sub_topic_id": "%s",
  "archetype": "%s",
  "briefing_style": "%s",
  "is_true_positive": %t,
  "tp_fp_explanation": "Detailed explanation why this is a True or False Positive incident",
  "tp_fp_explanation_th": "คำอธิบายเหตุผลอย่างละเอียดว่าเป็น True Positive หรือ False Positive",
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
      "description_th": "จุดสังเกตหลักที่สอดคล้องกับหลักฐานใน log",
      "domain_id": "domain4",
      "points": 10,
      "evidence_log_messages": ["Primary suspicious or decisive execution event on SRV-WEB01"],
      "verdict_support": "malicious",
      "explanation": "Why this evidence proves the finding",
      "explanation_th": "เหตุผลที่หลักฐานนี้ยืนยันข้อสรุปของจุดสังเกต"
    }
  ],
  "containment_steps": [
    {"label": "Detailed correct containment procedure referencing SRV-WEB01 and 185.220.101.45...", "label_th": "ขั้นตอน Containment ที่ถูกต้องเป็นภาษาไทยโดยระบุ SRV-WEB01 และ 185.220.101.45...", "is_correct": true, "explanation": "Why this containment action is effective", "explanation_th": "เหตุผลที่ขั้นตอน Containment นี้มีประสิทธิภาพ"},
    {"label": "Plausible but flawed containment procedure referencing SRV-WEB01 and 185.220.101.45...", "label_th": "ตัวเลือกหลอกของ Containment ที่ดูสมจริงแต่มีข้อผิดพลาด...", "is_correct": false, "explanation": "Why this distractor is flawed", "explanation_th": "เหตุผลที่ตัวเลือกหลอกนี้ไม่เหมาะสม"},
    {"label": "Second plausible but flawed containment procedure referencing SRV-WEB01 and 185.220.101.45...", "label_th": "ตัวเลือกหลอกที่สองของ Containment ที่ดูสมจริงแต่มีข้อผิดพลาด...", "is_correct": false, "explanation": "Why this distractor is flawed", "explanation_th": "เหตุผลที่ตัวเลือกหลอกนี้ไม่เหมาะสม"}
  ],
  "eradication_steps": [
    {"label": "Detailed correct eradication procedure referencing SRV-WEB01...", "label_th": "ขั้นตอน Eradication ที่ถูกต้องเป็นภาษาไทยโดยระบุ SRV-WEB01...", "is_correct": true, "explanation": "Why this eradication action is effective", "explanation_th": "เหตุผลที่ขั้นตอน Eradication นี้มีประสิทธิภาพ"},
    {"label": "Plausible but flawed eradication procedure referencing SRV-WEB01...", "label_th": "ตัวเลือกหลอกของ Eradication ที่ดูสมจริงแต่มีข้อผิดพลาด...", "is_correct": false, "explanation": "Why this distractor is flawed", "explanation_th": "เหตุผลที่ตัวเลือกหลอกนี้ไม่เหมาะสม"},
    {"label": "Second plausible but flawed eradication procedure referencing SRV-WEB01...", "label_th": "ตัวเลือกหลอกที่สองของ Eradication ที่ดูสมจริงแต่มีข้อผิดพลาด...", "is_correct": false, "explanation": "Why this distractor is flawed", "explanation_th": "เหตุผลที่ตัวเลือกหลอกนี้ไม่เหมาะสม"}
  ],
  "recovery_steps": [
    {"label": "Detailed correct recovery procedure referencing SRV-WEB01...", "label_th": "ขั้นตอน Recovery ที่ถูกต้องเป็นภาษาไทยโดยระบุ SRV-WEB01...", "is_correct": true, "explanation": "Why this recovery action is effective", "explanation_th": "เหตุผลที่ขั้นตอน Recovery นี้มีประสิทธิภาพ"},
    {"label": "Plausible but flawed recovery procedure referencing SRV-WEB01...", "label_th": "ตัวเลือกหลอกของ Recovery ที่ดูสมจริงแต่มีข้อผิดพลาด...", "is_correct": false, "explanation": "Why this distractor is flawed", "explanation_th": "เหตุผลที่ตัวเลือกหลอกนี้ไม่เหมาะสม"},
    {"label": "Second plausible but flawed recovery procedure referencing SRV-WEB01...", "label_th": "ตัวเลือกหลอกที่สองของ Recovery ที่ดูสมจริงแต่มีข้อผิดพลาด...", "is_correct": false, "explanation": "Why this distractor is flawed", "explanation_th": "เหตุผลที่ตัวเลือกหลอกนี้ไม่เหมาะสม"}
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

// ValidateScenarioCandidate validates scenario structure, bounds, anti-predictability rules, referential integrity, and bilingual translations.
func ValidateScenarioCandidate(s *models.Scenario) error {
	if err := ValidateScenarioBaseCandidate(s); err != nil {
		return err
	}
	return ValidateScenarioTranslationsTh(s)
}

// ValidateScenarioBaseCandidate validates canonical English scenario structure, zero-Thai-in-logs invariant,
// bounds, anti-predictability rules, and evidence integrity.
func ValidateScenarioBaseCandidate(s *models.Scenario) error {
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

	// Validate Logs length, zero-Thai rule, and anti-repetition rules
	initialLogs := normalizePhaseStepSlice(s.InitialLogs)
	logCount := len(initialLogs)
	if logCount < 10 {
		return fmt.Errorf("content_validation_failed: insufficient logs count (%d)", logCount)
	}
	for idx, entry := range initialLogs {
		if containsThaiInValue(entry) {
			return fmt.Errorf("content_validation_failed: initial_logs entry %d must be strictly English-only (found Thai characters)", idx)
		}
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

// ValidateScenarioTranslationsTh checks that translations.th completely covers title, description,
// tp_fp_explanation, all playbook_steps action IDs, and all expected_outcomes findings IDs.
func ValidateScenarioTranslationsTh(s *models.Scenario) error {
	if s == nil {
		return fmt.Errorf("translation_validation_failed: scenario is nil")
	}
	translations := normalizeAnyStringMap(s.Translations)
	if len(translations) == 0 {
		return fmt.Errorf("translation_validation_failed: translations map is required")
	}
	thMap := normalizeAnyStringMap(translations["th"])
	if len(thMap) == 0 {
		return fmt.Errorf("translation_validation_failed: translations.th is required")
	}

	thTitle, _ := thMap["title"].(string)
	if strings.TrimSpace(thTitle) == "" || !ContainsThaiRune(thTitle) {
		return fmt.Errorf("translation_validation_failed: translations.th.title must be non-empty Thai text")
	}
	thDesc, _ := thMap["description"].(string)
	if strings.TrimSpace(thDesc) == "" || !ContainsThaiRune(thDesc) {
		return fmt.Errorf("translation_validation_failed: translations.th.description must be non-empty Thai text")
	}
	thTpFp, _ := thMap["tp_fp_explanation"].(string)
	if strings.TrimSpace(thTpFp) == "" || !ContainsThaiRune(thTpFp) {
		return fmt.Errorf("translation_validation_failed: translations.th.tp_fp_explanation must be non-empty Thai text")
	}

	thPlaybook := normalizeAnyNestedStringMap(thMap["playbook_steps"])
	playbook := normalizeAnyStringMap(s.PlaybookSteps)
	for _, phase := range []string{"containment", "eradication", "recovery"} {
		steps := normalizePhaseStepSlice(playbook[phase])
		for _, step := range steps {
			id, _ := step["id"].(string)
			id = strings.TrimSpace(id)
			if id == "" {
				return fmt.Errorf("translation_validation_failed: playbook step in phase %q is missing id", phase)
			}
			entry := thPlaybook[id]
			if strings.TrimSpace(entry["label"]) == "" || strings.TrimSpace(entry["explanation"]) == "" {
				return fmt.Errorf("translation_validation_failed: translations.th.playbook_steps[%q] must include non-empty label and explanation", id)
			}
		}
	}

	thFindings := normalizeAnyNestedStringMap(thMap["findings"])
	outcomes := normalizeAnyStringMap(s.ExpectedOutcomes)
	findings := normalizePhaseStepSlice(outcomes["key_findings"])
	for _, f := range findings {
		id, _ := f["id"].(string)
		id = strings.TrimSpace(id)
		if id == "" {
			return fmt.Errorf("translation_validation_failed: finding is missing id")
		}
		entry := thFindings[id]
		if strings.TrimSpace(entry["description"]) == "" || strings.TrimSpace(entry["explanation"]) == "" {
			return fmt.Errorf("translation_validation_failed: translations.th.findings[%q] must include non-empty description and explanation", id)
		}
	}

	return nil
}

func containsThaiInValue(v interface{}) bool {
	switch typed := v.(type) {
	case string:
		return ContainsThaiRune(typed)
	case []interface{}:
		for _, item := range typed {
			if containsThaiInValue(item) {
				return true
			}
		}
	case []map[string]interface{}:
		for _, item := range typed {
			if containsThaiInValue(item) {
				return true
			}
		}
	case map[string]interface{}:
		for k, val := range typed {
			if ContainsThaiRune(k) || containsThaiInValue(val) {
				return true
			}
		}
	case map[string]string:
		for k, val := range typed {
			if ContainsThaiRune(k) || ContainsThaiRune(val) {
				return true
			}
		}
	}
	return false
}

func normalizeAnyStringMap(raw interface{}) map[string]interface{} {
	switch typed := raw.(type) {
	case map[string]interface{}:
		return typed
	case map[string]string:
		out := make(map[string]interface{}, len(typed))
		for k, v := range typed {
			out[k] = v
		}
		return out
	}
	return nil
}

func normalizeAnyNestedStringMap(raw interface{}) map[string]map[string]string {
	out := make(map[string]map[string]string)
	switch typed := raw.(type) {
	case map[string]map[string]string:
		return typed
	case map[string]interface{}:
		for key, item := range typed {
			sub := make(map[string]string)
			switch subTyped := item.(type) {
			case map[string]string:
				for sk, sv := range subTyped {
					sub[sk] = sv
				}
			case map[string]interface{}:
				for sk, sv := range subTyped {
					if str, ok := sv.(string); ok {
						sub[sk] = str
					}
				}
			}
			if len(sub) > 0 {
				out[key] = sub
			}
		}
	}
	return out
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
	steps, _ := buildPhaseStepsWithDistractorsWithTh(phase, idPrefix, domainID, legacyCorrectAction, "", providedSteps, ai)
	return steps
}

func buildPhaseStepsWithDistractorsWithTh(phase string, idPrefix string, domainID string, legacyCorrectAction string, legacyCorrectActionTh string, providedSteps []aiResponseActionOption, ai aiMicroScenario) ([]map[string]interface{}, map[string]map[string]string) {
	steps := make([]map[string]interface{}, 0, 4)
	thSteps := make(map[string]map[string]string)
	seen := make(map[string]bool)
	correctCount := 0
	distractorCount := 0

	addOption := func(label string, labelTh string, isCorrect bool, explanation string, explanationTh string) {
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

		stepID := fmt.Sprintf("%s%d", idPrefix, len(steps)+1)
		steps = append(steps, map[string]interface{}{
			"id":          stepID,
			"label":       trimmed,
			"is_correct":  isCorrect,
			"domain_id":   domainID,
			"points":      pts,
			"explanation": explanation,
		})

		trimmedTh := strings.TrimSpace(labelTh)
		trimmedExpTh := strings.TrimSpace(explanationTh)
		if trimmedTh != "" || trimmedExpTh != "" {
			thSteps[stepID] = map[string]string{
				"label":       trimmedTh,
				"explanation": trimmedExpTh,
			}
		}
	}

	for _, opt := range providedSteps {
		addOption(opt.Label, opt.LabelTh, opt.IsCorrect, opt.Explanation, opt.ExplanationTh)
	}

	if correctCount == 0 && strings.TrimSpace(legacyCorrectAction) != "" {
		legacyExpTh := ""
		if strings.TrimSpace(legacyCorrectActionTh) != "" {
			phaseTh := map[string]string{
				"containment": "Containment (การจำกัดวงความเสียหาย)",
				"eradication": "Eradication (การกำจัดภัยคุกคาม)",
				"recovery":    "Recovery (การฟื้นฟูระบบ)",
			}[phase]
			if phaseTh == "" {
				phaseTh = phase
			}
			legacyExpTh = fmt.Sprintf("ดำเนินการตามขั้นตอน %s ที่ผ่านการตรวจสอบแล้วบน %s โดยตรง", phaseTh, ai.TargetHost)
		}
		addOption(
			legacyCorrectAction,
			legacyCorrectActionTh,
			true,
			fmt.Sprintf("Directly executes validated %s procedure on %s.", phase, ai.TargetHost),
			legacyExpTh,
		)
	}

	if correctCount == 0 {
		defaultCorrect := defaultContextualCorrectOption(phase, ai)
		addOption(defaultCorrect.Label, defaultCorrect.LabelTh, true, defaultCorrect.Explanation, defaultCorrect.ExplanationTh)
	}

	if distractorCount < 2 || len(steps) < 3 {
		candidates := buildContextualDistractors(phase, domainID, ai)
		for _, cand := range candidates {
			if distractorCount >= 2 && len(steps) >= 3 {
				break
			}
			addOption(cand.Label, cand.LabelTh, false, cand.Explanation, cand.ExplanationTh)
		}
	}

	return steps, thSteps
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
				Label:         fmt.Sprintf("Verify authorized change control records for traffic between %s and %s while keeping production services online without isolating the host.", host, ip),
				LabelTh:       fmt.Sprintf("ตรวจสอบบันทึกการขออนุมัติเปลี่ยนแปลงระบบ (Change Control) สำหรับทราฟฟิกระหว่าง %s และ %s โดยยังคงให้บริการระบบงานจริงตามปกติและไม่แยกโฮสต์ออกจากเครือข่าย", host, ip),
				IsCorrect:     true,
				Explanation:   "Confirms benign activity against authorized change records without causing unnecessary production downtime.",
				ExplanationTh: "ยืนยันว่ากิจกรรมดังกล่าวปลอดภัยและได้รับอนุญาตตามบันทึก Change Records โดยไม่ก่อให้เกิดการหยุดชะงักของระบบงานจริงโดยไม่จำเป็น",
			}
		case "eradication":
			return aiResponseActionOption{
				Label:         fmt.Sprintf("Document the verified benign telemetry pattern from %s on %s and refine SIEM correlation exclusions to reduce false positive alert fatigue.", ip, host),
				LabelTh:       fmt.Sprintf("จัดทำเอกสารบันทึกพฤติกรรม telemetry ที่ตรวจสอบแล้วว่าปลอดภัยจาก %s บน %s พร้อมปรับจูน SIEM correlation exclusions เพื่อลดการแจ้งเตือนเท็จซ้ำซ้อน", ip, host),
				IsCorrect:     true,
				Explanation:   "Improves detection fidelity without modifying legitimate system binaries or configurations.",
				ExplanationTh: "เพิ่มความแม่นยำในการตรวจจับโดยไม่แก้ไขไฟล์ระบบหรือการตั้งค่าที่ถูกต้องตามกฎหมาย",
			}
		default:
			return aiResponseActionOption{
				Label:         fmt.Sprintf("Close the incident ticket as a verified False Positive and confirm normal baseline monitoring remains active for %s and source %s.", host, ip),
				LabelTh:       fmt.Sprintf("ปิด incident ticket ในสถานะ False Positive ที่ตรวจสอบแล้ว พร้อมยืนยันว่าระบบเฝ้าระวังมาตรฐานสำหรับ %s และต้นทาง %s ยังคงทำงานปกติ", host, ip),
				IsCorrect:     true,
				Explanation:   "Completes the incident lifecycle cleanly while preserving standard monitoring coverage.",
				ExplanationTh: "ปิดวงจรการจัดการ incident ได้อย่างครบถ้วนโดยยังคงความครอบคลุมของการเฝ้าระวังตามปกติ",
			}
		}
	}

	switch phase {
	case "containment":
		return aiResponseActionOption{
			Label:         fmt.Sprintf("Isolate %s from the production network VLAN immediately and block all inbound and outbound sessions involving %s at the perimeter firewall.", host, ip),
			LabelTh:       fmt.Sprintf("แยกโฮสต์ %s ออกจาก VLAN ระบบงานจริงทันที (Network Isolation) และปิดกั้นการเชื่อมต่อทั้งขาเข้าและขาออกที่เกี่ยวข้องกับ %s ที่ perimeter firewall", host, ip),
			IsCorrect:     true,
			Explanation:   "Immediately halts active command-and-control communication and prevents lateral movement from the compromised host.",
			ExplanationTh: "หยุดยั้งการสื่อสาร command-and-control ที่กำลังทำงานและป้องกันการเคลื่อนที่ข้ามระบบ (lateral movement) จากโฮสต์ที่ถูกบุกรุกได้ทันที",
		}
	case "eradication":
		return aiResponseActionOption{
			Label:         fmt.Sprintf("Terminate unauthorized processes on %s, remove persistence artifacts linked to %s, and rotate all credentials exposed on the host.", host, ip),
			LabelTh:       fmt.Sprintf("ยุติโปรเซสที่ไม่ได้รับอนุญาตบน %s ลบกลไก persistence ที่เชื่อมโยงกับ %s และหมุนเวียน credentials ทั้งหมดที่อาจรั่วไหลบนโฮสต์ดังกล่าว", host, ip),
			IsCorrect:     true,
			Explanation:   "Eliminates attacker persistence mechanisms and invalidates potentially harvested credentials.",
			ExplanationTh: "กำจัดกลไกการฝังตัว (persistence) ของผู้โจมตีและยกเลิกอายุการใช้งาน credentials ที่อาจถูกขโมยออกไป",
		}
	default:
		return aiResponseActionOption{
			Label:         fmt.Sprintf("Restore %s from a verified clean image, validate service integrity, and enforce enhanced SIEM telemetry monitoring for traffic from %s.", host, ip),
			LabelTh:       fmt.Sprintf("กู้คืน %s จาก clean image ที่ตรวจสอบแล้ว ทดสอบความสมบูรณ์ของบริการ และเปิดการเฝ้าระวัง SIEM telemetry เชิงลึกสำหรับทราฟฟิกจาก %s", host, ip),
			IsCorrect:     true,
			Explanation:   "Safely returns the service to production with heightened visibility against re-infection.",
			ExplanationTh: "นำระบบกลับเข้าสู่การใช้งานจริงอย่างปลอดภัยพร้อมเพิ่มระดับการเฝ้าระวังเพื่อป้องกันการโจมตีซ้ำ",
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

	domainFocusTh := map[string]string{
		"domain1": "การควบคุมสิทธิ์และตัวตน (IAM)",
		"domain2": "การป้องกันการโจมตีบน endpoint",
		"domain3": "สถาปัตยกรรมโซนเครือข่าย",
		"domain4": "การเชื่อมโยงเหตุการณ์ SIEM telemetry",
		"domain5": "การกำกับดูแลและการตรวจสอบ compliance",
	}[domainID]
	if domainFocusTh == "" {
		domainFocusTh = "การปฏิบัติการด้านความมั่นคงปลอดภัย (SOC)"
	}

	if !ai.IsTruePositive {
		switch phase {
		case "containment":
			return []aiResponseActionOption{
				{
					Label:         fmt.Sprintf("Immediately quarantine %s at the EDR layer and blackhole all subnet routing toward %s before reviewing %s change logs.", host, ip, domainFocus),
					LabelTh:       fmt.Sprintf("กักกัน %s ในระดับ EDR ทันทีและตัดเส้นทาง routing ทั้ง subnet ไปยัง %s ก่อนตรวจสอบบันทึกการเปลี่ยนแปลงด้าน %s", host, ip, domainFocusTh),
					IsCorrect:     false,
					Explanation:   "Prematurely isolates a production system during a benign event without verifying authorization logs.",
					ExplanationTh: "ตัดแยกระบบงานจริงออกก่อนเวลาอันควรในเหตุการณ์ที่ปลอดภัยโดยไม่ตรวจสอบบันทึกการอนุมัติ",
				},
				{
					Label:         fmt.Sprintf("Revoke all active Kerberos tickets and disable service accounts operating on %s in response to the benign traffic from %s.", host, ip),
					LabelTh:       fmt.Sprintf("เพิกถอน Kerberos tickets ทั้งหมดและปิดการใช้งาน service accounts ที่ทำงานบน %s เพื่อตอบสนองต่อทราฟฟิกปกติจาก %s", host, ip),
					IsCorrect:     false,
					Explanation:   "Disrupts legitimate business authentication workflows for a verified False Positive alert.",
					ExplanationTh: "ทำให้ระบบพิสูจน์ตัวตนของการทำงานตามปกติหยุดชะงักสำหรับเหตุการณ์ False Positive ที่ตรวจสอบแล้ว",
				},
				{
					Label:         fmt.Sprintf("Configure an emergency perimeter firewall drop rule for the entire internal subnet hosting %s to block connections from %s.", host, ip),
					LabelTh:       fmt.Sprintf("ตั้งค่ากฎฉุกเฉินบน perimeter firewall เพื่อทิ้งแพ็กเก็ตทั้งหมดของ subnet ภายในที่มี %s เพื่อบล็อกการเชื่อมต่อจาก %s", host, ip),
					IsCorrect:     false,
					Explanation:   "Causes a broad network outage for benign authorized activity.",
					ExplanationTh: "ก่อให้เกิดเครือข่ายล่มเป็นวงกว้างทั้งที่กิจกรรมดังกล่าวได้รับอนุญาตและปลอดภัย",
				},
			}
		case "eradication":
			return []aiResponseActionOption{
				{
					Label:         fmt.Sprintf("Wipe the operating system partition on %s and revoke TLS certificates associated with %s despite confirmed benign %s verification.", host, ip, domainFocus),
					LabelTh:       fmt.Sprintf("ล้างพาร์ทิชันระบบปฏิบัติการบน %s และเพิกถอนใบรับรอง TLS ที่เกี่ยวข้องกับ %s ทั้งที่ผลการตรวจสอบ %s ยืนยันว่าปลอดภัย", host, ip, domainFocusTh),
					IsCorrect:     false,
					Explanation:   "Destructive eradication actions are inappropriate when telemetry confirms a False Positive.",
					ExplanationTh: "การลบหรือทำลายระบบไม่เหมาะสมอย่างยิ่งเมื่อข้อมูล telemetry ยืนยันว่าเป็น False Positive",
				},
				{
					Label:         fmt.Sprintf("Permanently disable the SIEM correlation rule and EDR telemetry agent on %s so future events from %s are ignored.", host, ip),
					LabelTh:       fmt.Sprintf("ปิดกฎ SIEM correlation และ EDR telemetry agent บน %s อย่างถาวรเพื่อให้ระบบละเลยเหตุการณ์จาก %s ในอนาคต", host, ip),
					IsCorrect:     false,
					Explanation:   "Disabling security sensors creates a critical visibility blind spot rather than properly tuning the rule.",
					ExplanationTh: "การปิดเซ็นเซอร์รักษาความปลอดภัยทำให้เกิดจุดบอดในการตรวจจับแทนที่จะปรับจูนกฎอย่างเหมาะสม",
				},
				{
					Label:         fmt.Sprintf("Force a global password reset across all domain users who accessed %s during the window of activity from %s.", host, ip),
					LabelTh:       fmt.Sprintf("บังคับรีเซ็ตรหัสผ่านผู้ใช้ทั้งหมดในโดเมนที่เข้าถึง %s ในช่วงเวลาที่มีทราฟฟิกจาก %s", host, ip),
					IsCorrect:     false,
					Explanation:   "Unnecessary enterprise-wide disruption for a benign operational event.",
					ExplanationTh: "สร้างความปั่นป่วนทั่วทั้งองค์กรโดยไม่จำเป็นสำหรับกิจกรรมปฏิบัติการตามปกติ",
				},
			}
		default:
			return []aiResponseActionOption{
				{
					Label:         fmt.Sprintf("Roll back %s to a week-old backup snapshot and purge all database transactions recorded during the connection from %s.", host, ip),
					LabelTh:       fmt.Sprintf("ย้อนสถานะ %s กลับไปยัง backup snapshot เก่าเมื่อสัปดาห์ที่แล้วและลบ transaction ฐานข้อมูลทั้งหมดในช่วงที่เชื่อมต่อกับ %s", host, ip),
					IsCorrect:     false,
					Explanation:   "Causes severe production data loss when no malicious tampering occurred.",
					ExplanationTh: "ก่อให้เกิดการสูญหายของข้อมูลการผลิตอย่างรุนแรงทั้งที่ไม่มีการดัดแปลงระบบที่เป็นอันตราย",
				},
				{
					Label:         fmt.Sprintf("Migrate %s to an unmonitored DMZ VLAN and bypass %s inspection for traffic originating from %s.", host, domainFocus, ip),
					LabelTh:       fmt.Sprintf("ย้าย %s ไปยัง DMZ VLAN ที่ไม่มีการเฝ้าระวังและข้ามการตรวจสอบ %s สำหรับทราฟฟิกที่มาจาก %s", host, domainFocusTh, ip),
					IsCorrect:     false,
					Explanation:   "Weakens network architecture security controls and removes inspection visibility.",
					ExplanationTh: "ลดระดับการควบคุมความมั่นคงปลอดภัยของสถาปัตยกรรมเครือข่ายและทำลายความสามารถในการตรวจสอบทราฟฟิก",
				},
				{
					Label:         fmt.Sprintf("Grant permanent local administrator exceptions on %s for all future connections matching %s without MFA verification.", host, ip),
					LabelTh:       fmt.Sprintf("อนุญาตสิทธิ์ local administrator แบบถาวรบน %s สำหรับการเชื่อมต่อในอนาคตที่ตรงกับ %s โดยไม่ต้องยืนยัน MFA", host, ip),
					IsCorrect:     false,
					Explanation:   "Violates least-privilege principles and introduces unnecessary access risk.",
					ExplanationTh: "ละเมิดหลักการ least privilege และเพิ่มความเสี่ยงด้านการเข้าถึงระบบโดยไม่จำเป็น",
				},
			}
		}
	}

	switch phase {
	case "containment":
		return []aiResponseActionOption{
			{
				Label:         fmt.Sprintf("Perform an immediate hard power-off of %s and flush volatile RAM caches to terminate active sockets connected to %s.", host, ip),
				LabelTh:       fmt.Sprintf("ปิดเครื่อง %s แบบฉับพลัน (Hard Power-Off) และล้างข้อมูล RAM ชั่วคราวทันทีเพื่อตัดการเชื่อมต่อ socket กับ %s", host, ip),
				IsCorrect:     false,
				Explanation:   "Destroys volatile memory artifacts and active network state required for forensic root-cause investigation.",
				ExplanationTh: "ทำลายหลักฐานในหน่วยความจำชั่วคราว (volatile RAM) และสถานะเครือข่ายที่จำเป็นต่อการสืบสวนหาสาเหตุรากเหง้า",
			},
			{
				Label:         fmt.Sprintf("Reroute suspicious traffic from %s targeting %s into the internal corporate backup VLAN while bypassing %s controls.", ip, host, domainFocus),
				LabelTh:       fmt.Sprintf("เปลี่ยนเส้นทางทราฟฟิกต้องสงสัยจาก %s ที่มุ่งเป้าไปยัง %s เข้าสู่ corporate backup VLAN โดยข้ามการควบคุมด้าน %s", ip, host, domainFocusTh),
				IsCorrect:     false,
				Explanation:   "Exposes the internal backup network segment to lateral movement instead of containing the threat.",
				ExplanationTh: "เปิดโอกาสให้ภัยคุกคามแพร่กระจายเข้าสู่ส่วนเครือข่ายสำรองข้อมูลภายในแทนที่จะควบคุมขอบเขตความเสียหาย",
			},
			{
				Label:         fmt.Sprintf("Block the internal default gateway and core DNS resolver for the entire subnet of %s rather than filtering attacker endpoint %s.", host, ip),
				LabelTh:       fmt.Sprintf("บล็อก internal default gateway และ core DNS resolver สำหรับทั้ง subnet ของ %s แทนการปิดกั้นเฉพาะปลายทางของผู้โจมตี %s", host, ip),
				IsCorrect:     false,
				Explanation:   "Causes collateral outage across healthy production hosts without isolating the compromised endpoint.",
				ExplanationTh: "ก่อให้เกิดระบบล่มในวงกว้างกับโฮสต์ที่ปกติทั้งหมดโดยไม่ได้แยกกักกันเครื่องที่ถูกบุกรุก",
			},
		}
	case "eradication":
		return []aiResponseActionOption{
			{
				Label:         fmt.Sprintf("Clear all local security event logs and %s history on %s before restarting the service connected to %s.", domainFocus, host, ip),
				LabelTh:       fmt.Sprintf("ลบ security event logs ในเครื่องและประวัติ %s ทั้งหมดบน %s ก่อนรีสตาร์ทบริการที่เชื่อมต่อกับ %s", domainFocusTh, host, ip),
				IsCorrect:     false,
				Explanation:   "Erases critical audit trails without removing the underlying vulnerability or persistence mechanism.",
				ExplanationTh: "ทำลายร่องรอยการตรวจสอบที่สำคัญโดยไม่ได้กำจัดช่องโหว่หรือกลไก persistence ที่แฝงอยู่",
			},
			{
				Label:         fmt.Sprintf("Restart the affected application daemon on %s and whitelist %s in the host firewall to suppress recurring IDS alerts.", host, ip),
				LabelTh:       fmt.Sprintf("รีสตาร์ท application daemon ที่ได้รับผลกระทบบน %s และเพิ่ม %s เข้า whitelist ใน host firewall เพื่อลดการแจ้งเตือน IDS ซ้ำ", host, ip),
				IsCorrect:     false,
				Explanation:   "Leaves persistent backdoors intact on disk and explicitly permits further attacker access.",
				ExplanationTh: "ปล่อยให้ backdoor ยังคงอยู่บนดิสก์และยังเปิดทางให้ผู้โจมตีเชื่อมต่อเข้ามาได้ซ้ำ",
			},
			{
				Label:         fmt.Sprintf("Delete only temporary browser cache files on %s while leaving scheduled tasks and service accounts touched by %s unmodified.", host, ip),
				LabelTh:       fmt.Sprintf("ลบเฉพาะไฟล์แคชชั่วคราวของเบราว์เซอร์บน %s โดยปล่อยให้ scheduled tasks และ service accounts ที่ถูกแก้ไขโดย %s คงเดิม", host, ip),
				IsCorrect:     false,
				Explanation:   "Fails to remediate privilege escalation and persistence mechanisms established during the intrusion.",
				ExplanationTh: "ไม่สามารถกำจัดกลไกการยกระดับสิทธิ์และการฝังตัวที่ถูกสร้างขึ้นระหว่างการบุกรุกได้",
			},
		}
	default:
		return []aiResponseActionOption{
			{
				Label:         fmt.Sprintf("Restore %s directly from the most recent unverified snapshot taken after %s initiated exploitation without applying %s patches.", host, ip, domainFocus),
				LabelTh:       fmt.Sprintf("กู้คืน %s โดยตรงจาก snapshot ล่าสุดที่ยังไม่ได้ตรวจสอบซึ่งสร้างหลัง %s เริ่มการโจมตีโดยไม่อัปเดตแพตช์ด้าน %s", host, ip, domainFocusTh),
				IsCorrect:     false,
				Explanation:   "Reintroduces the unpatched vulnerability and potentially restores dormant attacker persistence.",
				ExplanationTh: "นำช่องโหว่ที่ยังไม่ได้แก้ไขกลับมาใช้อีกครั้งและอาจนำกลไก persistence ที่ซ่อนอยู่กลับมาด้วย",
			},
			{
				Label:         fmt.Sprintf("Reconnect %s to the production VLAN immediately and suspend EDR behavioral monitoring for 48 hours to avoid performance overhead after %s.", host, ip),
				LabelTh:       fmt.Sprintf("เชื่อมต่อ %s กลับเข้า production VLAN ทันทีและระงับการเฝ้าระวังพฤติกรรมของ EDR เป็นเวลา 48 ชั่วโมงเพื่อลดภาระเครื่องหลังเหตุการณ์ %s", host, ip),
				IsCorrect:     false,
				Explanation:   "Removes endpoint visibility during the critical post-incident observation window.",
				ExplanationTh: "ปิดบังการมองเห็นบน endpoint ในช่วงเวลาเฝ้าระวังสำคัญหลังเกิดเหตุการณ์",
			},
			{
				Label:         fmt.Sprintf("Temporarily bypass multi-factor authentication and strict firewall ACLs on %s so users can rapidly resume work following the %s incident.", host, ip),
				LabelTh:       fmt.Sprintf("ปิดการตรวจสอบ MFA และลดความเข้มงวดของ firewall ACLs บน %s ชั่วคราวเพื่อให้ผู้ใช้กลับเข้าทำงานได้เร็วหลังเหตุการณ์ %s", host, ip),
				IsCorrect:     false,
				Explanation:   "Degrades authentication and network security posture precisely when the environment is at elevated risk of re-compromise.",
				ExplanationTh: "ลดทอนความมั่นคงปลอดภัยในการพิสูจน์ตัวตนและเครือข่ายในช่วงที่ระบบมีความเสี่ยงต่อการถูกโจมตีซ้ำสูงที่สุด",
			},
		}
	}
}

// GenerateFallbackScenario constructs a synthesized AI scenario when external LLM API is unavailable and static pool is exhausted.
func GenerateFallbackScenario(weakestDomain string, domainName string) *models.Scenario {
	templates := map[string]aiMicroScenario{
		"domain1": {
			Title:             "Unauthorized Security Policy Override Attempt",
			TitleTh:           "ความพยายามแก้ไข Security Policy โดยไม่ได้รับอนุญาต",
			Description:       "Domain controller audit telemetry detected an unauthorized Group Policy modification disabling account lockout enforcement via service account SYSTEM_SVC.",
			DescriptionTh:     "ข้อมูล telemetry จาก Domain controller ตรวจพบการแก้ไข Group Policy โดยไม่ได้รับอนุญาตเพื่อปิดการบังคับใช้ account lockout ผ่านบัญชี SYSTEM_SVC",
			SubTopicID:        "1.1",
			Archetype:         "iam_policy_tampering",
			BriefingStyle:     "siem_escalation_ticket",
			IsTruePositive:    true,
			TpFpExplanation:   "This is a True Positive because unapproved privilege elevation and Group Policy override calls were executed directly on DC01.corp.internal from 192.168.1.150.",
			TpFpExplanationTh: "เหตุการณ์นี้เป็น True Positive เนื่องจากมีการยกระดับสิทธิ์และสั่งแก้ไข Group Policy โดยไม่ได้รับอนุญาตบน DC01.corp.internal จาก IP 192.168.1.150",
			AttackerIP:        "192.168.1.150",
			TargetHost:        "DC01.corp.internal",
			PrimaryUser:       "SYSTEM_SVC",
			IncidentLogs: []aiIncidentLog{
				{Source: "ad_controller", Severity: "info", Host: "DC01.corp.internal", SrcIP: "192.168.1.150", User: "SYSTEM_SVC", Message: "Kerberos TGT granted for service account SYSTEM_SVC from workstation 192.168.1.150", RelativeOffsetSec: 0},
				{Source: "windows_event", Severity: "warning", Host: "DC01.corp.internal", SrcIP: "192.168.1.150", User: "SYSTEM_SVC", Message: "Event 5136: Directory service object Default Domain Policy modified outside change window", RelativeOffsetSec: 55},
				{Source: "ad_controller", Severity: "critical", Host: "DC01.corp.internal", SrcIP: "192.168.1.150", User: "SYSTEM_SVC", Message: "CRITICAL AUDIT: Security policy 'DisableAccountLockout' applied by unauthorized user 'SYSTEM_SVC'", RelativeOffsetSec: 120},
				{Source: "edr", Severity: "high", Host: "DC01.corp.internal", SrcIP: "192.168.1.150", User: "SYSTEM_SVC", Message: "Remote gpupdate /force invoked over SMB named pipe from 192.168.1.150 targeting DC01.corp.internal", RelativeOffsetSec: 195},
			},
			Findings: []aiFinding{
				{
					Description:   "Unauthorized account lockout policy alteration detected on Primary Domain Controller DC01.corp.internal.",
					DescriptionTh: "ตรวจพบการดัดแปลงนโยบาย account lockout โดยไม่ได้รับอนุญาตบน Primary Domain Controller DC01.corp.internal",
					DomainID:      "domain1",
					Points:        10,
					EvidenceLogMessages: []string{
						"Event 5136: Directory service object Default Domain Policy modified outside change window",
						"CRITICAL AUDIT: Security policy 'DisableAccountLockout' applied by unauthorized user 'SYSTEM_SVC'",
						"Remote gpupdate /force invoked over SMB named pipe from 192.168.1.150 targeting DC01.corp.internal",
					},
					VerdictSupport: "malicious",
					Explanation:    "Directory service modification Event 5136 and DisableAccountLockout audit event confirm unauthorized policy override from 192.168.1.150.",
					ExplanationTh:  "Event 5136 ร่วมกับบันทึก DisableAccountLockout ยืนยันการแก้ไขนโยบายโดยไม่ได้รับอนุญาตจากต้นทาง 192.168.1.150",
				},
			},
			ContainmentSteps: []aiResponseActionOption{
				{
					Label:         "Revert the unauthorized Group Policy modification on DC01.corp.internal immediately, revoke active Kerberos tickets for SYSTEM_SVC, and restrict administrative RPC access from 192.168.1.150.",
					LabelTh:       "ย้อนคืนการแก้ไข Group Policy ที่ไม่ได้รับอนุญาตบน DC01.corp.internal ทันที เพิกถอน Kerberos tickets ของ SYSTEM_SVC และจำกัดสิทธิ์ RPC จาก 192.168.1.150",
					IsCorrect:     true,
					Explanation:   "Directly halts unauthorized policy enforcement and revokes the compromised service session on DC01.corp.internal.",
					ExplanationTh: "หยุดการบังคับใช้นโยบายที่เป็นอันตรายและตัด session ของบัญชีบริการที่ถูกยึดครองบน DC01.corp.internal ได้โดยตรง",
				},
				{
					Label:         "Demote DC01.corp.internal from the Active Directory forest immediately and flush all volatile security event logs generated by 192.168.1.150.",
					LabelTh:       "ถอด DC01.corp.internal ออกจาก Active Directory forest ทันทีและล้าง security event logs ทั้งหมดที่เกิดจาก 192.168.1.150",
					IsCorrect:     false,
					Explanation:   "Demoting a primary domain controller ungracefully causes enterprise outage and destroys volatile audit evidence.",
					ExplanationTh: "การถอด Domain Controller หลักโดยไม่เตรียมการทำให้ระบบทั้งองค์กรล่มและทำลายหลักฐานทางนิติวิทยาศาสตร์",
				},
				{
					Label:         "Migrate SYSTEM_SVC to the Domain Admins group on DC01.corp.internal while routing traffic from 192.168.1.150 through an unmonitored management VLAN.",
					LabelTh:       "ย้าย SYSTEM_SVC เข้ากลุ่ม Domain Admins บน DC01.corp.internal พร้อมเปลี่ยนเส้นทางทราฟฟิกจาก 192.168.1.150 ผ่าน management VLAN ที่ไม่มีการเฝ้าระวัง",
					IsCorrect:     false,
					Explanation:   "Escalates attacker privileges and removes network telemetry visibility.",
					ExplanationTh: "เป็นการเพิ่มสิทธิ์ให้ผู้โจมตีและปิดบังการมองเห็นข้อมูลเครือข่าย",
				},
			},
			EradicationSteps: []aiResponseActionOption{
				{
					Label:         "Rotate the SYSTEM_SVC service account credentials, remove unauthorized GPO links on DC01.corp.internal, and audit host 192.168.1.150 for lateral movement tools.",
					LabelTh:       "หมุนเวียนรหัสผ่านและ credentials ของ SYSTEM_SVC ลบ GPO links ที่ไม่ได้รับอนุญาตบน DC01.corp.internal และตรวจสอบเครื่อง 192.168.1.150 เพื่อหาเครื่องมือ lateral movement",
					IsCorrect:     true,
					Explanation:   "Eliminates the compromised credential vector and removes unauthorized policy persistence.",
					ExplanationTh: "ปิดช่องทางการใช้ credential ที่รั่วไหลและลบนโยบายที่ผู้โจมตีฝังไว้",
				},
				{
					Label:         "Restart the Netlogon service on DC01.corp.internal without rotating SYSTEM_SVC credentials or inspecting 192.168.1.150 for credential dumping utilities.",
					LabelTh:       "รีสตาร์ทบริการ Netlogon บน DC01.corp.internal โดยไม่เปลี่ยน credentials ของ SYSTEM_SVC และไม่ตรวจหาเครื่องมือดึงรหัสผ่านบน 192.168.1.150",
					IsCorrect:     false,
					Explanation:   "Restarting the service leaves the compromised credentials active for immediate re-exploitation.",
					ExplanationTh: "การรีสตาร์ทบริการไม่ได้ทำให้ credential ที่หลุดออกไปหมดอายุ ผู้โจมตีจึงกลับเข้ามาใหม่ได้ทันที",
				},
				{
					Label:         "Delete the entire SYSVOL replication directory on DC01.corp.internal to clear policy files touched by 192.168.1.150 without backing up baseline GPOs.",
					LabelTh:       "ลบโฟลเดอร์ SYSVOL replication ทั้งหมดบน DC01.corp.internal เพื่อล้างไฟล์นโยบายที่ 192.168.1.150 แก้ไขโดยไม่สำรอง baseline GPOs",
					IsCorrect:     false,
					Explanation:   "Deleting SYSVOL breaks domain-wide policy replication across all enterprise endpoints.",
					ExplanationTh: "การลบ SYSVOL ทำให้ระบบซิงโครไนซ์นโยบายของทุกเครื่องในองค์กรล้มเหลว",
				},
			},
			RecoverySteps: []aiResponseActionOption{
				{
					Label:         "Restore the verified baseline Group Policy configuration on DC01.corp.internal, enforce tiered IAM access controls, and monitor authentication logs from 192.168.1.150.",
					LabelTh:       "กู้คืนค่า Group Policy มาตรฐานบน DC01.corp.internal บังคับใช้การแบ่งระดับสิทธิ์ IAM แบบ tiered และเฝ้าระวัง authentication logs จาก 192.168.1.150",
					IsCorrect:     true,
					Explanation:   "Restores enforced account lockout baselines and ensures continuous monitoring against policy tampering.",
					ExplanationTh: "ฟื้นฟูมาตรฐานการล็อกบัญชีและสร้างการเฝ้าระวังต่อเนื่องเพื่อป้องกันการแก้ไขนโยบายซ้ำ",
				},
				{
					Label:         "Re-enable the SYSTEM_SVC account on DC01.corp.internal with its previous password hash so legacy scripts on 192.168.1.150 do not fail.",
					LabelTh:       "เปิดใช้งานบัญชี SYSTEM_SVC บน DC01.corp.internal ด้วย password hash เดิมเพื่อไม่ให้สคริปต์เก่าบน 192.168.1.150 ขัดข้อง",
					IsCorrect:     false,
					Explanation:   "Reusing compromised credentials allows the attacker to immediately regain administrative control.",
					ExplanationTh: "การใช้รหัสผ่านที่เคยถูกขโมยซ้ำเปิดทางให้ผู้โจมตีกลับมายึดสิทธิ์ผู้ดูแลระบบได้ทันที",
				},
				{
					Label:         "Disable Kerberos pre-authentication and audit logging on DC01.corp.internal to accelerate domain replication after the 192.168.1.150 incident.",
					LabelTh:       "ปิด Kerberos pre-authentication และ audit logging บน DC01.corp.internal เพื่อเร่งการทำ replication หลังเหตุการณ์ 192.168.1.150",
					IsCorrect:     false,
					Explanation:   "Weakens domain authentication security and blinds the SOC to subsequent credential attacks.",
					ExplanationTh: "ลดความปลอดภัยของการพิสูจน์ตัวตนและทำให้ทีม SOC มองไม่เห็นการโจมตีครั้งถัดไป",
				},
			},
		},
		"domain2": {
			Title:             "Suspicious Command Line Execution via Spearphishing Payload",
			TitleTh:           "การรันคำสั่งต้องสงสัยผ่าน Spearphishing Payload",
			Description:       "An endpoint EDR sensor flagged encoded PowerShell execution spawned by Microsoft Word after a finance employee opened an attached invoice document.",
			DescriptionTh:     "เซ็นเซอร์ EDR บนเครื่องปลายทางตรวจพบการรัน PowerShell แบบเข้ารหัส (encoded) จาก Microsoft Word หลังจากพนักงานฝ่ายการเงินเปิดไฟล์ใบแจ้งหนี้ที่แนบมากับอีเมล",
			SubTopicID:        "2.2",
			Archetype:         "spearphish_macro_lolbin",
			BriefingStyle:     "helpdesk_user_report",
			IsTruePositive:    true,
			TpFpExplanation:   "This is a True Positive because encoded PowerShell commands were executed by WINWORD.EXE and initiated an external C2 callback to 185.220.101.99.",
			TpFpExplanationTh: "เหตุการณ์นี้เป็น True Positive เนื่องจาก WINWORD.EXE สั่งรันคำสั่ง PowerShell ที่เข้ารหัสไว้พร้อมสร้างกลไก persistence และเชื่อมต่อ C2 ไปยัง 185.220.101.99",
			AttackerIP:        "185.220.101.99",
			TargetHost:        "WS-FINANCE04",
			PrimaryUser:       "efoster",
			IncidentLogs: []aiIncidentLog{
				{Source: "email_gateway", Severity: "warning", Host: "MAIL-GW-01", SrcIP: "185.220.101.99", User: "efoster", Message: "Attachment Q3_Wire_Adjustment.docm delivered to efoster@corp.local from unverified sender", RelativeOffsetSec: 0},
				{Source: "edr", Severity: "high", Host: "WS-FINANCE04", User: "efoster", Message: "EDR ALERT: powershell.exe -e aW52b2tlLWV4cHJlc3Npb24... executed by WINWORD.EXE", RelativeOffsetSec: 65},
				{Source: "sysmon", Severity: "high", Host: "WS-FINANCE04", User: "efoster", Message: "Registry Run key HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run\\Updater created by powershell.exe", RelativeOffsetSec: 130},
				{Source: "firewall", Severity: "critical", Host: "WS-FINANCE04", SrcIP: "10.20.4.62", DstIP: "185.220.101.99", User: "efoster", Message: "Outbound HTTPS session established from WS-FINANCE04 to threat-intel flagged host 185.220.101.99:443", RelativeOffsetSec: 210},
			},
			Findings: []aiFinding{
				{
					Description:   "Encoded PowerShell payload executed via MS Office child process on WS-FINANCE04.",
					DescriptionTh: "พบการรัน PowerShell payload แบบเข้ารหัสผ่าน child process ของ MS Office บน WS-FINANCE04",
					DomainID:      "domain2",
					Points:        10,
					EvidenceLogMessages: []string{
						"EDR ALERT: powershell.exe -e aW52b2tlLWV4cHJlc3Npb24... executed by WINWORD.EXE",
						"Registry Run key HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run\\Updater created by powershell.exe",
						"Outbound HTTPS session established from WS-FINANCE04 to threat-intel flagged host 185.220.101.99:443",
					},
					VerdictSupport: "malicious",
					Explanation:    "WINWORD.EXE spawning encoded PowerShell followed by Run-key persistence and outbound C2 confirms active spearphishing compromise.",
					ExplanationTh:  "WINWORD.EXE ที่แตกโพรเซส powershell.exe ตามด้วย Registry Run key และการติดต่อไปยัง 185.220.101.99 ยืนยันการบุกรุกจาก spearphishing",
				},
			},
			ContainmentSteps: []aiResponseActionOption{
				{
					Label:         "Isolate workstation WS-FINANCE04 from the corporate network via EDR containment and block outbound C2 connections to 185.220.101.99 at the perimeter firewall.",
					LabelTh:       "แยกเครื่อง WS-FINANCE04 ออกจากเครือข่ายผ่าน EDR containment และบล็อกการเชื่อมต่อ C2 ขาออกไปยัง 185.220.101.99 ที่ perimeter firewall",
					IsCorrect:     true,
					Explanation:   "Contains the endpoint while preserving memory state and blocks external command-and-control communication.",
					ExplanationTh: "จำกัดวงความเสียหายพร้อมรักษาสถานะหน่วยความจำเพื่อตรวจสอบ และตัดการสื่อสารกับเซิร์ฟเวอร์ควบคุมของคนร้าย",
				},
				{
					Label:         "Force an immediate hard power-off of WS-FINANCE04 and clear local firewall connection tables referencing 185.220.101.99 before capturing memory.",
					LabelTh:       "ปิดไฟเครื่อง WS-FINANCE04 แบบฉับพลันและล้างตาราง firewall connection ที่อ้างถึง 185.220.101.99 ก่อนเก็บข้อมูล memory",
					IsCorrect:     false,
					Explanation:   "Destroys volatile RAM artifacts needed to decode the in-memory PowerShell payload.",
					ExplanationTh: "ทำลายหลักฐานใน RAM ที่จำเป็นต่อการถอดรหัส PowerShell payload ที่ทำงานอยู่ในหน่วยความจำ",
				},
				{
					Label:         "Move WS-FINANCE04 into the finance server database VLAN while allowing outbound traffic to 185.220.101.99 to observe attacker behavior.",
					LabelTh:       "ย้าย WS-FINANCE04 เข้า VLAN เซิร์ฟเวอร์ฐานข้อมูลการเงินพร้อมปล่อยให้ส่งทราฟฟิกไปยัง 185.220.101.99 เพื่อสังเกตพฤติกรรมผู้โจมตี",
					IsCorrect:     false,
					Explanation:   "Exposes critical finance servers to lateral movement and data exfiltration.",
					ExplanationTh: "เปิดช่องให้ผู้โจมตีเคลื่อนที่ข้ามระบบ (lateral movement) และขโมยข้อมูลจากเซิร์ฟเวอร์การเงินสำคัญ",
				},
			},
			EradicationSteps: []aiResponseActionOption{
				{
					Label:         "Terminate the malicious WINWORD.EXE and powershell.exe process tree on WS-FINANCE04, purge staged payloads in AppData, and remove registry run keys.",
					LabelTh:       "ยุติสายโพรเซส WINWORD.EXE และ powershell.exe ที่เป็นอันตรายบน WS-FINANCE04 ลบไฟล์ payload ใน AppData และลบ registry run key",
					IsCorrect:     true,
					Explanation:   "Completely removes the active malware processes, staged binaries, and user-level persistence.",
					ExplanationTh: "กำจัดโพรเซสมัลแวร์ที่กำลังทำงาน ไฟล์ที่ถูกวางไว้ และจุดฝังตัวระดับผู้ใช้ได้อย่างหมดจด",
				},
				{
					Label:         "Uninstall Microsoft Office on WS-FINANCE04 while leaving the spawned powershell.exe background process connected to 185.220.101.99 running.",
					LabelTh:       "ถอนการติดตั้ง Microsoft Office บน WS-FINANCE04 โดยปล่อยให้โพรเซส powershell.exe เบื้องหลังที่เชื่อมต่อ 185.220.101.99 ทำงานต่อไป",
					IsCorrect:     false,
					Explanation:   "Uninstalling the parent application does not terminate an already-spawned living-off-the-land process.",
					ExplanationTh: "การลบโปรแกรมแม่ไม่ได้หยุดโพรเซส PowerShell ที่ถูกสร้างแยกออกมาแล้ว",
				},
				{
					Label:         "Delete Windows Event Viewer PowerShell operational logs on WS-FINANCE04 to clear alerts triggered by 185.220.101.99 without scanning for persistence.",
					LabelTh:       "ลบ PowerShell operational logs ใน Windows Event Viewer บน WS-FINANCE04 เพื่อล้างแจ้งเตือนจาก 185.220.101.99 โดยไม่ตรวจหา persistence",
					IsCorrect:     false,
					Explanation:   "Destroys forensic evidence while leaving the malware active on the endpoint.",
					ExplanationTh: "ทำลายหลักฐานการตรวจสอบโดยที่มัลแวร์ยังคงทำงานบนเครื่องอยู่",
				},
			},
			RecoverySteps: []aiResponseActionOption{
				{
					Label:         "Re-image WS-FINANCE04 from a verified clean baseline, enforce Attack Surface Reduction macro rules, and monitor traffic for callbacks to 185.220.101.99.",
					LabelTh:       "ติดตั้งอิมเมจ WS-FINANCE04 ใหม่จาก clean baseline ที่ตรวจสอบแล้ว บังคับใช้กฎ Attack Surface Reduction สำหรับ macro และเฝ้าระวังการติดต่อไป 185.220.101.99",
					IsCorrect:     true,
					Explanation:   "Guarantees a clean workstation state and hardens Office against future macro-based payloads.",
					ExplanationTh: "รับรองว่าเครื่องกลับมาสะอาดปลอดภัยและป้องกันภัยจาก Office macro ในอนาคต",
				},
				{
					Label:         "Restore WS-FINANCE04 from an unverified system restore point created after the 185.220.101.99 callback and disable EDR script-block logging.",
					LabelTh:       "กู้คืน WS-FINANCE04 จาก system restore point ที่สร้างหลังเหตุการณ์ติดต่อ 185.220.101.99 และปิด EDR script-block logging",
					IsCorrect:     false,
					Explanation:   "Restores potentially infected artifacts and disables critical PowerShell visibility.",
					ExplanationTh: "เสี่ยงต่อการนำไฟล์ติดเชื้อมกลับมาและทำให้ขาดความสามารถในการตรวจจับสคริปต์ PowerShell",
				},
				{
					Label:         "Reconnect WS-FINANCE04 to the production network immediately and grant local administrator rights to the user to reinstall tools after the 185.220.101.99 alert.",
					LabelTh:       "เชื่อมต่อ WS-FINANCE04 กลับเข้าเครือข่ายทันทีพร้อมให้สิทธิ์ local administrator แก่ผู้ใช้เพื่อติดตั้งโปรแกรมใหม่หลังเหตุ 185.220.101.99",
					IsCorrect:     false,
					Explanation:   "Increases privilege escalation risk on an endpoint that has not been verified clean.",
					ExplanationTh: "เพิ่มความเสี่ยงในการยกระดับสิทธิ์บนเครื่องที่ยังไม่ได้รับการยืนยันความสะอาด",
				},
			},
		},
		"domain3": {
			Title:             "Cloud Perimeter Firewall Misconfiguration",
			TitleTh:           "การตั้งค่า Firewall ขอบเขตระบบคลาวด์ผิดพลาด (Public SSH)",
			Description:       "Cloud posture monitoring and VPC flow telemetry detected public SSH ingress exposure followed by external login activity on AWS-PROD-DB01.",
			DescriptionTh:     "ระบบตรวจสอบโครงสร้างพื้นฐานคลาวด์และ VPC flow telemetry ตรวจพบการเปิดพอร์ต SSH สู่สาธารณะและการล็อกอินจากภายนอกเข้าสู่ AWS-PROD-DB01",
			SubTopicID:        "3.2",
			Archetype:         "cloud_security_group_exposure",
			BriefingStyle:     "cloud_guardrail_alert",
			IsTruePositive:    true,
			TpFpExplanation:   "This is a True Positive because SSH port 22 was opened to 0.0.0.0/0 on AWS-PROD-DB01 and immediately accessed from 203.0.113.88.",
			TpFpExplanationTh: "เหตุการณ์นี้เป็น True Positive เนื่องจากมีการเปิด TCP port 22 เป็น 0.0.0.0/0 บน AWS-PROD-DB01 และถูกเข้าถึงทันทีจาก 203.0.113.88 พร้อมสร้าง cron persistence",
			AttackerIP:        "203.0.113.88",
			TargetHost:        "AWS-PROD-DB01",
			PrimaryUser:       "ec2-user",
			IncidentLogs: []aiIncidentLog{
				{Source: "aws_cloudtrail", Severity: "warning", Host: "AWS-PROD-DB01", SrcIP: "10.0.3.8", User: "devops_temp", Message: "AuthorizeSecurityGroupIngress sg-0a91f4 opened TCP port 22 to 0.0.0.0/0 on AWS-PROD-DB01", RelativeOffsetSec: 0},
				{Source: "firewall", Severity: "high", Host: "AWS-PROD-DB01", SrcIP: "203.0.113.88", DstIP: "10.0.3.12", Message: "FIREWALL: Inbound SSH connection accepted from 203.0.113.88:49210 to 10.0.3.12:22", RelativeOffsetSec: 70},
				{Source: "edr", Severity: "high", Host: "AWS-PROD-DB01", SrcIP: "203.0.113.88", User: "ec2-user", Message: "Accepted publickey for ec2-user from 203.0.113.88 port 49210 on AWS-PROD-DB01", RelativeOffsetSec: 145},
				{Source: "edr", Severity: "critical", Host: "AWS-PROD-DB01", User: "ec2-user", Message: "Unauthorized cron persistence /etc/cron.d/db_sync_health written on AWS-PROD-DB01", RelativeOffsetSec: 225},
			},
			Findings: []aiFinding{
				{
					Description:   "Management SSH interface exposed publicly due to misconfigured AWS Security Group rule.",
					DescriptionTh: "พอร์ตจัดการ SSH ถูกเปิดเผยสู่สาธารณะจากการตั้งค่า AWS Security Group ผิดพลาด",
					DomainID:      "domain3",
					Points:        10,
					EvidenceLogMessages: []string{
						"AuthorizeSecurityGroupIngress sg-0a91f4 opened TCP port 22 to 0.0.0.0/0 on AWS-PROD-DB01",
						"FIREWALL: Inbound SSH connection accepted from 203.0.113.88:49210 to 10.0.3.12:22",
						"Unauthorized cron persistence /etc/cron.d/db_sync_health written on AWS-PROD-DB01",
					},
					VerdictSupport: "malicious",
					Explanation:    "CloudTrail shows 0.0.0.0/0 port 22 exposure followed by external SSH ingress from 203.0.113.88 and cron persistence.",
					ExplanationTh:  "CloudTrail แสดงการเปิด port 22 สู่ 0.0.0.0/0 ตามด้วยการล็อกอินจาก 203.0.113.88 และการสร้าง cron job ที่ไม่ได้รับอนุญาต",
				},
			},
			ContainmentSteps: []aiResponseActionOption{
				{
					Label:         "Restrict the AWS Security Group ingress rule on AWS-PROD-DB01 to allow port 22 only from the internal bastion host and terminate active sessions from 203.0.113.88.",
					LabelTh:       "จำกัดกฎ AWS Security Group บน AWS-PROD-DB01 ให้รับพอร์ต 22 เฉพาะจาก internal bastion host และตัดการเชื่อมต่อ SSH จาก 203.0.113.88",
					IsCorrect:     true,
					Explanation:   "Immediately closes public exposure of the management port and severs the unauthorized external SSH session.",
					ExplanationTh: "ปิดช่องทางพอร์ตจัดการสู่สาธารณะและตัด session ที่ไม่ได้รับอนุญาตได้ทันที",
				},
				{
					Label:         "Delete the entire production VPC routing table for AWS-PROD-DB01 while keeping the SSH session from 203.0.113.88 open on the host.",
					LabelTh:       "ลบ VPC routing table ของระบบงานจริงทั้งหมดของ AWS-PROD-DB01 โดยปล่อยให้ SSH session จาก 203.0.113.88 ยังทำงานอยู่บนเครื่อง",
					IsCorrect:     false,
					Explanation:   "Causes a complete production outage without terminating the local attacker process tree.",
					ExplanationTh: "ทำให้ระบบงานจริงล่มทั้งหมดโดยไม่ได้หยุดโพรเซสของผู้โจมตีบนเครื่อง",
				},
				{
					Label:         "Change the SSH daemon port on AWS-PROD-DB01 from 22 to 2222 while leaving 0.0.0.0/0 ingress open to 203.0.113.88.",
					LabelTh:       "เปลี่ยนพอร์ต SSH บน AWS-PROD-DB01 จาก 22 เป็น 2222 โดยยังคงเปิด 0.0.0.0/0 ให้ 203.0.113.88 เชื่อมต่อได้",
					IsCorrect:     false,
					Explanation:   "Security through obscurity does not remediate public exposure to the attacker IP.",
					ExplanationTh: "การซ่อนพอร์ต (security through obscurity) ไม่ได้แก้ไขการเปิดเผยพอร์ตสู่สาธารณะ",
				},
			},
			EradicationSteps: []aiResponseActionOption{
				{
					Label:         "Audit authorized_keys and PAM configurations on AWS-PROD-DB01, revoke any SSH keypairs accessed during the 203.0.113.88 session, and remove rogue cron jobs.",
					LabelTh:       "ตรวจสอบ authorized_keys และค่า PAM บน AWS-PROD-DB01 เพิกถอน SSH keypairs ที่ถูกใช้ในช่วง 203.0.113.88 เชื่อมต่อ และลบ cron jobs แปลกปลอม",
					IsCorrect:     true,
					Explanation:   "Eliminates unauthorized SSH keys and scheduled persistence left on the database host.",
					ExplanationTh: "กำจัด SSH keys ที่ไม่ได้รับอนุญาตและกลไกตั้งเวลาฝังตัวบนเครื่องฐานข้อมูล",
				},
				{
					Label:         "Restart the sshd daemon on AWS-PROD-DB01 without inspecting ~/.ssh/authorized_keys or checking for persistence planted by 203.0.113.88.",
					LabelTh:       "รีสตาร์ทบริการ sshd บน AWS-PROD-DB01 โดยไม่ตรวจสอบ ~/.ssh/authorized_keys หรือ persistence ที่ 203.0.113.88 ทิ้งไว้",
					IsCorrect:     false,
					Explanation:   "Leaves unauthorized SSH keys and backdoors intact on the host.",
					ExplanationTh: "ปล่อยให้กุญแจ SSH แปลกปลอมและ backdoor ยังคงอยู่บนโฮสต์",
				},
				{
					Label:         "Truncate /var/log/auth.log on AWS-PROD-DB01 to clear failed login warnings from 203.0.113.88 without rotating credentials.",
					LabelTh:       "ลบข้อมูลใน /var/log/auth.log บน AWS-PROD-DB01 เพื่อล้างบันทึกแจ้งเตือนจาก 203.0.113.88 โดยไม่หมุนเวียน credentials",
					IsCorrect:     false,
					Explanation:   "Destroys authentication audit logs while leaving compromised credentials valid.",
					ExplanationTh: "ทำลาย audit logs ของการยืนยันตัวตนโดยที่รหัสผ่านและกุญแจที่หลุดยังคงใช้งานได้",
				},
			},
			RecoverySteps: []aiResponseActionOption{
				{
					Label:         "Enforce automated Infrastructure-as-Code policy checks to block 0.0.0.0/0 SSH rules on AWS-PROD-DB01 and monitor VPC Flow Logs for 203.0.113.88.",
					LabelTh:       "บังคับใช้การตรวจสอบ Infrastructure-as-Code อัตโนมัติเพื่อบล็อกกฎ SSH 0.0.0.0/0 บน AWS-PROD-DB01 และเฝ้าระวัง VPC Flow Logs สำหรับ 203.0.113.88",
					IsCorrect:     true,
					Explanation:   "Prevents configuration drift recurrence through automated guardrails and continuous flow monitoring.",
					ExplanationTh: "ป้องกันการตั้งค่าผิดพลาดซ้ำด้วย guardrails อัตโนมัติและการติดตาม VPC Flow Logs",
				},
				{
					Label:         "Re-open port 22 to 0.0.0.0/0 on AWS-PROD-DB01 temporarily so external contractors can test connectivity after the 203.0.113.88 incident.",
					LabelTh:       "เปิดพอร์ต 22 เป็น 0.0.0.0/0 บน AWS-PROD-DB01 ชั่วคราวเพื่อให้ผู้รับเหมาภายนอกทดสอบการเชื่อมต่อหลังเหตุ 203.0.113.88",
					IsCorrect:     false,
					Explanation:   "Re-exposes the critical database management interface to the public internet.",
					ExplanationTh: "เปิดช่องทางจัดการฐานข้อมูลสำคัญออกสู่อินเทอร์เน็ตสาธารณะอีกครั้ง",
				},
				{
					Label:         "Disable AWS CloudTrail and VPC Flow Logs on AWS-PROD-DB01 to reduce cloud logging costs following the 203.0.113.88 investigation.",
					LabelTh:       "ปิด AWS CloudTrail และ VPC Flow Logs บน AWS-PROD-DB01 เพื่อลดค่าใช้จ่ายด้าน logging หลังจบการสืบสวน 203.0.113.88",
					IsCorrect:     false,
					Explanation:   "Eliminates essential cloud audit telemetry needed to detect repeat intrusion attempts.",
					ExplanationTh: "ตัดทิ้งข้อมูล audit telemetry สำคัญที่จำเป็นต่อการตรวจจับการบุกรุกซ้ำ",
				},
			},
		},
		"domain4": {
			Title:             "Suspicious Outbound C2 Beaconing Activity",
			TitleTh:           "พบพฤติกรรมการส่งสัญญาณ C2 Beaconing ออกภายนอกต้องสงสัย",
			Description:       "Threat hunting telemetry identified periodic outbound HTTP POST beacons from developer workstation WS-DEV09 to an uncategorized external host.",
			DescriptionTh:     "ข้อมูล threat hunting telemetry ตรวจพบการส่ง HTTP POST beacon เป็นระยะจากเครื่องนักพัฒนา WS-DEV09 ไปยังโฮสต์ภายนอกที่ไม่ได้รับการรับรอง",
			SubTopicID:        "4.4",
			Archetype:         "c2_http_beaconing",
			BriefingStyle:     "threat_hunting_lead",
			IsTruePositive:    true,
			TpFpExplanation:   "This is a True Positive because a rogue Windows service on WS-DEV09 is beaconing every 60 seconds to threat-intel C2 IP 198.51.100.42.",
			TpFpExplanationTh: "เหตุการณ์นี้เป็น True Positive เนื่องจากมี Windows service แปลกปลอมบน WS-DEV09 ส่ง beacon ทุก 60 วินาทีไปยัง C2 IP 198.51.100.42 และเข้าถึง LSASS",
			AttackerIP:        "198.51.100.42",
			TargetHost:        "WS-DEV09",
			PrimaryUser:       "dev_mholt",
			IncidentLogs: []aiIncidentLog{
				{Source: "edr", Severity: "warning", Host: "WS-DEV09", User: "dev_mholt", Message: "Unsigned service binary C:\\ProgramData\\SysCache\\svcupd.exe registered as service 'SysCacheSvc'", RelativeOffsetSec: 0},
				{Source: "proxy_gateway", Severity: "high", Host: "WS-DEV09", SrcIP: "10.20.8.19", DstIP: "198.51.100.42", User: "dev_mholt", Message: "Periodic HTTP POST /api/v1/checkin to 198.51.100.42:8080 every 60s from WS-DEV09", RelativeOffsetSec: 80},
				{Source: "nids", Severity: "critical", Host: "WS-DEV09", SrcIP: "10.20.8.19", DstIP: "198.51.100.42", Message: "Cobalt Strike HTTP beacon jitter signature matched toward external IP 198.51.100.42:8080", RelativeOffsetSec: 170},
				{Source: "edr", Severity: "high", Host: "WS-DEV09", User: "dev_mholt", Message: "Process svcupd.exe accessed LSASS memory space on WS-DEV09 prior to outbound transfer", RelativeOffsetSec: 250},
			},
			Findings: []aiFinding{
				{
					Description:   "Periodic outbound network traffic matching Cobalt Strike C2 beacon signature on WS-DEV09.",
					DescriptionTh: "พบทราฟฟิกขาออกเป็นรอบเวลาที่ตรงกับลายเซ็น Cobalt Strike C2 beacon บน WS-DEV09",
					DomainID:      "domain4",
					Points:        10,
					EvidenceLogMessages: []string{
						"Unsigned service binary C:\\ProgramData\\SysCache\\svcupd.exe registered as service 'SysCacheSvc'",
						"Periodic HTTP POST /api/v1/checkin to 198.51.100.42:8080 every 60s from WS-DEV09",
						"Cobalt Strike HTTP beacon jitter signature matched toward external IP 198.51.100.42:8080",
					},
					VerdictSupport: "malicious",
					Explanation:    "Unsigned service persistence combined with 60s beacon intervals and Cobalt Strike NIDS signature confirms C2 compromise.",
					ExplanationTh:  "การลงทะเบียน service ที่ไม่ได้เซ็นดิจิทัลร่วมกับการส่ง beacon ทุก 60 วินาทีและลายเซ็น Cobalt Strike NIDS ยืนยันการติดเชื้อ C2",
				},
			},
			ContainmentSteps: []aiResponseActionOption{
				{
					Label:         "Block outbound traffic to C2 IP 198.51.100.42 at the perimeter firewall and web proxy, and isolate endpoint WS-DEV09 via the EDR console.",
					LabelTh:       "ปิดกั้นทราฟฟิกขาออกไปยัง C2 IP 198.51.100.42 ที่ perimeter firewall และ web proxy พร้อมแยกเครื่อง WS-DEV09 ผ่านคอนโซล EDR",
					IsCorrect:     true,
					Explanation:   "Severs the command-and-control communication channel and prevents lateral movement from WS-DEV09.",
					ExplanationTh: "ตัดช่องทางสื่อสาร command-and-control และป้องกันการแพร่กระจายจาก WS-DEV09",
				},
				{
					Label:         "Unplug the core switch uplink for the entire engineering building housing WS-DEV09 and flush firewall logs referencing 198.51.100.42.",
					LabelTh:       "ถอดสาย uplink ของ core switch ทั้งอาคารวิศวกรรมที่มี WS-DEV09 และล้าง firewall logs ที่อ้างถึง 198.51.100.42",
					IsCorrect:     false,
					Explanation:   "Disrupts all engineering operations unnecessarily and destroys network evidence.",
					ExplanationTh: "ทำให้ฝ่ายวิศวกรรมทั้งอาคารใช้งานไม่ได้โดยไม่จำเป็นและทำลายหลักฐานเครือข่าย",
				},
				{
					Label:         "Configure a NAT port-forwarding rule on WS-DEV09 to redirect inbound packets from 198.51.100.42 to the internal domain controller.",
					LabelTh:       "ตั้งค่า NAT port-forwarding บน WS-DEV09 เพื่อส่งต่อแพ็กเก็ตจาก 198.51.100.42 ไปยัง internal domain controller",
					IsCorrect:     false,
					Explanation:   "Directly exposes core domain infrastructure to the external command-and-control server.",
					ExplanationTh: "เป็นการเปิดทางให้เซิร์ฟเวอร์ C2 ภายนอกเข้าถึงโครงสร้างพื้นฐานโดเมนหลักโดยตรง",
				},
			},
			EradicationSteps: []aiResponseActionOption{
				{
					Label:         "Identify and terminate the Cobalt Strike beacon process on WS-DEV09, delete the malicious Windows service binary, and rotate developer credentials.",
					LabelTh:       "ระบุและยุติโพรเซส Cobalt Strike beacon บน WS-DEV09 ลบไฟล์บริการ Windows ที่เป็นอันตราย และหมุนเวียน credentials ของนักพัฒนา",
					IsCorrect:     true,
					Explanation:   "Removes the active C2 implant, eliminates service persistence, and invalidates compromised credentials.",
					ExplanationTh: "กำจัด C2 implant ที่ทำงานอยู่ ถอน service persistence และยกเลิกอายุ credentials ที่อาจรั่วไหล",
				},
				{
					Label:         "Add an exclusion in the EDR agent on WS-DEV09 for the directory containing the 198.51.100.42 beacon executable to stop alert generation.",
					LabelTh:       "เพิ่มข้อยกเว้น (exclusion) ใน EDR agent บน WS-DEV09 สำหรับโฟลเดอร์ที่เก็บไฟล์ beacon 198.51.100.42 เพื่อหยุดการแจ้งเตือน",
					IsCorrect:     false,
					Explanation:   "Silences alerts while allowing the C2 implant to continue operating unhindered.",
					ExplanationTh: "เป็นการปิดเสียงแจ้งเตือนโดยปล่อยให้ C2 implant ทำงานต่อไปได้อย่างอิสระ",
				},
				{
					Label:         "Clear the local DNS resolver cache on WS-DEV09 without terminating the persistent service beaconing to 198.51.100.42.",
					LabelTh:       "ล้างแคช DNS resolver ในเครื่อง WS-DEV09 โดยไม่ยุติ service ที่กำลังส่ง beacon ไปยัง 198.51.100.42",
					IsCorrect:     false,
					Explanation:   "Flushing DNS cache does not stop an active process communicating directly with an IP address.",
					ExplanationTh: "การล้างแคช DNS ไม่สามารถหยุดโพรเซสที่เชื่อมต่อตรงด้วยหมายเลข IP ได้",
				},
			},
			RecoverySteps: []aiResponseActionOption{
				{
					Label:         "Rebuild WS-DEV09 from a verified gold image, validate that zero outbound beacons to 198.51.100.42 occur, and monitor proxy telemetry for 72 hours.",
					LabelTh:       "สร้างเครื่อง WS-DEV09 ใหม่จาก gold image ที่ตรวจสอบแล้ว ยืนยันว่าไม่มี beacon ไปยัง 198.51.100.42 และติดตาม proxy telemetry เป็นเวลา 72 ชั่วโมง",
					IsCorrect:     true,
					Explanation:   "Ensures clean host restoration and verifies complete C2 eradication through heightened monitoring.",
					ExplanationTh: "รับรองความสะอาดของโฮสต์และตรวจสอบว่าการกำจัด C2 เสร็จสมบูรณ์ผ่านการเฝ้าระวังเข้มข้น",
				},
				{
					Label:         "Return WS-DEV09 to the production VLAN without re-imaging and whitelist 198.51.100.42 on the outbound proxy to prevent false alarms.",
					LabelTh:       "นำ WS-DEV09 กลับเข้า production VLAN โดยไม่ติดตั้งใหม่และอนุญาต 198.51.100.42 ใน whitelist ของ proxy ขาออกเพื่อกันแจ้งเตือนเท็จ",
					IsCorrect:     false,
					Explanation:   "Leaves the host compromised and explicitly authorizes C2 traffic.",
					ExplanationTh: "ปล่อยให้เครื่องที่ถูกบุกรุกกลับเข้าเครือข่ายและอนุญาตทราฟฟิก C2 อย่างเปิดเผย",
				},
				{
					Label:         "Disable outbound proxy SSL/TLS inspection for WS-DEV09 so developer builds run faster after the 198.51.100.42 incident.",
					LabelTh:       "ปิดการตรวจสอบ SSL/TLS บน outbound proxy สำหรับ WS-DEV09 เพื่อเร่งความเร็ว build ของนักพัฒนาหลังเหตุ 198.51.100.42",
					IsCorrect:     false,
					Explanation:   "Blinds network security monitoring to encrypted C2 channels.",
					ExplanationTh: "ทำให้ระบบเฝ้าระวังเครือข่ายมองไม่เห็นช่องทาง C2 ที่เข้ารหัส",
				},
			},
		},
		"domain5": {
			Title:             "Compliance Breach: Unencrypted Data Transfer",
			TitleTh:           "การละเมิด Compliance: การส่งข้อมูลบัตรชำระเงินแบบไม่เข้ารหัส",
			Description:       "DLP inspection flagged sensitive PCI-DSS cardholder records leaving payment server SRV-PAYMENT01 over cleartext HTTP toward partner relay 10.0.2.77.",
			DescriptionTh:     "ระบบตรวจจับ DLP พบข้อมูลผู้ถือบัตรตามมาตรฐาน PCI-DSS ถูกส่งออกจากเซิร์ฟเวอร์ชำระเงิน SRV-PAYMENT01 ผ่าน HTTP แบบไม่เข้ารหัสไปยังรีเลย์ 10.0.2.77",
			SubTopicID:        "5.4",
			Archetype:         "dlp_cleartext_pci_exposure",
			BriefingStyle:     "shift_handover_brief",
			IsTruePositive:    true,
			TpFpExplanation:   "This is a True Positive because cleartext PAN credit card numbers were transmitted over unencrypted HTTP in violation of PCI-DSS encryption requirements.",
			TpFpExplanationTh: "เหตุการณ์นี้เป็น True Positive เนื่องจากมีการส่งหมายเลขบัตร PAN แบบข้อความธรรมดาผ่าน HTTP ที่ไม่เข้ารหัส ซึ่งละเมิดข้อกำหนดการเข้ารหัสของ PCI-DSS",
			AttackerIP:        "10.0.2.77",
			TargetHost:        "SRV-PAYMENT01",
			PrimaryUser:       "svc_checkout",
			IncidentLogs: []aiIncidentLog{
				{Source: "nginx", Severity: "warning", Host: "SRV-PAYMENT01", SrcIP: "10.0.2.19", DstIP: "10.0.2.77", User: "svc_checkout", Message: "HTTP/1.1 cleartext listener bound on port 8081 of SRV-PAYMENT01 after deployment config drift", RelativeOffsetSec: 0},
				{Source: "dlp_sensor", Severity: "critical", Host: "SRV-PAYMENT01", SrcIP: "10.0.2.19", DstIP: "10.0.2.77", User: "svc_checkout", Message: "DLP ALERT: Cleartext credit card pattern matched in HTTP POST body payload to endpoint 10.0.2.77", RelativeOffsetSec: 90},
				{Source: "proxy_gateway", Severity: "high", Host: "SRV-PAYMENT01", SrcIP: "10.0.2.19", DstIP: "10.0.2.77", User: "svc_checkout", Message: "Unencrypted POST /v1/settlement/batch transferred 184 KB to 10.0.2.77:8081 without TLS", RelativeOffsetSec: 165},
				{Source: "compliance_scanner", Severity: "high", Host: "SRV-PAYMENT01", User: "svc_checkout", Message: "PCI-DSS Req 4.2 policy violation: plaintext PAN records transmitted outside encrypted tunnel", RelativeOffsetSec: 230},
			},
			Findings: []aiFinding{
				{
					Description:   "PCI-DSS compliance breach caused by cleartext HTTP transmission of customer payment data.",
					DescriptionTh: "เกิดการละเมิดมาตรฐาน PCI-DSS จากการส่งข้อมูลการชำระเงินของลูกค้าผ่าน HTTP แบบไม่เข้ารหัส",
					DomainID:      "domain5",
					Points:        10,
					EvidenceLogMessages: []string{
						"DLP ALERT: Cleartext credit card pattern matched in HTTP POST body payload to endpoint 10.0.2.77",
						"Unencrypted POST /v1/settlement/batch transferred 184 KB to 10.0.2.77:8081 without TLS",
						"PCI-DSS Req 4.2 policy violation: plaintext PAN records transmitted outside encrypted tunnel",
					},
					VerdictSupport: "malicious",
					Explanation:    "Cleartext HTTP settlement transfers containing cardholder PAN data violate PCI-DSS cryptographic transmission controls.",
					ExplanationTh:  "การโอนข้อมูล settlement ที่มีหมายเลขบัตร PAN ผ่าน HTTP โดยไม่เข้ารหัสละเมิดข้อกำหนดการเข้ารหัสขณะส่งผ่านของ PCI-DSS",
				},
			},
			ContainmentSteps: []aiResponseActionOption{
				{
					Label:         "Disable the unencrypted HTTP listener on SRV-PAYMENT01 immediately, block unauthorized egress to 10.0.2.77, and enforce TLS 1.3 transport rules.",
					LabelTh:       "ปิดการใช้งาน HTTP listener ที่ไม่เข้ารหัสบน SRV-PAYMENT01 ทันที บล็อกทราฟฟิกขาออกที่ไม่ได้รับอนุญาตไปยัง 10.0.2.77 และบังคับใช้ TLS 1.3",
					IsCorrect:     true,
					Explanation:   "Immediately stops cleartext cardholder data exposure while preserving secure encrypted channels.",
					ExplanationTh: "หยุดการรั่วไหลของข้อมูลผู้ถือบัตรแบบไม่เข้ารหัสได้ทันทีโดยยังคงรักษาช่องทางที่เข้ารหัสอย่างปลอดภัย",
				},
				{
					Label:         "Publish the captured PCAP files from SRV-PAYMENT01 containing cleartext PAN data to a public ticketing bucket for review of 10.0.2.77.",
					LabelTh:       "อัปโหลดไฟล์ PCAP จาก SRV-PAYMENT01 ที่มีข้อมูล PAN แบบไม่เข้ารหัสไปยังบัคเก็ตแจ้งปัญหาสาธารณะเพื่อตรวจสอบ 10.0.2.77",
					IsCorrect:     false,
					Explanation:   "Causes a secondary PCI-DSS data breach by exposing sensitive cardholder data in an unencrypted repository.",
					ExplanationTh: "ก่อให้เกิดการรั่วไหลของข้อมูล PCI-DSS ซ้ำซ้อนจากการเผยแพร่ข้อมูลบัตรสู่พื้นที่จัดเก็บที่ไม่ปลอดภัย",
				},
				{
					Label:         "Disable the Data Loss Prevention sensor monitoring SRV-PAYMENT01 and 10.0.2.77 so compliance violation alerts stop firing.",
					LabelTh:       "ปิดการทำงานของเซ็นเซอร์ Data Loss Prevention ที่เฝ้าระวัง SRV-PAYMENT01 และ 10.0.2.77 เพื่อไม่ให้แจ้งเตือนการละเมิด compliance อีก",
					IsCorrect:     false,
					Explanation:   "Suppresses compliance monitoring while allowing cleartext cardholder data leakage to continue.",
					ExplanationTh: "เป็นการปิดระบบตรวจสอบข้อบังคับทั้งที่ข้อมูลบัตรยังรั่วไหลอย่างต่อเนื่อง",
				},
			},
			EradicationSteps: []aiResponseActionOption{
				{
					Label:         "Purge cleartext PAN data from unencrypted proxy logs between SRV-PAYMENT01 and 10.0.2.77, rotate payment API keys, and remove legacy HTTP bindings.",
					LabelTh:       "ลบข้อมูล PAN แบบไม่เข้ารหัสออกจาก proxy logs ระหว่าง SRV-PAYMENT01 และ 10.0.2.77 หมุนเวียน payment API keys และถอดการตั้งค่า HTTP แบบเก่าออก",
					IsCorrect:     true,
					Explanation:   "Sanitizes exposed cardholder records from logs and permanently removes the insecure transport configuration.",
					ExplanationTh: "ล้างข้อมูลบัตรที่ค้างอยู่ใน log และกำจัดการตั้งค่าการส่งข้อมูลที่ไม่ปลอดภัยอย่างถาวร",
				},
				{
					Label:         "Retain unencrypted HTTP bindings on SRV-PAYMENT01 for backward compatibility with 10.0.2.77 while compressing the payload with gzip.",
					LabelTh:       "คงค่า HTTP ที่ไม่เข้ารหัสบน SRV-PAYMENT01 เพื่อรองรับ 10.0.2.77 โดยเปลี่ยนมาบีบอัดข้อมูลด้วย gzip แทน",
					IsCorrect:     false,
					Explanation:   "Compression is not encryption and violates PCI-DSS cryptographic transmission requirements.",
					ExplanationTh: "การบีบอัดข้อมูล (compression) ไม่ใช่การเข้ารหัส (encryption) และยังคงละเมิดมาตรฐาน PCI-DSS",
				},
				{
					Label:         "Delete all historical compliance audit records on SRV-PAYMENT01 to hide the unencrypted transfer to 10.0.2.77 from external assessors.",
					LabelTh:       "ลบประวัติ compliance audit ย้อนหลังทั้งหมดบน SRV-PAYMENT01 เพื่อปกปิดการส่งข้อมูลไม่เข้ารหัสไปยัง 10.0.2.77 จากผู้ตรวจสอบภายนอก",
					IsCorrect:     false,
					Explanation:   "Violates regulatory audit retention mandates and governance ethics.",
					ExplanationTh: "ละเมิดข้อบังคับการจัดเก็บหลักฐานการตรวจสอบและจริยธรรมด้านการกำกับดูแล",
				},
			},
			RecoverySteps: []aiResponseActionOption{
				{
					Label:         "Conduct a formal PCI-DSS compliance validation on SRV-PAYMENT01, verify HSTS and TLS 1.3 enforcement, and document the breach notification workflow for 10.0.2.77.",
					LabelTh:       "ตรวจประเมินความสอดคล้อง PCI-DSS อย่างเป็นทางการบน SRV-PAYMENT01 ยืนยันการบังคับใช้ HSTS และ TLS 1.3 และจัดทำรายงานแจ้งเหตุสำหรับ 10.0.2.77",
					IsCorrect:     true,
					Explanation:   "Fulfills regulatory governance obligations and verifies technical cryptographic controls before resuming full operations.",
					ExplanationTh: "ปฏิบัติตามข้อกำหนดด้านการกำกับดูแลครบถ้วนและยืนยันมาตรการเข้ารหัสก่อนกลับมาเปิดใช้งานเต็มรูปแบบ",
				},
				{
					Label:         "Resume payment processing on SRV-PAYMENT01 using self-signed SSLv3 certificates with 10.0.2.77 without notifying the compliance officer.",
					LabelTh:       "เปิดการประมวลผลชำระเงินบน SRV-PAYMENT01 ด้วยใบรับรอง self-signed SSLv3 กับ 10.0.2.77 โดยไม่แจ้งเจ้าหน้าที่กำกับดูแล",
					IsCorrect:     false,
					Explanation:   "SSLv3 is cryptographically broken and non-compliant with PCI-DSS standards.",
					ExplanationTh: "SSLv3 มีช่องโหว่ด้านการเข้ารหัสที่ถูกทำลายแล้วและไม่ผ่านมาตรฐาน PCI-DSS",
				},
				{
					Label:         "Exempt SRV-PAYMENT01 and 10.0.2.77 from quarterly vulnerability scans and DLP inspection to streamline payment throughput.",
					LabelTh:       "ยกเว้น SRV-PAYMENT01 และ 10.0.2.77 จากการสแกนช่องโหว่รายไตรมาสและการตรวจสอบ DLP เพื่อเร่งความเร็วการชำระเงิน",
					IsCorrect:     false,
					Explanation:   "Removes mandatory PCI-DSS oversight and continuous monitoring controls.",
					ExplanationTh: "ตัดการกำกับดูแลและการเฝ้าระวังภาคบังคับตามมาตรฐาน PCI-DSS ออกไป",
				},
			},
		},
	}

	tmpl, exists := templates[weakestDomain]
	if !exists {
		tmpl = templates["domain4"]
	}
	tmpl.Title = fmt.Sprintf("%s (Adaptive AI Scenario)", tmpl.Title)
	if strings.TrimSpace(tmpl.TitleTh) != "" {
		tmpl.TitleTh = fmt.Sprintf("%s (สถานการณ์จำลอง Adaptive AI)", tmpl.TitleTh)
	}

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
			thExisting := normalizeAnyNestedStringMap(normalizeAnyStringMap(normalizeAnyStringMap(s.Translations)["th"])["playbook_steps"])
			for _, rawStep := range normalizePhaseStepSlice(existingPB[phase]) {
				label, _ := rawStep["label"].(string)
				trimmed := strings.TrimSpace(label)
				if trimmed == "" || IsBannedBoilerplateLabel(trimmed) {
					continue
				}
				stepID, _ := rawStep["id"].(string)
				isCorrect, _ := rawStep["is_correct"].(bool)
				explanation, _ := rawStep["explanation"].(string)
				preserved = append(preserved, aiResponseActionOption{
					Label:         trimmed,
					LabelTh:       thExisting[stepID]["label"],
					IsCorrect:     isCorrect,
					Explanation:   explanation,
					ExplanationTh: thExisting[stepID]["explanation"],
				})
			}
		}

		steps, thSteps := buildPhaseStepsWithDistractorsWithTh(phase, phasePrefixes[phase], domainID, "", "", preserved, ai)
		newPB[phase] = steps
		if len(thSteps) > 0 {
			if transMap := normalizeAnyStringMap(s.Translations); transMap != nil {
				if thMap := normalizeAnyStringMap(transMap["th"]); thMap != nil {
					mergedPB := normalizeAnyStringMap(thMap["playbook_steps"])
					if mergedPB == nil {
						mergedPB = make(map[string]interface{})
					}
					for k, v := range thSteps {
						mergedPB[k] = v
					}
					thMap["playbook_steps"] = mergedPB
				}
			}
		}
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
