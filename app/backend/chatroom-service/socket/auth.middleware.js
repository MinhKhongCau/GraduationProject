import jwt from "jsonwebtoken";
import crypto from "crypto";
import { ENV } from "../config/env.js";

function loadPublicKey() {
  const raw = (ENV.JWT_PUBLIC_KEY || "").replace(/\s+/g, "");
  if (!raw) return null;
  try {
    // Primary: raw base64 X.509/SPKI DER, as issued by auth-service's
    // RsaKeyConfig.getPublicKeyBase64() — no PEM armor.
    return crypto.createPublicKey({
      key: Buffer.from(raw, "base64"),
      format: "der",
      type: "spki",
    });
  } catch {
    // Fallback: in case JWT_PUBLIC_KEY is ever set as PEM directly.
    const pem = raw.includes("BEGIN PUBLIC KEY")
      ? raw
      : `-----BEGIN PUBLIC KEY-----\n${raw}\n-----END PUBLIC KEY-----`;
    return crypto.createPublicKey(pem);
  }
}

const publicKey = loadPublicKey();

// Socket.IO io.use() middleware — verifies the same RS256 JWT auth-service
// issues instead of trusting a client-supplied display name. Populates
// socket.data.user = { id: accountId, email, role } for every downstream
// controller (dm.*, presence.*) to address DMs/presence by real accountId.
export function socketAuthMiddleware(socket, next) {
  if (!publicKey && process.env.NODE_ENV !== "test") {
    return next(new Error("unauthorized: server misconfigured (JWT_PUBLIC_KEY not set)"));
  }

  const token = socket.handshake.auth?.token;
  if (!token) return next(new Error("unauthorized: missing token"));

  try {
    const claims = jwt.verify(token, publicKey || "dummy-key", { algorithms: ["RS256"] });
    const accountId = claims.accountId;
    const role = claims.role;
    if (!accountId || !role) {
      return next(new Error("unauthorized: token missing accountId/role"));
    }
    socket.data.user = { id: accountId, email: claims.sub, role };
    return next();
  } catch (err) {
    if (err.name === "TokenExpiredError") {
      return next(new Error("unauthorized: token expired"));
    }
    return next(new Error("unauthorized: invalid token"));
  }
}
