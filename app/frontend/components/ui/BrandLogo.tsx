import Link from "next/link";
import { Brain } from "lucide-react";
import clsx from "clsx";
import { ROUTES } from "@/constants";

export interface BrandLogoProps {
  label?: string;
  href?: string;
  className?: string;
}

export function BrandMark({ className }: { className?: string }) {
  return (
    <span className={clsx("inline-flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-white", className)}>
      <Brain className="h-5 w-5" aria-hidden="true" />
    </span>
  );
}

export function BrandLogo({ label = "MindCare", href = ROUTES.HOME, className }: BrandLogoProps) {
  return (
    <Link href={href} className={clsx("flex items-center gap-2 text-lg font-bold tracking-tight text-foreground", className)}>
      <BrandMark />
      {label}
    </Link>
  );
}
