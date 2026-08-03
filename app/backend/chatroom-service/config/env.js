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
  REDIS_URL: process.env.REDIS_URL || (process.env.NODE_ENV === "test" ? "redis://localhost:6379" : ""),
  // Raw base64 X.509/SPKI DER — same value auth-service signs JWTs with
  // (see auth-service's RsaKeyConfig). Verified locally at the Socket.IO
  // handshake instead of trusting client-supplied identity.
  JWT_PUBLIC_KEY: process.env.JWT_PUBLIC_KEY || "",
  DB_HOST: process.env.DB_HOST || "localhost",
  DB_PORT: Number(process.env.DB_PORT || 5432),
  DB_USER: process.env.DB_USER || "admin",
  DB_PASSWORD: process.env.DB_PASSWORD || "admin",
  DB_NAME: process.env.DB_NAME || "chatroom_db",
  DB_SSLMODE: process.env.DB_SSLMODE || "disable",
  CHATBOT_URL: process.env.CHATBOT_URL || "http://localhost:8086",
};

if (!ENV.REDIS_URL && process.env.NODE_ENV !== "test") throw new Error("❌ Missing REDIS_URL in .env");
if (!ENV.JWT_PUBLIC_KEY && process.env.NODE_ENV !== "test") {
  console.warn("⚠️  JWT_PUBLIC_KEY not set — every socket connection will be rejected.");
}