import { redisPub as r } from "../config/redis.js";

const key = (roomId) => `room:${roomId}:chat`;

export async function pushMessage(roomId, message) {
  await r.lPush(key(roomId), JSON.stringify(message));
  await r.lTrim(key(roomId), 0, 499);
  await r.expire(key(roomId), 60 * 60 * 24);
}

export async function getHistory(roomId, start = 0, end = 19) {
  const items = await r.lRange(key(roomId), start, end);
  return items.map((x) => JSON.parse(x)).reverse();
}