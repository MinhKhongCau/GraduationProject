import Link from "next/link";
import { Search, ClipboardCheck, FileText, MessageSquare, MessageCircle, CalendarPlus, Wallet, PenSquare } from "lucide-react";
import { Card } from "@/components/ui";
import { ROUTES } from "@/constants";

const QUICK_LINKS = [
  { id: "find-experts", label: "Find Experts", href: ROUTES.PATIENT.FIND_EXPERTS, icon: Search },
  { id: "book-appointment", label: "Booking", href: ROUTES.PATIENT.BOOK_APPOINTMENT, icon: CalendarPlus },
  { id: "wallet", label: "Wallet", href: ROUTES.PATIENT.WALLET, icon: Wallet },
  { id: "assessment", label: "Take an Assessment", href: ROUTES.PATIENT.ASSESSMENT, icon: ClipboardCheck },
  { id: "medical-history", label: "Medical History", href: ROUTES.PATIENT.MEDICAL_HISTORY, icon: FileText },
  { id: "messages", label: "Messages", href: ROUTES.PATIENT.MESSAGES, icon: MessageSquare },
  { id: "community", label: "Community", href: ROUTES.PATIENT.FORUM, icon: MessageCircle },
  { id: "my-posts", label: "My Posts", href: ROUTES.PATIENT.FORUM_MY_POSTS, icon: PenSquare },
];

export function QuickLinksGrid() {
  return (
    <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
      {QUICK_LINKS.map((link) => {
        const Icon = link.icon;
        return (
          <Link key={link.id} href={link.href} className="group block rounded-xl">
            <Card className="flex h-full flex-col items-center gap-3 p-5 text-center transition-all duration-200 group-hover:border-primary/40 group-hover:shadow-elevated">
              <span className="flex h-11 w-11 items-center justify-center rounded-lg bg-primary-soft text-primary">
                <Icon className="h-5 w-5" />
              </span>
              <span className="text-sm font-semibold text-foreground">{link.label}</span>
            </Card>
          </Link>
        );
      })}
    </div>
  );
}
