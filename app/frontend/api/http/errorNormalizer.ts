import { isAxiosError, type AxiosError } from "axios";
import type { ApiErrorResponse } from "@/types";

/**
 * Backend error shapes are inconsistent across services:
 * - auth-service (Spring) wraps most failures as { message } but returns a
 *   raw string body for register() failures.
 * - Go services (profile/payment/booking) use { error }.
 * This always produces the one shape the rest of the app consumes.
 */
export function normalizeError(error: unknown): ApiErrorResponse {
  if (isAxiosError(error)) {
    const axiosError = error as AxiosError;
    const data = axiosError.response?.data;
    const statusCode = axiosError.response?.status;

    if (typeof data === "string" && data.trim().length > 0) {
      return { message: data, statusCode };
    }
    if (data && typeof data === "object") {
      const record = data as Record<string, unknown>;
      if (typeof record.message === "string") {
        return { message: record.message, statusCode, details: record.details };
      }
      if (typeof record.error === "string") {
        return { message: record.error, statusCode, details: record.details };
      }
    }
    if (axiosError.code === "ECONNABORTED") {
      return { message: "Request timed out. Please try again.", statusCode };
    }
    if (!axiosError.response) {
      return { message: "Network error. Please check your connection.", statusCode };
    }
    return { message: axiosError.message || "Something went wrong.", statusCode };
  }

  if (error instanceof Error) {
    return { message: error.message };
  }

  return { message: "Something went wrong." };
}
