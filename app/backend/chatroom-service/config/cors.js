import { ENV } from "./env.js";

export const corsOptions = {
  origin: ENV.CORS_ORIGINS,
  credentials: true,
  methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
};
