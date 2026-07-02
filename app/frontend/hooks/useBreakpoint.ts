"use client";

import { useMediaQuery } from "./useMediaQuery";

export type Breakpoint = "mobile" | "tablet" | "desktop";

/**
 * Mirrors Tailwind's own sm(640)/lg(1024) boundaries so JS and CSS never
 * disagree. The three nav variants in ClientShell switch primarily via
 * Tailwind responsive classes (no hydration flash) — this hook is for
 * genuine JS-only behavioral branches. See DESIGN.md.
 */
export function useBreakpoint(): Breakpoint {
  const isDesktop = useMediaQuery("(min-width: 1024px)");
  const isTablet = useMediaQuery("(min-width: 640px) and (max-width: 1023px)");

  if (isDesktop) return "desktop";
  if (isTablet) return "tablet";
  return "mobile";
}
