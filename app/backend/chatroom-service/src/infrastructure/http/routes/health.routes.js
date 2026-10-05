import { Router } from "express";
import { healthCheck } from "../controllers/health.controller.js";

const router = Router();

/**
 * @openapi
 * /health:
 *   get:
 *     summary: Service health check
 *     tags: [Health]
 *     responses:
 *       200:
 *         description: Service is up
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 ok: { type: boolean }
 *                 message: { type: string }
 */
router.get("/", healthCheck);

export default router;