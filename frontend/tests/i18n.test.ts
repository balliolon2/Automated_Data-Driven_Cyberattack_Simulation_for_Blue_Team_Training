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
  resolveGuestDefaultLocale,
  resolveNotificationLocale,
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
    "EXAM_NOT_COMPLETED",
    "LEARN_PRETEST_REQUIRED",
    "LEARN_DOMAIN_NOT_FOUND",
    "LEARN_SUBTOPIC_NOT_FOUND",
    "LEARN_SUBTOPICS_INCOMPLETE",
    "LEARN_SAMPLE_NOT_FOUND",
    "LEARN_SAMPLE_INCOMPLETE",
    "LEARN_INVALID_ANSWER",
    "INTERNAL_SERVER_ERROR",
    "SIM_PRETEST_REQUIRED",
    "SIM_UNAUTHORIZED",
    "SIM_NO_SKILL_PROFILE",
    "SIM_THEORY_INCOMPLETE",
    "SIM_BAD_REQUEST",
    "SIM_SCENARIO_LOAD_FAILED",
    "SIM_SESSION_CREATE_FAILED",
    "SIM_SESSION_NOT_FOUND",
    "SIM_SESSION_FORBIDDEN",
    "SIM_SESSION_NOT_ACTIVE",
    "SIM_SESSION_NOT_COMPLETED",
    "SIM_SCENARIO_NOT_FOUND",
    "SIM_LOGS_UNAVAILABLE",
    "SIM_LOGS_LOAD_FAILED",
    "SIM_INVALID_LOG_INDEX",
    "SIM_SUBMIT_FAILED",
    "SIM_ACTION_RECORD_FAILED",
    "SIM_EVIDENCE_REPLAY_FAILED",
    "SIM_PROFILE_UPDATE_FAILED",
    "SIM_SESSION_UPDATE_FAILED",
    "SPECIALIST_BIO_REQUIRED",
    "SPECIALIST_RESUME_REQUIRED",
    "SPECIALIST_INVALID_PAYLOAD",
    "SPECIALIST_FILE_TOO_LARGE",
    "SPECIALIST_INVALID_FILE_TYPE",
    "SPECIALIST_ALREADY_PENDING",
    "SPECIALIST_ALREADY_APPROVED",
    "SPECIALIST_APPLICATION_FAILED",
    "SPECIALIST_APPLICATION_NOT_FOUND",
    "SPECIALIST_SUBMISSIONS_FETCH_FAILED",
    "SPECIALIST_SESSION_NOT_FOUND",
    "SPECIALIST_SCENARIO_NOT_FOUND",
    "ADMIN_APPLICATIONS_FETCH_FAILED",
    "ADMIN_APPLICATION_NOT_FOUND",
    "ADMIN_INVALID_FILE_TYPE",
    "ADMIN_DOCUMENT_NOT_FOUND",
    "ADMIN_DOCUMENT_MISSING",
    "ADMIN_ALREADY_APPROVED",
    "ADMIN_APPROVE_FAILED",
    "ADMIN_REJECTION_REASON_REQUIRED",
    "ADMIN_REJECT_FAILED",
    "THREAD_INVALID_PAYLOAD",
    "THREAD_TITLE_TOO_SHORT",
    "THREAD_CONTENT_TOO_SHORT",
    "THREAD_SCENARIO_NOT_FOUND",
    "THREAD_CREATE_FAILED",
    "THREAD_FETCH_FAILED",
    "THREAD_NOT_FOUND",
    "THREAD_FORBIDDEN",
    "THREAD_UPDATE_FAILED",
    "THREAD_DELETE_FAILED",
    "THREAD_UPVOTE_FAILED",
    "COMMENT_INVALID_PAYLOAD",
    "COMMENT_LENGTH_INVALID",
    "COMMENT_PARENT_NOT_FOUND",
    "COMMENT_CREATE_FAILED",
    "COMMENT_FETCH_FAILED",
    "COMMENT_NOT_FOUND",
    "COMMENT_FORBIDDEN",
    "COMMENT_UPDATE_FAILED",
    "COMMENT_DELETE_FAILED",
    "COMMENT_MODERATION_FORBIDDEN",
    "COMMENT_PIN_TOP_LEVEL_ONLY",
    "COMMENT_PIN_LIMIT_REACHED",
    "COMMENT_MODERATION_FAILED",
    "COMMENT_UPVOTE_FAILED",
    "NOTIF_FETCH_FAILED",
    "NOTIF_UPDATE_FAILED",
    "QUESTIONNAIRE_LEARNER_ONLY",
    "QUESTIONNAIRE_POST_TEST_REQUIRED",
    "QUESTIONNAIRE_ALREADY_SUBMITTED",
    "QUESTIONNAIRE_INVALID_CSUQ",
    "QUESTIONNAIRE_ALL_NA",
    "QUESTIONNAIRE_INVALID_TLX",
    "QUESTIONNAIRE_FEEDBACK_TOO_LONG",
    "EVALUATION_INVALID_EXPORT_FORMAT",
    "NFR_PROBE_ALREADY_RUNNING",
    "NFR_INVALID_PROBE_PARAMS",
    "NFR_INVALID_BENCHMARK_PAYLOAD",
    "NFR_PAYLOAD_TOO_LARGE",
    "NFR_RUN_NOT_FOUND",
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

test("resolveNotificationLocale renders dynamic templated notifications in en/th and falls back on legacy rows", () => {
  const commentNotif = {
    title: "New reply to your comment",
    message: "alice replied to your comment on 'DNS Tunneling Notes'",
    template_key: "comment_reply",
    template_params: { actor: "alice", thread_title: "DNS Tunneling Notes" },
  };

  const resolvedTh = resolveNotificationLocale(commentNotif, "th");
  assert.ok(/[\u0E00-\u0E7F]/.test(resolvedTh.title), "Expected Thai notification title");
  assert.ok(resolvedTh.message.includes("alice"), "Expected actor interpolation");
  assert.ok(resolvedTh.message.includes("DNS Tunneling Notes"), "Expected verbatim thread title");

  const resolvedEn = resolveNotificationLocale(commentNotif, "en");
  assert.equal(resolvedEn.title, "New reply to your comment");
  assert.equal(resolvedEn.message, "alice replied to your comment on 'DNS Tunneling Notes'");

  const legacyNotif = {
    title: "Legacy alert",
    message: "Pre-Spec-11 notification",
  };
  assert.deepEqual(resolveNotificationLocale(legacyNotif, "th"), {
    title: "Legacy alert",
    message: "Pre-Spec-11 notification",
  });
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

  // Pre-submission sanitized payload (no tp_fp_explanation / no findings / no action explanation)
  const sanitized = {
    title: "MFA Push Fatigue",
    description: "User ajones rejected multiple MFA prompts.",
    initial_logs: initialLogs,
    translations: {
      th: {
        title: "การโจมตี MFA Push Fatigue",
        description: "พบผู้ใช้ ajones ปฏิเสธคำขอ MFA หลายครั้ง",
        playbook_steps: {
          c1: { label: "เพิกถอน tokens" },
        },
      },
    },
  };
  const resolvedSanitized = resolveScenarioLocale(sanitized, "th");
  assert.equal(resolvedSanitized.title, "การโจมตี MFA Push Fatigue");
  assert.equal(resolvedSanitized.playbookLabel("c1", "Revoke tokens"), "เพิกถอน tokens");
  assert.equal(resolvedSanitized.tp_fp_explanation, "");
  assert.equal(resolvedSanitized.findingDescription("f1", "Fallback finding"), "Fallback finding");
});

test("resolveGuestDefaultLocale falls back to configured default only when explicit choice is absent", () => {
  assert.equal(resolveGuestDefaultLocale(null, "th"), "th");
  assert.equal(resolveGuestDefaultLocale(null, "TH"), "th");
  assert.equal(resolveGuestDefaultLocale("", "th"), "th");
  assert.equal(resolveGuestDefaultLocale(undefined, undefined), "th");
  assert.equal(resolveGuestDefaultLocale(null, "en"), "en");
  assert.equal(resolveGuestDefaultLocale("en", "th"), "en");
  assert.equal(resolveGuestDefaultLocale("th", "en"), "th");
});
