import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import axios from "axios";
import { formatDate as formatGregorianDate, normalizeLocale, translate } from "./format.ts";
import type { Locale, TranslationKey, TranslationParams } from "./types.ts";

const STORAGE_KEY = "cybersim_lang";

function readStoredLocale(): Locale {
  if (typeof window === "undefined") {
    return "en";
  }
  try {
    return normalizeLocale(window.localStorage.getItem(STORAGE_KEY));
  } catch {
    return "en";
  }
}

let activeLocale: Locale = readStoredLocale();

if (typeof document !== "undefined") {
  document.documentElement.lang = activeLocale;
}

axios.interceptors.request.use((config) => {
  config.headers = config.headers ?? {};
  config.headers["Accept-Language"] = activeLocale;
  return config;
});

interface I18nContextValue {
  locale: Locale;
  setLocale: (next: Locale) => void;
  syncLocaleFromUser: (serverLang?: string | null) => void;
  t: (key: TranslationKey | (string & {}), params?: TranslationParams) => string;
  formatDate: (iso: string | Date | number, options?: Intl.DateTimeFormatOptions) => string;
}

const I18nContext = createContext<I18nContextValue | null>(null);

export function I18nProvider({ children }: { children: ReactNode }) {
  const [locale, setLocaleState] = useState<Locale>(() => activeLocale);

  useEffect(() => {
    activeLocale = locale;
    if (typeof document !== "undefined") {
      document.documentElement.lang = locale;
    }
  }, [locale]);

  const applyLocaleLocally = useCallback((next: Locale) => {
    const normalized = normalizeLocale(next);
    activeLocale = normalized;
    if (typeof document !== "undefined") {
      document.documentElement.lang = normalized;
    }
    if (typeof window !== "undefined") {
      try {
        window.localStorage.setItem(STORAGE_KEY, normalized);
      } catch {
        // Ignore storage quota errors
      }
    }
    setLocaleState(normalized);
    return normalized;
  }, []);

  const setLocale = useCallback(
    (next: Locale) => {
      const normalized = applyLocaleLocally(next);
      if (typeof window !== "undefined") {
        const token = window.localStorage.getItem("token");
        if (token) {
          axios
            .put(
              "/api/user/language",
              { language: normalized },
              { headers: { Authorization: `Bearer ${token}` } },
            )
            .catch((err) => {
              console.warn("Failed to persist language preference:", err);
            });
        }
      }
    },
    [applyLocaleLocally],
  );

  const syncLocaleFromUser = useCallback(
    (serverLang?: string | null) => {
      if (serverLang === "en" || serverLang === "th") {
        applyLocaleLocally(serverLang);
      }
    },
    [applyLocaleLocally],
  );

  const t = useCallback(
    (key: TranslationKey | (string & {}), params?: TranslationParams) =>
      translate(locale, key, params),
    [locale],
  );

  const formatDate = useCallback(
    (iso: string | Date | number, options?: Intl.DateTimeFormatOptions) =>
      formatGregorianDate(iso, locale, options),
    [locale],
  );

  const value = useMemo<I18nContextValue>(
    () => ({ locale, setLocale, syncLocaleFromUser, t, formatDate }),
    [locale, setLocale, syncLocaleFromUser, t, formatDate],
  );

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

// eslint-disable-next-line react-refresh/only-export-components
export function useI18n(): I18nContextValue {
  const ctx = useContext(I18nContext);
  if (!ctx) {
    throw new Error("useI18n must be used within an I18nProvider");
  }
  return ctx;
}
