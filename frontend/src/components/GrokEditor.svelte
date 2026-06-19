<script>
  import highlightWords from "highlight-words";
  import { createEventDispatcher } from "svelte";

  export let value = "";

  const dispatch = createEventDispatcher();
  let textarea;
  let backdrop;
  let scrollTop = 0;
  let scrollLeft = 0;

  // Sync scroll positions
  function handleScroll() {
    if (textarea && backdrop) {
      backdrop.scrollTop = textarea.scrollTop;
      backdrop.scrollLeft = textarea.scrollLeft;
    }
  }

  // Get current selection start, end, and selected text
  export function getSelection() {
    if (!textarea) {
      return { index: 0, length: 0, text: "" };
    }
    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    return {
      index: start,
      length: end - start,
      text: value.substring(start, end)
    };
  }

  // Replace current selection and restore focus/cursor position
  export function replaceSelection(replacement) {
    if (!textarea) return;
    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;

    value = value.substring(0, start) + replacement + value.substring(end);

    setTimeout(() => {
      textarea.focus();
      textarea.setSelectionRange(start + replacement.length, start + replacement.length);
      handleScroll();
    }, 0);
  }

  // Keydown shortcuts (Cmd+W/Ctrl+W for Grok, Cmd+S/Ctrl+S for Wildcard)
  function handleKeyDown(e) {
    const isMeta = e.metaKey || e.ctrlKey;
    if (isMeta && e.key.toLowerCase() === "w") {
      e.preventDefault();
      dispatch("action", { type: "grok" });
    } else if (isMeta && e.key.toLowerCase() === "s") {
      e.preventDefault();
      dispatch("action", { type: "wildcard" });
    }
  }

  // Grok color mapping logic
  const getGrokColor = (c) => {
    if (!c.match) {
      return "";
    }
    if (c.text.includes("%{IP")) {
      return "#1a7f37";
    }
    if (c.text.includes("%{MAC")) {
      return "#8250df";
    }
    if (c.text.includes("%{INT") || c.text.includes("%{NUM")) {
      return "#2da44e";
    }
    if (c.text.includes("%{URL")) {
      return "#bc4c00";
    }
    if (c.text.includes("%{")) {
      return "#bf8722";
    }
    if (c.text.includes("s")) {
      return "#0969da";
    }
    return "#cf222e";
  };

  // Keep single-line pattern (grok patterns don't have newlines)
  $: {
    if (value && value.includes("\n")) {
      value = value.replaceAll("\n", "");
    }
  }

  // Highlight words based on grok regex chunking
  $: chunks = highlightWords({
    text: value || "",
    query: /(%\{.+?\}|\.\+|\\s\+)/,
  });
</script>

<div class="grok-editor">
  <div class="backdrop" bind:this={backdrop}>
    {#each chunks as chunk}
      {#if chunk.match}
        <span 
          style="color: {getGrokColor(chunk)}; text-decoration: {chunk.text.startsWith('%{') ? 'underline' : 'none'}"
          class="highlight"
        >{chunk.text}</span>
      {:else}
        <span>{chunk.text}</span>
      {/if}
    {/each}
  </div>
  <textarea
    bind:this={textarea}
    value={value || ""}
    on:input={(e) => value = e.target.value}
    on:scroll={handleScroll}
    on:keydown={handleKeyDown}
    autocomplete="off"
    autocorrect="off"
    autocapitalize="off"
    spellcheck="false"
  ></textarea>
</div>

<style>
  .grok-editor {
    position: relative;
    box-sizing: border-box;
    width: 100%;
    height: 80px;
    border: 1px solid #d0d7de;
    border-radius: 6px;
    background-color: #ffffff;
    overflow: hidden;
  }

  .grok-editor:focus-within {
    border-color: #0969da;
    box-shadow: 0 0 0 3px rgba(9, 105, 218, 0.3);
  }

  .backdrop, textarea {
    position: absolute;
    top: 0;
    left: 0;
    margin: 0;
    padding: 8px 12px;
    width: 100%;
    height: 100%;
    font-family: monospace, serif;
    font-size: 14px;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-all;
    box-sizing: border-box;
  }

  textarea {
    color: transparent;
    background: transparent;
    caret-color: #24292f;
    resize: none;
    border: none;
    outline: none;
    overflow: auto;
    display: block;
  }

  .backdrop {
    pointer-events: none;
    color: #24292f;
    overflow: hidden;
  }

  .highlight {
    font-weight: bold;
  }
</style>
