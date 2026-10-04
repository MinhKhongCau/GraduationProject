"use client";

import { PostForm } from "@/components/forum";
import { Card, PageHeader } from "@/components/ui";
import { ROUTES } from "@/constants";

export default function NewForumPostPage() {
  return (
    <div className="mx-auto max-w-2xl">
      <PageHeader title="New post" />
      <Card className="p-5 sm:p-6">
        <PostForm basePath={ROUTES.PATIENT.FORUM} />
      </Card>
    </div>
  );
}
