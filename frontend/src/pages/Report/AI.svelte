<script>
  import {
    X16,
    Copy16,
    ShieldLock16,
    ShieldCheck16,
    ChevronDown16,
    ChevronRight16,
    Play16,
  } from "svelte-octicons";
  import * as echarts from "echarts";
  import { createEventDispatcher, onMount, onDestroy } from "svelte";
  import { AnalyzeLogsWithAI } from "../../../wailsjs/go/main/App";
  import { EventsOn, EventsOff } from "../../../wailsjs/runtime/runtime";
  import { copyText } from "svelte-copy";
  import { _, getLocale } from "../../i18n/i18n";
  import { renderMarkdown } from "../../js/markdown";

  export let logs = [];
  export let dark = false;

  const dispatch = createEventDispatcher();
  let locale = getLocale();
  let lang = locale === "en" ? "en" : "ja";

  let busy = false;
  let errorMsg = "";
  let infoMsg = "";
  let maskPII = true;
  let sampleSize = 50;
  let topNError = 10;
  let customPrompt = "";
  let showPatterns = false;

  let summary = {
    TotalEntries: logs ? logs.length : 0,
    ErrorCount: 0,
    WarningCount: 0,
    StartTime: 0,
    EndTime: 0,
    TopErrors: [],
  };

  let markdown = "";
  let stats = "";

  $: presets = lang === "ja" ? [
    { label: $_("AI.PresetGeneral"), text: "" },
    { label: $_("AI.PresetIncident"), text: "重大なセキュリティインシデント、不正アクセス、攻撃の兆候に特化して分析し、危険度と証拠ログ、緊急対処策を提示してください。" },
    { label: $_("AI.PresetPerf"), text: "システムのパフォーマンス障害、レイテンシ悪化、リソース枯渇やエラー急増のボトルネックに特化して原因究明と是正策を提示してください。" },
    { label: $_("AI.PresetActionPlan"), text: "発見された問題点を解消しシステムの可用性・信頼性を高めるための具体的かつ実行可能な推奨アクションプラン・再発防止策を提示してください。" },
  ] : [
    { label: $_("AI.PresetGeneral"), text: "" },
    { label: $_("AI.PresetIncident"), text: "Focus specifically on critical security incidents, intrusion attempts, and anomalies. Highlight high-risk indicators and provide immediate containment procedures." },
    { label: $_("AI.PresetPerf"), text: "Focus specifically on performance bottlenecks, latency spikes, resource exhaustion, and error surges. Identify potential culprits and corrective actions." },
    { label: $_("AI.PresetActionPlan"), text: "Provide a concrete, actionable roadmap and preventative recommendations to improve overall system reliability and stability." },
  ];

  const applyPreset = (p) => {
    customPrompt = p.text;
    runAnalysis();
  };

  onMount(() => {
    EventsOn("ai_report_stream", (chunk) => {
      markdown += chunk;
    });
    if (logs && logs.length > 0) {
      runAnalysis();
    }
  });

  onDestroy(() => {
    EventsOff("ai_report_stream");
  });

  const runAnalysis = async () => {
    if (!logs || logs.length === 0) {
      errorMsg = $_("AI.NoAnswer");
      return;
    }
    busy = true;
    errorMsg = "";
    infoMsg = "";
    markdown = "";
    stats = "";

    try {
      const req = {
        Prompt: customPrompt,
        Lang: lang,
        MaskPII: maskPII,
        TopNError: topNError,
        SampleSize: sampleSize,
      };
      const res = await AnalyzeLogsWithAI(req, logs);
      busy = false;
      if (res) {
        if (res.Summary) {
          summary = res.Summary;
        }
        if (res.Stats) {
          stats = res.Stats;
        }
        if (res.Error) {
          errorMsg = res.Error;
        }
        if (res.Markdown && !markdown) {
          markdown = res.Markdown;
        }
      }
    } catch (e) {
      busy = false;
      errorMsg = e.message || String(e);
    }
  };

  const copy = async () => {
    if (!markdown) return;
    await copyText(markdown);
    infoMsg = $_("LogView.Copied") || "Copied";
    setTimeout(() => {
      infoMsg = "";
    }, 2000);
  };

  const downloadReport = (format) => {
    if (!markdown) return;
    const nowStr = echarts.time.format(new Date(), "{yyyy}{MM}{dd}_{HH}{mm}{ss}");
    const filename = `TWLogAIAN_AI_Report_${nowStr}.${format === "md" ? "md" : "txt"}`;
    const blob = new Blob([markdown], { type: "text/plain;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const formatTime = (nano) => {
    if (!nano || nano <= 0) return "-";
    return echarts.time.format(new Date(nano / (1000 * 1000)), "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}");
  };

  const back = () => {
    dispatch("done", {});
  };

  $: renderedMarkdown = renderMarkdown(markdown);
</script>

<div class="Box mx-auto Box--condensed" style="max-width: 99%; display: flex; flex-direction: column; max-height: calc(100vh - 40px);">
  <!-- Header -->
  <div class="Box-header d-flex flex-justify-between flex-items-center flex-wrap" style="gap: 8px;">
    <div class="d-flex flex-items-center" style="gap: 8px;">
      <h3 class="Box-title m-0">{$_("AI.ReportTitle")}</h3>
      <span class="Label Label--secondary">{$_("AI.TotalLogs")}: {summary.TotalEntries}</span>
      {#if summary.ErrorCount > 0}
        <span class="Label Label--danger">{$_("AI.ErrorLogs")}: {summary.ErrorCount}</span>
      {:else}
        <span class="Label Label--success">{$_("AI.ErrorLogs")}: 0</span>
      {/if}
      {#if summary.WarningCount > 0}
        <span class="Label Label--warning">{$_("AI.WarnLogs")}: {summary.WarningCount}</span>
      {/if}
      {#if summary.StartTime > 0 && summary.EndTime > 0}
        <span class="Label Label--secondary text-small" title="Time range">
          {formatTime(summary.StartTime)} 〜 {formatTime(summary.EndTime)}
        </span>
      {/if}
      {#if stats}
        <span class="Label Label--accent text-small">{stats}</span>
      {/if}
    </div>

    <!-- Controls -->
    <div class="d-flex flex-items-center" style="gap: 6px;">
      <!-- Language -->
      <select class="form-select select-sm" bind:value={lang} on:change={runAnalysis} disabled={busy}>
        <option value="ja">{$_("AI.LangJa")}</option>
        <option value="en">{$_("AI.LangEn")}</option>
      </select>

      <!-- Sample size -->
      <select class="form-select select-sm" bind:value={sampleSize} on:change={runAnalysis} disabled={busy} title={$_("AI.SampleSize")}>
        <option value={20}>20 sample</option>
        <option value={50}>50 sample</option>
        <option value={100}>100 sample</option>
        <option value={200}>200 sample</option>
      </select>

      <!-- PII Mask toggle -->
      <button
        class="btn btn-sm"
        class:btn-outline={maskPII}
        type="button"
        on:click={() => { maskPII = !maskPII; runAnalysis(); }}
        disabled={busy}
        title="Toggle PII Masking"
      >
        {#if maskPII}
          <ShieldLock16 />
          {$_("AI.MaskPII")} ON
        {:else}
          <ShieldCheck16 />
          {$_("AI.MaskPII")} OFF
        {/if}
      </button>

      <!-- Back button -->
      <button class="btn btn-sm btn-secondary" type="button" on:click={back}>
        <X16 />
        {$_("AI.BackBtn")}
      </button>
    </div>
  </div>

  <!-- Messages -->
  {#if errorMsg}
    <div class="flash flash-error m-2 p-2">
      {errorMsg}
    </div>
  {/if}
  {#if infoMsg}
    <div class="flash flash-success m-2 p-2">
      {infoMsg}
    </div>
  {/if}

  <!-- Main Scrollable Content -->
  <div class="Box-body p-3" style="overflow-y: auto; flex: 1;">
    <!-- Preset Action Bar -->
    <div class="mb-3 d-flex flex-wrap flex-items-center" style="gap: 6px;">
      <span class="text-bold text-small mr-1" style="color: var(--color-fg-muted);">Presets:</span>
      {#each presets as p}
        <button
          class="btn btn-sm"
          class:btn-primary={customPrompt === p.text}
          type="button"
          disabled={busy}
          on:click={() => applyPreset(p)}
        >
          {p.label}
        </button>
      {/each}
    </div>

    <!-- Collapsible Top Error Patterns -->
    <div class="mb-3 border rounded p-2" style="background-color: var(--color-canvas-subtle);">
      <div
        class="d-flex flex-items-center flex-justify-between"
        style="cursor: pointer;"
        on:click={() => showPatterns = !showPatterns}
      >
        <div class="d-flex flex-items-center" style="gap: 6px;">
          {#if showPatterns}
            <ChevronDown16 />
          {:else}
            <ChevronRight16 />
          {/if}
          <span class="text-bold">{$_("AI.TopErrors")} ({summary.TopErrors ? summary.TopErrors.length : 0})</span>
        </div>
        <span class="text-small text-muted">
          {showPatterns ? $_("AI.HidePatterns") : $_("AI.ShowPatterns")}
        </span>
      </div>

      {#if showPatterns}
        <div class="mt-2">
          {#if summary.TopErrors && summary.TopErrors.length > 0}
            <table class="table-sm width-full" style="font-size: 12px;">
              <thead>
                <tr>
                  <th style="width: 8%;">#</th>
                  <th style="width: 12%;">{$_("AI.Count")}</th>
                  <th style="width: 40%;">{$_("AI.Pattern")}</th>
                  <th style="width: 40%;">{$_("AI.Example")}</th>
                </tr>
              </thead>
              <tbody>
                {#each summary.TopErrors as ep, i}
                  <tr>
                    <td>{i + 1}</td>
                    <td><span class="Label Label--danger">{ep.Count}</span></td>
                    <td><code style="word-break: break-all;">{ep.Pattern}</code></td>
                    <td><div class="text-small text-muted" style="word-break: break-all; max-height: 60px; overflow-y: auto;">{ep.Example}</div></td>
                  </tr>
                {/each}
              </tbody>
            </table>
          {:else}
            <div class="text-small text-muted p-2">{$_("AI.NoErrors")}</div>
          {/if}
        </div>
      {/if}
    </div>

    <!-- AI Analysis Report Content View -->
    <div class="border rounded p-3 mb-3 markdown-body" style="background-color: var(--color-canvas-default); min-height: 250px;">
      {#if busy && !markdown}
        <div class="d-flex flex-items-center flex-justify-center p-4" style="color: var(--color-fg-muted);">
          <span>{$_("AI.Analyzing")}</span>
          <span class="AnimatedEllipsis ml-2"></span>
        </div>
      {:else if markdown}
        {@html renderedMarkdown}
        {#if busy}
          <div class="mt-2 text-small text-muted d-flex flex-items-center">
            <span>{$_("AI.AIThinking")}</span>
            <span class="AnimatedEllipsis ml-1"></span>
          </div>
        {/if}
      {:else}
        <div class="text-center text-muted p-4">
          {$_("AI.NoAnswer")}
        </div>
      {/if}
    </div>

    <!-- Custom Prompt Input Bar -->
    <div class="d-flex flex-items-center" style="gap: 8px;">
      <input
        type="text"
        class="form-control flex-auto"
        placeholder={$_("AI.CustomPromptPlaceholder")}
        bind:value={customPrompt}
        on:keydown={(e) => { if (e.key === "Enter" && !busy) runAnalysis(); }}
        disabled={busy}
      />
      <button class="btn btn-primary" type="button" on:click={runAnalysis} disabled={busy}>
        <Play16 />
        {$_("AI.AnalyzeLogs")}
      </button>
    </div>
  </div>

  <!-- Footer Actions -->
  <div class="Box-footer d-flex flex-justify-between flex-items-center">
    <div>
      <button class="btn btn-secondary" type="button" on:click={back}>
        <X16 />
        {$_("AI.BackBtn")}
      </button>
    </div>
    <div class="d-flex" style="gap: 8px;">
      <button class="btn btn-sm" type="button" on:click={copy} disabled={!markdown}>
        <Copy16 />
        {$_("AI.Copy")}
      </button>
      <button class="btn btn-sm" type="button" on:click={() => downloadReport("md")} disabled={!markdown}>
        {$_("AI.ExportMD")}
      </button>
      <button class="btn btn-sm" type="button" on:click={() => downloadReport("txt")} disabled={!markdown}>
        {$_("AI.ExportTXT")}
      </button>
    </div>
  </div>
</div>

<style>
  .flash {
    border-radius: 4px;
    font-size: 13px;
  }
  .flash-error {
    background-color: rgba(248, 81, 73, 0.15);
    border: 1px solid rgba(248, 81, 73, 0.4);
    color: #cf222e;
  }
  .flash-success {
    background-color: rgba(46, 160, 67, 0.15);
    border: 1px solid rgba(46, 160, 67, 0.4);
    color: #1a7f37;
  }
</style>
