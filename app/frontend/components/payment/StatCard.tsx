import type { ReactNode } from "react";
import clsx from "clsx";
import { Card } from "@/components/ui";

export interface StatCardProps {
  label: ReactNode;
  value: ReactNode;
  hint?: ReactNode;
  tone?: "default" | "success" | "danger" | "primary";
}

const TONES = {
  default: "text-foreground",
  success: "text-success",
  danger: "text-danger",
  primary: "text-primary",
};

/** One KPI tile in a transaction summary row. */
export function StatCard({ label, value, hint, tone = "default" }: StatCardProps) {
  return (
    <Card className="p-4">
      <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{label}</p>
      <p className={clsx("mt-2 text-xl font-bold tabular-nums", TONES[tone])}>{value}</p>
      {hint && <p className="mt-1 text-xs text-muted-foreground">{hint}</p>}
    </Card>
  );
}
