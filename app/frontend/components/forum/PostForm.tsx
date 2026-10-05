"use client";

import dynamic from "next/dynamic";
import { useRouter } from "next/navigation";
import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button, FieldError, Input, Label, Select } from "@/components/ui";
import { TagInput } from "./TagInput";
import { useForumCategories, useCreatePost, useUpdatePost } from "@/hooks";

const ContentEditor = dynamic(() => import("@/components/ui/MDXEditor"), { ssr: false });

const postSchema = z.object({
  categoryId: z.coerce.number().int().positive("Select a category"),
  title: z.string().min(1, "Title is required"),
  summary: z.string().optional(),
  content: z.string().min(1, "Content is required"),
  tags: z.array(z.string()).optional(),
});

type PostFormInput = z.input<typeof postSchema>;
type PostFormOutput = z.output<typeof postSchema>;

export interface PostFormInitialValues {
  categoryId: number;
  title: string;
  summary?: string;
  content: string;
  tags: string[];
}

export interface PostFormProps {
  /** "/patient/forum" or "/expert/forum" — redirected to after save. */
  basePath: string;
  mode?: "create" | "edit";
  /** Required when mode is "edit". */
  postId?: number;
  /** Required when mode is "edit" — prefills the form. */
  initialValues?: PostFormInitialValues;
}

export function PostForm({ basePath, mode = "create", postId, initialValues }: PostFormProps) {
  const router = useRouter();
  const { data: categories = [] } = useForumCategories();
  const createPost = useCreatePost();
  const updatePost = useUpdatePost(postId ?? 0);
  const isEdit = mode === "edit";

  const {
    register,
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<PostFormInput, unknown, PostFormOutput>({
    resolver: zodResolver(postSchema),
    defaultValues: initialValues ?? { content: "", tags: [] },
  });

  const onSubmit = (values: PostFormOutput) => {
    const payload = {
      categoryId: values.categoryId,
      title: values.title,
      summary: values.summary,
      content: values.content,
      tags: values.tags ?? [],
    };

    if (isEdit) {
      updatePost.mutate(payload, {
        onSuccess: (post) => router.push(`${basePath}/${post.slug}`),
      });
    } else {
      createPost.mutate(
        { ...payload, status: "PUBLISHED" },
        { onSuccess: (post) => router.push(`${basePath}/${post.slug}`) }
      );
    }
  };

  const isPending = isEdit ? updatePost.isPending : createPost.isPending;

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
      <div>
        <Label htmlFor="post-category">Category</Label>
        <Select id="post-category" invalid={!!errors.categoryId} {...register("categoryId")}>
          <option value="">Select a category</option>
          {categories.map((category) => (
            <option key={category.id} value={category.id}>
              {category.name}
            </option>
          ))}
        </Select>
        <FieldError>{errors.categoryId?.message}</FieldError>
      </div>

      <div>
        <Label htmlFor="post-title">Title</Label>
        <Input id="post-title" type="text" invalid={!!errors.title} {...register("title")} />
        <FieldError>{errors.title?.message}</FieldError>
      </div>

      <div>
        <Label htmlFor="post-summary">Summary</Label>
        <Input id="post-summary" type="text" {...register("summary")} />
      </div>

      <div>
        <Label>Content</Label>
        <Controller
          name="content"
          control={control}
          render={({ field }) => (
            <ContentEditor value={field.value ?? ""} onChange={field.onChange} placeholder="Share your thoughts..." />
          )}
        />
        <FieldError>{errors.content?.message}</FieldError>
      </div>

      <div>
        <Label>Tags</Label>
        <Controller
          name="tags"
          control={control}
          render={({ field }) => (
            <TagInput value={field.value ?? []} onChange={field.onChange} placeholder="Type # to add a tag" />
          )}
        />
      </div>

      <Button type="submit" size="lg" className="w-full" disabled={isPending}>
        {isEdit ? (isPending ? "Saving..." : "Save changes") : isPending ? "Publishing..." : "Publish post"}
      </Button>
    </form>
  );
}
