import { joinRoom, leaveRoom } from "../../services/room.service.js";
import { addChat, getChatHistory } from "../../services/chat.service.js";
import { setOnline, setOffline } from "../../stores/presence.store.js";
import { updateRoomUserCount, getRoomList } from "../../services/room.service.js";

export function roomSocketController(io, socket) {
  const sysMsg = async (roomId, text) => {// system message helper
    const msg = {
      id: `${Date.now()}-${Math.random()}`,
      roomId,
      type: "system",
      text,
      user: "system",
      createdAt: Date.now(),
    };
    await addChat(roomId, msg); // store message
    io.to(roomId).emit("chat:message", msg);// broadcast to room
  };

  socket.on("room:join", async ({ roomId, user }) => {// join room
    const name = user?.name?.trim();// sanitize
    if (!roomId || !name) return;// validate

    socket.data.roomId = roomId;// store roomId in socket session
    // NOTE: overwrites the verified-JWT identity from auth.middleware.js with
    // client-supplied `name` — inert today since the DM-only frontend never
    // emits room:join; would need reconciling with socket.data.user.id if
    // rooms are ever revived.
    socket.data.user = { name };// store user info
    socket.data.leftManually = false;// when the user leaves manually

    // presence mapping
    await setOnline(name, socket.id); 

    socket.join(roomId);

    // update room user list
    const users = await joinRoom(roomId, socket.id, { name });
    io.to(roomId).emit("room:users", users);

    // UPDATE ROOM USER COUNT
    await updateRoomUserCount(roomId, users.length);
    const rooms = await getRoomList();
    io.emit("rooms:list", rooms); // Broadcast to all clients

    // send history to late joiner
    const history = await getChatHistory(roomId);
    socket.emit("chat:history", history);

    // system join message
    await sysMsg(roomId, `${name} joined the room`);
  });

  // manual leave (disconnect button)
  socket.on("room:leave", async () => {
    const roomId = socket.data.roomId;
    const name = socket.data.user?.name;
    if (!roomId || !name) return;

    console.log(`[room:leave] ${name} manually leaving room ${roomId}`);

    socket.data.leftManually = true;

    // cleanup presence FIRST (before leaving room)
    await setOffline(name, socket.id);

    // update users list
    const users = await leaveRoom(roomId, socket.id);
    
    // UPDATE ROOM USER COUNT
    await updateRoomUserCount(roomId, users.length);
    const rooms = await getRoomList();
    io.emit("rooms:list", rooms); // Broadcast to all clients
    
    // notify room BEFORE socket leaves
    io.to(roomId).emit("room:users", users);

    // system left message
    await sysMsg(roomId, `${name} left the room`);

    // leave room in socket.io
    socket.leave(roomId);

    // clear server-side state
    socket.data.roomId = null;
    socket.data.user = null;
  });

  // tab close / refresh / sudden disconnect
  socket.on("disconnecting", async () => {
    // if we already handled leave via button, skip
    if (socket.data.leftManually) {
      console.log(`[disconnecting] ${socket.data.user?.name} - already handled via room:leave`);
      return;
    }

    const roomId = socket.data.roomId;
    const name = socket.data.user?.name;
    if (!roomId || !name) return;

    console.log(`[disconnecting] ${name} disconnecting from room ${roomId}`);

    await setOffline(name, socket.id);

    const users = await leaveRoom(roomId, socket.id);
    
    //UPDATE ROOM USER COUNT
    await updateRoomUserCount(roomId, users.length);
    const rooms = await getRoomList();
    io.emit("rooms:list", rooms); // Broadcast to all clients
    
    io.to(roomId).emit("room:users", users);

    await sysMsg(roomId, `${name} left the room`);
  });

  // handle dissonnect for final cleanup
  socket.on("disconnect", async () => {
    if (socket.data.leftManually) return;

    const roomId = socket.data.roomId;
    const name = socket.data.user?.name;
    
    if (roomId && name) {
      console.log(`[disconnect] Final cleanup for ${name}`);
      await setOffline(name, socket.id);
    }
  });
}