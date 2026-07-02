"use client";

import Link from "next/link";
import { useTranslation } from "@/hooks";
import { ROUTES } from "@/constants";
import { Button } from "@/components/ui";
import { LanguageSwitcher } from "../LanguageSwitcher";

export function LandingHeader() {
  const { t } = useTranslation();

  return (
    <header className="sticky top-0 z-30 border-b border-border bg-background/90 backdrop-blur">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
        <Link href={ROUTES.HOME} className="flex items-center gap-2 text-xl font-bold text-foreground">
          <span className="rounded-lg bg-primary p-1 text-white">🧠</span>
          MindCare
        </Link>

        <div className="flex items-center gap-3">
          <LanguageSwitcher className="hidden sm:inline-flex" />
          <Link href={ROUTES.AUTH.LOGIN}>
            <Button variant="ghost" size="sm">
              {t("landing.header.login", "Log in")}
            </Button>
          </Link>
          <Link href={ROUTES.AUTH.REGISTER}>
            <Button size="sm">{t("landing.header.register", "Sign up")}</Button>
          </Link>
        </div>
      </div>
    </header>
  );
}
