"use client";

import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { CalendarPlus } from "lucide-react";
import { Button, Card } from "@/components/ui";
import { useApiMutation } from "@/hooks";
import { bookingApi } from "@/api";
import { useErrorContext } from "@/context/ErrorContext";

/** Materializes the saved weekly template into concrete bookable slots for the next N days. */
export function GenerateSlotsPanel() {
  const [days, setDays] = useState(14);
  const { showSuccess } = useErrorContext();
  const queryClient = useQueryClient();

  const generateMutation = useApiMutation({
    mutationFn: () => bookingApi.generateSlots(days),
    onSuccess: (result) => {
      showSuccess(
        result.slotsCreated > 0
          ? `${result.slotsCreated} slots generated for the next ${days} days.`
          : "No new slots were generated — make sure your weekly template is saved first."
      );
      queryClient.invalidateQueries({ queryKey: ["booking", "expert-slots"] });
    },
  });

  return (
    <Card className="p-6">
      <h2 className="mb-1 text-lg font-bold text-foreground">Generate bookable slots</h2>
      <p className="mb-4 text-sm text-muted-foreground">
        Turns your saved weekly template into concrete slots patients can book, starting today.
      </p>
      <div className="flex flex-wrap items-center gap-3">
        <label className="flex items-center gap-2 text-sm font-semibold text-foreground">
          Days ahead
          <input
            type="number"
            min={1}
            max={30}
            value={days}
            onChange={(event) => setDays(Math.min(30, Math.max(1, Number(event.target.value) || 1)))}
            className="w-20 rounded-lg border border-border px-2 py-1.5 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
          />
        </label>
        <Button onClick={() => generateMutation.mutate()} disabled={generateMutation.isPending}>
          <CalendarPlus className="h-4 w-4" />
          {generateMutation.isPending ? "Generating..." : "Generate slots"}
        </Button>
      </div>
    </Card>
  );
}
