"use client";

import { useState } from "react";
import { useSearchParams } from "next/navigation";
import { ScheduleStep } from "./component/ScheduleStep";
import { ConfirmStep } from "./component/ConfirmStep";
import { PaymentStep } from "./component/PaymentStep";
import { useApiQuery, useApiMutation, usePaymentRedirect } from "@/hooks";
import { bookingApi, expertApi, paymentApi } from "@/api";
import { normalizeError } from "@/api/http/errorNormalizer";
import { QUERY_KEYS, CHANNELING_FEE } from "@/constants";
import { useErrorContext } from "@/context/ErrorContext";
import type {
  ExpertProfile,
  AvailableTimeSlot,
  CreateAppointmentResponse,
  PatientRecord,
  Specialization,
} from "@/types";

type WizardStep = "schedule" | "confirm" | "payment";

export default function BookAppointmentPage() {
  const searchParams = useSearchParams();
  const preselectedExpertId = searchParams.get("expertId");

  const [step, setStep] = useState<WizardStep>("schedule");
  const [selectedSpecialization, setSelectedSpecialization] = useState<Specialization | null>(null);
  const [selectedExpert, setSelectedExpert] = useState<ExpertProfile | null>(null);
  const [selectedDate, setSelectedDate] = useState<string | null>(null);
  const [selectedSlot, setSelectedSlot] = useState<AvailableTimeSlot | null>(null);
  const [selectedRecord, setSelectedRecord] = useState<PatientRecord | null>(null);
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

  /** The slot hold was lost (expired or taken): drop the slot and let the patient pick again. */
  function backToSlotSelection() {
    setSelectedSlot(null);
    setStep("schedule");
  }

  const lockMutation = useApiMutation({
    mutationFn: () => {
      if (!selectedSlot) throw new Error("Missing slot selection");
      return bookingApi.lockSlot(selectedSlot.slotId);
    },
    onSuccess: () => setStep("confirm"),
    onError: (error) => {
      if (normalizeError(error).statusCode === 409) backToSlotSelection();
    },
  });

  const bookMutation = useApiMutation({
    mutationFn: () => {
      if (!selectedExpert || !selectedSlot || !selectedRecord) throw new Error("Missing booking selection");
      return bookingApi.createAppointment({
        slotId: selectedSlot.slotId,
        expertId: selectedExpert.accountId,
        patientRecordId: selectedRecord.recordId,
        specializationId: selectedSpecialization?.specId,
      });
    },
    onSuccess: (appointment) => {
      setCreatedAppointment(appointment);
      setStep("payment");
    },
    onError: (error) => {
      if (normalizeError(error).statusCode === 409) backToSlotSelection();
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

  function selectSpecialization(specialization: Specialization | null) {
    const expertStillMatches =
      !specialization || selectedExpert?.specializations.some((spec) => spec.specId === specialization.specId);
    if (!expertStillMatches) {
      // Experts are filtered by specialization, so a previous pick may no longer apply.
      setSelectedExpert(null);
      setSelectedDate(null);
      setSelectedSlot(null);
    }
    setSelectedSpecialization(specialization);
  }

  function selectExpert(expert: ExpertProfile) {
    if (expert.accountId !== selectedExpert?.accountId) {
      setSelectedDate(null);
      setSelectedSlot(null);
    }
    setSelectedExpert(expert);
  }

  function selectDate(date: string) {
    setSelectedDate(date);
    setSelectedSlot(null);
  }

  return (
    <div className="mx-auto flex h-full max-w-5xl flex-col">
      {step === "schedule" && (
        <ScheduleStep
          isExpertPreselected={!!preselectedExpertId}
          selectedSpecialization={selectedSpecialization}
          onSelectSpecialization={selectSpecialization}
          selectedExpert={selectedExpert}
          onSelectExpert={selectExpert}
          selectedDate={selectedDate}
          onSelectDate={selectDate}
          selectedSlot={selectedSlot}
          onSelectSlot={setSelectedSlot}
          selectedRecord={selectedRecord}
          onSelectRecord={setSelectedRecord}
          onNext={() => lockMutation.mutate()}
          isSubmitting={lockMutation.isPending}
        />
      )}
      {step === "confirm" && selectedExpert && selectedSlot && selectedRecord && (
        <ConfirmStep
          params={{
            slotId: selectedSlot.slotId,
            expertId: selectedExpert.accountId,
            patientRecordId: selectedRecord.recordId,
            specializationId: selectedSpecialization?.specId,
          }}
          onBack={() => setStep("schedule")}
          onPickAnotherSlot={backToSlotSelection}
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
