import { APPOINTMENT_STATUS_TEXT } from "../status";
import type { Appointment } from "@/types";

export function shortBookingId(appointmentId: string): string {
  return `#${appointmentId.slice(0, 8).toUpperCase()}`;
}

/** The other party of a booking: the expert for patients, the patient for experts. */
export interface Counterpart<T extends Appointment> {
  /** CSV column header, e.g. "Expert". */
  label: string;
  name: (appointment: T) => string | undefined;
}

function saveFile(filename: string, content: string, type: string) {
  const url = URL.createObjectURL(new Blob([content], { type }));
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();
  URL.revokeObjectURL(url);
}

function csvCell(value: unknown): string {
  const text = value == null ? "" : String(value);
  return /[",\n]/.test(text) ? `"${text.replace(/"/g, '""')}"` : text;
}

const isoOrEmpty = (time?: number | null) => (time ? new Date(time).toISOString() : "");

/** Spreadsheet of the given bookings (UTF-8 BOM so Excel keeps Vietnamese text). */
export function downloadBookingsCsv<T extends Appointment>(appointments: T[], counterpart: Counterpart<T>) {
  const header = [
    "Booking ID",
    counterpart.label,
    "Specialization",
    "Start",
    "End",
    "Booked at",
    "Price (VND)",
    "Amount paid (VND)",
    "Payment status",
    "Status",
    "Meeting link",
  ];
  const rows = appointments.map((a) => [
    a.appointmentId,
    counterpart.name(a) ?? "",
    a.specializationName ?? "",
    isoOrEmpty(a.startTime),
    isoOrEmpty(a.endTime),
    isoOrEmpty(a.createdAt),
    a.price ?? "",
    a.payment?.amountVnd ?? "",
    a.payment?.status ?? "",
    APPOINTMENT_STATUS_TEXT[a.statusLabel] ?? a.statusLabel,
    a.meetingLink ?? "",
  ]);
  const csv = [header, ...rows].map((row) => row.map(csvCell).join(",")).join("\r\n");
  saveFile("bookings.csv", `﻿${csv}`, "text/csv;charset=utf-8");
}

const icsTime = (time: number) =>
  new Date(time)
    .toISOString()
    .replace(/[-:]/g, "")
    .replace(/\.\d{3}/, "");
const icsText = (text: string) => text.replace(/[\;,]/g, (c) => `\\${c}`).replace(/\n/g, "\\n");

/** Calendar "tickets" (.ics) for confirmed sessions, importable into any calendar app. */
export function downloadTicketsIcs<T extends Appointment>(appointments: T[], counterpart: Counterpart<T>): number {
  const tickets = appointments.filter((a) => a.statusLabel === "CONFIRMED" && a.startTime);
  if (tickets.length === 0) return 0;

  const events = tickets.map((a) => {
    const start = a.startTime as number;
    const end = a.endTime ?? start + 60 * 60 * 1000;
    const who = counterpart.name(a) ?? counterpart.label.toLowerCase();
    const summary = `Session with ${who}${a.specializationName ? ` (${a.specializationName})` : ""}`;
    return [
      "BEGIN:VEVENT",
      `UID:${a.appointmentId}`,
      `DTSTAMP:${icsTime(Date.now())}`,
      `DTSTART:${icsTime(start)}`,
      `DTEND:${icsTime(end)}`,
      `SUMMARY:${icsText(summary)}`,
      `DESCRIPTION:${icsText(`Booking ${shortBookingId(a.appointmentId)}`)}`,
      ...(a.meetingLink ? [`URL:${a.meetingLink}`, `LOCATION:${icsText(a.meetingLink)}`] : []),
      "END:VEVENT",
    ].join("\r\n");
  });
  const calendar = ["BEGIN:VCALENDAR", "VERSION:2.0", "PRODID:-//Bookings//EN", ...events, "END:VCALENDAR"].join(
    "\r\n"
  );
  saveFile("booking-tickets.ics", calendar, "text/calendar;charset=utf-8");
  return tickets.length;
}
