import clsx from "clsx";
import type { RegisterRequest } from "@/types";

export type RegisterRole = RegisterRequest["role"];

export interface RoleToggleProps {
  value: RegisterRole;
  onChange: (role: RegisterRole) => void;
}

export function RoleToggle({ value, onChange }: RoleToggleProps) {
  return (
    <div>
      <p className="mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">I am a:</p>
      <div className="flex gap-2">
        {(["PATIENT", "EXPERT"] as const).map((role) => (
          <button
            key={role}
            type="button"
            onClick={() => onChange(role)}
            className={clsx(
              "flex-1 rounded-lg py-2 text-sm font-semibold capitalize transition-all",
              value === role
                ? "bg-primary text-white shadow-card"
                : "border border-border bg-surface text-muted-foreground hover:bg-border/40"
            )}
          >
            {role === "PATIENT" ? "Patient" : "Expert"}
          </button>
        ))}
      </div>
    </div>
  );
}
