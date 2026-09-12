<script>
  import { getTableLimit } from "../../js/define";
  import { X16, FileDirectory16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, onDestroy, tick } from "svelte";
  import {
    showSigmaSeverityChart,
    showSigmaTagsChart,
    showSigmaTimelineChart,
    disposeSigmaChart,
    resizeSigmaChart,
    getSigmaChartImage,
  } from "./sigma";
  import Grid from "../../components/Grid.svelte";
  import { html } from "gridjs";
  import jaJP from "../../js/gridjsJaJP";
  import numeral from "numeral";
  import { _, getLocale } from "../../i18n/i18n";
  import {
    AnalyzeSigmaLogs,
    GetSigmaPacks,
    SelectSigmaRulesDir,
    Export,
    LLMExplainReport,
  } from "../../../wailsjs/go/main/App";
  import ReportAIDialog from "../../components/ReportAIDialog.svelte";
  import * as echarts from "echarts";

  let locale = getLocale();
  let gridLang = locale == "ja" ? jaJP : undefined;

  export let logs = [];
  export let dark = false;
  export let hasLLM = false;

  let showAIDialog = false;
  let aiLang = locale === "en" ? "en" : "ja";
  let selectAILang = "";

  const onSelectAILang = () => {
    if (!selectAILang) return;
    aiLang = selectAILang;
    showAIDialog = true;
    selectAILang = "";
  };

  const dispatch = createEventDispatcher();

  let reportContent = "threats"; // "all" | "threats" | "compliance" | "rules" | "tags"
  let chartType = "severity"; // "severity" | "tags" | "timeline"
  let analyzing = false;
  let saveBusy = false;
  let exportType = "";

  let customRulesPath = "";

  let reportResult = {
    Stats: {
      TotalLogs: 0,
      HitLogs: 0,
      TotalDetections: 0,
      ActiveRules: 0,
      Critical: 0,
      High: 0,
      Medium: 0,
      Low: 0,
      Informational: 0,
      ComplianceHits: 0,
      TopRules: [],
      TopTags: [],
      LogSources: [],
      Timeline: [],
    },
    Items: [],
  };

  // Pagination for right-side top rules list in overview
  let currentRulesPage = 0;
  const rulesPageSize = 5;
  $: totalRulesPages = Math.ceil((reportResult?.Stats?.TopRules?.length || 0) / rulesPageSize);
  $: paginatedTopRules = (reportResult?.Stats?.TopRules || []).slice(
    currentRulesPage * rulesPageSize,
    (currentRulesPage + 1) * rulesPageSize
  );

  const formatLevel = (level) => {
    const l = (level || "medium").toLowerCase();
    let bg = "#6e7681";
    let color = "#fff";
    if (l === "critical") {
      bg = "#cf222e";
    } else if (l === "high") {
      bg = "#f85149";
    } else if (l === "medium") {
      bg = "#d29922";
    } else if (l === "low") {
      bg = "#58a6ff";
    }
    return html(
      `<span style="display:inline-block;padding:2px 8px;border-radius:12px;font-size:11px;font-weight:600;background-color:${bg};color:${color};text-transform:uppercase;">${l}</span>`
    );
  };

  const formatTags = (tags) => {
    if (!tags || tags.length === 0) return "";
    const badges = tags.slice(0, 4).map((t) => {
      let bg = dark ? "#21262d" : "#f6f8fa";
      let color = dark ? "#c9d1d9" : "#24292f";
      let border = dark ? "#30363d" : "#d0d7de";
      const tl = t.toLowerCase();
      if (tl.startsWith("attack.")) {
        color = "#f85149";
      } else if (
        tl.startsWith("pci") ||
        tl.startsWith("nist") ||
        tl.startsWith("gdpr") ||
        tl.startsWith("cis") ||
        tl.startsWith("compliance")
      ) {
        color = "#2ea44f";
      }
      return `<span style="display:inline-block;padding:1px 5px;margin:1px;border-radius:4px;font-size:10px;border:1px solid ${border};background:${bg};color:${color};">${t}</span>`;
    });
    if (tags.length > 4) {
      badges.push(`<span style="font-size:10px;color:#8b949e;">+${tags.length - 4}</span>`);
    }
    return html(badges.join(" "));
  };

  let threatColumns = [
    {
      name: $_("Sigma.Level"),
      width: "9%",
      formatter: (cell) => formatLevel(cell),
    },
    {
      name: $_("Sigma.Time"),
      width: "14%",
      formatter: (cell) => {
        if (!cell) return "";
        const d = new Date(cell / (1000 * 1000));
        return echarts.time.format(d, "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}");
      },
    },
    {
      name: $_("Sigma.RuleTitle"),
      width: "22%",
    },
    {
      name: $_("Sigma.LogSource"),
      width: "12%",
    },
    {
      name: $_("Sigma.Tags"),
      width: "18%",
      formatter: (cell) => formatTags(cell),
    },
    {
      name: $_("Sigma.Log"),
      width: "25%",
    },
  ];

  let rulesColumns = [
    {
      name: $_("Sigma.Count"),
      width: "8%",
      formatter: (cell) => html(`<b>${cell}</b>`),
    },
    {
      name: $_("Sigma.Level"),
      width: "10%",
      formatter: (cell) => formatLevel(cell),
    },
    {
      name: $_("Sigma.RuleTitle"),
      width: "32%",
    },
    {
      name: $_("Sigma.Source"),
      width: "20%",
    },
    {
      name: $_("Sigma.Tags"),
      width: "30%",
      formatter: (cell) => formatTags(cell),
    },
  ];

  let tagsColumns = [
    {
      name: $_("Sigma.Count"),
      width: "12%",
      formatter: (cell) => html(`<b>${cell}</b>`),
    },
    {
      name: $_("Sigma.Category"),
      width: "18%",
      formatter: (cell) => {
        let label = cell;
        let bg = "#0969da";
        if (cell === "mitre") {
          label = "MITRE ATT&CK";
          bg = "#cf222e";
        } else if (cell === "compliance") {
          label = "Compliance";
          bg = "#2ea44f";
        }
        return html(
          `<span style="padding:2px 8px;border-radius:10px;font-size:11px;color:#fff;background:${bg}">${label}</span>`
        );
      },
    },
    {
      name: $_("Sigma.Tag"),
      width: "70%",
    },
  ];

  let columns = threatColumns;
  let data = [];
  let pagination = false;

  const updateTableData = () => {
    if (reportContent === "threats" || reportContent === "all") {
      columns = threatColumns;
      data = (reportResult.Items || []).map((item) => [
        item.Level,
        item.Time,
        item.Title,
        item.LogSource,
        item.Tags,
        item.Log,
      ]);
    } else if (reportContent === "compliance") {
      columns = threatColumns;
      const compItems = (reportResult.Items || []).filter((item) => item.IsCompliance);
      data = compItems.map((item) => [
        item.Level,
        item.Time,
        item.Title,
        item.LogSource,
        item.Tags,
        item.Log,
      ]);
    } else if (reportContent === "rules") {
      columns = rulesColumns;
      data = (reportResult.Stats.TopRules || []).map((r) => [
        r.Count,
        r.Level,
        r.Title,
        r.Source,
        r.Tags,
      ]);
    } else if (reportContent === "tags") {
      columns = tagsColumns;
      data = (reportResult.Stats.TopTags || []).map((t) => [
        t.Count,
        t.Category,
        t.Tag,
      ]);
    }

    if (data.length > 10) {
      pagination = {
        limit: getTableLimit(),
        enable: true,
      };
    } else {
      pagination = false;
    }
  };

  const updateChart = async () => {
    await tick();
    const chartEl = document.getElementById("sigma-chart");
    if (!chartEl) return;

    if (chartType === "severity") {
      showSigmaSeverityChart("sigma-chart", reportResult.Stats, dark);
    } else if (chartType === "tags") {
      showSigmaTagsChart(
        "sigma-chart",
        reportResult.Stats.TopTags,
        dark,
        $_("Sigma.TopTagsChart")
      );
    } else if (chartType === "timeline") {
      showSigmaTimelineChart("sigma-chart", reportResult.Stats.Timeline, dark);
    }
  };

  const runAnalysis = async () => {
    if (!logs || logs.length === 0) return;
    analyzing = true;
    try {
      const res = await AnalyzeSigmaLogs(logs, [], customRulesPath);
      if (res) {
        reportResult = res;
        currentRulesPage = 0;
      }
    } catch (e) {
      console.error("AnalyzeSigmaLogs error:", e);
    } finally {
      analyzing = false;
    }
    updateTableData();
    await updateChart();
  };

  const selectCustomDir = async () => {
    try {
      const dir = await SelectSigmaRulesDir();
      if (dir) {
        customRulesPath = dir;
        await runAnalysis();
      }
    } catch (e) {
      console.error("SelectSigmaRulesDir error:", e);
    }
  };

  const handleReportContentChange = () => {
    updateTableData();
  };

  const handleChartTypeChange = async () => {
    await updateChart();
  };

  const onResize = () => {
    resizeSigmaChart();
  };

  const back = () => {
    disposeSigmaChart();
    dispatch("done", {});
  };

  const exportReport = async () => {
    if (!exportType) return;
    saveBusy = true;
    try {
      let exportItems = [];
      if (reportContent === "threats" || reportContent === "all" || reportContent === "compliance") {
        const sourceItems =
          reportContent === "compliance"
            ? (reportResult.Items || []).filter((i) => i.IsCompliance)
            : reportResult.Items || [];
        exportItems = sourceItems.map((item) => ({
          Time:
            item.Time > 0
              ? echarts.time.format(
                  new Date(item.Time / (1000 * 1000)),
                  "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"
                )
              : "",
          Level: item.Level,
          RuleTitle: item.Title,
          RuleID: item.RuleID,
          LogSource: item.LogSource,
          Tags: (item.Tags || []).join(", "),
          Log: item.Log,
        }));
      } else if (reportContent === "rules") {
        exportItems = (reportResult.Stats.TopRules || []).map((r) => ({
          Count: r.Count,
          Level: r.Level,
          RuleTitle: r.Title,
          RuleID: r.ID,
          Source: r.Source,
          Tags: (r.Tags || []).join(", "),
        }));
      } else if (reportContent === "tags") {
        exportItems = (reportResult.Stats.TopTags || []).map((t) => ({
          Count: t.Count,
          Category: t.Category,
          Tag: t.Tag,
        }));
      }

      await Export(exportType, {
        Title: $_("Sigma.Title"),
        Header:
          exportItems.length > 0 ? Object.keys(exportItems[0]) : ["No Data"],
        Data: exportItems,
        Image: getSigmaChartImage(),
      });
    } catch (e) {
      console.error("Export error:", e);
    } finally {
      exportType = "";
      saveBusy = false;
    }
  };

  onMount(async () => {
    await runAnalysis();
  });

  onDestroy(() => {
    disposeSigmaChart();
  });
</script>

<svelte:window on:resize={onResize} />

<div class="Box mx-auto Box--condensed" style="max-width: 99%; display: flex; flex-direction: column; max-height: calc(100vh - 40px);">
  <!-- Box Header -->
  <div class="Box-header">
    <h3 class="Box-title">{$_('Sigma.Title')}</h3>
  </div>

  <!-- Box Body (Scrollable) -->
  <div style="overflow-y: auto; overflow-x: hidden; flex: 1;">
    <!-- Top Chart Row -->
    <div class="Box-row">
      <div id="sigma-chart"></div>
    </div>

    <!-- Middle Two-Column Overview & Top Rules Row -->
    <div class="Box-row" style="display: flex; gap: 16px; align-items: flex-start;">
      <!-- Left Column: Overview Table -->
      <div class="markdown-body log" style="flex: 1; min-width: 0;">
        <h5>{$_('Sigma.Overview')}</h5>
        <table class="overview-table" width="100%">
          <tbody>
            <tr>
              <th>{$_('Sigma.TotalScanned')}</th>
              <td>{numeral(reportResult.Stats.TotalLogs).format('0,0')}</td>
            </tr>
            <tr>
              <th>{$_('Sigma.TotalHits')}</th>
              <td class="font-weight-bold color-fg-danger">{numeral(reportResult.Stats.TotalDetections).format('0,0')}</td>
            </tr>
            <tr>
              <th>Critical ({$_('Sigma.LevelCritical')})</th>
              <td class="font-weight-bold" style="color: #cf222e;">{numeral(reportResult.Stats.Critical).format('0,0')}</td>
            </tr>
            <tr>
              <th>High ({$_('Sigma.LevelHigh')})</th>
              <td class="font-weight-bold" style="color: #f85149;">{numeral(reportResult.Stats.High).format('0,0')}</td>
            </tr>
            <tr>
              <th>Medium ({$_('Sigma.LevelMedium')})</th>
              <td class="font-weight-bold" style="color: #d29922;">{numeral(reportResult.Stats.Medium).format('0,0')}</td>
            </tr>
            <tr>
              <th>Low / Info ({$_('Sigma.LevelLow')})</th>
              <td style="color: #58a6ff;">{numeral(reportResult.Stats.Low + reportResult.Stats.Informational).format('0,0')}</td>
            </tr>
            <tr>
              <th>{$_('Sigma.ComplianceHits')}</th>
              <td class="font-weight-bold" style="color: #2ea44f;">{numeral(reportResult.Stats.ComplianceHits).format('0,0')}</td>
            </tr>
            <tr>
              <th>{$_('Sigma.ActiveRules')}</th>
              <td>{numeral(reportResult.Stats.ActiveRules).format('0,0')}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Right Column: Top Rules Table -->
      <div class="markdown-body log" style="flex: 1; min-width: 0;">
        <h5>{$_('Sigma.TopRulesTitle')}</h5>
        <table class="extract-table" width="100%">
          <thead>
            <tr>
              <th width="50%">{$_('Sigma.RuleTitle')}</th>
              <th width="15%">{$_('Sigma.Level')}</th>
              <th width="15%">{$_('Sigma.Count')}</th>
              <th width="20%">{$_('Sigma.Source')}</th>
            </tr>
          </thead>
          <tbody>
            {#if paginatedTopRules.length > 0}
              {#each paginatedTopRules as r}
                <tr>
                  <td title={r.Title} style="max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">{r.Title}</td>
                  <td>
                    <span class="level-badge level-{r.Level?.toLowerCase()}">{r.Level}</span>
                  </td>
                  <td><b>{numeral(r.Count).format('0,0')}</b></td>
                  <td style="font-size: 11px; color: var(--color-fg-muted);">{r.Source?.replace('pack:', '')}</td>
                </tr>
              {/each}
            {:else}
              <tr>
                <td colspan="4" class="text-center color-fg-muted">{$_('Sigma.NoHits')}</td>
              </tr>
            {/if}
          </tbody>
        </table>
        {#if (reportResult?.Stats?.TopRules?.length || 0) > rulesPageSize}
          <div class="d-flex flex-justify-center mt-2" style="align-items: center; gap: 8px;">
            <button
              class="btn btn-sm"
              type="button"
              disabled={currentRulesPage === 0}
              on:click={() => currentRulesPage--}
            >
              ◀
            </button>
            <span style="font-size: 14px;">{currentRulesPage + 1} / {totalRulesPages || 1}</span>
            <button
              class="btn btn-sm"
              type="button"
              disabled={currentRulesPage >= totalRulesPages - 1}
              on:click={() => currentRulesPage++}
            >
              ▶
            </button>
          </div>
        {/if}
      </div>
    </div>

    <!-- Bottom Event List Row -->
    <div class="Box-row markdown-body log">
      <h5>{$_('Sigma.EventList')}</h5>
      <Grid {columns} {data} {pagination} language={gridLang} autoWidth={false} />
    </div>
  </div>

  <!-- Box Footer (Buttons and Selects in Same Row) -->
  <div class="Box-footer text-left d-flex flex-justify-between flex-items-center">
    <div class="d-flex flex-items-center" style="gap: 8px;">
      <!-- Back Button -->
      <button class="btn" type="button" on:click={back}>
        <X16 />
        {$_('Sigma.BackBtn')}
      </button>

      <!-- Report Content Select (下段のボタンと同じ行にセレクトで配置) -->
      <select
        class="form-select"
        bind:value={reportContent}
        on:change={handleReportContentChange}
        title={$_('Sigma.ReportContent')}
      >
        <option value="threats">{$_('Sigma.ModeThreats')}</option>
        <option value="compliance">{$_('Sigma.ModeCompliance')}</option>
        <option value="all">{$_('Sigma.ModeAll')}</option>
        <option value="rules">{$_('Sigma.ModeRules')}</option>
        <option value="tags">{$_('Sigma.ModeTags')}</option>
      </select>

      <!-- Chart Type Select -->
      <select
        class="form-select"
        bind:value={chartType}
        on:change={handleChartTypeChange}
        title={$_('Sigma.ChartType')}
      >
        <option value="severity">{$_('Sigma.ChartSeverity')}</option>
        <option value="tags">{$_('Sigma.ChartTags')}</option>
        <option value="timeline">{$_('Sigma.ChartTimeline')}</option>
      </select>

      <!-- External Rules Button -->
      <button
        class="btn"
        type="button"
        on:click={selectCustomDir}
        title={customRulesPath ? customRulesPath : $_('Sigma.SelectCustomRules')}
      >
        <FileDirectory16 />
        {customRulesPath ? customRulesPath.split('/').pop() : $_('Sigma.ExternalRules')}
      </button>

      <!-- AI Explain Select -->
      {#if hasLLM}
        <select
          class="form-select"
          bind:value={selectAILang}
          on:change={onSelectAILang}
        >
          <option value="">{$_('ReportAI.AIExplain')}</option>
          <option value="ja">{$_('ReportAI.LangJa')}</option>
          <option value="en">{$_('ReportAI.LangEn')}</option>
        </select>
      {/if}

      <!-- Export Select -->
      <select
        class="form-select"
        bind:value={exportType}
        on:change={exportReport}
        disabled={saveBusy}
      >
        <option value="">{$_('Sigma.ExportBtn')}</option>
        <option value="csv">CSV</option>
        <option value="json">JSON</option>
        <option value="excel">Excel</option>
        <option value="image">{$_('Sigma.ExportImage')}</option>
      </select>
    </div>

    <!-- Right side status -->
    <div>
      {#if analyzing}
        <span class="color-fg-accent font-weight-bold">{$_('Sigma.Analyzing')}</span>
      {/if}
    </div>
  </div>
</div>

{#if showAIDialog}
  <ReportAIDialog
    bind:show={showAIDialog}
    bind:lang={aiLang}
    title={$_("ReportAI.Title") + " - " + $_("Sigma.Title")}
    exportFilename="sigma_threat_compliance_ai_explanation"
    analyzeFunc={(targetLang) =>
      LLMExplainReport({
        ReportType: "sigma",
        Title: $_("Sigma.Title"),
        Summary: `Scanned ${reportResult.Stats.TotalLogs} logs. Detected ${reportResult.Stats.TotalDetections} threats (${reportResult.Stats.Critical} critical, ${reportResult.Stats.High} high, ${reportResult.Stats.Medium} medium, ${reportResult.Stats.Low} low) and ${reportResult.Stats.ComplianceHits} compliance violations across ${reportResult.Stats.ActiveRules} active rules.`,
        Columns: ["Count", "Level", "RuleTitle", "Tags"],
        Data: (reportResult.Stats.TopRules || []).slice(0, 30).map((r) => [
          r.Count,
          r.Level,
          r.Title,
          (r.Tags || []).join(", "),
        ]),
        Lang: targetLang,
      })}
  />
{/if}

<style>
  #sigma-chart {
    width: 100%;
    height: 250px;
    margin: 5px auto;
  }
  .log table {
    width: 100% !important;
    display: table !important;
  }
  .overview-table th,
  .overview-table td {
    padding: 3px 8px !important;
    font-size: 12px;
    line-height: 1.3;
  }
  .overview-table th {
    width: 50%;
    white-space: nowrap;
  }
  .extract-table th,
  .extract-table td {
    padding: 3px 8px !important;
    font-size: 12px;
    line-height: 1.3;
  }
  .extract-table th {
    white-space: nowrap;
  }
  .level-badge {
    display: inline-block;
    padding: 1px 6px;
    border-radius: 10px;
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    color: #fff;
    background-color: #6e7681;
  }
  .level-critical {
    background-color: #cf222e;
  }
  .level-high {
    background-color: #f85149;
  }
  .level-medium {
    background-color: #d29922;
  }
  .level-low {
    background-color: #58a6ff;
  }
  :global(.log .gridjs-wrapper) {
    overflow-x: hidden;
  }
  :global(.log table.gridjs-table) {
    table-layout: fixed;
    width: 100% !important;
  }
  :global(.log table.gridjs-table th.gridjs-th) {
    padding: 6px 6px;
    box-sizing: border-box;
  }
  :global(.log table.gridjs-table td.gridjs-td) {
    word-break: break-all;
    box-sizing: border-box;
  }
</style>
