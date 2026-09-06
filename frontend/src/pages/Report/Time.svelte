<script>
  import * as echarts from 'echarts'
  import { getFieldName,getFields,getTableLimit } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, onDestroy, tick } from "svelte";
  import { showTimeChart, resizeTimeChart, getTimeChartImage, disposeTimeChart } from "./time";
  import Grid from "../../components/Grid.svelte";
  import jaJP from "../../js/gridjsJaJP";
  import { _,getLocale } from '../../i18n/i18n';
  import {Export, LLMExplainReport} from '../../../wailsjs/go/main/App';
  import ReportAIDialog from "../../components/ReportAIDialog.svelte";

  let locale = getLocale();
  let gridLang = locale == "ja" ? jaJP : undefined;

  export let logs = [];
  export let fields = [];
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

  let chartDiv;
  let numFields = [];
  let field = "";
  let chartType = "";

  const dispatch = createEventDispatcher();
  let data = [];
  let columns = [];
  let pagination = false;

  const updateTime = async () => {
    await tick();
    if (!chartDiv || field == "") {
      disposeTimeChart();
      return;
    }
    data = showTimeChart(chartDiv, logs, field, chartType, dark);
    if(chartType == "1h" || chartType == "1m" ){
      columns = [
        {
          name: $_('Time.DateAndTime'),
          formatter: (cell) => echarts.time.format(new Date(cell/(1000*1000)), '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
        },{
          name: $_('Time.MeanValue'),
        },{
          name: $_('Time.MaxValue'),
        },{
          name: $_('Time.MinValue'),
        },{
          name: $_('Time.MedianValue'),
        },{
          name: $_('Time.VariantValue'),
        }
      ];
    } else if (chartType != ""){
      columns = [
        {
          name: $_('Time.DateAndTime'),
          formatter: (cell) => echarts.time.format(new Date(cell), '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
        },{
          name: $_('Time.Regression ',{values:{chartType:chartType}}),
        }
      ];
    } else {
      columns = [
        {
          name: $_('Time.DateAndTime'),
          formatter: (cell) => echarts.time.format(new Date(cell/(1000*1000)), '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
        },{
          name: getFieldName(field),
        }
      ];

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

  onMount(async () => {
    numFields = getFields(fields,"number");
    if(field == "" && numFields.length > 0 ){
      field = numFields[0];
    }
    await updateTime();
  });

  onDestroy(() => {
    disposeTimeChart();
  });

  let exportType = '';
  let saveBusy = false;
  const exportReport = async () => {
    if (exportType == "") {
      return;
    }
    saveBusy = true;
    const exportData = {
      Type: $_('Time.ExportType'),
      Title: $_('Time.ExportTitle'),
      Header: [],
      Data: [],
      Image: "",
    };
    if (exportType == "excel") {
      exportData.Image = getTimeChartImage();
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
    resizeTimeChart();
  };

  const back = () => {
    disposeTimeChart();
    dispatch("done", {});
  };

</script>

<svelte:window on:resize={onResize} />
<div class="Box mx-auto Box--condensed" style="max-width: 99%; display: flex; flex-direction: column; max-height: calc(100vh - 40px);">
  <!-- svelte-ignore a11y-no-onchange -->
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">{$_('Time.Title')}</h3>
    <select
      class="form-select"
      bind:value={field}
      on:change="{updateTime}"
    >
      <option value="">{$_('Time.SelectItemMsg')}</option>
      {#each numFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <select
      class="form-select ml-2"
      bind:value={chartType}
      on:change="{updateTime}"
    >
      <option value="">{$_('Time.RealData')}</option>
      <option value="1m">{$_('Time.Sum1Min')}</option>
      <option value="1h">{$_('Time.Sum1H')}</option>
      <option value="linear">{$_('Time.RegressionLinear')}</option>
      <option value="exponential">{$_('Time.RegressionExponential')}</option>
      <option value="logarithmic">{$_('Time.RegrfessionLogarithmic')}</option>
      <option value="polynomial">{$_('Time.RegressionPlynomial')}</option>
    </select>
  </div>
  <div style="overflow-y: auto; flex: 1;">
    <div class="Box-row">
      <div id="timeChart" bind:this={chartDiv}></div>
    </div>
    <div class="Box-row markdown-body log">
      <Grid {data} sort search {pagination} {columns} language={gridLang} />
    </div>
  </div>
  <div class="Box-footer d-flex flex-justify-between">
    <div class="d-flex flex-items-center">
      <button class="btn btn-secondary" type="button" on:click={back}>
        <X16 />
        {$_('Time.BackBtn')}
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
          <span>{$_('Time.Saving')}</span><span class="AnimatedEllipsis"></span>
        {:else}
          <select class="form-select" bind:value={exportType} on:change="{exportReport}">
            <option value="">{$_('Time.ExportBtn')}</option>
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
  title={$_('ReportAI.Title') + ' - ' + $_('Time.Title')}
  exportFilename="time_ai_explanation"
  analyzeFunc={(targetLang) => LLMExplainReport({
    ReportType: "time",
    Title: $_('Time.Title'),
    Field: getFieldName(field) || field,
    Params: { chartType: chartType },
    Columns: columns.map(c => c.name),

    Data: data.slice(0, 60),
    Lang: targetLang
  })}
/>

<style>

  #timeChart {
    width: 100%;
    height: 300px;
    margin: 5px auto;
  }
</style>
