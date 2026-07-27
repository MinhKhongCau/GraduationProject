"use client";

import { CalendarCheck } from "lucide-react";
import { Button, Card } from "@/components/ui";
import { CHANNELING_FEE, BOOKING_TOPICS } from "@/constants";
import type { ExpertProfile, AvailableTimeSlot } from "@/types";

export interface ReviewStepProps {
  expert: ExpertProfile;
  slot: AvailableTimeSlot;
  date: string;
  topics: string[];
  onBack: () => void;
  onConfirm: () => void;
  isSubmitting: boolean;
}

export function ReviewStep({ expert, slot, date, topics, onBack, onConfirm, isSubmitting }: ReviewStepProps) {
  const topicLabels = BOOKING_TOPICS.filter((topic) => topics.includes(topic.id))
    .map((topic) => topic.label)
    .join(", ");

  const timeLabel = new Date(slot.startTime).toLocaleTimeString("en-GB", { hour: "2-digit", minute: "2-digit" });
  const dateLabel = new Date(date).toLocaleDateString("en-GB", { day: "2-digit", month: "short", year: "numeric" });

  return (
    <div className="flex h-full flex-col">
      <h1 className="mb-6 text-2xl font-bold text-foreground">Review your booking</h1>

      <Card className="mb-8 max-w-xl space-y-4 p-6">
        <div className="flex items-center justify-between">
          <span className="text-sm text-muted-foreground">Expert</span>
          <span className="text-sm font-bold text-foreground">{expert.fullName}</span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-sm text-muted-foreground">Topics</span>
          <span className="text-sm font-bold text-foreground">{topicLabels}</span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-sm text-muted-foreground">Date &amp; time</span>
          <span className="flex items-center gap-1.5 text-sm font-bold text-foreground">
            <CalendarCheck className="h-4 w-4 text-primary" />
            {dateLabel} · {timeLabel}
          </span>
        </div>
        <div className="flex items-center justify-between border-t border-border pt-4">
          <span className="text-sm text-muted-foreground">Channeling fee</span>
          <span className="text-lg font-extrabold text-foreground">
            {(slot.price ?? CHANNELING_FEE).toLocaleString("vi-VN")} đ
          </span>
        </div>
      </Card>

      <p className="mb-6 max-w-xl text-xs text-muted-foreground">
        Confirming holds this slot for 15 minutes and takes you straight to VNPay to complete payment.
      </p>

      <div className="mt-auto flex justify-between border-t border-border pt-6">
        <Button variant="outline" onClick={onBack} disabled={isSubmitting}>
          Back
        </Button>
        <Button onClick={onConfirm} disabled={isSubmitting}>
          {isSubmitting ? "Booking..." : "Confirm booking"}
        </Button>
      </div>
    </div>
  );
}
