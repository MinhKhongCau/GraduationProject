"use client";

import { PostForm } from "@/components/forum";
import { Card, PageHeader } from "@/components/ui";
import { ROUTES } from "@/constants";

export default function NewExpertForumPostPage() {
  return (
    <div className="mx-auto max-w-2xl">
      <PageHeader title="New post" />
      <Card className="p-6">
        <PostForm basePath={ROUTES.EXPERT.FORUM} />
      </Card>
    </div>
  );
}
