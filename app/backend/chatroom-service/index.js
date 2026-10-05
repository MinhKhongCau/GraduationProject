import http from "http";
import { createApp } from "./src/app.js";
import { ENV } from "./config/env.js";
import { connectRedis } from "./config/redis.js";
import { initSocket } from "./src/infrastructure/socket/socket.js";
import { initDb } from "./config/db.js";

const app = createApp();
const server = http.createServer(app);

await connectRedis();
await initDb();
initSocket(server);

server.listen(ENV.PORT, () => {
  console.log(`✅ Server running on http://localhost:${ENV.PORT}`);
});