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
	"sync"
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

type scenarioStoryBeats struct {
	RoleContext     string `json:"role_context"`
	CurrentActivity string `json:"current_activity"`
	WhatHappened    string `json:"what_happened"`
	PotentialImpact string `json:"potential_impact"`
}

type aiMicroScenario struct {
	Title               string                   `json:"title"`
	TitleTh             string                   `json:"title_th,omitempty"`
	Description         string                   `json:"description"`
	DescriptionTh       string                   `json:"description_th,omitempty"`
	Story               *scenarioStoryBeats      `json:"story,omitempty"`
	StoryTh             *scenarioStoryBeats      `json:"story_th,omitempty"`
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
	"siem_escalation_ticket": "Open Beat 1 & 2 (`role_context`, `current_activity`) as a SOC Analyst on a monitoring shift working through the SIEM alert queue when a high-priority ticket flags unusual system behavior.",
	"shift_handover_brief":   "Open Beat 1 & 2 (`role_context`, `current_activity`) as a SOC Analyst who just started a shift and is reviewing outgoing handover notes when an unfolding anomaly catches your attention.",
	"threat_hunting_lead":    "Open Beat 1 & 2 (`role_context`, `current_activity`) as a SOC Analyst conducting a scheduled baseline telemetry health check when an unusual pattern stands out from normal activity.",
	"helpdesk_user_report":   "Open Beat 1 & 2 (`role_context`, `current_activity`) as a SOC Analyst assisting IT Support with an escalated user or asset-owner report when initial logs corroborate unexpected behavior.",
	"cloud_guardrail_alert":  "Open Beat 1 & 2 (`role_context`, `current_activity`) as a SOC Analyst monitoring cloud deployments and compliance dashboards alongside DevOps when an automated guardrail fires.",
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

type llmClientConfig struct {
	URL    string
	APIKey string
	Model  string
}

func resolveLLMClientConfig() (llmClientConfig, error) {
	rawApiKey := os.Getenv("LLM_API_KEY")
	if rawApiKey == "" {
		return llmClientConfig{}, fmt.Errorf("LLM_API_KEY is not configured")
	}
	cleanKey := strings.TrimSpace(strings.Trim(rawApiKey, "\"'"))
	if cleanKey == "" {
		return llmClientConfig{}, fmt.Errorf("LLM_API_KEY is not configured")
	}

	baseURL := os.Getenv("LLM_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	baseURL = strings.TrimSpace(strings.Trim(baseURL, "\"'"))
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/chat/completions", baseURL)
	if strings.HasSuffix(baseURL, "/chat/completions") {
		url = baseURL
	}

	model := os.Getenv("LLM_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}
	model = strings.TrimSpace(strings.Trim(model, "\"'"))

	return llmClientConfig{
		URL:    url,
		APIKey: cleanKey,
		Model:  model,
	}, nil
}

// GenerateLLMScenarioWithContext calls an OpenAI-compatible API using rotated SY0-701 sub-topic,
// archetype, briefing style, TP/FP cadence, and session history exclusions.
func GenerateLLMScenarioWithContext(ctx context.Context, weakestDomain string, domainName string, proficiency float64, profiles []models.UserSkillProfile, divCtx ScenarioDiversityContext) (*models.Scenario, error) {
	cfg, err := resolveLLMClientConfig()
	if err != nil {
		return nil, err
	}
	cleanKey, url, model := cfg.APIKey, cfg.URL, cfg.Model

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

	systemPrompt := "You are a specialized CompTIA Security+ 701 bilingual (English + Thai) cybersecurity scenario generator. Output ONLY raw JSON. Every scenario MUST author: (1) a smooth, friendly 4-beat narrative story (`story` in English and `story_th` in Thai) written in second-person voice ('You...' / 'คุณ...') positioning the learner as a SOC Analyst (explaining infrastructure/protocols plainly inline, preserving key host/IP/user tokens verbatim in English, and framing Balanced Dual-Stakes impact without spoiling the TP/FP verdict), (2) a realistic 4-to-8 event English-only incident_logs chain matching its domain and environment (never generic web/CORR-9901 boilerplate, zero Thai characters in incident_logs), (3) full Hybrid Bilingual Thai translations (*_th fields), and (4) 3 realistic SOC playbook options per IR phase (1 correct, 2 plausible technical distractors, 15-30 words each referencing entities)."
	userPrompt := fmt.Sprintf(`Generate a unique SOC scenario for target domain %s (%s). Current proficiency: %.1f%%.

Target Scenario Blueprint:
- SY0-701 Sub-Topic: %s (%s)
- Operational Archetype: %s
- Required Verdict (is_true_positive): %s
- Narrative Briefing Style (%s): %s

Rules for 4-Beat Incident Storytelling (story & story_th):
1. Provide BOTH "story" (English) and "story_th" (Thai) objects with exact keys: role_context, current_activity, what_happened, potential_impact. (The backend composes display descriptions automatically from these 4 fields.)
2. Second-Person Voice & Persona: Address the learner directly ("You..." in English / "คุณ..." in Thai) as an organizational SOC Analyst in a believable setting (matching the briefing_style above). Never start abruptly with raw technical jargon without introducing your role and ongoing task.
3. Friendly Guided Prose (No Tool Commands / No Bullets): Write naturally like a story for human readers (do NOT use beat labels/numbers inside the text, do NOT append "INVESTIGATION CLUES:", and do NOT instruct the learner to run specific EDR/firewall commands in the narrative). Briefly explain specialized systems/protocols inline in plain terms while keeping primary target_host, attacker_ip, and primary_user verbatim in English inside "what_happened" (both EN and TH).
4. Balanced Dual-Stakes in "potential_impact" (Anti-Spoiler): ALWAYS describe BOTH conditional possibilities evenly — (a) what harm could happen if this activity turns out to be a genuine attack, AND (b) what business/operational workflow could break if this is authorized legitimate activity that gets blocked prematurely. Never hint at whether is_true_positive is true or false.

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
  "story": {
    "role_context": "You are an on-duty SOC Analyst overseeing enterprise web infrastructure for an e-commerce company.",
    "current_activity": "While clearing the morning alert queue during a routine software release window, a priority notification lands on your dashboard.",
    "what_happened": "Security logs show deployment account svc_deploy logging into public storefront server SRV-WEB01 from external IP 185.220.101.45, followed by unexpected system file edits.",
    "potential_impact": "If this is a real intrusion, attackers could hijack customer checkouts or plant backdoors; however, if it is an emergency hotfix deployment, locking SRV-WEB01 immediately could take the store offline—verify the telemetry before acting."
  },
  "story_th": {
    "role_context": "เช้าวันทำงาน คุณปฏิบัติหน้าที่เป็น SOC Analyst ดูแลระบบความปลอดภัยให้กับโครงสร้างพื้นฐานเว็บของบริษัท",
    "current_activity": "ระหว่างที่กำลังตรวจคิวแจ้งเตือนประจำวันในช่วงเวลาปรับปรุงซอฟต์แวร์ตามรอบ ระบบแจ้งเตือนความเสี่ยงระดับสูงก็เด้งขึ้นมาให้คุณตรวจสอบ",
    "what_happened": "ระบบพบการใช้งานบัญชีจัดการอัตโนมัติ svc_deploy จาก IP ภายนอก 185.220.101.45 เข้ามาบนเซิร์ฟเวอร์หน้าร้าน SRV-WEB01 พร้อมมีการแก้ไขไฟล์ระบบสำคัญที่ไม่ได้แจ้งไว้ล่วงหน้า",
    "potential_impact": "หากเป็นการเจาะระบบจริง ผู้โจมตีอาจฝังช่องทางลอบเข้าสู่ระบบหรือขโมยข้อมูลลูกค้าได้ แต่หากเป็นการอัปเดตฉุกเฉินของทีมพัฒนา การสั่งตัดเซิร์ฟเวอร์ SRV-WEB01 ทันทีอาจทำให้บริการหน้าร้านหยุดชะงักได้ คุณจึงต้องตรวจ Log เพื่อยืนยันให้แน่ใจก่อน"
  },
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

	llmStart := time.Now()
	totalAttempts := 0
	recordOutcome := func(outcome string, err error) {
		reason := ""
		if err != nil {
			reason = strings.ReplaceAll(err.Error(), cleanKey, "[REDACTED]")
		}
		retries := 0
		if totalAttempts > 1 {
			retries = totalAttempts - 1
		}
		GlobalNFRTelemetry().RecordLLMCall(LLMCallRecord{
			DurationMs:     float64(time.Since(llmStart).Microseconds()) / 1000.0,
			Attempts:       totalAttempts,
			Retries:        retries,
			Outcome:        outcome,
			FallbackReason: reason,
			Provider:       url,
			Model:          model,
		})
	}

	rawJSON, attempts, err := callChatCompletionTracked(ctx, url, cleanKey, model, messages, 0.3)
	totalAttempts += attempts
	if err != nil {
		recordOutcome("error", err)
		return nil, err
	}

	micro, err := parseAIMicroScenario(rawJSON)
	if err != nil {
		recordOutcome("fallback", err)
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

		retryRaw, retryAttempts, retryErr := callChatCompletionTracked(ctx, url, cleanKey, model, retryMessages, 0.1)
		totalAttempts += retryAttempts
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
		valWrap := fmt.Errorf("scenario_validation_failed: %w", valErr)
		recordOutcome("fallback", valWrap)
		return nil, valWrap
	}

	recordOutcome("ai_success", nil)
	return scenario, nil
}

// callChatCompletion performs one OpenAI-compatible chat completion with network
// retries and returns the raw message content.
func callChatCompletion(ctx context.Context, url string, apiKey string, model string, messages []chatMessage, temperature float64) (string, error) {
	raw, _, err := callChatCompletionTracked(ctx, url, apiKey, model, messages, temperature)
	return raw, err
}

func callChatCompletionTracked(ctx context.Context, url string, apiKey string, model string, messages []chatMessage, temperature float64) (string, int, error) {
	reqBody := chatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temperature,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal LLM request: %w", err)
	}

	client := &http.Client{Timeout: 145 * time.Second}

	var resp *http.Response
	maxRetries := 2
	attempts := 0
	for attempt := 1; attempt <= maxRetries; attempt++ {
		attempts = attempt
		// Rebuild request for each attempt (body is consumed after first read)
		retryReq, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBytes))
		if reqErr != nil {
			return "", attempts, fmt.Errorf("failed to create HTTP request (attempt %d): %w", attempt, reqErr)
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
				return "", attempts, fmt.Errorf("LLM request failed after %d attempt(s): %w", attempt, ctx.Err())
			case <-time.After(3 * time.Second):
			}
		}
	}
	if err != nil {
		return "", attempts, fmt.Errorf("LLM request failed after %d attempts: %w", maxRetries, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("[LLM Error Details] URL: %s, Status: %d, Response: %s", url, resp.StatusCode, string(bodyBytes))
		return "", attempts, fmt.Errorf("LLM API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", attempts, fmt.Errorf("failed to decode LLM response: %w", err)
	}
	if len(chatResp.Choices) == 0 {
		return "", attempts, fmt.Errorf("LLM returned empty choices")
	}

	return stripJSONFences(strings.TrimSpace(chatResp.Choices[0].Message.Content)), attempts, nil
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
			Title:   "Unauthorized Security Policy Override Attempt",
			TitleTh: "ความพยายามแก้ไข Security Policy โดยไม่ได้รับอนุญาต",
			Story: &scenarioStoryBeats{
				RoleContext:     "You are an on-duty SOC Analyst responsible for monitoring identity and access controls across the corporate headquarters network.",
				CurrentActivity: "While reviewing incoming high-priority tickets in the SIEM queue during the late afternoon shift, a critical directory policy alert draws your focus.",
				WhatHappened:    "Audit telemetry shows service account SYSTEM_SVC connecting from internal workstation 192.168.1.150 and modifying the domain's main login controller (DC01.corp.internal) to disable password lockout protections outside the normal maintenance schedule.",
				PotentialImpact: "If this change is malicious, an attacker could freely run brute-force password guessing attacks across employee accounts without getting locked out; however, if an administrator triggered an emergency recovery adjustment, revoking SYSTEM_SVC prematurely could disrupt automated directory services—verify the logs first.",
			},
			StoryTh: &scenarioStoryBeats{
				RoleContext:     "ช่วงบ่ายวันทำงาน คุณปฏิบัติหน้าที่เป็น SOC Analyst ดูแลความปลอดภัยด้านสิทธิ์การเข้าถึงและระบบยืนยันตัวตนของสำนักงานใหญ่",
				CurrentActivity: "ระหว่างที่คุณกำลังไล่ตรวจสอบรายการแจ้งเตือนสำคัญบนหน้าจอ SIEM ประจำกะ ก็มี Ticket ระดับวิกฤตเด้งขึ้นมาให้ตรวจสอบทันที",
				WhatHappened:    "ระบบบันทึกเหตุการณ์พบว่าบัญชีบริการ SYSTEM_SVC จากเครื่องภายใน 192.168.1.150 เข้าไปปรับแก้นโยบายความปลอดภัยบนเครื่องแม่ข่ายหลักควบคุมสิทธิ์ (DC01.corp.internal) เพื่อปิดระบบล็อกบัญชีเมื่อใส่รหัสผ่านผิด นอกช่วงเวลาบำรุงรักษาปกติ",
				PotentialImpact: "หากเป็นการลักลอบแก้ไขโดยผู้โจมตี คนร้ายจะสามารถสุ่มเดารหัสผ่านบัญชีพนักงานได้ไม่จำกัดโดยไม่ถูกล็อก แต่หากเป็นการปรับค่าฉุกเฉินของผู้ดูแลระบบ การสั่งปิดบัญชี SYSTEM_SVC ทันทีอาจกระทบงานอัตโนมัติของระบบยืนยันตัวตนในองค์กรได้ คุณจึงต้องตรวจ Log เพื่อยืนยันความจริงก่อน",
			},
			Description:       "You are an on-duty SOC Analyst responsible for monitoring identity and access controls across the corporate headquarters network. While reviewing incoming high-priority tickets in the SIEM queue during the late afternoon shift, a critical directory policy alert draws your focus.\n\nAudit telemetry shows service account SYSTEM_SVC connecting from internal workstation 192.168.1.150 and modifying the domain's main login controller (DC01.corp.internal) to disable password lockout protections outside the normal maintenance schedule.\n\nIf this change is malicious, an attacker could freely run brute-force password guessing attacks across employee accounts without getting locked out; however, if an administrator triggered an emergency recovery adjustment, revoking SYSTEM_SVC prematurely could disrupt automated directory services—verify the logs first.",
			DescriptionTh:     "ช่วงบ่ายวันทำงาน คุณปฏิบัติหน้าที่เป็น SOC Analyst ดูแลความปลอดภัยด้านสิทธิ์การเข้าถึงและระบบยืนยันตัวตนของสำนักงานใหญ่ ระหว่างที่คุณกำลังไล่ตรวจสอบรายการแจ้งเตือนสำคัญบนหน้าจอ SIEM ประจำกะ ก็มี Ticket ระดับวิกฤตเด้งขึ้นมาให้ตรวจสอบทันที\n\nระบบบันทึกเหตุการณ์พบว่าบัญชีบริการ SYSTEM_SVC จากเครื่องภายใน 192.168.1.150 เข้าไปปรับแก้นโยบายความปลอดภัยบนเครื่องแม่ข่ายหลักควบคุมสิทธิ์ (DC01.corp.internal) เพื่อปิดระบบล็อกบัญชีเมื่อใส่รหัสผ่านผิด นอกช่วงเวลาบำรุงรักษาปกติ\n\nหากเป็นการลักลอบแก้ไขโดยผู้โจมตี คนร้ายจะสามารถสุ่มเดารหัสผ่านบัญชีพนักงานได้ไม่จำกัดโดยไม่ถูกล็อก แต่หากเป็นการปรับค่าฉุกเฉินของผู้ดูแลระบบ การสั่งปิดบัญชี SYSTEM_SVC ทันทีอาจกระทบงานอัตโนมัติของระบบยืนยันตัวตนในองค์กรได้ คุณจึงต้องตรวจ Log เพื่อยืนยันความจริงก่อน",
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
			Title:   "Suspicious Command Line Execution via Spearphishing Payload",
			TitleTh: "การรันคำสั่งต้องสงสัยผ่าน Spearphishing Payload",
			Story: &scenarioStoryBeats{
				RoleContext:     "You are a SOC Analyst collaborating with the corporate IT Support desk to triage escalated employee workstation reports.",
				CurrentActivity: "While reviewing a helpdesk ticket submitted after finance staff noticed their computer briefly stalling upon opening an email attachment, you correlate endpoint alerts on WS-FINANCE04.",
				WhatHappened:    "Security monitoring shows that right after employee efoster opened an invoice document on WS-FINANCE04, Microsoft Word silently launched a hidden automation script that began connecting out to unfamiliar external address 185.220.101.99.",
				PotentialImpact: "If this is a targeted phishing attack, attackers could gain remote control over WS-FINANCE04 and access sensitive accounting records; on the other hand, if it belongs to a benign macro reporting tool, cutting network access immediately could interrupt urgent financial closing tasks—review the telemetry first.",
			},
			StoryTh: &scenarioStoryBeats{
				RoleContext:     "เช้าวันทำงาน คุณปฏิบัติหน้าที่เป็น SOC Analyst ที่ทำงานประสานกับทีม IT Helpdesk เพื่อช่วยตรวจสอบแจ้งเหตุผิดปกติจากเครื่องพนักงาน",
				CurrentActivity: "ระหว่างที่คุณกำลังไล่เช็กเคสที่ทีม Helpdesk ส่งต่อมาหลังจากพนักงานฝ่ายการเงินแจ้งว่าเครื่องค้างชั่วครู่ตอนเปิดไฟล์แนบอีเมล คุณจึงเปิดดูข้อมูลเหตุการณ์บนเครื่อง WS-FINANCE04",
				WhatHappened:    "ระบบตรวจพบว่าทันทีที่ผู้ใช้ efoster เปิดเอกสารใบแจ้งหนี้บนเครื่อง WS-FINANCE04 โปรแกรม Word ได้เรียกใช้สคริปต์คำสั่งแบบซ่อนเบื้องหลัง พร้อมความพยายามเชื่อมต่อออกไปยัง IP ภายนอกที่ไม่รู้จัก 185.220.101.99",
				PotentialImpact: "หากเป็นการโจมตีผ่านอีเมลหลอกลวงจริง ผู้โจมตีอาจยึดควบคุมเครื่อง WS-FINANCE04 และเข้าถึงข้อมูลบัญชีการเงินสำคัญได้ แต่หากเป็นไฟล์รายงานอัตโนมัติขององค์กร การตัดเครื่องออกจากเครือข่ายทันทีอาจทำให้งานปิดงบการเงินสะดุดได้ คุณจึงควรตรวจหลักฐานใน Log ให้แน่ชัดก่อน",
			},
			Description:       "You are a SOC Analyst collaborating with the corporate IT Support desk to triage escalated employee workstation reports. While reviewing a helpdesk ticket submitted after finance staff noticed their computer briefly stalling upon opening an email attachment, you correlate endpoint alerts on WS-FINANCE04.\n\nSecurity monitoring shows that right after employee efoster opened an invoice document on WS-FINANCE04, Microsoft Word silently launched a hidden automation script that began connecting out to unfamiliar external address 185.220.101.99.\n\nIf this is a targeted phishing attack, attackers could gain remote control over WS-FINANCE04 and access sensitive accounting records; on the other hand, if it belongs to a benign macro reporting tool, cutting network access immediately could interrupt urgent financial closing tasks—review the telemetry first.",
			DescriptionTh:     "เช้าวันทำงาน คุณปฏิบัติหน้าที่เป็น SOC Analyst ที่ทำงานประสานกับทีม IT Helpdesk เพื่อช่วยตรวจสอบแจ้งเหตุผิดปกติจากเครื่องพนักงาน ระหว่างที่คุณกำลังไล่เช็กเคสที่ทีม Helpdesk ส่งต่อมาหลังจากพนักงานฝ่ายการเงินแจ้งว่าเครื่องค้างชั่วครู่ตอนเปิดไฟล์แนบอีเมล คุณจึงเปิดดูข้อมูลเหตุการณ์บนเครื่อง WS-FINANCE04\n\nระบบตรวจพบว่าทันทีที่ผู้ใช้ efoster เปิดเอกสารใบแจ้งหนี้บนเครื่อง WS-FINANCE04 โปรแกรม Word ได้เรียกใช้สคริปต์คำสั่งแบบซ่อนเบื้องหลัง พร้อมความพยายามเชื่อมต่อออกไปยัง IP ภายนอกที่ไม่รู้จัก 185.220.101.99\n\nหากเป็นการโจมตีผ่านอีเมลหลอกลวงจริง ผู้โจมตีอาจยึดควบคุมเครื่อง WS-FINANCE04 และเข้าถึงข้อมูลบัญชีการเงินสำคัญได้ แต่หากเป็นไฟล์รายงานอัตโนมัติขององค์กร การตัดเครื่องออกจากเครือข่ายทันทีอาจทำให้งานปิดงบการเงินสะดุดได้ คุณจึงควรตรวจหลักฐานใน Log ให้แน่ชัดก่อน",
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
			Title:   "Cloud Perimeter Firewall Misconfiguration",
			TitleTh: "การตั้งค่า Firewall ขอบเขตระบบคลาวด์ผิดพลาด (Public SSH)",
			Story: &scenarioStoryBeats{
				RoleContext:     "You are a SOC Analyst responsible for monitoring cloud infrastructure and compliance dashboards across production workloads.",
				CurrentActivity: "While watching automated security guardrails during a morning infrastructure deployment window, a high-severity network exposure alert fires.",
				WhatHappened:    "Cloud firewall configuration rules on primary database host AWS-PROD-DB01 were changed to allow remote login traffic from anywhere on the public internet, followed quickly by remote login activity under ec2-user from external IP 203.0.113.88.",
				PotentialImpact: "If this remote access is unauthorized, an attacker could extract confidential customer records or alter backend databases; however, if DevOps opened temporary troubleshooting access for an active outage, locking the rule blindly could break urgent database recovery work—inspect the timeline first.",
			},
			StoryTh: &scenarioStoryBeats{
				RoleContext:     "ช่วงเช้าในห้องปฏิบัติการความปลอดภัย คุณทำหน้าที่เป็น SOC Analyst ดูแลระบบแจ้งเตือนและมาตรฐานความปลอดภัยบนโครงสร้างพื้นฐานคลาวด์ของบริษัท",
				CurrentActivity: "ระหว่างที่กำลังเฝ้าดูแผงควบคุมแจ้งเตือนความปลอดภัยอัตโนมัติ (Cloud Guardrail) ช่วงที่มีการปรับระบบงานประจำวัน ระบบก็เตือนความเสี่ยงระดับสูงเข้ามา",
				WhatHappened:    "กฎความปลอดภัยหน้าด่านของเซิร์ฟเวอร์ฐานข้อมูลหลัก (AWS-PROD-DB01) ถูกเปลี่ยนให้เปิดรับการเชื่อมต่อระยะไกลจากอินเทอร์เน็ตสาธารณะทั้งหมด และไม่นานนักก็มีการล็อกอินเข้าใช้งานบัญชี ec2-user จาก IP ภายนอก 203.0.113.88",
				PotentialImpact: "หากเป็นการเจาะเข้ามาโดยไม่ได้รับอนุญาต ผู้โจมตีอาจเข้าถึงหรือดูดข้อมูลฐานข้อมูลสำคัญของลูกค้าออกไปได้ แต่หากเป็นทีม DevOps เปิดช่องทางฉุกเฉินเพื่อแก้ไขระบบงานจริงที่กำลังมีปัญหา การสั่งปิดกฎทันทีอาจขัดขวางการกู้ระบบได้ คุณจึงต้องตรวจ Log ยืนยันก่อนตัดสินใจ",
			},
			Description:       "You are a SOC Analyst responsible for monitoring cloud infrastructure and compliance dashboards across production workloads. While watching automated security guardrails during a morning infrastructure deployment window, a high-severity network exposure alert fires.\n\nCloud firewall configuration rules on primary database host AWS-PROD-DB01 were changed to allow remote login traffic from anywhere on the public internet, followed quickly by remote login activity under ec2-user from external IP 203.0.113.88.\n\nIf this remote access is unauthorized, an attacker could extract confidential customer records or alter backend databases; however, if DevOps opened temporary troubleshooting access for an active outage, locking the rule blindly could break urgent database recovery work—inspect the timeline first.",
			DescriptionTh:     "ช่วงเช้าในห้องปฏิบัติการความปลอดภัย คุณทำหน้าที่เป็น SOC Analyst ดูแลระบบแจ้งเตือนและมาตรฐานความปลอดภัยบนโครงสร้างพื้นฐานคลาวด์ของบริษัท ระหว่างที่กำลังเฝ้าดูแผงควบคุมแจ้งเตือนความปลอดภัยอัตโนมัติ (Cloud Guardrail) ช่วงที่มีการปรับระบบงานประจำวัน ระบบก็เตือนความเสี่ยงระดับสูงเข้ามา\n\nกฎความปลอดภัยหน้าด่านของเซิร์ฟเวอร์ฐานข้อมูลหลัก (AWS-PROD-DB01) ถูกเปลี่ยนให้เปิดรับการเชื่อมต่อระยะไกลจากอินเทอร์เน็ตสาธารณะทั้งหมด และไม่นานนักก็มีการล็อกอินเข้าใช้งานบัญชี ec2-user จาก IP ภายนอก 203.0.113.88\n\nหากเป็นการเจาะเข้ามาโดยไม่ได้รับอนุญาต ผู้โจมตีอาจเข้าถึงหรือดูดข้อมูลฐานข้อมูลสำคัญของลูกค้าออกไปได้ แต่หากเป็นทีม DevOps เปิดช่องทางฉุกเฉินเพื่อแก้ไขระบบงานจริงที่กำลังมีปัญหา การสั่งปิดกฎทันทีอาจขัดขวางการกู้ระบบได้ คุณจึงต้องตรวจ Log ยืนยันก่อนตัดสินใจ",
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
			Title:   "Suspicious Outbound C2 Beaconing Activity",
			TitleTh: "พบพฤติกรรมการส่งสัญญาณ C2 Beaconing ออกภายนอกต้องสงสัย",
			Story: &scenarioStoryBeats{
				RoleContext:     "You are a SOC Analyst monitoring corporate software engineering endpoints and outbound internet gateways.",
				CurrentActivity: "While running a routine weekly baseline health review across developer machines to spot abnormal network patterns, one repeated traffic rhythm catches your eye.",
				WhatHappened:    "Monitoring telemetry indicates developer workstation WS-DEV09 (logged in as dev_mholt) has begun sending automated check-in signals every 60 seconds to unrecognized external server 198.51.100.42 via an unsigned background service.",
				PotentialImpact: "If this traffic comes from remote attacker malware, the adversary could harvest developer passwords or pivot into source-code servers; however, if dev_mholt is running an internal sandbox test script, isolating WS-DEV09 prematurely could disrupt an active release build—verify the logs first.",
			},
			StoryTh: &scenarioStoryBeats{
				RoleContext:     "ช่วงบ่ายในศูนย์ปฏิบัติการความปลอดภัย คุณทำหน้าที่เป็น SOC Analyst ดูแลความปลอดภัยเครื่องคอมพิวเตอร์ทีมวิศวกรซอฟต์แวร์และประตูเชื่อมต่ออินเทอร์เน็ตขาออกของบริษัท",
				CurrentActivity: "ระหว่างที่คุณกำลังไล่ตรวจสอบพฤติกรรมเครือข่ายประจำสัปดาห์เพื่อค้นหาความผิดปกติจากค่ามาตรฐาน (Threat Hunting) รูปแบบการส่งข้อมูลเป็นจังหวะสม่ำเสมอรายการหนึ่งก็สะดุดตาขึ้นมา",
				WhatHappened:    "ข้อมูลตรวจจับแสดงว่าเครื่องนักพัฒนา WS-DEV09 ของผู้ใช้ dev_mholt มีโปรแกรมเบื้องหลังที่ไม่ได้ลงลายเซ็นดิจิทัลส่งสัญญาณติดต่อออกไปหาเซิร์ฟเวอร์ภายนอกที่ไม่มีประวัติ 198.51.100.42 ทุกๆ 60 วินาที",
				PotentialImpact: "หากเป็นโปรแกรมควบคุมระยะไกลของคนร้าย ผู้โจมตีอาจดึงรหัสผ่านหรือเจาะต่อเข้าสู่ระบบซอร์สโค้ดของบริษัทได้ แต่หาก dev_mholt กำลังทดสอบสคริปต์งานในเครื่องจริง การสั่งตัดเครือข่าย WS-DEV09 ทันทีอาจทำให้งาน build ซอฟต์แวร์สำคัญสะดุดได้ คุณจึงควรตรวจ Log ก่อนดำเนินการ",
			},
			Description:       "You are a SOC Analyst monitoring corporate software engineering endpoints and outbound internet gateways. While running a routine weekly baseline health review across developer machines to spot abnormal network patterns, one repeated traffic rhythm catches your eye.\n\nMonitoring telemetry indicates developer workstation WS-DEV09 (logged in as dev_mholt) has begun sending automated check-in signals every 60 seconds to unrecognized external server 198.51.100.42 via an unsigned background service.\n\nIf this traffic comes from remote attacker malware, the adversary could harvest developer passwords or pivot into source-code servers; however, if dev_mholt is running an internal sandbox test script, isolating WS-DEV09 prematurely could disrupt an active release build—verify the logs first.",
			DescriptionTh:     "ช่วงบ่ายในศูนย์ปฏิบัติการความปลอดภัย คุณทำหน้าที่เป็น SOC Analyst ดูแลความปลอดภัยเครื่องคอมพิวเตอร์ทีมวิศวกรซอฟต์แวร์และประตูเชื่อมต่ออินเทอร์เน็ตขาออกของบริษัท ระหว่างที่คุณกำลังไล่ตรวจสอบพฤติกรรมเครือข่ายประจำสัปดาห์เพื่อค้นหาความผิดปกติจากค่ามาตรฐาน (Threat Hunting) รูปแบบการส่งข้อมูลเป็นจังหวะสม่ำเสมอรายการหนึ่งก็สะดุดตาขึ้นมา\n\nข้อมูลตรวจจับแสดงว่าเครื่องนักพัฒนา WS-DEV09 ของผู้ใช้ dev_mholt มีโปรแกรมเบื้องหลังที่ไม่ได้ลงลายเซ็นดิจิทัลส่งสัญญาณติดต่อออกไปหาเซิร์ฟเวอร์ภายนอกที่ไม่มีประวัติ 198.51.100.42 ทุกๆ 60 วินาที\n\nหากเป็นโปรแกรมควบคุมระยะไกลของคนร้าย ผู้โจมตีอาจดึงรหัสผ่านหรือเจาะต่อเข้าสู่ระบบซอร์สโค้ดของบริษัทได้ แต่หาก dev_mholt กำลังทดสอบสคริปต์งานในเครื่องจริง การสั่งตัดเครือข่าย WS-DEV09 ทันทีอาจทำให้งาน build ซอฟต์แวร์สำคัญสะดุดได้ คุณจึงควรตรวจ Log ก่อนดำเนินการ",
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
			Title:   "Compliance Breach: Unencrypted Data Transfer",
			TitleTh: "การละเมิด Compliance: การส่งข้อมูลบัตรชำระเงินแบบไม่เข้ารหัส",
			Story: &scenarioStoryBeats{
				RoleContext:     "You are a SOC Analyst starting the morning shift for a retail payment platform required to follow strict PCI-DSS security controls.",
				CurrentActivity: "While reviewing the handover briefing from the overnight team along with open data-protection notifications, one settlement transfer warning stands out.",
				WhatHappened:    "Data loss sensors report that payment processing server SRV-PAYMENT01 (running under service account svc_checkout) began sending customer payment batches in unencrypted plain text to relay endpoint 10.0.2.77 after a recent deployment.",
				PotentialImpact: "If sensitive card numbers are leaving without encryption, anyone inspecting network traffic could capture payment data and trigger severe compliance violations; yet if 10.0.2.77 is an internal encrypted-tunnel test fixture, shutting down the checkout listener abruptly could halt customer transactions—verify the logs first.",
			},
			StoryTh: &scenarioStoryBeats{
				RoleContext:     "ช่วงเช้าตอนเปลี่ยนกะ คุณปฏิบัติหน้าที่เป็น SOC Analyst ดูแลความปลอดภัยระบบชำระเงินของบริษัทซึ่งอยู่ภายใต้มาตรฐานคุ้มครองข้อมูลบัตร PCI-DSS",
				CurrentActivity: "ระหว่างที่กำลังอ่านสรุปส่งมอบงานจากทีมกะดึกร่วมกับรายการแจ้งเตือนเฝ้าระวังข้อมูลรั่วไหลประจำวัน ระบบแจ้งเตือนข้อมูลชำระเงินรายการหนึ่งก็ปรากฏขึ้น",
				WhatHappened:    "ระบบตรวจสอบข้อมูลสำคัญรายงานว่าเซิร์ฟเวอร์ประมวลผลการชำระเงินหลัก (SRV-PAYMENT01) ภายใต้บัญชีบริการ svc_checkout มีการส่งชุดข้อมูลการชำระเงินแบบไม่เข้ารหัสผ่าน HTTP ออกไปยังรีเลย์ปลายทาง 10.0.2.77 หลังมีการปรับตั้งค่าระบบ",
				PotentialImpact: "หากข้อมูลบัตรลูกค้ากำลังถูกส่งออกโดยไม่เข้ารหัสจริง อาจถูกดักจับข้อมูลกลางทางและละเมิดข้อกำหนดกฎหมายรุนแรงได้ แต่หากปลายทาง 10.0.2.77 เป็นโหนดทดสอบภายในที่มีการเข้ารหัสชั้นนอกอยู่ การปิดช่องทางชำระเงินทันทีอาจทำให้ระบบตัดเงินลูกค้าขัดข้องได้ คุณจึงควรตรวจ Log ยืนยันก่อน",
			},
			Description:       "You are a SOC Analyst starting the morning shift for a retail payment platform required to follow strict PCI-DSS security controls. While reviewing the handover briefing from the overnight team along with open data-protection notifications, one settlement transfer warning stands out.\n\nData loss sensors report that payment processing server SRV-PAYMENT01 (running under service account svc_checkout) began sending customer payment batches in unencrypted plain text to relay endpoint 10.0.2.77 after a recent deployment.\n\nIf sensitive card numbers are leaving without encryption, anyone inspecting network traffic could capture payment data and trigger severe compliance violations; yet if 10.0.2.77 is an internal encrypted-tunnel test fixture, shutting down the checkout listener abruptly could halt customer transactions—verify the logs first.",
			DescriptionTh:     "ช่วงเช้าตอนเปลี่ยนกะ คุณปฏิบัติหน้าที่เป็น SOC Analyst ดูแลความปลอดภัยระบบชำระเงินของบริษัทซึ่งอยู่ภายใต้มาตรฐานคุ้มครองข้อมูลบัตร PCI-DSS ระหว่างที่กำลังอ่านสรุปส่งมอบงานจากทีมกะดึกร่วมกับรายการแจ้งเตือนเฝ้าระวังข้อมูลรั่วไหลประจำวัน ระบบแจ้งเตือนข้อมูลชำระเงินรายการหนึ่งก็ปรากฏขึ้น\n\nระบบตรวจสอบข้อมูลสำคัญรายงานว่าเซิร์ฟเวอร์ประมวลผลการชำระเงินหลัก (SRV-PAYMENT01) ภายใต้บัญชีบริการ svc_checkout มีการส่งชุดข้อมูลการชำระเงินแบบไม่เข้ารหัสผ่าน HTTP ออกไปยังรีเลย์ปลายทาง 10.0.2.77 หลังมีการปรับตั้งค่าระบบ\n\nหากข้อมูลบัตรลูกค้ากำลังถูกส่งออกโดยไม่เข้ารหัสจริง อาจถูกดักจับข้อมูลกลางทางและละเมิดข้อกำหนดกฎหมายรุนแรงได้ แต่หากปลายทาง 10.0.2.77 เป็นโหนดทดสอบภายในที่มีการเข้ารหัสชั้นนอกอยู่ การปิดช่องทางชำระเงินทันทีอาจทำให้ระบบตัดเงินลูกค้าขัดข้องได้ คุณจึงควรตรวจ Log ยืนยันก่อน",
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

type scenarioPlaybookStepTranslation struct {
	Label       string `json:"label"`
	Explanation string `json:"explanation"`
}

type scenarioFindingTranslation struct {
	Description string `json:"description"`
	Explanation string `json:"explanation"`
}

type scenarioTranslationPayloadTh struct {
	Title           string                                     `json:"title"`
	Description     string                                     `json:"description"`
	TpFpExplanation string                                     `json:"tp_fp_explanation"`
	PlaybookSteps   map[string]scenarioPlaybookStepTranslation `json:"playbook_steps"`
	Findings        map[string]scenarioFindingTranslation      `json:"findings"`
}

// TranslateScenarioToThai translates an existing English scenario's learner-facing narrative,
// triage explanation, playbook steps, and key findings into Hybrid Bilingual Thai without altering
// canonical English logs, IDs, or scoring rubrics.
func TranslateScenarioToThai(ctx context.Context, scenario *models.Scenario) (map[string]any, error) {
	cfg, err := resolveLLMClientConfig()
	if err != nil {
		return nil, err
	}
	return translateScenarioToThaiWithCaller(ctx, scenario, func(callCtx context.Context, messages []chatMessage, temperature float64) (string, error) {
		return callChatCompletion(callCtx, cfg.URL, cfg.APIKey, cfg.Model, messages, temperature)
	})
}

func translateScenarioToThaiWithCaller(
	ctx context.Context,
	scenario *models.Scenario,
	callFn func(ctx context.Context, messages []chatMessage, temperature float64) (string, error),
) (map[string]any, error) {
	if scenario == nil {
		return nil, fmt.Errorf("scenario is nil")
	}
	if callFn == nil {
		return nil, fmt.Errorf("translation caller is nil")
	}
	messages, err := buildScenarioTranslationMessages(scenario)
	if err != nil {
		return nil, err
	}
	rawJSON, err := callFn(ctx, messages, 0.2)
	if err != nil {
		return nil, err
	}
	return parseAndValidateScenarioTranslationTh(rawJSON, scenario)
}

func extractScenarioTranslationSources(
	scenario *models.Scenario,
) (string, map[string]scenarioPlaybookStepTranslation, map[string]scenarioFindingTranslation, error) {
	if scenario == nil {
		return "", nil, nil, fmt.Errorf("scenario is nil")
	}

	playbookSource := make(map[string]scenarioPlaybookStepTranslation)
	playbook := normalizeAnyStringMap(scenario.PlaybookSteps)
	for _, phase := range []string{"containment", "eradication", "recovery"} {
		for _, step := range normalizePhaseStepSlice(playbook[phase]) {
			id, _ := step["id"].(string)
			id = strings.TrimSpace(id)
			if id == "" {
				return "", nil, nil, fmt.Errorf("scenario playbook step in phase %q is missing id", phase)
			}
			label, _ := step["label"].(string)
			exp, _ := step["explanation"].(string)
			if strings.TrimSpace(exp) == "" {
				exp = fmt.Sprintf("Explain clearly in Thai whether %s action %q is appropriate or flawed for this %s phase.", id, strings.TrimSpace(label), phase)
			}
			playbookSource[id] = scenarioPlaybookStepTranslation{
				Label:       strings.TrimSpace(label),
				Explanation: strings.TrimSpace(exp),
			}
		}
	}

	findingsSource := make(map[string]scenarioFindingTranslation)
	outcomes := normalizeAnyStringMap(scenario.ExpectedOutcomes)
	for _, f := range normalizePhaseStepSlice(outcomes["key_findings"]) {
		id, _ := f["id"].(string)
		id = strings.TrimSpace(id)
		if id == "" {
			return "", nil, nil, fmt.Errorf("scenario finding is missing id")
		}
		desc, _ := f["description"].(string)
		exp, _ := f["explanation"].(string)
		if strings.TrimSpace(exp) == "" {
			exp = fmt.Sprintf("Explain clearly in Thai why finding %q is significant.", strings.TrimSpace(desc))
		}
		findingsSource[id] = scenarioFindingTranslation{
			Description: strings.TrimSpace(desc),
			Explanation: strings.TrimSpace(exp),
		}
	}

	tpFpExp := strings.TrimSpace(scenario.TpFpExplanation)
	if tpFpExp == "" {
		if scenario.IsTruePositive {
			tpFpExp = "Explain concisely in Thai why this incident is a True Positive security threat."
		} else {
			tpFpExp = "Explain concisely in Thai why this incident is a False Positive benign operational event."
		}
	}

	return tpFpExp, playbookSource, findingsSource, nil
}

func buildScenarioTranslationMessages(scenario *models.Scenario) ([]chatMessage, error) {
	tpFpExp, playbookSource, findingsSource, err := extractScenarioTranslationSources(scenario)
	if err != nil {
		return nil, err
	}

	sourcePayload := scenarioTranslationPayloadTh{
		Title:           strings.TrimSpace(scenario.Title),
		Description:     strings.TrimSpace(scenario.Description),
		TpFpExplanation: tpFpExp,
		PlaybookSteps:   playbookSource,
		Findings:        findingsSource,
	}
	payloadBytes, err := json.MarshalIndent(sourcePayload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scenario translation source: %w", err)
	}

	systemPrompt := "You are a specialized CompTIA Security+ 701 bilingual cybersecurity translator (English to Thai). Output ONLY raw valid JSON matching the exact input schema (`title`, `description`, `tp_fp_explanation`, `playbook_steps`, `findings`) with identical action IDs and finding IDs. Follow Hybrid Bilingual Terminology: translate explanations, narratives, and action text into clear natural Thai while preserving hostnames (e.g., DMZ-HIST01), IP addresses, usernames, service accounts, firewall/rule names, scheduled tasks, filenames, registry paths, MITRE ATT&CK IDs, CVE IDs, CLI commands, and standard security terms (such as True Positive, False Positive, EDR, SIEM, NIDS, DMZ, VLAN, PLC, AD) verbatim in English. Every translated string MUST contain Thai characters and MUST NOT be empty or left entirely in English."
	userPrompt := fmt.Sprintf("Translate all fields of this SOC scenario into Hybrid Bilingual Thai. Preserve every key under `playbook_steps` and `findings` verbatim:\n\n%s", string(payloadBytes))

	return []chatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}, nil
}

func extractJSONObject(raw string) string {
	stripped := stripJSONFences(raw)
	if strings.HasPrefix(stripped, "{") && strings.HasSuffix(stripped, "}") {
		return stripped
	}
	start := strings.Index(stripped, "{")
	end := strings.LastIndex(stripped, "}")
	if start >= 0 && end > start {
		return strings.TrimSpace(stripped[start : end+1])
	}
	return stripped
}

func parseAndValidateScenarioTranslationTh(rawJSON string, scenario *models.Scenario) (map[string]any, error) {
	if scenario == nil {
		return nil, fmt.Errorf("scenario is nil")
	}
	var payload scenarioTranslationPayloadTh
	cleaned := extractJSONObject(rawJSON)
	if err := json.Unmarshal([]byte(cleaned), &payload); err != nil {
		return nil, fmt.Errorf("invalid_translation_json: %w", err)
	}

	playbookOut, err := validateTranslatedPlaybookMap(payload.PlaybookSteps)
	if err != nil {
		return nil, err
	}
	findingsOut, err := validateTranslatedFindingsMap(payload.Findings)
	if err != nil {
		return nil, err
	}

	thMap := map[string]interface{}{
		"title":             strings.TrimSpace(payload.Title),
		"description":       strings.TrimSpace(payload.Description),
		"tp_fp_explanation": strings.TrimSpace(payload.TpFpExplanation),
		"playbook_steps":    playbookOut,
		"findings":          findingsOut,
	}

	merged := make(map[string]any)
	for k, v := range normalizeAnyStringMap(scenario.Translations) {
		merged[k] = v
	}
	merged["th"] = thMap

	candidate := *scenario
	candidate.Translations = merged
	if err := ValidateScenarioTranslationsTh(&candidate); err != nil {
		return nil, err
	}

	return merged, nil
}

func validateTranslatedPlaybookMap(raw map[string]scenarioPlaybookStepTranslation) (map[string]interface{}, error) {
	out := make(map[string]interface{}, len(raw))
	for id, step := range raw {
		trimmedID := strings.TrimSpace(id)
		label := strings.TrimSpace(step.Label)
		exp := strings.TrimSpace(step.Explanation)
		if trimmedID == "" || label == "" || exp == "" {
			continue
		}
		if !ContainsThaiRune(label) {
			return nil, fmt.Errorf("translation_validation_failed: playbook_steps[%q].label must contain Thai text", trimmedID)
		}
		if !ContainsThaiRune(exp) {
			return nil, fmt.Errorf("translation_validation_failed: playbook_steps[%q].explanation must contain Thai text", trimmedID)
		}
		out[trimmedID] = map[string]interface{}{
			"label":       label,
			"explanation": exp,
		}
	}
	return out, nil
}

func validateTranslatedFindingsMap(raw map[string]scenarioFindingTranslation) (map[string]interface{}, error) {
	out := make(map[string]interface{}, len(raw))
	for id, f := range raw {
		trimmedID := strings.TrimSpace(id)
		desc := strings.TrimSpace(f.Description)
		exp := strings.TrimSpace(f.Explanation)
		if trimmedID == "" || desc == "" || exp == "" {
			continue
		}
		if !ContainsThaiRune(desc) {
			return nil, fmt.Errorf("translation_validation_failed: findings[%q].description must contain Thai text", trimmedID)
		}
		if !ContainsThaiRune(exp) {
			return nil, fmt.Errorf("translation_validation_failed: findings[%q].explanation must contain Thai text", trimmedID)
		}
		out[trimmedID] = map[string]interface{}{
			"description": desc,
			"explanation": exp,
		}
	}
	return out, nil
}

func hasAllStoryBeats(b *scenarioStoryBeats) bool {
	if b == nil {
		return false
	}
	return strings.TrimSpace(b.RoleContext) != "" &&
		strings.TrimSpace(b.CurrentActivity) != "" &&
		strings.TrimSpace(b.WhatHappened) != "" &&
		strings.TrimSpace(b.PotentialImpact) != ""
}

func hasAllThaiStoryBeats(b *scenarioStoryBeats) bool {
	if !hasAllStoryBeats(b) {
		return false
	}
	return ContainsThaiRune(b.RoleContext) &&
		ContainsThaiRune(b.CurrentActivity) &&
		ContainsThaiRune(b.WhatHappened) &&
		ContainsThaiRune(b.PotentialImpact)
}

// ComposeStoryDescription joins 4 structured story beats into 3 smooth reader-friendly
// paragraphs separated by blank lines without visual beat tags.
func ComposeStoryDescription(b *scenarioStoryBeats) string {
	if b == nil {
		return ""
	}
	p1 := strings.TrimSpace(strings.TrimSpace(b.RoleContext) + " " + strings.TrimSpace(b.CurrentActivity))
	p2 := strings.TrimSpace(b.WhatHappened)
	p3 := strings.TrimSpace(b.PotentialImpact)
	parts := make([]string, 0, 3)
	for _, p := range []string{p1, p2, p3} {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, "\n\n")
}

func storyBeatsToMap(b *scenarioStoryBeats) map[string]interface{} {
	if b == nil {
		return nil
	}
	return map[string]interface{}{
		"role_context":     strings.TrimSpace(b.RoleContext),
		"current_activity": strings.TrimSpace(b.CurrentActivity),
		"what_happened":    strings.TrimSpace(b.WhatHappened),
		"potential_impact": strings.TrimSpace(b.PotentialImpact),
	}
}

func extractStoryBeatsFromLocaleMap(localeMap map[string]interface{}) *scenarioStoryBeats {
	if len(localeMap) == 0 {
		return nil
	}
	rawStory, ok := localeMap["story"]
	if !ok || rawStory == nil {
		return nil
	}
	switch s := rawStory.(type) {
	case *scenarioStoryBeats:
		return s
	case scenarioStoryBeats:
		return &s
	case map[string]interface{}:
		role, _ := s["role_context"].(string)
		act, _ := s["current_activity"].(string)
		what, _ := s["what_happened"].(string)
		impact, _ := s["potential_impact"].(string)
		return &scenarioStoryBeats{
			RoleContext:     role,
			CurrentActivity: act,
			WhatHappened:    what,
			PotentialImpact: impact,
		}
	case map[string]string:
		return &scenarioStoryBeats{
			RoleContext:     s["role_context"],
			CurrentActivity: s["current_activity"],
			WhatHappened:    s["what_happened"],
			PotentialImpact: s["potential_impact"],
		}
	default:
		return nil
	}
}

// HasCompleteStory reports whether both translations.en.story and translations.th.story
// contain all 4 non-empty narrative beats (with Thai text in translations.th.story).
func HasCompleteStory(s *models.Scenario) bool {
	if s == nil {
		return false
	}
	trans := normalizeAnyStringMap(s.Translations)
	if len(trans) == 0 {
		return false
	}
	enStory := extractStoryBeatsFromLocaleMap(normalizeAnyStringMap(trans["en"]))
	thStory := extractStoryBeatsFromLocaleMap(normalizeAnyStringMap(trans["th"]))
	return hasAllStoryBeats(enStory) && hasAllThaiStoryBeats(thStory)
}

func ensureStoryEntityMentions(b *scenarioStoryBeats, targetHost, attackerIP string) {
	if b == nil {
		return
	}
	what := strings.TrimSpace(b.WhatHappened)
	host := strings.TrimSpace(targetHost)
	ip := strings.TrimSpace(attackerIP)
	if host != "" && !strings.Contains(strings.ToLower(what), strings.ToLower(host)) {
		what += fmt.Sprintf(" (%s)", host)
	}
	if ip != "" && !strings.Contains(strings.ToLower(what), strings.ToLower(ip)) {
		what += fmt.Sprintf(" [%s]", ip)
	}
	b.WhatHappened = what
}

// UpgradeScenarioStoryBilingual rewrites a legacy scenario into smooth 4-beat storytelling
// (both EN and TH) and concurrently backfills any missing Thai title/playbook/finding translations
// without touching canonical initial_logs, playbook options, findings rubrics, or status.
func UpgradeScenarioStoryBilingual(ctx context.Context, scenario *models.Scenario) (string, map[string]any, error) {
	cfg, err := resolveLLMClientConfig()
	if err != nil {
		return "", nil, err
	}
	return upgradeScenarioStoryWithCaller(ctx, scenario, func(callCtx context.Context, messages []chatMessage, temperature float64) (string, error) {
		return callChatCompletion(callCtx, cfg.URL, cfg.APIKey, cfg.Model, messages, temperature)
	})
}

type storyOnlyUpgradePayload struct {
	Story   scenarioStoryBeats `json:"story"`
	StoryTh scenarioStoryBeats `json:"story_th"`
	TitleTh string             `json:"title_th,omitempty"`
}

type playbookFindingsTranslationPayload struct {
	TpFpExplanation string                                     `json:"tp_fp_explanation"`
	PlaybookSteps   map[string]scenarioPlaybookStepTranslation `json:"playbook_steps"`
	Findings        map[string]scenarioFindingTranslation      `json:"findings"`
}

func needsNonNarrativeThaiTranslation(s *models.Scenario) bool {
	if s == nil {
		return true
	}
	thMap := normalizeAnyStringMap(normalizeAnyStringMap(s.Translations)["th"])
	if len(thMap) == 0 {
		return true
	}
	thTpFp, _ := thMap["tp_fp_explanation"].(string)
	if strings.TrimSpace(thTpFp) == "" || !ContainsThaiRune(thTpFp) {
		return true
	}
	thPlaybook := normalizeAnyNestedStringMap(thMap["playbook_steps"])
	playbook := normalizeAnyStringMap(s.PlaybookSteps)
	for _, phase := range []string{"containment", "eradication", "recovery"} {
		for _, step := range normalizePhaseStepSlice(playbook[phase]) {
			id, _ := step["id"].(string)
			entry := thPlaybook[strings.TrimSpace(id)]
			if strings.TrimSpace(entry["label"]) == "" || strings.TrimSpace(entry["explanation"]) == "" {
				return true
			}
		}
	}
	thFindings := normalizeAnyNestedStringMap(thMap["findings"])
	outcomes := normalizeAnyStringMap(s.ExpectedOutcomes)
	for _, f := range normalizePhaseStepSlice(outcomes["key_findings"]) {
		id, _ := f["id"].(string)
		entry := thFindings[strings.TrimSpace(id)]
		if strings.TrimSpace(entry["description"]) == "" || strings.TrimSpace(entry["explanation"]) == "" {
			return true
		}
	}
	return false
}

func upgradeScenarioStoryWithCaller(
	ctx context.Context,
	scenario *models.Scenario,
	callFn func(ctx context.Context, messages []chatMessage, temperature float64) (string, error),
) (string, map[string]any, error) {
	if scenario == nil {
		return "", nil, fmt.Errorf("scenario is nil")
	}
	if callFn == nil {
		return "", nil, fmt.Errorf("story upgrader caller is nil")
	}

	thMapExisting := normalizeAnyStringMap(normalizeAnyStringMap(scenario.Translations)["th"])
	existingTitleTh, _ := thMapExisting["title"].(string)
	needTitleTh := strings.TrimSpace(existingTitleTh) == "" || !ContainsThaiRune(existingTitleTh)
	needPlaybookFindings := needsNonNarrativeThaiTranslation(scenario)

	storyMsgs, err := buildStoryOnlyMessages(scenario, needTitleTh)
	if err != nil {
		return "", nil, err
	}

	var pbMsgs []chatMessage
	if needPlaybookFindings {
		pbMsgs, err = buildPlaybookFindingsMessages(scenario)
		if err != nil {
			return "", nil, err
		}
	}

	var (
		storyRaw string
		storyErr error
		pbRaw    string
		pbErr    error
		wg       sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		storyRaw, storyErr = callFn(ctx, storyMsgs, 0.25)
	}()

	if needPlaybookFindings {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pbRaw, pbErr = callFn(ctx, pbMsgs, 0.2)
		}()
	}
	wg.Wait()

	if storyErr != nil {
		return "", nil, storyErr
	}
	if pbErr != nil {
		return "", nil, pbErr
	}

	return parseAndMergeParallelStoryUpgrade(storyRaw, pbRaw, scenario, needTitleTh, needPlaybookFindings)
}

func buildStoryOnlyMessages(scenario *models.Scenario, needTitleTh bool) ([]chatMessage, error) {
	targetHost, attackerIP := ExtractPrimaryScenarioEntities(scenario)
	inputObj := map[string]interface{}{
		"title":              strings.TrimSpace(scenario.Title),
		"domain_id":          strings.TrimSpace(scenario.DomainID),
		"source_description": strings.TrimSpace(scenario.Description),
		"target_host":        targetHost,
		"attacker_ip":        attackerIP,
	}

	schemaHint := `{
  "story": {
    "role_context": "Who the learner is as an organizational SOC Analyst and shift/sector context",
    "current_activity": "What routine operational task you are performing right before the alert hits",
    "what_happened": "What was detected in plain friendly language keeping target_host and attacker_ip verbatim in English",
    "potential_impact": "Balanced Dual-Stakes: what breaks if malicious AND what business workflow breaks if legitimate work is blocked (no TP/FP spoilers)"
  },
  "story_th": {
    "role_context": "บทบาทของคุณในฐานะ SOC Analyst และบริบทขององค์กร/กะงาน",
    "current_activity": "สิ่งที่คุณกำลังทำอยู่ก่อนเกิดเหตุ",
    "what_happened": "สิ่งที่ตรวจพบด้วยภาษาที่เข้าใจง่าย คงชื่อ host, user และ IP หลักไว้เป็นภาษาอังกฤษ",
    "potential_impact": "ความเสี่ยงสองด้านอย่างสมดุล (หากโจมตีจริง vs หากเป็นการทำงานปกติแล้วสั่งปิดผิดพลาด) โดยไม่สปอยล์ TP/FP"
  }`
	if needTitleTh {
		schemaHint += `,
  "title_th": "ชื่อสถานการณ์ภาษาไทยแบบ Hybrid Bilingual"`
	}
	schemaHint += "\n}"

	payloadBytes, err := json.Marshal(inputObj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal story prompt input: %w", err)
	}

	systemPrompt := "You are a bilingual cybersecurity storytelling writer. Output ONLY raw valid JSON. Rewrite the scenario into a smooth, reader-friendly 4-beat story (`story` in English and `story_th` in Thai) in second-person voice ('You...' / 'คุณ...') positioning the learner as an organizational SOC Analyst. Explain technical components inline in plain words while preserving target_host, primary_user, and attacker_ip verbatim in English. Never include tool commands or bullet clues, and ALWAYS write Balanced Dual-Stakes impact without spoiling TP/FP."
	userPrompt := fmt.Sprintf("Rewrite this scenario narrative matching this exact schema:\n%s\n\nScenario:\n%s", schemaHint, string(payloadBytes))

	return []chatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}, nil
}

func buildPlaybookFindingsMessages(scenario *models.Scenario) ([]chatMessage, error) {
	tpFpExp, playbookSource, findingsSource, err := extractScenarioTranslationSources(scenario)
	if err != nil {
		return nil, err
	}

	inputObj := map[string]interface{}{
		"tp_fp_explanation": tpFpExp,
		"playbook_steps":    playbookSource,
		"findings":          findingsSource,
	}
	payloadBytes, err := json.Marshal(inputObj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal playbook/findings translation source: %w", err)
	}

	systemPrompt := "You are a CompTIA Security+ 701 cybersecurity translator (English to Thai). Output ONLY raw valid JSON matching the exact input schema (`tp_fp_explanation`, `playbook_steps`, `findings`) preserving every ID verbatim. Translate into Hybrid Bilingual Thai (natural Thai prose keeping hostnames, IPs, accounts, rule names, files, commands, and MITRE IDs verbatim in English). Keep translations concise and accurate; every label, description, and explanation must contain Thai text."
	userPrompt := fmt.Sprintf("Translate these triage explanations, playbook steps, and findings into Thai preserving every key verbatim:\n%s", string(payloadBytes))

	return []chatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}, nil
}

func parseAndMergeParallelStoryUpgrade(
	storyRaw string,
	pbRaw string,
	scenario *models.Scenario,
	needTitleTh bool,
	needPlaybookFindings bool,
) (string, map[string]any, error) {
	if scenario == nil {
		return "", nil, fmt.Errorf("scenario is nil")
	}

	var storyPayload storyOnlyUpgradePayload
	if err := json.Unmarshal([]byte(extractJSONObject(storyRaw)), &storyPayload); err != nil {
		return "", nil, fmt.Errorf("invalid_story_upgrade_json: %w", err)
	}
	if !hasAllStoryBeats(&storyPayload.Story) {
		return "", nil, fmt.Errorf("story_validation_failed: all 4 English story beats are required")
	}
	if !hasAllThaiStoryBeats(&storyPayload.StoryTh) {
		return "", nil, fmt.Errorf("story_validation_failed: all 4 Thai story beats must contain Thai text")
	}

	targetHost, attackerIP := ExtractPrimaryScenarioEntities(scenario)
	ensureStoryEntityMentions(&storyPayload.Story, targetHost, attackerIP)
	ensureStoryEntityMentions(&storyPayload.StoryTh, targetHost, attackerIP)

	enDesc := ComposeStoryDescription(&storyPayload.Story)
	thDesc := ComposeStoryDescription(&storyPayload.StoryTh)

	merged := make(map[string]any)
	for k, v := range normalizeAnyStringMap(scenario.Translations) {
		merged[k] = v
	}

	enMap := make(map[string]interface{})
	for k, v := range normalizeAnyStringMap(merged["en"]) {
		enMap[k] = v
	}
	enMap["description"] = enDesc
	enMap["story"] = storyBeatsToMap(&storyPayload.Story)
	merged["en"] = enMap

	thMap := make(map[string]interface{})
	for k, v := range normalizeAnyStringMap(merged["th"]) {
		thMap[k] = v
	}
	thMap["description"] = thDesc
	thMap["story"] = storyBeatsToMap(&storyPayload.StoryTh)
	if needTitleTh {
		thMap["title"] = strings.TrimSpace(storyPayload.TitleTh)
	}

	if needPlaybookFindings {
		var pbPayload playbookFindingsTranslationPayload
		if err := json.Unmarshal([]byte(extractJSONObject(pbRaw)), &pbPayload); err != nil {
			return "", nil, fmt.Errorf("invalid_playbook_translation_json: %w", err)
		}
		playbookOut, err := validateTranslatedPlaybookMap(pbPayload.PlaybookSteps)
		if err != nil {
			return "", nil, err
		}
		findingsOut, err := validateTranslatedFindingsMap(pbPayload.Findings)
		if err != nil {
			return "", nil, err
		}
		thMap["tp_fp_explanation"] = strings.TrimSpace(pbPayload.TpFpExplanation)
		thMap["playbook_steps"] = playbookOut
		thMap["findings"] = findingsOut
	}
	merged["th"] = thMap

	candidate := *scenario
	candidate.Description = enDesc
	candidate.Translations = merged
	if err := ValidateScenarioTranslationsTh(&candidate); err != nil {
		return "", nil, err
	}
	if !HasCompleteStory(&candidate) {
		return "", nil, fmt.Errorf("story_validation_failed: upgraded scenario lacks complete EN/TH story beats")
	}
	return enDesc, merged, nil
}
