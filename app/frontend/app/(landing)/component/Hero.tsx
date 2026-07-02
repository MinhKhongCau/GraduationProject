"use client";

import Link from "next/link";
import { ArrowRight, Search } from "lucide-react";
import { Button } from "@/components/ui";
import { ROUTES, HERO_CONTENT } from "@/constants";
import { useTranslation } from "@/hooks";

export function Hero() {
  const { t } = useTranslation();

  return (
    <section className="bg-gradient-to-b from-primary-soft to-background px-6 py-20">
      <div className="mx-auto flex max-w-4xl flex-col items-center text-center">
        <span className="mb-4 rounded-full bg-primary-soft px-4 py-1.5 text-xs font-bold uppercase tracking-wider text-primary-soft-text">
          {t(HERO_CONTENT.eyebrowKey, HERO_CONTENT.eyebrow)}
        </span>
        <h1 className="mb-6 text-4xl font-extrabold leading-tight text-foreground sm:text-5xl">
          {t(HERO_CONTENT.titleKey, HERO_CONTENT.title)}
        </h1>
        <p className="mb-10 max-w-2xl text-lg text-muted-foreground">
          {t(HERO_CONTENT.subtitleKey, HERO_CONTENT.subtitle)}
        </p>
        <div className="flex flex-col gap-3 sm:flex-row">
          <Link href={ROUTES.AUTH.REGISTER}>
            <Button size="lg">
              {t(HERO_CONTENT.primaryCtaKey, HERO_CONTENT.primaryCta)}
              <ArrowRight className="h-4 w-4" />
            </Button>
          </Link>
          <Link href={ROUTES.AUTH.REGISTER}>
            <Button size="lg" variant="outline">
              <Search className="h-4 w-4" />
              {t(HERO_CONTENT.secondaryCtaKey, HERO_CONTENT.secondaryCta)}
            </Button>
          </Link>
        </div>
      </div>
    </section>
  );
}
