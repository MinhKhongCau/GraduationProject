import { Card } from "@/components/ui";
import type { PatientProfile } from "@/types";

export function PersonalInfoSection({ profile }: { profile: PatientProfile }) {
  const rows: [string, string][] = [
    ["Date of birth", profile.dateOfBirth ?? "—"],
    ["Gender", profile.gender ?? "—"],
    ["Phone number", profile.phoneNumber ?? "—"],
    ["Address", profile.address ?? "—"],
  ];

  return (
    <Card className="overflow-hidden">
      <table className="w-full text-left text-sm">
        <tbody className="divide-y divide-border">
          {rows.map(([label, value]) => (
            <tr key={label} className="transition-colors hover:bg-surface/50">
              <td className="w-1/3 px-6 py-4 text-muted-foreground">{label}</td>
              <td className="px-6 py-4 font-medium text-foreground">{value}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </Card>
  );
}
