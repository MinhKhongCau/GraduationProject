"use client";

import { PostForm } from "@/components/forum";
import { Card } from "@/components/ui";
import { ROUTES } from "@/constants";

export default function NewForumPostPage() {
  return (
    <div className="mx-auto max-w-2xl">
      <h1 className="mb-6 text-2xl font-bold text-foreground">New post</h1>
      <Card className="p-6">
        <PostForm basePath={ROUTES.PATIENT.FORUM} />
      </Card>
    </div>
  );
}
