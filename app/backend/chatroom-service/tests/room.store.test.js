import { jest } from "@jest/globals";
import { getAllRooms, addRoom, updateUsers, removeRoom } from "../src/infrastructure/persistence/stores/room.store.js";
import { redisPub as r } from "../config/redis.js";
import { dataSource } from "../config/db.js";

describe("Room Store Unit Tests", () => {
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

  test("getAllRooms - should initialize default rooms if no data in Redis", async () => {
    jest.spyOn(r, "get").mockResolvedValue(null);
    jest.spyOn(r, "set").mockResolvedValue("OK");

    const rooms = await getAllRooms();
    expect(rooms).toHaveLength(3);
    expect(rooms[0].id).toBe("general");
    expect(r.get).toHaveBeenCalledWith("app:rooms");
    expect(r.set).toHaveBeenCalled();
  });

  test("getAllRooms - should return parsed rooms from Redis if present", async () => {
    const customRooms = [{ id: "custom", name: "Custom", icon: "✨", users: 5 }];
    jest.spyOn(r, "get").mockResolvedValue(JSON.stringify(customRooms));

    const rooms = await getAllRooms();
    expect(rooms).toEqual(customRooms);
    expect(r.get).toHaveBeenCalledWith("app:rooms");
  });

  test("addRoom - should not add room if it already exists", async () => {
    const customRooms = [{ id: "general", name: "General", icon: "💬", users: 0 }];
    jest.spyOn(r, "get").mockResolvedValue(JSON.stringify(customRooms));
    jest.spyOn(r, "set").mockResolvedValue("OK");

    const rooms = await addRoom({ id: "general", name: "General", icon: "💬", users: 0 });
    expect(rooms).toEqual(customRooms);
    expect(r.set).not.toHaveBeenCalled();
  });

  test("addRoom - should add new room if it does not exist", async () => {
    const initialRooms = [{ id: "general", name: "General", icon: "💬", users: 0 }];
    jest.spyOn(r, "get").mockResolvedValue(JSON.stringify(initialRooms));
    jest.spyOn(r, "set").mockResolvedValue("OK");

    const newRoom = { id: "new-room", name: "New Room", icon: "🚀", users: 0 };
    const expectedRooms = [...initialRooms, newRoom];
    mockRoomRepository.find.mockResolvedValue(expectedRooms);

    const rooms = await addRoom(newRoom);
    
    expect(rooms).toHaveLength(2);
    expect(rooms[1]).toEqual(newRoom);
    expect(r.set).toHaveBeenCalledWith("app:rooms", JSON.stringify(rooms));
  });

  test("updateUsers - should update users count for an existing room", async () => {
    const initialRooms = [{ id: "general", name: "General", icon: "💬", users: 0 }];
    jest.spyOn(r, "get").mockResolvedValue(JSON.stringify(initialRooms));
    jest.spyOn(r, "set").mockResolvedValue("OK");

    const expectedRooms = [{ id: "general", name: "General", icon: "💬", users: 5 }];
    mockRoomRepository.find.mockResolvedValue(expectedRooms);

    const rooms = await updateUsers("general", 5);
    expect(rooms[0].users).toBe(5);
    expect(r.set).toHaveBeenCalledWith("app:rooms", JSON.stringify(rooms));
  });

  test("removeRoom - should remove room by id", async () => {
    const initialRooms = [
      { id: "general", name: "General", icon: "💬", users: 0 },
      { id: "gaming", name: "Gaming", icon: "🎮", users: 0 }
    ];
    jest.spyOn(r, "get").mockResolvedValue(JSON.stringify(initialRooms));
    jest.spyOn(r, "set").mockResolvedValue("OK");

    const expectedRooms = [{ id: "general", name: "General", icon: "💬", users: 0 }];
    mockRoomRepository.find.mockResolvedValue(expectedRooms);

    const rooms = await removeRoom("gaming");
    expect(rooms).toHaveLength(1);
    expect(rooms[0].id).toBe("general");
    expect(r.set).toHaveBeenCalledWith("app:rooms", JSON.stringify(rooms));
  });
});
