import * as store from "../../infrastructure/persistence/stores/chat.store.js";

export async function addChat(roomId, message) {
  await store.pushMessage(roomId, message);
}

export async function getChatHistory(roomId, page = 1, limit = 20, lastMessageId = null) {
  return store.getHistory(roomId, page, limit, lastMessageId);
}