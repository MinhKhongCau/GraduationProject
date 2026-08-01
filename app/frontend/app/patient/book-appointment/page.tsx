"use client";

import { useState } from "react";
import { useSearchParams } from "next/navigation";
import { TopicStep } from "./component/TopicStep";
import { ExpertStep } from "./component/ExpertStep";
import { SlotStep } from "./component/SlotStep";
import { ReviewStep } from "./component/ReviewStep";
import { PaymentStep } from "./component/PaymentStep";
import { useApiQuery, useApiMutation, usePaymentRedirect } from "@/hooks";
import { bookingApi, expertApi, paymentApi } from "@/api";
import { QUERY_KEYS, CHANNELING_FEE } from "@/constants";
import { useErrorContext } from "@/context/ErrorContext";
import type { ExpertProfile, AvailableTimeSlot, CreateAppointmentResponse } from "@/types";

type WizardStep = "topic" | "expert" | "slot" | "review" | "payment";

export default function BookAppointmentPage() {
  const searchParams = useSearchParams();
  const preselectedExpertId = searchParams.get("expertId");

  const [step, setStep] = useState<WizardStep>("topic");
  const [selectedTopics, setSelectedTopics] = useState<string[]>([]);
  const [selectedExpert, setSelectedExpert] = useState<ExpertProfile | null>(null);
  const [selectedDate, setSelectedDate] = useState<string | null>(null);
  const [selectedSlot, setSelectedSlot] = useState<AvailableTimeSlot | null>(null);
  const [createdAppointment, setCreatedAppointment] = useState<CreateAppointmentResponse | null>(null);
  const { showError } = useErrorContext();
  const { isPaymentBrowserOpen, openPayment } = usePaymentRedirect({
    onBrowserFinished: () => {
      showError({
        message: "Payment browser closed. Your booking may still be pending payment; check My Bookings before trying again.",
      });
    },
  });

  useApiQuery({
    queryKey: QUERY_KEYS.expertProfile(preselectedExpertId ?? ""),
    queryFn: () => expertApi.getExpertProfile(preselectedExpertId!),
    enabled: !!preselectedExpertId && !selectedExpert,
    onSuccess: setSelectedExpert,
  });

  const bookMutation = useApiMutation({
    mutationFn: async () => {
      if (!selectedExpert || !selectedSlot) throw new Error("Missing expert or slot selection");
      await bookingApi.lockSlot(selectedSlot.slotId);
      return bookingApi.createAppointment({
        slotId: selectedSlot.slotId,
        expertId: selectedExpert.accountId,
      });
    },
    onSuccess: (appointment) => {
      setCreatedAppointment(appointment);
      setStep("payment");
    },
  });

  const payMutation = useApiMutation({
    mutationFn: () => {
      if (!selectedExpert || !createdAppointment) throw new Error("Missing appointment");
      return paymentApi.createOrder({
        expertId: selectedExpert.accountId,
        amount: selectedSlot?.price ?? CHANNELING_FEE,
        gateway: "VNPAY",
        appointmentId: createdAppointment.appointmentId,
      });
    },
    onSuccess: async (order) => {
      const opened = await openPayment(order.paymentUrl);
      if (!opened) showError({ message: "Unable to open the payment page. Please try again." });
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

  function selectDate(date: string) {
    setSelectedDate(date);
    setSelectedSlot(null);
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
          selectedDate={selectedDate}
          onSelectDate={selectDate}
          selectedSlot={selectedSlot}
          onSelectSlot={setSelectedSlot}
          onBack={() => setStep(preselectedExpertId ? "topic" : "expert")}
          onNext={() => setStep("review")}
        />
      )}
      {step === "review" && selectedExpert && selectedSlot && selectedDate && (
        <ReviewStep
          expert={selectedExpert}
          slot={selectedSlot}
          date={selectedDate}
          topics={selectedTopics}
          onBack={() => setStep("slot")}
          onConfirm={() => bookMutation.mutate()}
          isSubmitting={bookMutation.isPending}
        />
      )}
      {step === "payment" && (
        <PaymentStep
          onPay={() => payMutation.mutate()}
          isPending={payMutation.isPending || isPaymentBrowserOpen}
          isError={payMutation.isError}
        />
      )}
    </div>
  );
}
