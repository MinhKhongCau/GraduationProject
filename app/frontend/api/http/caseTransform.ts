/**
 * auth-service (Spring) returns camelCase JSON; profile/payment/booking
 * (Go) and assessment (FastAPI) return snake_case. These converters let
 * every types/*.ts interface and api/*.ts call site stay camelCase
 * (per document/code-convention.md) with no per-DTO manual mapping —
 * see DESIGN.md "Case transform".
 */

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return (
    typeof value === "object" &&
    value !== null &&
    !Array.isArray(value) &&
    !(value instanceof Date) &&
    !(typeof File !== "undefined" && value instanceof File) &&
    !(typeof Blob !== "undefined" && value instanceof Blob)
  );
}

function snakeToCamelKey(key: string): string {
  return key.replace(/_([a-z0-9])/g, (_match, char: string) => char.toUpperCase());
}

function camelToSnakeKey(key: string): string {
  return key.replace(/[A-Z]/g, (char) => `_${char.toLowerCase()}`);
}

function deepTransform(value: unknown, transformKey: (key: string) => string): unknown {
  if (Array.isArray(value)) {
    return value.map((item) => deepTransform(item, transformKey));
  }
  if (isPlainObject(value)) {
    const result: Record<string, unknown> = {};
    for (const [key, val] of Object.entries(value)) {
      result[transformKey(key)] = deepTransform(val, transformKey);
    }
    return result;
  }
  return value;
}

export function toCamelCase<T = unknown>(value: unknown): T {
  return deepTransform(value, snakeToCamelKey) as T;
}

export function toSnakeCase<T = unknown>(value: unknown): T {
  return deepTransform(value, camelToSnakeKey) as T;
}
