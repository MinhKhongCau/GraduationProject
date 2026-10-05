import type { HTMLAttributes } from "react";
import clsx from "clsx";

export type BadgeTone = "neutral" | "primary" | "success" | "warning" | "danger";

const TONE_CLASSES: Record<BadgeTone, string> = {
  neutral: "bg-surface text-muted-foreground border-border",
  primary: "bg-primary-soft text-primary-soft-text border-primary/20",
  success: "bg-success-soft text-success border-success/20",
  warning: "bg-warning-soft text-warning border-warning/20",
  danger: "bg-danger-soft text-danger border-danger/20",
};

export interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  tone?: BadgeTone;
}

/** Small status pill. Always pair the tone with readable text — never color alone. */
export function Badge({ tone = "neutral", className, ...props }: BadgeProps) {
  return (
    <span
      className={clsx(
        "inline-flex items-center gap-1 rounded-full border px-2.5 py-0.5 text-xs font-semibold whitespace-nowrap",
        TONE_CLASSES[tone],
        className
      )}
      {...props}
    />
  );
}
