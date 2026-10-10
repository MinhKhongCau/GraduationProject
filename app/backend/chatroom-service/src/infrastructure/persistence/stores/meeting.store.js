import { redisPub as r } from "../../../../config/redis.js";

// Expert presence per appointment, shared by every chatroom-service instance:
//   meet:<appointmentId>:expert  (hash)
//     socket:<socketId>  joinedAt of each open expert socket
//     activeCount        number of open expert sockets
//     activeSince        when the expert (first socket) entered the room
//     accumulatedMs      presence from earlier, already closed intervals
// Overlapping sockets (two tabs) count once: an interval opens when the first
// expert socket joins and closes when the last one leaves.
const key = (appointmentId) => `meet:${appointmentId}:expert`;
const endedKey = (appointmentId) => `meet:${appointmentId}:ended`;
const TTL_SECONDS = 24 * 60 * 60;

const JOIN_SCRIPT = `
if redis.call('HEXISTS', KEYS[1], 'socket:' .. ARGV[1]) == 1 then return 0 end
redis.call('HSET', KEYS[1], 'socket:' .. ARGV[1], ARGV[2])
local active = tonumber(redis.call('HGET', KEYS[1], 'activeCount') or '0')
if active <= 0 then redis.call('HSET', KEYS[1], 'activeSince', ARGV[2]) end
redis.call('HSET', KEYS[1], 'activeCount', active + 1)
redis.call('EXPIRE', KEYS[1], ARGV[3])
return 1
`;

const LEAVE_SCRIPT = `
if redis.call('HDEL', KEYS[1], 'socket:' .. ARGV[1]) == 0 then return 0 end
local active = tonumber(redis.call('HGET', KEYS[1], 'activeCount') or '1') - 1
if active <= 0 then
  local since = tonumber(redis.call('HGET', KEYS[1], 'activeSince') or ARGV[2])
  local now = tonumber(ARGV[2])
  if now > since then redis.call('HINCRBY', KEYS[1], 'accumulatedMs', now - since) end
  redis.call('HDEL', KEYS[1], 'activeSince')
  active = 0
end
redis.call('HSET', KEYS[1], 'activeCount', active)
return 1
`;

// countFromMs: presence before the slot starts does not count, so join/leave
// timestamps are clamped to the slot start time.
const effectiveTime = (nowMs, countFromMs) => Math.max(nowMs, countFromMs || 0);

export async function expertJoined(appointmentId, socketId, nowMs, countFromMs) {
  await r.eval(JOIN_SCRIPT, {
    keys: [key(appointmentId)],
    arguments: [socketId, String(effectiveTime(nowMs, countFromMs)), String(TTL_SECONDS)],
  });
}

export async function expertLeft(appointmentId, socketId, nowMs, countFromMs) {
  await r.eval(LEAVE_SCRIPT, {
    keys: [key(appointmentId)],
    arguments: [socketId, String(effectiveTime(nowMs, countFromMs))],
  });
}

export async function getExpertPresenceMs(appointmentId, nowMs, countFromMs) {
  const [accumulated, activeSince] = await r.hmGet(key(appointmentId), ["accumulatedMs", "activeSince"]);
  let total = Number(accumulated || 0);
  if (activeSince) {
    total += Math.max(0, effectiveTime(nowMs, countFromMs) - Number(activeSince));
  }
  return total;
}

export async function markEnded(appointmentId) {
  await r.set(endedKey(appointmentId), "1", { EX: TTL_SECONDS });
}

export async function isEnded(appointmentId) {
  return (await r.get(endedKey(appointmentId))) === "1";
}
