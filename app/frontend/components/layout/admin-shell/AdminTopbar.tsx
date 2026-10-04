"use client";

import { useRouter, usePathname } from "next/navigation";
import { LogOut } from "lucide-react";
import { useAuth, useTranslation } from "@/hooks";
import { ADMIN_NAV_ITEMS } from "@/constants";
import { LanguageSwitcher } from "../LanguageSwitcher";
import { Button, Select } from "@/components/ui";

export function AdminTopbar() {
  const { logout } = useAuth();
  const { t } = useTranslation();
  const router = useRouter();
  const pathname = usePathname();

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-border bg-background/95 px-4 backdrop-blur sm:px-6">
      {/* Admin is desktop-first (the sidebar is lg+ only); this select is a
          minimal, low-effort navigation fallback for smaller screens. */}
      <Select
        aria-label="Admin navigation"
        value={pathname}
        onChange={(event) => router.push(event.target.value)}
        className="h-9 w-auto font-semibold lg:hidden"
      >
        {ADMIN_NAV_ITEMS.map((item) => (
          <option key={item.id} value={item.href}>
            {t(item.labelKey, item.label)}
          </option>
        ))}
      </Select>
      <span className="hidden text-sm font-semibold text-foreground lg:inline">
        {ADMIN_NAV_ITEMS.find((item) => item.href === pathname)?.label ?? "Admin"}
      </span>

      <div className="flex items-center gap-3">
        <LanguageSwitcher />
        <Button size="sm" variant="ghost" onClick={logout}>
          <LogOut className="h-4 w-4" />
          {t("nav.logout", "Log out")}
        </Button>
      </div>
    </header>
  );
}
