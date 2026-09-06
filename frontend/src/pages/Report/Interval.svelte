<script>
  import { getTableLimit } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, onDestroy, tick } from "svelte";
  import {
    calculateIntervals,
    showIntervalChart,
    showIntervalHistogramChart,
    disposeIntervalChart,
    resizeIntervalChart,
    getIntervalChartImage,
  } from "./interval";
  import Grid from "../../components/Grid.svelte";
  import jaJP from "../../js/gridjsJaJP";
  import { _, getLocale } from "../../i18n/i18n";
  import { Export, LLMExplainReport } from "../../../wailsjs/go/main/App";
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
  let chartType = "scatter"; // "scatter" | "histogram"
  let data = [];
  let intervalResult = {
    items: [],
    deltas: [],
    stats: {
      count: 0,
      mean: 0,
      median: 0,
      mode: 0,
      min: 0,
      max: 0,
      stddev: 0,
    },
  };

  let columns = [
    {
      name: $_("Interval.Time"),
      width: "20%",
      formatter: (cell) => {
        if (!cell) return "";
        const d = new Date(cell);
        return echarts.time.format(d, "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}");
      },
    },
    {
      name: $_("Interval.DeltaSec"),
      width: "12%",
      formatter: (cell) => (cell !== undefined ? Number(cell).toFixed(3) : ""),
    },
    {
      name: $_("Interval.Log"),
      width: "68%",
    },
  ];

  let pagination = false;

  const updateChart = async () => {
    await tick();
    if (!chartDiv || intervalResult.deltas.length === 0) {
      disposeIntervalChart();
      return;
    }
    if (chartType === "scatter") {
      showIntervalChart(chartDiv, intervalResult.items, dark, "scatter");
    } else {
      showIntervalHistogramChart(chartDiv, intervalResult.deltas, dark);
    }
  };

  const initData = async () => {
    intervalResult = calculateIntervals(logs);
    data = intervalResult.items.map((item) => [
      item.time / (1000 * 1000),
      item.delta,
      item.log,
    ]);

    if (data.length > 10) {
      pagination = {
        limit: getTableLimit(),
        enable: true,
      };
    } else {
      pagination = false;
    }

    await updateChart();
  };

  onMount(() => {
    initData();
  });

  onDestroy(() => {
    disposeIntervalChart();
  });

  let exportType = "";
  let saveBusy = false;
  const exportReport = async () => {
    if (exportType == "") {
      return;
    }
    saveBusy = true;
    const exportData = {
      Type: $_("Interval.ExportType"),
      Title: $_("Interval.ExportTitle"),
      Header: [],
      Data: [],
      Image: "",
    };
    if (exportType == "excel") {
      exportData.Image = getIntervalChartImage();
    }
    columns.forEach((e) => {
      exportData.Header.push(e.name);
    });
    data.forEach((l) => {
      const row = [];
      l.forEach((e, i) => {
        const v = columns[i] && columns[i].formatter ? columns[i].formatter(e) : e;
        row.push(v);
      });
      exportData.Data.push(row);
    });
    await Export(exportType, exportData);
    saveBusy = false;
    exportType = "";
  };

  const onResize = () => {
    resizeIntervalChart();
  };

  const back = () => {
    disposeIntervalChart();
    dispatch("done", {});
  };
</script>

<svelte:window on:resize={onResize} />
<div
  class="Box mx-auto Box--condensed"
  style="max-width: 99%; display: flex; flex-direction: column; max-height: calc(100vh - 40px);"
>
  <div class="Box-header d-flex flex-items-center flex-wrap">
    <h3 class="Box-title overflow-hidden mr-3">{$_("Interval.Title")}</h3>

    <!-- 統計サマリーバッジ -->
    <div class="d-flex flex-items-center flex-wrap flex-auto f6" style="gap: 6px;">
      <span class="Label Label--info" title="Mean">
        {$_("Interval.Mean")}: {intervalResult.stats.mean.toFixed(2)}s
      </span>
      <span class="Label Label--info" title="Median">
        {$_("Interval.Median")}: {intervalResult.stats.median.toFixed(2)}s
      </span>
      <span class="Label Label--info" title="Mode">
        {$_("Interval.Mode")}: ~{Number(intervalResult.stats.mode).toFixed(2)}s
      </span>
      <span class="Label" title="StdDev">
        {$_("Interval.StdDev")}: {intervalResult.stats.stddev.toFixed(2)}s
      </span>
      <span class="Label" title="Min/Max">
        Min: {intervalResult.stats.min.toFixed(2)}s / Max: {intervalResult.stats.max.toFixed(2)}s
      </span>
    </div>

    <!-- グラフ種別セレクタ -->
    <div class="ml-2">
      <!-- svelte-ignore a11y-no-onchange -->
      <select class="form-select select-sm" bind:value={chartType} on:change={updateChart}>
        <option value="scatter">{$_("Interval.ScatterView")}</option>
        <option value="histogram">{$_("Interval.HistogramView")}</option>
      </select>
    </div>
  </div>

  <div style="overflow-y: auto; flex: 1;">
    <div class="Box-row">
      <div id="intervalChart" bind:this={chartDiv}></div>
    </div>
    <div class="Box-row markdown-body log">
      <Grid {data} sort search {pagination} {columns} language={gridLang} autoWidth={false} />
    </div>
  </div>

  <div class="Box-footer d-flex flex-justify-between">
    <div class="d-flex flex-items-center">
      <button class="btn btn-secondary" type="button" on:click={back}>
        <X16 />
        {$_("Interval.BackBtn")}
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
          <span>{$_("Interval.Saving")}</span><span class="AnimatedEllipsis"></span>
        {:else}
          <select class="form-select" bind:value={exportType} on:change={exportReport}>
            <option value="">{$_("Interval.ExportBtn")}</option>
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
  title={$_('ReportAI.Title') + ' - ' + $_('Interval.Title')}
  exportFilename="interval_ai_explanation"
  analyzeFunc={(targetLang) => LLMExplainReport({
    ReportType: "interval",
    Title: $_('Interval.Title'),
    Params: { chartType: chartType },
    Summary: intervalResult ? `Count: ${intervalResult.count}, Mean: ${intervalResult.mean?.toFixed(2)}s, Median: ${intervalResult.median?.toFixed(2)}s, Variance: ${intervalResult.variance?.toFixed(2)}, Min: ${intervalResult.min?.toFixed(2)}s, Max: ${intervalResult.max?.toFixed(2)}s` : "",
    Columns: columns.map(c => c.name),
    Data: data.slice(0, 60),
    Lang: targetLang
  })}
/>

<style>
  #intervalChart {
    width: 100%;
    height: 260px;
    margin: 5px auto;
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
