<script>
  import * as echarts from "echarts";
  import { getFieldName, getFields,getTableLimit } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, onDestroy, tick } from "svelte";
  import { getFFTMap, showFFTChart, resizeFFTChart, getFFTChartImage, disposeFFTChart } from "./fft";
  import Grid from "../../components/Grid.svelte";
  import jaJP from "../../js/gridjsJaJP";
  import { _,getLocale } from '../../i18n/i18n';
  import {Export, LLMExplainReport} from '../../../wailsjs/go/main/App';
  import ReportAIDialog from "../../components/ReportAIDialog.svelte";

  export let logs = [];
  export let fields = [];
  export let dark = false;
  export let hasLLM = false;

  let locale = getLocale();
  let showAIDialog = false;
  let aiLang = locale === "en" ? "en" : "ja";
  let selectAILang = "";

  const onSelectAILang = () => {
    if (!selectAILang) return;
    aiLang = selectAILang;
    showAIDialog = true;
    selectAILang = "";
  };
  let catFields = [];

  let keyList = [];
  let selected = "";
  let key = "";
  let fftType = "hz";
  let fftMap;

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
    columns[1].name = selected === "_None" ? $_('FFT.NoItem') : (getFieldName(selected) || selected);
    logs.forEach((l) => {
      if (selected === "_None") {
        data.push([l.Time, "-"]);
        return;
      }
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
    const mode = showFFTChart("chart", selected, fftMap, key, fftType, dark);
    if (mode === "fallback_2d") {
      key = "Total";
    }
  };

  onMount(() => {
    catFields = getFields(fields, "string");
  });

  onDestroy(() => {
    disposeFFTChart();
  });

  let exportType = '';
  let saveBusy = false;
  const exportReport = async () => {
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
    await Export(exportType,exportData);
    saveBusy = false;
    exportType = "";
  }

  const onResize = () => {
    resizeFFTChart();
  };

  const back = () => {
    dispatch("done", {});
  };
</script>

<svelte:window on:resize={onResize} />
<div class="Box mx-auto Box--condensed" style="max-width: 99%; display: flex; flex-direction: column; max-height: calc(100vh - 40px);">
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">{$_('FFT.Title')}</h3>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select"
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
        bind:value={fftType}
        on:change={updateFFTChart}
      >
        <option value="hz">{$_('FFT.Freq')}</option>
        <option value="time">{$_('FFT.Cycle')}</option>
      </select>
    {/if}
  </div>
  <div style="overflow-y: auto; flex: 1;">
    <div class="Box-row">
      <div id="chart"></div>
    </div>
    <div class="Box-row markdown-body log">
      <Grid {data} sort search {pagination} {columns} language={gridLang} />
    </div>
  </div>
  <div class="Box-footer d-flex flex-justify-between">
    <div class="d-flex flex-items-center">
      <button class="btn btn-secondary" type="button" on:click={back}>
        <X16 />
        {$_('FFT.BackBtn')}
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
          <span>{$_('FFT.Saving')}</span><span class="AnimatedEllipsis"></span>
        {:else}
          <select class="form-select" bind:value={exportType} on:change="{exportReport}">
            <option value="">{$_('FFT.ExportBtn')}</option>
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
  title={$_('ReportAI.Title') + ' - ' + $_('FFT.Title')}
  exportFilename="fft_ai_explanation"
  analyzeFunc={(targetLang) => LLMExplainReport({
    ReportType: "fft",
    Title: $_('FFT.Title'),
    Field: getFieldName(selected) || selected,
    Params: { key: key, fftType: fftType },
    Columns: columns.map(c => c.name),

    Data: data.slice(0, 60),
    Lang: targetLang
  })}
/>

<style>

  #chart {
    width: 100%;
    height: 350px;
    margin: 5px auto;
  }
</style>
