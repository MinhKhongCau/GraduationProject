"use client";

import { useState } from "react";
import Link from "next/link";
import { Plus, Trash2, HelpCircle, FileText, CheckSquare } from "lucide-react";
import { Card, Button, Spinner, Modal, Label, PageHeader, Input, Select, Badge } from "@/components/ui";
import { useApiQuery, useApiMutation } from "@/hooks";
import { assessmentApi } from "@/api";
import { QUERY_KEYS, ROUTES } from "@/constants";
import { useErrorContext } from "@/context/ErrorContext";

interface QuestionFormState {
  content: string;
  dimensionId: string; // holds dimension slug
  questionOrder: number;
}

export default function AdminQuestionsPage() {
  const { showSuccess, showError } = useErrorContext();
  const [selectedTemplateSlug, setSelectedTemplateSlug] = useState("");
  const [isBulkCreateOpen, setIsBulkCreateOpen] = useState(false);

  // Bulk form states
  const [bulkGroupId, setBulkGroupId] = useState("");
  const [bulkQuestions, setBulkQuestions] = useState<QuestionFormState[]>([
    { content: "", dimensionId: "", questionOrder: 1 },
  ]);

  // Fetch templates for selector
  const { data: templates = [], isLoading: isLoadingTemplates } = useApiQuery({
    queryKey: QUERY_KEYS.assessmentTemplates(),
    queryFn: () => assessmentApi.getTemplates(),
  });

  // Fetch option groups for selector
  const { data: groups = [], isLoading: isLoadingGroups } = useApiQuery({
    queryKey: QUERY_KEYS.assessmentOptionGroups(),
    queryFn: () => assessmentApi.getOptionGroups(),
  });

  // Fetch dimensions for selector — admin must create these first (see /admin/dimensions)
  const { data: dimensions = [] } = useApiQuery({
    queryKey: QUERY_KEYS.assessmentDimensions(),
    queryFn: () => assessmentApi.getDimensions(),
  });

  // Fetch questions of the selected template
  const { data: questions = [], isLoading: isLoadingQuestions, refetch: refetchQuestions } = useApiQuery({
    queryKey: QUERY_KEYS.assessmentQuestions(selectedTemplateSlug),
    queryFn: () => assessmentApi.getQuestionsByTemplate(selectedTemplateSlug),
    enabled: !!selectedTemplateSlug,
  });

  const createBulkMutation = useApiMutation({
    mutationFn: (payload: any) => assessmentApi.createBulkQuestions(payload),
    onSuccess: () => {
      showSuccess("Thêm danh sách câu hỏi thành công!");
      setIsBulkCreateOpen(false);
      resetBulkForm();
      refetchQuestions();
    },
    onError: (err: any) => {
      showError({ message: err.message || "Lỗi thêm câu hỏi" });
    },
  });

  const deleteMutation = useApiMutation({
    mutationFn: (slug: string) => assessmentApi.deleteQuestion(slug),
    onSuccess: () => {
      showSuccess("Xóa câu hỏi thành công!");
      refetchQuestions();
    },
    onError: (err: any) => {
      showError({ message: err.message || "Lỗi xóa câu hỏi" });
    },
  });

  function resetBulkForm() {
    setBulkGroupId("");
    setBulkQuestions([{ content: "", dimensionId: dimensions[0]?.slug ?? "", questionOrder: 1 }]);
  }

  function handleAddQuestionRow() {
    setBulkQuestions((current) => [
      ...current,
      { content: "", dimensionId: dimensions[0]?.slug ?? "", questionOrder: current.length + 1 },
    ]);
  }

  function handleRemoveQuestionRow(index: number) {
    setBulkQuestions((current) => current.filter((_, i) => i !== index));
  }

  function handleQuestionChange(index: number, field: keyof QuestionFormState, val: any) {
    setBulkQuestions((current) =>
      current.map((q, i) => (i === index ? { ...q, [field]: val } : q))
    );
  }

  function handleBulkSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!selectedTemplateSlug || !bulkGroupId || bulkQuestions.length === 0) return;
    createBulkMutation.mutate({
      templateId: selectedTemplateSlug,
      groupId: bulkGroupId,
      questions: bulkQuestions,
    });
  }

  function handleDelete(slug: string) {
    if (confirm("Bạn có chắc chắn muốn xóa câu hỏi này?")) {
      deleteMutation.mutate(slug);
    }
  }

  return (
    <div className="mx-auto max-w-6xl space-y-6">
      <PageHeader
        className="mb-0"
        title="Quản lý câu hỏi (Questions)"
        description="Thiết lập bộ câu hỏi và thang đo chuyên sâu cho từng bài test."
        actions={
          <Button
            disabled={!selectedTemplateSlug || dimensions.length === 0}
            onClick={() => { resetBulkForm(); setIsBulkCreateOpen(true); }}
          >
            <Plus className="h-4 w-4" /> Thêm câu hỏi (Bulk)
          </Button>
        }
      />

      {dimensions.length === 0 && (
        <Card className="p-4 text-sm text-muted-foreground">
          Chưa có khía cạnh (dimension) nào. Hãy tạo khía cạnh ở trang{" "}
          <Link href={ROUTES.ADMIN.DIMENSIONS} className="font-semibold text-primary hover:underline">
            Quản lý Khía Cạnh
          </Link>{" "}
          trước khi thêm câu hỏi.
        </Card>
      )}

      <Card className="p-5">
        <Label className="mb-2 font-semibold" htmlFor="questions-selectedTemplateSlug">Chọn bài test để quản lý câu hỏi</Label>
        <Select
          id="questions-selectedTemplateSlug"
          aria-label="Chọn bài test để quản lý câu hỏi"
          value={selectedTemplateSlug}
          onChange={(e) => setSelectedTemplateSlug(e.target.value)}
          className="max-w-md"
        >
          <option value="">-- Hãy chọn một bài đánh giá --</option>
          {templates.map((t) => (
            <option key={t.slug} value={t.slug}>
              {t.title} ({t.code})
            </option>
          ))}
        </Select>
      </Card>

      {!selectedTemplateSlug ? (
        <Card className="p-8 text-center text-muted-foreground">
          Hãy chọn một bài đánh giá ở trên để xem danh sách câu hỏi.
        </Card>
      ) : isLoadingQuestions ? (
        <div className="flex justify-center py-20">
          <Spinner className="h-8 w-8" />
        </div>
      ) : (
        <div className="space-y-4">
          <h2 className="text-lg font-bold text-foreground">
            Danh sách câu hỏi ({questions.length} câu)
          </h2>
          <div className="grid grid-cols-1 gap-4">
            {questions.map((question) => (
              <Card key={question.slug} className="p-5">
                <div className="flex items-start justify-between">
                  <div className="flex items-start gap-3">
                    <div className="mt-0.5 shrink-0 rounded-lg bg-primary-soft px-2 py-1 font-mono text-xs font-bold text-primary-soft-text">
                      Q{question.questionOrder}
                    </div>
                    <div>
                      <p className="font-medium text-foreground text-base">
                        {question.content}
                      </p>
                      <div className="mt-2 flex flex-wrap gap-2 text-xs">
                        <span className="rounded-md bg-surface px-1.5 py-0.5 font-mono text-muted-foreground">
                          Slug: {question.slug}
                        </span>
                        {question.dimension && (
                          <Badge tone="success" className="uppercase">
                            Thang đo: {question.dimension}
                          </Badge>
                        )}
                      </div>
                    </div>
                  </div>

                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handleDelete(question.slug)}
                    className="border-danger/30 text-danger hover:bg-danger-soft"
                  >
                    <Trash2 className="h-3.5 w-3.5" /> Xóa
                  </Button>
                </div>

                <div className="mt-4 border-t border-border pt-3">
                  <span className="text-xs font-semibold text-muted-foreground">
                    Các phương án trả lời tương ứng:
                  </span>
                  <div className="mt-2 flex flex-wrap gap-2">
                    {question.options.map((opt) => (
                      <span
                        key={opt.slug}
                        className="inline-flex items-center gap-1 rounded-lg border border-border bg-surface px-2 py-1 text-xs font-medium text-foreground"
                      >
                        {opt.label}{" "}
                        <span className="font-bold text-primary">({opt.scoreValue}đ)</span>
                      </span>
                    ))}
                  </div>
                </div>
              </Card>
            ))}

            {questions.length === 0 && (
              <Card className="p-8 text-center text-muted-foreground">
                Bài đánh giá này chưa có câu hỏi nào. Hãy thêm câu hỏi mới!
              </Card>
            )}
          </div>
        </div>
      )}

      {/* Bulk Create Modal */}
      <Modal open={isBulkCreateOpen} onOpenChange={setIsBulkCreateOpen} title="Thêm danh sách câu hỏi (Bulk)">
        <form onSubmit={handleBulkSubmit} className="space-y-4 pt-2">
          <div>
            <Label htmlFor="questions-bulkGroupId">Chọn nhóm phương án trả lời chung</Label>
            <Select
              id="questions-bulkGroupId"
              required
              value={bulkGroupId}
              onChange={(e) => setBulkGroupId(e.target.value)}
            >
              <option value="">-- Chọn nhóm phương án --</option>
              {groups.map((g) => (
                <option key={g.slug} value={g.slug}>
                  {g.groupName} ({g.groupCode})
                </option>
              ))}
            </Select>
          </div>

          <div className="border-t border-border pt-4">
            <div className="mb-3 flex items-center justify-between">
              <Label className="mb-0 font-semibold">Danh sách câu hỏi</Label>
              <Button type="button" variant="outline" size="sm" onClick={handleAddQuestionRow}>
                <Plus className="h-3.5 w-3.5" /> Thêm câu
              </Button>
            </div>

            <div className="max-h-80 space-y-3 overflow-y-auto pr-1">
              {bulkQuestions.map((q, index) => (
                <div key={index} className="flex flex-col gap-2 rounded-xl border border-border bg-surface/50 p-3 sm:flex-row sm:items-start">
                  <div className="w-12 text-center text-xs font-mono font-bold text-muted-foreground pt-3 shrink-0">
                    #{index + 1}
                  </div>
                  <div className="flex-1">
                    <Input
                      type="text"
                      required
                      placeholder="Nội dung câu hỏi..."
                      value={q.content}
                      onChange={(e) => handleQuestionChange(index, "content", e.target.value)}
                      className="h-9 text-xs"
                    />
                  </div>
                  <div className="w-36 shrink-0">
                    <Select
                      required
                      value={q.dimensionId}
                      onChange={(e) => handleQuestionChange(index, "dimensionId", e.target.value)}
                      className="h-9 text-xs"
                    >
                      <option value="">-- Khía cạnh --</option>
                      {dimensions.map((d) => (
                        <option key={d.slug} value={d.slug}>
                          {d.name} ({d.code})
                        </option>
                      ))}
                    </Select>
                  </div>
                  <div className="w-16 shrink-0">
                    <Input
                      type="number"
                      required
                      min={1}
                      placeholder="Thứ tự"
                      value={q.questionOrder}
                      onChange={(e) => handleQuestionChange(index, "questionOrder", parseInt(e.target.value) || 1)}
                      className="h-9 text-center text-xs"
                    />
                  </div>
                  <button
                    type="button"
                    onClick={() => handleRemoveQuestionRow(index)}
                    disabled={bulkQuestions.length === 1}
                    className="inline-flex h-9 w-9 shrink-0 items-center justify-center self-end rounded-lg text-danger transition-colors hover:bg-danger-soft disabled:cursor-not-allowed disabled:opacity-40 sm:self-auto"
                    aria-label="Remove question"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              ))}
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-2 border-t border-border">
            <Button type="button" variant="outline" onClick={() => setIsBulkCreateOpen(false)}>
              Hủy
            </Button>
            <Button type="submit" disabled={createBulkMutation.isPending || !bulkGroupId || bulkQuestions.length === 0}>
              {createBulkMutation.isPending ? "Đang lưu..." : "Lưu danh sách"}
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
