// Meet-room rules between a patient and an expert for a paid appointment.
// Booking-service is the source of truth for the appointment; this module
// only decides who may enter and when the expert may end the session.

export class MeetError extends Error {
  constructor(code, message) {
    super(message);
    this.code = code;
  }
}

export const meetRoom = (appointmentId) => `meet:${appointmentId}`;

// Returns "PATIENT" | "EXPERT" for the account allowed into the meeting.
export function resolveParticipant(meeting, user, nowMs, { earlyMs, lateMs }) {
  if (!meeting) throw new MeetError("not_found", "Meeting not found");

  let role = null;
  if (user?.id && user.id === meeting.expert_id) role = "EXPERT";
  else if (user?.id && user.id === meeting.patient_id) role = "PATIENT";
  if (!role) throw new MeetError("forbidden", "You are not a participant of this meeting");

  if (meeting.session_completed_at) throw new MeetError("ended", "This session has already ended");
  if (meeting.status !== "CONFIRMED" && meeting.status !== "COMPLETED") {
    throw new MeetError("not_paid", "The appointment is not paid or was cancelled");
  }
  if (nowMs < meeting.start_time - earlyMs) {
    throw new MeetError("too_early", "The meeting room is not open yet");
  }
  if (nowMs > meeting.end_time + lateMs) {
    throw new MeetError("expired", "The meeting room has closed");
  }
  return role;
}

export function presenceSummary(presenceMs, minimumPresenceSeconds) {
  const presenceSeconds = Math.floor(presenceMs / 1000);
  const requiredSeconds = Number(minimumPresenceSeconds || 0);
  return {
    presenceSeconds,
    requiredSeconds,
    remainingSeconds: Math.max(0, requiredSeconds - presenceSeconds),
    canComplete: presenceSeconds >= requiredSeconds,
  };
}
