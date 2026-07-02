import {
  CloudRain,
  Ghost,
  HeartCrack,
  Users,
  Leaf,
  BedDouble,
  Smile,
  BrainCircuit,
  Briefcase,
  MessageSquareQuote,
  type LucideIcon,
} from "lucide-react";

export interface BookingTopic {
  id: string;
  labelKey: string;
  label: string;
  icon: LucideIcon;
}

/** Migrated from the original app/(patient)/book-appointment prototype. */
export const BOOKING_TOPICS: BookingTopic[] = [
  { id: "stress", labelKey: "patient.topics.stress", label: "Stress Management", icon: CloudRain },
  { id: "anxiety", labelKey: "patient.topics.anxiety", label: "Anxiety & Panic", icon: Ghost },
  { id: "depression", labelKey: "patient.topics.depression", label: "Depression", icon: HeartCrack },
  { id: "relationships", labelKey: "patient.topics.relationships", label: "Relationship Issues", icon: Users },
  { id: "grief", labelKey: "patient.topics.grief", label: "Grief & Loss", icon: Leaf },
  { id: "sleep", labelKey: "patient.topics.sleep", label: "Sleep Problems", icon: BedDouble },
  { id: "self-esteem", labelKey: "patient.topics.selfEsteem", label: "Self-Esteem", icon: Smile },
  { id: "trauma", labelKey: "patient.topics.trauma", label: "Trauma & PTSD", icon: BrainCircuit },
  { id: "work-life", labelKey: "patient.topics.workLife", label: "Work-Life Balance", icon: Briefcase },
  { id: "communication", labelKey: "patient.topics.communication", label: "Communication Skills", icon: MessageSquareQuote },
];
