"use client";

import { useState } from "react";
import { Plus, Edit2, Trash2, ShieldAlert } from "lucide-react";
import { Card, Button, Spinner, PageHeader, Badge } from "@/components/ui";
import { useApiQuery, useApiMutation } from "@/hooks";
import { assessmentApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useErrorContext } from "@/context/ErrorContext";
import type { AssessmentTemplate } from "@/types";
import { TemplateModal } from "./component/TemplateModal";

export default function AdminTemplatesPage() {
  const { showSuccess, showError } = useErrorContext();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isEditOpen, setIsEditOpen] = useState(false);
  const [selectedTemplate, setSelectedTemplate] = useState<AssessmentTemplate | null>(null);

  const { data: templates = [], isLoading, refetch } = useApiQuery({
    queryKey: QUERY_KEYS.assessmentTemplates(),
    queryFn: () => assessmentApi.getTemplates(),
  });

  const createMutation = useApiMutation({
    mutationFn: (payload: { code: string; title: string; description?: string; instruction?: string; certification?: string }) =>
      assessmentApi.createTemplate(payload),
    onSuccess: () => {
      showSuccess("Tạo bài test thành công!");
      setIsCreateOpen(false);
      refetch();
    },
    onError: (err: any) => {
      showError({ message: err.message || "Lỗi tạo bài test" });
    },
  });

  const updateMutation = useApiMutation({
    mutationFn: ({ slug, payload }: { slug: string; payload: any }) =>
      assessmentApi.updateTemplate(slug, payload),
    onSuccess: () => {
      showSuccess("Cập nhật bài test thành công!");
      setIsEditOpen(false);
      setSelectedTemplate(null);
      refetch();
    },
    onError: (err: any) => {
      showError({ message: err.message || "Lỗi cập nhật bài test" });
    },
  });

  const deleteMutation = useApiMutation({
    mutationFn: (slug: string) => assessmentApi.deleteTemplate(slug),
    onSuccess: () => {
      showSuccess("Xóa bài test thành công!");
      refetch();
    },
    onError: (err: any) => {
      showError({ message: err.message || "Lỗi xóa bài test" });
    },
  });

  function handleOpenCreate() {
    setIsCreateOpen(true);
  }

  function handleOpenEdit(template: AssessmentTemplate) {
    setSelectedTemplate(template);
    setIsEditOpen(true);
  }

  function handleCreateSubmit(values: {
    code: string;
    title: string;
    description: string;
    instruction: string;
    certification: string;
  }) {
    createMutation.mutate(values);
  }

  function handleEditSubmit(values: {
    code: string;
    title: string;
    description: string;
    instruction: string;
    certification: string;
  }) {
    if (!selectedTemplate) return;
    updateMutation.mutate({
      slug: selectedTemplate.slug,
      payload: values,
    });
  }

  function handleDelete(slug: string) {
    if (confirm("Bạn có chắc chắn muốn xóa bài test này?")) {
      deleteMutation.mutate(slug);
    }
  }

  return (
    <div className="mx-auto max-w-6xl space-y-6">
      <PageHeader
        className="mb-0"
        title="Quản lý bài test (Templates)"
        description="Danh sách và thông tin các bài đánh giá tâm lý."
        actions={
          <Button onClick={handleOpenCreate}>
            <Plus className="h-4 w-4" /> Tạo bài test mới
          </Button>
        }
      />

      {isLoading ? (
        <div className="flex justify-center py-20">
          <Spinner className="h-8 w-8" />
        </div>
      ) : (
        <Card className="overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full border-collapse text-left text-sm">
              <thead className="border-b border-border bg-surface text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                <tr>
                  <th className="px-4 py-3">Mã Code</th>
                  <th className="px-4 py-3">Tên bài test</th>
                  <th className="px-4 py-3">Mô tả</th>
                  <th className="px-4 py-3">Slug</th>
                  <th className="px-4 py-3">Trạng thái</th>
                  <th className="px-4 py-3 text-right">Thao tác</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border bg-background">
                {templates.filter((t: any) => t.isActive !== false).map((template) => (
                  <tr key={template.slug} className="transition-colors hover:bg-surface/60">
                    <td className="px-4 py-3 font-mono text-xs font-bold text-primary">
                      {template.code}
                    </td>
                    <td className="px-4 py-3 font-medium text-foreground">
                      {template.title}
                    </td>
                    <td className="max-w-xs truncate px-4 py-3 text-muted-foreground">
                      {template.description || "—"}
                    </td>
                    <td className="px-4 py-3 text-muted-foreground font-mono text-xs">
                      {template.slug}
                    </td>
                    <td className="px-4 py-3">
                      <Badge tone="success">Hoạt động</Badge>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-end gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleOpenEdit(template)}
                          className="text-primary hover:bg-primary-soft"
                        >
                          <Edit2 className="h-3.5 w-3.5" /> Sửa
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDelete(template.slug)}
                          className="text-danger hover:bg-danger-soft"
                        >
                          <Trash2 className="h-3.5 w-3.5" /> Xóa
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
                {templates.filter((t: any) => t.isActive !== false).length === 0 && (
                  <tr>
                    <td colSpan={6} className="px-6 py-10 text-center text-muted-foreground">
                      Không tìm thấy bài test nào. Hãy tạo một cái!
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {/* Create Modal */}
      <TemplateModal
        open={isCreateOpen}
        onOpenChange={setIsCreateOpen}
        title="Tạo bài test mới"
        submitLabel="Tạo mới"
        isPending={createMutation.isPending}
        onSubmit={handleCreateSubmit}
      />

      {/* Edit Modal */}
      <TemplateModal
        open={isEditOpen}
        onOpenChange={setIsEditOpen}
        title="Chỉnh sửa bài test"
        submitLabel="Cập nhật"
        isPending={updateMutation.isPending}
        initialValues={
          selectedTemplate
            ? {
                code: selectedTemplate.code,
                title: selectedTemplate.title,
                description: selectedTemplate.description || "",
                instruction: selectedTemplate.instruction || "",
                certification: selectedTemplate.certification || "",
              }
            : undefined
        }
        onSubmit={handleEditSubmit}
      />
    </div>
  );
}
