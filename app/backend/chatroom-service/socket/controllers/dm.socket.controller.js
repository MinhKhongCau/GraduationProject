import { getDMHistory, pushDM, makeDmId } from "../../stores/dm.store.js";
import { getUserSocket } from "../../stores/presence.store.js";
import { toggleReaction } from "../../services/reactions.service.js";

// NOTE: `toUser` on every incoming payload is the recipient's accountId (a
// JWT `accountId` claim), not a display name — kept the field name as-is to
// minimize event-shape churn, only the semantics changed. The sender's
// identity (`fromId`) always comes from socket.data.user, set by the
// verified-JWT auth.middleware.js, never from client input.

export function dmSocketController(io, socket) {
  // DM history
  socket.on("dm:history", async ({ toUser, page = 1, limit = 20 }) => {
    const fromId = socket.data.user?.id;

    if (!fromId || !toUser) return;

    const dmId = makeDmId(fromId, toUser);
    const start = (page - 1) * limit;
    const end = start + limit - 1;
    const history = await getDMHistory(dmId, start, end);

    socket.emit("dm:history", { dmId, history, page, limit });
  });

  //SEND DM message
  socket.on("dm:send", async ({ toUser, text }) => {
    const fromId = socket.data.user?.id;

    if (!fromId || !toUser) return;

    const clean = (text || "").trim();
    if (!clean) return;

    const dmId = makeDmId(fromId, toUser);

    const msg = {
      id: `${Date.now()}-${Math.random()}`,
      dmId,
      type: "chat",
      text: clean,
      fromId,
      toId: toUser,
      createdAt: Date.now(),
    };

    await pushDM(dmId, msg); // store DM message

    // Send to sender
    socket.emit("dm:message", msg);

    // Send to receiver if online
    const toSocketId = await getUserSocket(toUser);
    if (toSocketId) {
      io.to(toSocketId).emit("dm:message", msg);
    }
  });

  //SEND VOICE DM
  socket.on("dm:send:voice", async ({ toUser, audio, duration, mimeType }) => {
    const fromId = socket.data.user?.id;

    if (!fromId || !toUser) {
      console.error("dm:send:voice - missing fromId or toUser");
      return;
    }

    if (!audio) {
      console.error("dm:send:voice - no audio data");
      return;
    }

    const dmId = makeDmId(fromId, toUser);

    const msg = {
      id: `${Date.now()}-${Math.random()}`,
      dmId,
      type: "voice",
      audio,
      duration: duration || 0,
      mimeType: mimeType || "audio/webm",
      fromId,
      toId: toUser,
      createdAt: Date.now(),
    };

    await pushDM(dmId, msg);

    // Send to sender
    socket.emit("dm:message", msg);

    // Send to receiver if online
    const toSocketId = await getUserSocket(toUser);
    if (toSocketId) {
      io.to(toSocketId).emit("dm:message", msg);
    }
  });

  // DM REACTIONS
  socket.on("dm:react", async ({ toUser, messageId, emoji }) => {
    const fromId = socket.data.user?.id;

    if (!fromId || !toUser || !messageId || !emoji) return;

    const dmId = makeDmId(fromId, toUser);

    await toggleReaction(dmId, messageId, emoji, fromId);

    // Send to both users
    socket.emit("dm:reaction", { messageId, emoji, userId: fromId }); // to sender

    const toSocketId = await getUserSocket(toUser); //to receiver
    if (toSocketId) {
      io.to(toSocketId).emit("dm:reaction", { messageId, emoji, userId: fromId });
    }
  });

  // DM TYPING INDICATOR — mirrors typing.socket.controller.js's room-scoped
  // pattern; no equivalent existed for DMs since the reference app only had
  // room-scoped typing.
  socket.on("dm:typing", async ({ toUser, isTyping }) => {
    const fromId = socket.data.user?.id;
    if (!fromId || !toUser) return;

    const toSocketId = await getUserSocket(toUser);
    if (toSocketId) {
      io.to(toSocketId).emit("dm:typing:status", { fromId, isTyping: !!isTyping });
    }
  });
}
