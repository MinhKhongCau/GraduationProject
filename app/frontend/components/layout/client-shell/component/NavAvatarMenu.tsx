"use client";

import Link from "next/link";
import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import { LogOut, Settings } from "lucide-react";
import { useAuth, useTranslation } from "@/hooks";
import type { NavItem } from "@/constants/nav";

export interface NavAvatarMenuProps {
  settingsItem: NavItem;
}

function initials(fullName: string): string {
  return fullName
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join("");
}

export function NavAvatarMenu({ settingsItem }: NavAvatarMenuProps) {
  const { user, logout } = useAuth();
  const { t } = useTranslation();

  if (!user) return null;

  return (
    <DropdownMenu.Root>
      <DropdownMenu.Trigger asChild>
        <button
          type="button"
          className="flex h-9 w-9 items-center justify-center rounded-full border border-border bg-surface text-sm font-bold text-muted-foreground outline-none focus-visible:ring-2 focus-visible:ring-primary-soft"
          aria-label="Account menu"
        >
          {initials(user.fullName) || "?"}
        </button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Portal>
        <DropdownMenu.Content
          align="end"
          sideOffset={8}
          className="z-50 min-w-[200px] rounded-xl border border-border bg-background p-1.5 shadow-elevated"
        >
          <div className="px-3 py-2">
            <p className="truncate text-sm font-bold text-foreground">{user.fullName}</p>
            <p className="truncate text-xs text-muted-foreground">{user.email}</p>
          </div>
          <DropdownMenu.Separator className="my-1 h-px bg-border" />
          <DropdownMenu.Item asChild>
            <Link
              href={settingsItem.href}
              className="flex cursor-pointer items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium text-foreground outline-none hover:bg-surface"
            >
              <Settings className="h-4 w-4" />
              {t(settingsItem.labelKey, settingsItem.label)}
            </Link>
          </DropdownMenu.Item>
          <DropdownMenu.Item
            onSelect={logout}
            className="flex cursor-pointer items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium text-danger outline-none hover:bg-danger-soft"
          >
            <LogOut className="h-4 w-4" />
            {t("nav.logout", "Log out")}
          </DropdownMenu.Item>
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  );
}
