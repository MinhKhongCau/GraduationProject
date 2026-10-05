"use client";

import Link from "next/link";
import { BrandLogo } from "@/components/ui";
import { usePathname } from "next/navigation";
import clsx from "clsx";
import { useTranslation } from "@/hooks";
import { ADMIN_HEADER_NAV_ITEMS, isNavItemActive } from "@/constants";

export function AdminSidebar() {
  const pathname = usePathname();
  const { t } = useTranslation();

  return (
    <aside className="sticky top-0 hidden h-screen w-64 shrink-0 flex-col border-r border-border bg-background lg:flex">
      <div className="flex h-16 items-center border-b border-border px-6">
        <BrandLogo label="MindCare Admin" />
      </div>
      <nav className="flex-1 space-y-1 overflow-y-auto p-3">
        {ADMIN_HEADER_NAV_ITEMS.map((item) => {
          const isActive = isNavItemActive(item, pathname);
          const Icon = item.icon;
          return (
            <Link
              key={item.id}
              href={item.href}
              aria-current={isActive ? "page" : undefined}
              className={clsx(
                "flex h-10 items-center gap-3 rounded-lg px-3 text-sm font-medium transition-colors",
                isActive ? "bg-primary-soft text-primary-soft-text font-semibold" : "text-muted-foreground hover:bg-surface hover:text-foreground"
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
