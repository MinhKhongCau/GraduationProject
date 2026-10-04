"use client";

import { useTranslation } from "@/hooks";
import clsx from "clsx";

export function LanguageSwitcher({ className }: { className?: string }) {
  const { locale, setLocale } = useTranslation();

  return (
    <div className={clsx("inline-flex h-9 items-center rounded-lg border border-border bg-surface p-0.5 text-xs font-semibold", className)}>
      {(["vi", "en"] as const).map((option) => (
        <button
          key={option}
          type="button"
          onClick={() => setLocale(option)}
          aria-pressed={locale === option}
          className={clsx(
            "h-full rounded-md px-2.5 uppercase transition-colors",
            locale === option ? "bg-background text-primary shadow-card" : "text-muted-foreground hover:text-foreground"
          )}
        >
          {option}
        </button>
      ))}
    </div>
  );
}
