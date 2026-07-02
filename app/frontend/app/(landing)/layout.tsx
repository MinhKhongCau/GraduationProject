import type { ReactNode } from "react";
import { LandingHeader } from "@/components/layout/landing/LandingHeader";
import { LandingFooter } from "@/components/layout/landing/LandingFooter";

export default function LandingLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex flex-1 flex-col">
      <LandingHeader />
      <main className="flex-1">{children}</main>
      <LandingFooter />
    </div>
  );
}
