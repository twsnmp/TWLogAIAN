function escapeHtml(str) {
  return str
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

function parseInline(text) {
  let s = escapeHtml(text);

  // Inline code: `code`
  s = s.replace(/`([^`]+)`/g, '<code class="color-bg-subtle border px-1 py-0 rounded-1">$1</code>');

  // Bold + Italic: ***text***
  s = s.replace(/\*\*\*([^*]+)\*\*\*/g, '<strong><em>$1</em></strong>');

  // Bold: **text**
  s = s.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');

  // Bold with underscores: isolated __text__
  s = s.replace(/(^|[\s(])__([^_]+)__(?=[)\s.,:;!?]|$)/g, '$1<strong>$2</strong>');

  // Italic: *text*
  s = s.replace(/\*([^*\n]+)\*/g, '<em>$1</em>');

  // Italic with underscores: isolated _text_ (do not match inside words like [HOST_1] or snake_case)
  s = s.replace(/(^|[\s(])_([^_]+)_(?=[)\s.,:;!?]|$)/g, '$1<em>$2</em>');

  // Strikethrough: ~~text~~
  s = s.replace(/~~([^~]+)~~/g, '<del>$1</del>');

  // Links: [text](url)
  s = s.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener noreferrer" class="Link--in-modal">$1</a>');

  return s;
}

export function renderMarkdown(md) {
  if (!md) return "";

  const lines = md.split(/\r?\n/);
  const out = [];

  let inCodeBlock = false;
  let codeBlockLang = "";
  let codeBlockLines = [];

  let inUl = false;
  let inOl = false;
  let inTable = false;
  let pendingBlankInList = false;

  const closeList = () => {
    if (inUl) {
      out.push("</ul>");
      inUl = false;
    }
    if (inOl) {
      out.push("</ol>");
      inOl = false;
    }
    pendingBlankInList = false;
  };

  const closeTable = () => {
    if (inTable) {
      out.push("</tbody></table></div>");
      inTable = false;
    }
  };

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmed = line.trim();

    // 1. Fenced Code Block
    const codeMatch = trimmed.match(/^(`{3,}|~{3,})\s*([\w-]*)/);
    if (codeMatch) {
      if (inCodeBlock) {
        // End code block
        const codeContent = escapeHtml(codeBlockLines.join("\n"));
        out.push(`<div class="my-2"><pre class="p-3 color-bg-subtle border rounded-2" style="overflow-x: auto; font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, Liberation Mono, monospace; font-size: 12px; line-height: 1.4;"><code>${codeContent}</code></pre></div>`);
        inCodeBlock = false;
        codeBlockLines = [];
      } else {
        closeList();
        closeTable();
        inCodeBlock = true;
        codeBlockLang = codeMatch[2];
        codeBlockLines = [];
      }
      continue;
    }

    if (inCodeBlock) {
      codeBlockLines.push(line);
      continue;
    }

    // 2. Blank line
    if (trimmed === "") {
      closeTable();
      if (inUl || inOl) {
        // Keep list open for potential next list item (loose list)
        pendingBlankInList = true;
      }
      continue;
    }

    // 3. Horizontal Rule
    if (/^(?:---|\*\*\*|___)$/.test(trimmed)) {
      closeList();
      closeTable();
      out.push('<hr class="my-3" />');
      continue;
    }

    // 4. Headings
    const headingMatch = line.match(/^(#{1,6})\s+(.*)$/);
    if (headingMatch) {
      closeList();
      closeTable();
      const level = headingMatch[1].length;
      const headingText = parseInline(headingMatch[2]);
      const headingClass = level === 1 ? "pb-2 border-bottom f2" : (level === 2 ? "pb-1 border-bottom f3" : "f4");
      out.push(`<h${level} class="${headingClass} mt-3 mb-2 font-weight-bold">${headingText}</h${level}>`);
      continue;
    }

    // 5. Blockquote
    const quoteMatch = line.match(/^>\s*(.*)$/);
    if (quoteMatch) {
      closeList();
      closeTable();
      out.push(`<blockquote class="color-fg-muted pl-3 my-2" style="border-left: 3px solid var(--color-border-default, #30363d);">${parseInline(quoteMatch[1])}</blockquote>`);
      continue;
    }

    // 6. Tables
    if (trimmed.startsWith("|") && trimmed.endsWith("|")) {
      closeList();
      const cells = trimmed.slice(1, -1).split("|").map(c => c.trim());
      const isDivider = cells.every(c => /^:?-+:?$/.test(c));

      if (!inTable) {
        inTable = true;
        out.push('<div class="my-2" style="overflow-x: auto;"><table class="table-sm width-full border"><thead><tr>');
        for (const cell of cells) {
          out.push(`<th class="p-2 border font-weight-bold">${parseInline(cell)}</th>`);
        }
        out.push('</tr></thead><tbody>');
      } else if (!isDivider) {
        out.push('<tr>');
        for (const cell of cells) {
          out.push(`<td class="p-2 border">${parseInline(cell)}</td>`);
        }
        out.push('</tr>');
      }
      continue;
    } else {
      closeTable();
    }

    // 7. Unordered List item
    const ulMatch = line.match(/^(\s*)[-*+]\s+(.*)$/);
    if (ulMatch) {
      if (inOl) {
        closeList();
      }
      if (!inUl) {
        out.push('<ul class="pl-4 my-2">');
        inUl = true;
      }
      pendingBlankInList = false;
      out.push(`<li class="my-1">${parseInline(ulMatch[2])}</li>`);
      continue;
    }

    // 8. Ordered List item (supports "1.", "1)", "2.", etc.)
    const olMatch = line.match(/^(\s*)\d+[.)]\s+(.*)$/);
    if (olMatch) {
      if (inUl) {
        closeList();
      }
      if (!inOl) {
        out.push('<ol class="pl-4 my-2">');
        inOl = true;
      }
      pendingBlankInList = false;
      out.push(`<li class="my-1">${parseInline(olMatch[2])}</li>`);
      continue;
    }

    // 9. List continuation or regular paragraph
    if ((inUl || inOl) && !pendingBlankInList && /^\s{2,}/.test(line)) {
      // Indented text inside current list item
      out.push(`<div class="pl-3 my-1 color-fg-muted">${parseInline(trimmed)}</div>`);
      continue;
    }

    // Regular paragraph
    closeList();
    out.push(`<p class="my-2" style="line-height: 1.6;">${parseInline(line)}</p>`);
  }

  if (inCodeBlock) {
    const codeContent = escapeHtml(codeBlockLines.join("\n"));
    out.push(`<div class="my-2"><pre class="p-3 color-bg-subtle border rounded-2" style="overflow-x: auto; font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, Liberation Mono, monospace; font-size: 12px; line-height: 1.4;"><code>${codeContent}</code></pre></div>`);
  }
  closeList();
  closeTable();

  return out.join("\n");
}
