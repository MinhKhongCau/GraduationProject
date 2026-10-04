import type { LucideIcon } from "lucide-react";
import { Card } from "@/components/ui";

export interface StatCardProps {
  label: string;
  value: string | number;
  icon: LucideIcon;
}

export function StatCard({ label, value, icon: Icon }: StatCardProps) {
  return (
    <Card className="p-5">
      <div className="mb-3 flex h-9 w-9 items-center justify-center rounded-lg bg-primary-soft text-primary-soft-text">
        <Icon className="h-4 w-4" aria-hidden="true" />
      </div>
      <p className="mb-1 text-xs font-semibold uppercase tracking-wide text-muted-foreground">{label}</p>
      <h3 className="text-2xl font-bold tracking-tight text-foreground">{value}</h3>
    </Card>
  );
}
