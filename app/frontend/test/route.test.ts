import { describe, it, expect } from "vitest";
import { ROUTES } from "@/constants/route";
import { resolveRequiredRole, dashboardForRole } from "@/router/routes.config";

function collectStrings(node: unknown, acc: string[] = []): string[] {
  if (typeof node === "string") acc.push(node);
  if (typeof node === "function") acc.push((node as (...args: string[]) => string)("x"));
  if (node && typeof node === "object") {
    Object.values(node).forEach((value) => collectStrings(value, acc));
  }
  return acc;
}

describe("ROUTES", () => {
  it("has no duplicate path constants", () => {
    const paths = collectStrings(ROUTES);
    expect(new Set(paths).size).toBe(paths.length);
  });

  it("every non-public route path resolves to a required role", () => {
    expect(resolveRequiredRole(ROUTES.PATIENT.DASHBOARD)).toBe("PATIENT");
    expect(resolveRequiredRole(ROUTES.EXPERT.DASHBOARD)).toBe("EXPERT");
    expect(resolveRequiredRole(ROUTES.ADMIN.DASHBOARD)).toBe("ADMIN");
    expect(resolveRequiredRole(ROUTES.HOME)).toBeNull();
    expect(resolveRequiredRole(ROUTES.AUTH.LOGIN)).toBeNull();
  });
});

describe("dashboardForRole", () => {
  it("maps each role to its own dashboard", () => {
    expect(dashboardForRole("PATIENT")).toBe(ROUTES.PATIENT.DASHBOARD);
    expect(dashboardForRole("EXPERT")).toBe(ROUTES.EXPERT.DASHBOARD);
    expect(dashboardForRole("ADMIN")).toBe(ROUTES.ADMIN.DASHBOARD);
  });
});
