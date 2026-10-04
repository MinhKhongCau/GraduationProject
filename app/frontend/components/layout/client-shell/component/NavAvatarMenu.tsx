"use client";

import Link from "next/link";
import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import { LogOut, Settings } from "lucide-react";
import { useAuth, useTranslation } from "@/hooks";
import { useQuery } from "@tanstack/react-query";
import { patientApi, expertApi } from "@/api";
import type { NavItem } from "@/constants/nav";
import type { PatientProfile, ExpertProfile } from "@/types";

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

  const isPatient = user?.role === "PATIENT";
  const isExpert = user?.role === "EXPERT";

  const { data: profile } = useQuery<PatientProfile | ExpertProfile>({
    queryKey: isPatient ? ["patient", "profile", "me"] : ["expert", "profile", "me"],
    queryFn: () => (isPatient ? patientApi.getMyProfile() : expertApi.getMyProfile()),
    enabled: !!user && (isPatient || isExpert),
    staleTime: 5 * 60 * 1000,
  });

  if (!user) return null;

  const displayAvatar = profile?.avatarUrl;
  const displayFullName = profile?.fullName || user.fullName;

  return (
    <DropdownMenu.Root>
      <DropdownMenu.Trigger asChild>
        <button
          type="button"
          className="relative flex h-10 w-10 items-center justify-center overflow-hidden rounded-full border border-border-strong bg-surface text-sm font-bold text-muted-foreground transition-shadow hover:ring-2 hover:ring-primary/20"
          aria-label="Account menu"
        >
          {displayAvatar ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={displayAvatar}
              alt={displayFullName}
              className="h-full w-full object-cover"
            />
          ) : (
            initials(displayFullName) || "?"
          )}
        </button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Portal>
        <DropdownMenu.Content
          align="end"
          sideOffset={8}
          className="z-50 min-w-[200px] rounded-xl border border-border bg-background p-1.5 shadow-elevated"
        >
          <div className="px-3 py-2 flex items-center gap-3">
            {displayAvatar && (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                src={displayAvatar}
                alt={displayFullName}
                className="h-8 w-8 rounded-full object-cover border border-border"
              />
            )}
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-bold text-foreground">{displayFullName}</p>
              <p className="truncate text-xs text-muted-foreground">{user.email}</p>
            </div>
          </div>
          <DropdownMenu.Separator className="my-1 h-px bg-border" />
          <DropdownMenu.Item asChild>
            <Link
              href={settingsItem.href}
              className="flex h-10 cursor-pointer items-center gap-2 rounded-lg px-3 text-sm font-medium text-foreground outline-none hover:bg-surface data-[highlighted]:bg-surface"
            >
              <Settings className="h-4 w-4" />
              {t(settingsItem.labelKey, settingsItem.label)}
            </Link>
          </DropdownMenu.Item>
          <DropdownMenu.Item
            onSelect={logout}
            className="flex h-10 cursor-pointer items-center gap-2 rounded-lg px-3 text-sm font-medium text-danger outline-none hover:bg-danger-soft data-[highlighted]:bg-danger-soft"
          >
            <LogOut className="h-4 w-4" />
            {t("nav.logout", "Log out")}
          </DropdownMenu.Item>
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  );
}
