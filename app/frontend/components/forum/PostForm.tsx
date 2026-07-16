"use client";

import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui";
import { useForumCategories, useCreatePost } from "@/hooks";

const postSchema = z.object({
  categoryId: z.coerce.number().int().positive("Select a category"),
  title: z.string().min(1, "Title is required"),
  summary: z.string().optional(),
  content: z.string().min(1, "Content is required"),
  tags: z.string().optional(),
});

type PostFormInput = z.input<typeof postSchema>;
type PostFormOutput = z.output<typeof postSchema>;

export interface PostFormProps {
  /** "/patient/forum" or "/expert/forum" — redirected to after creation. */
  basePath: string;
}

export function PostForm({ basePath }: PostFormProps) {
  const router = useRouter();
  const { data: categories = [] } = useForumCategories();
  const createPost = useCreatePost();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<PostFormInput, unknown, PostFormOutput>({ resolver: zodResolver(postSchema) });

  const onSubmit = (values: PostFormOutput) => {
    const tags = (values.tags ?? "")
      .split(",")
      .map((t) => t.trim())
      .filter(Boolean);

    createPost.mutate(
      {
        categoryId: values.categoryId,
        title: values.title,
        summary: values.summary,
        content: values.content,
        status: "PUBLISHED",
        tags,
      },
      {
        onSuccess: (post) => router.push(`${basePath}/${post.slug}`),
      }
    );
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div>
        <label className="mb-1 block text-xs font-bold text-foreground">Category</label>
        <select
          {...register("categoryId")}
          className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
        >
          <option value="">Select a category</option>
          {categories.map((category) => (
            <option key={category.id} value={category.id}>
              {category.name}
            </option>
          ))}
        </select>
        {errors.categoryId && <p className="mt-1 text-xs text-danger">{errors.categoryId.message}</p>}
      </div>

      <div>
        <label className="mb-1 block text-xs font-bold text-foreground">Title</label>
        <input
          type="text"
          {...register("title")}
          className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
        />
        {errors.title && <p className="mt-1 text-xs text-danger">{errors.title.message}</p>}
      </div>

      <div>
        <label className="mb-1 block text-xs font-bold text-foreground">Summary</label>
        <input
          type="text"
          {...register("summary")}
          className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
        />
      </div>

      <div>
        <label className="mb-1 block text-xs font-bold text-foreground">Content</label>
        <textarea
          {...register("content")}
          rows={8}
          className="w-full resize-none rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
        />
        {errors.content && <p className="mt-1 text-xs text-danger">{errors.content.message}</p>}
      </div>

      <div>
        <label className="mb-1 block text-xs font-bold text-foreground">Tags (comma-separated)</label>
        <input
          type="text"
          {...register("tags")}
          placeholder="mindfulness, self-care"
          className="w-full rounded-xl border border-border px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary-soft"
        />
      </div>

      <Button type="submit" className="w-full" disabled={createPost.isPending}>
        {createPost.isPending ? "Publishing..." : "Publish post"}
      </Button>
    </form>
  );
}
