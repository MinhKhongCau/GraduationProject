import express from "express";
import cors from "cors";
import swaggerUi from "swagger-ui-express";
import routes from "./routes/index.js";
import { corsOptions } from "./config/cors.js";
import { errorMiddleware } from "./middlewares/error.middleware.js";
import { healthCheck } from "./controllers/health.controller.js";
import { swaggerSpec } from "./config/swagger.js";

export function createApp() {
  const app = express();

  app.use(cors(corsOptions));
  app.use(express.json());

  app.get("/", (req, res) => res.send("Chatroom service running ✅"));

  // Root-level health check — matches the /health convention every other
  // MindCare service exposes for the api-gateway's upstream healthchecks.
  app.get("/health", healthCheck);

  app.get("/api-docs.json", (req, res) => res.json(swaggerSpec));
  app.use("/api-docs", swaggerUi.serve, swaggerUi.setup(swaggerSpec));

  app.use("/api/v1/chatroom", routes);

  app.use(errorMiddleware);
  return app;
}
