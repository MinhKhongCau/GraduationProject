import { EVENTS } from "../events.js";
import { ENV } from "../../../../config/env.js";
import { MeetError, meetRoom, presenceSummary, resolveParticipant } from "../../../application/services/meeting.service.js";
import { BookingClientError, createBookingClient } from "../../clients/booking.client.js";
import * as meetingStore from "../../persistence/stores/meeting.store.js";

const SIGNAL_TYPES = new Set(["offer", "answer", "ice"]);

let defaultBookingClient = null;

function defaultDeps() {
  if (!defaultBookingClient) defaultBookingClient = createBookingClient();
  return {
    booking: defaultBookingClient,
    store: meetingStore,
    now: Date.now,
    window: {
      earlyMs: ENV.MEET_JOIN_EARLY_MINUTES * 60_000,
      lateMs: ENV.MEET_JOIN_LATE_MINUTES * 60_000,
    },
  };
}

function toReply(err) {
  if (err instanceof MeetError) return { ok: false, code: err.code, message: err.message };
  if (err instanceof BookingClientError) {
    if (err.status === 404) return { ok: false, code: "not_found", message: "Meeting not found" };
    if (err.status === 409) return { ok: false, code: "not_completable", message: err.message };
    if (err.status === 502) {
      // booking recorded the session; only the payout failed — meet:end can be retried.
      return { ok: false, code: "payout_pending", retryable: true, message: err.message };
    }
  }
  console.error("meet error:", err);
  return { ok: false, code: "unavailable", message: "Meeting service is unavailable, try again" };
}

// Meet room for a paid appointment, joined through the link booking-service
// generates on payment (…/meet/<meetingToken>):
//   meet:join     { meetingToken }        → ack { ok, appointmentId, role, presence }
//   meet:signal   { type, payload }       → relayed to the other participant (WebRTC)
//   meet:presence                         → ack { ok, presence }
//   meet:end      (expert only)           → ack { ok, settled } once presence ≥ minimum
//   meet:leave
// Server → client: meet:participant-joined, meet:participant-left,
// meet:can-complete (expert reached the minimum presence), meet:ended.
export function meetSocketController(io, socket, deps = defaultDeps()) {
  const user = () => socket.data.user;

  function clearCompletionTimer() {
    if (socket.data.meetTimer) clearTimeout(socket.data.meetTimer);
    socket.data.meetTimer = null;
  }

  async function currentPresence(meet) {
    const presenceMs = await deps.store.getExpertPresenceMs(meet.appointmentId, deps.now(), meet.startTime);
    return presenceSummary(presenceMs, meet.minimumPresenceSeconds);
  }

  async function scheduleCompletionNotice(meet) {
    clearCompletionTimer();
    const presence = await currentPresence(meet);
    // Presence only accrues from the slot start, so wait for that too.
    const untilStart = Math.max(0, meet.startTime - deps.now());
    const delayMs = untilStart + presence.remainingSeconds * 1000;
    const timer = setTimeout(async () => {
      socket.data.meetTimer = null;
      if (socket.data.meet !== meet) return;
      const latest = await currentPresence(meet);
      if (latest.canComplete) socket.emit(EVENTS.MEET_CAN_COMPLETE, { appointmentId: meet.appointmentId, presence: latest });
      else scheduleCompletionNotice(meet);
    }, delayMs);
    timer.unref?.();
    socket.data.meetTimer = timer;
  }

  async function leave() {
    const meet = socket.data.meet;
    if (!meet) return;
    socket.data.meet = null;
    clearCompletionTimer();
    const room = meetRoom(meet.appointmentId);
    await socket.leave(room);
    if (meet.role === "EXPERT") {
      await deps.store.expertLeft(meet.appointmentId, socket.id, deps.now(), meet.startTime);
    }
    io.to(room).emit(EVENTS.MEET_PARTICIPANT_LEFT, { userId: user()?.id, role: meet.role });
  }

  socket.on(EVENTS.MEET_JOIN, async (payload, ack) => {
    const reply = typeof ack === "function" ? ack : () => {};
    try {
      const meetingToken = String(payload?.meetingToken || "").trim();
      if (!meetingToken) throw new MeetError("bad_request", "meetingToken is required");

      const meeting = await deps.booking.getMeeting(meetingToken);
      if (meeting && (await deps.store.isEnded(meeting.appointment_id))) {
        throw new MeetError("ended", "This session has already ended");
      }
      const role = resolveParticipant(meeting, user(), deps.now(), deps.window);

      await leave(); // one meeting per socket
      const meet = {
        appointmentId: meeting.appointment_id,
        role,
        startTime: meeting.start_time,
        endTime: meeting.end_time,
        minimumPresenceSeconds: meeting.minimum_presence_seconds,
      };
      socket.data.meet = meet;
      const room = meetRoom(meet.appointmentId);
      await socket.join(room);
      if (role === "EXPERT") {
        await deps.store.expertJoined(meet.appointmentId, socket.id, deps.now(), meet.startTime);
        await scheduleCompletionNotice(meet);
      }
      socket.to(room).emit(EVENTS.MEET_PARTICIPANT_JOINED, { userId: user().id, role });

      reply({
        ok: true,
        appointmentId: meet.appointmentId,
        role,
        startTime: meet.startTime,
        endTime: meet.endTime,
        presence: await currentPresence(meet),
      });
    } catch (err) {
      reply(toReply(err));
    }
  });

  socket.on(EVENTS.MEET_SIGNAL, (payload) => {
    const meet = socket.data.meet;
    if (!meet || !SIGNAL_TYPES.has(payload?.type)) return;
    socket.to(meetRoom(meet.appointmentId)).emit(EVENTS.MEET_SIGNAL, {
      from: user().id,
      role: meet.role,
      type: payload.type,
      payload: payload.payload,
    });
  });

  socket.on(EVENTS.MEET_PRESENCE, async (_payload, ack) => {
    const reply = typeof ack === "function" ? ack : () => {};
    const meet = socket.data.meet;
    if (!meet) return reply({ ok: false, code: "not_joined", message: "Join the meeting first" });
    try {
      reply({ ok: true, presence: await currentPresence(meet) });
    } catch (err) {
      reply(toReply(err));
    }
  });

  socket.on(EVENTS.MEET_END, async (_payload, ack) => {
    const reply = typeof ack === "function" ? ack : () => {};
    const meet = socket.data.meet;
    if (!meet) return reply({ ok: false, code: "not_joined", message: "Join the meeting first" });
    if (meet.role !== "EXPERT") return reply({ ok: false, code: "forbidden", message: "Only the expert can end the session" });
    try {
      const presence = await currentPresence(meet);
      if (!presence.canComplete) {
        return reply({ ok: false, code: "too_short", message: "The expert has not been in the meeting long enough", presence });
      }
      let result;
      try {
        result = await deps.booking.completeSession(meet.appointmentId, user().id, presence.presenceSeconds);
      } catch (err) {
        if (!(err instanceof BookingClientError && err.status === 502)) throw err;
        // Session is recorded as completed; tell the room, the payout is retried by meet:end.
        io.to(meetRoom(meet.appointmentId)).emit(EVENTS.MEET_ENDED, { appointmentId: meet.appointmentId, endedBy: user().id });
        return reply(toReply(err));
      }
      await deps.store.markEnded(meet.appointmentId);
      io.to(meetRoom(meet.appointmentId)).emit(EVENTS.MEET_ENDED, { appointmentId: meet.appointmentId, endedBy: user().id });
      reply({ ok: true, appointmentId: meet.appointmentId, settled: Boolean(result?.settled), presence });
    } catch (err) {
      reply(toReply(err));
    }
  });

  socket.on(EVENTS.MEET_LEAVE, () => {
    leave().catch((err) => console.error("meet leave failed:", err));
  });

  socket.on("disconnect", () => {
    leave().catch((err) => console.error("meet leave on disconnect failed:", err));
  });
}
