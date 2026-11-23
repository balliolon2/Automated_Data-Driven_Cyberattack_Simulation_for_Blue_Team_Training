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
	// MaxLogMessageRepeats caps identical normalized messages across a scenario's logs.
	MaxLogMessageRepeats = 3
)

// BannedLegacyLogSubstrings lists legacy hardcoded log strings prohibited by Spec 10.
var BannedLegacyLogSubstrings = []string{
	"siem alert rule [corr-9901]: multi-stage incident correlation triggered",
	"routine network telemetry event ",
	"kerberos ticket tgt requested for user account user",
}

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

func normalizeLogMessage(message string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(message))), " ")
}

func containsBannedLegacyBoilerplate(logs []map[string]interface{}) (string, bool) {
	for _, entry := range logs {
		msg, _ := entry["message"].(string)
		norm := normalizeLogMessage(msg)
		for _, banned := range BannedLegacyLogSubstrings {
			if strings.Contains(norm, banned) {
				return banned, true
			}
		}
	}
	return "", false
}

func validateScenarioAntiRepetition(logs []map[string]interface{}) error {
	if banned, found := containsBannedLegacyBoilerplate(logs); found {
		return fmt.Errorf("content_validation_failed: log contains banned legacy boilerplate %q", banned)
	}
	counts := make(map[string]int, len(logs))
	for _, entry := range logs {
		msg, _ := entry["message"].(string)
		norm := normalizeLogMessage(msg)
		if norm == "" {
			continue
		}
		counts[norm]++
		if counts[norm] > MaxLogMessageRepeats {
			return fmt.Errorf("content_validation_failed: log message %q repeated %d times (max %d)", strings.TrimSpace(msg), counts[norm], MaxLogMessageRepeats)
		}
	}
	return nil
}

func scenarioSeedHash(ai aiMicroScenario, domainID string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(strings.Join([]string{
		ai.Title, ai.TargetHost, ai.AttackerIP, ai.SubTopicID, ai.BriefingStyle, domainID,
	}, "|")))
	return h.Sum64()
}

func severityLevelForName(severity string) int {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "critical":
		return 5
	case "high":
		return 4
	case "warning":
		return 3
	case "medium":
		return 2
	default:
		return 1
	}
}

func normalizeIncidentEvents(ai aiMicroScenario, domainID string) []map[string]interface{} {
	host := strings.TrimSpace(ai.TargetHost)
	if host == "" {
		host = "SRV-WEB01"
	}
	ip := strings.TrimSpace(ai.AttackerIP)
	if ip == "" {
		ip = "185.220.101.45"
	}
	user := strings.TrimSpace(ai.PrimaryUser)
	if user == "" {
		user = "svc_ops"
	}

	if len(ai.IncidentLogs) > 0 {
		limit := len(ai.IncidentLogs)
		if limit > 8 {
			limit = 8
		}
		events := make([]map[string]interface{}, 0, 8)
		seenMsgs := make(map[string]bool, limit)
		for i := 0; i < limit; i++ {
			ev := ai.IncidentLogs[i]
			evMsg := strings.TrimSpace(ev.Message)
			evSource := strings.TrimSpace(ev.Source)
			if evSource == "" {
				evSource = "edr"
			}
			evSeverity := strings.ToLower(strings.TrimSpace(ev.Severity))
			if evSeverity == "" {
				evSeverity = "warning"
			}
			evHost := strings.TrimSpace(ev.Host)
			if evHost == "" {
				evHost = host
			}
			evSrcIP := strings.TrimSpace(ev.SrcIP)
			if evSrcIP == "" {
				evSrcIP = ip
			}
			evUser := strings.TrimSpace(ev.User)
			if evUser == "" {
				evUser = user
			}
			row := map[string]interface{}{
				"source":         evSource,
				"severity":       evSeverity,
				"severity_level": severityLevelForName(evSeverity),
				"message":        evMsg,
				"src_ip":         evSrcIP,
				"user":           evUser,
				"host":           evHost,
			}
			if strings.TrimSpace(ev.DstIP) != "" {
				row["dst_ip"] = strings.TrimSpace(ev.DstIP)
			}
			if ev.RelativeOffsetSec > 0 {
				row["relative_offset_sec"] = ev.RelativeOffsetSec
			}
			if evMsg != "" {
				seenMsgs[normalizeLogMessage(evMsg)] = true
			}
			events = append(events, row)
		}
		if len(events) < 4 {
			attackMsg := strings.TrimSpace(ai.AttackLogMessage)
			if attackMsg == "" {
				attackMsg = fmt.Sprintf("Corroborating security telemetry observation on %s involving %s", host, ip)
			}
			var pad []map[string]interface{}
			if ai.IsTruePositive {
				pad = buildDomainContextualFallbackChain(domainID, host, ip, user, attackMsg)
			} else {
				pad = buildFalsePositiveFallbackChain(domainID, host, ip, user, attackMsg)
			}
			for _, candidate := range pad {
				if len(events) >= 4 {
					break
				}
				norm := normalizeLogMessage(fmt.Sprint(candidate["message"]))
				if norm == "" || seenMsgs[norm] {
					continue
				}
				seenMsgs[norm] = true
				events = append(events, candidate)
			}
		}
		return events
	}

	// Legacy single-message compatibility without CORR-9901 or static 5-block boilerplate.
	attackMsg := strings.TrimSpace(ai.AttackLogMessage)
	if attackMsg == "" {
		attackMsg = fmt.Sprintf("Anomalous security telemetry event observed on %s involving %s", host, ip)
	}

	if !ai.IsTruePositive {
		return buildFalsePositiveFallbackChain(domainID, host, ip, user, attackMsg)
	}

	return buildDomainContextualFallbackChain(domainID, host, ip, user, attackMsg)
}

func buildFalsePositiveFallbackChain(domainID, host, ip, user, attackMsg string) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"source": "nids", "severity": "medium", "severity_level": 2,
			"message": fmt.Sprintf("Threshold telemetry alert triggered between %s and %s on %s", ip, host, domainID),
			"src_ip":  ip, "dst_ip": "10.20.4.18", "user": user, "host": host,
		},
		{
			"source": "edr", "severity": "info", "severity_level": 1,
			"message": attackMsg,
			"src_ip":  ip, "dst_ip": "10.20.4.18", "user": user, "host": host,
		},
		{
			"source": "iam_directory", "severity": "info", "severity_level": 1,
			"message": fmt.Sprintf("Service principal session verified against authorized directory policy on %s for %s", host, user),
			"src_ip":  "10.20.4.18", "dst_ip": ip, "user": user, "host": host,
		},
		{
			"source": "proxy_gateway", "severity": "info", "severity_level": 1,
			"message": fmt.Sprintf("Verified TLS certificate and baseline signature for %s session from %s", host, ip),
			"src_ip":  "10.20.4.18", "dst_ip": ip, "user": user, "host": host,
		},
	}
}

func buildDomainContextualFallbackChain(domainID, host, ip, user, attackMsg string) []map[string]interface{} {
	switch domainID {
	case "domain1":
		return []map[string]interface{}{
			{"source": "ad_controller", "severity": "warning", "severity_level": 3, "message": fmt.Sprintf("Repeated privilege delegation query from %s targeting %s under %s", ip, host, user), "src_ip": ip, "user": user, "host": host},
			{"source": "edr", "severity": "high", "severity_level": 4, "message": attackMsg, "src_ip": ip, "user": user, "host": host},
			{"source": "windows_event", "severity": "high", "severity_level": 4, "message": fmt.Sprintf("Audit Event 4732: Privileged security group membership altered on %s by %s", host, user), "src_ip": ip, "user": user, "host": host},
			{"source": "okta", "severity": "critical", "severity_level": 5, "message": fmt.Sprintf("Session token elevation observed for %s from %s after policy change on %s", user, ip, host), "src_ip": ip, "user": user, "host": host},
		}
	case "domain3":
		return []map[string]interface{}{
			{"source": "aws_cloudtrail", "severity": "warning", "severity_level": 3, "message": fmt.Sprintf("Network ACL boundary modified on %s permitting ingress from %s", host, ip), "src_ip": ip, "user": user, "host": host},
			{"source": "edr", "severity": "high", "severity_level": 4, "message": attackMsg, "src_ip": ip, "user": user, "host": host},
			{"source": "s3_access", "severity": "high", "severity_level": 4, "message": fmt.Sprintf("Unrestricted internal zone enumeration traced from %s across %s", ip, host), "src_ip": ip, "user": user, "host": host},
			{"source": "k8s_audit", "severity": "critical", "severity_level": 5, "message": fmt.Sprintf("Cross-namespace secrets read request initiated on %s via %s", host, ip), "src_ip": ip, "user": user, "host": host},
		}
	case "domain5":
		return []map[string]interface{}{
			{"source": "compliance_scanner", "severity": "warning", "severity_level": 3, "message": fmt.Sprintf("Governance policy deviation flagged for %s involving transfer to %s", host, ip), "src_ip": ip, "user": user, "host": host},
			{"source": "dlp_sensor", "severity": "high", "severity_level": 4, "message": attackMsg, "src_ip": ip, "user": user, "host": host},
			{"source": "proxy_gateway", "severity": "high", "severity_level": 4, "message": fmt.Sprintf("Unapproved external transfer channel opened from %s to %s without DLP ticket", host, ip), "src_ip": "10.20.5.12", "dst_ip": ip, "user": user, "host": host},
			{"source": "change_management", "severity": "warning", "severity_level": 3, "message": fmt.Sprintf("No matching approved change ticket found for configuration drift on %s", host), "src_ip": ip, "user": user, "host": host},
		}
	default:
		return []map[string]interface{}{
			{"source": "firewall", "severity": "warning", "severity_level": 3, "message": fmt.Sprintf("Outbound session anomaly observed between %s and remote peer %s", host, ip), "src_ip": "10.20.4.55", "dst_ip": ip, "user": user, "host": host},
			{"source": "edr", "severity": "high", "severity_level": 4, "message": attackMsg, "src_ip": "10.20.4.55", "dst_ip": ip, "user": user, "host": host},
			{"source": "sysmon", "severity": "high", "severity_level": 4, "message": fmt.Sprintf("Suspicious child process tree spawned on %s communicating with %s", host, ip), "src_ip": "10.20.4.55", "dst_ip": ip, "user": user, "host": host},
			{"source": "nids", "severity": "critical", "severity_level": 5, "message": fmt.Sprintf("Lateral or C2 protocol anomaly confirmed between %s and %s", host, ip), "src_ip": ip, "dst_ip": "10.20.4.55", "user": user, "host": host},
		}
	}
}

func selectNoisePacks(domainID, subTopicID string, incidentEvents []map[string]interface{}) []string {
	packSet := make(map[string]bool)
	addPack := func(p string) { packSet[p] = true }

	switch strings.TrimSpace(subTopicID) {
	case "1.2", "4.6", "5.6":
		addPack("identity_directory")
	case "1.3", "5.1", "5.2", "5.3", "5.4", "5.5":
		addPack("governance_itsm")
	case "1.4", "3.1", "4.4", "4.5":
		addPack("network_perimeter")
	case "2.2", "2.4", "2.5", "4.8", "4.9":
		addPack("endpoint_workstation")
	case "2.3", "3.3":
		addPack("application_data")
	case "3.2", "3.4", "4.1", "4.2", "4.3", "4.7":
		addPack("cloud_infrastructure")
	}

	for _, ev := range incidentEvents {
		src := strings.ToLower(strings.TrimSpace(fmt.Sprint(ev["source"])))
		switch {
		case strings.Contains(src, "okta") || strings.Contains(src, "ad_") || strings.Contains(src, "vpn") || strings.Contains(src, "sso"):
			addPack("identity_directory")
		case strings.Contains(src, "aws") || strings.Contains(src, "s3") || strings.Contains(src, "k8s") || strings.Contains(src, "iam"):
			addPack("cloud_infrastructure")
		case strings.Contains(src, "edr") || strings.Contains(src, "sysmon") || strings.Contains(src, "file") || strings.Contains(src, "windows"):
			addPack("endpoint_workstation")
		case strings.Contains(src, "nginx") || strings.Contains(src, "web") || strings.Contains(src, "postgres") || strings.Contains(src, "email") || strings.Contains(src, "dlp") || strings.Contains(src, "crm"):
			addPack("application_data")
		case strings.Contains(src, "change") || strings.Contains(src, "hr") || strings.Contains(src, "itsm") || strings.Contains(src, "compliance") || strings.Contains(src, "backup"):
			addPack("governance_itsm")
		default:
			addPack("network_perimeter")
		}
	}

	domainDefaults := map[string][]string{
		"domain1": {"identity_directory", "governance_itsm"},
		"domain2": {"endpoint_workstation", "application_data"},
		"domain3": {"cloud_infrastructure", "network_perimeter"},
		"domain4": {"network_perimeter", "endpoint_workstation"},
		"domain5": {"governance_itsm", "application_data"},
	}
	for _, p := range domainDefaults[domainID] {
		addPack(p)
	}
	if len(packSet) < 2 {
		addPack("network_perimeter")
		addPack("identity_directory")
	}

	ordered := []string{"identity_directory", "endpoint_workstation", "cloud_infrastructure", "network_perimeter", "application_data", "governance_itsm"}
	out := make([]string, 0, len(packSet))
	for _, p := range ordered {
		if packSet[p] {
			out = append(out, p)
		}
	}
	if len(out) > 3 {
		return out[:3]
	}
	return out
}

func buildNoiseLogEntry(pack string, idx int, seed uint64, avoidHost, avoidIP, avoidUser string) map[string]interface{} {
	users := []string{"tchen", "lnguyen", "dpatel", "srodriguez", "bwilson", "kpark", "amartinez", "jkim", "cmiller", "rthomas", "npatel", "edavis", "hlee", "cwright"}
	hosts := []string{"FIN-WS-101", "ENG-MAC-204", "HR-LAP-309", "OPS-SRV-412", "MKT-WS-118", "DEV-LIN-221", "SUP-WS-105", "QA-NODE-330", "SEC-MON-04", "BUILD-AG-09", "APP-NODE-07", "AUTH-REP-02"}
	domains := []string{"api.github.com", "packages.ubuntu.com", "login.microsoftonline.com", "cdn.cloudflare.com", "slack.com", "registry.npmjs.org", "docs.corp.internal", "zoom.us"}

	n := idx + int(seed%97)
	u := users[n%len(users)]
	if strings.EqualFold(u, avoidUser) {
		u = users[(n+1)%len(users)]
	}
	h := hosts[(n*3)%len(hosts)]
	if strings.EqualFold(h, avoidHost) {
		h = hosts[(n*3+1)%len(hosts)]
	}
	ip := fmt.Sprintf("10.%d.%d.%d", 20+(n%5), 10+((n*3)%40), 15+((n*7)%200))
	if ip == avoidIP {
		ip = fmt.Sprintf("10.%d.%d.%d", 26+(n%4), 12+(n%30), 22+(n%180))
	}
	dom := domains[n%len(domains)]
	reqID := 5000 + n*17 + idx

	var source, message string
	switch pack {
	case "identity_directory":
		opts := [][2]string{
			{"ad_controller", fmt.Sprintf("Kerberos service ticket TGS issued for %s@corp.local on %s (kvno=%d)", u, h, (idx%4)+1)},
			{"okta", fmt.Sprintf("OAuth2 refresh token verified for %s@corp.local on trusted endpoint #%d", u, reqID)},
			{"vpn_gateway", fmt.Sprintf("Split-tunnel VPN keepalive acknowledged for %s from %s (session #%d)", u, ip, reqID)},
			{"sso_idp", fmt.Sprintf("SAML SSO assertion completed for %s accessing internal wiki portal", u)},
		}
		pick := opts[idx%len(opts)]
		source, message = pick[0], pick[1]
	case "endpoint_workstation":
		opts := [][2]string{
			{"edr", fmt.Sprintf("Signed system binary svchost.exe verified on %s (PID %d)", h, 2400+idx*19)},
			{"sysmon", fmt.Sprintf("Process create: EdgeUpdate.exe /ua on %s under %s (signature valid)", h, u)},
			{"windows_event", fmt.Sprintf("Scheduled Defender signature update KB2267602-v%d applied on %s", 1400+idx, h)},
			{"file_integrity", fmt.Sprintf("Baseline checksum verified on %s: /etc/systemd/system/node-exporter unchanged", h)},
		}
		pick := opts[idx%len(opts)]
		source, message = pick[0], pick[1]
	case "cloud_infrastructure":
		opts := [][2]string{
			{"aws_cloudtrail", fmt.Sprintf("AssumeRole completed for CI runner role-deploy-%d (RequestId=req-%d)", (idx%5)+1, reqID)},
			{"s3_access", fmt.Sprintf("S3 HeadObject on static-assets-cdn key=bundles/app-v%d.js HTTP 200", idx)},
			{"k8s_audit", fmt.Sprintf("EKS kubelet liveness check passed for pod api-worker-%d in prod", 10+idx)},
			{"iam_service", fmt.Sprintf("KMS key alias/app-secrets automated rotation health check passed (#%d)", reqID)},
		}
		pick := opts[idx%len(opts)]
		source, message = pick[0], pick[1]
	case "application_data":
		opts := [][2]string{
			{"nginx", fmt.Sprintf("GET /api/v2/health?node=%s HTTP/1.1 200 OK (%dms)", h, 8+(idx%15))},
			{"postgresql", fmt.Sprintf("Read-replica WAL checkpoint completed on db-replica-0%d (lag=3ms)", (idx%3)+1)},
			{"email_gateway", fmt.Sprintf("Inbound email SPF/DKIM/DMARC pass from alerts@%s to %s@corp.local", dom, u)},
			{"dlp_sensor", fmt.Sprintf("Outbound report summary-q%d-%d.pdf cleared DLP policy scan for %s", (idx%4)+1, reqID, u)},
		}
		pick := opts[idx%len(opts)]
		source, message = pick[0], pick[1]
	case "governance_itsm":
		opts := [][2]string{
			{"change_management", fmt.Sprintf("Routine ITSM ticket INC-%d closed: printer driver update verified for %s", reqID, u)},
			{"backup_service", fmt.Sprintf("Nightly incremental backup snap-%d finished cleanly on %s", reqID, h)},
			{"compliance_scanner", fmt.Sprintf("CIS benchmark posture check passed on %s: auditd active", h)},
			{"itsm_ticketing", fmt.Sprintf("Quarterly access attestation #AR-%d approved by manager for %s", reqID, u)},
		}
		pick := opts[idx%len(opts)]
		source, message = pick[0], pick[1]
	default:
		opts := [][2]string{
			{"dns_server", fmt.Sprintf("DNS Query %s Type=AAAA resolved from %s (%s)", dom, ip, h)},
			{"firewall", fmt.Sprintf("Outbound TLS 1.3 session closed cleanly: %s -> 198.18.%d.%d:443 [bytes=%d]", ip, (idx%8)+1, 20+idx, 2048+idx*512)},
			{"proxy_gateway", fmt.Sprintf("CONNECT %s:443 completed cleanly for %s on %s", dom, u, h)},
			{"dhcp_server", fmt.Sprintf("DHCP lease renewed cleanly for %s (%s) on workstation VLAN 20", h, ip)},
		}
		pick := opts[idx%len(opts)]
		source, message = pick[0], pick[1]
	}

	return map[string]interface{}{
		"source":         source,
		"severity":       "info",
		"severity_level": 1,
		"message":        message,
		"src_ip":         ip,
		"user":           u,
		"host":           h,
	}
}

func buildDecoyWarningLogs(count int, seed uint64, avoidHost, avoidIP, avoidUser string) []map[string]interface{} {
	tag := int(seed % 1000)
	candidates := []map[string]interface{}{
		{
			"source": "ad_controller", "severity": "warning", "severity_level": 2,
			"message": fmt.Sprintf("Single invalid password entry for kpark on FIN-WS-104 followed by valid MFA (#%d)", 1100+tag),
			"src_ip":  "10.21.14.44", "user": "kpark", "host": "FIN-WS-104",
		},
		{
			"source": "firewall", "severity": "medium", "severity_level": 2,
			"message": fmt.Sprintf("Perimeter firewall dropped unsolicited internet scan from 198.51.100.241 to port 23 (#%d)", 1200+tag),
			"src_ip":  "198.51.100.241", "dst_ip": "10.21.1.5", "user": "net_ops", "host": "FW-PERIM-01",
		},
		{
			"source": "web_app", "severity": "warning", "severity_level": 2,
			"message": fmt.Sprintf("Transient HTTP 429 rate-limit triggered by partner crawler on /public/status (#%d)", 1300+tag),
			"src_ip":  "10.22.5.33", "user": "svc_crawler", "host": "PROXY-EDGE-03",
		},
		{
			"source": "compliance_scanner", "severity": "warning", "severity_level": 2,
			"message": fmt.Sprintf("Internal TLS certificate on intranet-staging.corp.local expires in 14 days (#%d)", 1400+tag),
			"src_ip":  "10.23.2.12", "user": "cert_bot", "host": "SEC-MON-04",
		},
		{
			"source": "edr", "severity": "warning", "severity_level": 3,
			"message": fmt.Sprintf("Sandboxed EICAR validation signature quarantined on QA-NODE-330 during scheduled test (#%d)", 1500+tag),
			"src_ip":  "10.24.8.71", "user": "qa_runner", "host": "QA-NODE-330",
		},
	}

	out := make([]map[string]interface{}, 0, count)
	start := int(seed % uint64(len(candidates)))
	for i := 0; i < len(candidates) && len(out) < count; i++ {
		c := candidates[(start+i)%len(candidates)]
		ch, _ := c["host"].(string)
		cip, _ := c["src_ip"].(string)
		cu, _ := c["user"].(string)
		if strings.EqualFold(ch, avoidHost) || (avoidIP != "" && cip == avoidIP) || strings.EqualFold(cu, avoidUser) {
			continue
		}
		out = append(out, c)
	}
	return out
}

// buildScenarioLogs constructs a variable-length (28-46 row), chronologically interleaved
// SIEM timeline combining multi-event incident logs, decoy warnings, and environment noise.
func buildScenarioLogs(ai aiMicroScenario, now time.Time) []map[string]interface{} {
	return buildScenarioLogsWithDomain(ai, "domain4", now)
}

func buildScenarioLogsWithDomain(ai aiMicroScenario, domainID string, now time.Time) []map[string]interface{} {
	incidentEvents := normalizeIncidentEvents(ai, domainID)
	return assembleInterleavedScenarioLogs(ai, domainID, incidentEvents, now)
}

func assembleInterleavedScenarioLogs(ai aiMicroScenario, domainID string, incidentEvents []map[string]interface{}, now time.Time) []map[string]interface{} {
	seed := scenarioSeedHash(ai, domainID)
	decoyCount := 2 + int((seed>>8)%3) // 2..4 decoy warning/medium logs
	decoys := buildDecoyWarningLogs(decoyCount, seed, ai.TargetHost, ai.AttackerIP, ai.PrimaryUser)

	targetTotal := 28 + int(seed%19) // 28..46 rows
	minRequired := len(incidentEvents) + len(decoys) + 8
	if targetTotal < minRequired {
		targetTotal = minRequired
	}
	if targetTotal > 46 {
		targetTotal = 46
	}

	slots := make([]map[string]interface{}, targetTotal)
	used := make([]bool, targetTotal)

	// Spread incident events chronologically across 10%..90% of the timeline (non-contiguous).
	m := len(incidentEvents)
	if m > 0 {
		lowBound := targetTotal / 10
		if lowBound < 2 {
			lowBound = 2
		}
		highBound := (targetTotal * 9) / 10
		if highBound >= targetTotal {
			highBound = targetTotal - 1
		}
		span := highBound - lowBound
		lastPos := lowBound - 1
		for idx, ev := range incidentEvents {
			var pos int
			if m == 1 {
				pos = targetTotal / 2
			} else {
				pos = lowBound + (idx*span)/(m-1)
			}
			// Keep at least 1 noise row gap on early transitions so events don't form a solid block
			minGap := 1
			if idx == 1 && m >= 2 && targetTotal-lastPos > (m-idx)*2 {
				minGap = 2
			}
			if pos <= lastPos+minGap-1 {
				pos = lastPos + minGap
			}
			if pos >= targetTotal-(m-idx) {
				pos = targetTotal - (m - idx)
			}
			if pos < 0 {
				pos = 0
			}
			for pos < targetTotal && used[pos] {
				pos++
			}
			if pos < targetTotal {
				slots[pos] = ev
				used[pos] = true
				lastPos = pos
			}
		}
	}

	// Place decoys across early, middle, and late free slots.
	freeSlots := make([]int, 0, targetTotal)
	for i := 0; i < targetTotal; i++ {
		if !used[i] {
			freeSlots = append(freeSlots, i)
		}
	}
	if len(freeSlots) > 0 && len(decoys) > 0 {
		step := len(freeSlots) / (len(decoys) + 1)
		if step < 1 {
			step = 1
		}
		for dIdx, dec := range decoys {
			pick := (dIdx + 1) * step
			if pick >= len(freeSlots) {
				pick = len(freeSlots) - 1
			}
			for used[freeSlots[pick]] {
				pick = (pick + 1) % len(freeSlots)
			}
			pos := freeSlots[pick]
			slots[pos] = dec
			used[pos] = true
		}
	}

	// Fill remaining slots with unique environment-matched noise logs.
	packs := selectNoisePacks(domainID, ai.SubTopicID, incidentEvents)
	usedMsgs := make(map[string]bool, targetTotal)
	for _, entry := range slots {
		if entry != nil {
			if msg, _ := entry["message"].(string); msg != "" {
				usedMsgs[normalizeLogMessage(msg)] = true
			}
		}
	}

	noiseCounter := 0
	for i := 0; i < targetTotal; i++ {
		if slots[i] != nil {
			continue
		}
		pack := packs[(i+noiseCounter)%len(packs)]
		var entry map[string]interface{}
		for attempt := 0; attempt < 25; attempt++ {
			candidate := buildNoiseLogEntry(pack, noiseCounter+attempt*7, seed, ai.TargetHost, ai.AttackerIP, ai.PrimaryUser)
			norm := normalizeLogMessage(fmt.Sprint(candidate["message"]))
			if !usedMsgs[norm] {
				usedMsgs[norm] = true
				entry = candidate
				break
			}
		}
		if entry == nil {
			entry = buildNoiseLogEntry(pack, noiseCounter+1000+i, seed, ai.TargetHost, ai.AttackerIP, ai.PrimaryUser)
		}
		slots[i] = entry
		noiseCounter++
	}

	// Stamp monotonically increasing RFC3339 timestamps across all rows, honoring relative_offset_sec when larger.
	offsetSec := 0
	for i, entry := range slots {
		if i > 0 {
			jitter := 5 + int((seed+uint64(i*11))%23)
			offsetSec += jitter
		}
		if rel, ok := entry["relative_offset_sec"].(int); ok && rel > offsetSec {
			offsetSec = rel
		}
		delete(entry, "relative_offset_sec")
		entry["timestamp"] = now.Add(time.Duration(offsetSec) * time.Second).Format(time.RFC3339)
	}

	return slots
}

// normalizeAIFindings returns the explicit findings, or synthesizes one finding
// from the legacy key_finding / attack_log_message fields.
func normalizeAIFindings(ai aiMicroScenario) []aiFinding {
	return normalizeAIFindingsForDomain(ai, "domain4")
}

func normalizeAIFindingsForDomain(ai aiMicroScenario, domainID string) []aiFinding {
	if len(ai.Findings) > 0 {
		out := make([]aiFinding, len(ai.Findings))
		copy(out, ai.Findings)
		return out
	}

	description := strings.TrimSpace(ai.KeyFinding)
	descriptionTh := strings.TrimSpace(ai.KeyFindingTh)
	if description == "" {
		if ai.IsTruePositive {
			description = fmt.Sprintf("Identified malicious activity originating from %s targeting %s", ai.AttackerIP, ai.TargetHost)
			if descriptionTh == "" && strings.TrimSpace(ai.TitleTh) != "" {
				descriptionTh = fmt.Sprintf("ตรวจพบกิจกรรมที่เป็นอันตรายจากต้นทาง %s ซึ่งมุ่งเป้าไปยัง %s", ai.AttackerIP, ai.TargetHost)
			}
		} else {
			description = fmt.Sprintf("Verified the activity involving %s and %s is benign and authorized", ai.TargetHost, ai.AttackerIP)
			if descriptionTh == "" && strings.TrimSpace(ai.TitleTh) != "" {
				descriptionTh = fmt.Sprintf("ยืนยันแล้วว่ากิจกรรมระหว่าง %s และ %s มีความปลอดภัยและได้รับอนุญาต", ai.TargetHost, ai.AttackerIP)
			}
		}
	} else if descriptionTh == "" && strings.TrimSpace(ai.TitleTh) != "" {
		if ai.IsTruePositive {
			descriptionTh = fmt.Sprintf("ตรวจพบกิจกรรมที่เป็นอันตรายจากต้นทาง %s ซึ่งมุ่งเป้าไปยัง %s", ai.AttackerIP, ai.TargetHost)
		} else {
			descriptionTh = fmt.Sprintf("ยืนยันแล้วว่ากิจกรรมระหว่าง %s และ %s มีความปลอดภัยและได้รับอนุญาต", ai.TargetHost, ai.AttackerIP)
		}
	}

	support := "benign"
	if ai.IsTruePositive {
		support = "malicious"
	}

	messages := make([]string, 0, 4)
	if ai.IsTruePositive {
		if len(ai.IncidentLogs) > 0 {
			for _, ev := range ai.IncidentLogs {
				if msg := strings.TrimSpace(ev.Message); msg != "" {
					messages = append(messages, msg)
				}
			}
		} else {
			messages = append(messages, legacyAttackEvidenceMessagesForDomain(ai, domainID)...)
		}
	} else if strings.TrimSpace(ai.KeyFinding) != "" {
		messages = append(messages, ai.KeyFinding)
	} else if strings.TrimSpace(ai.AttackLogMessage) != "" {
		messages = append(messages, ai.AttackLogMessage)
	}

	return []aiFinding{
		{
			Description:         description,
			DescriptionTh:       descriptionTh,
			DomainID:            "",
			Points:              DefaultAIFindingPoints,
			EvidenceLogMessages: messages,
			VerdictSupport:      support,
			Explanation:         ai.TpFpExplanation,
			ExplanationTh:       ai.TpFpExplanationTh,
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
func ensureBenignContextIncidentEvents(ai *aiMicroScenario, events []map[string]interface{}) []map[string]interface{} {
	if ai.IsTruePositive {
		return events
	}
	for _, entry := range events {
		if isStrictBenignContextLog(entry) {
			return events
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
			if _, ok := resolveEvidenceMessageIndex(events, trimmed); !ok {
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
	events = append(events, map[string]interface{}{
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

	return events
}

func legacyAttackEvidenceMessages(ai aiMicroScenario) []string {
	return legacyAttackEvidenceMessagesForDomain(ai, "domain4")
}

func legacyAttackEvidenceMessagesForDomain(ai aiMicroScenario, domainID string) []string {
	events := buildDomainContextualFallbackChain(domainID, ai.TargetHost, ai.AttackerIP, ai.PrimaryUser, ai.AttackLogMessage)
	messages := make([]string, 0, len(events))
	for _, ev := range events {
		if msg, _ := ev["message"].(string); strings.TrimSpace(msg) != "" {
			messages = append(messages, msg)
		}
	}
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
	descriptionTh := strings.TrimSpace(ai.KeyFindingTh)
	domain := weakestDomain
	points := DefaultAIFindingPoints
	explanation := ai.TpFpExplanation
	explanationTh := ai.TpFpExplanationTh

	if len(ai.Findings) > 0 {
		first := ai.Findings[0]
		if trimmed := strings.TrimSpace(first.Description); trimmed != "" {
			description = trimmed
		}
		if trimmed := strings.TrimSpace(first.DescriptionTh); trimmed != "" {
			descriptionTh = trimmed
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
		if trimmed := strings.TrimSpace(first.ExplanationTh); trimmed != "" {
			explanationTh = trimmed
		}
	}
	if description == "" {
		if ai.IsTruePositive {
			description = fmt.Sprintf("Identified malicious activity originating from %s targeting %s", ai.AttackerIP, ai.TargetHost)
			if descriptionTh == "" && strings.TrimSpace(ai.TitleTh) != "" {
				descriptionTh = fmt.Sprintf("ตรวจพบกิจกรรมที่เป็นอันตรายจากต้นทาง %s ซึ่งมุ่งเป้าไปยัง %s", ai.AttackerIP, ai.TargetHost)
			}
		} else {
			description = fmt.Sprintf("Verified the activity involving %s and %s is benign and authorized", ai.TargetHost, ai.AttackerIP)
			if descriptionTh == "" && strings.TrimSpace(ai.TitleTh) != "" {
				descriptionTh = fmt.Sprintf("ยืนยันแล้วว่ากิจกรรมระหว่าง %s และ %s มีความปลอดภัยและได้รับอนุญาต", ai.TargetHost, ai.AttackerIP)
			}
		}
	} else if descriptionTh == "" && strings.TrimSpace(ai.TitleTh) != "" {
		if ai.IsTruePositive {
			descriptionTh = fmt.Sprintf("ตรวจพบกิจกรรมที่เป็นอันตรายจากต้นทาง %s ซึ่งมุ่งเป้าไปยัง %s", ai.AttackerIP, ai.TargetHost)
		} else {
			descriptionTh = fmt.Sprintf("ยืนยันแล้วว่ากิจกรรมระหว่าง %s และ %s มีความปลอดภัยและได้รับอนุญาต", ai.TargetHost, ai.AttackerIP)
		}
	}

	support := "malicious"
	var indices []int
	if ai.IsTruePositive {
		indices = fallbackAttackEvidenceIndicesForAI(ai, logs)
	} else {
		support = "benign"
		indices = benignContextEvidenceIndices(logs)
	}

	return mappedAIFinding{
		finding: aiFinding{
			Description:         description,
			DescriptionTh:       descriptionTh,
			DomainID:            domain,
			Points:              float64(points),
			EvidenceLogMessages: messagesAtIndices(logs, indices),
			VerdictSupport:      support,
			Explanation:         explanation,
			ExplanationTh:       explanationTh,
		},
		indices: indices,
	}
}

func fallbackAttackEvidenceIndicesForAI(ai aiMicroScenario, logs []map[string]interface{}) []int {
	capLimit := len(logs) / 2
	if capLimit < 1 {
		capLimit = 1
	}

	matched := make([]int, 0, len(ai.IncidentLogs)+4)
	seen := make(map[int]bool)
	addIdx := func(idx int) {
		if idx < 0 || idx >= len(logs) || seen[idx] || len(matched) >= capLimit {
			return
		}
		seen[idx] = true
		matched = append(matched, idx)
	}

	for _, ev := range ai.IncidentLogs {
		if idx, ok := resolveEvidenceMessageIndex(logs, ev.Message); ok {
			addIdx(idx)
		}
	}
	if strings.TrimSpace(ai.AttackLogMessage) != "" {
		if idx, ok := resolveEvidenceMessageIndex(logs, ai.AttackLogMessage); ok {
			addIdx(idx)
		}
	}

	for i, entry := range logs {
		if scenarioLogSeverityBucket(entry) == "high" {
			addIdx(i)
		}
	}
	if len(matched) > 0 {
		return matched
	}
	return fallbackAttackEvidenceIndices(logs)
}

func fallbackAttackEvidenceIndices(logs []map[string]interface{}) []int {
	high := make([]int, 0, len(logs))
	for i, entry := range logs {
		if scenarioLogSeverityBucket(entry) == "high" {
			high = append(high, i)
		}
	}
	if len(high) > 0 {
		return high
	}

	start := len(logs) - 4
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
	outcomes, _ := buildExpectedOutcomesWithTranslations(mapped, ai, weakestDomain)
	return outcomes
}

func buildExpectedOutcomesWithTranslations(mapped []mappedAIFinding, ai aiMicroScenario, weakestDomain string) (map[string]interface{}, map[string]map[string]string) {
	findings := make([]map[string]interface{}, 0, len(mapped))
	thFindings := make(map[string]map[string]string)
	for i, item := range mapped {
		findingID := fmt.Sprintf("f%d", i+1)
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
			"id":                   findingID,
			"description":          description,
			"domain_id":            domain,
			"points":               normalizeAIFindingPoints(item.finding.Points),
			"evidence_log_indices": item.indices,
			"verdict_support":      support,
			"explanation":          explanation,
		})

		descriptionTh := strings.TrimSpace(item.finding.DescriptionTh)
		explanationTh := strings.TrimSpace(item.finding.ExplanationTh)
		if explanationTh == "" && descriptionTh != "" {
			explanationTh = strings.TrimSpace(ai.TpFpExplanationTh)
		}
		if descriptionTh != "" || explanationTh != "" {
			thFindings[findingID] = map[string]string{
				"description": descriptionTh,
				"explanation": explanationTh,
			}
		}
	}

	return map[string]interface{}{"key_findings": findings}, thFindings
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

	if hasAllStoryBeats(ai.Story) {
		ensureStoryEntityMentions(ai.Story, ai.TargetHost, ai.AttackerIP)
		ai.Description = ComposeStoryDescription(ai.Story)
	} else if strings.TrimSpace(ai.Description) == "" {
		ai.Description = fmt.Sprintf("Security telemetry detected an event requiring triage on %s involving %s.", ai.TargetHost, ai.AttackerIP)
	} else if !strings.Contains(strings.ToLower(ai.Description), strings.ToLower(ai.TargetHost)) {
		lead := formatStyleEntityLead(ai.BriefingStyle, ai.TargetHost, ai.AttackerIP)
		ai.Description = strings.TrimSpace(ai.Description) + " " + lead
	}

	ai.Findings = normalizeAIFindingsForDomain(ai, weakestDomain)
	now := time.Now().UTC()
	incidentEvents := normalizeIncidentEvents(ai, weakestDomain)
	incidentEvents = ensureBenignContextIncidentEvents(&ai, incidentEvents)
	logs := assembleInterleavedScenarioLogs(ai, weakestDomain, incidentEvents, now)

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

	containmentSteps, containmentTh := buildPhaseStepsWithDistractorsWithTh("containment", "c", weakestDomain, ai.ContainmentAction, ai.ContainmentActionTh, ai.ContainmentSteps, ai)
	eradicationSteps, eradicationTh := buildPhaseStepsWithDistractorsWithTh("eradication", "e", weakestDomain, ai.EradicationAction, ai.EradicationActionTh, ai.EradicationSteps, ai)
	recoverySteps, recoveryTh := buildPhaseStepsWithDistractorsWithTh("recovery", "r", weakestDomain, ai.RecoveryAction, ai.RecoveryActionTh, ai.RecoverySteps, ai)

	expectedOutcomes, thFindings := buildExpectedOutcomesWithTranslations(mapped, ai, weakestDomain)

	thPlaybook := make(map[string]map[string]string, len(containmentTh)+len(eradicationTh)+len(recoveryTh))
	for k, v := range containmentTh {
		thPlaybook[k] = v
	}
	for k, v := range eradicationTh {
		thPlaybook[k] = v
	}
	for k, v := range recoveryTh {
		thPlaybook[k] = v
	}

	var translations map[string]interface{}
	if hasAllStoryBeats(ai.Story) {
		translations = map[string]interface{}{
			"en": map[string]interface{}{
				"description": ai.Description,
				"story":       storyBeatsToMap(ai.Story),
			},
		}
	}

	thTitle := strings.TrimSpace(ai.TitleTh)
	var thDesc string
	var thStoryMap map[string]interface{}
	if hasAllThaiStoryBeats(ai.StoryTh) {
		ensureStoryEntityMentions(ai.StoryTh, ai.TargetHost, ai.AttackerIP)
		thDesc = ComposeStoryDescription(ai.StoryTh)
		thStoryMap = storyBeatsToMap(ai.StoryTh)
	} else {
		thDesc = strings.TrimSpace(ai.DescriptionTh)
		if thDesc != "" && !strings.Contains(strings.ToLower(thDesc), strings.ToLower(ai.TargetHost)) {
			thDesc = strings.TrimSpace(thDesc) + " (" + ai.TargetHost + " / " + ai.AttackerIP + ")"
		}
	}
	thTpFp := strings.TrimSpace(ai.TpFpExplanationTh)
	if thTitle != "" || thDesc != "" || thTpFp != "" || len(thFindings) > 0 {
		if translations == nil {
			translations = make(map[string]interface{})
		}
		thMap := map[string]interface{}{
			"title":             thTitle,
			"description":       thDesc,
			"tp_fp_explanation": thTpFp,
			"playbook_steps":    thPlaybook,
			"findings":          thFindings,
		}
		if thStoryMap != nil {
			thMap["story"] = thStoryMap
		}
		translations["th"] = thMap
	}

	return &models.Scenario{
		Title:            ai.Title,
		Description:      ai.Description,
		DomainID:         weakestDomain,
		Difficulty:       3,
		IsTruePositive:   ai.IsTruePositive,
		TpFpExplanation:  ai.TpFpExplanation,
		Translations:     translations,
		InitialLogs:      logs,
		ExpectedOutcomes: expectedOutcomes,
		PlaybookSteps: map[string]interface{}{
			"containment": containmentSteps,
			"eradication": eradicationSteps,
			"recovery":    recoverySteps,
		},
		Status: "active",
	}, errs
}

func formatStyleEntityLead(briefingStyle, targetHost, attackerIP string) string {
	switch briefingStyle {
	case "shift_handover_brief":
		return fmt.Sprintf("Outgoing shift notes flag host %s and peer address %s for immediate review.", targetHost, attackerIP)
	case "threat_hunting_lead":
		return fmt.Sprintf("Hunt scope centers on telemetry from %s and connections involving %s.", targetHost, attackerIP)
	case "helpdesk_user_report":
		return fmt.Sprintf("Affected endpoint %s exhibited anomalous activity linked to %s.", targetHost, attackerIP)
	case "cloud_guardrail_alert":
		return fmt.Sprintf("Guardrail context identifies resource %s with activity from %s.", targetHost, attackerIP)
	default:
		return fmt.Sprintf("Primary alert indicators reference asset %s and remote address %s.", targetHost, attackerIP)
	}
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
	builder.WriteString("\nReturn the complete scenario JSON again (including 4-8 chronological incident_logs). Every findings[].evidence_log_messages entry MUST be copied verbatim from your incident_logs[].message or these generated log messages:\n")
	for _, message := range availableMessages {
		builder.WriteString("- ")
		builder.WriteString(message)
		builder.WriteString("\n")
	}
	builder.WriteString("\nDo not invent messages. Return 1-4 findings. A True Positive needs at least one malicious finding. A False Positive needs at least one benign finding pointing to a benign context log (approved change, HR travel, scheduled maintenance, authorized scan) inside incident_logs. Keep the evidence union at or below 50% of the logs.")
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
	seedByTitle, err := loadSeedScenariosByTitleFromJSON(seedJSON)
	if err != nil {
		return err
	}
	if db == nil {
		return nil
	}

	seedTitles := make(map[string]struct{}, len(seedByTitle))
	for title := range seedByTitle {
		seedTitles[title] = struct{}{}
	}

	var stored []models.Scenario
	if err := db.Find(&stored).Error; err != nil {
		return err
	}

	archivedCount := 0
	syncedStoryCount := 0
	for i := range stored {
		scenario := &stored[i]
		normTitle := normalizeScenarioTitle(scenario.Title)
		if seedScen, isStatic := seedByTitle[normTitle]; isStatic {
			if !HasCompleteStory(scenario) && HasCompleteStory(&seedScen) {
				if syncErr := db.Model(&models.Scenario{}).
					Where("scenario_id = ?", scenario.ScenarioID).
					Updates(map[string]interface{}{
						"description":  seedScen.Description,
						"translations": seedScen.Translations,
					}).Error; syncErr == nil {
					syncedStoryCount++
				} else {
					log.Printf("[Scenario Story Sync] Failed to refresh narrative for seed scenario %q (%s): %v", scenario.Title, scenario.ScenarioID, syncErr)
				}
			}
		}

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
			log.Printf("[Scenario Evidence Quality Gate] Archived invalid legacy scenario %q (%s): %v", scenario.Title, scenario.ScenarioID, validationErr)
		}
	}

	if syncedStoryCount > 0 {
		log.Printf("[Scenario Story Sync] Upgraded narrative stories on %d static seed scenario(s).", syncedStoryCount)
	}
	if archivedCount > 0 {
		log.Printf("[Scenario Evidence Quality Gate] Archived %d invalid legacy scenario(s).", archivedCount)
	}
	return nil
}

// shouldArchiveScenarioEvidence is the pure remediation policy used at startup.
// Static authored Scenarios (matched to the seed by title) and already archived
// rows are never candidates. Active non-seed rows that contain banned legacy
// boilerplate log substrings or fail ValidateScenarioCandidate are archived.
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

	logs := normalizePhaseStepSlice(scenario.InitialLogs)
	if banned, found := containsBannedLegacyBoilerplate(logs); found {
		return true, fmt.Errorf("contains banned legacy boilerplate %q", banned)
	}
	if err := ValidateScenarioCandidate(scenario); err != nil {
		return true, err
	}
	return false, nil
}

func loadSeedScenarioTitles(seedPath string) (map[string]struct{}, error) {
	rawBytes, err := os.ReadFile(seedPath)
	if err != nil {
		return nil, err
	}
	return loadSeedScenarioTitlesFromJSON(rawBytes)
}

func loadSeedScenarioTitlesFromJSON(seedJSON []byte) (map[string]struct{}, error) {
	byTitle, err := loadSeedScenariosByTitleFromJSON(seedJSON)
	if err != nil {
		return nil, err
	}
	titles := make(map[string]struct{}, len(byTitle))
	for title := range byTitle {
		titles[title] = struct{}{}
	}
	return titles, nil
}

func loadSeedScenariosByTitleFromJSON(seedJSON []byte) (map[string]models.Scenario, error) {
	if len(seedJSON) == 0 {
		return nil, fmt.Errorf("scenario seed is empty")
	}

	var seedScenarios []models.Scenario
	if err := json.Unmarshal(seedJSON, &seedScenarios); err != nil {
		return nil, fmt.Errorf("scenario seed is not valid JSON: %w", err)
	}

	byTitle := make(map[string]models.Scenario, len(seedScenarios))
	for _, scenario := range seedScenarios {
		title := normalizeScenarioTitle(scenario.Title)
		if title == "" {
			continue
		}
		byTitle[title] = scenario
	}
	if len(byTitle) == 0 {
		return nil, fmt.Errorf("scenario seed contains no scenario titles")
	}
	return byTitle, nil
}

func normalizeScenarioTitle(title string) string {
	return strings.ToLower(strings.TrimSpace(title))
}
