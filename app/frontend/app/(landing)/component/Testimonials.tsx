"use client";

import { Quote } from "lucide-react";
import { TESTIMONIALS } from "@/constants";
import { useTranslation } from "@/hooks";
import { Card } from "@/components/ui";

export function Testimonials() {
  const { t } = useTranslation();

  return (
    <section className="px-6 py-20">
      <div className="mx-auto max-w-6xl">
        <h2 className="mb-12 text-center text-3xl font-bold text-foreground">What our users say</h2>
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-3">
          {TESTIMONIALS.map((testimonial) => (
            <Card key={testimonial.id} className="p-6">
              <Quote className="mb-3 h-5 w-5 text-primary" />
              <p className="mb-4 text-sm text-muted-foreground">
                {t(testimonial.quoteKey, testimonial.quote)}
              </p>
              <p className="text-sm font-bold text-foreground">
                {t(testimonial.nameKey, testimonial.name)}
              </p>
            </Card>
          ))}
        </div>
      </div>
    </section>
  );
}
