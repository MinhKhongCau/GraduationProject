"use client";

import { WeeklyScheduleEditor } from "./component/WeeklyScheduleEditor";
import { Spinner } from "@/components/ui";
import { useApiQuery, useApiMutation } from "@/hooks";
import { expertApi } from "@/api";
import { useErrorContext } from "@/context/ErrorContext";
import type { WeeklySlotInput } from "@/types";

export default function ExpertSchedulePage() {
  const { showSuccess } = useErrorContext();

  const { data: weeklySlots, isLoading } = useApiQuery({
    queryKey: ["expert", "weekly-schedule"],
    queryFn: () => expertApi.getWeeklySchedule(),
  });

  const saveMutation = useApiMutation({
    mutationFn: (weeklySlots: WeeklySlotInput[]) => expertApi.updateWeeklySchedule({ weeklySlots }),
    onSuccess: () => showSuccess("Weekly schedule saved."),
  });

  return (
    <div className="mx-auto max-w-3xl">
      <h1 className="mb-2 text-2xl font-bold text-foreground">Weekly Schedule</h1>
      <p className="mb-6 text-sm text-muted-foreground">
        Set your recurring weekly availability. Patients can book sessions during these hours.
      </p>

      {isLoading ? (
        <Spinner className="h-6 w-6" />
      ) : (
        <WeeklyScheduleEditor
          initialSlots={weeklySlots ?? []}
          onSave={(slots) => saveMutation.mutate(slots)}
          isSubmitting={saveMutation.isPending}
        />
      )}
    </div>
  );
}
