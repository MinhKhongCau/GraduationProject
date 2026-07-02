"use client";

import { useState } from "react";
import { useSearchParams } from "next/navigation";
import { TopicStep } from "./component/TopicStep";
import { ExpertStep } from "./component/ExpertStep";
import { SlotStep } from "./component/SlotStep";
import { ReviewStep } from "./component/ReviewStep";
import { SuccessStep } from "./component/SuccessStep";
import { useApiQuery, useApiMutation } from "@/hooks";
import { bookingApi, expertApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import type { ExpertProfile, ExpertSlot, Appointment } from "@/types";

type WizardStep = "topic" | "expert" | "slot" | "review" | "success";

export default function BookAppointmentPage() {
  const searchParams = useSearchParams();
  const preselectedExpertId = searchParams.get("expertId");

  const [step, setStep] = useState<WizardStep>("topic");
  const [selectedTopics, setSelectedTopics] = useState<string[]>([]);
  const [selectedExpert, setSelectedExpert] = useState<ExpertProfile | null>(null);
  const [selectedSlot, setSelectedSlot] = useState<ExpertSlot | null>(null);
  const [completedAppointment, setCompletedAppointment] = useState<Appointment | null>(null);
  const [confirmedQueueNumber, setConfirmedQueueNumber] = useState<number | null>(null);

  useApiQuery({
    queryKey: QUERY_KEYS.expertProfile(preselectedExpertId ?? ""),
    queryFn: () => expertApi.getExpertProfile(preselectedExpertId!),
    enabled: !!preselectedExpertId && !selectedExpert,
    onSuccess: setSelectedExpert,
  });

  const confirmMutation = useApiMutation({
    mutationFn: async () => {
      if (!selectedExpert || !selectedSlot) throw new Error("Missing expert or slot selection");
      // Capture the queue number before insertion — getBookingQueueNumber
      // counts existing bookings for this slot, so calling it again after
      // confirmBookingMock adds this one would be off by one.
      const queueNumber = bookingApi.getBookingQueueNumber(selectedSlot.slotId);
      const appointment = bookingApi.confirmBookingMock({
        slotId: selectedSlot.slotId,
        expertId: selectedExpert.accountId,
        expertName: selectedExpert.fullName,
        topic: selectedTopics.join(", "),
      });
      return { appointment, queueNumber };
    },
    onSuccess: ({ appointment, queueNumber }) => {
      setCompletedAppointment(appointment);
      setConfirmedQueueNumber(queueNumber);
      setStep("success");
    },
  });

  function toggleTopic(topicId: string) {
    setSelectedTopics((current) =>
      current.includes(topicId) ? current.filter((id) => id !== topicId) : [...current, topicId]
    );
  }

  function goToExpertOrSlot() {
    setStep(selectedExpert ? "slot" : "expert");
  }

  return (
    <div className="mx-auto flex h-full max-w-5xl flex-col">
      {step === "topic" && (
        <TopicStep selectedTopics={selectedTopics} onToggleTopic={toggleTopic} onNext={goToExpertOrSlot} />
      )}
      {step === "expert" && (
        <ExpertStep
          selectedExpert={selectedExpert}
          onSelectExpert={setSelectedExpert}
          onBack={() => setStep("topic")}
          onNext={() => setStep("slot")}
        />
      )}
      {step === "slot" && selectedExpert && (
        <SlotStep
          expertId={selectedExpert.accountId}
          selectedSlot={selectedSlot}
          onSelectSlot={setSelectedSlot}
          onBack={() => setStep(preselectedExpertId ? "topic" : "expert")}
          onNext={() => setStep("review")}
        />
      )}
      {step === "review" && selectedExpert && selectedSlot && (
        <ReviewStep
          expert={selectedExpert}
          slot={selectedSlot}
          topics={selectedTopics}
          onBack={() => setStep("slot")}
          onConfirm={() => confirmMutation.mutate()}
          isSubmitting={confirmMutation.isPending}
        />
      )}
      {step === "success" && completedAppointment && confirmedQueueNumber !== null && (
        <SuccessStep queueNumber={confirmedQueueNumber} />
      )}
    </div>
  );
}
