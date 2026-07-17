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
  onChange?: (value: string) => void;
  placeholder?: string;
  /** Renders the markdown without a toolbar or editing affordances. */
  readOnly?: boolean;
}

export default function MDXEditorWrapper({ value, onChange, placeholder, readOnly = false }: MDXEditorWrapperProps) {
  const ref = useRef<MDXEditorMethods>(null);

  const lastValueRef = useRef(value);

  // Sync value from parent if it changes externally
  useEffect(() => {
    if (ref.current && value !== lastValueRef.current) {
      ref.current.setMarkdown(value);
      lastValueRef.current = value;
    }
  }, [value]);

  const handleEditorChange = (val: string) => {
    lastValueRef.current = val;
    onChange?.(val);
  };

  return (
    <div
      className={
        readOnly
          ? "w-full"
          : "w-full rounded-lg border border-border bg-background overflow-hidden focus-within:border-primary transition-colors"
      }
    >
      <MDXEditor
        ref={ref}
        markdown={value}
        onChange={handleEditorChange}
        placeholder={placeholder}
        readOnly={readOnly}
        contentEditableClassName={
          readOnly
            ? "prose dark:prose-invert max-w-none text-sm leading-relaxed text-foreground"
            : "prose dark:prose-invert max-w-none min-h-[180px] max-h-[300px] overflow-y-auto p-4 focus:outline-none text-sm text-foreground"
        }
        plugins={
          readOnly
            ? [headingsPlugin(), listsPlugin(), quotePlugin(), thematicBreakPlugin()]
            : [
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
              ]
        }
      />
    </div>
  );
}
