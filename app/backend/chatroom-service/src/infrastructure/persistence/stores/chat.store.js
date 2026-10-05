import { dataSource } from "../../../../config/db.js";
import { MessageEntity } from "../entities/message.entity.js";
import { redisPub as r } from "../../../../config/redis.js";
import { LessThan } from "typeorm";

const key = (roomId) => `room:${roomId}:chat`;

export async function pushMessage(roomId, message) {
  // 1. Save to PostgreSQL using TypeORM
  const messageRepo = dataSource.getRepository(MessageEntity);
  await messageRepo.save({
    id: message.id,
    room_id: roomId,
    type: message.type,
    text: message.text || null,
    audio: message.audio || null,
    duration: message.duration || 0,
    mime_type: message.mimeType || null,
    user_name: message.user || null,
    from_id: message.from || message.fromId,
    created_at: String(message.createdAt),
  });

  // 2. Push to Redis (keeping N = 100 messages newest)
  const redisKey = key(roomId);
  await r.lPush(redisKey, JSON.stringify(message));
  await r.lTrim(redisKey, 0, 99); // Keep latest 100 messages
  await r.expire(redisKey, 60 * 60 * 24); // 24 hours TTL
}

// Convert DB message to UI message format
function mapToUiMessage(msg) {
  return {
    id: msg.id,
    roomId: msg.room_id,
    type: msg.type,
    text: msg.text,
    audio: msg.audio,
    duration: msg.duration,
    mimeType: msg.mime_type,
    user: msg.user_name,
    from: msg.from_id,
    createdAt: Number(msg.created_at),
  };
}

export async function getHistory(roomId, page = 1, limit = 20, lastMessageId = null) {
  const redisKey = key(roomId);
  const messageRepo = dataSource.getRepository(MessageEntity);

  // Case 2: Cursor pagination (lastMessageId is provided)
  if (lastMessageId) {
    // Query directly from PostgreSQL using cursor
    const lastMsg = await messageRepo.findOneBy({ id: lastMessageId, room_id: roomId });
    if (!lastMsg) {
      return []; // message not found
    }

    const messages = await messageRepo.find({
      where: {
        room_id: roomId,
        created_at: LessThan(lastMsg.created_at),
      },
      order: {
        created_at: "DESC",
      },
      take: limit,
    });
    return messages.map(mapToUiMessage).reverse();
  }

  // Case 1: First page (page = 1 or no lastMessageId)
  if (page === 1) {
    // Try to get from Redis
    const items = await r.lRange(redisKey, 0, limit - 1);
    if (items && items.length > 0) {
      return items.map((x) => JSON.parse(x)).reverse();
    }

    // Redis MISS: Query PostgreSQL for 20 newest messages
    const messages = await messageRepo.find({
      where: { room_id: roomId },
      order: { created_at: "DESC" },
      take: limit,
    });

    // Warm up Redis cache: Query top 100 newest messages and reload them to Redis
    const populateMessages = await messageRepo.find({
      where: { room_id: roomId },
      order: { created_at: "DESC" },
      take: 100,
    });
    if (populateMessages.length > 0) {
      await r.del(redisKey);
      const toPush = populateMessages.map(mapToUiMessage).reverse();
      for (const msg of toPush) {
        await r.lPush(redisKey, JSON.stringify(msg));
      }
      await r.lTrim(redisKey, 0, 99);
      await r.expire(redisKey, 60 * 60 * 24);
    }

    return messages.map(mapToUiMessage).reverse();
  }

  // Fallback: Page-based offset query on PostgreSQL if page > 1 but no lastMessageId
  const offset = (page - 1) * limit;
  const messages = await messageRepo.find({
    where: { room_id: roomId },
    order: { created_at: "DESC" },
    skip: offset,
    take: limit,
  });
  return messages.map(mapToUiMessage).reverse();
}