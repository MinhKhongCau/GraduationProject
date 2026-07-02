"use client";

import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { LOCALE_COOKIE, STORAGE_KEYS } from "@/constants/api";

export type Locale = "en" | "vi";

interface LocaleContextValue {
  locale: Locale;
  setLocale: (locale: Locale) => void;
}

const LocaleContext = createContext<LocaleContextValue | null>(null);

function readCookie(name: string): string | null {
  if (typeof document === "undefined") return null;
  const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`));
  return match ? decodeURIComponent(match[1]) : null;
}

function detectInitialLocale(): Locale {
  if (typeof window === "undefined") return "vi";
  const fromCookie = readCookie(LOCALE_COOKIE);
  if (fromCookie === "en" || fromCookie === "vi") return fromCookie;
  const fromStorage = window.localStorage.getItem(STORAGE_KEYS.LOCALE);
  if (fromStorage === "en" || fromStorage === "vi") return fromStorage;
  return window.navigator.language.toLowerCase().startsWith("en") ? "en" : "vi";
}

export interface LocaleProviderProps {
  children: ReactNode;
  /**
   * Read from the mc_locale cookie server-side in app/layout.tsx and
   * passed down, so SSR output already matches the visitor's saved
   * preference instead of always starting from a hardcoded default and
   * flashing to the right language after hydration.
   */
  initialLocale: Locale;
}

export function LocaleProvider({ children, initialLocale }: LocaleProviderProps) {
  const [locale, setLocaleState] = useState<Locale>(initialLocale);

  useEffect(() => {
    // Reconciles against localStorage for the one case the server can't
    // see: a first-ever visit with no cookie yet, where localStorage or
    // navigator.language may disagree with the server's fallback default.
    const detected = detectInitialLocale();
    if (detected !== locale) {
      setLocaleState(detected);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    document.documentElement.lang = locale;
  }, [locale]);

  const setLocale = (next: Locale) => {
    setLocaleState(next);
    window.localStorage.setItem(STORAGE_KEYS.LOCALE, next);
    document.cookie = `${LOCALE_COOKIE}=${next}; path=/; max-age=${60 * 60 * 24 * 365}; samesite=lax`;
  };

  const value = useMemo(() => ({ locale, setLocale }), [locale]);

  return <LocaleContext.Provider value={value}>{children}</LocaleContext.Provider>;
}

export function useLocaleContext(): LocaleContextValue {
  const context = useContext(LocaleContext);
  if (!context) {
    throw new Error("useLocaleContext must be used within a LocaleProvider");
  }
  return context;
}
