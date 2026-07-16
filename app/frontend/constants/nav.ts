import {
  LayoutDashboard,
  Search,
  CalendarPlus,
  ClipboardList,
  Wallet,
  ClipboardCheck,
  FileText,
  MessageSquare,
  Settings,
  Users,
  CalendarClock,
  CalendarCheck,
  Layers,
  MessageCircle,
  type LucideIcon,
} from "lucide-react";
import { ROUTES } from "./route";

export interface NavItem {
  id: string;
  labelKey: string;
  label: string;
  href: string;
  icon: LucideIcon;
}

export const PATIENT_NAV_ITEMS: NavItem[] = [
  { id: "dashboard", labelKey: "nav.dashboard", label: "Dashboard", href: ROUTES.PATIENT.DASHBOARD, icon: LayoutDashboard },
  { id: "find-experts", labelKey: "nav.findExperts", label: "Find Experts", href: ROUTES.PATIENT.FIND_EXPERTS, icon: Search },
  { id: "book-appointment", labelKey: "nav.bookAppointment", label: "Book Appointment", href: ROUTES.PATIENT.BOOK_APPOINTMENT, icon: CalendarPlus },
  { id: "my-bookings", labelKey: "nav.myBookings", label: "My Bookings", href: ROUTES.PATIENT.MY_BOOKINGS, icon: ClipboardList },
  { id: "wallet", labelKey: "nav.wallet", label: "Wallet", href: ROUTES.PATIENT.WALLET, icon: Wallet },
  { id: "assessment", labelKey: "nav.assessment", label: "Assessment", href: ROUTES.PATIENT.ASSESSMENT, icon: ClipboardCheck },
  { id: "medical-history", labelKey: "nav.medicalHistory", label: "Medical History", href: ROUTES.PATIENT.MEDICAL_HISTORY, icon: FileText },
  { id: "messages", labelKey: "nav.messages", label: "Messages", href: ROUTES.PATIENT.MESSAGES, icon: MessageSquare },
  { id: "forum", labelKey: "nav.forum", label: "Community", href: ROUTES.PATIENT.FORUM, icon: MessageCircle },
];

export const PATIENT_BOTTOM_NAV_ITEMS: NavItem[] = [
  PATIENT_NAV_ITEMS[0],
  PATIENT_NAV_ITEMS[1],
  PATIENT_NAV_ITEMS[2],
  PATIENT_NAV_ITEMS[3],
  PATIENT_NAV_ITEMS[4],
];

export const EXPERT_NAV_ITEMS: NavItem[] = [
  { id: "dashboard", labelKey: "nav.dashboard", label: "Dashboard", href: ROUTES.EXPERT.DASHBOARD, icon: LayoutDashboard },
  { id: "schedule", labelKey: "nav.schedule", label: "Schedule", href: ROUTES.EXPERT.SCHEDULE, icon: CalendarClock },
  { id: "appointments", labelKey: "nav.appointments", label: "Appointments", href: ROUTES.EXPERT.APPOINTMENTS, icon: CalendarCheck },
  { id: "patients", labelKey: "nav.patients", label: "My Patients", href: ROUTES.EXPERT.PATIENTS, icon: Users },
  { id: "clinical-records", labelKey: "nav.clinicalRecords", label: "Clinical Records", href: ROUTES.EXPERT.CLINICAL_RECORDS, icon: FileText },
  { id: "wallet", labelKey: "nav.wallet", label: "Wallet", href: ROUTES.EXPERT.WALLET, icon: Wallet },
  { id: "messages", labelKey: "nav.messages", label: "Messages", href: ROUTES.EXPERT.MESSAGES, icon: MessageSquare },
  { id: "forum", labelKey: "nav.forum", label: "Community", href: ROUTES.EXPERT.FORUM, icon: MessageCircle },
];

export const EXPERT_BOTTOM_NAV_ITEMS: NavItem[] = [
  EXPERT_NAV_ITEMS[0],
  EXPERT_NAV_ITEMS[1],
  EXPERT_NAV_ITEMS[2],
  EXPERT_NAV_ITEMS[3],
];

export const ADMIN_NAV_ITEMS: NavItem[] = [
  { id: "dashboard", labelKey: "nav.dashboard", label: "Dashboard", href: ROUTES.ADMIN.DASHBOARD, icon: LayoutDashboard },
  { id: "experts", labelKey: "nav.experts", label: "Experts", href: ROUTES.ADMIN.EXPERTS, icon: Users },
  { id: "specializations", labelKey: "nav.specializations", label: "Specializations", href: ROUTES.ADMIN.SPECIALIZATIONS, icon: ClipboardList },
  { id: "patients", labelKey: "nav.patients", label: "Patients", href: ROUTES.ADMIN.PATIENTS, icon: Users },
  { id: "appointments", labelKey: "nav.appointments", label: "Appointments", href: ROUTES.ADMIN.APPOINTMENTS, icon: CalendarCheck },
  { id: "schedules", labelKey: "nav.schedules", label: "Schedules", href: ROUTES.ADMIN.SCHEDULES, icon: CalendarClock },
  { id: "withdrawals", labelKey: "nav.withdrawals", label: "Withdrawals", href: ROUTES.ADMIN.WITHDRAWALS, icon: Wallet },
  { id: "templates", labelKey: "nav.templates", label: "Templates", href: ROUTES.ADMIN.TEMPLATES, icon: ClipboardCheck },
  { id: "dimensions", labelKey: "nav.dimensions", label: "Dimensions", href: ROUTES.ADMIN.DIMENSIONS, icon: Layers },
  { id: "option-groups", labelKey: "nav.optionGroups", label: "Option Groups", href: ROUTES.ADMIN.OPTION_GROUPS, icon: ClipboardList },
  { id: "questions", labelKey: "nav.questions", label: "Questions", href: ROUTES.ADMIN.QUESTIONS, icon: FileText },
];

export const SETTINGS_NAV_ITEM = {
  patient: { id: "settings", labelKey: "nav.settings", label: "Settings", href: ROUTES.PATIENT.SETTINGS, icon: Settings },
  expert: { id: "settings", labelKey: "nav.settings", label: "Settings", href: ROUTES.EXPERT.SETTINGS, icon: Settings },
};
