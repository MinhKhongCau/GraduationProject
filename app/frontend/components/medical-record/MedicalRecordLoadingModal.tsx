"use client";

import { Modal, Spinner } from "@/components/ui";

/** Shown while a medical record opened from a list row is still loading (or could not be found). */
export function MedicalRecordLoadingModal({
  open,
  onOpenChange,
  isLoading,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  isLoading: boolean;
}) {
  return (
    <Modal open={open} onOpenChange={onOpenChange} title="Chi tiết Hồ sơ Bệnh án" size="2xl">
      {isLoading ? (
        <div className="flex justify-center py-10">
          <Spinner className="h-6 w-6" />
        </div>
      ) : (
        <p className="py-6 text-center text-sm text-muted-foreground">Không tải được hồ sơ bệnh án.</p>
      )}
    </Modal>
  );
}
