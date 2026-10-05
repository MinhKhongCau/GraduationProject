// server/socket/controllers/call.socket.controller.js

const onlineUsers = new Map(); // username -> socketId

export function registerCallSocketHandlers(io, socket) {
  console.log("✅ Call handlers registered for socket:", socket.id);

  // Register user identity
  socket.on("user:online", ({ name }) => {
    if (!name) return;
    socket.data.username = name;
    onlineUsers.set(name, socket.id);
    console.log(`👤 ${name} is online (${socket.id})`);
    console.log("📊 Online users:", Array.from(onlineUsers.keys()));
  });

  const getSocketId = (username) => onlineUsers.get(username);// Helper to get socket ID by username

  // Call request (caller initiates)
  socket.on("call:request", ({ to }) => {
    const fromUsername = socket.data.username;
    const toSocketId = getSocketId(to);
    
    console.log(`📞 Call request: ${fromUsername} → ${to}`);
    console.log(`   Target socket ID: ${toSocketId}`);

    if (!toSocketId) {
      console.log(`❌ ${to} not found online`);
      socket.emit("call:unavailable", { to });
      return;
    }

    io.to(toSocketId).emit("call:incoming", { from: fromUsername });
    console.log(`✅ Call notification sent to ${to}`);
  });

  // Accept call
  socket.on("call:accept", ({ to }) => {
    const fromUsername = socket.data.username;
    const toSocketId = getSocketId(to);
    
    console.log(`✅ ${fromUsername} accepted call from ${to}`);

    if (toSocketId) {
      io.to(toSocketId).emit("call:accepted", { from: fromUsername });
    }
  });

  // Reject call
  socket.on("call:reject", ({ to }) => {
    const fromUsername = socket.data.username;
    const toSocketId = getSocketId(to);
    
    console.log(`❌ ${fromUsername} rejected call from ${to}`);

    if (toSocketId) {
      io.to(toSocketId).emit("call:rejected", { from: fromUsername });
    }
  });

  // WebRTC Offer
  socket.on("webrtc:offer", ({ to, offer }) => {
    const fromUsername = socket.data.username;
    const toSocketId = getSocketId(to);
    
    console.log(`🔄 WebRTC offer: ${fromUsername} → ${to}`);

    if (toSocketId) {
      io.to(toSocketId).emit("webrtc:offer", { from: fromUsername, offer });
    }
  });

  // WebRTC Answer
  socket.on("webrtc:answer", ({ to, answer }) => {
    const fromUsername = socket.data.username;
    const toSocketId = getSocketId(to);
    
    console.log(`🔄 WebRTC answer: ${fromUsername} → ${to}`);

    if (toSocketId) {
      io.to(toSocketId).emit("webrtc:answer", { from: fromUsername, answer });
    }
  });

  // ICE(Interactive Connectivity Establishment) Candidate
  socket.on("webrtc:ice", ({ to, candidate }) => {
    const fromUsername = socket.data.username;
    const toSocketId = getSocketId(to);

    if (toSocketId) {
      io.to(toSocketId).emit("webrtc:ice", { from: fromUsername, candidate });
    }
  });

  // End call
  socket.on("call:end", ({ to }) => {
    const fromUsername = socket.data.username;
    const toSocketId = getSocketId(to);
    
    console.log(`📴 ${fromUsername} ended call with ${to}`);

    if (toSocketId) {
      io.to(toSocketId).emit("call:ended", { from: fromUsername });
    }
  });

  // Cleanup on disconnect
  socket.on("disconnect", () => {
    const username = socket.data.username;
    if (username) {
      onlineUsers.delete(username);
      console.log(`👋 ${username} disconnected. Online users:`, Array.from(onlineUsers.keys()));
    }
  });
}