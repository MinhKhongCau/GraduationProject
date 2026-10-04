"use client";

import { useState } from "react";
import { Plus, Trash2, Layers } from "lucide-react";
import { Card, Button, Spinner, Modal, Label, PageHeader, Input } from "@/components/ui";
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
      <PageHeader
        className="mb-0"
        title="Quản lý Khía Cạnh (Dimensions)"
        description="Thiết lập các khía cạnh (thang đo) để gán cho câu hỏi trong bài test, ví dụ: Trầm cảm, Lo âu, Căng thẳng."
        actions={
          <Button onClick={() => { resetForm(); setIsCreateOpen(true); }}>
            <Plus className="h-4 w-4" /> Tạo khía cạnh mới
          </Button>
        }
      />

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
                  <div className="rounded-lg bg-primary-soft p-2 text-primary-soft-text">
                    <Layers className="h-5 w-5" aria-hidden="true" />
                  </div>
                  <div>
                    <h3 className="font-bold text-foreground">{dimension.name}</h3>
                    <span className="rounded-md bg-primary-soft px-1.5 py-0.5 font-mono text-xs text-primary-soft-text">
                      {dimension.code}
                    </span>
                  </div>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  aria-label="Xóa khía cạnh"
                  onClick={() => handleDelete(dimension.slug)}
                  className="w-8 px-0 border-danger/30 text-danger hover:bg-danger-soft"
                >
                  <Trash2 className="h-3.5 w-3.5" />
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
            <Label htmlFor="dimensions-code">Mã Code (e.g. DEPRESSION)</Label>
            <Input
              id="dimensions-code"
              type="text"
              required
              placeholder="Nhập mã khía cạnh..."
              value={code}
              onChange={(e) => setCode(e.target.value)}
            />
          </div>
          <div>
            <Label htmlFor="dimensions-name">Tên hiển thị</Label>
            <Input
              id="dimensions-name"
              type="text"
              required
              placeholder="e.g. Trầm cảm"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div>
            <Label htmlFor="dimensions-description">Mô tả</Label>
            <Input
              id="dimensions-description"
              type="text"
              placeholder="Nhập mô tả ngắn..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
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
