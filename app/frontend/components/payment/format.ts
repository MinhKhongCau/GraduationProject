/** VND amounts are integers from payment-service. */
export function formatVnd(amount: number): string {
  return `${amount.toLocaleString("vi-VN")} đ`;
}

/** payment-service timestamps are epoch milliseconds. */
export function formatDateTime(ms?: number): string {
  if (!ms) return "—";
  return new Date(ms).toLocaleString("vi-VN", { dateStyle: "short", timeStyle: "short" });
}

function toDateInput(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

/**
 * History screens look backwards: the last 30 days up to today. (payment-service's own default
 * window starts today and runs forward, which suits order lookups but not history.)
 */
export function defaultDateRange(): { from: string; to: string } {
  const to = new Date();
  const from = new Date();
  from.setDate(to.getDate() - 29);
  return { from: toDateInput(from), to: toDateInput(to) };
}

export function shortId(id?: string): string {
  return id ? id.slice(0, 8) : "—";
}
