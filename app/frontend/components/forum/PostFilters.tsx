"use client";

import { Search } from "lucide-react";
import type { Category, Tag } from "@/types";
import { Input, Select } from "@/components/ui";

export interface PostFiltersProps {
  search: string;
  onSearchChange: (search: string) => void;
  categoryId: number | null;
  onCategoryChange: (categoryId: number | null) => void;
  tag: string | null;
  onTagChange: (tag: string | null) => void;
  categories: Category[];
  tags: Tag[];
}

export function PostFilters({
  search,
  onSearchChange,
  categoryId,
  onCategoryChange,
  tag,
  onTagChange,
  categories,
  tags,
}: PostFiltersProps) {
  return (
    <div className="mb-6 space-y-3">
      <div className="flex flex-col gap-3 sm:flex-row">
        <div className="relative flex-1">
          <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-muted-foreground">
            <Search className="h-4 w-4" />
          </div>
          <Input
            type="search"
            aria-label="Search posts"
            value={search}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder="Search posts..."
            className="pl-10"
          />
        </div>
        <Select
          aria-label="Category"
          value={categoryId ?? ""}
          onChange={(e) => onCategoryChange(e.target.value ? Number(e.target.value) : null)}
          className="sm:w-56"
        >
          <option value="">All categories</option>
          {categories.map((category) => (
            <option key={category.id} value={category.id}>
              {category.name}
            </option>
          ))}
        </Select>
      </div>

      {tags.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          <button
            type="button"
            onClick={() => onTagChange(null)}
            aria-pressed={!tag}
            className={`inline-flex h-8 items-center rounded-full border px-3 text-xs font-medium transition-colors ${
              !tag ? "border-primary bg-primary text-white" : "border-border bg-background text-muted-foreground hover:bg-surface hover:text-foreground"
            }`}
          >
            All tags
          </button>
          {tags.map((t) => (
            <button
              key={t.id}
              type="button"
              onClick={() => onTagChange(t.slug)}
              aria-pressed={tag === t.slug}
              className={`inline-flex h-8 items-center rounded-full border px-3 text-xs font-medium transition-colors ${
                tag === t.slug ? "border-primary bg-primary text-white" : "border-border bg-background text-muted-foreground hover:bg-surface hover:text-foreground"
              }`}
            >
              {t.name}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
