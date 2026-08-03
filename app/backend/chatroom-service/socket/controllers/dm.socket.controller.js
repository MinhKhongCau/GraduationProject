import { getDMHistory, pushDM, makeDmId } from "../../stores/dm.store.js";
import { getUserSocket } from "../../stores/presence.store.js";
import { toggleReaction } from "../../services/reactions.service.js";
import { CHATBOT_ID } from "../../types/index.js";
import { ENV } from "../../config/env.js";

// Helper function to read the chatbot API streaming response and emit chunks to the client
async function handleChatbotStream(fromId, toUser, dmId, response, socket) {
  try {
    if (!response.ok) {
      throw new Error(`Chatbot API error: ${response.status} ${response.statusText}`);
    }

    let accumulatedText = "";
    const botMsgId = `${Date.now()}-${Math.random()}`;
    const botMsg = {
      id: botMsgId,
      dmId,
      type: "chat",
      text: "",
      fromId: toUser,
      toId: fromId,
      createdAt: Date.now(),
    };

    // Emit initial empty message to sender so UI can instantiate the bubble
    socket.emit("dm:message", botMsg);

    if (response.body && typeof response.body.getReader === "function") {
      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        const chunk = decoder.decode(value, { stream: true });
        accumulatedText += chunk;
        botMsg.text = accumulatedText;
        socket.emit("dm:message", botMsg);
      }
    } else if (response.body) {
      // Node.js Readable stream fallback
      for await (const chunk of response.body) {
        accumulatedText += chunk.toString();
        botMsg.text = accumulatedText;
        socket.emit("dm:message", botMsg);
      }
    } else {
      throw new Error("Chatbot API returned empty body");
    }

    // Save completed message to DB
    await pushDM(dmId, botMsg);
  } catch (error) {
    console.error("Error reading chatbot stream:", error);
    const errorMsg = {
      id: `chatbot-msg-err-${Date.now()}`,
      dmId,
      type: "chat",
      text: "Sorry, I am having trouble connecting right now. Please try again later.",
      fromId: toUser,
      toId: fromId,
      createdAt: Date.now(),
    };
    socket.emit("dm:message", errorMsg);
  } finally {
    // Stop typing indicator
    socket.emit("dm:typing:status", { fromId: toUser, isTyping: false });
  }
}

// NOTE: `toUser` on every incoming payload is the recipient's accountId (a
// JWT `accountId` claim), not a display name — kept the field name as-is to
// minimize event-shape churn, only the semantics changed. The sender's
// identity (`fromId`) always comes from socket.data.user, set by the
// verified-JWT auth.middleware.js, never from client input.

export function dmSocketController(io, socket) {
  // DM history
  socket.on("dm:history", async ({ toUser, page = 1, limit = 20, lastMessageId = null }) => {
    const fromId = socket.data.user?.id;

    if (!fromId || !toUser) return;

    const dmId = makeDmId(fromId, toUser);
    const history = await getDMHistory(dmId, page, limit, lastMessageId);

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

    const isChatbot = toUser === CHATBOT_ID || toUser.startsWith(CHATBOT_ID + ":");

    if (isChatbot) {
      // Trigger chatbot stream API
      socket.emit("dm:typing:status", { fromId: toUser, isTyping: true });
      try {
        const response = await fetch(`${ENV.CHATBOT_URL}/api/v1/chat/stream`, {
          method: "POST",
          headers: {
            "Accept": "application/json",
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            message: clean,
            session_id: fromId,
          }),
        });
        await handleChatbotStream(fromId, toUser, dmId, response, socket);
      } catch (error) {
        console.error("Error calling chatbot text API:", error);
        socket.emit("dm:typing:status", { fromId: toUser, isTyping: false });
        socket.emit("dm:message", {
          id: `chatbot-msg-err-${Date.now()}`,
          dmId,
          type: "chat",
          text: "Sorry, I am having trouble connecting right now. Please try again later.",
          fromId: toUser,
          toId: fromId,
          createdAt: Date.now(),
        });
      }
    } else {
      // Send to receiver if online
      const toSocketId = await getUserSocket(toUser);
      if (toSocketId) {
        io.to(toSocketId).emit("dm:message", msg);
      }
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

    const isChatbot = toUser === CHATBOT_ID || toUser.startsWith(CHATBOT_ID + ":");

    if (isChatbot) {
      socket.emit("dm:typing:status", { fromId: toUser, isTyping: true });
      try {
        const buffer = Buffer.from(audio, "base64");
        const blob = new Blob([buffer], { type: mimeType || "audio/webm" });
        const file = new File([blob], "voice.webm", { type: mimeType || "audio/webm" });

        const formData = new FormData();
        formData.append("audio_file", file);
        formData.append("session_id", fromId);

        const response = await fetch(`${ENV.CHATBOT_URL}/api/v1/chat/voice`, {
          method: "POST",
          body: formData,
        });
        await handleChatbotStream(fromId, toUser, dmId, response, socket);
      } catch (error) {
        console.error("Error calling chatbot voice API:", error);
        socket.emit("dm:typing:status", { fromId: toUser, isTyping: false });
        socket.emit("dm:message", {
          id: `chatbot-msg-err-${Date.now()}`,
          dmId,
          type: "chat",
          text: "Sorry, I am having trouble processing your voice message right now. Please try again later.",
          fromId: toUser,
          toId: fromId,
          createdAt: Date.now(),
        });
      }
    } else {
      // Send to receiver if online
      const toSocketId = await getUserSocket(toUser);
      if (toSocketId) {
        io.to(toSocketId).emit("dm:message", msg);
      }
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
