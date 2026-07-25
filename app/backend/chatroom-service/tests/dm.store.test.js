import { jest } from "@jest/globals";
import { makeDmId, pushDM, getDMHistory } from "../stores/dm.store.js";
import { redisPub as r } from "../config/redis.js";

describe("UC-19: DM Store & Redis History Management", () => {
  test("TC-CHAT-MNG-01 - Tạo DM ID đồng bộ hai chiều bất kể thứ tự tham số", () => {
    const id1 = makeDmId("account-111", "account-222");
    const id2 = makeDmId("account-222", "account-111");

    expect(id1).toBe("account-111:account-222");
    expect(id2).toBe("account-111:account-222");
    expect(id1).toBe(id2);
  });

  test("TC-CHAT-MNG-04 - Tạo DM ID với mã UUID hệ thống thực tế", () => {
    const uuid1 = "b1111111-1111-1111-1111-111111111111";
    const uuid2 = "a2222222-2222-2222-2222-222222222222";

    const dmId = makeDmId(uuid1, uuid2);
    expect(dmId).toBe("a2222222-2222-2222-2222-222222222222:b1111111-1111-1111-1111-111111111111");
  });

  test("TC-CHAT-MNG-03 - Lưu tin nhắn DM mới vào Redis và giới hạn tối đa 100 tin nhắn (pushDM & lTrim)", async () => {
    jest.spyOn(r, "lPush").mockResolvedValue(1);
    jest.spyOn(r, "lTrim").mockResolvedValue("OK");
    jest.spyOn(r, "expire").mockResolvedValue(true);

    const message = {
      id: "msg-1",
      text: "Xin chào",
      fromId: "account-111",
      toId: "account-222",
      createdAt: 1700000000000,
    };

    await pushDM("account-111:account-222", message);

    expect(r.lPush).toHaveBeenCalledWith("dm:account-111:account-222:messages", JSON.stringify(message));
    expect(r.lTrim).toHaveBeenCalledWith("dm:account-111:account-222:messages", 0, 99);
    expect(r.expire).toHaveBeenCalledWith("dm:account-111:account-222:messages", 604800);

    r.lPush.mockRestore();
    r.lTrim.mockRestore();
    r.expire.mockRestore();
  });

  test("TC-CHAT-MNG-02 - Lấy lịch sử tin nhắn từ Redis và đảo ngược mảng theo thứ tự thời gian tăng dần (getDMHistory & reverse)", async () => {
    const mockRedisItems = [
      JSON.stringify({ id: "msg-2", text: "Tin nhắn mới hơn (LIFO index 0)", createdAt: 1700000002000 }),
      JSON.stringify({ id: "msg-1", text: "Tin nhắn cũ hơn (LIFO index 1)", createdAt: 1700000001000 }),
    ];

    jest.spyOn(r, "lRange").mockResolvedValue(mockRedisItems);

    const history = await getDMHistory("account-111:account-222");

    expect(r.lRange).toHaveBeenCalledWith("dm:account-111:account-222:messages", 0, 99);
    expect(history).toHaveLength(2);
    // Sau khi reverse(), msg-1 (cũ) đứng trước msg-2 (mới)
    expect(history[0].id).toBe("msg-1");
    expect(history[1].id).toBe("msg-2");

    r.lRange.mockRestore();
  });

  test("TC-CHAT-MNG-05 - Xử lý lấy lịch sử DM khi Redis trả về danh sách rỗng", async () => {
    jest.spyOn(r, "lRange").mockResolvedValue([]);

    const history = await getDMHistory("new-dm-id");

    expect(history).toEqual([]);

    r.lRange.mockRestore();
  });
});
