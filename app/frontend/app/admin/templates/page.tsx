"use client";

import { useState } from "react";
import dynamic from "next/dynamic";
import { Plus, Edit2, Trash2, ShieldAlert } from "lucide-react";
import { Card, Button, Spinner, Modal } from "@/components/ui";
import { useApiQuery, useApiMutation } from "@/hooks";
import { assessmentApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useErrorContext } from "@/context/ErrorContext";
import type { AssessmentTemplate } from "@/types";

const MDXEditor = dynamic(() => import("@/components/ui/MDXEditor"), {
  ssr: false,
});

export default function AdminTemplatesPage() {
  const { showSuccess, showError } = useErrorContext();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isEditOpen, setIsEditOpen] = useState(false);
  const [selectedTemplate, setSelectedTemplate] = useState<AssessmentTemplate | null>(null);

  // Form states
  const [code, setCode] = useState("");
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [instruction, setInstruction] = useState("");
  const [certification, setCertification] = useState("");

  const countWords = (text: string) => {
    return text.trim() === "" ? 0 : text.trim().split(/\s+/).length;
  };

  const isWordCountInvalid = countWords(instruction) > 3000 || countWords(certification) > 3000;

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
      resetForm();
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
      resetForm();
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

  function resetForm() {
    setCode("");
    setTitle("");
    setDescription("");
    setInstruction("");
    setCertification("");
  }

  function handleOpenCreate() {
    resetForm();
    setIsCreateOpen(true);
  }

  function handleOpenEdit(template: AssessmentTemplate) {
    setSelectedTemplate(template);
    setCode(template.code);
    setTitle(template.title);
    setDescription(template.description || "");
    setInstruction(template.instruction || "");
    setCertification(template.certification || "");
    setIsEditOpen(true);
  }

  function handleCreateSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!code.trim() || !title.trim() || isWordCountInvalid) return;
    createMutation.mutate({ code, title, description, instruction, certification });
  }

  function handleEditSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!selectedTemplate || !code.trim() || !title.trim() || isWordCountInvalid) return;
    updateMutation.mutate({
      slug: selectedTemplate.slug,
      payload: { code, title, description, instruction, certification },
    });
  }

  function handleDelete(slug: string) {
    if (confirm("Bạn có chắc chắn muốn xóa bài test này?")) {
      deleteMutation.mutate(slug);
    }
  }

  return (
    <div className="mx-auto max-w-6xl space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Quản lý bài test (Templates)</h1>
          <p className="text-sm text-muted-foreground">
            Danh sách và thông tin các bài đánh giá tâm lý.
          </p>
        </div>
        <Button onClick={handleOpenCreate} className="flex items-center gap-1.5 self-start sm:self-auto">
          <Plus className="h-4 w-4" /> Tạo bài test mới
        </Button>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-20">
          <Spinner className="h-8 w-8" />
        </div>
      ) : (
        <Card className="overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full border-collapse text-left text-sm">
              <thead className="border-b border-border bg-surface text-xs font-semibold uppercase text-muted-foreground">
                <tr>
                  <th className="px-6 py-4">Mã Code</th>
                  <th className="px-6 py-4">Tên bài test</th>
                  <th className="px-6 py-4">Mô tả</th>
                  <th className="px-6 py-4">Slug</th>
                  <th className="px-6 py-4">Trạng thái</th>
                  <th className="px-6 py-4 text-right">Thao tác</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border bg-background">
                {templates.filter((t: any) => t.isActive !== false).map((template) => (
                  <tr key={template.slug} className="hover:bg-surface/30">
                    <td className="px-6 py-4 font-mono text-xs font-bold text-primary">
                      {template.code}
                    </td>
                    <td className="px-6 py-4 font-medium text-foreground">
                      {template.title}
                    </td>
                    <td className="max-w-xs truncate px-6 py-4 text-muted-foreground">
                      {template.description || "—"}
                    </td>
                    <td className="px-6 py-4 font-mono text-xs text-muted-foreground">
                      {template.slug}
                    </td>
                    <td className="px-6 py-4">
                      <span className="inline-flex rounded-full bg-success-soft px-2 py-1 text-xs font-semibold text-success">
                        Active
                      </span>
                    </td>
                    <td className="px-6 py-4 text-right">
                      <div className="flex justify-end gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleOpenEdit(template)}
                          className="flex items-center gap-1 px-2.5 py-1 text-xs"
                        >
                          <Edit2 className="h-3 w-3" /> Sửa
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleDelete(template.slug)}
                          className="flex items-center gap-1 border-danger/20 px-2.5 py-1 text-xs text-danger hover:bg-danger-soft"
                        >
                          <Trash2 className="h-3 w-3" /> Xóa
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
      <Modal open={isCreateOpen} onOpenChange={setIsCreateOpen} title="Tạo bài test mới" size="3xl">
        <form onSubmit={handleCreateSubmit} className="space-y-4 pt-2">
          <div>
            <label className="mb-1 block text-sm font-medium text-foreground">Mã Code (e.g. PHQ_9)</label>
            <input
              type="text"
              required
              placeholder="Nhập mã code viết hoa..."
              value={code}
              onChange={(e) => setCode(e.target.value)}
              className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:border-primary"
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-foreground">Tên bài test</label>
            <input
              type="text"
              required
              placeholder="Nhập tên tiêu đề bài test..."
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:border-primary"
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-foreground">Mô tả chi tiết</label>
            <textarea
              rows={3}
              placeholder="Mô tả công dụng và hướng dẫn bài test..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:border-primary"
            />
          </div>
          <div>
            <div className="flex justify-between items-center mb-1">
              <label className="text-sm font-medium text-foreground">Hướng dẫn (Instruction - Markdown)</label>
              <span className={`text-xs ${countWords(instruction) > 3000 ? "text-danger font-bold" : "text-muted-foreground"}`}>
                {countWords(instruction)}/3000 từ
              </span>
            </div>
            <MDXEditor
              value={instruction}
              onChange={setInstruction}
              placeholder="Hướng dẫn thực hiện bài test..."
            />
          </div>
          <div>
            <div className="flex justify-between items-center mb-1">
              <label className="text-sm font-medium text-foreground">Chứng nhận (Certification - Markdown)</label>
              <span className={`text-xs ${countWords(certification) > 3000 ? "text-danger font-bold" : "text-muted-foreground"}`}>
                {countWords(certification)}/3000 từ
              </span>
            </div>
            <MDXEditor
              value={certification}
              onChange={setCertification}
              placeholder="Thông tin chứng nhận chuyên môn..."
            />
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="outline" onClick={() => setIsCreateOpen(false)}>
              Hủy
            </Button>
            <Button type="submit" disabled={createMutation.isPending || isWordCountInvalid}>
              {createMutation.isPending ? "Đang lưu..." : "Tạo mới"}
            </Button>
          </div>
        </form>
      </Modal>

      {/* Edit Modal */}
      <Modal open={isEditOpen} onOpenChange={setIsEditOpen} title="Chỉnh sửa bài test" size="3xl">
        <form onSubmit={handleEditSubmit} className="space-y-4 pt-2">
          <div>
            <label className="mb-1 block text-sm font-medium text-foreground">Mã Code</label>
            <input
              type="text"
              required
              value={code}
              onChange={(e) => setCode(e.target.value)}
              className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:border-primary"
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-foreground">Tên bài test</label>
            <input
              type="text"
              required
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:border-primary"
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-foreground">Mô tả</label>
            <textarea
              rows={3}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:border-primary"
            />
          </div>
          <div>
            <div className="flex justify-between items-center mb-1">
              <label className="text-sm font-medium text-foreground">Hướng dẫn (Instruction - Markdown)</label>
              <span className={`text-xs ${countWords(instruction) > 3000 ? "text-danger font-bold" : "text-muted-foreground"}`}>
                {countWords(instruction)}/3000 từ
              </span>
            </div>
            <MDXEditor
              value={instruction}
              onChange={setInstruction}
              placeholder="Hướng dẫn thực hiện bài test..."
            />
          </div>
          <div>
            <div className="flex justify-between items-center mb-1">
              <label className="text-sm font-medium text-foreground">Chứng nhận (Certification - Markdown)</label>
              <span className={`text-xs ${countWords(certification) > 3000 ? "text-danger font-bold" : "text-muted-foreground"}`}>
                {countWords(certification)}/3000 từ
              </span>
            </div>
            <MDXEditor
              value={certification}
              onChange={setCertification}
              placeholder="Thông tin chứng nhận chuyên môn..."
            />
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="outline" onClick={() => setIsEditOpen(false)}>
              Hủy
            </Button>
            <Button type="submit" disabled={updateMutation.isPending || isWordCountInvalid}>
              {updateMutation.isPending ? "Đang lưu..." : "Cập nhật"}
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
