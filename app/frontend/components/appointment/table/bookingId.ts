/** "#1A2B3C4D": the short booking reference shown in tables and menus. */
export function shortBookingId(appointmentId: string): string {
  return `#${appointmentId.slice(0, 8).toUpperCase()}`;
}
