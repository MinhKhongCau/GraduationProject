"use client";

import { useState } from "react";
import { Plus, Trash2, Layers } from "lucide-react";
import { Card, Button, Spinner, Modal } from "@/components/ui";
import { useApiQuery, useApiMutation } from "@/hooks";
import { assessmentApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useErrorContext } from "@/context/ErrorContext";

export default function AdminDimensionsPage() {
  const { showSuccess, showError } = useErrorContext();
  const [isCreateOpen, setIsCreateOpen] = useState(false);

  const [code, setCode] = useState("");
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");

  const { data: dimensions = [], isLoading, refetch } = useApiQuery({
    queryKey: QUERY_KEYS.assessmentDimensions(),
    queryFn: () => assessmentApi.getDimensions(),
  });

  const createMutation = useApiMutation({
    mutationFn: (payload: any) => assessmentApi.createDimension(payload),
    onSuccess: () => {
      showSuccess("Tạo khía cạnh (dimension) thành công!");
      setIsCreateOpen(false);
      resetForm();
      refetch();
    },
    onError: (err: any) => {
      showError({ message: err.message || "Lỗi tạo khía cạnh (dimension)" });
    },
  });

  const deleteMutation = useApiMutation({
    mutationFn: (slug: string) => assessmentApi.deleteDimension(slug),
    onSuccess: () => {
      showSuccess("Xóa khía cạnh (dimension) thành công!");
      refetch();
    },
    onError: (err: any) => {
      showError({ message: err.message || "Lỗi xóa khía cạnh (dimension)" });
    },
  });

  function resetForm() {
    setCode("");
    setName("");
    setDescription("");
  }

  function handleCreateSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!code.trim() || !name.trim()) return;
    createMutation.mutate({ code, name, description });
  }

  function handleDelete(slug: string) {
    if (confirm("Bạn có chắc chắn muốn xóa khía cạnh (dimension) này?")) {
      deleteMutation.mutate(slug);
    }
  }

  return (
    <div className="mx-auto max-w-5xl space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Quản lý Khía Cạnh (Dimensions)</h1>
          <p className="text-sm text-muted-foreground">
            Thiết lập các khía cạnh (thang đo) để gán cho câu hỏi trong bài test, ví dụ: Trầm cảm, Lo âu, Căng thẳng.
          </p>
        </div>
        <Button onClick={() => { resetForm(); setIsCreateOpen(true); }} className="flex items-center gap-1.5 self-start sm:self-auto">
          <Plus className="h-4 w-4" /> Tạo khía cạnh mới
        </Button>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-20">
          <Spinner className="h-8 w-8" />
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {dimensions.map((dimension) => (
            <Card key={dimension.slug} className="p-5">
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <div className="rounded-lg bg-primary-soft p-2 text-primary">
                    <Layers className="h-5 w-5" />
                  </div>
                  <div>
                    <h3 className="font-bold text-foreground">{dimension.name}</h3>
                    <span className="font-mono text-xs text-primary bg-primary-soft/40 px-1.5 py-0.5 rounded">
                      {dimension.code}
                    </span>
                  </div>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleDelete(dimension.slug)}
                  className="flex items-center gap-1 border-danger/20 px-2.5 py-1 text-xs text-danger hover:bg-danger-soft shrink-0"
                >
                  <Trash2 className="h-3 w-3" />
                </Button>
              </div>
              {dimension.description && (
                <p className="mt-3 text-sm text-muted-foreground">{dimension.description}</p>
              )}
            </Card>
          ))}

          {dimensions.length === 0 && (
            <Card className="p-8 text-center text-muted-foreground sm:col-span-2 lg:col-span-3">
              Chưa có khía cạnh (dimension) nào. Hãy tạo một cái để gán cho câu hỏi!
            </Card>
          )}
        </div>
      )}

      {/* Create Modal */}
      <Modal open={isCreateOpen} onOpenChange={setIsCreateOpen} title="Tạo khía cạnh mới">
        <form onSubmit={handleCreateSubmit} className="space-y-4 pt-2">
          <div>
            <label className="mb-1 block text-sm font-medium text-foreground">Mã Code (e.g. DEPRESSION)</label>
            <input
              type="text"
              required
              placeholder="Nhập mã khía cạnh..."
              value={code}
              onChange={(e) => setCode(e.target.value)}
              className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:border-primary"
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-foreground">Tên hiển thị</label>
            <input
              type="text"
              required
              placeholder="e.g. Trầm cảm"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:border-primary"
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-foreground">Mô tả</label>
            <input
              type="text"
              placeholder="Nhập mô tả ngắn..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:border-primary"
            />
          </div>

          <div className="flex justify-end gap-2 pt-2 border-t border-border">
            <Button type="button" variant="outline" onClick={() => setIsCreateOpen(false)}>
              Hủy
            </Button>
            <Button type="submit" disabled={createMutation.isPending}>
              {createMutation.isPending ? "Đang lưu..." : "Tạo mới"}
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
