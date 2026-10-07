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
  PenSquare,
  GraduationCap,
  ListChecks,
  type LucideIcon,
} from "lucide-react";
import { ROUTES } from "./route";

export interface NavItem {
  id: string;
  labelKey: string;
  label: string;
  href: string;
  icon: LucideIcon;
  /** One-line summary shown on dashboard feature cards. */
  descriptionKey?: string;
  description?: string;
  /** Route exists but only renders the ComingSoon placeholder. */
  comingSoon?: boolean;
}

/** A titled group of feature cards on a role's dashboard. */
export interface FeatureGroup {
  id: string;
  labelKey: string;
  label: string;
  items: NavItem[];
}

function pick(items: NavItem[], ids: string[]): NavItem[] {
  return ids.map((id) => {
    const item = items.find((candidate) => candidate.id === id);
    if (!item) throw new Error(`Unknown nav item "${id}"`);
    return item;
  });
}

// ---------------------------------------------------------------------------
// Patient
// ---------------------------------------------------------------------------

/** Every patient page — the header shows only the core subset, the dashboard lists the rest. */
export const PATIENT_NAV_ITEMS: NavItem[] = [
  { id: "dashboard", labelKey: "nav.dashboard", label: "Dashboard", href: ROUTES.PATIENT.DASHBOARD, icon: LayoutDashboard },
  { id: "find-experts", labelKey: "nav.findExperts", label: "Find Experts", href: ROUTES.PATIENT.FIND_EXPERTS, icon: Search },
  {
    id: "book-appointment", labelKey: "nav.bookAppointment", label: "Book Appointment", href: ROUTES.PATIENT.BOOK_APPOINTMENT, icon: CalendarPlus,
    descriptionKey: "nav.desc.bookAppointment", description: "Pick a topic, an expert and a time slot.",
  },
  { id: "my-bookings", labelKey: "nav.myBookings", label: "My Bookings", href: ROUTES.PATIENT.MY_BOOKINGS, icon: ClipboardList },
  {
    id: "wallet", labelKey: "nav.wallet", label: "Wallet", href: ROUTES.PATIENT.WALLET, icon: Wallet,
    descriptionKey: "nav.desc.wallet", description: "Check your balance, top up or withdraw.",
  },
  {
    id: "assessment", labelKey: "nav.assessment", label: "Assessment", href: ROUTES.PATIENT.ASSESSMENT, icon: ClipboardCheck,
    descriptionKey: "nav.desc.assessment", description: "Take a self-assessment and review past results.",
  },
  {
    id: "medical-history", labelKey: "nav.medicalHistory", label: "Medical History", href: ROUTES.PATIENT.MEDICAL_HISTORY, icon: FileText,
    descriptionKey: "nav.desc.medicalHistory", description: "Your health history and consultation records.",
  },
  { id: "messages", labelKey: "nav.messages", label: "Messages", href: ROUTES.PATIENT.MESSAGES, icon: MessageSquare },
  { id: "forum", labelKey: "nav.forum", label: "Community", href: ROUTES.PATIENT.FORUM, icon: MessageCircle },
  {
    id: "my-posts", labelKey: "nav.myPosts", label: "My Posts", href: ROUTES.PATIENT.FORUM_MY_POSTS, icon: PenSquare,
    descriptionKey: "nav.desc.myPosts", description: "Posts you have shared with the community.",
  },
];

export const PATIENT_HEADER_NAV_ITEMS = pick(PATIENT_NAV_ITEMS, ["dashboard", "find-experts", "my-bookings", "messages", "forum"]);
export const PATIENT_BOTTOM_NAV_ITEMS = PATIENT_HEADER_NAV_ITEMS;

export const PATIENT_DASHBOARD_FEATURES: FeatureGroup[] = [
  { id: "care", labelKey: "nav.group.care", label: "Care", items: pick(PATIENT_NAV_ITEMS, ["book-appointment", "assessment", "medical-history"]) },
  { id: "account", labelKey: "nav.group.account", label: "Account & community", items: pick(PATIENT_NAV_ITEMS, ["wallet", "my-posts"]) },
];

// ---------------------------------------------------------------------------
// Expert
// ---------------------------------------------------------------------------

export const EXPERT_NAV_ITEMS: NavItem[] = [
  { id: "dashboard", labelKey: "nav.dashboard", label: "Dashboard", href: ROUTES.EXPERT.DASHBOARD, icon: LayoutDashboard },
  { id: "schedule", labelKey: "nav.schedule", label: "Schedule", href: ROUTES.EXPERT.SCHEDULE, icon: CalendarClock },
  { id: "appointments", labelKey: "nav.appointments", label: "Appointments", href: ROUTES.EXPERT.APPOINTMENTS, icon: CalendarCheck },
  {
    id: "patients", labelKey: "nav.patients", label: "My Patients", href: ROUTES.EXPERT.PATIENTS, icon: Users, comingSoon: true,
    descriptionKey: "nav.desc.expertPatients", description: "Everyone you have counselled so far.",
  },
  {
    id: "clinical-records", labelKey: "nav.clinicalRecords", label: "Clinical Records", href: ROUTES.EXPERT.CLINICAL_RECORDS, icon: FileText,
    descriptionKey: "nav.desc.clinicalRecords", description: "Session notes and medical records.",
  },
  {
    id: "wallet", labelKey: "nav.wallet", label: "Wallet", href: ROUTES.EXPERT.WALLET, icon: Wallet, comingSoon: true,
    descriptionKey: "nav.desc.expertWallet", description: "Earnings and withdrawal requests.",
  },
  { id: "messages", labelKey: "nav.messages", label: "Messages", href: ROUTES.EXPERT.MESSAGES, icon: MessageSquare },
  {
    id: "forum", labelKey: "nav.forum", label: "Community", href: ROUTES.EXPERT.FORUM, icon: MessageCircle,
    descriptionKey: "nav.desc.forum", description: "Answer questions and share knowledge.",
  },
];

export const EXPERT_HEADER_NAV_ITEMS = pick(EXPERT_NAV_ITEMS, ["dashboard", "schedule", "appointments", "messages"]);
export const EXPERT_BOTTOM_NAV_ITEMS = EXPERT_HEADER_NAV_ITEMS;

export const EXPERT_DASHBOARD_FEATURES: FeatureGroup[] = [
  { id: "care", labelKey: "nav.group.patientCare", label: "Patient care", items: pick(EXPERT_NAV_ITEMS, ["clinical-records", "patients"]) },
  { id: "account", labelKey: "nav.group.account", label: "Account & community", items: pick(EXPERT_NAV_ITEMS, ["forum", "wallet"]) },
];

// ---------------------------------------------------------------------------
// Admin
// ---------------------------------------------------------------------------

export const ADMIN_NAV_ITEMS: NavItem[] = [
  { id: "dashboard", labelKey: "nav.dashboard", label: "Dashboard", href: ROUTES.ADMIN.DASHBOARD, icon: LayoutDashboard },
  { id: "experts", labelKey: "nav.experts", label: "Experts", href: ROUTES.ADMIN.EXPERTS, icon: GraduationCap },
  {
    id: "specializations", labelKey: "nav.specializations", label: "Specializations", href: ROUTES.ADMIN.SPECIALIZATIONS, icon: ClipboardList,
    descriptionKey: "nav.desc.specializations", description: "Areas of expertise experts can list.",
  },
  { id: "patients", labelKey: "nav.patients", label: "Patients", href: ROUTES.ADMIN.PATIENTS, icon: Users },
  {
    id: "appointments", labelKey: "nav.appointments", label: "Appointments", href: ROUTES.ADMIN.APPOINTMENTS, icon: CalendarCheck, comingSoon: true,
    descriptionKey: "nav.desc.adminAppointments", description: "Every booking across the platform.",
  },
  {
    id: "schedules", labelKey: "nav.schedules", label: "Schedules", href: ROUTES.ADMIN.SCHEDULES, icon: CalendarClock, comingSoon: true,
    descriptionKey: "nav.desc.schedules", description: "Shift templates and expert availability.",
  },
  {
    id: "withdrawals", labelKey: "nav.withdrawals", label: "Withdrawals", href: ROUTES.ADMIN.WITHDRAWALS, icon: Wallet, comingSoon: true,
    descriptionKey: "nav.desc.withdrawals", description: "Review and process payout requests.",
  },
  {
    id: "templates", labelKey: "nav.templates", label: "Templates", href: ROUTES.ADMIN.TEMPLATES, icon: ClipboardCheck,
    descriptionKey: "nav.desc.templates", description: "Assessment templates patients can take.",
  },
  {
    id: "dimensions", labelKey: "nav.dimensions", label: "Dimensions", href: ROUTES.ADMIN.DIMENSIONS, icon: Layers,
    descriptionKey: "nav.desc.dimensions", description: "Scoring dimensions used by templates.",
  },
  {
    id: "option-groups", labelKey: "nav.optionGroups", label: "Option Groups", href: ROUTES.ADMIN.OPTION_GROUPS, icon: ListChecks,
    descriptionKey: "nav.desc.optionGroups", description: "Reusable answer scales for questions.",
  },
  {
    id: "questions", labelKey: "nav.questions", label: "Questions", href: ROUTES.ADMIN.QUESTIONS, icon: FileText,
    descriptionKey: "nav.desc.questions", description: "The question bank behind every assessment.",
  },
  { id: "forum-posts", labelKey: "nav.forumPosts", label: "Forum Posts", href: ROUTES.ADMIN.FORUM_POSTS, icon: MessageCircle },
];

export const ADMIN_HEADER_NAV_ITEMS = pick(ADMIN_NAV_ITEMS, ["dashboard", "experts", "patients", "forum-posts"]);

export const ADMIN_DASHBOARD_FEATURES: FeatureGroup[] = [
  {
    id: "assessment", labelKey: "nav.group.assessmentBuilder", label: "Assessment builder",
    items: pick(ADMIN_NAV_ITEMS, ["templates", "questions", "dimensions", "option-groups"]),
  },
  {
    id: "operations", labelKey: "nav.group.operations", label: "Operations",
    items: pick(ADMIN_NAV_ITEMS, ["appointments", "schedules", "specializations", "withdrawals"]),
  },
];

export const SETTINGS_NAV_ITEM = {
  patient: { id: "settings", labelKey: "nav.settings", label: "Settings", href: ROUTES.PATIENT.SETTINGS, icon: Settings },
  expert: { id: "settings", labelKey: "nav.settings", label: "Settings", href: ROUTES.EXPERT.SETTINGS, icon: Settings },
};

/** Dashboard links match exactly; every other item also stays active on its sub-pages. */
export function isNavItemActive(item: NavItem, pathname: string): boolean {
  if (item.id === "dashboard") return pathname === item.href;
  return pathname === item.href || pathname.startsWith(`${item.href}/`);
}
