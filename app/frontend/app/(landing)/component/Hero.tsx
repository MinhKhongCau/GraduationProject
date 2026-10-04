"use client";

import Link from "next/link";
import { ArrowRight, Search } from "lucide-react";
import { buttonClasses } from "@/components/ui";
import { ROUTES, HERO_CONTENT } from "@/constants";
import { useTranslation } from "@/hooks";

export function Hero() {
  const { t } = useTranslation();

  return (
    <section className="bg-gradient-to-b from-primary-soft to-background px-4 py-16 sm:px-6 sm:py-24">
      <div className="mx-auto flex max-w-4xl flex-col items-center text-center">
        <span className="mb-4 rounded-full border border-primary/20 bg-background px-4 py-1.5 text-xs font-bold uppercase tracking-wider text-primary-soft-text">
          {t(HERO_CONTENT.eyebrowKey, HERO_CONTENT.eyebrow)}
        </span>
        <h1 className="mb-6 text-4xl font-extrabold leading-tight tracking-tight text-foreground sm:text-5xl">
          {t(HERO_CONTENT.titleKey, HERO_CONTENT.title)}
        </h1>
        <p className="mb-10 max-w-2xl text-lg text-muted-foreground">
          {t(HERO_CONTENT.subtitleKey, HERO_CONTENT.subtitle)}
        </p>
        <div className="flex w-full flex-col gap-3 sm:w-auto sm:flex-row">
          <Link href={ROUTES.AUTH.REGISTER} className={buttonClasses("primary", "lg")}>
            {t(HERO_CONTENT.primaryCtaKey, HERO_CONTENT.primaryCta)}
            <ArrowRight className="h-4 w-4" aria-hidden="true" />
          </Link>
          <Link href={ROUTES.AUTH.REGISTER} className={buttonClasses("outline", "lg")}>
            <Search className="h-4 w-4" aria-hidden="true" />
            {t(HERO_CONTENT.secondaryCtaKey, HERO_CONTENT.secondaryCta)}
          </Link>
        </div>
      </div>
    </section>
  );
}
