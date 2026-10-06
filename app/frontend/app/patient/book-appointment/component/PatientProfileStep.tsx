"use client";

import { useState, type FormEvent } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Check, Plus, UserRound } from "lucide-react";
import { Badge, Button, Card, FieldError, Input, Label, PageHeader, Select, Spinner } from "@/components/ui";
import { useApiMutation, useApiQuery } from "@/hooks";
import { patientRecordApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import type { CreatePatientRecordRequest, PatientRecord, PatientRecordRelationship } from "@/types";

export const RELATIONSHIP_LABELS: Record<PatientRecordRelationship, string> = {
  SELF: "Myself",
  PARENT: "Parent",
  CHILD: "Child",
  SPOUSE: "Spouse",
  SIBLING: "Sibling",
  RELATIVE: "Relative",
  OTHER: "Other",
};

const NEW_RECORD_RELATIONSHIPS = ["PARENT", "CHILD", "SPOUSE", "SIBLING", "RELATIVE", "OTHER"] as const;

const EMPTY_FORM: CreatePatientRecordRequest = {
  fullName: "",
  dateOfBirth: "",
  gender: "MALE",
  phoneNumber: "",
  email: "",
  address: "",
  relationship: "PARENT",
};

type FormErrors = Partial<Record<keyof CreatePatientRecordRequest, string>>;

function validate(form: CreatePatientRecordRequest): FormErrors {
  const errors: FormErrors = {};
  if (!form.fullName.trim()) errors.fullName = "Full name is required.";
  if (!form.dateOfBirth) errors.dateOfBirth = "Date of birth is required.";
  else if (new Date(`${form.dateOfBirth}T00:00:00`) > new Date()) errors.dateOfBirth = "Date of birth can't be in the future.";
  if (!form.phoneNumber.trim()) errors.phoneNumber = "Phone number is required.";
  else if (form.phoneNumber.trim().length > 20) errors.phoneNumber = "Phone number is too long.";
  if (form.email && !/^\S+@\S+\.\S+$/.test(form.email)) errors.email = "Enter a valid email.";
  return errors;
}

function formatDateOfBirth(value: string | null): string | null {
  if (!value) return null;
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? value
    : date.toLocaleDateString("en-GB", { day: "2-digit", month: "short", year: "numeric" });
}

export interface PatientProfileStepProps {
  selectedRecord: PatientRecord | null;
  onSelectRecord: (record: PatientRecord) => void;
  onBack: () => void;
  onNext: () => void;
  /** True while the slot is being locked before moving to the confirm page. */
  isSubmitting: boolean;
}

export function PatientProfileStep({ selectedRecord, onSelectRecord, onBack, onNext, isSubmitting }: PatientProfileStepProps) {
  const queryClient = useQueryClient();
  const [isCreating, setIsCreating] = useState(false);
  const [form, setForm] = useState<CreatePatientRecordRequest>(EMPTY_FORM);
  const [errors, setErrors] = useState<FormErrors>({});

  const { data: records = [], isLoading } = useApiQuery({
    queryKey: QUERY_KEYS.myPatientRecords(),
    queryFn: () => patientRecordApi.getMyPatientRecords(),
  });

  const createMutation = useApiMutation({
    mutationFn: (payload: CreatePatientRecordRequest) => patientRecordApi.createMyPatientRecord(payload),
    onSuccess: (record) => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.myPatientRecords() });
      onSelectRecord(record);
      setIsCreating(false);
      setForm(EMPTY_FORM);
    },
  });

  function updateField<K extends keyof CreatePatientRecordRequest>(key: K, value: CreatePatientRecordRequest[K]) {
    setForm((current) => ({ ...current, [key]: value }));
    setErrors((current) => ({ ...current, [key]: undefined }));
  }

  function submitNewRecord(event: FormEvent) {
    event.preventDefault();
    const nextErrors = validate(form);
    setErrors(nextErrors);
    if (Object.keys(nextErrors).length > 0) return;
    createMutation.mutate({
      ...form,
      fullName: form.fullName.trim(),
      phoneNumber: form.phoneNumber.trim(),
      email: form.email?.trim() || undefined,
      address: form.address?.trim() || undefined,
    });
  }

  return (
    <div className="flex h-full flex-col">
      <PageHeader title="Who is this appointment for?" description="Choose a saved profile or add a new one for a family member." />

      {isLoading ? (
        <Spinner className="h-6 w-6" />
      ) : (
        <div className="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {records.map((record) => {
            const isSelected = selectedRecord?.recordId === record.recordId;
            const dateOfBirth = formatDateOfBirth(record.dateOfBirth);
            return (
              <button
                key={record.recordId}
                type="button"
                onClick={() => onSelectRecord(record)}
                aria-pressed={isSelected}
                className={`relative flex items-start gap-3 rounded-xl border p-4 text-left transition-all ${
                  isSelected ? "border-primary bg-primary-soft ring-1 ring-primary" : "border-border-strong bg-background hover:border-primary/40 hover:shadow-card"
                }`}
              >
                {isSelected && (
                  <div className="absolute right-3 top-3 text-primary">
                    <Check className="h-4 w-4 stroke-[3]" />
                  </div>
                )}
                <UserRound className="mt-0.5 h-5 w-5 shrink-0 text-muted-foreground" />
                <div className="min-w-0 space-y-1">
                  <p className="truncate pr-5 text-sm font-semibold text-foreground">{record.fullName || "Unnamed profile"}</p>
                  <Badge tone={record.relationship === "SELF" ? "primary" : "neutral"}>{RELATIONSHIP_LABELS[record.relationship]}</Badge>
                  <p className="text-xs text-muted-foreground">
                    {[dateOfBirth, record.phoneNumber].filter(Boolean).join(" · ") || "No contact details yet"}
                  </p>
                </div>
              </button>
            );
          })}

          {!isCreating && (
            <button
              type="button"
              onClick={() => setIsCreating(true)}
              className="flex min-h-24 items-center justify-center gap-2 rounded-xl border border-dashed border-border-strong bg-background p-4 text-sm font-semibold text-muted-foreground transition-colors hover:border-primary/40 hover:text-primary"
            >
              <Plus className="h-4 w-4" />
              Create new profile
            </button>
          )}
        </div>
      )}

      {isCreating && (
        <Card className="mb-6 max-w-2xl p-6">
          <form onSubmit={submitNewRecord} noValidate className="space-y-4">
            <p className="text-sm font-semibold text-foreground">New patient profile</p>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div className="sm:col-span-2">
                <Label htmlFor="record-full-name">Full name *</Label>
                <Input id="record-full-name" value={form.fullName} invalid={!!errors.fullName} onChange={(e) => updateField("fullName", e.target.value)} />
                <FieldError>{errors.fullName}</FieldError>
              </div>
              <div>
                <Label htmlFor="record-relationship">Relationship *</Label>
                <Select
                  id="record-relationship"
                  value={form.relationship}
                  onChange={(e) => updateField("relationship", e.target.value as CreatePatientRecordRequest["relationship"])}
                >
                  {NEW_RECORD_RELATIONSHIPS.map((relationship) => (
                    <option key={relationship} value={relationship}>
                      {RELATIONSHIP_LABELS[relationship]}
                    </option>
                  ))}
                </Select>
              </div>
              <div>
                <Label htmlFor="record-gender">Gender *</Label>
                <Select id="record-gender" value={form.gender} onChange={(e) => updateField("gender", e.target.value as CreatePatientRecordRequest["gender"])}>
                  <option value="MALE">Male</option>
                  <option value="FEMALE">Female</option>
                  <option value="OTHER">Other</option>
                </Select>
              </div>
              <div>
                <Label htmlFor="record-dob">Date of birth *</Label>
                <Input
                  id="record-dob"
                  type="date"
                  value={form.dateOfBirth}
                  invalid={!!errors.dateOfBirth}
                  onChange={(e) => updateField("dateOfBirth", e.target.value)}
                />
                <FieldError>{errors.dateOfBirth}</FieldError>
              </div>
              <div>
                <Label htmlFor="record-phone">Phone number *</Label>
                <Input
                  id="record-phone"
                  type="tel"
                  value={form.phoneNumber}
                  invalid={!!errors.phoneNumber}
                  onChange={(e) => updateField("phoneNumber", e.target.value)}
                />
                <FieldError>{errors.phoneNumber}</FieldError>
              </div>
              <div>
                <Label htmlFor="record-email">Email</Label>
                <Input id="record-email" type="email" value={form.email} invalid={!!errors.email} onChange={(e) => updateField("email", e.target.value)} />
                <FieldError>{errors.email}</FieldError>
              </div>
              <div>
                <Label htmlFor="record-address">Address</Label>
                <Input id="record-address" value={form.address} onChange={(e) => updateField("address", e.target.value)} />
              </div>
            </div>
            <div className="flex justify-end gap-3">
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setIsCreating(false);
                  setErrors({});
                }}
                disabled={createMutation.isPending}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={createMutation.isPending}>
                {createMutation.isPending ? "Saving..." : "Save profile"}
              </Button>
            </div>
          </form>
        </Card>
      )}

      <div className="mt-auto flex justify-between border-t border-border pt-6">
        <Button variant="outline" onClick={onBack} disabled={isSubmitting}>
          Back
        </Button>
        <Button onClick={onNext} disabled={!selectedRecord || isSubmitting || isCreating}>
          {isSubmitting ? "Holding your slot..." : "Next"}
          {!isSubmitting && <Check className="h-4 w-4" />}
        </Button>
      </div>
    </div>
  );
}
