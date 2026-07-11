// Safe, lightweight custom markdown parser (headings, bullet lists, bold/italic)
export function MarkdownRenderer({ text }: { text: string }) {
  if (!text) return null;

  const lines = text.split("\n");
  const processed = lines.map((line, idx) => {
    const trimmed = line.trim();
    if (trimmed.startsWith("# ")) {
      return <h1 key={idx} className="text-lg font-bold text-foreground mt-4 mb-2">{trimmed.slice(2)}</h1>;
    }
    if (trimmed.startsWith("## ")) {
      return <h2 key={idx} className="text-base font-bold text-foreground mt-3 mb-2">{trimmed.slice(3)}</h2>;
    }
    if (trimmed.startsWith("### ")) {
      return <h3 key={idx} className="text-sm font-bold text-foreground mt-2 mb-1">{trimmed.slice(4)}</h3>;
    }
    if (trimmed.startsWith("- ") || trimmed.startsWith("* ")) {
      return <li key={idx} className="ml-4 list-disc text-sm text-muted-foreground mb-1.5">{trimmed.slice(2)}</li>;
    }
    if (trimmed === "") {
      return <div key={idx} className="h-2" />;
    }

    // Bold/Italic parser
    const formattedHtml = trimmed
      .replace(/\*\*(.*?)\*\*/g, "<strong>$1</strong>")
      .replace(/\*(.*?)\*/g, "<em>$1</em>");

    return (
      <p
        key={idx}
        className="text-sm text-muted-foreground leading-relaxed mb-2"
        dangerouslySetInnerHTML={{ __html: formattedHtml }}
      />
    );
  });

  return <div className="space-y-1">{processed}</div>;
}
