"use client";

import {
  MDXEditor,
  type MDXEditorMethods,
  headingsPlugin,
  listsPlugin,
  quotePlugin,
  thematicBreakPlugin,
  markdownShortcutPlugin,
  toolbarPlugin,
  UndoRedo,
  BoldItalicUnderlineToggles,
  BlockTypeSelect,
  ListsToggle,
} from "@mdxeditor/editor";
import "@mdxeditor/editor/style.css";
import { useEffect, useRef } from "react";

interface MDXEditorWrapperProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
}

export default function MDXEditorWrapper({ value, onChange, placeholder }: MDXEditorWrapperProps) {
  const ref = useRef<MDXEditorMethods>(null);

  // Sync value from parent if it changes externally
  useEffect(() => {
    if (ref.current) {
      const currentVal = ref.current.getMarkdown();
      if (currentVal !== value) {
        ref.current.setMarkdown(value);
      }
    }
  }, [value]);

  return (
    <div className="w-full rounded-lg border border-border bg-background overflow-hidden focus-within:border-primary transition-colors">
      <MDXEditor
        ref={ref}
        markdown={value}
        onChange={onChange}
        placeholder={placeholder}
        contentEditableClassName="prose dark:prose-invert max-w-none min-h-[180px] max-h-[300px] overflow-y-auto p-4 focus:outline-none text-sm text-foreground"
        plugins={[
          headingsPlugin(),
          listsPlugin(),
          quotePlugin(),
          thematicBreakPlugin(),
          markdownShortcutPlugin(),
          toolbarPlugin({
            toolbarContents: () => (
              <div className="flex flex-wrap items-center gap-1.5 border-b border-border bg-surface/50 p-1.5 w-full">
                <UndoRedo />
                <span className="w-px h-4 bg-border mx-1" />
                <BoldItalicUnderlineToggles />
                <span className="w-px h-4 bg-border mx-1" />
                <BlockTypeSelect />
                <span className="w-px h-4 bg-border mx-1" />
                <ListsToggle />
              </div>
            ),
          }),
        ]}
      />
    </div>
  );
}
