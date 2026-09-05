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
  import { AnalyzeLogsWithAI, CancelAIReport, GetLogSummary } from "../../../wailsjs/go/main/App";
  import { EventsOn, EventsOff } from "../../../wailsjs/runtime/runtime";
  import { copyText } from "svelte-copy";
  import { _, getLocale } from "../../i18n/i18n";
  import { renderMarkdown } from "../../js/markdown";
  import { getLogLevel } from "../Search/logview.js";

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
  let selectedPresetIndex = -1;

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

  const jaPresets = [
    { label: "全体サマリー・異常検知", text: "ログ全体の傾向、エラーや異常パターンの有無、推奨される対応策を総合的に分析してください。" },
    { label: "重大インシデント・不正アクセス調査", text: "重大なセキュリティインシデント、不正アクセス、攻撃の兆候に特化して分析し、危険度と証拠ログ、緊急対処策を提示してください。" },
    { label: "性能障害・エラー急増の特定", text: "システムのパフォーマンス障害、レイテンシ悪化、リソース枯渇やエラー急増のボトルネックに特化して原因究明と是正策を提示してください。" },
    { label: "推奨アクション・再発防止策", text: "発見された問題点を解消しシステムの可用性・信頼性を高めるための具体的かつ実行可能な推奨アクションプラン・再発防止策を提示してください。" },
  ];

  const enPresets = [
    { label: "Overall Summary & Anomalies", text: "Analyze overall log trends, identify anomalies and error patterns, and summarize actionable recommendations." },
    { label: "Critical Incidents & Security", text: "Focus specifically on critical security incidents, intrusion attempts, and anomalies. Highlight high-risk indicators and provide immediate containment procedures." },
    { label: "Performance Bottlenecks & Spikes", text: "Focus specifically on performance bottlenecks, latency spikes, resource exhaustion, and error surges. Identify potential culprits and corrective actions." },
    { label: "Recommendations & Action Plan", text: "Provide a concrete, actionable roadmap and preventative recommendations to improve overall system reliability and stability." },
  ];

  $: presets = lang === "ja" ? jaPresets : enPresets;

  const selectPreset = (p, idx) => {
    selectedPresetIndex = idx;
    customPrompt = p.text;
  };

  const setLang = (newLang) => {
    if (lang === newLang) return;
    const oldPresets = lang === "ja" ? jaPresets : enPresets;
    const newPresets = newLang === "ja" ? jaPresets : enPresets;
    const matchIdx = oldPresets.findIndex((p) => p.text === customPrompt);
    if (matchIdx !== -1) {
      customPrompt = newPresets[matchIdx].text;
      selectedPresetIndex = matchIdx;
    }
    lang = newLang;
  };

  $: t = lang === "en" ? {
    title: "AI Analysis Report",
    totalLogs: "Total",
    errorLogs: "Errors",
    warnLogs: "Warnings",
    timeRange: "Time Range",
    presets: "Presets:",
    placeholder: "Enter additional questions or analysis instructions (leave blank for standard analysis)",
    analyzeBtn: "Analyze Logs with AI",
    analyzingBtn: "Analyzing...",
    topErrors: "Top Error Patterns",
    showPatterns: "Show Error Patterns",
    hidePatterns: "Hide Error Patterns",
    count: "Count",
    pattern: "Pattern",
    example: "Example",
    noErrors: "No error logs detected",
    analyzingMsg: "AI is analyzing logs...",
    thinkingMsg: "AI is thinking...",
    emptyPrompt: "Click \"Analyze Logs with AI\" button above to start AI log analysis.",
    back: "Back",
    copy: "Copy",
    copied: "Copied to clipboard",
    exportMD: "Markdown",
    exportTXT: "Text",
    lang: "Lang:",
    maskPII: "Mask PII",
  } : {
    title: $_("AI.ReportTitle") || "AI分析レポート",
    totalLogs: $_("AI.TotalLogs") || "総ログ数",
    errorLogs: $_("AI.ErrorLogs") || "エラー数",
    warnLogs: $_("AI.WarnLogs") || "警告数",
    timeRange: $_("AI.TimeRange") || "期間",
    presets: "プリセット:",
    placeholder: "追加の質問や分析の指示を入力してください（空欄の場合は標準分析を実行します）",
    analyzeBtn: $_("AI.AnalyzeLogs") || "AI分析を実行",
    analyzingBtn: "分析中...",
    topErrors: $_("AI.TopErrors") || "頻出エラーパターン (上位)",
    showPatterns: $_("AI.ShowPatterns") || "頻出エラーパターンを表示",
    hidePatterns: $_("AI.HidePatterns") || "頻出エラーパターンを閉じる",
    count: $_("AI.Count") || "発生回数",
    pattern: $_("AI.Pattern") || "パターン",
    example: $_("AI.Example") || "該当例",
    noErrors: $_("AI.NoErrors") || "エラーログは検出されませんでした",
    analyzingMsg: "AIがログ全体を分析しています...",
    thinkingMsg: $_("AI.AIThinking") || "AIが考えています...",
    emptyPrompt: "「AI分析を実行」ボタンをクリックして分析を開始してください。",
    back: $_("AI.BackBtn") || "戻る",
    copy: $_("AI.Copy") || "コピー",
    copied: $_("LogView.Copied") || "コピーしました",
    exportMD: "Markdown",
    exportTXT: "テキスト",
    lang: "言語:",
    maskPII: $_("AI.MaskPII") || "PIIマスク",
  };

  const calculateInitialSummary = () => {
    if (!logs || logs.length === 0) {
      summary = {
        TotalEntries: 0,
        ErrorCount: 0,
        WarningCount: 0,
        StartTime: 0,
        EndTime: 0,
        TopErrors: [],
      };
      return;
    }

    let errCount = 0;
    let warnCount = 0;
    let minTime = logs[0].Time || 0;
    let maxTime = logs[0].Time || 0;

    for (const l of logs) {
      if (l.Time > 0) {
        if (minTime === 0 || l.Time < minTime) minTime = l.Time;
        if (l.Time > maxTime) maxTime = l.Time;
      }
      const lvl = getLogLevel(l);
      if (lvl === "error") {
        errCount++;
      } else if (lvl === "warn") {
        warnCount++;
      }
    }

    summary = {
      TotalEntries: logs.length,
      ErrorCount: errCount,
      WarningCount: warnCount,
      StartTime: minTime,
      EndTime: maxTime,
      TopErrors: summary.TopErrors || [],
    };
  };

  const loadInitialSummary = async () => {
    if (!logs || logs.length === 0) return;
    try {
      const preparedLogs = logs.map((l) => {
        const kv = l.KeyValue ? { ...l.KeyValue } : {};
        kv["_level"] = getLogLevel(l);
        return {
          ...l,
          KeyValue: kv,
        };
      });
      const s = await GetLogSummary({
        Prompt: "",
        Lang: lang,
        MaskPII: maskPII,
        TopNError: topNError,
        SampleSize: sampleSize,
      }, preparedLogs);
      if (s) {
        summary = s;
        if (summary.TopErrors && summary.TopErrors.length > 0) {
          showPatterns = true;
        }
      }
    } catch (e) {
      console.warn("GetLogSummary error:", e);
    }
  };

  $: if (logs) {
    calculateInitialSummary();
    loadInitialSummary();
  }

  onMount(() => {
    EventsOn("ai_report_stream", (chunk) => {
      markdown += chunk;
    });
    EventsOn("ai_report_summary", (s) => {
      if (s) {
        summary = s;
        if (summary.TopErrors && summary.TopErrors.length > 0) {
          showPatterns = true;
        }
      }
    });
    calculateInitialSummary();
    loadInitialSummary();
  });

  onDestroy(() => {
    EventsOff("ai_report_stream");
    EventsOff("ai_report_summary");
    try {
      CancelAIReport();
    } catch (e) {}
  });

  const runAnalysis = async () => {
    if (!logs || logs.length === 0) {
      errorMsg = $_("AI.NoAnswer");
      return;
    }

    try {
      await CancelAIReport();
    } catch (e) {}

    busy = true;
    errorMsg = "";
    infoMsg = "";
    markdown = "";
    stats = "";

    try {
      const preparedLogs = logs.map((l) => {
        const kv = l.KeyValue ? { ...l.KeyValue } : {};
        kv["_level"] = getLogLevel(l);
        return {
          ...l,
          KeyValue: kv,
        };
      });

      const req = {
        Prompt: customPrompt,
        Lang: lang,
        MaskPII: maskPII,
        TopNError: topNError,
        SampleSize: sampleSize,
      };
      const res = await AnalyzeLogsWithAI(req, preparedLogs);
      busy = false;
      if (res) {
        if (res.Summary) {
          summary = res.Summary;
          if (summary.TopErrors && summary.TopErrors.length > 0) {
            showPatterns = true;
          }
        }
        if (res.Stats) {
          stats = res.Stats;
        }
        if (res.Error) {
          if (res.Error !== "Analysis canceled") {
            errorMsg = res.Error;
          }
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
    infoMsg = t.copied;
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
    try {
      CancelAIReport();
    } catch (e) {}
    dispatch("done", {});
  };

  $: renderedMarkdown = renderMarkdown(markdown);
</script>

<div class="Box mx-auto Box--condensed" style="max-width: 99%; display: flex; flex-direction: column; max-height: calc(100vh - 40px);">
  <!-- Header -->
  <div class="Box-header d-flex flex-items-center flex-wrap py-2" style="row-gap: 8px; column-gap: 12px;">
    <h3 class="Box-title mr-2" style="white-space: nowrap; flex-shrink: 0;">
      {t.title}
      {#if busy}
        <span class="Label Label--warning ml-2">{t.analyzingBtn}</span>
      {/if}
    </h3>

    <!-- Summary Badges -->
    <div class="d-flex flex-items-center flex-wrap flex-auto f6" style="gap: 6px;">
      <span class="Label Label--info">{t.totalLogs}: {summary.TotalEntries}</span>
      {#if summary.ErrorCount > 0}
        <span class="Label Label--danger">{t.errorLogs}: {summary.ErrorCount}</span>
      {:else if summary.TotalEntries > 0}
        <span class="Label">{t.errorLogs}: 0</span>
      {/if}
      {#if summary.WarningCount > 0}
        <span class="Label Label--warning">{t.warnLogs}: {summary.WarningCount}</span>
      {/if}
      {#if summary.StartTime > 0 && summary.EndTime > 0}
        <span class="Label text-small" title={t.timeRange}>
          {formatTime(summary.StartTime)} 〜 {formatTime(summary.EndTime)}
        </span>
      {/if}
      {#if stats}
        <span class="Label Label--accent text-small">{stats}</span>
      {/if}
    </div>

    <!-- Right Controls -->
    <div class="d-flex flex-items-center ml-auto" style="gap: 8px; flex-shrink: 0;">
      <!-- Language Toggle Group -->
      <div class="d-flex flex-items-center">
        <span class="text-small color-fg-muted font-weight-bold mr-1">{t.lang}</span>
        <div class="BtnGroup">
          <button
            class="btn btn-sm BtnGroup-item"
            class:btn-primary={lang === "ja"}
            type="button"
            on:click={() => setLang("ja")}
            disabled={busy}
          >
            日本語
          </button>
          <button
            class="btn btn-sm BtnGroup-item"
            class:btn-primary={lang === "en"}
            type="button"
            on:click={() => setLang("en")}
            disabled={busy}
          >
            English
          </button>
        </div>
      </div>

      <!-- Sample size -->
      <select class="form-select select-sm" bind:value={sampleSize} disabled={busy} title="Sample Size">
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
        on:click={() => { maskPII = !maskPII; }}
        disabled={busy}
        title="Toggle PII Masking"
      >
        {#if maskPII}
          <ShieldLock16 />
          {t.maskPII} ON
        {:else}
          <ShieldCheck16 />
          {t.maskPII} OFF
        {/if}
      </button>

      <!-- Back button -->
      <button class="btn btn-sm btn-secondary" type="button" on:click={back}>
        <X16 />
        {t.back}
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
    <div class="mb-2 d-flex flex-wrap flex-items-center" style="gap: 6px;">
      <span class="text-bold text-small mr-1" style="color: var(--color-fg-muted);">{t.presets}</span>
      {#each presets as p, idx}
        <button
          class="btn btn-sm"
          class:btn-outline={customPrompt === p.text}
          type="button"
          disabled={busy}
          on:click={() => selectPreset(p, idx)}
        >
          {p.label}
        </button>
      {/each}
    </div>

    <!-- Custom Prompt Input Bar (Swapped ABOVE Results) -->
    <div class="d-flex flex-items-center mb-3" style="gap: 8px;">
      <input
        type="text"
        class="form-control flex-auto"
        placeholder={t.placeholder}
        bind:value={customPrompt}
        on:keydown={(e) => { if (e.key === "Enter" && !busy) runAnalysis(); }}
        disabled={busy}
      />
      <button
        class="btn btn-primary d-flex flex-items-center flex-justify-center"
        style="white-space: nowrap; flex-shrink: 0; min-width: 130px; gap: 6px;"
        type="button"
        on:click={runAnalysis}
        disabled={busy}
      >
        {#if busy}
          <span class="AnimatedEllipsis"></span>
          <span>{t.analyzingBtn}</span>
        {:else}
          <Play16 />
          <span>{t.analyzeBtn}</span>
        {/if}
      </button>
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
          <span class="text-bold">{t.topErrors} ({summary.TopErrors ? summary.TopErrors.length : 0})</span>
        </div>
        <span class="text-small text-muted">
          {showPatterns ? t.hidePatterns : t.showPatterns}
        </span>
      </div>

      {#if showPatterns}
        <div class="mt-2">
          {#if summary.TopErrors && summary.TopErrors.length > 0}
            <table class="table-sm width-full" style="font-size: 12px;">
              <thead>
                <tr>
                  <th style="width: 8%;">#</th>
                  <th style="width: 12%;">{t.count}</th>
                  <th style="width: 40%;">{t.pattern}</th>
                  <th style="width: 40%;">{t.example}</th>
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
            <div class="text-small text-muted p-2">{t.noErrors}</div>
          {/if}
        </div>
      {/if}
    </div>

    <!-- AI Analysis Report Content View -->
    <div class="border rounded p-3 mb-2 markdown-body" style="background-color: var(--color-canvas-default); min-height: 280px;">
      {#if busy && !markdown}
        <div class="d-flex flex-items-center flex-justify-center p-4" style="color: var(--color-fg-muted);">
          <span>{t.analyzingMsg}</span>
          <span class="AnimatedEllipsis ml-2"></span>
        </div>
      {:else if markdown}
        {@html renderedMarkdown}
        {#if busy}
          <div class="mt-2 text-small text-muted d-flex flex-items-center">
            <span>{t.thinkingMsg}</span>
            <span class="AnimatedEllipsis ml-1"></span>
          </div>
        {/if}
      {:else}
        <div class="text-center text-muted p-4">
          <p class="mb-2">{t.emptyPrompt}</p>
        </div>
      {/if}
    </div>
  </div>

  <!-- Footer Actions -->
  <div class="Box-footer d-flex flex-justify-between flex-items-center">
    <div>
      <button class="btn btn-secondary" type="button" on:click={back}>
        <X16 />
        {t.back}
      </button>
    </div>
    <div class="d-flex" style="gap: 8px;">
      <button class="btn btn-sm" type="button" on:click={copy} disabled={!markdown}>
        <Copy16 />
        {t.copy}
      </button>
      <button class="btn btn-sm" type="button" on:click={() => downloadReport("md")} disabled={!markdown}>
        {t.exportMD}
      </button>
      <button class="btn btn-sm" type="button" on:click={() => downloadReport("txt")} disabled={!markdown}>
        {t.exportTXT}
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
