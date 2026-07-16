"use client";

import { Search } from "lucide-react";
import type { Category, Tag } from "@/types";

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
          <input
            type="text"
            value={search}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder="Search posts..."
            className="w-full rounded-xl border border-border py-2.5 pl-10 pr-3 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
          />
        </div>
        <select
          value={categoryId ?? ""}
          onChange={(e) => onCategoryChange(e.target.value ? Number(e.target.value) : null)}
          className="rounded-xl border border-border px-3 py-2.5 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
        >
          <option value="">All categories</option>
          {categories.map((category) => (
            <option key={category.id} value={category.id}>
              {category.name}
            </option>
          ))}
        </select>
      </div>

      {tags.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          <button
            onClick={() => onTagChange(null)}
            className={`rounded-full px-2.5 py-1 text-xs font-medium ${
              !tag ? "bg-primary text-white" : "bg-surface text-muted-foreground hover:bg-border/50"
            }`}
          >
            All tags
          </button>
          {tags.map((t) => (
            <button
              key={t.id}
              onClick={() => onTagChange(t.slug)}
              className={`rounded-full px-2.5 py-1 text-xs font-medium ${
                tag === t.slug ? "bg-primary text-white" : "bg-surface text-muted-foreground hover:bg-border/50"
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
