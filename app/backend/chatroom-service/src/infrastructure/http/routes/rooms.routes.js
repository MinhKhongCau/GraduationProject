import { Router } from "express";
import { listRooms } from "../controllers/rooms.controller.js";

const router = Router();

/**
 * @openapi
 * /rooms:
 *   get:
 *     summary: List active chat rooms
 *     tags: [Rooms]
 *     responses:
 *       200:
 *         description: Active rooms with live user counts
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 rooms:
 *                   type: array
 *                   items:
 *                     type: object
 *                     properties:
 *                       id: { type: string }
 *                       name: { type: string }
 *                       icon: { type: string }
 *                       users: { type: integer }
 */
router.get("/", listRooms);

export default router;