"use client";

import { Construction } from "lucide-react";
import { useTranslation } from "@/hooks";

export function ComingSoon({ title }: { title: string }) {
  const { t } = useTranslation();

  return (
    <div className="flex min-h-[50vh] flex-col items-center justify-center gap-3 text-center">
      <span className="flex h-16 w-16 items-center justify-center rounded-2xl bg-surface">
        <Construction className="h-8 w-8 text-muted-foreground" aria-hidden="true" />
      </span>
      <h1 className="text-xl font-bold text-foreground">{title}</h1>
      <p className="max-w-sm text-sm text-muted-foreground">
        {t("common.comingSoon", "Coming soon")} — {t("common.comingSoonDescription", "This page is planned but not built in this pass. See README for the full roadmap.")}
      </p>
    </div>
  );
}
