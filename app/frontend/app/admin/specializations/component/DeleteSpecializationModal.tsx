"use client";

import { Modal, Button } from "@/components/ui";
import type { AdminSpecialization } from "@/types";

export interface DeleteSpecializationModalProps {
  specialization: AdminSpecialization | null;
  onOpenChange: (open: boolean) => void;
  isPending: boolean;
  onConfirm: () => void;
  onDeactivate: () => void;
}

/** Deleting is only allowed while no expert uses the specialization; otherwise offer deactivation. */
export function DeleteSpecializationModal({
  specialization,
  onOpenChange,
  isPending,
  onConfirm,
  onDeactivate,
}: DeleteSpecializationModalProps) {
  const inUse = (specialization?.expertCount ?? 0) > 0;

  return (
    <Modal open={specialization !== null} onOpenChange={onOpenChange} title="Xoá chuyên khoa" size="md">
      {specialization && (
        <div className="space-y-4 pt-2">
          {inUse ? (
            <p className="text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">{specialization.name}</span> đang được{" "}
              {specialization.expertCount} chuyên gia sử dụng nên không thể xoá. Hãy ngừng hoạt động chuyên khoa
              để ẩn nó khỏi bệnh nhân; các chuyên gia đã đăng ký vẫn giữ chuyên khoa này.
            </p>
          ) : (
            <p className="text-sm text-muted-foreground">
              Xoá vĩnh viễn chuyên khoa{" "}
              <span className="font-semibold text-foreground">{specialization.name}</span>? Thao tác này không thể
              hoàn tác.
            </p>
          )}

          <div className="flex justify-end gap-2 border-t border-border pt-4">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Hủy
            </Button>
            {inUse ? (
              specialization.isActive && (
                <Button type="button" onClick={onDeactivate} disabled={isPending}>
                  Ngừng hoạt động
                </Button>
              )
            ) : (
              <Button type="button" variant="danger" onClick={onConfirm} disabled={isPending}>
                {isPending ? "Đang xoá..." : "Xoá"}
              </Button>
            )}
          </div>
        </div>
      )}
    </Modal>
  );
}
