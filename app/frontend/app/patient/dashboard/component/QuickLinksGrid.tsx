import Link from "next/link";
import { Search, ClipboardCheck, FileText, MessageSquare } from "lucide-react";
import { Card } from "@/components/ui";
import { ROUTES } from "@/constants";

const QUICK_LINKS = [
  { id: "find-experts", label: "Find Experts", href: ROUTES.PATIENT.FIND_EXPERTS, icon: Search },
  { id: "assessment", label: "Take an Assessment", href: ROUTES.PATIENT.ASSESSMENT, icon: ClipboardCheck },
  { id: "medical-history", label: "Medical History", href: ROUTES.PATIENT.MEDICAL_HISTORY, icon: FileText },
  { id: "messages", label: "Messages", href: ROUTES.PATIENT.MESSAGES, icon: MessageSquare },
];

export function QuickLinksGrid() {
  return (
    <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
      {QUICK_LINKS.map((link) => {
        const Icon = link.icon;
        return (
          <Link key={link.id} href={link.href}>
            <Card className="flex flex-col items-center gap-2 p-5 text-center transition-shadow hover:shadow-elevated">
              <Icon className="h-6 w-6 text-primary" />
              <span className="text-sm font-semibold text-foreground">{link.label}</span>
            </Card>
          </Link>
        );
      })}
    </div>
  );
}
