<script>
  import { getTableLimit } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, onDestroy, tick } from "svelte";
  import {
    showEmailSPFChart,
    showEmailCountBarChart,
    showEmailDelayChart,
    disposeEmailChart,
    resizeEmailChart,
    getEmailChartImage,
  } from "./email";
  import Grid from "../../components/Grid.svelte";
  import jaJP from "../../js/gridjsJaJP";
  import { _, getLocale } from "../../i18n/i18n";
  import { AnalyzeEmailLogs, Export, LLMExplainReport } from "../../../wailsjs/go/main/App";
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

  let chartDiv;
  let chartType = "spf"; // "spf" | "domain" | "from" | "to" | "senderIP" | "delay"
  let tableMode = "items"; // "items" | "matrix"
  let checkSPF = false;
  let analyzing = false;
  let saveBusy = false;
  let exportType = "";

  let reportResult = {
    Stats: {
      TotalEmails: 0,
      SPFPass: 0,
      SPFFail: 0,
      SPFSoftFail: 0,
      SPFNone: 0,
      SPFOther: 0,
      AvgDelay: 0,
      MaxDelay: 0,
      AvgRelay: 0,
      MaxRelay: 0,
      TopFrom: [],
      TopTo: [],
      TopDomain: [],
      TopSenderIP: [],
      Matrix: [],
    },
    Items: [],
  };

  let itemsColumns = [
    {
      name: $_("Email.Time"),
      width: "15%",
      formatter: (cell) => {
        if (!cell) return "";
        const d = new Date(cell);
        return echarts.time.format(d, "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}");
      },
    },
    {
      name: $_("Email.From"),
      width: "18%",
    },
    {
      name: $_("Email.To"),
      width: "18%",
    },
    {
      name: $_("Email.Subject"),
      width: "20%",
    },
    {
      name: $_("Email.DelaySec"),
      width: "8%",
      formatter: (cell) => (cell !== undefined ? `${cell}s` : ""),
    },
    {
      name: $_("Email.Relay"),
      width: "6%",
    },
    {
      name: $_("Email.SenderIP"),
      width: "9%",
    },
    {
      name: "SPF",
      width: "6%",
      formatter: (cell) => {
        if (!cell) return "";
        let color = "#8b949e";
        if (cell === "pass") color = "#2ea44f";
        else if (cell === "fail") color = "#cf222e";
        else if (cell === "softfail") color = "#d29922";
        return `<span style="font-weight: bold; color: ${color};">${cell}</span>`;
      },
    },
  ];

  let matrixColumns = [
    {
      name: $_("Email.From"),
      width: "45%",
    },
    {
      name: $_("Email.To"),
      width: "45%",
    },
    {
      name: $_("Email.Count"),
      width: "10%",
    },
  ];

  let columns = itemsColumns;
  let data = [];
  let pagination = false;

  const updateChart = async () => {
    await tick();
    if (!chartDiv || !reportResult || reportResult.Stats.TotalEmails === 0) {
      disposeEmailChart();
      return;
    }
    if (chartType === "spf") {
      showEmailSPFChart(chartDiv, reportResult.Stats, dark);
    } else if (chartType === "domain") {
      showEmailCountBarChart(
        chartDiv,
        reportResult.Stats.TopDomain,
        dark,
        $_("Email.ModeDomain")
      );
    } else if (chartType === "from") {
      showEmailCountBarChart(
        chartDiv,
        reportResult.Stats.TopFrom,
        dark,
        $_("Email.ModeFrom")
      );
    } else if (chartType === "to") {
      showEmailCountBarChart(
        chartDiv,
        reportResult.Stats.TopTo,
        dark,
        $_("Email.ModeTo")
      );
    } else if (chartType === "senderIP") {
      showEmailCountBarChart(
        chartDiv,
        reportResult.Stats.TopSenderIP,
        dark,
        $_("Email.ModeSenderIP")
      );
    } else if (chartType === "delay") {
      showEmailDelayChart(chartDiv, reportResult.Items, dark);
    }
  };

  const updateTableData = () => {
    if (tableMode === "items") {
      columns = itemsColumns;
      data = (reportResult.Items || []).map((item) => [
        item.Time > 0 ? item.Time / (1000 * 1000) : 0,
        item.From,
        item.To,
        item.Subject,
        item.Delay,
        item.Relay,
        item.SenderIP,
        item.SPF,
      ]);
    } else {
      columns = matrixColumns;
      data = (reportResult.Stats.Matrix || []).map((m) => [
        m.From,
        m.To,
        m.Count,
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

  const loadReport = async (spfCheck) => {
    if (!logs || logs.length === 0) return;
    analyzing = true;
    try {
      const res = await AnalyzeEmailLogs(logs, spfCheck);
      if (res) {
        reportResult = res;
      }
    } catch (e) {
      console.error("AnalyzeEmailLogs error:", e);
    } finally {
      analyzing = false;
    }
    updateTableData();
    await updateChart();
  };

  const handleSPFCheckChange = async () => {
    await loadReport(checkSPF);
  };

  const onResize = () => {
    resizeEmailChart();
  };

  const back = () => {
    disposeEmailChart();
    dispatch("done", {});
  };

  const exportReport = async () => {
    if (!exportType) return;
    saveBusy = true;
    try {
      let exportItems = [];
      if (tableMode === "items") {
        exportItems = (reportResult.Items || []).map((item) => ({
          Time:
            item.Time > 0
              ? echarts.time.format(
                  new Date(item.Time / (1000 * 1000)),
                  "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"
                )
              : "",
          From: item.From,
          To: item.To,
          Subject: item.Subject,
          DelaySec: item.Delay,
          RelayCount: item.Relay,
          SenderIP: item.SenderIP,
          SPF: item.SPF,
          SPFError: item.SPFError,
        }));
      } else {
        exportItems = (reportResult.Stats.Matrix || []).map((m) => ({
          From: m.From,
          To: m.To,
          Count: m.Count,
        }));
      }

      const image = getEmailChartImage();
      await Export(
        exportType,
        $_("Email.ExportTitle"),
        $_("Email.ExportType"),
        image,
        exportItems
      );
    } catch (e) {
      console.error("Export error:", e);
    } finally {
      saveBusy = false;
      exportType = "";
    }
  };

  onMount(async () => {
    await loadReport(false);
  });

  onDestroy(() => {
    disposeEmailChart();
  });
</script>

<svelte:window on:resize={onResize} />
<div
  class="Box mx-auto Box--condensed"
  style="max-width: 99%; display: flex; flex-direction: column; max-height: calc(100vh - 40px);"
>
  <div class="Box-header d-flex flex-items-center flex-wrap">
    <h3 class="Box-title overflow-hidden mr-3">
      {$_("Email.Title")}
      {#if analyzing}
        <span class="Label Label--warning ml-2">{$_("Email.Analyzing")}</span>
      {/if}
    </h3>

    <!-- 統計サマリーバッジ -->
    <div class="d-flex flex-items-center flex-wrap flex-auto f6" style="gap: 6px;">
      <span class="Label Label--info" title="Total">
        {$_("Email.TotalEmails")}: {reportResult.Stats.TotalEmails}
      </span>
      <span class="Label Label--success" title="SPF Pass">
        SPF Pass: {reportResult.Stats.SPFPass}
        {#if reportResult.Stats.TotalEmails > 0}
          ({((reportResult.Stats.SPFPass / reportResult.Stats.TotalEmails) * 100).toFixed(1)}%)
        {/if}
      </span>
      <span class="Label Label--danger" title="SPF Fail">
        Fail: {reportResult.Stats.SPFFail} / Soft: {reportResult.Stats.SPFSoftFail}
      </span>
      <span class="Label" title="Avg Delay">
        {$_("Email.AvgDelay")}: {(reportResult.Stats.AvgDelay || 0).toFixed(1)}s (Max: {reportResult.Stats.MaxDelay || 0}s)
      </span>
      <span class="Label" title="Avg Relay">
        {$_("Email.AvgRelay")}: {(reportResult.Stats.AvgRelay || 0).toFixed(1)} (Max: {reportResult.Stats.MaxRelay || 0})
      </span>
    </div>

    <!-- コントロール（グラフ種類、テーブル種類、SPF検証チェックボックス） -->
    <div class="d-flex flex-items-center ml-2" style="gap: 8px;">
      <!-- svelte-ignore a11y-no-onchange -->
      <select class="form-select select-sm" bind:value={chartType} on:change={updateChart}>
        <option value="spf">{$_("Email.ModeSPF")}</option>
        <option value="domain">{$_("Email.ModeDomain")}</option>
        <option value="from">{$_("Email.ModeFrom")}</option>
        <option value="to">{$_("Email.ModeTo")}</option>
        <option value="senderIP">{$_("Email.ModeSenderIP")}</option>
        <option value="delay">{$_("Email.ModeDelay")}</option>
      </select>

      <!-- svelte-ignore a11y-no-onchange -->
      <select class="form-select select-sm" bind:value={tableMode} on:change={updateTableData}>
        <option value="items">{$_("Email.TableItems")}</option>
        <option value="matrix">{$_("Email.TableMatrix")}</option>
      </select>

      <label class="f6 d-flex flex-items-center mb-0" style="cursor: pointer;">
        <input
          type="checkbox"
          class="mr-1"
          bind:checked={checkSPF}
          on:change={handleSPFCheckChange}
          disabled={analyzing}
        />
        {$_("Email.CheckSPF")}
      </label>
    </div>
  </div>

  <div style="overflow-y: auto; flex: 1;">
    {#if reportResult.Stats.TotalEmails === 0}
      <div class="flash flash-warn m-3">
        {$_("Email.NoData")}
      </div>
    {/if}
    <div class="Box-row" class:d-none={reportResult.Stats.TotalEmails === 0}>
      <div id="emailChart" bind:this={chartDiv}></div>
    </div>
    {#if data.length > 0}
      <div class="Box-row markdown-body log">
        <Grid {data} sort search {pagination} {columns} language={gridLang} />
      </div>
    {/if}
  </div>

  <div class="Box-footer d-flex flex-justify-between">
    <div class="d-flex flex-items-center">
      <button class="btn btn-secondary" type="button" on:click={back}>
        <X16 />
        {$_("Email.BackBtn")}
      </button>
      {#if hasLLM}
        <!-- svelte-ignore a11y-no-onchange -->
        <select class="form-select ml-2" bind:value={selectAILang} on:change={onSelectAILang}>
          <option value="">{$_('ReportAI.AIExplain')}</option>
          <option value="ja">{$_('ReportAI.LangJa')}</option>
          <option value="en">{$_('ReportAI.LangEn')}</option>
        </select>
      {/if}
    </div>
    <div>
      {#if data.length > 0}
        <!-- svelte-ignore a11y-no-onchange -->
        {#if saveBusy}
          <span>{$_("Email.Saving")}</span><span class="AnimatedEllipsis"></span>
        {:else}
          <select class="form-select" bind:value={exportType} on:change={exportReport}>
            <option value="">{$_("Email.ExportBtn")}</option>
            <option value="csv">CSV</option>
            <option value="excel">Excel</option>
          </select>
        {/if}
      {/if}
    </div>
  </div>
</div>

<ReportAIDialog
  bind:show={showAIDialog}
  bind:lang={aiLang}
  title={$_('ReportAI.Title') + ' - ' + $_('Email.Title')}
  exportFilename="email_ai_explanation"
  analyzeFunc={(targetLang) => LLMExplainReport({
    ReportType: "email",
    Title: $_('Email.Title'),
    Params: { chartType: chartType, tableMode: tableMode },
    Summary: reportResult && reportResult.Stats ? `Total: ${reportResult.Stats.TotalEmails}, SPF Pass: ${reportResult.Stats.SPFPass}, Fail: ${reportResult.Stats.SPFFail}, SoftFail: ${reportResult.Stats.SPFSoftFail}, None: ${reportResult.Stats.SPFNone}` : "",
    Columns: columns.map(c => c.name),
    Data: data.slice(0, 60),
    Lang: targetLang
  })}
/>

<style>

  #emailChart {
    width: 100%;
    height: 260px;
    margin: 5px auto;
  }
</style>
