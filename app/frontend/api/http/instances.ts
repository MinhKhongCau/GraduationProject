import { createHttpClient } from "./client";

export const authClient = createHttpClient({
  baseURL: process.env.NEXT_PUBLIC_AUTH_API_URL ?? "http://localhost:8080/api/v1",
  transformCase: false, // Spring/Jackson already returns camelCase.
});

export const profileClient = createHttpClient({
  baseURL: process.env.NEXT_PUBLIC_PROFILE_API_URL ?? "http://localhost:8081/api/v1",
  transformCase: true, // Go/Gin returns snake_case.
});

export const paymentClient = createHttpClient({
  baseURL: process.env.NEXT_PUBLIC_PAYMENT_API_URL ?? "http://localhost:8082/api/v1",
  transformCase: true,
});

export const bookingClient = createHttpClient({
  baseURL: process.env.NEXT_PUBLIC_BOOKING_API_URL ?? "http://localhost:8083/api/v1",
  transformCase: true,
});

export const assessmentClient = createHttpClient({
  baseURL: process.env.NEXT_PUBLIC_ASSESSMENT_API_URL ?? "http://localhost:5000/api/v1",
  transformCase: true, // FastAPI/Pydantic returns snake_case.
});
