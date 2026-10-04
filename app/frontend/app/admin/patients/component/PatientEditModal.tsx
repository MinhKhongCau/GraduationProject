"use client";

import { useEffect, useState, type FormEvent } from "react";
import { Modal, Button, Spinner, Label, Input, Select } from "@/components/ui";
import { useApiQuery } from "@/hooks";
import { patientApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import type { UpdatePatientProfileRequest } from "@/types";

export interface PatientEditModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  accountId: string | null;
  isPending: boolean;
  onSubmit: (values: UpdatePatientProfileRequest) => void;
}

export function PatientEditModal({
  open,
  onOpenChange,
  accountId,
  isPending,
  onSubmit,
}: PatientEditModalProps) {
  const { data: patient, isLoading } = useApiQuery({
    queryKey: accountId ? QUERY_KEYS.patientProfile(accountId) : ["patient", "profile", "none"],
    queryFn: () => patientApi.getPatientProfile(accountId!),
    enabled: open && !!accountId,
  });

  const [fullName, setFullName] = useState("");
  const [phoneNumber, setPhoneNumber] = useState("");
  const [email, setEmail] = useState("");
  const [dateOfBirth, setDateOfBirth] = useState("");
  const [gender, setGender] = useState("");
  const [address, setAddress] = useState("");

  useEffect(() => {
    if (open && patient) {
      setFullName(patient.fullName);
      setPhoneNumber(patient.phoneNumber ?? "");
      setEmail(patient.email ?? "");
      setDateOfBirth(patient.dateOfBirth ? patient.dateOfBirth.slice(0, 10) : "");
      setGender(patient.gender ?? "");
      setAddress(patient.address ?? "");
    }
  }, [open, patient]);

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    onSubmit({
      fullName,
      phoneNumber: phoneNumber || undefined,
      email: email || undefined,
      dateOfBirth,
      gender: gender || undefined,
      address: address || undefined,
    });
  }

  return (
    <Modal open={open} onOpenChange={onOpenChange} title="Hồ sơ bệnh nhân" size="lg">
      {isLoading ? (
        <div className="flex justify-center py-10">
          <Spinner className="h-6 w-6" />
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          <div>
            <Label htmlFor="patienteditmodal-fullName">Họ và tên</Label>
            <Input
              id="patienteditmodal-fullName"
              required
              value={fullName}
              onChange={(e) => setFullName(e.target.value)}
            />
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <Label htmlFor="patienteditmodal-phoneNumber">Số điện thoại</Label>
              <Input
                id="patienteditmodal-phoneNumber"
                value={phoneNumber}
                onChange={(e) => setPhoneNumber(e.target.value)}
              />
            </div>
            <div>
              <Label htmlFor="patienteditmodal-email">Email</Label>
              <Input
                id="patienteditmodal-email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <Label htmlFor="patienteditmodal-dateOfBirth">Ngày sinh</Label>
              <Input
                id="patienteditmodal-dateOfBirth"
                type="date"
                required
                value={dateOfBirth}
                onChange={(e) => setDateOfBirth(e.target.value)}
              />
            </div>
            <div>
              <Label htmlFor="patienteditmodal-gender">Giới tính</Label>
              <Select
                id="patienteditmodal-gender"
                value={gender}
                onChange={(e) => setGender(e.target.value)}
              >
                <option value="">—</option>
                <option value="MALE">Nam</option>
                <option value="FEMALE">Nữ</option>
                <option value="OTHER">Khác</option>
              </Select>
            </div>
          </div>

          <div>
            <Label htmlFor="patienteditmodal-address">Địa chỉ</Label>
            <Input
              id="patienteditmodal-address"
              value={address}
              onChange={(e) => setAddress(e.target.value)}
            />
          </div>

          {patient?.medicalHistories && patient.medicalHistories.length > 0 && (
            <div className="border-t border-border pt-4">
              <h4 className="mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                Tiền sử bệnh
              </h4>
              <div className="max-h-40 space-y-2 overflow-y-auto">
                {patient.medicalHistories.map((history) => (
                  <div
                    key={history.historyId}
                    className="rounded-xl border border-border bg-surface/50 p-3 text-xs"
                  >
                    <div className="flex items-center justify-between">
                      <span className="font-semibold text-foreground">{history.conditionName}</span>
                      <span className="text-muted-foreground">{history.diagnosedAt?.slice(0, 10)}</span>
                    </div>
                    {history.description && (
                      <p className="mt-1 text-muted-foreground">{history.description}</p>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          <div className="flex justify-end gap-2 border-t border-border pt-4">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Hủy
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? "Đang lưu..." : "Lưu thay đổi"}
            </Button>
          </div>
        </form>
      )}
    </Modal>
  );
}
