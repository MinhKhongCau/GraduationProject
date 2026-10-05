import { dataSource } from "../../../../config/db.js";
import { MessageEntity } from "../entities/message.entity.js";
import { redisPub as r } from "../../../../config/redis.js";
import { LessThan } from "typeorm";

// Single source of truth for DM id derivation — keyed by accountId pair,
// order-independent, so either side computes the same id.
export const makeDmId = (accountIdA, accountIdB) => {
  const [x, y] = [accountIdA, accountIdB].sort();
  return `${x}:${y}`;
};

const key = (dmId) => `dm:${dmId}:messages`;

export async function pushDM(dmId, message) {
  // 1. Save to PostgreSQL using TypeORM
  const messageRepo = dataSource.getRepository(MessageEntity);
  await messageRepo.save({
    id: message.id,
    dm_id: dmId,
    type: message.type,
    text: message.text || null,
    audio: message.audio || null,
    duration: message.duration || 0,
    mime_type: message.mimeType || null,
    from_id: message.fromId,
    to_id: message.toId,
    created_at: String(message.createdAt),
  });

  // 2. Push to Redis (keeping N = 100 messages newest)
  const redisKey = key(dmId);
  await r.lPush(redisKey, JSON.stringify(message));
  await r.lTrim(redisKey, 0, 99); // Keep latest 100 messages
  await r.expire(redisKey, 60 * 60 * 24 * 7); // 7 days TTL for DMs
}

// Convert DB message to UI DM format
function mapToUiDM(msg) {
  return {
    id: msg.id,
    dmId: msg.dm_id,
    type: msg.type,
    text: msg.text,
    audio: msg.audio,
    duration: msg.duration,
    mimeType: msg.mime_type,
    fromId: msg.from_id,
    toId: msg.to_id,
    createdAt: Number(msg.created_at),
  };
}

export async function getDMHistory(dmId, page = 1, limit = 20, lastMessageId = null) {
  const redisKey = key(dmId);
  const messageRepo = dataSource.getRepository(MessageEntity);

  // Case 2: Cursor pagination (lastMessageId is provided)
  if (lastMessageId) {
    // Query directly from PostgreSQL using cursor
    const lastMsg = await messageRepo.findOneBy({ id: lastMessageId, dm_id: dmId });
    if (!lastMsg) {
      return []; // message not found
    }

    const messages = await messageRepo.find({
      where: {
        dm_id: dmId,
        created_at: LessThan(lastMsg.created_at),
      },
      order: {
        created_at: "DESC",
      },
      take: limit,
    });
    return messages.map(mapToUiDM).reverse();
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
      where: { dm_id: dmId },
      order: {
        created_at: "DESC",
      },
      take: limit,
    });

    // Warm up Redis cache: Query top 100 newest messages and reload them to Redis
    const populateMessages = await messageRepo.find({
      where: { dm_id: dmId },
      order: {
        created_at: "DESC",
      },
      take: 100,
    });
    if (populateMessages.length > 0) {
      await r.del(redisKey);
      const toPush = populateMessages.map(mapToUiDM).reverse();
      for (const msg of toPush) {
        await r.lPush(redisKey, JSON.stringify(msg));
      }
      await r.lTrim(redisKey, 0, 99);
      await r.expire(redisKey, 60 * 60 * 24 * 7);
    }

    return messages.map(mapToUiDM).reverse();
  }

  // Fallback: Page-based offset query on PostgreSQL if page > 1 but no lastMessageId
  const offset = (page - 1) * limit;
  const messages = await messageRepo.find({
    where: { dm_id: dmId },
    order: {
      created_at: "DESC",
    },
    skip: offset,
    take: limit,
  });
  return messages.map(mapToUiDM).reverse();
}