import dotenv from "dotenv";
dotenv.config();

// Root .env sets CORS_ORIGINS as a comma-separated list shared by every
// MindCare service; CLIENT_ORIGIN is kept as a fallback for standalone runs.
function parseOrigins(value) {
  if (!value || value === "*") return "*";
  return value.split(",").map((origin) => origin.trim()).filter(Boolean);
}

export const ENV = {
  PORT: Number(process.env.PORT || 5000),
  NODE_ENV: process.env.NODE_ENV || "development",
  CORS_ORIGINS: parseOrigins(process.env.CORS_ORIGINS || process.env.CLIENT_ORIGIN),
  REDIS_URL: process.env.REDIS_URL,
  // Raw base64 X.509/SPKI DER — same value auth-service signs JWTs with
  // (see auth-service's RsaKeyConfig). Verified locally at the Socket.IO
  // handshake instead of trusting client-supplied identity.
  JWT_PUBLIC_KEY: process.env.JWT_PUBLIC_KEY || "",
};

if (!ENV.REDIS_URL) throw new Error("❌ Missing REDIS_URL in .env");
if (!ENV.JWT_PUBLIC_KEY) {
  console.warn("⚠️  JWT_PUBLIC_KEY not set — every socket connection will be rejected.");
}