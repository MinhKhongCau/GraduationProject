import { EVENTS } from "../events.js";
import { setSpeaking, getSpeaking } from "../../services/presence.service.js";
import { getUserSocket } from "../../stores/presence.store.js";

export function presenceSocketController(io, socket) {
  socket.on(EVENTS.PRESENCE_SPEAKING, async ({ isSpeaking }) => {
    const roomId = socket.data.roomId;
    if (!roomId) return;

    await setSpeaking(roomId, socket.id, !!isSpeaking);
    const speaking = await getSpeaking(roomId);

    io.to(roomId).emit(EVENTS.PRESENCE_SPEAKING_LIST, speaking);
  });

  // Poll-based presence for the DM contact list — there's no room-based
  // broadcast to piggyback on once rooms are out of scope, and a full
  // pub/sub online/offline fan-out isn't needed for a small contact list.
  socket.on("presence:query", async ({ userIds }) => {
    const pairs = await Promise.all(
      (userIds || []).map(async (id) => [id, !!(await getUserSocket(id))])
    );
    socket.emit("presence:status", Object.fromEntries(pairs));
  });
}
