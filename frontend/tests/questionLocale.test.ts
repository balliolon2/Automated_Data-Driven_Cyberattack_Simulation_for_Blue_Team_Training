import test from "node:test";
import assert from "node:assert/strict";
import { resolveLocalizedQuestion, type LocalizableQuestion } from "../src/lib/questionLocale.ts";

const sampleQuestion: LocalizableQuestion = {
  question_text: "What enforces Least Privilege in IAM?",
  options: {
    A: "Granting admin rights to every user",
    B: "Assigning only permissions needed for a job role",
    C: "Disabling MFA",
    D: "Sharing root credentials",
  },
  explanation: "Correct Answer: B. Least Privilege minimizes blast radius.",
  translations: {
    th: {
      question_text: "หลักการใดที่ใช้บังคับ Least Privilege ในระบบ IAM?",
      options: {
        A: "ให้สิทธิ์ Admin แก่ผู้ใช้ทุกคน",
        B: "กำหนดสิทธิ์เท่าที่จำเป็นต่อหน้าที่งานเท่านั้น",
        C: "ปิดการใช้งาน MFA",
        D: "ใช้บัญชี root ร่วมกัน",
      },
      explanation: "คำตอบที่ถูกต้อง: B. Least Privilege ช่วยจำกัดขอบเขตความเสียหาย",
    },
  },
};

test("resolveLocalizedQuestion returns Thai content when locale is th and falls back cleanly for EN or override", () => {
  const thView = resolveLocalizedQuestion(sampleQuestion, "th");
  assert.equal(thView.questionText, "หลักการใดที่ใช้บังคับ Least Privilege ในระบบ IAM?");
  assert.equal(thView.options.B, "กำหนดสิทธิ์เท่าที่จำเป็นต่อหน้าที่งานเท่านั้น");
  assert.equal(thView.explanation, "คำตอบที่ถูกต้อง: B. Least Privilege ช่วยจำกัดขอบเขตความเสียหาย");
  assert.equal(thView.hasThaiTranslation, true);
  assert.equal(thView.showingOriginalEn, false);

  const toggledEn = resolveLocalizedQuestion(sampleQuestion, "th", true);
  assert.equal(toggledEn.questionText, "What enforces Least Privilege in IAM?");
  assert.equal(toggledEn.options.B, "Assigning only permissions needed for a job role");
  assert.equal(toggledEn.explanation, "Correct Answer: B. Least Privilege minimizes blast radius.");
  assert.equal(toggledEn.hasThaiTranslation, true);
  assert.equal(toggledEn.showingOriginalEn, true);

  const enLocale = resolveLocalizedQuestion(sampleQuestion, "en", false);
  assert.equal(enLocale.questionText, "What enforces Least Privilege in IAM?");
  assert.equal(enLocale.showingOriginalEn, false);
});

test("resolveLocalizedQuestion gracefully falls back to canonical EN when translations are absent or explanation stripped", () => {
  const untranslated: LocalizableQuestion = {
    question_text: "Canonical EN only",
    options: { A: "One", B: "Two" },
  };
  const fallbackView = resolveLocalizedQuestion(untranslated, "th");
  assert.equal(fallbackView.questionText, "Canonical EN only");
  assert.equal(fallbackView.options.A, "One");
  assert.equal(fallbackView.hasThaiTranslation, false);
  assert.equal(fallbackView.showingOriginalEn, false);

  const sanitizedActive: LocalizableQuestion = {
    question_text: "Canonical EN active",
    options: { A: "One", B: "Two" },
    translations: {
      th: {
        question_text: "คำถามที่กำลังสอบ",
        options: { A: "หนึ่ง", B: "สอง" },
      },
    },
  };
  const activeView = resolveLocalizedQuestion(sanitizedActive, "th");
  assert.equal(activeView.questionText, "คำถามที่กำลังสอบ");
  assert.equal(activeView.options.B, "สอง");
  assert.equal(activeView.explanation, "");
  assert.equal(activeView.hasThaiTranslation, true);

  const nullView = resolveLocalizedQuestion(null, "th", true);
  assert.deepEqual(nullView, {
    questionText: "",
    options: {},
    explanation: "",
    hasThaiTranslation: false,
    showingOriginalEn: false,
  });
});
