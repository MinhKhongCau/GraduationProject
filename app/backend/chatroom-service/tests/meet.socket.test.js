import { jest } from "@jest/globals";
import { meetSocketController } from "../src/infrastructure/socket/controllers/meet.socket.controller.js";
import { BookingClientError } from "../src/infrastructure/clients/booking.client.js";

const MIN = 60_000;
const NOW = Date.UTC(2026, 9, 10, 9, 45);

function meeting(overrides = {}) {
  return {
    appointment_id: "appt-1",
    patient_id: "patient-1",
    expert_id: "expert-1",
    status: "CONFIRMED",
    start_time: NOW - 40 * MIN,
    end_time: NOW + 20 * MIN,
    session_completed_at: null,
    minimum_presence_seconds: 1800,
    ...overrides,
  };
}

function setup({ user, meetingData = meeting(), presenceMs = 0 } = {}) {
  const handlers = {};
  const roomEmit = jest.fn();
  const peerEmit = jest.fn();
  const socket = {
    id: `socket-${user.id}`,
    data: { user },
    on: jest.fn((event, handler) => {
      handlers[event] = handler;
    }),
    emit: jest.fn(),
    join: jest.fn().mockResolvedValue(undefined),
    leave: jest.fn().mockResolvedValue(undefined),
    to: jest.fn().mockReturnValue({ emit: peerEmit }),
  };
  const io = { to: jest.fn().mockReturnValue({ emit: roomEmit }) };
  const deps = {
    booking: {
      getMeeting: jest.fn().mockResolvedValue(meetingData),
      completeSession: jest.fn().mockResolvedValue({ settled: true }),
    },
    store: {
      expertJoined: jest.fn().mockResolvedValue(undefined),
      expertLeft: jest.fn().mockResolvedValue(undefined),
      getExpertPresenceMs: jest.fn().mockResolvedValue(presenceMs),
      markEnded: jest.fn().mockResolvedValue(undefined),
      isEnded: jest.fn().mockResolvedValue(false),
    },
    now: () => NOW,
  };
  meetSocketController(io, socket, deps);

  const call = (event, payload) =>
    new Promise((resolve) => {
      handlers[event](payload, resolve);
    });
  return { socket, io, deps, handlers, call, roomEmit, peerEmit };
}

const expert = { id: "expert-1", role: "EXPERT" };
const patient = { id: "patient-1", role: "PATIENT" };

afterEach(() => {
  jest.useRealTimers();
});

describe("Meet room (paid appointment link)", () => {
  test("patient joins the meeting room through the link token", async () => {
    const { call, socket, deps, peerEmit } = setup({ user: patient });

    const reply = await call("meet:join", { meetingToken: "tok-1" });

    expect(deps.booking.getMeeting).toHaveBeenCalledWith("tok-1");
    expect(reply).toMatchObject({ ok: true, appointmentId: "appt-1", role: "PATIENT" });
    expect(socket.join).toHaveBeenCalledWith("meet:appt-1");
    expect(deps.store.expertJoined).not.toHaveBeenCalled();
    expect(peerEmit).toHaveBeenCalledWith("meet:participant-joined", { socketId: "socket-patient-1", userId: "patient-1", role: "PATIENT" });
  });

  test("expert join starts presence tracking from the slot start", async () => {
    jest.useFakeTimers();
    const { call, deps, socket } = setup({ user: expert, presenceMs: 10 * MIN });

    const reply = await call("meet:join", { meetingToken: "tok-1" });

    expect(reply).toMatchObject({ ok: true, role: "EXPERT", presence: { presenceSeconds: 600, remainingSeconds: 1200, canComplete: false } });
    expect(deps.store.expertJoined).toHaveBeenCalledWith("appt-1", socket.id, NOW, NOW - 40 * MIN);
    expect(socket.data.meetTimer).toBeTruthy();
  });

  test("someone else's account cannot join", async () => {
    const { call, socket } = setup({ user: { id: "stranger", role: "PATIENT" } });
    const reply = await call("meet:join", { meetingToken: "tok-1" });
    expect(reply).toMatchObject({ ok: false, code: "forbidden" });
    expect(socket.join).not.toHaveBeenCalled();
  });

  test.each([
    ["unpaid appointment", meeting({ status: "PENDING_PAYMENT" }), "not_paid"],
    ["session already ended", meeting({ session_completed_at: NOW - MIN }), "ended"],
  ])("rejects join: %s", async (_name, meetingData, code) => {
    const { call } = setup({ user: patient, meetingData });
    expect(await call("meet:join", { meetingToken: "tok-1" })).toMatchObject({ ok: false, code });
  });

  test.each([
    ["days before the slot", meeting({ start_time: NOW + 3 * 24 * 60 * MIN, end_time: NOW + (3 * 24 * 60 + 60) * MIN })],
    ["days after the slot", meeting({ start_time: NOW - 3 * 24 * 60 * MIN, end_time: NOW - (3 * 24 * 60 - 60) * MIN })],
  ])("joins regardless of time window: %s", async (_name, meetingData) => {
    const { call, socket } = setup({ user: patient, meetingData });
    expect(await call("meet:join", { meetingToken: "tok-1" })).toMatchObject({ ok: true, role: "PATIENT" });
    expect(socket.join).toHaveBeenCalled();
  });

  test("expert joining weeks early arms a capped completion timer", async () => {
    jest.useFakeTimers();
    try {
      const start = NOW + 40 * 24 * 60 * MIN;
      const { call, socket } = setup({ user: expert, meetingData: meeting({ start_time: start, end_time: start + 60 * MIN }) });
      const spy = jest.spyOn(global, "setTimeout");
      await call("meet:join", { meetingToken: "tok-1" });
      expect(socket.data.meetTimer).toBeTruthy();
      expect(spy.mock.calls.at(-1)[1]).toBe(2 ** 31 - 1);
      spy.mockRestore();
    } finally {
      jest.useRealTimers();
    }
  });

  test("unknown link returns not_found", async () => {
    const { call, deps } = setup({ user: patient });
    deps.booking.getMeeting.mockRejectedValue(new BookingClientError(404, "not_found", "Meeting not found"));
    expect(await call("meet:join", { meetingToken: "nope" })).toMatchObject({ ok: false, code: "not_found" });
  });

  test("expert cannot end the session before 30 minutes of presence", async () => {
    const { call, deps } = setup({ user: expert, presenceMs: 29 * MIN });
    await call("meet:join", { meetingToken: "tok-1" });

    const reply = await call("meet:end");

    expect(reply).toMatchObject({ ok: false, code: "too_short", presence: { remainingSeconds: 60 } });
    expect(deps.booking.completeSession).not.toHaveBeenCalled();
  });

  test("expert ends the session after 30 minutes and booking pays out", async () => {
    const { call, deps, roomEmit } = setup({ user: expert, presenceMs: 31 * MIN });
    await call("meet:join", { meetingToken: "tok-1" });

    const reply = await call("meet:end");

    expect(deps.booking.completeSession).toHaveBeenCalledWith("appt-1", "expert-1", 1860);
    expect(deps.store.markEnded).toHaveBeenCalledWith("appt-1");
    expect(roomEmit).toHaveBeenCalledWith("meet:ended", { appointmentId: "appt-1", endedBy: "expert-1" });
    expect(reply).toMatchObject({ ok: true, settled: true });
  });

  test("payout failure is reported as retryable", async () => {
    const { call, deps } = setup({ user: expert, presenceMs: 31 * MIN });
    deps.booking.completeSession.mockRejectedValue(new BookingClientError(502, "bad_gateway", "payout failed"));
    await call("meet:join", { meetingToken: "tok-1" });

    expect(await call("meet:end")).toMatchObject({ ok: false, code: "payout_pending", retryable: true });
    expect(deps.store.markEnded).not.toHaveBeenCalled();
  });

  test("patient cannot end the session", async () => {
    const { call, deps } = setup({ user: patient, presenceMs: 60 * MIN });
    await call("meet:join", { meetingToken: "tok-1" });
    expect(await call("meet:end")).toMatchObject({ ok: false, code: "forbidden" });
    expect(deps.booking.completeSession).not.toHaveBeenCalled();
  });

  test("WebRTC signals are relayed only to the meeting room", async () => {
    const { call, handlers, socket, peerEmit } = setup({ user: patient });
    handlers["meet:signal"]({ type: "offer", payload: { sdp: "x" } });
    expect(peerEmit).not.toHaveBeenCalledWith("meet:signal", expect.anything());

    await call("meet:join", { meetingToken: "tok-1" });
    handlers["meet:signal"]({ type: "offer", payload: { sdp: "x" } });
    expect(socket.to).toHaveBeenLastCalledWith("meet:appt-1");
    expect(peerEmit).toHaveBeenLastCalledWith("meet:signal", {
      from: "patient-1",
      fromSocketId: "socket-patient-1",
      role: "PATIENT",
      type: "offer",
      payload: { sdp: "x" },
    });
  });

  test("WebRTC signals addressed to a peer go only to that socket in the room", async () => {
    const { call, handlers, io, socket, peerEmit } = setup({ user: patient });
    io.sockets = { adapter: { rooms: new Map([["meet:appt-1", new Set(["socket-patient-1", "socket-expert-1"])]]) } };
    await call("meet:join", { meetingToken: "tok-1" });
    peerEmit.mockClear();

    handlers["meet:signal"]({ to: "socket-outsider", type: "offer", payload: { sdp: "x" } });
    expect(peerEmit).not.toHaveBeenCalled();

    handlers["meet:signal"]({ to: "socket-expert-1", type: "answer", payload: { sdp: "y" } });
    expect(socket.to).toHaveBeenLastCalledWith("socket-expert-1");
    expect(peerEmit).toHaveBeenLastCalledWith("meet:signal", expect.objectContaining({ fromSocketId: "socket-patient-1", type: "answer" }));
  });

  test("expert disconnect stops presence tracking", async () => {
    const { call, handlers, deps, socket, roomEmit } = setup({ user: expert });
    await call("meet:join", { meetingToken: "tok-1" });

    await handlers["disconnect"]();
    await Promise.resolve();

    expect(deps.store.expertLeft).toHaveBeenCalledWith("appt-1", socket.id, NOW, NOW - 40 * MIN);
    expect(roomEmit).toHaveBeenCalledWith("meet:participant-left", { socketId: "socket-expert-1", userId: "expert-1", role: "EXPERT" });
    expect(socket.data.meet).toBeNull();
  });
});
