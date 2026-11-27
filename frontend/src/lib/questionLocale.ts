import type { Locale } from "../i18n/types.ts";

export interface QuestionTranslationEntry {
  question_text?: string;
  options?: Record<string, string>;
  explanation?: string;
}

export interface QuestionTranslationsMap {
  th?: QuestionTranslationEntry;
  [locale: string]: QuestionTranslationEntry | undefined;
}

export interface LocalizableQuestion {
  question_text: string;
  options?: Record<string, string> | null;
  explanation?: string | null;
  translations?: QuestionTranslationsMap | null;
}

export interface ResolvedQuestionContent {
  questionText: string;
  options: Record<string, string>;
  explanation: string;
  hasThaiTranslation: boolean;
  showingOriginalEn: boolean;
}

export function resolveLocalizedQuestion(
  question: LocalizableQuestion | null | undefined,
  locale: Locale,
  forceOriginalEn = false,
): ResolvedQuestionContent {
  if (!question) {
    return {
      questionText: "",
      options: {},
      explanation: "",
      hasThaiTranslation: false,
      showingOriginalEn: false,
    };
  }
  const canonicalOptions = question.options ?? {};
  const thEntry = question.translations?.th;
  const thStem = typeof thEntry?.question_text === "string" ? thEntry.question_text.trim() : "";
  const hasThaiTranslation = thStem.length > 0;

  if (locale === "th" && hasThaiTranslation && !forceOriginalEn) {
    const mergedOptions: Record<string, string> = {};
    const thOpts = thEntry?.options ?? {};
    const orderedKeys = Object.keys(canonicalOptions).length > 0 ? Object.keys(canonicalOptions) : Object.keys(thOpts);
    for (const key of orderedKeys) {
      const candidate = thOpts[key];
      mergedOptions[key] = typeof candidate === "string" && candidate.trim().length > 0 ? candidate : canonicalOptions[key] ?? "";
    }

    const thExp = typeof thEntry?.explanation === "string" && thEntry.explanation.trim().length > 0
      ? thEntry.explanation
      : (question.explanation ?? "");

    return {
      questionText: thEntry!.question_text!,
      options: mergedOptions,
      explanation: thExp,
      hasThaiTranslation,
      showingOriginalEn: false,
    };
  }

  return {
    questionText: question.question_text,
    options: canonicalOptions,
    explanation: question.explanation ?? "",
    hasThaiTranslation,
    showingOriginalEn: locale === "th" && hasThaiTranslation && forceOriginalEn,
  };
}
