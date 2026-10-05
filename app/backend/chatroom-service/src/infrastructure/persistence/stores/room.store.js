import { dataSource } from "../../../../config/db.js";
import { RoomEntity } from "../entities/room.entity.js";
import { redisPub as r } from "../../../../config/redis.js";

const ROOMS_KEY = "app:rooms";

const defaultRooms = [
  { id: "general", name: "General", icon: "💬", users: 0 },
  { id: "gaming", name: "Gaming", icon: "🎮", users: 0 },
  { id: "music", name: "Music", icon: "🎵", users: 0 },
];

export async function getAllRooms() {
  // Try to get from Redis cache first
  const roomsData = await r.get(ROOMS_KEY);
  if (roomsData) {
    try {
      return JSON.parse(roomsData);
    } catch (err) {
      console.error("Failed to parse rooms cache:", err);
    }
  }

  // Fallback to TypeORM
  const roomRepo = dataSource.getRepository(RoomEntity);
  const rooms = await roomRepo.find();
  if (rooms.length === 0) {
    // Initialize with default rooms
    await roomRepo.save(defaultRooms);
    await r.set(ROOMS_KEY, JSON.stringify(defaultRooms));
    return defaultRooms;
  }

  // Set cache
  await r.set(ROOMS_KEY, JSON.stringify(rooms));
  return rooms;
}

export async function addRoom(room) {
  const rooms = await getAllRooms();
  if (rooms.some((r) => r.id === room.id)) return rooms;

  const roomRepo = dataSource.getRepository(RoomEntity);
  await roomRepo.save({
    id: room.id,
    name: room.name,
    icon: room.icon,
    users: room.users || 0,
  });
  
  // Refresh cache
  const updatedRooms = await roomRepo.find();
  await r.set(ROOMS_KEY, JSON.stringify(updatedRooms));
  return updatedRooms;
}

export async function updateUsers(roomId, count) {
  const roomRepo = dataSource.getRepository(RoomEntity);
  await roomRepo.update(roomId, { users: count });
  
  // Refresh cache
  const rooms = await roomRepo.find();
  await r.set(ROOMS_KEY, JSON.stringify(rooms));
  return rooms;
}

export async function removeRoom(roomId) {
  const roomRepo = dataSource.getRepository(RoomEntity);
  await roomRepo.delete(roomId);
  
  // Refresh cache
  const rooms = await roomRepo.find();
  await r.set(ROOMS_KEY, JSON.stringify(rooms));
  return rooms;
}