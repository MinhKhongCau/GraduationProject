import { ENV } from "../../../config/env.js";
import { createTokenManager } from "./internal-token.js";

export class BookingClientError extends Error {
  constructor(status, code, message) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

// booking-service internal API used by the meet room.
export function createBookingClient({
  baseUrl = ENV.BOOKING_SERVICE_INTERNAL_URL,
  tokens = createTokenManager(),
  fetchImpl = fetch,
} = {}) {
  async function call(method, path, body, retried = false) {
    const token = await tokens.getToken();
    const res = await fetchImpl(`${baseUrl}${path}`, {
      method,
      headers: {
        Authorization: `Bearer ${token}`,
        ...(body ? { "Content-Type": "application/json" } : {}),
      },
      body: body ? JSON.stringify(body) : undefined,
    });
    if (res.status === 401 && !retried) {
      tokens.invalidate();
      return call(method, path, body, true);
    }
    const payload = await res.json().catch(() => ({}));
    if (!res.ok) {
      throw new BookingClientError(res.status, payload.error || "booking_error", payload.message || `booking-service returned ${res.status}`);
    }
    return payload.data;
  }

  return {
    // → { appointment_id, patient_id, expert_id, status, start_time, end_time,
    //     session_completed_at, minimum_presence_seconds }
    getMeeting(meetingToken) {
      return call("GET", `/internal/meetings/${encodeURIComponent(meetingToken)}`);
    },
    // Marks the appointment COMPLETED and triggers the expert payout. Idempotent.
    completeSession(appointmentId, expertId, presenceSeconds) {
      return call("POST", `/internal/appointments/${encodeURIComponent(appointmentId)}/complete-session`, {
        expert_id: expertId,
        presence_seconds: presenceSeconds,
      });
    },
  };
}
