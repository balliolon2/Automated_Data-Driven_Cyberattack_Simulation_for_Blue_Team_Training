import test from "node:test";
import assert from "node:assert/strict";
import { en } from "../src/i18n/en.ts";
import { th } from "../src/i18n/th.ts";
import {
  interpolate,
  translate,
  formatDate,
  resolveApiError,
  resolveDomainName,
  resolveScenarioLocale,
} from "../src/i18n/format.ts";

function flattenDict(obj: Record<string, unknown>, prefix = ""): Record<string, string> {
  const out: Record<string, string> = {};
  for (const [k, v] of Object.entries(obj)) {
    const key = prefix ? `${prefix}.${k}` : k;
    if (v !== null && typeof v === "object" && !Array.isArray(v)) {
      Object.assign(out, flattenDict(v as Record<string, unknown>, key));
    } else {
      out[key] = String(v ?? "");
    }
  }
  return out;
}

function extractPlaceholders(str: string): string[] {
  const matches = str.match(/\{\{([^}]+)\}\}/g) ?? [];
  return matches.map((m) => m.slice(2, -2).trim()).sort();
}

test("en and th dictionaries have 100% key parity, non-empty leaves, and identical {{param}} placeholders", () => {
  const flatEn = flattenDict(en as unknown as Record<string, unknown>);
  const flatTh = flattenDict(th as unknown as Record<string, unknown>);

  const enKeys = Object.keys(flatEn).sort();
  const thKeys = Object.keys(flatTh).sort();

  assert.deepEqual(thKeys, enKeys, "en and th key sets must match identically");
  assert.ok(enKeys.length > 50, "dictionary should contain M1 namespaces");

  for (const key of enKeys) {
    assert.ok(flatEn[key].trim().length > 0, `en.${key} must not be empty`);
    assert.ok(flatTh[key].trim().length > 0, `th.${key} must not be empty`);
    assert.deepEqual(
      extractPlaceholders(flatTh[key]),
      extractPlaceholders(flatEn[key]),
      `{{param}} placeholders must match for ${key}`,
    );
  }
});

test("canonical M1 error codes and bilingual domain names are present", () => {
  const requiredCodes = [
    "AUTH_UNAUTHORIZED",
    "AUTH_FORBIDDEN",
    "AUTH_INVALID_PAYLOAD",
    "AUTH_INVALID_CREDENTIALS",
    "AUTH_EMAIL_EXISTS",
    "AUTH_NICKNAME_REQUIRED",
    "AUTH_NICKNAME_TAKEN",
    "AUTH_NICKNAME_INVALID",
    "INVALID_LANGUAGE",
    "USER_NOT_FOUND",
    "EXAM_PRETEST_REQUIRED",
    "EXAM_POSTTEST_LOCKED",
    "EXAM_SESSION_NOT_FOUND",
    "EXAM_QUESTION_NOT_FOUND",
    "EXAM_INVALID_ANSWER",
    "EXAM_SESSION_ACTIVE",
    "LEARN_PRETEST_REQUIRED",
    "LEARN_DOMAIN_NOT_FOUND",
    "LEARN_SUBTOPIC_NOT_FOUND",
    "LEARN_SUBTOPICS_INCOMPLETE",
    "LEARN_SAMPLE_NOT_FOUND",
    "LEARN_SAMPLE_INCOMPLETE",
    "LEARN_INVALID_ANSWER",
    "INTERNAL_SERVER_ERROR",
    "SIM_PRETEST_REQUIRED",
    "SPECIALIST_BIO_REQUIRED",
    "SPECIALIST_RESUME_REQUIRED",
  ] as const;

  for (const code of requiredCodes) {
    assert.ok(code in en.errors, `en.errors missing ${code}`);
    assert.ok(code in th.errors, `th.errors missing ${code}`);
  }

  assert.equal(resolveDomainName("domain1", "Fallback", "th"), "แนวคิดความปลอดภัยทั่วไป (General Security Concepts)");
  assert.equal(resolveDomainName("domain2", "Fallback", "th"), "ภัยคุกคาม ช่องโหว่ และการลดความเสี่ยง (Threats, Vulnerabilities, and Mitigations)");
  assert.equal(resolveDomainName("domain3", "Fallback", "th"), "สถาปัตยกรรมความปลอดภัย (Security Architecture)");
  assert.equal(resolveDomainName("domain4", "Fallback", "th"), "ปฏิบัติการด้านความปลอดภัย (Security Operations)");
  assert.equal(resolveDomainName("domain5", "Fallback", "th"), "การบริหารจัดการและกำกับดูแลความปลอดภัย (Security Program Management and Oversight)");
  assert.equal(resolveDomainName("domain1", "Fallback", "en"), "General Security Concepts");
  assert.equal(resolveDomainName("unknown_domain", "Custom Fallback", "th"), "Custom Fallback");
});

test("interpolate and translate handle placeholders and unknown keys safely", () => {
  assert.equal(interpolate("Tier {{tier}} Analyst", { tier: 2 }), "Tier 2 Analyst");
  assert.equal(interpolate("{{a}} / {{b}}", { a: 3 }), "3 / {{b}}");
  assert.equal(translate("en", "profile.tierBadge", { tier: 3 }), "Tier 3 Analyst");
  assert.equal(translate("th", "profile.tierBadge", { tier: 3 }), "นักวิเคราะห์ระดับ Tier 3");
  assert.equal(translate("th", "missing.nonexistent.key"), "missing.nonexistent.key");
});

test("formatDate uses Gregorian calendar for both en-US and th-TH-u-ca-gregory", () => {
  const iso = "2026-11-20T13:47:45Z";
  assert.equal(formatDate(iso, "en"), "Nov 20, 2026");

  const thaiFormatted = formatDate(iso, "th");
  assert.ok(thaiFormatted.includes("2026"), `Expected Gregorian year 2026 in ${thaiFormatted}`);
  assert.ok(!thaiFormatted.includes("2569"), `Must NOT render Buddhist Era 2569 in ${thaiFormatted}`);
  assert.ok(/[\u0E00-\u0E7F]/.test(thaiFormatted), `Expected Thai month in ${thaiFormatted}`);

  assert.equal(formatDate("not-a-date", "th"), "not-a-date");
});

test("resolveApiError resolves code + params and falls back gracefully", () => {
  const thT = (key: string, params?: Record<string, string | number>) => translate("th", key, params);
  const enT = (key: string, params?: Record<string, string | number>) => translate("en", key, params);

  const lockedErr = {
    response: {
      data: {
        error: "You are not eligible for the Post-Test.",
        code: "EXAM_POSTTEST_LOCKED",
        params: { domain_id: "domain2", proficiency_score: 55, threshold: 70 },
      },
    },
  };

  const resolvedTh = resolveApiError(lockedErr, thT);
  assert.ok(/[\u0E00-\u0E7F]/.test(resolvedTh), `Expected Thai translation, got ${resolvedTh}`);
  assert.ok(resolvedTh.includes("70"), `Expected interpolated threshold 70 in ${resolvedTh}`);

  const resolvedEn = resolveApiError(lockedErr, enT);
  assert.ok(resolvedEn.includes("70"));

  const unknownCodeErr = {
    response: {
      data: {
        error: "Legacy server message",
        code: "SOME_UNMAPPED_CODE",
      },
    },
  };
  assert.equal(resolveApiError(unknownCodeErr, thT), "Legacy server message");

  const noCodeErr = {
    response: {
      data: {
        message: "Session missing",
      },
    },
  };
  assert.equal(resolveApiError(noCodeErr, thT), "Session missing");
  assert.equal(resolveApiError(new Error("boom"), thT, "Custom fallback"), "Custom fallback");
});

test("resolveScenarioLocale selects Thai display fields while preserving initial_logs untouched", () => {
  const initialLogs = [{ timestamp: "2026-11-20T10:00:00Z", source: "okta", severity: "high", host: "auth-01", message: "MFA push denied" }];
  const scenario = {
    title: "MFA Push Fatigue",
    description: "User ajones rejected multiple MFA prompts.",
    tp_fp_explanation: "True Positive due to repeated rejections followed by login.",
    initial_logs: initialLogs,
    playbook_steps: {
      containment: [{ action_id: "c1", label: "Revoke tokens", explanation: "Stops attacker" }],
    },
    findings: [{ finding_id: "f1", description: "Repeated denials", explanation: "Shows fatigue" }],
    translations: {
      th: {
        title: "การโจมตี MFA Push Fatigue",
        description: "พบผู้ใช้ ajones ปฏิเสธคำขอ MFA หลายครั้ง",
        tp_fp_explanation: "เป็น True Positive",
        playbook_steps: {
          c1: { label: "เพิกถอน tokens", explanation: "หยุดผู้โจมตี" },
        },
        findings: {
          f1: { description: "ปฏิเสธซ้ำ", explanation: "บ่งชี้ Push Fatigue" },
        },
      },
    },
  };

  const resolvedTh = resolveScenarioLocale(scenario, "th");
  assert.equal(resolvedTh.title, "การโจมตี MFA Push Fatigue");
  assert.equal(resolvedTh.description, "พบผู้ใช้ ajones ปฏิเสธคำขอ MFA หลายครั้ง");
  assert.equal(resolvedTh.playbookLabel("c1", "Revoke tokens"), "เพิกถอน tokens");
  assert.equal(resolvedTh.findingDescription("f1", "Repeated denials"), "ปฏิเสธซ้ำ");
  assert.strictEqual(scenario.initial_logs, initialLogs);

  const resolvedEn = resolveScenarioLocale(scenario, "en");
  assert.equal(resolvedEn.title, "MFA Push Fatigue");
  assert.equal(resolvedEn.playbookLabel("c1", "Revoke tokens"), "Revoke tokens");
});
