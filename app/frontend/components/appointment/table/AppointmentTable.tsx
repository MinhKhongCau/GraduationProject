"use client";

import type { ReactNode } from "react";
import clsx from "clsx";
import { Pagination, Spinner } from "@/components/ui";
import { OPTIONAL_COLUMNS, type AppointmentListState } from "./useAppointmentListState";

export interface AppointmentTableProps {
  state: AppointmentListState;
  /** Header of the other-party column ("Expert" / "Patient"). */
  counterpartLabel: string;
  visibleIds: string[];
  totalPages: number;
  isLoading: boolean;
  isFetching: boolean;
  emptyText: string;
  /** One row per appointment. */
  children: ReactNode;
  /** Optional card list shown instead of the table on phones (below `md`). */
  mobileCards?: ReactNode;
}

/** Bordered table frame: header, loading/empty rows, rows, then pagination. */
export function AppointmentTable({
  state,
  counterpartLabel,
  visibleIds,
  totalPages,
  isLoading,
  isFetching,
  emptyText,
  children,
  mobileCards,
}: AppointmentTableProps) {
  const shownColumns = OPTIONAL_COLUMNS.filter((column) => state.columns[column.id]);
  // Booking ID + counterpart + Actions, plus the optional columns shown.
  const columnCount = 3 + shownColumns.length;

  const hasCards = mobileCards !== undefined;

  return (
    <div className={clsx("overflow-hidden", hasCards ? "md:rounded-xl md:border md:border-border md:bg-background" : "rounded-xl border border-border bg-background")}>
      {hasCards && (
        <div
          className={clsx("space-y-3 transition-opacity md:hidden", isFetching && !isLoading && "opacity-60")}
          aria-busy={isFetching}
        >
          {isLoading ? (
            <div className="rounded-xl border border-border bg-background px-3 py-10">
              <Spinner className="mx-auto h-6 w-6" />
            </div>
          ) : visibleIds.length === 0 ? (
            <p className="rounded-xl border border-dashed border-border bg-background px-4 py-10 text-center text-sm text-muted-foreground">
              {emptyText}
            </p>
          ) : (
            mobileCards
          )}
        </div>
      )}
      <div
        className={clsx("overflow-x-auto transition-opacity", hasCards && "hidden md:block", isFetching && !isLoading && "opacity-60")}
        aria-busy={isFetching}
      >
        <table className="w-full border-collapse text-left">
          <thead className="border-b border-border bg-surface">
            <tr className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              <th scope="col" className="px-3 py-2.5">
                Booking ID
              </th>
              <th scope="col" className="px-3 py-2.5">
                {counterpartLabel}
              </th>
              {shownColumns.map((column) => (
                <th
                  key={column.id}
                  scope="col"
                  className={clsx("px-3 py-2.5", column.align === "right" && "text-right")}
                >
                  {column.label}
                </th>
              ))}
              <th scope="col" className="px-3 py-2.5 text-right">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {isLoading ? (
              <tr>
                <td colSpan={columnCount} className="px-3 py-10">
                  <Spinner className="mx-auto h-6 w-6" />
                </td>
              </tr>
            ) : visibleIds.length === 0 ? (
              <tr>
                <td colSpan={columnCount} className="px-3 py-10 text-center text-sm text-muted-foreground">
                  {emptyText}
                </td>
              </tr>
            ) : (
              children
            )}
          </tbody>
        </table>
      </div>
      <div
        className={clsx(
          hasCards &&
            "mt-3 overflow-hidden rounded-xl border border-border bg-background empty:hidden md:mt-0 md:rounded-none md:border-0 [&>div]:border-t-0 md:[&>div]:border-t"
        )}
      >
        <Pagination page={state.page + 1} totalPages={totalPages} onPageChange={(next) => state.goToPage(next - 1)} />
      </div>
    </div>
  );
}
