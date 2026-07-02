"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import clsx from "clsx";
import { useTranslation } from "@/hooks";
import { ROUTES, ADMIN_NAV_ITEMS } from "@/constants";

export function AdminSidebar() {
  const pathname = usePathname();
  const { t } = useTranslation();

  return (
    <aside className="hidden w-64 shrink-0 flex-col border-r border-border bg-background lg:flex">
      <Link href={ROUTES.HOME} className="flex items-center gap-2 border-b border-border px-6 py-5 text-lg font-bold text-foreground">
        <span className="rounded-lg bg-primary p-1 text-white">🧠</span>
        MindCare Admin
      </Link>
      <nav className="flex-1 space-y-1 p-4">
        {ADMIN_NAV_ITEMS.map((item) => {
          const isActive = pathname === item.href;
          const Icon = item.icon;
          return (
            <Link
              key={item.id}
              href={item.href}
              className={clsx(
                "flex items-center gap-3 rounded-lg px-4 py-2.5 text-sm font-medium transition-colors",
                isActive ? "bg-primary text-white shadow-card" : "text-muted-foreground hover:bg-surface"
              )}
            >
              <Icon className="h-4 w-4" />
              {t(item.labelKey, item.label)}
            </Link>
          );
        })}
      </nav>
    </aside>
  );
}
