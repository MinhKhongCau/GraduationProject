"use client";

import { useLocaleContext } from "@/context/LocaleContext";
import { LOCALE_DICTS } from "@/locales";

function resolvePath(dict: Record<string, unknown>, path: string): unknown {
  return path.split(".").reduce<unknown>((node, segment) => {
    if (node && typeof node === "object") {
      return (node as Record<string, unknown>)[segment];
    }
    return undefined;
  }, dict);
}

function interpolate(template: string, vars?: Record<string, string | number>): string {
  if (!vars) return template;
  return template.replace(/\{\{(\w+)\}\}/g, (match, key) =>
    key in vars ? String(vars[key]) : match
  );
}

export function useTranslation() {
  const { locale, setLocale } = useLocaleContext();

  /**
   * `defaultText` is required (not optional) so pages read correctly in
   * English immediately, even before locales/*.json has a translation for
   * `key` — see DESIGN.md "i18n". Once a dictionary entry exists it wins.
   */
  function t(key: string, defaultText: string, vars?: Record<string, string | number>): string {
    const dict = LOCALE_DICTS[locale] as Record<string, unknown>;
    const value = resolvePath(dict, key);
    return interpolate(typeof value === "string" ? value : defaultText, vars);
  }

  return { t, locale, setLocale };
}
