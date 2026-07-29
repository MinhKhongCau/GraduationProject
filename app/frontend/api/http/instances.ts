import { createHttpClient } from "./client";
import { API_BASE_URL } from "./config";

export const authClient = createHttpClient({
  baseURL: API_BASE_URL,
  transformCase: false, // Spring/Jackson already returns camelCase.
});

export const profileClient = createHttpClient({
  baseURL: API_BASE_URL,
  transformCase: true, // Go/Gin returns snake_case.
});

export const paymentClient = createHttpClient({
  baseURL: API_BASE_URL,
  transformCase: true,
});

export const bookingClient = createHttpClient({
  baseURL: API_BASE_URL,
  transformCase: true,
});

export const assessmentClient = createHttpClient({
  baseURL: API_BASE_URL,
  transformCase: true, // FastAPI/Pydantic returns snake_case.
  unwrapEnvelope: "result", // assessment-service wraps responses in {statusCode, result, message, ...}.
});

export const forumClient = createHttpClient({
  baseURL: API_BASE_URL,
  transformCase: false, // forum-service DTOs are already camelCase (see internal/api/dto/*.go).
  unwrapEnvelope: "success-data", // forum-service always wraps responses in {success, message, data, error}.
});
