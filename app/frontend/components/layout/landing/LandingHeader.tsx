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
      <div className="mx-auto flex max-w-6xl items-center justify-between px-4 py-3 sm:px-6">
        <BrandLogo className="text-xl" />

        <div className="flex items-center gap-3">
          <LanguageSwitcher className="hidden sm:inline-flex" />
          <Link href={ROUTES.AUTH.LOGIN} className={buttonClasses("ghost", "md")}>
            {t("landing.header.login", "Log in")}
          </Link>
          <Link href={ROUTES.AUTH.REGISTER} className={buttonClasses("primary", "md")}>
            {t("landing.header.register", "Sign up")}
          </Link>
        </div>
      </div>
    </header>
  );
}
