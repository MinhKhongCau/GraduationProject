"use client";

import { useState } from "react";
import { Plus, Trash2, Layers, ChevronDown, ChevronUp } from "lucide-react";
import { Card, Button, Spinner, Modal, Label, PageHeader, Input } from "@/components/ui";
import { useApiQuery, useApiMutation } from "@/hooks";
import { assessmentApi } from "@/api";
import { QUERY_KEYS } from "@/constants";
import { useErrorContext } from "@/context/ErrorContext";

interface OptionFormState {
  label: string;
  value: string;
  scoreValue: number;
  orderIndex: number;
}

export default function AdminOptionGroupsPage() {
  const { showSuccess, showError } = useErrorContext();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [expandedGroups, setExpandedGroups] = useState<Record<string, boolean>>({});

  // Form states
  const [groupCode, setGroupCode] = useState("");
  const [groupName, setGroupName] = useState("");
  const [description, setDescription] = useState("");
  const [options, setOptions] = useState<OptionFormState[]>([
    { label: "Không ngày nào", value: "0", scoreValue: 0, orderIndex: 1 },
    { label: "Vài ngày", value: "1", scoreValue: 1, orderIndex: 2 },
  ]);

  const { data: groups = [], isLoading, refetch } = useApiQuery({
    queryKey: QUERY_KEYS.assessmentOptionGroups(),
    queryFn: () => assessmentApi.getOptionGroups(),
  });

  const createMutation = useApiMutation({
    mutationFn: (payload: any) => assessmentApi.createOptionGroup(payload),
    onSuccess: () => {
      showSuccess("Tạo nhóm phương án thành công!");
      setIsCreateOpen(false);
      resetForm();
      refetch();
    },
    onError: (err: any) => {
      showError({ message: err.message || "Lỗi tạo nhóm phương án" });
    },
  });

  const deleteMutation = useApiMutation({
    mutationFn: (slug: string) => assessmentApi.deleteOptionGroup(slug),
    onSuccess: () => {
      showSuccess("Xóa nhóm phương án thành công!");
      refetch();
    },
    onError: (err: any) => {
      showError({ message: err.message || "Lỗi xóa nhóm phương án" });
    },
  });

  function resetForm() {
    setGroupCode("");
    setGroupName("");
    setDescription("");
    setOptions([
      { label: "Không ngày nào", value: "0", scoreValue: 0, orderIndex: 1 },
      { label: "Vài ngày", value: "1", scoreValue: 1, orderIndex: 2 },
    ]);
  }

  function toggleExpand(slug: string) {
    setExpandedGroups((current) => ({ ...current, [slug]: !current[slug] }));
  }

  function handleAddOption() {
    setOptions((current) => [
      ...current,
      { label: "", value: String(current.length), scoreValue: current.length, orderIndex: current.length + 1 },
    ]);
  }

  function handleRemoveOption(index: number) {
    setOptions((current) => current.filter((_, i) => i !== index));
  }

  function handleOptionChange(index: number, field: keyof OptionFormState, val: any) {
    setOptions((current) =>
      current.map((opt, i) => (i === index ? { ...opt, [field]: val } : opt))
    );
  }

  function handleCreateSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!groupCode.trim() || !groupName.trim() || options.length === 0) return;
    createMutation.mutate({ groupCode, groupName, description, options });
  }

  function handleDelete(slug: string) {
    if (confirm("Bạn có chắc chắn muốn xóa nhóm phương án này cùng tất cả tùy chọn?")) {
      deleteMutation.mutate(slug);
    }
  }

  return (
    <div className="mx-auto max-w-6xl space-y-6">
      <PageHeader
        className="mb-0"
        title="Quản lý Nhóm Phương Án (Option Groups)"
        description="Thiết lập danh sách và điểm số cho các phương án lựa chọn trong câu hỏi."
        actions={
          <Button onClick={() => { resetForm(); setIsCreateOpen(true); }}>
            <Plus className="h-4 w-4" /> Tạo nhóm mới
          </Button>
        }
      />

      {isLoading ? (
        <div className="flex justify-center py-20">
          <Spinner className="h-8 w-8" />
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4">
          {groups.map((group) => {
            const isExpanded = expandedGroups[group.slug];
            return (
              <Card key={group.slug} className="p-5">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-3">
                    <div className="rounded-lg bg-primary-soft p-2 text-primary-soft-text">
                      <Layers className="h-5 w-5" />
                    </div>
                    <div>
                      <h3 className="font-bold text-foreground">{group.groupName}</h3>
                      <div className="mt-1 flex flex-wrap gap-2 text-xs">
                        <span className="rounded-md bg-primary-soft px-1.5 py-0.5 font-mono text-primary-soft-text">
                          Code: {group.groupCode}
                        </span>
                        <span className="font-mono text-muted-foreground">
                          Slug: {group.slug}
                        </span>
                        <span className="text-muted-foreground">
                          ({group.options.length} phương án)
                        </span>
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => toggleExpand(group.slug)}
                    >
                      {isExpanded ? (
                        <>Thu gọn <ChevronUp className="h-3.5 w-3.5" /></>
                      ) : (
                        <>Chi tiết <ChevronDown className="h-3.5 w-3.5" /></>
                      )}
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleDelete(group.slug)}
                      className="border-danger/30 text-danger hover:bg-danger-soft"
                    >
                      <Trash2 className="h-3.5 w-3.5" /> Xóa
                    </Button>
                  </div>
                </div>

                {group.description && (
                  <p className="mt-3 text-sm text-muted-foreground pl-12">
                    {group.description}
                  </p>
                )}

                {isExpanded && (
                  <div className="mt-4 border-t border-border pt-4 pl-12">
                    <h4 className="mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                      Danh sách phương án trả lời:
                    </h4>
                    <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-4">
                      {group.options
                        .slice()
                        .sort((a: any, b: any) => a.orderIndex - b.orderIndex)
                        .map((opt: any) => (
                          <div
                            key={opt.slug}
                            className="flex flex-col rounded-xl border border-border bg-surface p-3"
                          >
                            <span className="text-xs font-semibold text-muted-foreground">
                              Thứ tự {opt.orderIndex}
                            </span>
                            <span className="mt-1 font-medium text-foreground">
                              {opt.label}
                            </span>
                            <div className="mt-2 flex items-center justify-between text-xs border-t border-border pt-1.5 font-mono">
                              <span className="text-muted-foreground">Value: "{opt.value}"</span>
                              <span className="font-bold text-primary">Score: {opt.scoreValue}</span>
                            </div>
                          </div>
                        ))}
                    </div>
                  </div>
                )}
              </Card>
            );
          })}

          {groups.length === 0 && (
            <Card className="p-8 text-center text-muted-foreground">
              Không tìm thấy nhóm phương án nào. Hãy tạo một cái!
            </Card>
          )}
        </div>
      )}

      {/* Create Modal */}
      <Modal open={isCreateOpen} onOpenChange={setIsCreateOpen} title="Tạo nhóm phương án mới">
        <form onSubmit={handleCreateSubmit} className="space-y-4 pt-2">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <Label htmlFor="option-groups-groupCode">Mã Code (e.g. GAD7_FREQ)</Label>
              <Input
                id="option-groups-groupCode"
                type="text"
                required
                placeholder="Nhập mã nhóm..."
                value={groupCode}
                onChange={(e) => setGroupCode(e.target.value)}
              />
            </div>
            <div>
              <Label htmlFor="option-groups-groupName">Tên nhóm</Label>
              <Input
                id="option-groups-groupName"
                type="text"
                required
                placeholder="Tên hiển thị nhóm phương án..."
                value={groupName}
                onChange={(e) => setGroupName(e.target.value)}
              />
            </div>
          </div>
          <div>
            <Label htmlFor="option-groups-description">Mô tả</Label>
            <Input
              id="option-groups-description"
              type="text"
              placeholder="Nhập mô tả ngắn..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>

          <div className="border-t border-border pt-4">
            <div className="mb-3 flex items-center justify-between">
              <Label className="mb-0 font-semibold">Các tùy chọn trả lời</Label>
              <Button type="button" variant="outline" size="sm" onClick={handleAddOption}>
                <Plus className="h-3.5 w-3.5" /> Thêm tùy chọn
              </Button>
            </div>

            <div className="max-h-60 space-y-2 overflow-y-auto pr-1">
              {options.map((opt, index) => (
                <div key={index} className="flex flex-col gap-2 rounded-xl border border-border bg-surface/50 p-3 sm:flex-row sm:items-center">
                  <div className="flex-1">
                    <Input
                      type="text"
                      required
                      placeholder="Nhãn phương án (e.g. Rất nhiều)"
                      value={opt.label}
                      onChange={(e) => handleOptionChange(index, "label", e.target.value)}
                      className="h-9 text-xs"
                    />
                  </div>
                  <div className="w-20">
                    <Input
                      type="text"
                      required
                      placeholder="Value"
                      value={opt.value}
                      onChange={(e) => handleOptionChange(index, "value", e.target.value)}
                      className="h-9 text-center text-xs"
                    />
                  </div>
                  <div className="w-20">
                    <Input
                      type="number"
                      required
                      min={0}
                      placeholder="Score"
                      value={opt.scoreValue}
                      onChange={(e) => handleOptionChange(index, "scoreValue", parseInt(e.target.value) || 0)}
                      className="h-9 text-center text-xs"
                    />
                  </div>
                  <div className="w-16">
                    <Input
                      type="number"
                      required
                      min={1}
                      placeholder="Order"
                      value={opt.orderIndex}
                      onChange={(e) => handleOptionChange(index, "orderIndex", parseInt(e.target.value) || 1)}
                      className="h-9 text-center text-xs"
                    />
                  </div>
                  <button
                    type="button"
                    onClick={() => handleRemoveOption(index)}
                    className="inline-flex h-9 w-9 shrink-0 items-center justify-center self-end rounded-lg text-danger transition-colors hover:bg-danger-soft sm:self-auto"
                    aria-label="Remove option"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              ))}
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-2 border-t border-border">
            <Button type="button" variant="outline" onClick={() => setIsCreateOpen(false)}>
              Hủy
            </Button>
            <Button type="submit" disabled={createMutation.isPending || options.length === 0}>
              {createMutation.isPending ? "Đang lưu..." : "Tạo mới"}
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
