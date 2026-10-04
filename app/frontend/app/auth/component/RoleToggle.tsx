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
      <p id="register-role-label" className="mb-1.5 text-sm font-medium text-foreground">I am a:</p>
      <div role="group" aria-labelledby="register-role-label" className="flex gap-1 rounded-lg border border-border bg-surface p-1">
        {(["PATIENT", "EXPERT"] as const).map((role) => (
          <button
            key={role}
            type="button"
            onClick={() => onChange(role)}
            aria-pressed={value === role}
            className={clsx(
              "h-9 flex-1 rounded-md text-sm font-semibold capitalize transition-colors",
              value === role
                ? "bg-background text-primary shadow-card"
                : "text-muted-foreground hover:text-foreground"
            )}
          >
            {role === "PATIENT" ? "Patient" : "Expert"}
          </button>
        ))}
      </div>
    </div>
  );
}
