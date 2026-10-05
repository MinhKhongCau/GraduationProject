import { jest } from "@jest/globals";
import { listRooms } from "../src/infrastructure/http/controllers/rooms.controller.js";
import { redisPub as r } from "../config/redis.js";

describe("Rooms Controller Unit Tests", () => {
  afterEach(() => {
    jest.restoreAllMocks();
  });

  test("listRooms - should return active rooms list successfully", async () => {
    const mockRooms = [{ id: "general", name: "General", icon: "💬", users: 0 }];
    jest.spyOn(r, "get").mockResolvedValue(JSON.stringify(mockRooms));

    const req = {};
    const res = {
      json: jest.fn()
    };
    const next = jest.fn();

    await listRooms(req, res, next);

    expect(r.get).toHaveBeenCalledWith("app:rooms");
    expect(res.json).toHaveBeenCalledWith({ rooms: mockRooms });
    expect(next).not.toHaveBeenCalled();
  });

  test("listRooms - should call next with error if Redis get throws", async () => {
    const error = new Error("Redis connection failed");
    jest.spyOn(r, "get").mockRejectedValue(error);

    const req = {};
    const res = {
      json: jest.fn()
    };
    const next = jest.fn();

    await listRooms(req, res, next);

    expect(r.get).toHaveBeenCalledWith("app:rooms");
    expect(res.json).not.toHaveBeenCalled();
    expect(next).toHaveBeenCalledWith(error);
  });
});
