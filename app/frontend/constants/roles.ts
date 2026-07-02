import type { UserRole } from "@/types";

/** Source of truth: auth-service Account.role. Note API-document.md/root README
 * say "CLIENT" instead of "PATIENT" — code is ground truth, see DESIGN.md. */
export const ROLES: Record<UserRole, UserRole> = {
  PATIENT: "PATIENT",
  EXPERT: "EXPERT",
  ADMIN: "ADMIN",
};
