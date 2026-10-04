"use client";

import { useState, useEffect } from "react";
import dynamic from "next/dynamic";
import { Modal, Button, Label, Input, Textarea, Badge } from "@/components/ui";
import { BookOpen, Award } from "lucide-react";

const MDXEditor = dynamic(() => import("@/components/ui/MDXEditor"), {
  ssr: false,
});

interface TemplateModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  submitLabel: string;
  isPending: boolean;
  initialValues?: {
    code: string;
    title: string;
    description: string;
    instruction: string;
    certification: string;
  };
  onSubmit: (values: {
    code: string;
    title: string;
    description: string;
    instruction: string;
    certification: string;
  }) => void;
}

// Custom Markdown preview parser matching the patient detail page structure
function PreviewMarkdownRenderer({ text }: { text: string }) {
  if (!text) return <p className="text-xs text-muted-foreground italic">Chưa nhập nội dung.</p>;
  
  const lines = text.split("\n");
  const processed = lines.map((line, idx) => {
    const trimmed = line.trim();
    if (trimmed.startsWith("# ")) {
      return <h1 key={idx} className="text-base font-bold text-foreground mt-3 mb-1">{trimmed.slice(2)}</h1>;
    }
    if (trimmed.startsWith("## ")) {
      return <h2 key={idx} className="text-sm font-bold text-foreground mt-2 mb-1">{trimmed.slice(3)}</h2>;
    }
    if (trimmed.startsWith("### ")) {
      return <h3 key={idx} className="text-xs font-bold text-foreground mt-1.5 mb-1">{trimmed.slice(4)}</h3>;
    }
    if (trimmed.startsWith("- ") || trimmed.startsWith("* ")) {
      return <li key={idx} className="ml-4 list-disc text-xs text-muted-foreground mb-1">{trimmed.slice(2)}</li>;
    }
    if (trimmed === "") {
      return <div key={idx} className="h-1.5" />;
    }
    
    const formattedHtml = trimmed
      .replace(/\*\*(.*?)\*\*/g, "<strong>$1</strong>")
      .replace(/\*(.*?)\*/g, "<em>$1</em>");
      
    return (
      <p 
        key={idx} 
        className="text-xs text-muted-foreground leading-relaxed mb-1"
        dangerouslySetInnerHTML={{ __html: formattedHtml }}
      />
    );
  });

  return <div className="space-y-0.5">{processed}</div>;
}

export function TemplateModal({
  open,
  onOpenChange,
  title: modalTitle,
  submitLabel,
  isPending,
  initialValues,
  onSubmit,
}: TemplateModalProps) {
  const [activeTab, setActiveTab] = useState<"edit" | "preview">("edit");
  const [code, setCode] = useState("");
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [instruction, setInstruction] = useState("");
  const [certification, setCertification] = useState("");

  useEffect(() => {
    if (open) {
      setActiveTab("edit");
      setCode(initialValues?.code || "");
      setTitle(initialValues?.title || "");
      setDescription(initialValues?.description || "");
      setInstruction(initialValues?.instruction || "");
      setCertification(initialValues?.certification || "");
    }
  }, [open, initialValues]);

  const countWords = (text: string) => {
    return text.trim() === "" ? 0 : text.trim().split(/\s+/).length;
  };

  const isWordCountInvalid = countWords(instruction) > 3000 || countWords(certification) > 3000;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!code.trim() || !title.trim() || isWordCountInvalid) return;
    onSubmit({ code, title, description, instruction, certification });
  };

  return (
    <Modal open={open} onOpenChange={onOpenChange} title={modalTitle} size="3xl">
      <div role="tablist" className="mb-4 flex flex-shrink-0 gap-1 border-b border-border">
        <button
          type="button"
          role="tab"
          aria-selected={activeTab === "edit"}
          onClick={() => setActiveTab("edit")}
          className={`-mb-px h-10 border-b-2 px-4 text-sm font-semibold transition-colors ${
            activeTab === "edit"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          Chỉnh sửa nội dung
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={activeTab === "preview"}
          onClick={() => setActiveTab("preview")}
          className={`-mb-px h-10 border-b-2 px-4 text-sm font-semibold transition-colors ${
            activeTab === "preview"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          Xem trước giao diện
        </button>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4 pt-2">
        {activeTab === "edit" ? (
          <>
            <div>
              <Label htmlFor="templatemodal-code">Mã Code (e.g. PHQ_9)</Label>
              <Input
                id="templatemodal-code"
                type="text"
                required
                placeholder="Nhập mã code viết hoa..."
                value={code}
                onChange={(e) => setCode(e.target.value)}
              />
            </div>
            <div>
              <Label htmlFor="templatemodal-title">Tên bài test</Label>
              <Input
                id="templatemodal-title"
                type="text"
                required
                placeholder="Nhập tên tiêu đề bài test..."
                value={title}
                onChange={(e) => setTitle(e.target.value)}
              />
            </div>
            <div>
              <Label htmlFor="templatemodal-description">Mô tả chi tiết</Label>
              <Textarea
                id="templatemodal-description"
                rows={3}
                placeholder="Mô tả công dụng và hướng dẫn bài test..."
                value={description}
                onChange={(e) => setDescription(e.target.value)}
              />
            </div>
            <div>
              <div className="flex justify-between items-center mb-1">
                <Label className="mb-0">Hướng dẫn (Instruction - Markdown)</Label>
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
                <Label className="mb-0">Chứng nhận (Certification - Markdown)</Label>
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
          </>
        ) : (
          <div className="max-h-[500px] space-y-5 overflow-y-auto rounded-xl border border-border bg-surface p-5">
            <div className="border-b border-border pb-4">
              <Badge tone="primary" className="mb-2 uppercase">
                {code || "CODE_TEST"}
              </Badge>
              <h2 className="text-lg font-bold text-foreground">{title || "Chưa nhập tên tiêu đề"}</h2>
              <p className="mt-1 text-xs text-muted-foreground leading-relaxed">
                {description || "Chưa nhập mô tả chi tiết bài đánh giá."}
              </p>
            </div>

            {/* Instruction block preview */}
            <div className="space-y-2">
              <h3 className="flex items-center gap-2 text-xs font-bold text-foreground">
                <BookOpen className="h-4 w-4 text-primary" /> Hướng dẫn làm bài
              </h3>
              <div className="rounded-xl border border-border bg-background p-3.5">
                <PreviewMarkdownRenderer text={instruction} />
              </div>
            </div>

            {/* Certification block preview */}
            <div className="space-y-2">
              <h3 className="flex items-center gap-2 text-xs font-bold text-foreground">
                <Award className="h-4 w-4 text-success" /> Chứng nhận chuyên môn & Cơ sở khoa học
              </h3>
              <div className="rounded-xl border border-success/20 bg-success-soft p-3.5">
                <PreviewMarkdownRenderer text={certification} />
              </div>
            </div>
          </div>
        )}

        <div className="flex justify-end gap-2 pt-2 border-t border-border mt-6 flex-shrink-0">
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            Hủy
          </Button>
          <Button type="submit" disabled={isPending || isWordCountInvalid}>
            {isPending ? "Đang lưu..." : submitLabel}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
