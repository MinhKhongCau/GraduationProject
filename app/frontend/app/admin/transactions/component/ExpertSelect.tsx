"use client";

import { Label, Select } from "@/components/ui";
import type { ExpertProfile } from "@/types";

export interface ExpertSelectProps {
  experts: ExpertProfile[];
  value?: string;
  onChange: (expertId?: string) => void;
}

/** Only the experts this admin manages can be picked — the API answers 403 for any other. */
export function ExpertSelect({ experts, value, onChange }: ExpertSelectProps) {
  return (
    <div>
      <Label htmlFor="transactions-expert">Chuyên gia</Label>
      <Select id="transactions-expert" value={value ?? ""} onChange={(e) => onChange(e.target.value || undefined)}>
        <option value="">Tất cả chuyên gia tôi quản lý</option>
        {experts.map((expert) => (
          <option key={expert.accountId} value={expert.accountId}>
            {expert.fullName || expert.email}
          </option>
        ))}
      </Select>
    </div>
  );
}
