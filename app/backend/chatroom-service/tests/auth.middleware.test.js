import { jest } from "@jest/globals";
import { socketAuthMiddleware } from "../socket/auth.middleware.js";
import jwt from "jsonwebtoken";

describe("UC-09: Socket Authentication Middleware", () => {
  test("TC-CHAT-AUTH-02 - Từ chối kết nối Socket do thiếu JWT Token", () => {
    const socket = {
      handshake: { auth: {} },
      data: {},
    };

    const next = jest.fn();
    socketAuthMiddleware(socket, next);

    expect(next).toHaveBeenCalledWith(expect.any(Error));
    expect(next.mock.calls[0][0].message).toBe("unauthorized: missing token");
    expect(socket.data.user).toBeUndefined();
  });

  test("TC-CHAT-AUTH-03 - Từ chối kết nối Socket do JWT Token bị hết hạn", () => {
    const socket = {
      handshake: { auth: { token: "expired.jwt.token" } },
      data: {},
    };

    jest.spyOn(jwt, "verify").mockImplementation(() => {
      const err = new Error("jwt expired");
      err.name = "TokenExpiredError";
      throw err;
    });

    const next = jest.fn();
    socketAuthMiddleware(socket, next);

    expect(next).toHaveBeenCalledWith(expect.any(Error));
    expect(next.mock.calls[0][0].message).toBe("unauthorized: token expired");

    jwt.verify.mockRestore();
  });

  test("TC-CHAT-AUTH-01 - Xác thực JWT Token hợp lệ khi kết nối Socket (Happy Case)", () => {
    const mockClaims = {
      accountId: "user-uuid-1",
      sub: "patient@example.com",
      role: "PATIENT",
    };

    jest.spyOn(jwt, "verify").mockReturnValue(mockClaims);

    const socket = {
      handshake: { auth: { token: "valid.jwt.token" } },
      data: {},
    };

    const next = jest.fn();
    socketAuthMiddleware(socket, next);

    expect(next).toHaveBeenCalledWith();
    expect(socket.data.user).toEqual({
      id: "user-uuid-1",
      email: "patient@example.com",
      role: "PATIENT",
    });

    jwt.verify.mockRestore();
  });

  test("TC-CHAT-AUTH-04 - Từ chối khi token thiếu accountId hoặc role", () => {
    const mockClaims = {
      sub: "user@example.com",
      // thiếu accountId và role
    };

    jest.spyOn(jwt, "verify").mockReturnValue(mockClaims);

    const socket = {
      handshake: { auth: { token: "invalid.claims.token" } },
      data: {},
    };

    const next = jest.fn();
    socketAuthMiddleware(socket, next);

    expect(next).toHaveBeenCalledWith(expect.any(Error));
    expect(next.mock.calls[0][0].message).toBe("unauthorized: token missing accountId/role");

    jwt.verify.mockRestore();
  });
});
