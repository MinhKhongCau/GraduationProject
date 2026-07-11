import { createHttpClient } from "./client";

const baseURL = process.env.REACT_APP_API_URL ?? "https://api.qmcloud.io.vn/api/v1";

export const authClient = createHttpClient({
  baseURL,
  transformCase: false, // Spring/Jackson already returns camelCase.
});

export const profileClient = createHttpClient({
  baseURL,
  transformCase: true, // Go/Gin returns snake_case.
});

export const paymentClient = createHttpClient({
  baseURL,
  transformCase: true,
});

export const bookingClient = createHttpClient({
  baseURL,
  transformCase: true,
});

export const assessmentClient = createHttpClient({
  baseURL,
  transformCase: true, // FastAPI/Pydantic returns snake_case.
  unwrapEnvelope: true, // assessment-service wraps responses in {statusCode, result, message, ...}.
});
