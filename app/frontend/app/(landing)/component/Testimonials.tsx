"use client";

import { Quote } from "lucide-react";
import { TESTIMONIALS } from "@/constants";
import { useTranslation } from "@/hooks";
import { Card } from "@/components/ui";

export function Testimonials() {
  const { t } = useTranslation();

  return (
    <section className="px-4 py-16 sm:px-6 sm:py-20">
      <div className="mx-auto max-w-6xl">
        <h2 className="mb-12 text-center text-3xl font-bold tracking-tight text-foreground">What our users say</h2>
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-3">
          {TESTIMONIALS.map((testimonial) => (
            <Card key={testimonial.id} className="flex flex-col p-6">
              <Quote className="mb-3 h-5 w-5 text-primary" aria-hidden="true" />
              <p className="mb-4 flex-1 text-sm leading-relaxed text-muted-foreground">
                {t(testimonial.quoteKey, testimonial.quote)}
              </p>
              <p className="border-t border-border pt-4 text-sm font-semibold text-foreground">
                {t(testimonial.nameKey, testimonial.name)}
              </p>
            </Card>
          ))}
        </div>
      </div>
    </section>
  );
}
