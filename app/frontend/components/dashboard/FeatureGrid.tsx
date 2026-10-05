"use client";

import Link from "next/link";
import { ChevronRight } from "lucide-react";
import { Badge } from "@/components/ui";
import { useTranslation } from "@/hooks";
import type { FeatureGroup } from "@/constants/nav";

export interface FeatureGridProps {
  groups: FeatureGroup[];
}

/**
 * Dashboard entry point for every feature that isn't in the header nav —
 * the header stays short and the dashboard is where the rest is found.
 */
export function FeatureGrid({ groups }: FeatureGridProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-8">
      {groups.map((group) => (
        <section key={group.id} aria-labelledby={`feature-group-${group.id}`}>
          <h2
            id={`feature-group-${group.id}`}
            className="mb-3 text-sm font-semibold uppercase tracking-wide text-muted-foreground"
          >
            {t(group.labelKey, group.label)}
          </h2>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {group.items.map((item) => {
              const Icon = item.icon;
              return (
                <Link
                  key={item.id}
                  href={item.href}
                  className="group flex items-center gap-4 rounded-xl border border-border bg-background p-4 shadow-card transition-colors duration-150 hover:border-primary/40 hover:bg-primary-soft/40"
                >
                  <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-primary-soft text-primary-soft-text">
                    <Icon className="h-5 w-5" aria-hidden="true" />
                  </span>
                  <span className="min-w-0 flex-1">
                    <span className="flex items-center gap-2">
                      <span className="truncate text-sm font-semibold text-foreground">{t(item.labelKey, item.label)}</span>
                      {item.comingSoon && <Badge tone="neutral">{t("common.comingSoon", "Coming soon")}</Badge>}
                    </span>
                    {item.description && (
                      <span className="mt-0.5 block text-xs text-muted-foreground">
                        {t(item.descriptionKey ?? "", item.description)}
                      </span>
                    )}
                  </span>
                  <ChevronRight
                    className="h-4 w-4 shrink-0 text-muted-foreground transition-transform duration-150 group-hover:translate-x-0.5 group-hover:text-primary"
                    aria-hidden="true"
                  />
                </Link>
              );
            })}
          </div>
        </section>
      ))}
    </div>
  );
}
