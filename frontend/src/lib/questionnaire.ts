export type CSUQSubscale = "SYSUSE" | "INFOQUAL" | "INTERQUAL" | "OVERALL";
export type CSUQAnswerValue = number | null; // 1..7 or null for N/A
export type NASATLXKey =
  | "mental_demand"
  | "physical_demand"
  | "temporal_demand"
  | "performance_rating"
  | "effort"
  | "frustration";

export interface CSUQItemSpec {
  key: string;
  number: number;
  subscale: CSUQSubscale;
  en: string;
  th: string;
}

export interface NASATLXDimSpec {
  key: NASATLXKey;
  invertForWorkload: boolean;
  en: {
    label: string;
    prompt: string;
    lowAnchor: string;
    highAnchor: string;
  };
  th: {
    label: string;
    prompt: string;
    lowAnchor: string;
    highAnchor: string;
  };
}

export interface QualitativeDraft {
  most_impressive: string;
  areas_to_improve: string;
  additional_comments: string;
}

export interface QuestionnaireDraft {
  step: 1 | 2 | 3;
  csuq: Record<string, CSUQAnswerValue | undefined>;
  tlx: Record<NASATLXKey, number>;
  tlxConfirmed: Record<NASATLXKey, boolean>;
  qualitative: QualitativeDraft;
}

export interface CSUQScoresPayload {
  schema_version: string;
  scale_min: number;
  scale_max: number;
  direction: string;
  locale: string;
  items: Record<string, number | null>;
  subscales: {
    sysuse: number | null;
    infoqual: number | null;
    interqual: number | null;
    overall: number | null;
  };
  valid_counts: {
    sysuse: number;
    infoqual: number;
    interqual: number;
    overall: number;
  };
}

export interface NASATLXScoresPayload {
  schema_version: string;
  variant: string;
  scale_min: number;
  scale_max: number;
  step: number;
  dimensions: {
    mental_demand: number;
    physical_demand: number;
    temporal_demand: number;
    performance_rating: number;
    performance_workload: number;
    effort: number;
    frustration: number;
  };
  overall_workload: number;
}

export interface EvaluationMetricSummary {
  eval_id: string;
  pre_score: number;
  post_score: number;
  improvement_pct: number;
  submitted_locale: string;
  completed_scenarios: number;
  average_proficiency: number;
  csuq_scores: CSUQScoresPayload;
  nasa_tlx_scores: NASATLXScoresPayload;
  qualitative_feedback: QualitativeDraft;
  notes: string;
  eval_date: string;
}

export interface QuestionnaireStatusResponse {
  role: string;
  is_preview_only: boolean;
  has_completed_post_test: boolean;
  latest_post_session_id: string | null;
  eligible: boolean;
  ineligible_reason: string;
  submitted: boolean;
  submission: EvaluationMetricSummary | null;
}

export const MAX_QUALITATIVE_CHARS = 2000;

export const CSUQ_ITEMS: readonly CSUQItemSpec[] = [
  {
    key: "q1",
    number: 1,
    subscale: "SYSUSE",
    en: "Overall, I am satisfied with how easy it is to use this system.",
    th: "โดยรวมแล้ว ฉันพึงพอใจกับความง่ายในการใช้งานระบบนี้",
  },
  {
    key: "q2",
    number: 2,
    subscale: "SYSUSE",
    en: "It was simple to use this system.",
    th: "ระบบนี้ใช้งานได้ง่ายและไม่ซับซ้อน",
  },
  {
    key: "q3",
    number: 3,
    subscale: "SYSUSE",
    en: "I was able to complete the tasks and scenarios quickly using this system.",
    th: "ฉันสามารถทำภารกิจและสถานการณ์จำลอง (Scenarios) ได้อย่างรวดเร็วผ่านระบบนี้",
  },
  {
    key: "q4",
    number: 4,
    subscale: "SYSUSE",
    en: "I felt comfortable using this system.",
    th: "ฉันรู้สึกสะดวกสบายและไม่ติดขัดในการใช้งานระบบนี้",
  },
  {
    key: "q5",
    number: 5,
    subscale: "SYSUSE",
    en: "It was easy to learn to use this system.",
    th: "การเรียนรู้วิธีใช้งานระบบนี้ทำได้ง่าย",
  },
  {
    key: "q6",
    number: 6,
    subscale: "SYSUSE",
    en: "I believe I could become productive quickly using this system.",
    th: "ฉันเชื่อว่าฉันสามารถใช้งานระบบนี้ให้เกิดประสิทธิภาพได้อย่างรวดเร็ว",
  },
  {
    key: "q7",
    number: 7,
    subscale: "INFOQUAL",
    en: "The system gave error messages that clearly told me how to fix problems.",
    th: "ระบบแสดงข้อความแจ้งเตือนข้อผิดพลาด (Error Messages) ที่บอกวิธีแก้ไขปัญหาได้อย่างชัดเจน",
  },
  {
    key: "q8",
    number: 8,
    subscale: "INFOQUAL",
    en: "Whenever I made a mistake using the system, I could recover easily and quickly.",
    th: "เมื่อฉันทำผิดพลาดระหว่างใช้งานระบบ ฉันสามารถแก้ไขหรือย้อนกลับได้ง่ายและรวดเร็ว",
  },
  {
    key: "q9",
    number: 9,
    subscale: "INFOQUAL",
    en: "The information (such as online help, on-screen messages, and other documentation) provided with this system was clear.",
    th: "ข้อมูลที่ระบบแสดง (เช่น คำแนะนำบนหน้าจอ บทเรียน และคำอธิบายผล) มีความชัดเจน",
  },
  {
    key: "q10",
    number: 10,
    subscale: "INFOQUAL",
    en: "It was easy to find the information I needed.",
    th: "ฉันสามารถค้นหาข้อมูลที่ต้องการภายในระบบได้ง่าย",
  },
  {
    key: "q11",
    number: 11,
    subscale: "INFOQUAL",
    en: "The information was effective in helping me complete the tasks and scenarios.",
    th: "ข้อมูลในระบบช่วยให้ฉันทำภารกิจและสถานการณ์จำลองสำเร็จได้อย่างมีประสิทธิภาพ",
  },
  {
    key: "q12",
    number: 12,
    subscale: "INFOQUAL",
    en: "The organization of information on the system screens was clear.",
    th: "การจัดวางโครงสร้างข้อมูลบนหน้าจอของระบบมีความเป็นระเบียบและชัดเจน",
  },
  {
    key: "q13",
    number: 13,
    subscale: "INTERQUAL",
    en: "The interface of this system was pleasant.",
    th: "ส่วนติดต่อผู้ใช้ (Interface) ของระบบนี้มีความสวยงามและน่าใช้งาน",
  },
  {
    key: "q14",
    number: 14,
    subscale: "INTERQUAL",
    en: "I liked using the interface of this system.",
    th: "ฉันชอบใช้งานส่วนติดต่อผู้ใช้ (Interface) ของระบบนี้",
  },
  {
    key: "q15",
    number: 15,
    subscale: "INTERQUAL",
    en: "This system has all the functions and capabilities I expect it to have.",
    th: "ระบบนี้มีฟังก์ชันและความสามารถครบถ้วนตามที่ฉันคาดหวัง",
  },
  {
    key: "q16",
    number: 16,
    subscale: "OVERALL",
    en: "Overall, I am satisfied with this system.",
    th: "โดยรวมแล้ว ฉันพึงพอใจกับแพลตฟอร์มนี้",
  },
] as const;

export const NASA_TLX_DIMENSIONS: readonly NASATLXDimSpec[] = [
  {
    key: "mental_demand",
    invertForWorkload: false,
    en: {
      label: "Mental Demand",
      prompt: "How mentally demanding was the training and investigation?",
      lowAnchor: "Very Low",
      highAnchor: "Very High",
    },
    th: {
      label: "ภาระทางความคิด (Mental Demand)",
      prompt: "การฝึกอบรมและการสืบสวนสถานการณ์จำลองต้องใช้ความคิดและการวิเคราะห์มากเพียงใด?",
      lowAnchor: "น้อยมาก",
      highAnchor: "สูงมาก",
    },
  },
  {
    key: "physical_demand",
    invertForWorkload: false,
    en: {
      label: "Physical Demand",
      prompt: "How physically demanding was using the platform?",
      lowAnchor: "Very Low",
      highAnchor: "Very High",
    },
    th: {
      label: "ภาระทางร่างกาย (Physical Demand)",
      prompt: "การใช้งานแพลตฟอร์มนี้ทำให้เกิดความเหนื่อยล้าทางร่างกายมากเพียงใด?",
      lowAnchor: "น้อยมาก",
      highAnchor: "สูงมาก",
    },
  },
  {
    key: "temporal_demand",
    invertForWorkload: false,
    en: {
      label: "Temporal Demand",
      prompt: "How hurried or rushed was the pace of the tasks and exams?",
      lowAnchor: "Very Low",
      highAnchor: "Very High",
    },
    th: {
      label: "แรงกดดันด้านเวลา (Temporal Demand)",
      prompt: "คุณรู้สึกเร่งรีบหรือถูกกดดันเรื่องเวลาในการทำข้อสอบและภารกิจมากเพียงใด?",
      lowAnchor: "น้อยมาก",
      highAnchor: "สูงมาก",
    },
  },
  {
    key: "performance_rating",
    invertForWorkload: true,
    en: {
      label: "Performance",
      prompt: "How successful were you in accomplishing what you were asked to do?",
      lowAnchor: "Failure (Poor)",
      highAnchor: "Perfect (Excellent)",
    },
    th: {
      label: "ประสิทธิภาพของตนเอง (Performance)",
      prompt: "คุณคิดว่าตนเองทำภารกิจและการทดสอบได้สำเร็จมากน้อยเพียงใด?",
      lowAnchor: "ไม่สำเร็จ (แย่มาก)",
      highAnchor: "สำเร็จสมบูรณ์ (ดีเยี่ยม)",
    },
  },
  {
    key: "effort",
    invertForWorkload: false,
    en: {
      label: "Effort",
      prompt: "How hard did you have to work to accomplish your level of performance?",
      lowAnchor: "Very Low",
      highAnchor: "Very High",
    },
    th: {
      label: "ความพยายามที่ใช้ (Effort)",
      prompt: "คุณต้องใช้ความพยายามมากเพียงใดเพื่อให้ได้ผลลัพธ์ระดับนี้?",
      lowAnchor: "น้อยมาก",
      highAnchor: "สูงมาก",
    },
  },
  {
    key: "frustration",
    invertForWorkload: false,
    en: {
      label: "Frustration",
      prompt: "How insecure, discouraged, irritated, stressed, and annoyed were you?",
      lowAnchor: "Very Low",
      highAnchor: "Very High",
    },
    th: {
      label: "ความหงุดหงิดหรือเครียด (Frustration)",
      prompt: "คุณรู้สึกสับสน หงุดหงิด เครียด หรือท้อแท้มากเพียงใดระหว่างใช้งาน?",
      lowAnchor: "น้อยมาก",
      highAnchor: "สูงมาก",
    },
  },
] as const;

export function round2(n: number): number {
  return Math.round(n * 100) / 100;
}

export function getQuestionnaireDraftStorageKey(userId: string): string {
  const clean = userId.trim();
  return `cybersim_questionnaire_draft:${clean || "anon"}`;
}

export function createEmptyQuestionnaireDraft(): QuestionnaireDraft {
  return {
    step: 1,
    csuq: {},
    tlx: {
      mental_demand: 50,
      physical_demand: 50,
      temporal_demand: 50,
      performance_rating: 50,
      effort: 50,
      frustration: 50,
    },
    tlxConfirmed: {
      mental_demand: false,
      physical_demand: false,
      temporal_demand: false,
      performance_rating: false,
      effort: false,
      frustration: false,
    },
    qualitative: {
      most_impressive: "",
      areas_to_improve: "",
      additional_comments: "",
    },
  };
}

export function serializeQuestionnaireDraft(draft: QuestionnaireDraft): string {
  return JSON.stringify(draft);
}

export function parseQuestionnaireDraft(raw: string | null): QuestionnaireDraft {
  const base = createEmptyQuestionnaireDraft();
  if (!raw) return base;
  try {
    const parsed = JSON.parse(raw) as Partial<QuestionnaireDraft>;
    if (parsed.step === 1 || parsed.step === 2 || parsed.step === 3) {
      base.step = parsed.step;
    }
    if (parsed.csuq && typeof parsed.csuq === "object") {
      for (const item of CSUQ_ITEMS) {
        const val = parsed.csuq[item.key];
        if (val === null) {
          base.csuq[item.key] = null;
        } else if (typeof val === "number" && Number.isInteger(val) && val >= 1 && val <= 7) {
          base.csuq[item.key] = val;
        }
      }
    }
    if (parsed.tlx && typeof parsed.tlx === "object") {
      for (const dim of NASA_TLX_DIMENSIONS) {
        const v = parsed.tlx[dim.key];
        if (typeof v === "number" && Number.isInteger(v) && v >= 0 && v <= 100 && v % 5 === 0) {
          base.tlx[dim.key] = v;
        }
      }
    }
    if (parsed.tlxConfirmed && typeof parsed.tlxConfirmed === "object") {
      for (const dim of NASA_TLX_DIMENSIONS) {
        if (typeof parsed.tlxConfirmed[dim.key] === "boolean") {
          base.tlxConfirmed[dim.key] = parsed.tlxConfirmed[dim.key];
        }
      }
    }
    if (parsed.qualitative && typeof parsed.qualitative === "object") {
      base.qualitative = {
        most_impressive: typeof parsed.qualitative.most_impressive === "string" ? parsed.qualitative.most_impressive : "",
        areas_to_improve: typeof parsed.qualitative.areas_to_improve === "string" ? parsed.qualitative.areas_to_improve : "",
        additional_comments: typeof parsed.qualitative.additional_comments === "string" ? parsed.qualitative.additional_comments : "",
      };
    }
    return base;
  } catch {
    return createEmptyQuestionnaireDraft();
  }
}

export function countCSUQAnswered(csuq: Record<string, CSUQAnswerValue | undefined>): {
  answered: number;
  numericCount: number;
} {
  let answered = 0;
  let numericCount = 0;
  for (const item of CSUQ_ITEMS) {
    if (Object.prototype.hasOwnProperty.call(csuq, item.key) && csuq[item.key] !== undefined) {
      answered++;
      if (typeof csuq[item.key] === "number") {
        numericCount++;
      }
    }
  }
  return { answered, numericCount };
}

export function canAdvanceCSUQStep(csuq: Record<string, CSUQAnswerValue | undefined>): boolean {
  const { answered, numericCount } = countCSUQAnswered(csuq);
  return answered === CSUQ_ITEMS.length && numericCount >= 1;
}

export function canAdvanceTLXStep(tlxConfirmed: Record<NASATLXKey, boolean>): boolean {
  return NASA_TLX_DIMENSIONS.every((d) => tlxConfirmed[d.key] === true);
}

function meanForKeys(csuq: Record<string, CSUQAnswerValue | undefined>, keys: string[]): number | null {
  let sum = 0;
  let count = 0;
  for (const key of keys) {
    const v = csuq[key];
    if (typeof v === "number") {
      sum += v;
      count++;
    }
  }
  return count === 0 ? null : round2(sum / count);
}

export function computeCSUQScoresPreview(csuq: Record<string, CSUQAnswerValue | undefined>): {
  sysuse: number | null;
  infoqual: number | null;
  interqual: number | null;
  overall: number | null;
  validOverallCount: number;
} {
  const { numericCount } = countCSUQAnswered(csuq);
  return {
    sysuse: meanForKeys(csuq, ["q1", "q2", "q3", "q4", "q5", "q6"]),
    infoqual: meanForKeys(csuq, ["q7", "q8", "q9", "q10", "q11", "q12"]),
    interqual: meanForKeys(csuq, ["q13", "q14", "q15"]),
    overall: meanForKeys(csuq, CSUQ_ITEMS.map((i) => i.key)),
    validOverallCount: numericCount,
  };
}

export function computeNASATLXPreview(tlx: Record<NASATLXKey, number>): {
  performanceWorkload: number;
  overallWorkload: number;
} {
  const performanceWorkload = 100 - tlx.performance_rating;
  const total =
    tlx.mental_demand +
    tlx.physical_demand +
    tlx.temporal_demand +
    performanceWorkload +
    tlx.effort +
    tlx.frustration;
  return {
    performanceWorkload,
    overallWorkload: round2(total / 6),
  };
}

export function countQualitativeChars(text: string): number {
  return Array.from(text).length;
}

export function isQualitativeStepValid(q: QualitativeDraft): boolean {
  return (
    countQualitativeChars(q.most_impressive.trim()) <= MAX_QUALITATIVE_CHARS &&
    countQualitativeChars(q.areas_to_improve.trim()) <= MAX_QUALITATIVE_CHARS &&
    countQualitativeChars(q.additional_comments.trim()) <= MAX_QUALITATIVE_CHARS
  );
}
