import type { en } from "./en.ts";

export type Locale = "en" | "th";

export type TranslationParams = Record<string, string | number>;

type WidenStrings<T> = {
  readonly [K in keyof T]: T[K] extends string ? string : WidenStrings<T[K]>;
};

export type TranslationDictionary = WidenStrings<typeof en>;

type DotPaths<T, Prefix extends string = ""> = {
  [K in keyof T & string]: T[K] extends string
    ? `${Prefix}${K}`
    : DotPaths<T[K], `${Prefix}${K}.`>;
}[keyof T & string];

export type TranslationKey = DotPaths<typeof en>;

export interface ApiErrorLike {
  response?: {
    data?: {
      error?: string;
      message?: string;
      code?: string;
      params?: Record<string, unknown>;
    };
  };
}

export interface ScenarioTranslationPlaybookEntry {
  label?: string;
  explanation?: string;
}

export interface ScenarioTranslationFindingEntry {
  description?: string;
  explanation?: string;
}

export interface ScenarioTranslationPayload {
  title?: string;
  description?: string;
  tp_fp_explanation?: string;
  playbook_steps?: Record<string, ScenarioTranslationPlaybookEntry>;
  findings?: Record<string, ScenarioTranslationFindingEntry>;
}

export interface LocalizableScenario {
  title?: string;
  description?: string;
  tp_fp_explanation?: string;
  initial_logs?: readonly Record<string, unknown>[];
  translations?: {
    th?: ScenarioTranslationPayload;
    [locale: string]: ScenarioTranslationPayload | undefined;
  };
}

export interface LocalizableNotification {
  title?: string;
  message?: string;
  template_key?: string;
  template_params?: Record<string, unknown>;
}
