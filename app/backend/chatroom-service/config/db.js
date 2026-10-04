import "reflect-metadata";
import { DataSource } from "typeorm";
import { ENV } from "./env.js";
import { RoomEntity } from "../entities/room.entity.js";
import { MessageEntity } from "../entities/message.entity.js";
import { CreateChatroomTables1700000000000 } from "../migrations/1700000000000-CreateChatroomTables.js";
import { CreateOutboxInboxTables1791136060543 } from "../migrations/1791136060543-CreateOutboxInboxTables.js";

export const dataSource = new DataSource({
  type: "postgres",
  host: ENV.DB_HOST,
  port: ENV.DB_PORT,
  username: ENV.DB_USER,
  password: ENV.DB_PASSWORD,
  database: ENV.DB_NAME,
  ssl: ENV.DB_SSLMODE === "disable" ? false : { rejectUnauthorized: false },
  synchronize: false, // Do not synchronize automatically, use migrations!
  logging: ENV.NODE_ENV === "development" ? ["query", "error"] : ["error"],
  entities: [RoomEntity, MessageEntity],
  migrations: [CreateChatroomTables1700000000000, CreateOutboxInboxTables1791136060543],
});

export async function initDb() {
  try {
    console.log("⏳ Initializing TypeORM DataSource...");
    await dataSource.initialize();
    console.log("✅ TypeORM DataSource initialized");

    console.log("⏳ Running TypeORM migrations...");
    await dataSource.runMigrations();
    console.log("✅ TypeORM migrations finished successfully");
  } catch (err) {
    console.error("❌ Failed to initialize database:", err);
    throw err;
  }
}
