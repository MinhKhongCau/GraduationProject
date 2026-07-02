import { describe, it, expect } from "vitest";
import { toCamelCase, toSnakeCase } from "@/api/http/caseTransform";

describe("toCamelCase", () => {
  it("converts snake_case keys to camelCase", () => {
    const input = { full_name: "An Binh", account_id: "abc-123", is_active: true };
    expect(toCamelCase(input)).toEqual({ fullName: "An Binh", accountId: "abc-123", isActive: true });
  });

  it("converts nested objects and arrays deeply", () => {
    const input = {
      expert_id: "e1",
      specializations: [
        { spec_id: "s1", is_active: true },
        { spec_id: "s2", is_active: false },
      ],
    };
    expect(toCamelCase(input)).toEqual({
      expertId: "e1",
      specializations: [
        { specId: "s1", isActive: true },
        { specId: "s2", isActive: false },
      ],
    });
  });

  it("matches a real profile-service Expert DTO shape", () => {
    const input = {
      expert_id: "e1",
      account_id: "a1",
      full_name: "Dr. Tran Minh Tam",
      verification_status: "VERIFIED",
      introduction_video_url: null,
    };
    expect(toCamelCase(input)).toEqual({
      expertId: "e1",
      accountId: "a1",
      fullName: "Dr. Tran Minh Tam",
      verificationStatus: "VERIFIED",
      introductionVideoUrl: null,
    });
  });

  it("leaves primitives and null untouched", () => {
    expect(toCamelCase("hello")).toBe("hello");
    expect(toCamelCase(42)).toBe(42);
    expect(toCamelCase(null)).toBe(null);
  });
});

describe("toSnakeCase", () => {
  it("converts camelCase keys to snake_case", () => {
    const input = { fullName: "An Binh", accountId: "abc-123", isActive: true };
    expect(toSnakeCase(input)).toEqual({ full_name: "An Binh", account_id: "abc-123", is_active: true });
  });

  it("round-trips through camel -> snake -> camel", () => {
    const original = { patientId: "p1", medicalHistories: [{ conditionName: "Insomnia" }] };
    expect(toCamelCase(toSnakeCase(original))).toEqual(original);
  });
});
