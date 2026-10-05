import { Server } from "socket.io";
import { createAdapter } from "@socket.io/redis-adapter";
import { redisPub, redisSub } from "../../../config/redis.js";
import { ENV } from "../../../config/env.js";
import { socketAuthMiddleware } from "./auth.middleware.js";
import { setOnline, setOffline } from "../persistence/stores/presence.store.js";

import { roomSocketController } from "./controllers/room.socket.controller.js";
import { chatSocketController } from "./controllers/chat.socket.controller.js";
import { dmSocketController } from "./controllers/dm.socket.controller.js";
import { presenceSocketController } from "./controllers/presence.socket.controller.js";
import { webrtcSocketController } from "./controllers/webrtc.socket.controller.js";
import { registerCallSocketHandlers } from "./controllers/call.socket.controller.js";

export function initSocket(server) {
  const io = new Server(server, {
    cors: {
      origin: ENV.CORS_ORIGINS,
      methods: ["GET", "POST"],
      credentials: true,
    },
    // Default 1e6 is too small for base64-encoded voice-message payloads.
    maxHttpBufferSize: 8 * 1024 * 1024,
  });

  io.adapter(createAdapter(redisPub, redisSub));

  io.use(socketAuthMiddleware);

  io.on("connection", (socket) => {
    // DM/presence identity comes only from the verified JWT (auth.middleware.js),
    // never from client input — mark this accountId online for the whole
    // connection lifetime so presence:query works without requiring room:join
    // (rooms are unused by the DM-only frontend).
    const accountId = socket.data.user?.id;
    if (accountId) setOnline(accountId, socket.id);

    socket.on("disconnect", () => {
      if (accountId) setOffline(accountId, socket.id);
    });

    roomSocketController(io, socket);
    chatSocketController(io, socket);
    dmSocketController(io, socket);
    presenceSocketController(io, socket);
    webrtcSocketController(io, socket);
    registerCallSocketHandlers(io, socket);
  });

  return io;
}
