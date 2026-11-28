import { normalizeLocale } from "../i18n/format.ts";
import type { Locale } from "../i18n/types.ts";

export type MiniCaseVerdict = "tp" | "fp";

export interface MiniCaseLogLine {
  timestamp: string;
  source: string;
  severity: "INFO" | "WARN" | "ALERT";
  message: string;
}

export interface PublicLandingMiniCase {
  id: string;
  title: string;
  briefing: string;
  logs: MiniCaseLogLine[];
}

export interface LandingMiniCaseResult {
  isCorrect: boolean;
  expectedVerdict: MiniCaseVerdict;
  rationale: string;
}

interface InternalMiniCase {
  id: string;
  expectedVerdict: MiniCaseVerdict;
  en: {
    title: string;
    briefing: string;
    rationale: string;
  };
  th: {
    title: string;
    briefing: string;
    rationale: string;
  };
  logs: MiniCaseLogLine[];
}

const MINI_CASES: readonly InternalMiniCase[] = [
  {
    id: "midnight-club-login",
    expectedVerdict: "tp",
    en: {
      title: "Case 01: Midnight Login & Suspicious ZIP",
      briefing:
        "At 02:14 AM while the school computer club is asleep, someone logs into the club server and downloads a sensitive file. Check the 3 SIEM logs below — real attack or normal?",
      rationale:
        "Real Attack (True Positive): 18 failed logins followed by a sudden login from overseas IP 185.220.101.44 at 02:14 AM, plus immediate exfiltration of passwords_backup.zip, clearly confirms unauthorized access.",
    },
    th: {
      title: "คดีที่ 01: ล็อกอินตีสองพร้อมดูดไฟล์ลับ",
      briefing:
        "เวลา 02:14 น. ตอนที่ทุกคนหลับหมดแล้ว มีคนล็อกอินเข้าเซิร์ฟเวอร์ชมรมคอมฯ และโหลดไฟล์สำคัญออกไป ลองเช็ก SIEM Logs 3 บรรทัดด้านล่าง — แฮกเกอร์ตัวจริง หรือใช้งานปกติ?",
      rationale:
        "เหตุโจมตีจริง (True Positive): บรรทัดที่ 1–2 พบการเดารหัสผ่านผิด 18 ครั้งแล้วเข้าสำเร็จจาก IP ต่างประเทศ 185.220.101.44 ตอนตี 2 และบรรทัดที่ 3 โหลดไฟล์ passwords_backup.zip ทันที ซึ่งเป็นพฤติกรรมผู้บุกรุกชัดเจน",
    },
    logs: [
      {
        timestamp: "02:14:02Z",
        source: "auth.sshd",
        severity: "WARN",
        message: "Failed password x18 for admin_club from 185.220.101.44 (Tor exit node)",
      },
      {
        timestamp: "02:14:19Z",
        source: "auth.sshd",
        severity: "ALERT",
        message: "LOGIN_SUCCESS user=admin_club src_ip=185.220.101.44 method=password",
      },
      {
        timestamp: "02:14:47Z",
        source: "file.sftp",
        severity: "ALERT",
        message: "DOWNLOAD path=/srv/club/passwords_backup.zip size=4.2MB src_ip=185.220.101.44",
      },
    ],
  },
  {
    id: "scheduled-it-backup",
    expectedVerdict: "fp",
    en: {
      title: "Case 02: Overnight Backup Job Alert",
      briefing:
        "An automated high-volume file transfer alert triggered on the school server at 02:00 AM. Inspect all 3 SIEM logs carefully before sounding the alarm!",
      rationale:
        "Safe Routine (False Positive): The transfer ran inside the scheduled maintenance window by verified internal service account svc_school_backup (10.10.1.5) matching approved change ticket CHG-402.",
    },
    th: {
      title: "คดีที่ 02: แจ้งเตือนโอนย้ายไฟล์ตอนตีสอง",
      briefing:
        "ระบบแจ้งเตือนพบการคัดลอกข้อมูลขนาดใหญ่จากเซิร์ฟเวอร์โรงเรียนตอน 02:00 น. ลองอ่าน SIEM Logs ให้ครบทั้ง 3 บรรทัดก่อนตัดสินใจ — แฮกเกอร์หรือระบบสำรองข้อมูลปกติ?",
      rationale:
        "เหตุการณ์ปกติ (False Positive): แม้จะทำงานตอนตี 2 แต่ Log บรรทัดที่ 2–3 ยืนยันว่ามาจากบัญชีบริการภายใน svc_school_backup (10.10.1.5) ที่มีลายเซ็นดิจิทัลยืนยันตัวตนถูกต้อง (sig=VERIFIED) และตรงกับใบงานบำรุงรักษา CHG-402",
    },
    logs: [
      {
        timestamp: "02:00:01Z",
        source: "cron.backup",
        severity: "INFO",
        message: "Scheduled job nightly_db_snapshot started on host=school-db01",
      },
      {
        timestamp: "02:00:18Z",
        source: "iam.audit",
        severity: "INFO",
        message: "Service account svc_school_backup authenticated from internal 10.10.1.5 (sig=VERIFIED)",
      },
      {
        timestamp: "02:01:05Z",
        source: "net.monitor",
        severity: "WARN",
        message: "Volume transfer 1.8GB to backup-vault-01 matched approved ticket CHG-402",
      },
    ],
  },
];

/** Returns spoiler-free rotatable beginner detective mini-cases for the Landing workbench. */
export function getLandingMiniCases(locale: Locale): PublicLandingMiniCase[] {
  const lang = normalizeLocale(locale);
  return MINI_CASES.map((item) => ({
    id: item.id,
    title: item[lang].title,
    briefing: item[lang].briefing,
    logs: item.logs,
  }));
}

/** Evaluates learner TP/FP decision on a mini-case and returns log-referenced rationale. */
export function evaluateLandingMiniCase(
  caseId: string,
  decision: MiniCaseVerdict,
  locale: Locale,
): LandingMiniCaseResult | null {
  const found = MINI_CASES.find((c) => c.id === caseId);
  if (!found) return null;
  const lang = normalizeLocale(locale);
  return {
    isCorrect: found.expectedVerdict === decision,
    expectedVerdict: found.expectedVerdict,
    rationale: found[lang].rationale,
  };
}
