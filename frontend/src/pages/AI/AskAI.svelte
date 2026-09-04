<script>
  import { X16, Question16, Copy16, ShieldLock16, ShieldCheck16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, onDestroy } from "svelte";
  import { AskAIAboutLog } from "../../../wailsjs/go/main/App";
  import { EventsOn, EventsOff } from "../../../wailsjs/runtime/runtime";
  import { copyText } from "svelte-copy";
  import { _, getLocale } from '../../i18n/i18n';
  import { renderMarkdown } from '../../js/markdown';
  import { maskPII } from '../../js/maskpii';

  export let log = "";
  let originalLog = "";
  $: if (log && !originalLog) {
    originalLog = log;
  }

  let errorMsg = "";
  let infoMsg = "";
  let answer = "";

  const dispatch = createEventDispatcher();
  let locale = getLocale();
  let askLang = (locale === "en") ? "en" : "ja";

  let prompt = "";
  let busy = false;
  let isMasked = false;

  onMount(() => {
    EventsOn("ask_ai_stream", (chunk) => {
      answer += chunk;
    });
  });

  onDestroy(() => {
    EventsOff("ask_ai_stream");
  });

  $: presets = askLang === "ja" ? [
    { label: $_('AI.PresetExplain'), text: 'このログの解説をしてください。' },
    { label: $_('AI.PresetCause'), text: 'このログが発生した原因を分析してください。' },
    { label: $_('AI.PresetAction'), text: 'このログに対する対策や対処方法を教えてください。' },
    { label: $_('AI.PresetSecurity'), text: 'セキュリティ上のリスクや脅威があるか評価してください。' },
    { label: $_('AI.PresetPattern'), text: 'このログから重要なパターンや正規表現、抽出可能なフィールドを提示してください。' },
    { label: $_('AI.PresetSummary'), text: 'このログの内容を要約してください。' },
  ] : [
    { label: $_('AI.PresetExplain'), text: 'Please explain this log in detail.' },
    { label: $_('AI.PresetCause'), text: 'Please analyze the cause of this log event.' },
    { label: $_('AI.PresetAction'), text: 'What countermeasures or actions should be taken for this log?' },
    { label: $_('AI.PresetSecurity'), text: 'Please assess the security risks or threats associated with this log.' },
    { label: $_('AI.PresetPattern'), text: 'Please extract key patterns, regex, or structured fields from this log.' },
    { label: $_('AI.PresetSummary'), text: 'Please summarize this log.' },
  ];

  const selectPreset = (p) => {
    prompt = p.text;
  };

  const setAskLang = (lang) => {
    askLang = lang;
  };

  const toggleMask = () => {
    isMasked = !isMasked;
  };

  const askAI = async (withMask = false) => {
    busy = true;
    clearMsg();
    answer = "";
    try {
      const targetLog = (withMask || isMasked) ? maskPII(originalLog || log) : (originalLog || log);
      const r = await AskAIAboutLog(prompt, targetLog, askLang);
      busy = false;
      if (r) {
        if (r.Error) {
          errorMsg = r.Error;
        }
        if (r.Answer && !answer) {
          answer = r.Answer;
        }
        return;
      }
      if (!answer) {
        errorMsg = $_('AI.NoAnswer');
      }
    } catch (e) {
      busy = false;
      errorMsg = e.message || String(e);
    }
  };

  const close = () => {
    dispatch("done", {});
  };

  const clearMsg = () => {
    errorMsg = "";
    infoMsg = "";
  };

  const copy = async () => {
    await copyText(answer);
    infoMsg = $_("LogView.Copied") || "Copied";
    setTimeout(() => {
      infoMsg = "";
    }, 2000);
  };

  $: displayLog = isMasked ? maskPII(originalLog || log) : (originalLog || log);
  $: renderedAnswer = renderMarkdown(answer);
  $: placeholderText = askLang === "ja" ? $_('AI.BlankPromptNote') : 'Leave blank to explain the log';
</script>

<div class="Box mx-auto Box--condensed" style="max-width: 99%; display: flex; flex-direction: column; max-height: calc(100vh - 40px);">
  <div class="Box-header d-flex flex-justify-between flex-items-center">
    <h3 class="Box-title">{$_('AI.AskAI')}</h3>
    <button class="btn btn-sm btn-invisible" type="button" on:click={close}>
      <X16 />
    </button>
  </div>

  {#if errorMsg != ""}
    <div class="flash flash-error m-2">
      {errorMsg}
      <button class="flash-close js-flash-close" type="button" on:click={clearMsg}>
        <X16 />
      </button>
    </div>
  {/if}

  {#if infoMsg != ""}
    <div class="flash m-2">
      {infoMsg}
      <button class="flash-close js-flash-close" type="button" on:click={clearMsg}>
        <X16 />
      </button>
    </div>
  {/if}

  <div class="Box-body p-3" style="overflow-y: auto; flex: 1;">
    <!-- Presets and Language Selector -->
    <div class="d-flex flex-justify-between flex-items-center mb-2 flex-wrap" style="gap: 8px;">
      <div class="d-flex flex-items-center flex-wrap" style="gap: 6px;">
        <span class="text-small color-fg-muted font-weight-bold">{$_('AI.Quattion')} プリセット / Presets:</span>
        {#each presets as p}
          <button
            class="btn btn-sm"
            type="button"
            on:click={() => selectPreset(p)}
          >
            {p.label}
          </button>
        {/each}
      </div>
      <div class="d-flex flex-items-center">
        <span class="text-small color-fg-muted font-weight-bold mr-2">{$_('AI.Lang')}:</span>
        <div class="BtnGroup">
          <button
            class="btn btn-sm BtnGroup-item"
            class:btn-primary={askLang === 'ja'}
            type="button"
            on:click={() => setAskLang('ja')}
          >
            {$_('AI.LangJa')}
          </button>
          <button
            class="btn btn-sm BtnGroup-item"
            class:btn-primary={askLang === 'en'}
            type="button"
            on:click={() => setAskLang('en')}
          >
            {$_('AI.LangEn')}
          </button>
        </div>
      </div>
    </div>

    <!-- Question Input (Compact) -->
    <div class="form-group my-1">
      <div class="form-group-header d-flex flex-justify-between flex-items-center mb-1">
        <label for="ai-prompt-input" class="font-weight-bold text-small">{$_('AI.Quattion')}</label>
        <span class="f6 color-fg-muted">{placeholderText}</span>
      </div>
      <div class="form-group-body">
        <textarea
          id="ai-prompt-input"
          class="form-control"
          style="width: 100%; height: 38px; min-height: 38px; resize: vertical; padding: 4px 8px; font-size: 13px;"
          placeholder={placeholderText}
          bind:value={prompt}
        ></textarea>
      </div>
    </div>

    <!-- Log Display (Read Only & Compact) -->
    <div class="form-group my-1">
      <div class="form-group-header d-flex flex-justify-between flex-items-center mb-1">
        <div>
          <label for="ai-log-input" class="font-weight-bold text-small">{$_('AI.Log')}</label>
          <span class="Label Label--secondary ml-2">{$_('AI.ReadOnly')}</span>
          {#if isMasked}
            <span class="Label Label--accent ml-1">{$_('AI.Masked')}</span>
          {/if}
        </div>
        <div>
          <button class="btn btn-sm" type="button" on:click={toggleMask}>
            {#if isMasked}
              <ShieldCheck16 /> {$_('AI.OriginalLog')}
            {:else}
              <ShieldLock16 /> {$_('AI.MaskPII')}
            {/if}
          </button>
        </div>
      </div>
      <div class="form-group-body">
        <textarea
          id="ai-log-input"
          class="form-control"
          style="width: 100%; height: 42px; min-height: 38px; font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, Liberation Mono, monospace; font-size: 12px; line-height: 1.35; resize: vertical; padding: 4px 8px;"
          readonly
          value={displayLog}
        ></textarea>
      </div>
    </div>

    <!-- Action Buttons -->
    <div class="d-flex flex-items-center my-2" style="gap: 8px;">
      <button class="btn btn-primary" type="button" disabled={busy} on:click={() => askAI(false)}>
        <Question16 />
        {$_('AI.Ask')}
      </button>
      <button class="btn" type="button" disabled={busy} on:click={() => askAI(true)}>
        <ShieldLock16 />
        {$_('AI.AskWithMaskedPII')}
      </button>
      {#if busy}
        <div class="d-flex flex-items-center ml-2 color-fg-attention">
          <span class="AnimatedEllipsis mr-2"></span>
          <span>{$_('AI.AIThinking')}</span>
        </div>
      {/if}
    </div>

    <!-- Answer Section (Single-screen response & streaming) -->
    {#if busy || answer}
      <div class="Box Box--condensed mt-2" style="border: 1px solid var(--color-border-default, #30363d);">
        <div class="Box-header d-flex flex-justify-between flex-items-center">
          <h4 class="Box-title">
            {$_('AI.Answer')}
            {#if busy}
              <span class="AnimatedEllipsis"></span>
            {/if}
          </h4>
          {#if answer}
            <button class="btn btn-sm btn-secondary" type="button" on:click={copy}>
              <Copy16 />
              {$_('AI.Copy')}
            </button>
          {/if}
        </div>
        <div class="Box-body p-3" style="max-height: 460px; overflow-y: auto;">
          {#if answer}
            <div class="markdown-body" style="font-size: 14px;">
              {@html renderedAnswer}
              {#if busy}
                <span class="cursor-blink">▋</span>
              {/if}
            </div>
          {:else if busy}
            <div class="color-fg-muted p-2">
              {$_('AI.AIThinking')}
              <span class="AnimatedEllipsis"></span>
            </div>
          {/if}
        </div>
      </div>
    {/if}
  </div>

  <div class="Box-footer d-flex flex-justify-between">
    <div>
      <button class="btn btn-secondary" type="button" on:click={close}>
        <X16 />
        {$_('AI.Close')}
      </button>
    </div>
    {#if answer}
      <div>
        <button class="btn btn-secondary" type="button" on:click={copy}>
          <Copy16 />
          {$_('AI.Copy')}
        </button>
      </div>
    {/if}
  </div>
</div>

<style>
  .cursor-blink {
    display: inline-block;
    color: var(--color-accent-fg, #58a6ff);
    margin-left: 2px;
    animation: blink-animation 1s step-start infinite;
  }
  @keyframes blink-animation {
    50% {
      opacity: 0;
    }
  }
</style>
