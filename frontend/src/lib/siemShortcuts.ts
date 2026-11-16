export interface SiemShortcut {
  labelPrefix: string;
  label: string;
  query: string;
  tone: "emerald" | "amber" | "sky";
}

export interface ScenarioShortcutInput {
  title?: string;
  description?: string;
  initial_logs?: readonly Record<string, unknown>[];
}

interface EntityStats {
  bestLogScore: number;
  highSignalCount: number;
  totalCount: number;
  inContext: boolean;
  isExternalIp?: boolean;
}

interface RankedEntity {
  value: string;
  score: number;
  bestLogScore: number;
  inContext: boolean;
  isExternalIp?: boolean;
}

const SEVERITY_WEIGHTS: Record<string, number> = {
  critical: 5,
  high: 4,
  medium: 2,
  warning: 2,
  low: 1,
  info: 1,
};

const CONTEXT_HIGH_SIGNAL_SOURCES = new Set(["change_management", "hr_portal"]);

const GENERIC_USERS = new Set([
  "",
  "-",
  "n/a",
  "na",
  "none",
  "null",
  "unknown",
  "system",
  "nt authority\\system",
  "local service",
  "network service",
  "anonymous",
]);

const GENERIC_HOSTS = new Set(["", "-", "n/a", "na", "none", "null", "unknown", "localhost"]);

const STOP_WORDS = new Set([
  "about", "after", "alert", "alerts", "amid", "another", "between", "could",
  "during", "event", "events", "flags", "from", "high", "incident", "into",
  "logs", "multiple", "observed", "over", "report", "security", "server",
  "should", "system", "target", "that", "their", "these", "this", "through",
  "unexpected", "user", "using", "warning", "when", "with", "within",
]);

const SOURCE_ACRONYMS: Record<string, string> = {
  ad: "AD",
  api: "API",
  aws: "AWS",
  cloudtrail: "CloudTrail",
  crm: "CRM",
  db: "DB",
  dhcp: "DHCP",
  dlp: "DLP",
  dns: "DNS",
  edr: "EDR",
  hr: "HR",
  iam: "IAM",
  ids: "IDS",
  ips: "IPS",
  itsm: "ITSM",
  k8s: "K8s",
  ldap: "LDAP",
  mdm: "MDM",
  mfa: "MFA",
  nids: "NIDS",
  ntp: "NTP",
  postgresql: "PostgreSQL",
  rdp: "RDP",
  s3: "S3",
  siem: "SIEM",
  smb: "SMB",
  sql: "SQL",
  ssh: "SSH",
  sso: "SSO",
  tls: "TLS",
  vpn: "VPN",
  waf: "WAF",
};

function normalizeToken(raw: unknown): string {
  if (typeof raw !== "string") {
    return "";
  }
  return raw.trim();
}

function isValidIp(value: string): boolean {
  if (!value || GENERIC_HOSTS.has(value.toLowerCase())) {
    return false;
  }
  const parts = value.split(".");
  if (parts.length !== 4) {
    return value.includes(":");
  }
  return parts.every((part) => {
    if (!/^\d{1,3}$/.test(part)) return false;
    const num = Number(part);
    return num >= 0 && num <= 255;
  });
}

function isPrivateIp(ip: string): boolean {
  if (!ip || ip.includes(":")) {
    return (
      ip === "::1" ||
      ip.toLowerCase().startsWith("fe80:") ||
      ip.toLowerCase().startsWith("fc") ||
      ip.toLowerCase().startsWith("fd")
    );
  }
  const parts = ip.split(".").map(Number);
  if (parts.length !== 4 || parts.some(Number.isNaN)) {
    return true;
  }
  const [a, b] = parts;
  if (a === 10 || a === 127 || a === 0) return true;
  if (a === 192 && b === 168) return true;
  if (a === 172 && b >= 16 && b <= 31) return true;
  if (a === 169 && b === 254) return true;
  return false;
}

function extractContextKeywords(contextLower: string): string[] {
  const tokens = contextLower.match(/[a-z0-9_-]{3,}/g) ?? [];
  const unique = new Set<string>();
  for (const tok of tokens) {
    if (!STOP_WORDS.has(tok) && !/^\d+$/.test(tok)) {
      unique.add(tok);
    }
  }
  return Array.from(unique);
}

function tokenAppearsInContext(token: string, contextLower: string): boolean {
  const normalized = token.trim().toLowerCase();
  if (normalized.length < 3) {
    return false;
  }
  return contextLower.includes(normalized);
}

export function formatSourceLabel(source: string): string {
  const cleaned = source.trim();
  if (!cleaned) return "Event";
  return cleaned
    .split(/[_-]+/)
    .filter(Boolean)
    .map((word) => {
      const lower = word.toLowerCase();
      if (SOURCE_ACRONYMS[lower]) {
        return SOURCE_ACRONYMS[lower];
      }
      return lower.charAt(0).toUpperCase() + lower.slice(1);
    })
    .join(" ");
}

function makeSourceShortcut(source: string): SiemShortcut {
  return {
    labelPrefix: "Source:",
    label: formatSourceLabel(source),
    query: `source == "${source}"`,
    tone: "sky",
  };
}

function recordEntityScore(
  map: Map<string, EntityStats>,
  value: string,
  logScore: number,
  inContext: boolean,
  isExternalIp = false
): void {
  const key = value.trim();
  if (!key) return;
  const existing = map.get(key);
  const highSignal = logScore >= 3 ? 1 : 0;
  if (!existing) {
    map.set(key, {
      bestLogScore: logScore,
      highSignalCount: highSignal,
      totalCount: 1,
      inContext,
      isExternalIp,
    });
    return;
  }
  existing.bestLogScore = Math.max(existing.bestLogScore, logScore);
  existing.highSignalCount += highSignal;
  existing.totalCount += 1;
  existing.inContext = existing.inContext || inContext;
  existing.isExternalIp = Boolean(existing.isExternalIp || isExternalIp);
}

function rankEntities(map: Map<string, EntityStats>, preferExternalIp = false): RankedEntity[] {
  return Array.from(map.entries())
    .map(([value, stats]) => {
      const externalBonus = preferExternalIp && stats.isExternalIp ? 50 : 0;
      const score =
        stats.bestLogScore * 100 +
        stats.highSignalCount * 10 +
        Math.min(stats.totalCount, 9) +
        externalBonus;
      return {
        value,
        score,
        bestLogScore: stats.bestLogScore,
        inContext: stats.inContext,
        isExternalIp: stats.isExternalIp,
      };
    })
    .sort((a, b) => b.score - a.score);
}

function buildSeverityShortcut(logs: readonly Record<string, unknown>[]): SiemShortcut | null {
  let hasCritical = false;
  let hasHigh = false;
  let hasWarningOrMedium = false;

  for (const log of logs) {
    const sev = normalizeToken(log?.severity).toLowerCase();
    if (sev === "critical") hasCritical = true;
    else if (sev === "high") hasHigh = true;
    else if (sev === "warning" || sev === "medium") hasWarningOrMedium = true;
  }

  if (hasCritical && hasHigh) {
    return {
      labelPrefix: "Filter:",
      label: "High & Critical Alerts",
      query: 'severity == "high" or severity == "critical"',
      tone: "amber",
    };
  }
  if (hasCritical) {
    return {
      labelPrefix: "Filter:",
      label: "Critical Alerts",
      query: 'severity == "critical"',
      tone: "amber",
    };
  }
  if (hasHigh) {
    return {
      labelPrefix: "Filter:",
      label: "High Severity Alerts",
      query: 'severity == "high"',
      tone: "amber",
    };
  }
  if (hasWarningOrMedium) {
    return {
      labelPrefix: "Filter:",
      label: "Warning Alerts",
      query: 'severity == "warning" or severity == "medium"',
      tone: "amber",
    };
  }
  return null;
}

export function buildInvestigationShortcuts(scenario: ScenarioShortcutInput | null | undefined): SiemShortcut[] {
  if (!scenario || !Array.isArray(scenario.initial_logs) || scenario.initial_logs.length === 0) {
    return [];
  }

  const contextLower = `${scenario.title ?? ""} ${scenario.description ?? ""}`.toLowerCase();
  const contextKeywords = extractContextKeywords(contextLower);

  const hostMap = new Map<string, EntityStats>();
  const ipMap = new Map<string, EntityStats>();
  const userMap = new Map<string, EntityStats>();
  const sourceMap = new Map<string, EntityStats>();

  for (const log of scenario.initial_logs) {
    if (!log || typeof log !== "object") continue;

    const sev = normalizeToken(log.severity).toLowerCase();
    const baseWeight = SEVERITY_WEIGHTS[sev] ?? 1;
    const source = normalizeToken(log.source);
    const sourceLower = source.toLowerCase();
    const msgLower = normalizeToken(log.message).toLowerCase();
    const contextSourceBoost =
      baseWeight <= 1 &&
      CONTEXT_HIGH_SIGNAL_SOURCES.has(sourceLower) &&
      (msgLower.includes("approved") || msgLower.includes("authorized") || msgLower.includes("cr-"))
        ? 1
        : 0;
    let keywordHits = 0;
    for (const kw of contextKeywords) {
      if (msgLower.includes(kw)) {
        keywordHits += 1;
        if (keywordHits >= 3) break;
      }
    }
    const baseLogWeight = baseWeight + contextSourceBoost + keywordHits * 3;

    const host = normalizeToken(log.host);
    const hostValid = Boolean(host && !GENERIC_HOSTS.has(host.toLowerCase()));
    const hostInContext = Boolean(hostValid && tokenAppearsInContext(host, contextLower));

    const user = normalizeToken(log.user);
    const userValid = Boolean(user && !GENERIC_USERS.has(user.toLowerCase()));
    const userInContext = Boolean(userValid && tokenAppearsInContext(user, contextLower));

    const candidateIps = [normalizeToken(log.dst_ip), normalizeToken(log.src_ip), normalizeToken(log.ip)].filter(isValidIp);
    const anyIpInContext = candidateIps.some((ip) => tokenAppearsInContext(ip, contextLower));

    if (hostValid) {
      recordEntityScore(hostMap, host, baseLogWeight + (hostInContext ? 20 : 0), hostInContext);
    }

    for (const ip of candidateIps) {
      const thisIpInContext = tokenAppearsInContext(ip, contextLower);
      recordEntityScore(ipMap, ip, baseLogWeight + (thisIpInContext ? 20 : 0), thisIpInContext, !isPrivateIp(ip));
    }

    if (userValid) {
      recordEntityScore(userMap, user, baseLogWeight + (userInContext ? 20 : 0), userInContext);
    }

    if (source && !GENERIC_HOSTS.has(sourceLower)) {
      const sourceInContext = tokenAppearsInContext(sourceLower.replace(/_/g, " "), contextLower);
      const anyEntityInContext = hostInContext || userInContext || anyIpInContext;
      const sourceScore = baseLogWeight + (sourceInContext ? 10 : 0) + (anyEntityInContext ? 5 : 0);
      recordEntityScore(sourceMap, source, sourceScore, sourceInContext || anyEntityInContext);
    }
  }

  const rankedHosts = rankEntities(hostMap);
  const rankedIps = rankEntities(ipMap, true);
  const rankedUsers = rankEntities(userMap);
  const rankedSources = rankEntities(sourceMap);

  const shortcuts: SiemShortcut[] = [];
  const usedQueries = new Set<string>();

  const pushUnique = (shortcut: SiemShortcut | null) => {
    if (!shortcut || shortcuts.length >= 4) return;
    const normalizedQuery = shortcut.query.trim().toLowerCase();
    if (!normalizedQuery || usedQueries.has(normalizedQuery)) return;
    usedQueries.add(normalizedQuery);
    shortcuts.push(shortcut);
  };

  if (rankedHosts.length > 0) {
    const topHost = rankedHosts[0];
    pushUnique({
      labelPrefix: "Host:",
      label: topHost.value,
      query: topHost.value,
      tone: "emerald",
    });
  }

  const topIp = rankedIps[0];
  const topUser = rankedUsers[0];
  if (topIp && (!topUser || topIp.score >= topUser.score)) {
    pushUnique({
      labelPrefix: "IP:",
      label: topIp.value,
      query: topIp.value,
      tone: "emerald",
    });
  } else if (topUser) {
    pushUnique({
      labelPrefix: "User:",
      label: topUser.value,
      query: `user == "${topUser.value}"`,
      tone: "emerald",
    });
  }

  if (rankedSources.length > 0) {
    pushUnique(makeSourceShortcut(rankedSources[0].value));
  }

  const secondarySource = rankedSources[1];
  const hasHighSignalSecondarySource = Boolean(secondarySource && secondarySource.bestLogScore >= 3);
  if (hasHighSignalSecondarySource) {
    pushUnique(makeSourceShortcut(secondarySource.value));
  }

  pushUnique(buildSeverityShortcut(scenario.initial_logs));

  if (shortcuts.length < 4 && secondarySource) {
    pushUnique(makeSourceShortcut(secondarySource.value));
  }

  if (shortcuts.length < 4 && topUser) {
    pushUnique({
      labelPrefix: "User:",
      label: topUser.value,
      query: `user == "${topUser.value}"`,
      tone: "emerald",
    });
  }

  return shortcuts;
}
