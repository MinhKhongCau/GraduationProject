"use client";

import { useQueryClient } from "@tanstack/react-query";
import { WeekTemplateGrid, type WeekTemplateState } from "./component/WeekTemplateGrid";
import { GenerateSlotsPanel } from "./component/GenerateSlotsPanel";
import { SlotCalendarPreview } from "./component/SlotCalendarPreview";
import { PageHeader, Spinner } from "@/components/ui";
import { useApiQuery, useApiMutation } from "@/hooks";
import { bookingApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useErrorContext } from "@/context/ErrorContext";
import type { DayOfWeek } from "@/types";

const DAY_VALUES: DayOfWeek[] = [1, 2, 3, 4, 5, 6, 7];

export default function ExpertSchedulePage() {
  const queryClient = useQueryClient();
  const { showSuccess } = useErrorContext();

  const { data: templates, isLoading: isLoadingTemplates } = useApiQuery({
    queryKey: QUERY_KEYS.shiftTemplates(),
    queryFn: () => bookingApi.getShiftTemplates(),
  });

  const { data: availabilities, isLoading: isLoadingAvailabilities } = useApiQuery({
    queryKey: QUERY_KEYS.myAvailabilities(),
    queryFn: () => bookingApi.getMyAvailabilities(),
  });

  const saveMutation = useApiMutation({
    mutationFn: async (state: WeekTemplateState) => {
      await Promise.all(
        DAY_VALUES.map((day) => {
          const config = state[day];
          const existing = availabilities?.find((availability) => availability.dayOfWeek === day);

          if (!config.enabled) {
            if (existing?.isEnabled) {
              return bookingApi.updateAvailability(existing.availabilityId, { isEnabled: false });
            }
            return Promise.resolve();
          }

          const effectiveFrom = config.effectiveFrom.getTime();
          const effectiveUntil = config.effectiveUntil ? config.effectiveUntil.getTime() : null;

          if (!existing) {
            return bookingApi.createAvailability({
              templateId: config.templateId,
              dayOfWeek: day,
              effectiveFrom,
              effectiveUntil,
              price: config.price,
            });
          }

          const changed =
            !existing.isEnabled ||
            existing.templateId !== config.templateId ||
            existing.effectiveFrom !== effectiveFrom ||
            (existing.effectiveUntil ?? null) !== effectiveUntil ||
            existing.price !== config.price;

          if (!changed) return Promise.resolve();

          return bookingApi.updateAvailability(existing.availabilityId, {
            isEnabled: true,
            templateId: config.templateId,
            effectiveFrom,
            effectiveUntil,
            price: config.price,
          });
        })
      );
    },
    onSuccess: () => {
      showSuccess("Weekly template saved.");
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.myAvailabilities() });
    },
  });

  const isLoading = isLoadingTemplates || isLoadingAvailabilities;

  return (
    <div className="mx-auto max-w-5xl space-y-6">
      <PageHeader
        className="mb-0"
        title="Weekly Schedule Template"
        description="Build your recurring weekly availability once, then generate bookable slots for patients from it."
      />

      {isLoading ? (
        <Spinner className="h-6 w-6" />
      ) : (
        <>
          <WeekTemplateGrid
            templates={templates ?? []}
            availabilities={availabilities ?? []}
            onSave={(state) => saveMutation.mutate(state)}
            isSubmitting={saveMutation.isPending}
          />
          <GenerateSlotsPanel />
          <SlotCalendarPreview />
        </>
      )}
    </div>
  );
}
