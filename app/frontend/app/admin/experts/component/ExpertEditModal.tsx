"use client";

import { useEffect, useState, type FormEvent } from "react";
import { Modal, Button, Spinner, Label, Input, Textarea } from "@/components/ui";
import { useApiQuery } from "@/hooks";
import { expertApi, specializationApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import type { UpdateExpertProfileRequest } from "@/types";

export interface ExpertEditModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  accountId: string | null;
  isPending: boolean;
  onSubmit: (values: UpdateExpertProfileRequest) => void;
}

export function ExpertEditModal({
  open,
  onOpenChange,
  accountId,
  isPending,
  onSubmit,
}: ExpertEditModalProps) {
  const { data: expert, isLoading } = useApiQuery({
    queryKey: accountId ? QUERY_KEYS.expertProfile(accountId) : ["expert", "profile", "none"],
    queryFn: () => expertApi.getExpertProfile(accountId!),
    enabled: open && !!accountId,
  });

  const { data: allSpecializations = [] } = useApiQuery({
    queryKey: QUERY_KEYS.specializations(),
    queryFn: () => specializationApi.getAllSpecializations(),
    enabled: open,
  });

  const [fullName, setFullName] = useState("");
  const [phoneNumber, setPhoneNumber] = useState("");
  const [email, setEmail] = useState("");
  const [introductionVideoUrl, setIntroductionVideoUrl] = useState("");
  const [bio, setBio] = useState("");
  const [specializationIds, setSpecializationIds] = useState<string[]>([]);

  useEffect(() => {
    if (open && expert) {
      setFullName(expert.fullName);
      setPhoneNumber(expert.phoneNumber ?? "");
      setEmail(expert.email ?? "");
      setIntroductionVideoUrl(expert.introductionVideoUrl ?? "");
      setBio(expert.bio ?? "");
      setSpecializationIds(expert.specializations.map((spec) => spec.specId));
    }
  }, [open, expert]);

  function toggleSpecialization(specId: string) {
    setSpecializationIds((current) =>
      current.includes(specId) ? current.filter((id) => id !== specId) : [...current, specId]
    );
  }

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    onSubmit({
      fullName,
      phoneNumber: phoneNumber || undefined,
      email: email || undefined,
      introductionVideoUrl: introductionVideoUrl || undefined,
      bio: bio || undefined,
      specializationIds,
    });
  }

  return (
    <Modal open={open} onOpenChange={onOpenChange} title="Hồ sơ chuyên gia" size="lg">
      {isLoading ? (
        <div className="flex justify-center py-10">
          <Spinner className="h-6 w-6" />
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          <div>
            <Label htmlFor="experteditmodal-fullName">Họ và tên</Label>
            <Input
              id="experteditmodal-fullName"
              required
              value={fullName}
              onChange={(e) => setFullName(e.target.value)}
            />
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <Label htmlFor="experteditmodal-phoneNumber">Số điện thoại</Label>
              <Input
                id="experteditmodal-phoneNumber"
                value={phoneNumber}
                onChange={(e) => setPhoneNumber(e.target.value)}
              />
            </div>
            <div>
              <Label htmlFor="experteditmodal-email">Email</Label>
              <Input
                id="experteditmodal-email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
          </div>

          <div>
            <Label htmlFor="experteditmodal-introductionVideoUrl">Video giới thiệu (URL)</Label>
            <Input
              id="experteditmodal-introductionVideoUrl"
              value={introductionVideoUrl}
              onChange={(e) => setIntroductionVideoUrl(e.target.value)}
            />
          </div>

          <div>
            <Label htmlFor="experteditmodal-bio">Tiểu sử</Label>
            <Textarea
              id="experteditmodal-bio"
              rows={3}
              value={bio}
              onChange={(e) => setBio(e.target.value)}
            />
          </div>

          <div>
            <Label className="mb-2">Chuyên khoa</Label>
            <div className="flex flex-wrap gap-2">
              {allSpecializations.map((spec) => {
                const isSelected = specializationIds.includes(spec.specId);
                return (
                  <button
                    key={spec.specId}
                    type="button"
                    aria-pressed={isSelected}
                    onClick={() => toggleSpecialization(spec.specId)}
                    className={`h-8 rounded-lg border px-3 text-xs font-medium transition-colors ${
                      isSelected
                        ? "border-primary bg-primary-soft text-primary-soft-text"
                        : "border-border text-muted-foreground hover:bg-surface"
                    }`}
                  >
                    {spec.name}
                  </button>
                );
              })}
              {allSpecializations.length === 0 && (
                <p className="text-xs text-muted-foreground">Chưa có chuyên khoa nào.</p>
              )}
            </div>
          </div>

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
