import { en } from "./en.ts";
import { th } from "./th.ts";
import type {
  ApiErrorLike,
  LocalizableNotification,
  LocalizableScenario,
  Locale,
  ScenarioTranslationPayload,
  TranslationDictionary,
  TranslationParams,
} from "./types.ts";

const DICTIONARIES: Record<Locale, TranslationDictionary> = { en, th };

function parseLocaleToken(raw: unknown): Locale | null {
  const cleaned = typeof raw === "string" ? raw.trim().toLowerCase() : "";
  if (cleaned === "en" || cleaned === "th") {
    return cleaned;
  }
  return null;
}

export function normalizeLocale(raw: unknown): Locale {
  return raw === "th" ? "th" : "en";
}

/** Resolves initial guest locale from storage, falling back to normalized configuredDefault ("th" when unset). */
export function resolveGuestDefaultLocale(
  stored: unknown,
  configuredDefault?: unknown,
): Locale {
  const explicitStored = parseLocaleToken(stored);
  if (explicitStored !== null) {
    return explicitStored;
  }
  const hasConfigured =
    configuredDefault !== undefined &&
    configuredDefault !== null &&
    String(configuredDefault).trim().length > 0;
  if (!hasConfigured) {
    return "th";
  }
  return normalizeLocale(String(configuredDefault).trim().toLowerCase());
}

function lookupPath(dict: Record<string, unknown>, path: string): string | undefined {
  const parts = path.split(".");
  let current: unknown = dict;
  for (const part of parts) {
    if (current === null || typeof current !== "object" || !(part in current)) {
      return undefined;
    }
    current = (current as Record<string, unknown>)[part];
  }
  return typeof current === "string" ? current : undefined;
}

export function interpolate(template: string, params?: TranslationParams): string {
  if (!params) return template;
  return template.replace(/\{\{\s*([a-zA-Z0-9_]+)\s*\}\}/g, (match, key: string) => {
    if (Object.prototype.hasOwnProperty.call(params, key)) {
      const value = params[key];
      return value === undefined || value === null ? match : String(value);
    }
    return match;
  });
}

export function translate(locale: Locale, key: string, params?: TranslationParams): string {
  const activeDict = DICTIONARIES[normalizeLocale(locale)] as unknown as Record<string, unknown>;
  const fallbackDict = DICTIONARIES.en as unknown as Record<string, unknown>;
  const template = lookupPath(activeDict, key) ?? lookupPath(fallbackDict, key);
  if (template === undefined) {
    return key;
  }
  return interpolate(template, params);
}

export function formatDate(
  iso: string | Date | number,
  locale: Locale,
  options: Intl.DateTimeFormatOptions = { year: "numeric", month: "short", day: "numeric" },
): string {
  const date = iso instanceof Date ? iso : new Date(iso);
  if (Number.isNaN(date.getTime())) {
    return typeof iso === "string" ? iso : String(iso);
  }
  const intlLocale = normalizeLocale(locale) === "th" ? "th-TH-u-ca-gregory" : "en-US";
  return new Intl.DateTimeFormat(intlLocale, options).format(date);
}

function sanitizeParams(raw?: Record<string, unknown>): TranslationParams | undefined {
  if (!raw || typeof raw !== "object") return undefined;
  const out: TranslationParams = {};
  for (const [k, v] of Object.entries(raw)) {
    if (typeof v === "string" || typeof v === "number") {
      out[k] = v;
    }
  }
  return out;
}

export function resolveApiError(
  err: unknown,
  t: (key: string, params?: TranslationParams) => string,
  fallback?: string,
): string {
  const apiErr = err as ApiErrorLike | undefined;
  const data = apiErr?.response?.data;
  const code = typeof data?.code === "string" ? data.code.trim() : "";

  if (code) {
    const errorKey = `errors.${code}`;
    const translated = t(errorKey, sanitizeParams(data?.params));
    if (translated !== errorKey) {
      return translated;
    }
  }

  if (typeof data?.error === "string" && data.error.trim().length > 0) {
    return data.error;
  }
  if (typeof data?.message === "string" && data.message.trim().length > 0) {
    return data.message;
  }
  if (fallback && fallback.trim().length > 0) {
    return fallback;
  }
  return t("errors.INTERNAL_SERVER_ERROR");
}

export function resolveDomainName(
  domainId: string,
  fallbackName: string,
  locale: Locale,
): string {
  const dict = DICTIONARIES[normalizeLocale(locale)].domains as Record<string, string>;
  if (Object.prototype.hasOwnProperty.call(dict, domainId)) {
    return dict[domainId];
  }
  return fallbackName || domainId;
}

export function resolveScenarioLocale(
  scenario: LocalizableScenario | null | undefined,
  locale: Locale,
): {
  title: string;
  description: string;
  tp_fp_explanation: string;
  playbookLabel: (actionId: string, fallback: string) => string;
  playbookExplanation: (actionId: string, fallback: string) => string;
  findingDescription: (findingId: string, fallback: string) => string;
  findingExplanation: (findingId: string, fallback: string) => string;
} {
  const activeLocale = normalizeLocale(locale);
  const tr: ScenarioTranslationPayload | undefined =
    activeLocale === "en" ? undefined : scenario?.translations?.[activeLocale];

  return {
    title: tr?.title?.trim() ? tr.title : scenario?.title ?? "",
    description: tr?.description?.trim() ? tr.description : scenario?.description ?? "",
    tp_fp_explanation: tr?.tp_fp_explanation?.trim()
      ? tr.tp_fp_explanation
      : scenario?.tp_fp_explanation ?? "",
    playbookLabel: (actionId: string, fallback: string) =>
      tr?.playbook_steps?.[actionId]?.label?.trim() || fallback,
    playbookExplanation: (actionId: string, fallback: string) =>
      tr?.playbook_steps?.[actionId]?.explanation?.trim() || fallback,
    findingDescription: (findingId: string, fallback: string) =>
      tr?.findings?.[findingId]?.description?.trim() || fallback,
    findingExplanation: (findingId: string, fallback: string) =>
      tr?.findings?.[findingId]?.explanation?.trim() || fallback,
  };
}

export function resolveNotificationLocale(
  notif: LocalizableNotification | null | undefined,
  locale: Locale,
): { title: string; message: string } {
  const activeLocale = normalizeLocale(locale);
  const templateKey = typeof notif?.template_key === "string" ? notif.template_key.trim() : "";
  if (templateKey) {
    const titleKey = `notifications.templates.${templateKey}.title`;
    const msgKey = `notifications.templates.${templateKey}.message`;
    const params = sanitizeParams(notif?.template_params);
    const localizedTitle = translate(activeLocale, titleKey, params);
    const localizedMsg = translate(activeLocale, msgKey, params);
    if (localizedTitle !== titleKey && localizedMsg !== msgKey) {
      return { title: localizedTitle, message: localizedMsg };
    }
  }
  return {
    title: notif?.title ?? "",
    message: notif?.message ?? "",
  };
}
