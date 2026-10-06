"use client";

import { useEffect, useState, type FormEvent } from "react";
import { Modal, Button, Label, Input, Textarea } from "@/components/ui";
import type { AdminSpecialization, CreateSpecializationRequest } from "@/types";

export interface SpecializationFormModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** null = create a new specialization. */
  specialization: AdminSpecialization | null;
  isPending: boolean;
  onSubmit: (values: CreateSpecializationRequest) => void;
}

/** Symptoms are edited as one comma-separated line and sent as a trimmed, de-duplicated list. */
function parseSymptoms(value: string): string[] {
  return Array.from(new Set(value.split(",").map((item) => item.trim()).filter(Boolean)));
}

export function SpecializationFormModal({
  open,
  onOpenChange,
  specialization,
  isPending,
  onSubmit,
}: SpecializationFormModalProps) {
  const [code, setCode] = useState("");
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [description, setDescription] = useState("");
  const [symptoms, setSymptoms] = useState("");
  const [location, setLocation] = useState("");
  const [imageUrl, setImageUrl] = useState("");

  useEffect(() => {
    if (!open) return;
    // Reset the form each time the modal opens, for the selected specialization or a blank create form.
    /* eslint-disable react-hooks/set-state-in-effect */
    setCode(specialization?.code ?? "");
    setName(specialization?.name ?? "");
    setSlug(specialization?.slug ?? "");
    setDescription(specialization?.description ?? "");
    setSymptoms(specialization?.symptoms.join(", ") ?? "");
    setLocation(specialization?.location ?? "");
    setImageUrl(specialization?.imageUrl ?? "");
    /* eslint-enable react-hooks/set-state-in-effect */
  }, [open, specialization]);

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    onSubmit({
      code: code.trim(),
      name: name.trim(),
      slug: slug.trim() || undefined,
      description: description.trim() || undefined,
      symptoms: parseSymptoms(symptoms),
      location: location.trim() || undefined,
      imageUrl: imageUrl.trim() || undefined,
    });
  }

  return (
    <Modal
      open={open}
      onOpenChange={onOpenChange}
      title={specialization ? "Sửa chuyên khoa" : "Thêm chuyên khoa"}
      size="lg"
    >
      <form onSubmit={handleSubmit} className="space-y-4 pt-2">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <div>
            <Label htmlFor="spec-code">Mã</Label>
            <Input
              id="spec-code"
              required
              maxLength={50}
              placeholder="SPEC-001"
              value={code}
              onChange={(e) => setCode(e.target.value)}
            />
          </div>
          <div className="sm:col-span-2">
            <Label htmlFor="spec-name">Tên chuyên khoa</Label>
            <Input
              id="spec-name"
              required
              maxLength={255}
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
        </div>

        <div>
          <Label htmlFor="spec-slug">Slug</Label>
          <Input
            id="spec-slug"
            maxLength={255}
            placeholder="Để trống để tự sinh từ tên"
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
          />
        </div>

        <div>
          <Label htmlFor="spec-description">Mô tả</Label>
          <Textarea
            id="spec-description"
            rows={3}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
        </div>

        <div>
          <Label htmlFor="spec-symptoms">Triệu chứng</Label>
          <Input
            id="spec-symptoms"
            placeholder="Mất ngủ, lo âu, căng thẳng"
            value={symptoms}
            onChange={(e) => setSymptoms(e.target.value)}
          />
          <p className="mt-1.5 text-xs text-muted-foreground">Phân tách các triệu chứng bằng dấu phẩy.</p>
        </div>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <Label htmlFor="spec-location">Vị trí</Label>
            <Input
              id="spec-location"
              maxLength={255}
              value={location}
              onChange={(e) => setLocation(e.target.value)}
            />
          </div>
          <div>
            <Label htmlFor="spec-imageUrl">Ảnh (URL)</Label>
            <Input id="spec-imageUrl" value={imageUrl} onChange={(e) => setImageUrl(e.target.value)} />
          </div>
        </div>

        <div className="flex justify-end gap-2 border-t border-border pt-4">
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            Hủy
          </Button>
          <Button type="submit" disabled={isPending}>
            {isPending ? "Đang lưu..." : specialization ? "Lưu thay đổi" : "Thêm chuyên khoa"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
