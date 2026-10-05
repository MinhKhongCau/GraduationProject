import { jest } from "@jest/globals";
import { 
  getRoomList, 
  getActiveRooms, 
  createRoom, 
  updateRoomUserCount, 
  deleteRoom, 
  joinRoom, 
  leaveRoom, 
  getRoomUsers 
} from "../src/application/services/room.service.js";
import { redisPub as r } from "../config/redis.js";
import { dataSource } from "../config/db.js";

describe("Room Service Unit Tests", () => {
  let mockRoomRepository;

  beforeEach(() => {
    mockRoomRepository = {
      find: jest.fn().mockResolvedValue([]),
      save: jest.fn().mockImplementation((val) => Promise.resolve(val)),
      update: jest.fn().mockResolvedValue({}),
      delete: jest.fn().mockResolvedValue({}),
    };

    jest.spyOn(dataSource, "getRepository").mockImplementation(() => mockRoomRepository);
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  test("getRoomList & getActiveRooms - should fetch all rooms from Redis", async () => {
    const mockRooms = [{ id: "general", name: "General", icon: "💬", users: 0 }];
    jest.spyOn(r, "get").mockResolvedValue(JSON.stringify(mockRooms));

    const list = await getRoomList();
    const active = await getActiveRooms();

    expect(list).toEqual(mockRooms);
    expect(active).toEqual(mockRooms);
    expect(r.get).toHaveBeenCalledWith("app:rooms");
  });

  test("createRoom - should add room to Redis", async () => {
    jest.spyOn(r, "get").mockResolvedValue(JSON.stringify([]));
    jest.spyOn(r, "set").mockResolvedValue("OK");
    mockRoomRepository.find.mockResolvedValue([{ id: "new-room", name: "New Room", icon: "🚀", users: 0 }]);

    await createRoom("new-room", "New Room", "🚀");
    expect(r.set).toHaveBeenCalledWith("app:rooms", expect.stringContaining("New Room"));
  });

  test("updateRoomUserCount - should update users count in Redis", async () => {
    const mockRooms = [{ id: "general", name: "General", icon: "💬", users: 0 }];
    jest.spyOn(r, "get").mockResolvedValue(JSON.stringify(mockRooms));
    jest.spyOn(r, "set").mockResolvedValue("OK");
    mockRoomRepository.find.mockResolvedValue([{ id: "general", name: "General", icon: "💬", users: 10 }]);

    await updateRoomUserCount("general", 10);
    expect(r.set).toHaveBeenCalledWith("app:rooms", expect.stringContaining('"users":10'));
  });

  test("deleteRoom - should remove room from Redis", async () => {
    const mockRooms = [
      { id: "general", name: "General", icon: "💬", users: 0 },
      { id: "gaming", name: "Gaming", icon: "🎮", users: 0 }
    ];
    jest.spyOn(r, "get").mockResolvedValue(JSON.stringify(mockRooms));
    jest.spyOn(r, "set").mockResolvedValue("OK");
    mockRoomRepository.find.mockResolvedValue([{ id: "general", name: "General", icon: "💬", users: 0 }]);

    await deleteRoom("gaming");
    expect(r.set).toHaveBeenCalledWith("app:rooms", expect.stringContaining("general"));
    expect(r.set).not.value = expect.stringContaining("gaming");
  });

  test("joinRoom, leaveRoom, getRoomUsers - should manage user map correctly", async () => {
    const user1 = { id: "user1", name: "User 1" };
    const user2 = { id: "user2", name: "User 2" };

    // Initial users list should be empty
    expect(getRoomUsers("room1")).toEqual([]);

    // Join user 1
    const listAfterJoin1 = await joinRoom("room1", "socket1", user1);
    expect(listAfterJoin1).toEqual([user1]);
    expect(getRoomUsers("room1")).toEqual([user1]);

    // Join user 2
    const listAfterJoin2 = await joinRoom("room1", "socket2", user2);
    expect(listAfterJoin2).toEqual([user1, user2]);

    // Leave user 1
    const listAfterLeave1 = await leaveRoom("room1", "socket1");
    expect(listAfterLeave1).toEqual([user2]);

    // Leave user 2 (room becomes empty)
    const listAfterLeave2 = await leaveRoom("room1", "socket2");
    expect(listAfterLeave2).toEqual([]);
    expect(getRoomUsers("room1")).toEqual([]);
  });

  test("leaveRoom - should return empty array if room does not exist", async () => {
    const list = await leaveRoom("non-existent-room", "socket1");
    expect(list).toEqual([]);
  });
});
