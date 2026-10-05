import { jest } from "@jest/globals";
import { dmSocketController } from "../src/infrastructure/socket/controllers/dm.socket.controller.js";
import { redisPub as r } from "../config/redis.js";
import { dataSource } from "../config/db.js";

describe("UC-09 & UC-10: Socket DM Controller (Thực hiện & Nhận tư vấn)", () => {
  let io;
  let socket;
  let handlers;
  let mockMessageRepository;

  beforeEach(() => {
    mockMessageRepository = {
      findOneBy: jest.fn().mockResolvedValue(null),
      find: jest.fn().mockResolvedValue([]),
      save: jest.fn().mockImplementation((val) => Promise.resolve(val)),
    };
    jest.spyOn(dataSource, "getRepository").mockImplementation(() => mockMessageRepository);
    handlers = {};
    socket = {
      id: "socket-sender-id",
      data: {
        user: { id: "expert-uuid-1", email: "expert@example.com", role: "EXPERT" },
      },
      on: jest.fn((event, handler) => {
        handlers[event] = handler;
      }),
      emit: jest.fn(),
    };

    io = {
      to: jest.fn().mockReturnValue({
        emit: jest.fn(),
      }),
    };

    dmSocketController(io, socket);
  });

  test("TC-CHAT-SND-01 - Chuyên gia gửi tin nhắn tư vấn thành công (Send DM)", async () => {
    jest.spyOn(r, "lPush").mockResolvedValue(1);
    jest.spyOn(r, "lTrim").mockResolvedValue("OK");
    jest.spyOn(r, "expire").mockResolvedValue(true);
    jest.spyOn(r, "get").mockResolvedValue("socket-receiver-id");

    const recipientSocketEmit = jest.fn();
    io.to.mockReturnValue({ emit: recipientSocketEmit });

    await handlers["dm:send"]({
      toUser: "patient-uuid-1",
      text: "  Chào bạn, tôi có thể giúp gì cho bạn?  ",
    });

    expect(r.lPush).toHaveBeenCalledWith(
      "dm:expert-uuid-1:patient-uuid-1:messages",
      expect.stringContaining("Chào bạn, tôi có thể giúp gì cho bạn?")
    );

    expect(socket.emit).toHaveBeenCalledWith("dm:message", expect.objectContaining({
      text: "Chào bạn, tôi có thể giúp gì cho bạn?",
    }));

    expect(io.to).toHaveBeenCalledWith("socket-receiver-id");
    expect(recipientSocketEmit).toHaveBeenCalledWith("dm:message", expect.any(Object));

    r.lPush.mockRestore();
    r.lTrim.mockRestore();
    r.expire.mockRestore();
    r.get.mockRestore();
  });

  test("TC-CHAT-SND-02 - Gửi tin nhắn thất bại do nội dung văn bản rỗng", async () => {
    jest.spyOn(r, "lPush").mockResolvedValue(1);

    await handlers["dm:send"]({
      toUser: "patient-uuid-1",
      text: "     ",
    });

    expect(r.lPush).not.toHaveBeenCalled();
    expect(socket.emit).not.toHaveBeenCalled();

    r.lPush.mockRestore();
  });

  test("TC-CHAT-SND-03 - Gửi tin nhắn thất bại do thiếu thông tin người nhận", async () => {
    jest.spyOn(r, "lPush").mockResolvedValue(1);

    await handlers["dm:send"]({
      toUser: "",
      text: "Tin nhắn không có người nhận",
    });

    expect(r.lPush).not.toHaveBeenCalled();
    expect(socket.emit).not.toHaveBeenCalled();

    r.lPush.mockRestore();
  });

  test("TC-CHAT-SND-04 - Chuyên gia gửi tin nhắn thoại tư vấn thành công (Send Voice DM)", async () => {
    jest.spyOn(r, "lPush").mockResolvedValue(1);
    jest.spyOn(r, "lTrim").mockResolvedValue("OK");
    jest.spyOn(r, "expire").mockResolvedValue(true);
    jest.spyOn(r, "get").mockResolvedValue("socket-receiver-id");

    await handlers["dm:send:voice"]({
      toUser: "patient-uuid-1",
      audio: "data:audio/webm;base64,GkXfo59ChoEBQveBA...",
      duration: 15,
      mimeType: "audio/webm",
    });

    expect(r.lPush).toHaveBeenCalledWith(
      "dm:expert-uuid-1:patient-uuid-1:messages",
      expect.stringContaining('"type":"voice"')
    );

    expect(socket.emit).toHaveBeenCalledWith("dm:message", expect.objectContaining({
      type: "voice",
    }));

    r.lPush.mockRestore();
    r.lTrim.mockRestore();
    r.expire.mockRestore();
    r.get.mockRestore();
  });

  test("TC-CHAT-SND-05 - Thả biểu cảm trên tin nhắn tư vấn thành công (React DM)", async () => {
    jest.spyOn(r, "get").mockImplementation(async (key) => {
      if (key.includes("presence")) return "socket-receiver-id";
      return JSON.stringify({});
    });
    jest.spyOn(r, "set").mockResolvedValue("OK");

    await handlers["dm:react"]({
      toUser: "patient-uuid-1",
      messageId: "msg-123",
      emoji: "❤️",
    });

    expect(socket.emit).toHaveBeenCalledWith("dm:reaction", {
      messageId: "msg-123",
      emoji: "❤️",
      userId: "expert-uuid-1",
    });

    r.get.mockRestore();
    r.set.mockRestore();
  });

  test("TC-CHAT-REC-02 - Gửi tín hiệu gõ bàn phím (DM Typing Indicator)", async () => {
    jest.spyOn(r, "get").mockResolvedValue("socket-receiver-id");

    const receiverEmit = jest.fn();
    io.to.mockReturnValue({ emit: receiverEmit });

    await handlers["dm:typing"]({
      toUser: "patient-uuid-1",
      isTyping: true,
    });

    expect(io.to).toHaveBeenCalledWith("socket-receiver-id");
    expect(receiverEmit).toHaveBeenCalledWith("dm:typing:status", {
      fromId: "expert-uuid-1",
      isTyping: true,
    });

    r.get.mockRestore();
  });
});
