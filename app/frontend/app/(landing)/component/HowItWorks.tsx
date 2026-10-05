"use client";

import { HOW_IT_WORKS_STEPS } from "@/constants";
import { useTranslation } from "@/hooks";
import { Card } from "@/components/ui";

export function HowItWorks() {
  const { t } = useTranslation();

  return (
    <section className="px-4 py-16 sm:px-6 sm:py-20">
      <div className="mx-auto max-w-6xl">
        <h2 className="mb-12 text-center text-3xl font-bold tracking-tight text-foreground">How MindCare works</h2>
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
          {HOW_IT_WORKS_STEPS.map((step, index) => (
            <Card key={step.id} className="p-6">
              <div className="mb-4 flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-sm font-bold text-white">
                {index + 1}
              </div>
              <h3 className="mb-2 font-semibold text-foreground">{t(step.titleKey, step.title)}</h3>
              <p className="text-sm text-muted-foreground">{t(step.descriptionKey, step.description)}</p>
            </Card>
          ))}
        </div>
      </div>
    </section>
  );
}
