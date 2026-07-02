"use client";

import { useTranslation } from "@/hooks";
import clsx from "clsx";

export function LanguageSwitcher({ className }: { className?: string }) {
  const { locale, setLocale } = useTranslation();

  return (
    <div className={clsx("inline-flex items-center rounded-full bg-surface p-0.5 text-xs font-semibold", className)}>
      {(["vi", "en"] as const).map((option) => (
        <button
          key={option}
          type="button"
          onClick={() => setLocale(option)}
          className={clsx(
            "rounded-full px-2.5 py-1 uppercase transition-colors",
            locale === option ? "bg-background text-primary shadow-card" : "text-muted-foreground"
          )}
        >
          {option}
        </button>
      ))}
    </div>
  );
}
