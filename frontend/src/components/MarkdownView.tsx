import React from "react";
import { cn } from "../lib/utils";

interface MarkdownViewProps {
  content: string;
  className?: string;
}

export default function MarkdownView({ content, className }: MarkdownViewProps) {
  if (!content) return null;

  // Render markdown lines
  const lines = content.split("\n");
  const elements: React.ReactNode[] = [];
  let inCodeBlock = false;
  let codeBlockLang = "";
  let codeBlockContent: string[] = [];
  let tableRows: string[][] = [];
  let inTable = false;

  const flushTable = (key: string) => {
    if (tableRows.length === 0) return;
    const headerRow = tableRows[0];
    const dataRows = tableRows.slice(1).filter((r) => !r.every((c) => c.match(/^[ -:]+$/)));

    elements.push(
      <div key={key} className="overflow-x-auto my-4 rounded border border-graphite-800">
        <table className="w-full text-left text-xs border-collapse">
          <thead>
            <tr className="bg-graphite-900 border-b border-graphite-800">
              {headerRow.map((col, idx) => (
                <th key={idx} className="p-2.5 font-semibold text-graphite-200">
                  {col.trim()}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-graphite-800/60 bg-graphite-950/40">
            {dataRows.map((row, rIdx) => (
              <tr key={rIdx} className="hover:bg-graphite-900/30 transition-colors">
                {row.map((cell, cIdx) => (
                  <td key={cIdx} className="p-2.5 text-graphite-300 font-mono text-[11px]">
                    {cell.trim()}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
    tableRows = [];
    inTable = false;
  };

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];

    // Code blocks ```
    if (line.startsWith("```")) {
      if (inCodeBlock) {
        // End code block
        elements.push(
          <div key={`code-${i}`} className="my-3 rounded-md bg-graphite-950 border border-graphite-800 overflow-hidden">
            {codeBlockLang && (
              <div className="px-3 py-1 bg-graphite-900/80 border-b border-graphite-800 text-[10px] font-mono text-graphite-400">
                {codeBlockLang}
              </div>
            )}
            <pre className="p-3 text-xs font-mono text-emerald-400 overflow-x-auto whitespace-pre leading-relaxed">
              {codeBlockContent.join("\n")}
            </pre>
          </div>
        );
        codeBlockContent = [];
        inCodeBlock = false;
        codeBlockLang = "";
      } else {
        if (inTable) flushTable(`table-${i}`);
        inCodeBlock = true;
        codeBlockLang = line.replace("```", "").trim();
      }
      continue;
    }

    if (inCodeBlock) {
      codeBlockContent.push(line);
      continue;
    }

    // Markdown Table Detection (| ... | ... |)
    if (line.trim().startsWith("|") && line.trim().endsWith("|")) {
      const cells = line
        .trim()
        .slice(1, -1)
        .split("|");
      tableRows.push(cells);
      inTable = true;
      continue;
    } else if (inTable) {
      flushTable(`table-${i}`);
    }

    // Empty line
    if (!line.trim()) {
      elements.push(<div key={`empty-${i}`} className="h-3" />);
      continue;
    }

    // Headers
    if (line.startsWith("### ")) {
      elements.push(
        <h3 key={`h3-${i}`} className="text-base font-semibold text-white mt-4 mb-2 tracking-tight">
          {renderInline(line.replace("### ", ""))}
        </h3>
      );
      continue;
    }
    if (line.startsWith("## ")) {
      elements.push(
        <h2 key={`h2-${i}`} className="text-lg font-bold text-white mt-5 mb-2.5 pb-1 border-b border-graphite-800 tracking-tight">
          {renderInline(line.replace("## ", ""))}
        </h2>
      );
      continue;
    }
    if (line.startsWith("# ")) {
      elements.push(
        <h1 key={`h1-${i}`} className="text-xl font-bold text-white mt-6 mb-3 pb-1.5 border-b border-graphite-800 tracking-tight">
          {renderInline(line.replace("# ", ""))}
        </h1>
      );
      continue;
    }

    // Blockquote
    if (line.startsWith("> ")) {
      elements.push(
        <blockquote
          key={`quote-${i}`}
          className="my-2.5 border-l-2 border-indigo-500 bg-indigo-950/20 px-3.5 py-2 rounded-r text-sm text-graphite-300 italic"
        >
          {renderInline(line.replace("> ", ""))}
        </blockquote>
      );
      continue;
    }

    // Unordered List (- or *)
    if (line.match(/^[-*]\s+/)) {
      elements.push(
        <li key={`li-${i}`} className="ml-5 list-disc text-sm text-graphite-300 my-1 leading-relaxed">
          {renderInline(line.replace(/^[-*]\s+/, ""))}
        </li>
      );
      continue;
    }

    // Numbered list
    if (line.match(/^\d+\.\s+/)) {
      elements.push(
        <li key={`oli-${i}`} className="ml-5 list-decimal text-sm text-graphite-300 my-1 leading-relaxed">
          {renderInline(line.replace(/^\d+\.\s+/, ""))}
        </li>
      );
      continue;
    }

    // Standard Paragraph
    elements.push(
      <p key={`p-${i}`} className="text-sm text-graphite-300 leading-relaxed my-1.5">
        {renderInline(line)}
      </p>
    );
  }

  if (inTable) {
    flushTable("table-end");
  }

  return <div className={cn("space-y-1", className)}>{elements}</div>;
}

// Inline formatting helper (bold, code, links)
function renderInline(text: string): React.ReactNode {
  const parts: React.ReactNode[] = [];
  let remaining = text;
  let keyIdx = 0;

  while (remaining.length > 0) {
    // Inline code `...`
    const codeMatch = remaining.match(/^`([^`]+)`/);
    if (codeMatch) {
      parts.push(
        <code
          key={keyIdx++}
          className="px-1.5 py-0.5 rounded bg-graphite-900 border border-graphite-800 text-xs font-mono text-emerald-400"
        >
          {codeMatch[1]}
        </code>
      );
      remaining = remaining.slice(codeMatch[0].length);
      continue;
    }

    // Bold **...**
    const boldMatch = remaining.match(/^\*\*([^*]+)\*\*/);
    if (boldMatch) {
      parts.push(
        <strong key={keyIdx++} className="font-semibold text-white">
          {boldMatch[1]}
        </strong>
      );
      remaining = remaining.slice(boldMatch[0].length);
      continue;
    }

    // Italic *...*
    const italicMatch = remaining.match(/^\*([^*]+)\*/);
    if (italicMatch) {
      parts.push(
        <em key={keyIdx++} className="italic text-graphite-200">
          {italicMatch[1]}
        </em>
      );
      remaining = remaining.slice(italicMatch[0].length);
      continue;
    }

    // Link [text](url)
    const linkMatch = remaining.match(/^\[([^\]]+)\]\(([^)]+)\)/);
    if (linkMatch) {
      parts.push(
        <a
          key={keyIdx++}
          href={linkMatch[2]}
          target="_blank"
          rel="noopener noreferrer"
          className="text-indigo-400 hover:text-indigo-300 underline underline-offset-2"
        >
          {linkMatch[1]}
        </a>
      );
      remaining = remaining.slice(linkMatch[0].length);
      continue;
    }

    // Plain text until next special character
    const nextSpecial = remaining.search(/[`*\[]/);
    if (nextSpecial === -1) {
      parts.push(remaining);
      break;
    } else if (nextSpecial === 0) {
      // Special character wasn't part of a valid match, consume 1 char
      parts.push(remaining[0]);
      remaining = remaining.slice(1);
    } else {
      parts.push(remaining.slice(0, nextSpecial));
      remaining = remaining.slice(nextSpecial);
    }
  }

  return <>{parts}</>;
}
