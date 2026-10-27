"use client";

import Link from "next/link";
import { useTranslation } from "@/hooks";
import { ROUTES } from "@/constants";
import { BrandLogo, buttonClasses } from "@/components/ui";
import { LanguageSwitcher } from "../LanguageSwitcher";

export function LandingHeader() {
  const { t } = useTranslation();

  return (
    <header className="sticky top-0 z-30 border-b border-border bg-background/90 backdrop-blur">
      <div className="mx-auto flex max-w-6xl items-center justify-between gap-2 px-4 py-3 pt-[max(0.75rem,env(safe-area-inset-top))] sm:px-6">
        <BrandLogo className="shrink-0 sm:text-xl" />

        <div className="flex min-w-0 items-center gap-1.5 sm:gap-3">
          {/* Wrapper hides it on phones; a "hidden" class on the switcher itself loses to its own inline-flex. */}
          <div className="hidden sm:block">
            <LanguageSwitcher />
          </div>
          <Link href={ROUTES.AUTH.LOGIN} className={buttonClasses("ghost", "sm", "sm:h-10 sm:px-4 sm:text-sm")}>
            {t("landing.header.login", "Log in")}
          </Link>
          <Link href={ROUTES.AUTH.REGISTER} className={buttonClasses("primary", "sm", "sm:h-10 sm:px-4 sm:text-sm")}>
            {t("landing.header.register", "Sign up")}
          </Link>
        </div>
      </div>
    </header>
  );
}
