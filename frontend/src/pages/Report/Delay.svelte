<script>
  import { getTableLimit, getFields, getFieldName } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, onDestroy, tick } from "svelte";
  import {
    calculateDelays,
    showDelayChart,
    disposeDelayChart,
    resizeDelayChart,
    getDelayChartImage,
  } from "./delay";
  import Grid from "../../components/Grid.svelte";
  import jaJP from "../../js/gridjsJaJP";
  import { _, getLocale } from "../../i18n/i18n";
  import { Export } from "../../../wailsjs/go/main/App";
  import * as echarts from "echarts";

  let locale = getLocale();
  let gridLang = locale == "ja" ? jaJP : undefined;

  export let logs = [];
  export let fields = [];
  export let dark = false;

  const dispatch = createEventDispatcher();

  let chartDiv;
  let mode = "delta"; // "delta" | "field"
  let selectedField = "";
  let numFields = [];
  let data = [];
  let delayResult = {
    items: [],
    delays: [],
    stats: {
      count: 0,
      max: 0,
      mean: 0,
      median: 0,
      min: 0,
    },
  };

  let columns = [
    {
      name: $_("Delay.Rank"),
      width: "10%",
    },
    {
      name: $_("Delay.Time"),
      width: "25%",
      formatter: (cell) => {
        if (!cell) return "";
        const d = new Date(cell);
        return echarts.time.format(d, "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}");
      },
    },
    {
      name: $_("Delay.DelaySec"),
      width: "15%",
      formatter: (cell) => (cell !== undefined ? Number(cell).toFixed(3) : ""),
    },
    {
      name: $_("Delay.Log"),
      width: "50%",
    },
  ];

  let pagination = false;

  const updateChart = async () => {
    await tick();
    if (!chartDiv || delayResult.stats.count === 0) {
      disposeDelayChart();
      return;
    }
    showDelayChart(
      chartDiv,
      delayResult.items,
      dark,
      mode === "field" ? selectedField : ""
    );
  };

  const updateDelayData = async () => {
    delayResult = calculateDelays(logs, mode, selectedField);
    data = delayResult.items.map((item, idx) => [
      idx + 1,
      item.time / (1000 * 1000),
      item.delay,
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

  const handleModeChange = () => {
    if (mode === "field" && !selectedField && numFields.length > 0) {
      selectedField = numFields[0];
    }
    updateDelayData();
  };

  onMount(() => {
    numFields = getFields(fields, "number").filter((f) => f !== "delta");
    updateDelayData();
  });

  onDestroy(() => {
    disposeDelayChart();
  });

  let exportType = "";
  let saveBusy = false;
  const exportReport = async () => {
    if (exportType == "") {
      return;
    }
    saveBusy = true;
    const exportData = {
      Type: $_("Delay.ExportType"),
      Title: $_("Delay.ExportTitle"),
      Header: [],
      Data: [],
      Image: "",
    };
    if (exportType == "excel") {
      exportData.Image = getDelayChartImage();
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
    resizeDelayChart();
  };

  const back = () => {
    disposeDelayChart();
    dispatch("done", {});
  };
</script>

<svelte:window on:resize={onResize} />
<div
  class="Box mx-auto Box--condensed"
  style="max-width: 99%; display: flex; flex-direction: column; max-height: calc(100vh - 40px);"
>
  <div class="Box-header d-flex flex-items-center flex-wrap">
    <h3 class="Box-title overflow-hidden mr-3">{$_("Delay.Title")}</h3>

    <!-- 検出モード選択 -->
    <div class="d-flex flex-items-center mr-3">
      <!-- svelte-ignore a11y-no-onchange -->
      <select class="form-select select-sm mr-2" bind:value={mode} on:change={handleModeChange}>
        <option value="delta">{$_("Delay.ModeDelta")}</option>
        {#if numFields.length > 0}
          <option value="field">{$_("Delay.ModeField")}</option>
        {/if}
      </select>

      {#if mode === "field"}
        <!-- svelte-ignore a11y-no-onchange -->
        <select
          class="form-select select-sm"
          bind:value={selectedField}
          on:change={updateDelayData}
        >
          {#each numFields as f}
            <option value={f}>{getFieldName(f)}</option>
          {/each}
        </select>
      {/if}
    </div>

    <!-- 統計サマリーバッジ -->
    <div class="d-flex flex-items-center flex-wrap flex-auto f6" style="gap: 6px;">
      <span class="Label Label--info" title="Hits">
        {$_("Delay.Hits")}: {delayResult.stats.count}
      </span>
      {#if delayResult.stats.count > 0}
        <span class="Label Label--danger" title="Max">
          Max: {delayResult.stats.max.toFixed(2)}s
        </span>
        <span class="Label Label--warning" title="Mean">
          {$_("Delay.Mean")}: {delayResult.stats.mean.toFixed(2)}s
        </span>
        <span class="Label" title="Median">
          {$_("Delay.Median")}: {delayResult.stats.median.toFixed(2)}s
        </span>
      {/if}
    </div>
  </div>

  <div style="overflow-y: auto; flex: 1;">
    {#if delayResult.stats.count === 0}
      <div class="flash flash-warn m-3">
        {$_("Delay.NoDelayMsg")}
      </div>
    {/if}
    <div class="Box-row" class:d-none={delayResult.stats.count === 0}>
      <div id="delayChart" bind:this={chartDiv}></div>
    </div>
    {#if delayResult.stats.count > 0}
      <div class="Box-row markdown-body log">
        <Grid {data} sort search {pagination} {columns} language={gridLang} />
      </div>
    {/if}
  </div>

  <div class="Box-footer d-flex flex-justify-between">
    <div>
      <button class="btn btn-secondary" type="button" on:click={back}>
        <X16 />
        {$_("Delay.BackBtn")}
      </button>
    </div>
    <div>
      {#if data.length > 0}
        <!-- svelte-ignore a11y-no-onchange -->
        {#if saveBusy}
          <span>{$_("Delay.Saving")}</span><span class="AnimatedEllipsis"></span>
        {:else}
          <select class="form-select" bind:value={exportType} on:change={exportReport}>
            <option value="">{$_("Delay.ExportBtn")}</option>
            <option value="csv">CSV</option>
            <option value="excel">Excel</option>
          </select>
        {/if}
      {/if}
    </div>
  </div>
</div>

<style>
  #delayChart {
    width: 100%;
    height: 260px;
    margin: 5px auto;
  }
</style>
