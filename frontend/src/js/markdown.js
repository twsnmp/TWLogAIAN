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

  const listStack = [];
  let inTable = false;
  let pendingBlankInList = false;

  const closeListToIndent = (indent) => {
    while (listStack.length > 0 && listStack[listStack.length - 1].indent > indent) {
      const top = listStack.pop();
      out.push(top.type === "ul" ? "</ul>" : "</ol>");
    }
  };

  const closeAllLists = () => {
    while (listStack.length > 0) {
      const top = listStack.pop();
      out.push(top.type === "ul" ? "</ul>" : "</ol>");
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
        closeAllLists();
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
      if (listStack.length > 0) {
        pendingBlankInList = true;
      }
      continue;
    }

    // 3. Horizontal Rule
    if (/^(?:---|\*\*\*|___)$/.test(trimmed)) {
      closeAllLists();
      closeTable();
      out.push('<hr class="my-3" />');
      continue;
    }

    // 4. Headings
    const headingMatch = line.match(/^(#{1,6})\s+(.*)$/);
    if (headingMatch) {
      closeAllLists();
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
      closeAllLists();
      closeTable();
      out.push(`<blockquote class="color-fg-muted pl-3 my-2" style="border-left: 3px solid var(--color-border-default, #30363d);">${parseInline(quoteMatch[1])}</blockquote>`);
      continue;
    }

    // 6. Tables
    if (trimmed.startsWith("|") && trimmed.endsWith("|")) {
      closeAllLists();
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
      const indent = ulMatch[1].length;
      closeListToIndent(indent);

      if (listStack.length === 0 || listStack[listStack.length - 1].indent < indent) {
        listStack.push({ type: "ul", indent });
        out.push('<ul class="pl-4 my-1">');
      } else if (listStack[listStack.length - 1].type !== "ul") {
        const old = listStack.pop();
        out.push(old.type === "ul" ? "</ul>" : "</ol>");
        listStack.push({ type: "ul", indent });
        out.push('<ul class="pl-4 my-1">');
      }
      pendingBlankInList = false;
      out.push(`<li class="my-1">${parseInline(ulMatch[2])}</li>`);
      continue;
    }

    // 8. Ordered List item (supports "1.", "1)", "2.", etc.)
    const olMatch = line.match(/^(\s*)(\d+)[.)]\s+(.*)$/);
    if (olMatch) {
      const indent = olMatch[1].length;
      const num = olMatch[2];
      closeListToIndent(indent);

      if (listStack.length === 0 || listStack[listStack.length - 1].indent < indent) {
        listStack.push({ type: "ol", indent });
        out.push(`<ol start="${num}" class="pl-4 my-1">`);
      } else if (listStack[listStack.length - 1].type !== "ol") {
        const old = listStack.pop();
        out.push(old.type === "ul" ? "</ul>" : "</ol>");
        listStack.push({ type: "ol", indent });
        out.push(`<ol start="${num}" class="pl-4 my-1">`);
      }
      pendingBlankInList = false;
      out.push(`<li value="${num}" class="my-1">${parseInline(olMatch[3])}</li>`);
      continue;
    }

    // 9. List continuation or regular paragraph
    if (listStack.length > 0 && !pendingBlankInList && /^\s{2,}/.test(line)) {
      // Indented text inside current list item
      out.push(`<div class="pl-3 my-1 color-fg-muted">${parseInline(trimmed)}</div>`);
      continue;
    }

    // Regular paragraph
    closeAllLists();
    out.push(`<p class="my-2" style="line-height: 1.6;">${parseInline(line)}</p>`);
  }

  if (inCodeBlock) {
    const codeContent = escapeHtml(codeBlockLines.join("\n"));
    out.push(`<div class="my-2"><pre class="p-3 color-bg-subtle border rounded-2" style="overflow-x: auto; font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, Liberation Mono, monospace; font-size: 12px; line-height: 1.4;"><code>${codeContent}</code></pre></div>`);
  }
  closeAllLists();
  closeTable();

  return out.join("\n");
}
