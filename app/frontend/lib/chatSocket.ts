import { io, type Socket } from "socket.io-client";
import { getAccessToken } from "@/api/http/session";

const WS_URL = process.env.NEXT_PUBLIC_CHATROOM_WS_URL ?? "http://localhost:8085";
const GLOBAL_KEY = "__mindcareChatSocket__";

/**
 * HMR/strict-mode-safe singleton — one Socket.IO connection per tab, shared
 * across every consumer of ChatContext. `auth` is a callback (not a static
 * object) so socket.io-client re-reads the access token on every (re)connect
 * attempt, picking up a token the axios interceptor refreshed meanwhile.
 */
export function getChatSocket(): Socket {
  const g = globalThis as unknown as Record<string, Socket | undefined>;
  if (g[GLOBAL_KEY]) return g[GLOBAL_KEY]!;

  const socket = io(WS_URL, {
    autoConnect: false,
    transports: ["websocket", "polling"],
    reconnection: true,
    reconnectionAttempts: 20,
    reconnectionDelay: 500,
    timeout: 20000,
    auth: (cb) => cb({ token: getAccessToken() }),
  });

  g[GLOBAL_KEY] = socket;
  return socket;
}

/** Same order-independent pairing scheme as chatroom-service's dm.store.js. */
export function makeDmId(accountIdA: string, accountIdB: string): string {
  const [x, y] = [accountIdA, accountIdB].sort();
  return `${x}:${y}`;
}
