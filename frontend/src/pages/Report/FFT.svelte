<script>
  import * as echarts from "echarts";
  import { getFieldName, getFields,getTableLimit } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { getFFTMap, showFFTChart, resizeFFTChart,getFFTChartImage } from "./fft";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";
  import { _,getLocale } from '../../i18n/i18n';

  export let logs = [];
  export let fields = [];
  let dark = false;
  let catFields = [];
  let keyList = [];
  let selected = "";
  let key = "";
  let fftType = "hz";
  let fftMap;

  let locale = getLocale();
  let gridLang = locale == "ja" ? jaJP : undefined;

  const dispatch = createEventDispatcher();
  let data = [];
  let columns = [
    {
      name: $_('FFT.DateTime'),
      width: "80%",
      formatter: (cell) =>
        echarts.time.format(
          new Date(cell / (1000 * 1000)),
          "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"
        ),
    },
    {
      name: $_('FFT.Item1'),
      width: "20%",
    },
  ];

  let pagination = false;

  const updateFFT = () => {
    if (selected == "") {
      return;
    }
    fftMap = getFFTMap(logs, selected);
    keyList = [];
    for (let key of fftMap.keys()) {
      keyList.push(key);
    }
    data = [];
    columns[1].name = getFieldName(selected) || selected;
    logs.forEach((l) => {
      const v = l.KeyValue[selected];
      switch (typeof v) {
        case "string":
        case "number":
        case "boolean":
          break;
        default:
          return;
      }
      data.push([l.Time, l.KeyValue[selected] || ""]);
    });
    if (data.length > 10) {
      pagination = {
        limit: getTableLimit(),
        enable: true,
      };
    } else {
      pagination = false;
    }
    updateFFTChart();
  };

  const updateFFTChart = async () => {
    if (!fftMap) {
      return;
    }
    await tick();
    showFFTChart("chart", selected, fftMap, key, fftType, dark);
  };

  onMount(() => {
    catFields = getFields(fields, "string");
    window.go.main.App.GetDark().then((v) => {
      dark = v;
    });
  });

  let exportType = '';
  let saveBusy = false;
  const exportReport = () => {
    if (exportType == "") {
      return;
    }
    saveBusy = true;
    const exportData = {
      Type: $_('FFT.ExportType'),
      Title: $_('FFT.ExportTitle'),
      Header: [],
      Data: [],
      Image: "",
    };
    if (exportType == "excel") {
      exportData.Image = getFFTChartImage();
    }
    columns.forEach((e)=>{
      exportData.Header.push(e.name);
    });
    data.forEach((l)=>{
      const row = [];
      l.forEach((e,i)=>{
        const v = columns[i] && columns[i].formatter ? columns[i].formatter(e) : e;
        row.push(v);
      });
      exportData.Data.push(row);
    });
    window.go.main.App.Export(exportType,exportData).then(()=>{
      saveBusy = false;
      exportType = "";
    });
  }

  const onResize = () => {
    resizeFFTChart();
  };

  const back = () => {
    dispatch("done", {});
  };
</script>

<svelte:window on:resize={onResize} />
<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">{$_('FFT.Title')}</h3>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select"
      aria-label="{$_('FFT.SumItem')}"
      bind:value={selected}
      on:change={updateFFT}
    >
      <option value="">{$_('FFT.SelectSumItem')}</option>
      {#if catFields.length < 1 }
        <option value="_None">{$_('FFT.NoItem')}</option>
      {/if}
      {#each catFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    {#if keyList.length > 0}
      <!-- svelte-ignore a11y-no-onchange -->
      <select
        class="form-select ml-2"
        aria-label="{$_('FFT.DisplayItem')}"
        bind:value={key}
        on:change={updateFFTChart}
      >
        <option value="">{$_('FFT.3DChart')}</option>
        {#each keyList as k}
          <option value={k}>{k}</option>
        {/each}
      </select>
      <!-- svelte-ignore a11y-no-onchange -->
      <select
        class="form-select ml-2"
        aria-label="{$_('FFT.CharMode')}"
        bind:value={fftType}
        on:change={updateFFTChart}
      >
        <option value="hz">{$_('FFT.Freq')}</option>
        <option value="time">{$_('FFT.Cycle')}</option>
      </select>
    {/if}
  </div>
  <div class="Box-row">
    <div id="chart" />
  </div>
  <div class="Box-row markdown-body log">
    <Grid {data} sort search {pagination} {columns} language={gridLang} />
  </div>
  <div class="Box-footer text-right">
    {#if data.length > 0}
      <!-- svelte-ignore a11y-no-onchange -->
      {#if saveBusy}
        <span>{$_('FFT.Saving')}</span><span class="AnimatedEllipsis"></span>
      {:else}
        <select class="form-select" bind:value={exportType} on:change="{exportReport}">
          <option value="">{$_('FFT.ExportBtn')}</option>
          <option value="csv">CSV</option>
          <option value="excel">Excel</option>
        </select>
      {/if}
    {/if}
    <button class="btn btn-secondary" type="button" on:click={back}>
      <X16 />
      {$_('FFT.BackBtn')}
    </button>
  </div>
</div>

<style>
  #chart {
    width: 100%;
    height: 500px;
    margin: 5px auto;
  }
</style>
