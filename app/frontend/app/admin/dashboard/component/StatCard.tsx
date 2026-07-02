import type { LucideIcon } from "lucide-react";
import { Card } from "@/components/ui";

export interface StatCardProps {
  label: string;
  value: string | number;
  icon: LucideIcon;
}

export function StatCard({ label, value, icon: Icon }: StatCardProps) {
  return (
    <Card className="p-6">
      <div className="mb-3 flex h-9 w-9 items-center justify-center rounded-full bg-primary-soft text-primary-soft-text">
        <Icon className="h-4 w-4" />
      </div>
      <p className="mb-1 text-xs font-bold uppercase tracking-wider text-muted-foreground">{label}</p>
      <h3 className="text-2xl font-extrabold text-foreground">{value}</h3>
    </Card>
  );
}
