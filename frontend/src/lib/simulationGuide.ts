import { normalizeLocale } from "../i18n/format.ts";
import type { Locale } from "../i18n/types.ts";

export const BEGINNER_GLOSSARY_TERM_IDS = [
  "siem",
  "kql",
  "ioc",
  "true_positive",
  "false_positive",
  "containment",
  "eradication",
  "recovery",
  "brute_force",
  "phishing",
] as const;

export type BeginnerGlossaryTermId = (typeof BEGINNER_GLOSSARY_TERM_IDS)[number];

export interface BeginnerGlossaryEntry {
  id: BeginnerGlossaryTermId;
  term: string;
  definition: string;
  example: string;
}

const STORAGE_KEY_PREFIX = "cybersim_guide_collapsed";

const GLOSSARY_CATALOG: Record<
  BeginnerGlossaryTermId,
  {
    en: { term: string; definition: string; example: string };
    th: { term: string; definition: string; example: string };
  }
> = {
  siem: {
    en: {
      term: "SIEM (Security Information & Event Management)",
      definition: "A central dashboard that collects login, firewall, and server logs from all computers into one searchable timeline.",
      example: "Checking the SIEM Log Terminal to see who logged in at 02:14 AM.",
    },
    th: {
      term: "SIEM (ศูนย์รวมบันทึกเหตุการณ์ความปลอดภัย)",
      definition: "หน้าจอกลางที่รวบรวม Log จากระบบล็อกอิน ไฟร์วอลล์ และเซิร์ฟเวอร์ทุกเครื่องมาไว้ในที่เดียวเพื่อสืบสวนได้ง่าย",
      example: "เปิด SIEM Log Terminal ดูว่าใครล็อกอินเข้าเครื่องตอนตีสอง",
    },
  },
  kql: {
    en: {
      term: "KQL (Kusto Query Language)",
      definition: "A simple search filter syntax used by SOC analysts to slice thousands of logs down to only suspicious rows.",
      example: "Clicking Quick-Filter chips to show only one IP address or Host.",
    },
    th: {
      term: "KQL (ภาษาคำสั่งกรองข้อมูล Log)",
      definition: "คำสั่งค้นหาสำหรับคัดกรอง Log นับพันบรรทัดให้เหลือเฉพาะเครื่อง IP หรือผู้ใช้ที่ต้องสงสัย",
      example: "กดปุ่ม Quick-Filter ด้านบนเพื่อกรองดูเฉพาะ Host หรือ IP เป้าหมาย",
    },
  },
  ioc: {
    en: {
      term: "IOC (Indicator of Compromise)",
      definition: "A digital footprint left behind by an attacker, such as a malicious IP address, strange file hash, or unauthorized account.",
      example: "An unfamiliar external IP address downloading passwords_backup.zip.",
    },
    th: {
      term: "IOC (ร่องรอยการบุกรุก / หลักฐานดิจิทัล)",
      definition: "ร่องรอยที่แฮกเกอร์ทิ้งไว้ เช่น IP แปลกปลอม ชื่อไฟล์ต้องสงสัย หรือคำสั่งแปลกๆ ในระบบ",
      example: "พบ IP ต่างประเทศที่ไม่รู้จักแอบดาวน์โหลดไฟล์รหัสผ่าน",
    },
  },
  true_positive: {
    en: {
      term: "True Positive (Real Attack / Threat)",
      definition: "The security alert detected an actual unauthorized or malicious event that needs immediate defender action.",
      example: "Brute force logins succeeded from an external attacker IP and stole data.",
    },
    th: {
      term: "True Positive (เหตุโจมตีจริง)",
      definition: "การแจ้งเตือนที่เกิดจากการโจมตีหรือการบุกรุกจริง ต้องรีบเลือกหลักฐานและดำเนินการหยุดยั้งทันที",
      example: "มีคนนอกเดารหัสผ่านสำเร็จและขโมยไฟล์ข้อมูลออกไปจริง",
    },
  },
  false_positive: {
    en: {
      term: "False Positive (Safe Routine / Benign)",
      definition: "The alert fired on legitimate, authorized activity (such as IT maintenance or normal scripts) rather than a real attack.",
      example: "High disk activity at 02:00 AM caused by an approved nightly IT backup ticket.",
    },
    th: {
      term: "False Positive (แจ้งเตือนจากงานปกติ — ไม่ใช่แฮกเกอร์)",
      definition: "ระบบแจ้งเตือนเพราะความเข้าใจผิด แต่จริงๆ เป็นการทำงานที่ได้รับอนุญาตตามปกติ",
      example: "ระบบสำรองข้อมูลอัตโนมัติของฝ่าย IT ที่ได้รับอนุมัติล่วงหน้า",
    },
  },
  containment: {
    en: {
      term: "Containment (Phase 1 Response)",
      definition: "Immediate short-term action to stop an attack from spreading further without destroying evidence.",
      example: "Blocking a suspicious IP at the firewall or isolating an infected host.",
    },
    th: {
      term: "Containment (ระยะที่ 1: จำกัดวงและหยุดการลุกลาม)",
      definition: "ขั้นตอนเร่งด่วนเพื่อตัดทางแฮกเกอร์ไม่ให้ลามไปเครื่องอื่น เช่น แยกเครื่องหรือล็อกบัญชีทันที",
      example: "บล็อก IP ผู้บุกรุกที่ Firewall หรือปิดใช้งานบัญชีที่ถูกยึด",
    },
  },
  eradication: {
    en: {
      term: "Eradication (Phase 2 Response)",
      definition: "Removing every trace of the attacker's access, malware, or backdoors from the environment cleanly.",
      example: "Revoking compromised tokens/SSH keys and deleting malicious scripts.",
    },
    th: {
      term: "Eradication (ระยะที่ 2: กำจัดต้นตอให้สิ้นซาก)",
      definition: "ลบไฟล์อันตราย ประตูหลัง (Backdoor) หรือสิทธิ์ปลอมที่ผู้โจมตีสร้างทิ้งไว้ออกทั้งหมด",
      example: "เพิกถอน SSH key ที่รั่วไหล และลบสคริปต์โจมตีออกจากเครื่อง",
    },
  },
  recovery: {
    en: {
      term: "Recovery (Phase 3 Response)",
      definition: "Safely bringing systems back to normal operation and monitoring closely after the threat is gone.",
      example: "Restoring clean verified backups and resetting impacted passwords.",
    },
    th: {
      term: "Recovery (ระยะที่ 3: ฟื้นฟูระบบกลับมาใช้งานปลอดภัย)",
      definition: "นำบริการที่สะอาดปลอดภัยกลับมาให้บริการตามปกติ พร้อมเฝ้าระวังเป็นพิเศษหลังกำจัดภัยคุกคามแล้ว",
      example: "กู้คืนระบบจาก Backup ที่ปลอดภัยและบังคับเปลี่ยนรหัสผ่านใหม่",
    },
  },
  brute_force: {
    en: {
      term: "Brute Force Attack",
      definition: "Automated repeated guessing of many usernames or passwords in seconds/minutes until one finally works.",
      example: "Seeing dozens of Failed password alerts right before one LOGIN_SUCCESS.",
    },
    th: {
      term: "Brute Force (การสุ่มเดารหัสผ่านรัวๆ)",
      definition: "การใช้โปรแกรมยิงเดารหัสผ่านซ้ำๆ หลายสิบหลายร้อยครั้งในเวลาสั้นๆ เพื่อเจาะเข้าบัญชีผู้ใช้",
      example: "พบ Log แจ้งรหัสผ่านผิดติดกันรัวๆ ก่อนจะมีบรรทัดเข้าสู่ระบบสำเร็จ",
    },
  },
  phishing: {
    en: {
      term: "Phishing (Deceptive Bait Link/Message)",
      definition: "Tricking victims with fake login pages, urgent SMS notices, or look-alike emails so they hand over passwords or run malware.",
      example: "A fake scholarship or parcel SMS tempting users to type their school login.",
    },
    th: {
      term: "Phishing (ลวงเหยื่อกรอกข้อมูล/คลิกลิงก์ปลอม)",
      definition: "การส่งลิงก์ปลอม หน้าเว็บล็อกอินปลอม หรือข้อความหลอกว่าได้รับรางวัล/พัสดุ เพื่อขโมยรหัสผ่านหรือหลอกให้รันไฟล์อันตราย",
      example: "SMS ลิงก์รับทุนการศึกษาปลอมที่หลอกให้กรอกรหัสผ่านบัญชีหลัก",
    },
  },
};

/** Returns per-learner localStorage dismissal key for the Beginner Guide drawer. */
export function getSimulationGuideDismissStorageKey(userId: string | null | undefined): string {
  const clean = typeof userId === "string" ? userId.trim() : "";
  return `${STORAGE_KEY_PREFIX}:${clean || "anon"}`;
}

/** Auto-expands ONLY on a learner's first Scenario (completedScenarios === 0) unless already dismissed. */
export function shouldAutoExpandBeginnerGuide(
  completedScenarios: number | null | undefined,
  isDismissed: boolean,
): boolean {
  return completedScenarios === 0 && !isDismissed;
}

/** Returns the static, zero-spoiler 10-term bilingual SOC beginner glossary. */
export function getBeginnerGlossaryTerms(locale: Locale): BeginnerGlossaryEntry[] {
  const lang = normalizeLocale(locale);
  return BEGINNER_GLOSSARY_TERM_IDS.map((id) => ({
    id,
    term: GLOSSARY_CATALOG[id][lang].term,
    definition: GLOSSARY_CATALOG[id][lang].definition,
    example: GLOSSARY_CATALOG[id][lang].example,
  }));
}
